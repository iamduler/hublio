import { type NextRequest } from "next/server";
import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

type Ctx = { params: Promise<{ intentId: string }> };

/** GET /api/intents/[intentId] → Go (JWT + X-Workspace-ID). */
export async function GET(_request: NextRequest, context: Ctx) {
  try {
    const { intentId } = await context.params;
    return await proxyGoWithJWT(`/intents/${intentId}`, {
      method: "GET",
      requireWorkspace: true,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
