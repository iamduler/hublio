import { type NextRequest } from "next/server";
import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

type Ctx = { params: Promise<{ executionId: string }> };

export async function POST(_request: NextRequest, context: Ctx) {
  try {
    const { executionId } = await context.params;
    return await proxyGoWithJWT(`/executions/${executionId}/cancel`, {
      method: "POST",
      body: {},
      requireWorkspace: true,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
