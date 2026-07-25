import { type NextRequest } from "next/server";
import {
  bffErrorResponse,
  proxyToGo,
  resolveWorkspaceContext,
} from "@/lib/api/bff";

/** GET /api/events?execution_id&limit → Go GET /events (workspace-scoped). */
export async function GET(request: NextRequest) {
  try {
    const ctx = await resolveWorkspaceContext();
    const { searchParams } = new URL(request.url);
    return await proxyToGo(ctx, "/events", {
      params: {
        execution_id: searchParams.get("execution_id") ?? undefined,
        limit: searchParams.get("limit") ?? undefined,
      },
    });
  } catch (err) {
    return bffErrorResponse(err);
  }
}
