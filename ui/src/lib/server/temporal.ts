import { Client, Connection, WorkflowIdReusePolicy } from '@temporalio/client';
import fs from 'node:fs';

// DocumentWorkflowSummary is returned by listDocumentWorkflows.
export interface DocumentWorkflowSummary {
  id: string;
  status: string;
  startTime: string; // ISO 8601
  closeTime: string | null;
}

const TASK_QUEUE = 'document-processing';

export type Scenario =
  | 'happy_path'
  | 'fail_once_summarise'
  | 'primary_provider_failure';

export type ProviderOverride = 'default' | 'openai' | 'anthropic';

// SessionState represents the current state of a document session as seen by
// the UI. The phase field drives which section of the page is shown.
export interface SessionState {
  documentId: string;
  // pending     – no workflow running for this document yet
  // processing  – workflow is running, summary not yet ready
  // summarised  – summary complete, Q&A available
  // ended       – end signal received, session closed
  // failed      – workflow failed or timed out
  phase: 'pending' | 'processing' | 'summarised' | 'ended' | 'failed';
  summary?: string;
  provider?: string;
  model?: string;
  fallbackOccurred?: boolean;
  qa: Array<{
    question: string;
    answer: string;
    provider?: string;
    model?: string;
  }>;
  providerOverride?: ProviderOverride;
}

// DocumentState mirrors the Go DocumentState struct returned by the getState query.
interface DocumentState {
  phase: 'processing' | 'summarised' | 'ended';
  summary?: string;
  provider?: string;
  model?: string;
  fallbackOccurred?: boolean;
  qa: Array<{
    question: string;
    answer: string;
    provider?: string;
    model?: string;
  }>;
  providerOverride?: ProviderOverride;
}

let clientInstance: Client | null = null;

async function getClient(): Promise<Client> {
  if (clientInstance) return clientInstance;

  const address = process.env.TEMPORAL_ADDRESS ?? 'localhost:7233';
  const namespace = process.env.TEMPORAL_NAMESPACE ?? 'default';

  let connection: Connection;

  if (process.env.TEMPORAL_API_KEY) {
    connection = await Connection.connect({
      address,
      tls: true,
      apiKey: process.env.TEMPORAL_API_KEY,
    });
  } else if (process.env.TEMPORAL_TLS_CERT && process.env.TEMPORAL_TLS_KEY) {
    connection = await Connection.connect({
      address,
      tls: {
        clientCertPair: {
          crt: fs.readFileSync(process.env.TEMPORAL_TLS_CERT),
          key: fs.readFileSync(process.env.TEMPORAL_TLS_KEY),
        },
      },
    });
  } else {
    connection = await Connection.connect({ address });
  }

  clientInstance = new Client({ connection, namespace });
  return clientInstance;
}

// getOrCreateDocumentWorkflow starts the document workflow for the given
// documentId if it is not already running. If a workflow for that ID is
// already running, this is a no-op (upsert behaviour).
//
// The documentId is used directly as the workflow ID so one document maps
// to exactly one workflow instance.
export async function getOrCreateDocumentWorkflow(
  documentId: string,
  content: string,
  scenario: Scenario,
  providerOverride: ProviderOverride,
): Promise<void> {
  const client = await getClient();
  const handle = client.workflow.getHandle(documentId);

  try {
    const description = await handle.describe();
    if (description.status.name === 'RUNNING') {
      // Already running — nothing to do.
      return;
    }
  } catch {
    // Workflow does not exist — fall through to start.
  }

  await client.workflow.start('document', {
    taskQueue: TASK_QUEUE,
    workflowId: documentId,
    args: [{ documentId, content, scenario, providerOverride }],
    workflowIdReusePolicy: WorkflowIdReusePolicy.ALLOW_DUPLICATE,
  });
}

// getDocumentState returns the current phase and result of the session.
// Returns phase 'pending' if no workflow has been started for this document ID.
export async function getDocumentState(
  documentId: string,
): Promise<SessionState> {
  const client = await getClient();
  const handle = client.workflow.getHandle(documentId);

  try {
    const description = await handle.describe();

    if (description.status.name === 'RUNNING') {
      const state = await handle.query<DocumentState>('getState');
      return {
        documentId,
        phase: state.phase,
        summary: state.summary,
        provider: state.provider,
        model: state.model,
        fallbackOccurred: state.fallbackOccurred,
        qa: state.qa ?? [],
        providerOverride: state.providerOverride,
      };
    }

    if (description.status.name === 'COMPLETED') {
      // The only normal completion is via the end signal. Query the completed
      // workflow to retrieve the full session state including Q&A history.
      // Temporal workers can answer queries on completed workflows by replaying
      // the workflow history.
      try {
        const state = await handle.query<DocumentState>('getState');
        return {
          documentId,
          phase: 'ended',
          summary: state.summary,
          provider: state.provider,
          model: state.model,
          fallbackOccurred: state.fallbackOccurred,
          qa: state.qa ?? [],
          providerOverride: state.providerOverride,
        };
      } catch {
        return { documentId, phase: 'ended', qa: [] };
      }
    }

    if (
      description.status.name === 'FAILED' ||
      description.status.name === 'TIMED_OUT'
    ) {
      return { documentId, phase: 'failed', qa: [] };
    }

    // CANCELLED or any other terminal state: treat as pending so the user
    // can re-upload to the same session ID.
    return { documentId, phase: 'pending', qa: [] };
  } catch {
    // Workflow does not exist yet.
    return { documentId, phase: 'pending', qa: [] };
  }
}

// askDocumentQuestion sends an update to the document workflow and returns
// the answer. The update is synchronous: it blocks until AnswerQuestionActivity
// completes inside the workflow and the answer is returned to the caller.
export async function askDocumentQuestion(
  documentId: string,
  question: string,
  scenario: Scenario,
  providerOverride: ProviderOverride,
): Promise<string> {
  const client = await getClient();
  const handle = client.workflow.getHandle(documentId);

  const result = await handle.executeUpdate<
    { answer: string },
    [
      {
        question: string;
        scenario: Scenario;
        providerOverride: ProviderOverride;
      },
    ]
  >('askQuestion', {
    args: [{ question, scenario, providerOverride }],
  });

  return result.answer;
}

// listDocumentWorkflows returns a summary of all document workflows, both
// running and completed, ordered by start time descending (most recent first).
// Temporal is the sole source of truth; no separate database is used.
export async function listDocumentWorkflows(): Promise<
  DocumentWorkflowSummary[]
> {
  const client = await getClient();
  const results: DocumentWorkflowSummary[] = [];

  for await (const wf of client.workflow.list({
    query: 'WorkflowType = "document"',
  })) {
    results.push({
      id: wf.workflowId,
      status: wf.status.name,
      startTime: wf.startTime.toISOString(),
      closeTime: wf.closeTime ? wf.closeTime.toISOString() : null,
    });
  }

  results.sort(
    (a, b) => new Date(b.startTime).getTime() - new Date(a.startTime).getTime(),
  );

  return results;
}

// endDocumentSession sends the "end" signal to the document workflow, causing
// it to exit cleanly. No-op if the workflow is not running.
export async function endDocumentSession(documentId: string): Promise<void> {
  const client = await getClient();
  const handle = client.workflow.getHandle(documentId);

  try {
    const description = await handle.describe();
    if (description.status.name === 'RUNNING') {
      await handle.signal('end');
    }
  } catch {
    // Workflow does not exist — nothing to end.
  }
}
