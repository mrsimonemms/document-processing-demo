import { endDocumentSession } from '$lib/server/temporal';
import { json } from '@sveltejs/kit';

import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ params }) => {
  await endDocumentSession(params.documentId);
  return json({ ok: true });
};
