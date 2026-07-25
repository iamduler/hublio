// Re-export of the generated OpenAPI SDK types (`@hublio/sdk`).
//
// The SDK types are generated from `api/openapi/openapi.yaml` and are the
// canonical shape of the HTTP API. Feature modules keep their hand-written
// fetch clients (thin wrappers over `lib/api/client`), but should reference
// these types where a DTO maps 1:1 to a spec schema, so drift is caught by
// the compiler. Regenerate with: `pnpm --filter @hublio/sdk generate`.
export type { paths, components, operations, Schemas } from "@hublio/sdk";
