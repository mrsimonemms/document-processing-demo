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
	w.RegisterActivity(activities.ExtractDocumentTextActivity)
	w.RegisterActivity(activities.ChunkDocumentActivity)
	w.RegisterActivity(activities.SummariseDocumentActivity)
	w.RegisterActivity(activities.AnswerQuestionActivity)

	if err := w.Start(); err != nil {
		return gh.FatalError{
			Cause: err,
			Msg:   "Failed to start Temporal worker",
		}
	}
	defer w.Stop()

	log.Info().Str("taskQueue", models.TaskQueue).Msg("Worker ready")

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
