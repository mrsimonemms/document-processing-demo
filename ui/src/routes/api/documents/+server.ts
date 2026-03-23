import { listDocumentWorkflows } from '$lib/server/temporal';
import { json } from '@sveltejs/kit';

import type { RequestHandler } from './$types';

export const GET: RequestHandler = async () => {
  const documents = await listDocumentWorkflows();
  return json(documents);
};
