import type { LayoutServerLoad } from './$types';

// Expose the Temporal connection details used by the server-side client so the
// root layout can render them in the footer. Values mirror the defaults in
// src/lib/server/temporal.ts.
export const load: LayoutServerLoad = () => {
  return {
    temporalAddress: process.env.TEMPORAL_ADDRESS ?? 'localhost:7233',
    temporalNamespace: process.env.TEMPORAL_NAMESPACE ?? 'default',
  };
};
