import { type NextRequest } from "next/server";
import {
  bffErrorResponse,
  proxyToGo,
  resolveWorkspaceContext,
} from "@/lib/api/bff";

/** POST /api/intents → Go POST /intents (X-API-KEY, requires Idempotency-Key). */
export async function POST(request: NextRequest) {
  try {
    const ctx = await resolveWorkspaceContext();
    const body = (await request.json()) as unknown;
    const idempotencyKey =
      request.headers.get("Idempotency-Key") ?? crypto.randomUUID();

    return await proxyToGo(ctx, "/intents", {
      method: "POST",
      body,
      headers: { "Idempotency-Key": idempotencyKey },
    });
  } catch (err) {
    return bffErrorResponse(err);
  }
}
