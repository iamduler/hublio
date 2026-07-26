import { type NextRequest } from "next/server";
import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

/** POST /api/intents → Go POST /intents (JWT + X-Workspace-ID). */
export async function POST(request: NextRequest) {
  try {
    const body = (await request.json()) as unknown;
    const idempotencyKey =
      request.headers.get("Idempotency-Key") ?? crypto.randomUUID();

    return await proxyGoWithJWT("/intents", {
      method: "POST",
      body,
      headers: { "Idempotency-Key": idempotencyKey },
      requireWorkspace: true,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
