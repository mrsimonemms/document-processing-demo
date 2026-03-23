import { getDocumentState } from '$lib/server/temporal';
import { json } from '@sveltejs/kit';

import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ params }) => {
  const state = await getDocumentState(params.documentId);
  return json(state);
};
