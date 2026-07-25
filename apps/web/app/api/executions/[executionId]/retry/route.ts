import { type NextRequest } from "next/server";
import {
  bffErrorResponse,
  proxyToGo,
  resolveWorkspaceContext,
} from "@/lib/api/bff";

/** POST /api/executions/:id/retry → Go POST /executions/:id/retry */
export async function POST(
  _request: NextRequest,
  { params }: { params: Promise<{ executionId: string }> },
) {
  try {
    const { executionId } = await params;
    const ctx = await resolveWorkspaceContext();
    return await proxyToGo(ctx, `/executions/${executionId}/retry`, {
      method: "POST",
      body: {},
    });
  } catch (err) {
    return bffErrorResponse(err);
  }
}
