package providers

import (
	"context"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

// Anthropic is a fake Anthropic provider with deterministic output.
// It is the default fallback provider in the demo chain.
// Replace Summarise and Answer with real API calls when integrating the Anthropic SDK.
type Anthropic struct{}

func (a *Anthropic) Name() models.ProviderName { return models.ProviderAnthropic }

func (a *Anthropic) Summarise(_ context.Context, req SummariseRequest) (SummariseResponse, error) {
	return SummariseResponse{Summary: buildSummary("anthropic", req)}, nil
}

func (a *Anthropic) Answer(_ context.Context, req AnswerRequest) (AnswerResponse, error) {
	return AnswerResponse{Answer: buildAnswer("anthropic", req)}, nil
}
