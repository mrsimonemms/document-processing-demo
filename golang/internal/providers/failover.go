package providers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

// SummariseWithFailover tries each provider in chain order.
// It returns the first successful result with metadata indicating which
// provider was used and whether a fallback occurred.
// If all providers fail, it returns a combined error.
func SummariseWithFailover(ctx context.Context, chain []Summariser, req SummariseRequest) (models.SummariseResult, error) {
	var errs []string

	for i, p := range chain {
		resp, err := p.Summarise(ctx, req)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p.Name(), err))
			continue
		}

		return models.SummariseResult{
			Summary:          resp.Summary,
			Provider:         p.Name(),
			FallbackOccurred: i > 0,
		}, nil
	}

	return models.SummariseResult{}, fmt.Errorf("all providers failed: %s", strings.Join(errs, "; "))
}

// AnswerWithFailover tries each provider in chain order for Q&A.
// It returns the first successful result with metadata indicating which
// provider was used and whether a fallback occurred.
// If all providers fail, it returns a combined error.
func AnswerWithFailover(ctx context.Context, chain []QuestionAnswerer, req AnswerRequest) (models.AnswerResult, error) {
	var errs []string

	for i, p := range chain {
		resp, err := p.Answer(ctx, req)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", p.Name(), err))
			continue
		}

		return models.AnswerResult{
			Answer:           resp.Answer,
			Provider:         p.Name(),
			FallbackOccurred: i > 0,
		}, nil
	}

	return models.AnswerResult{}, fmt.Errorf("all providers failed: %s", strings.Join(errs, "; "))
}

// FaultyProvider wraps any Summariser and always returns an error.
// It is used for deterministic failure injection in demo scenarios.
// The wrapped provider's Name() is preserved so logs and results correctly
// identify which provider was injected with a failure.
type FaultyProvider struct {
	inner   Summariser
	message string
}

// NewFaultyProvider returns a Summariser that always fails with the given message.
func NewFaultyProvider(inner Summariser, message string) Summariser {
	return &FaultyProvider{inner: inner, message: message}
}

func (f *FaultyProvider) Name() models.ProviderName { return f.inner.Name() }

func (f *FaultyProvider) Summarise(_ context.Context, _ SummariseRequest) (SummariseResponse, error) {
	return SummariseResponse{}, errors.New(f.message)
}

// FaultyQuestionProvider wraps any QuestionAnswerer and always returns an error.
// It is used for deterministic failure injection in demo scenarios.
type FaultyQuestionProvider struct {
	inner   QuestionAnswerer
	message string
}

// NewFaultyQuestionProvider returns a QuestionAnswerer that always fails with the given message.
func NewFaultyQuestionProvider(inner QuestionAnswerer, message string) QuestionAnswerer {
	return &FaultyQuestionProvider{inner: inner, message: message}
}

func (f *FaultyQuestionProvider) Name() models.ProviderName { return f.inner.Name() }

func (f *FaultyQuestionProvider) Answer(_ context.Context, _ AnswerRequest) (AnswerResponse, error) {
	return AnswerResponse{}, errors.New(f.message)
}
