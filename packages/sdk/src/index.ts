// Public entrypoint for the generated Hublio API SDK.
//
// Types are generated from `api/openapi/openapi.yaml` (the source of truth).
// Regenerate with: `pnpm --filter @hublio/sdk generate`.
export type { paths, components, operations } from "./schema";
import type { components } from "./schema";

/** Convenience alias for all named schema (DTO) types. */
export type Schemas = components["schemas"];
