package generator

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/humbleshuttler/repodiagram/internal/prompts"
	openai "github.com/sashabaranov/go-openai"
)

type Result struct {
	Diagram   string
	Explanation string
	Mapping   string
}

type Generator struct {
	client  *openai.Client
	model   string
	verbose bool
}

func New(apiKey, model string, verbose bool) *Generator {
	return &Generator{
		client:  openai.NewClient(apiKey),
		model:   model,
		verbose: verbose,
	}
}

func (g *Generator) GenerateDiagram(fileTree, readme, instructions string) (*Result, error) {
	ctx := context.Background()
	info := color.New(color.FgCyan)

	if g.verbose {
		info.Fprintf(os.Stderr, "  Phase 1/3: Analyzing repository structure...\n")
	}
	explanation, err := g.generateExplanation(ctx, fileTree, readme)
	if err != nil {
		return nil, fmt.Errorf("phase 1 failed: %w", err)
	}

	if g.verbose {
		info.Fprintf(os.Stderr, "  Phase 2/3: Mapping components to files...\n")
	}
	mapping, err := g.generateMapping(ctx, explanation, fileTree)
	if err != nil {
		return nil, fmt.Errorf("phase 2 failed: %w", err)
	}

	if g.verbose {
		info.Fprintf(os.Stderr, "  Phase 3/3: Generating Mermaid diagram...\n")
	}
	diagram, err := g.generateDiagram(ctx, explanation, mapping, instructions)
	if err != nil {
		return nil, fmt.Errorf("phase 3 failed: %w", err)
	}

	return &Result{
		Diagram:     diagram,
		Explanation: explanation,
		Mapping:     mapping,
	}, nil
}

func (g *Generator) generateExplanation(ctx context.Context, fileTree, readme string) (string, error) {
	userMsg := prompts.FormatFirstPrompt(fileTree, readme)

	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: g.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: prompts.SystemFirstPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMsg},
		},
	})
	if err != nil {
		return "", err
	}

	content := resp.Choices[0].Message.Content
	return extractTag(content, "explanation"), nil
}

func (g *Generator) generateMapping(ctx context.Context, explanation, fileTree string) (string, error) {
	userMsg := prompts.FormatSecondPrompt(explanation, fileTree)

	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: g.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: prompts.SystemSecondPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMsg},
		},
	})
	if err != nil {
		return "", err
	}

	content := resp.Choices[0].Message.Content
	return extractTag(content, "component_mapping"), nil
}

func (g *Generator) generateDiagram(ctx context.Context, explanation, mapping, instructions string) (string, error) {
	userMsg := prompts.FormatThirdPrompt(explanation, mapping, instructions)
	systemPrompt := prompts.GetThirdSystemPrompt(instructions != "")

	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: g.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMsg},
		},
	})
	if err != nil {
		return "", err
	}

	content := resp.Choices[0].Message.Content
	content = cleanMermaidOutput(content)

	if content == "BAD_INSTRUCTIONS" {
		return "", fmt.Errorf("the provided instructions were invalid or unclear")
	}

	return content, nil
}

func (g *Generator) GenerateDiagramStreaming(fileTree, readme, instructions string, w io.Writer) (*Result, error) {
	ctx := context.Background()

	explanation, err := g.streamPhase(ctx, "Phase 1/3: Analyzing...", prompts.SystemFirstPrompt, prompts.FormatFirstPrompt(fileTree, readme), w)
	if err != nil {
		return nil, err
	}
	explanation = extractTag(explanation, "explanation")

	mapping, err := g.streamPhase(ctx, "Phase 2/3: Mapping...", prompts.SystemSecondPrompt, prompts.FormatSecondPrompt(explanation, fileTree), w)
	if err != nil {
		return nil, err
	}
	mapping = extractTag(mapping, "component_mapping")

	systemPrompt := prompts.GetThirdSystemPrompt(instructions != "")
	diagram, err := g.streamPhase(ctx, "Phase 3/3: Generating...", systemPrompt, prompts.FormatThirdPrompt(explanation, mapping, instructions), w)
	if err != nil {
		return nil, err
	}
	diagram = cleanMermaidOutput(diagram)

	return &Result{
		Diagram:     diagram,
		Explanation: explanation,
		Mapping:     mapping,
	}, nil
}

func (g *Generator) streamPhase(ctx context.Context, phase, systemPrompt, userMsg string, w io.Writer) (string, error) {
	if g.verbose {
		fmt.Fprintf(w, "%s\n", phase)
	}

	stream, err := g.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: g.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMsg},
		},
		Stream: true,
	})
	if err != nil {
		return "", err
	}
	defer stream.Close()

	var result strings.Builder
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		chunk := response.Choices[0].Delta.Content
		result.WriteString(chunk)
	}

	return result.String(), nil
}

func extractTag(content, tag string) string {
	pattern := fmt.Sprintf(`<%s>([\s\S]*?)</%s>`, tag, tag)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(content)
}

func cleanMermaidOutput(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```mermaid") {
		content = strings.TrimPrefix(content, "```mermaid")
	}
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
	}
	if strings.HasSuffix(content, "```") {
		content = strings.TrimSuffix(content, "```")
	}
	return strings.TrimSpace(content)
}
