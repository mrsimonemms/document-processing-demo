package main

import (
	"os"

	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	gh "github.com/mrsimonemms/golang-helpers"
	"github.com/mrsimonemms/golang-helpers/temporal"
	temporalhelper "github.com/mrsimonemms/golang-helpers/temporal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/mrsimonemms/document-processing-demo/golang/internal/activities"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/models"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/providers"
	"github.com/mrsimonemms/document-processing-demo/golang/internal/workflows"
)

func main() {
	if err := exec(); err != nil {
		os.Exit(gh.HandleFatalError(err))
	}
}

func exec() error {
	logLevel, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logLevel = "info"
	}
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return gh.FatalError{
			Cause: err,
			Msg:   "Invalid log level",
		}
	}
	zerolog.SetGlobalLevel(level)

	openaiKey, ok := os.LookupEnv("OPENAI_API_KEY")
	if !ok || openaiKey == "" {
		return gh.FatalError{Msg: "OPENAI_API_KEY environment variable is required"}
	}

	// OPENAI_MODEL is optional; NewOpenAI defaults to gpt-4o.
	openaiModel := os.Getenv("OPENAI_MODEL")
	openaiProvider := providers.NewOpenAI(openaiKey, openaiModel)

	anthropicKey, ok := os.LookupEnv("ANTHROPIC_API_KEY")
	if !ok || anthropicKey == "" {
		return gh.FatalError{Msg: "ANTHROPIC_API_KEY environment variable is required"}
	}
	anthropicModel := os.Getenv("ANTHROPIC_MODEL")
	anthropicProvider := providers.NewAnthropic(anthropicKey, anthropicModel)

	summarisers := providers.NewChain(openaiProvider, anthropicProvider)
	questioners := providers.NewQuestionChain(openaiProvider, anthropicProvider)

	acts := activities.NewActivities(summarisers, questioners)

	// NewConnectionWithEnvvars reads TEMPORAL_ADDRESS, TEMPORAL_NAMESPACE,
	// TEMPORAL_TLS_CERT, TEMPORAL_TLS_KEY and TEMPORAL_API_KEY from the
	// environment. For local development the defaults (localhost:7233, default
	// namespace) are used automatically.
	temporalClient, err := temporalhelper.NewConnectionWithEnvvars(
		temporal.WithZerolog(&log.Logger),
	)
	if err != nil {
		return gh.FatalError{
			Cause: err,
			Msg:   "Failed to connect to Temporal",
		}
	}
	defer temporalClient.Close()

	// Register the single document workflow and all activities on the shared task queue.
	w := worker.New(temporalClient, models.TaskQueue, worker.Options{})
	w.RegisterWorkflowWithOptions(workflows.DocumentWorkflow, workflow.RegisterOptions{Name: "document"})
	w.RegisterActivity(acts)

	if err := w.Start(); err != nil {
		return gh.FatalError{
			Cause: err,
			Msg:   "Failed to start Temporal worker",
		}
	}
	defer w.Stop()

	log.Info().
		Str("taskQueue", models.TaskQueue).
		Str("provider", string(models.ProviderOpenAI)).
		Str("model", openaiModel).
		Msg("Worker ready")

	// Block until the process receives a shutdown signal.
	// worker.Run() handles SIGINT/SIGTERM and returns when signalled.
	if err := w.Run(worker.InterruptCh()); err != nil {
		return gh.FatalError{
			Cause: err,
			Msg:   "Worker failed to start listening",
		}
	}

	return nil
}
