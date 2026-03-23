import { getDocumentState } from '$lib/server/temporal';

import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params }) => {
  const state = await getDocumentState(params.documentId);
  return { ...state, title: 'Document session', subtitle: params.documentId };
};
