import type { ProviderOverride, Scenario } from '$lib/server/temporal';
import { getOrCreateDocumentWorkflow } from '$lib/server/temporal';
import { json } from '@sveltejs/kit';

import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ params, request }) => {
  const body = await request.json();
  const content: unknown = body.content;
  const scenario: Scenario = body.scenario ?? 'happy_path';
  const providerOverride: ProviderOverride = body.providerOverride ?? 'default';

  if (typeof content !== 'string' || !content.trim()) {
    return json({ error: 'Document content is required.' }, { status: 400 });
  }

  await getOrCreateDocumentWorkflow(
    params.documentId,
    content,
    scenario,
    providerOverride,
  );
  return json({ ok: true });
};
