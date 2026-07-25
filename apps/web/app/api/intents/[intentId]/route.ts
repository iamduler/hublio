import { type NextRequest } from "next/server";
import {
  bffErrorResponse,
  proxyToGo,
  resolveWorkspaceContext,
} from "@/lib/api/bff";

/** GET /api/intents/:id → Go GET /intents/:id */
export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ intentId: string }> },
) {
  try {
    const { intentId } = await params;
    const ctx = await resolveWorkspaceContext();
    return await proxyToGo(ctx, `/intents/${intentId}`);
  } catch (err) {
    return bffErrorResponse(err);
  }
}
