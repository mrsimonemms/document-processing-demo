import type { ProviderOverride, Scenario } from '$lib/server/temporal';
import { askDocumentQuestion } from '$lib/server/temporal';
import { json } from '@sveltejs/kit';

import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ params, request }) => {
  const body = await request.json();
  const question: unknown = body.question;
  const scenario: Scenario = body.scenario ?? 'happy_path';
  const providerOverride: ProviderOverride = body.providerOverride ?? 'default';

  if (typeof question !== 'string' || !question.trim()) {
    return json({ error: 'Question is required.' }, { status: 400 });
  }

  try {
    const answer = await askDocumentQuestion(
      params.documentId,
      question,
      scenario,
      providerOverride,
    );
    return json({ question, answer });
  } catch (err) {
    const message =
      err instanceof Error ? err.message : 'Question request failed.';
    return json({ error: message }, { status: 500 });
  }
};
