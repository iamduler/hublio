import { type NextRequest } from "next/server";
import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

/** GET /api/events → Go (JWT + X-Workspace-ID). */
export async function GET(request: NextRequest) {
  try {
    return await proxyGoWithJWT("/events", {
      method: "GET",
      params: {
        execution_id: request.nextUrl.searchParams.get("execution_id") ?? undefined,
        limit: request.nextUrl.searchParams.get("limit") ?? undefined,
      },
      requireWorkspace: true,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
