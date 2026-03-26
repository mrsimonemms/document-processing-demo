package providers

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

const defaultModel = "gpt-4o"

// OpenAI is a real OpenAI provider backed by the official Go client.
// Construct one via NewOpenAI and inject it into Activities.
// API key and model are read from environment variables during construction.
type OpenAI struct {
	client openai.Client
	model  string
}

// NewOpenAI constructs an OpenAI provider using the given API key and model name.
// If model is empty it defaults to gpt-4o.
func NewOpenAI(apiKey, model string) *OpenAI {
	if model == "" {
		model = defaultModel
	}

	return &OpenAI{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		model:  model,
	}
}

func (o *OpenAI) Name() models.ProviderName { return models.ProviderOpenAI }

// Summarise concatenates all chunks and asks OpenAI for a concise summary.
func (o *OpenAI) Summarise(ctx context.Context, req SummariseRequest) (SummariseResponse, error) {
	combined := strings.Join(req.Chunks, "\n\n")

	resp, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: o.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(strings.Join([]string{
				"You are a concise document summariser.",
				"Produce a short, clear summary of the document provided.",
				"Use British English spelling and punctuation throughout.",
				"Output only the summary text, nothing else.",
			}, " ")),
			openai.UserMessage(combined),
		},
	})
	if err != nil {
		return SummariseResponse{}, fmt.Errorf("openai summarise: %w", err)
	}

	if len(resp.Choices) == 0 {
		return SummariseResponse{}, fmt.Errorf("openai summarise: no choices in response")
	}

	return SummariseResponse{Summary: strings.TrimSpace(resp.Choices[0].Message.Content), Model: o.model}, nil
}

// Answer answers a question grounded in the provided document content.
// Recent Q&A history is included in the prompt so follow-up questions work correctly.
// If the answer is not present in the content, the model is instructed to say so.
func (o *OpenAI) Answer(ctx context.Context, req AnswerRequest) (AnswerResponse, error) {
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

	resp, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: o.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(strings.Join([]string{
				"You are a precise question answerer.",
				"The document is the sole source of truth.",
				"Answer the question using only the document provided.",
				"Use British English spelling and punctuation throughout.",
				"The conversation history is context only — do not treat prior answers as authoritative.",
				"If the current question challenges a previous answer, re-check the document and correct the answer if it was wrong.",
				"If a previous answer was wrong, say so clearly and provide the correct answer from the document.",
				"If the answer is not present in the document, say so clearly.",
				"Output only the answer, nothing else.",
			}, " ")),
			openai.UserMessage(prompt),
		},
	})
	if err != nil {
		return AnswerResponse{}, fmt.Errorf("openai answer: %w", err)
	}

	if len(resp.Choices) == 0 {
		return AnswerResponse{}, fmt.Errorf("openai answer: no choices in response")
	}

	return AnswerResponse{Answer: strings.TrimSpace(resp.Choices[0].Message.Content), Model: o.model}, nil
}
