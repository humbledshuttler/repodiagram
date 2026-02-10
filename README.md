# RepoDiagram

Inspited by [GitDiagram](https://gitdiagram.com),  a CLI tool that generates architecture diagrams from any local repository using AI.

Turn any local repository into a visual architecture diagram. RepoDiagram scans your file structure and README, then uses AI to generate interactive Mermaid.js diagrams - all from your terminal.

## Why RepoDiagram?

[GitDiagram](https://gitdiagram.com) is great for quickly visualizing public GitHub repositories. But it has limitations:

- **Private repos require a GitHub token** - not everyone wants to grant access
- **Local-only repos aren't supported** - projects not pushed to GitHub can't be visualized
- **Requires internet for repo access** - can't work with local-only codebases
- **No CLI workflow** - you have to leave your terminal and use a browser

RepoDiagram solves these by running entirely locally. Point it at any directory on your machine, and get the same AI-powered architecture diagrams - no GitHub, no tokens, no browser needed.

## Installation

### From Source

```bash
go install github.com/humbledshuttler/repodiagram@latest
```

### Build Locally

```bash
git clone https://github.com/humbledshuttler/repodiagram.git
cd repodiagram
make build
```

## Usage

```bash
# Generate diagram for current directory
repodiagram

# Generate for a specific directory
repodiagram ./my-project

# Output to file
repodiagram -o architecture.mmd

# Generate HTML preview
repodiagram -f html -o diagram.html

# With custom instructions
repodiagram -i "Focus on the API layer"

# Verbose mode
repodiagram -v
```

## Options

| Flag | Description |
|------|-------------|
| `-o, --output` | Output file (default: stdout) |
| `-f, --format` | Output format: `mermaid` (default), `html` |
| `-i, --instructions` | Custom instructions for diagram generation |
| `--api-key` | OpenAI API key (or use `OPENAI_API_KEY` env var) |
| `--model` | OpenAI model (default: `gpt-4o-mini`) |
| `-v, --verbose` | Show generation progress |
| `--no-click` | Disable click events in output |

## Configuration

Set your OpenAI API key:

```bash
export OPENAI_API_KEY=sk-...
```

Or pass it directly:

```bash
repodiagram --api-key sk-...
```

## Output Formats

### Mermaid (default)

Raw Mermaid.js code that can be used in:
- GitHub/GitLab markdown
- Notion
- Obsidian
- Any Mermaid-compatible tool

### HTML

Self-contained HTML file with embedded Mermaid.js renderer. Open in any browser.

## How It Works

1. **Scans** your repository's file structure (respects `.gitignore`)
2. **Reads** the README for context
3. **Analyzes** the structure using AI (3-phase prompt pipeline)
4. **Generates** a Mermaid.js diagram with clickable components

## Cross-Platform Builds

```bash
make build-all
```

Creates binaries for:
- macOS (Intel & Apple Silicon)
- Linux (AMD64 & ARM64)
- Windows (AMD64)

## License

MIT
