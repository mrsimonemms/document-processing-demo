import { listDocumentWorkflows } from '$lib/server/temporal';

import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
  const documents = await listDocumentWorkflows();
  return { documents, title: 'Document sessions' };
};
