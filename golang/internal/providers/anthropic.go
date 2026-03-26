package providers

import (
	"context"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

const defaultAnthropicModel = "claude-3-5-sonnet"

// Anthropic is a real Anthropic provider backed by the official Go client.
// Construct one via NewAnthropic and inject it into Activities.
// API key and model are read from environment variables during construction.
type Anthropic struct {
	client anthropic.Client
	model  string
}

// NewAnthropic constructs an Anthropic provider using the given API key and model name.
// If model is empty it defaults to claude-3-5-sonnet.
func NewAnthropic(apiKey, model string) *Anthropic {
	if model == "" {
		model = defaultAnthropicModel
	}

	return &Anthropic{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
	}
}

func (a *Anthropic) Name() models.ProviderName { return models.ProviderAnthropic }

// Summarise concatenates all chunks and asks Anthropic for a concise summary.
func (a *Anthropic) Summarise(ctx context.Context, req SummariseRequest) (SummariseResponse, error) {
	combined := strings.Join(req.Chunks, "\n\n")

	resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     a.model,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: strings.Join([]string{
				"You are a concise document summariser.",
				"Produce a short, clear summary of the document provided.",
				"Use British English spelling and punctuation throughout.",
				"Output only the summary text, nothing else.",
			}, " ")},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(combined)),
		},
	})
	if err != nil {
		return SummariseResponse{}, fmt.Errorf("anthropic summarise: %w", err)
	}

	if len(resp.Content) == 0 || resp.Content[0].Type != "text" {
		return SummariseResponse{}, fmt.Errorf("anthropic summarise: no text content in response")
	}

	return SummariseResponse{Summary: strings.TrimSpace(resp.Content[0].Text), Model: a.model}, nil
}

// Answer answers a question grounded in the provided document content.
// Recent Q&A history is included in the prompt so follow-up questions work correctly.
// If the answer is not present in the content, the model is instructed to say so.
func (a *Anthropic) Answer(ctx context.Context, req AnswerRequest) (AnswerResponse, error) {
	var sb strings.Builder
	sb.WriteString("Document:\n")
	sb.WriteString(req.Content)

	if len(req.History) > 0 {
		sb.WriteString("\n\nConversation so far:\n")
		for _, qa := range req.History {
			sb.WriteString("Q: ")
			sb.WriteString(qa.Question)
			sb.WriteString("\nA: ")
			sb.WriteString(qa.Answer)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nQuestion: ")
	sb.WriteString(req.Question)
	prompt := sb.String()

	resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     a.model,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: strings.Join([]string{
				"You are a precise question answerer.",
				"The document is the sole source of truth.",
				"Answer the question using only the document provided.",
				"Use British English spelling and punctuation throughout.",
				"The conversation history is context only — do not treat prior answers as authoritative.",
				"If the current question challenges a previous answer, re-check the document and correct the answer if it was wrong.",
				"If a previous answer was wrong, say so clearly and provide the correct answer from the document.",
				"If the answer is not present in the document, say so clearly.",
				"Output only the answer, nothing else.",
			}, " ")},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return AnswerResponse{}, fmt.Errorf("anthropic answer: %w", err)
	}

	if len(resp.Content) == 0 || resp.Content[0].Type != "text" {
		return AnswerResponse{}, fmt.Errorf("anthropic answer: no text content in response")
	}

	return AnswerResponse{Answer: strings.TrimSpace(resp.Content[0].Text), Model: a.model}, nil
}
