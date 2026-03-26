package activities

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/providers"
)

// AnswerQuestionActivity answers a question grounded in the supplied document content.
//
// Provider selection follows the same failover pattern as SummariseDocumentActivity:
// the primary provider is tried first, and the fallback is used if it fails.
// Under ScenarioProviderFailover, the primary provider is replaced with a faulty
// wrapper that always errors, forcing the fallback to be used.
// Provider selection is controlled by input.ProviderOverride before scenario injection.
func (a *Activities) AnswerQuestionActivity(ctx context.Context, input models.AnswerInput) (models.AnswerResult, error) {
	// Select providers then copy so scenario injection does not mutate the base slice.
	base := selectQuestioners(a.questioners, input.ProviderOverride)
	chain := make([]providers.QuestionAnswerer, len(base))
	copy(chain, base)

	if input.Scenario == models.ScenarioProviderFailover && len(chain) > 0 {
		chain[0] = providers.NewFaultyQuestionProvider(chain[0], "simulated primary provider failure")
	}

	if input.Scenario == models.ScenarioProviderDown && len(chain) > 0 {
		chain[0] = providers.NewDownQuestionProvider(chain[0], fmt.Sprintf("simulated %s provider down", chain[0].Name()))
	}

	if input.Scenario == models.ScenarioProviderRateLimit && len(chain) > 0 {
		chain[0] = providers.NewRateLimitQuestionProvider(chain[0], fmt.Sprintf("simulated %s provider rate limit", chain[0].Name()))
	}

	result, err := providers.AnswerWithFailover(ctx, chain, providers.AnswerRequest{
		Content:  input.Content,
		Question: input.Question,
		History:  input.History,
	})
	if err != nil {
		if providers.IsProviderError(err) {
			return models.AnswerResult{}, temporal.NewNonRetryableApplicationError(err.Error(), "ProviderError", err)
		}
		return models.AnswerResult{}, err
	}

	logger := activity.GetLogger(ctx)
	logger.Info("question answered",
		"provider", result.Provider,
		"fallbackOccurred", result.FallbackOccurred,
	)

	return result, nil
}
