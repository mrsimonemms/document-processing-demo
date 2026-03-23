import type { Scenario } from '$lib/server/temporal';
import { askDocumentQuestion } from '$lib/server/temporal';
import { json } from '@sveltejs/kit';

import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ params, request }) => {
  const body = await request.json();
  const question: unknown = body.question;
  const scenario: Scenario = body.scenario ?? 'happy_path';

  if (typeof question !== 'string' || !question.trim()) {
    return json({ error: 'Question is required.' }, { status: 400 });
  }

  const answer = await askDocumentQuestion(
    params.documentId,
    question,
    scenario,
  );
  return json({ question, answer });
};
