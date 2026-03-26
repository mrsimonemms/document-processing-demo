package providers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

const maxProviderAttempts = 3

// FailureKind identifies the small set of provider failure classes used by the
// demo's explicit failover logic.
type FailureKind string

const (
	FailureDown      FailureKind = "down"
	FailureRateLimit FailureKind = "rate_limit"
	FailureFatal     FailureKind = "fatal"
)

// ProviderError carries a failure kind so failover can decide whether to retry
// the current provider or move on immediately.
type ProviderError struct {
	Kind    FailureKind
	Message string
}

func (e *ProviderError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func NewProviderError(kind FailureKind, message string) error {
	return &ProviderError{Kind: kind, Message: message}
}

func providerErrorKind(err error) FailureKind {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr.Kind
	}
	return FailureFatal
}

func IsProviderError(err error) bool {
	var providerErr *ProviderError
	return errors.As(err, &providerErr)
}

// SummariseWithFailover tries each provider in chain order.
// It returns the first successful result with metadata indicating which
// provider was used and whether a fallback occurred.
// If all providers fail, it returns a combined error.
func SummariseWithFailover(ctx context.Context, chain []Summariser, req SummariseRequest) (models.SummariseResult, error) {
	var errs []string

	for i, p := range chain {
		for attempt := 1; attempt <= maxProviderAttempts; attempt++ {
			resp, err := p.Summarise(ctx, req)
			if err == nil {
				return models.SummariseResult{
					Summary:          resp.Summary,
					Provider:         p.Name(),
					Model:            resp.Model,
					FallbackOccurred: i > 0,
				}, nil
			}

			errs = append(errs, fmt.Sprintf("%s attempt %d/%d: %v", p.Name(), attempt, maxProviderAttempts, err))
			if providerErrorKind(err) != FailureDown {
				break
			}
		}
	}

	return models.SummariseResult{}, NewProviderError(FailureFatal, fmt.Sprintf("all providers failed: %s", strings.Join(errs, "; ")))
}

// AnswerWithFailover tries each provider in chain order for Q&A.
// It returns the first successful result with metadata indicating which
// provider was used and whether a fallback occurred.
// If all providers fail, it returns a combined error.
func AnswerWithFailover(ctx context.Context, chain []QuestionAnswerer, req AnswerRequest) (models.AnswerResult, error) {
	var errs []string

	for i, p := range chain {
		for attempt := 1; attempt <= maxProviderAttempts; attempt++ {
			resp, err := p.Answer(ctx, req)
			if err == nil {
				return models.AnswerResult{
					Answer:           resp.Answer,
					Provider:         p.Name(),
					Model:            resp.Model,
					FallbackOccurred: i > 0,
				}, nil
			}

			errs = append(errs, fmt.Sprintf("%s attempt %d/%d: %v", p.Name(), attempt, maxProviderAttempts, err))
			if providerErrorKind(err) != FailureDown {
				break
			}
		}
	}

	return models.AnswerResult{}, NewProviderError(FailureFatal, fmt.Sprintf("all providers failed: %s", strings.Join(errs, "; ")))
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
	return SummariseResponse{}, NewProviderError(FailureFatal, f.message)
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
	return AnswerResponse{}, NewProviderError(FailureFatal, f.message)
}

// DownProvider wraps any Summariser and always returns a retryable provider-down error.
type DownProvider struct {
	inner   Summariser
	message string
}

func NewDownProvider(inner Summariser, message string) Summariser {
	return &DownProvider{inner: inner, message: message}
}

func (d *DownProvider) Name() models.ProviderName { return d.inner.Name() }

func (d *DownProvider) Summarise(_ context.Context, _ SummariseRequest) (SummariseResponse, error) {
	return SummariseResponse{}, NewProviderError(FailureDown, d.message)
}

// DownQuestionProvider wraps any QuestionAnswerer and always returns a retryable provider-down error.
type DownQuestionProvider struct {
	inner   QuestionAnswerer
	message string
}

func NewDownQuestionProvider(inner QuestionAnswerer, message string) QuestionAnswerer {
	return &DownQuestionProvider{inner: inner, message: message}
}

func (d *DownQuestionProvider) Name() models.ProviderName { return d.inner.Name() }

func (d *DownQuestionProvider) Answer(_ context.Context, _ AnswerRequest) (AnswerResponse, error) {
	return AnswerResponse{}, NewProviderError(FailureDown, d.message)
}

// RateLimitProvider wraps any Summariser and always returns a non-retryable rate-limit error.
type RateLimitProvider struct {
	inner   Summariser
	message string
}

func NewRateLimitProvider(inner Summariser, message string) Summariser {
	return &RateLimitProvider{inner: inner, message: message}
}

func (r *RateLimitProvider) Name() models.ProviderName { return r.inner.Name() }

func (r *RateLimitProvider) Summarise(_ context.Context, _ SummariseRequest) (SummariseResponse, error) {
	return SummariseResponse{}, NewProviderError(FailureRateLimit, r.message)
}

// RateLimitQuestionProvider wraps any QuestionAnswerer and always returns a non-retryable rate-limit error.
type RateLimitQuestionProvider struct {
	inner   QuestionAnswerer
	message string
}

func NewRateLimitQuestionProvider(inner QuestionAnswerer, message string) QuestionAnswerer {
	return &RateLimitQuestionProvider{inner: inner, message: message}
}

func (r *RateLimitQuestionProvider) Name() models.ProviderName { return r.inner.Name() }

func (r *RateLimitQuestionProvider) Answer(_ context.Context, _ AnswerRequest) (AnswerResponse, error) {
	return AnswerResponse{}, NewProviderError(FailureRateLimit, r.message)
}
