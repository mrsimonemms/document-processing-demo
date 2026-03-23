package activities

import (
	"context"

	"go.temporal.io/sdk/activity"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/providers"
)

// AnswerQuestionActivity answers a question grounded in the supplied document content.
//
// Provider selection follows the same failover pattern as SummariseDocumentActivity:
// the primary provider is tried first, and the fallback is used if it fails.
// Under ScenarioProviderFailover, the primary provider is replaced with a faulty
// wrapper that always errors, forcing the fallback to be used.
func AnswerQuestionActivity(ctx context.Context, input models.AnswerInput) (models.AnswerResult, error) {
	chain := providers.DefaultQuestionChain()

	if input.Scenario == models.ScenarioProviderFailover && len(chain) > 0 {
		chain[0] = providers.NewFaultyQuestionProvider(chain[0], "simulated primary provider failure")
	}

	result, err := providers.AnswerWithFailover(ctx, chain, providers.AnswerRequest{
		Content:  input.Content,
		Question: input.Question,
	})
	if err != nil {
		return models.AnswerResult{}, err
	}

	logger := activity.GetLogger(ctx)
	logger.Info("question answered",
		"provider", result.Provider,
		"fallbackOccurred", result.FallbackOccurred,
	)

	return result, nil
}
