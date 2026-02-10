package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/humbledshuttler/repodiagram/internal/generator"
	"github.com/humbledshuttler/repodiagram/internal/output"
	"github.com/humbledshuttler/repodiagram/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	outputFile   string
	outputFormat string
	instructions string
	apiKey       string
	model        string
	verbose      bool
	noClick      bool
)

var rootCmd = &cobra.Command{
	Use:   "repodiagram [path]",
	Short: "Generate architecture diagrams from local repositories",
	Long: `RepoDiagram generates Mermaid.js architecture diagrams from any local repository.
It analyzes the file structure and README to create an interactive system design diagram.

Examples:
  repodiagram                              # Current directory
  repodiagram ./my-project                 # Specific directory
  repodiagram -o diagram.mmd              # Output to file
  repodiagram -f html -o diagram.html     # HTML preview
  repodiagram -i "Focus on the API layer" # Custom instructions`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "mermaid", "Output format: mermaid, html")
	rootCmd.Flags().StringVarP(&instructions, "instructions", "i", "", "Custom instructions for diagram generation")
	rootCmd.Flags().StringVar(&apiKey, "api-key", "", "OpenAI API key (or use OPENAI_API_KEY env var)")
	rootCmd.Flags().StringVar(&model, "model", "gpt-4o-mini", "OpenAI model to use")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show generation progress")
	rootCmd.Flags().BoolVar(&noClick, "no-click", false, "Disable click events in output")
}

func run(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	key := apiKey
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	if key == "" {
		return fmt.Errorf("OpenAI API key required. Set OPENAI_API_KEY env var or use --api-key flag")
	}

	info := color.New(color.FgCyan)
	success := color.New(color.FgGreen)
	
	if verbose {
		info.Fprintf(os.Stderr, "Scanning directory: %s\n", path)
	}

	fileTree, err := scanner.ScanDirectory(path)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if verbose {
		info.Fprintf(os.Stderr, "Found %d lines in file tree\n", countLines(fileTree))
	}

	readme, err := scanner.FindReadme(path)
	if err != nil {
		if verbose {
			info.Fprintf(os.Stderr, "No README found, continuing without it\n")
		}
		readme = ""
	} else if verbose {
		info.Fprintf(os.Stderr, "Found README (%d bytes)\n", len(readme))
	}

	if verbose {
		info.Fprintf(os.Stderr, "Generating diagram using %s...\n", model)
	}

	gen := generator.New(key, model, verbose)
	result, err := gen.GenerateDiagram(fileTree, readme, instructions)
	if err != nil {
		return fmt.Errorf("failed to generate diagram: %w", err)
	}

	diagram := result.Diagram
	if noClick {
		diagram = output.RemoveClickEvents(diagram)
	}

	var out string
	switch outputFormat {
	case "html":
		out = output.ToHTML(diagram)
	case "mermaid":
		out = diagram
	default:
		return fmt.Errorf("unknown format: %s (supported: mermaid, html)", outputFormat)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(out), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		if verbose {
			success.Fprintf(os.Stderr, "Diagram written to %s\n", outputFile)
		}
	} else {
		fmt.Print(out)
	}

	return nil
}

func countLines(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}
