package providers

import (
	"context"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

// OpenAI is a fake OpenAI provider with deterministic output.
// It is the default primary provider in the demo chain.
// Replace Summarise and Answer with real API calls when integrating the OpenAI SDK.
type OpenAI struct{}

func (o *OpenAI) Name() models.ProviderName { return models.ProviderOpenAI }

func (o *OpenAI) Summarise(_ context.Context, req SummariseRequest) (SummariseResponse, error) {
	return SummariseResponse{Summary: buildSummary("openai", req)}, nil
}

func (o *OpenAI) Answer(_ context.Context, req AnswerRequest) (AnswerResponse, error) {
	return AnswerResponse{Answer: buildAnswer("openai", req)}, nil
}
