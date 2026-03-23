package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/activities"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
)

func defaultActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
}

// DocumentWorkflow is the single long-lived workflow for a document session.
//
// It runs the full processing pipeline, then waits for questions via workflow
// updates and terminates when it receives the "end" signal.
//
// Lifecycle:
//  1. Extract text from the document input.
//  2. Chunk the extracted text.
//  3. Summarise via the provider chain (failure injection driven by input.Scenario).
//  4. Expose state via the "getState" query handler.
//  5. Accept "askQuestion" update requests, each answered by AnswerQuestionActivity.
//  6. Wait for the "end" signal, then exit cleanly.
//
// The workflow ID is the documentId, so one session maps to one workflow.
func DocumentWorkflow(ctx workflow.Context, input models.DocumentInput) error {
	// State visible to the UI via the "getState" query at any point.
	state := models.DocumentState{Phase: "processing", QA: []models.QA{}}

	if err := workflow.SetQueryHandler(ctx, "getState", func() (models.DocumentState, error) {
		return state, nil
	}); err != nil {
		return err
	}

	activityOpts := defaultActivityOptions()
	opts := workflow.WithActivityOptions(ctx, activityOpts)

	// Step 1: extract text
	var text string
	if err := workflow.ExecuteActivity(opts, activities.ExtractDocumentTextActivity, input).Get(opts, &text); err != nil {
		return err
	}

	// Step 2: chunk the extracted text
	var chunks []string
	if err := workflow.ExecuteActivity(opts, activities.ChunkDocumentActivity, text).Get(opts, &chunks); err != nil {
		return err
	}

	// Step 3: summarise via provider chain
	// Failure injection is handled inside the activity, driven by input.Scenario.
	summariseInput := models.SummariseInput{
		Chunks:   chunks,
		Scenario: input.Scenario,
	}

	var summariseResult models.SummariseResult
	if err := workflow.ExecuteActivity(opts, activities.SummariseDocumentActivity, summariseInput).Get(opts, &summariseResult); err != nil {
		return err
	}

	state = models.DocumentState{
		Phase:            "summarised",
		Summary:          summariseResult.Summary,
		Provider:         summariseResult.Provider,
		FallbackOccurred: summariseResult.FallbackOccurred,
	}

	// Step 4: register update handler for questions.
	// Each update runs AnswerQuestionActivity and returns the answer to the caller.
	// The document content is captured from the workflow input.
	if err := workflow.SetUpdateHandler(ctx, "askQuestion",
		func(ctx workflow.Context, req models.QuestionUpdate) (models.QuestionUpdateResult, error) {
			actCtx := workflow.WithActivityOptions(ctx, defaultActivityOptions())
			answerInput := models.AnswerInput{
				Content:  input.Content,
				Question: req.Question,
				Scenario: req.Scenario,
			}

			var result models.AnswerResult
			if err := workflow.ExecuteActivity(actCtx, activities.AnswerQuestionActivity, answerInput).Get(actCtx, &result); err != nil {
				return models.QuestionUpdateResult{}, err
			}

			state.QA = append(state.QA, models.QA{
				Question: req.Question,
				Answer:   result.Answer,
			})

			return models.QuestionUpdateResult{Answer: result.Answer}, nil
		},
	); err != nil {
		return err
	}

	// Step 5: wait for the "end" signal, then exit cleanly.
	workflow.GetSignalChannel(ctx, "end").Receive(ctx, nil)
	state.Phase = "ended"

	return nil
}
