import { type NextRequest } from "next/server";
import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

type Ctx = { params: Promise<{ executionId: string }> };

export async function GET(_request: NextRequest, context: Ctx) {
  try {
    const { executionId } = await context.params;
    return await proxyGoWithJWT(`/executions/${executionId}/timeline`, {
      method: "GET",
      requireWorkspace: true,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
