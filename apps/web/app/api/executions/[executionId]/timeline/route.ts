import { type NextRequest } from "next/server";
import {
  bffErrorResponse,
  proxyToGo,
  resolveWorkspaceContext,
} from "@/lib/api/bff";

/** GET /api/executions/:id/timeline → Go GET /executions/:id/timeline */
export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ executionId: string }> },
) {
  try {
    const { executionId } = await params;
    const ctx = await resolveWorkspaceContext();
    return await proxyToGo(ctx, `/executions/${executionId}/timeline`);
  } catch (err) {
    return bffErrorResponse(err);
  }
}
