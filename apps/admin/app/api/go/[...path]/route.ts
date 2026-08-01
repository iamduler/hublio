import { type NextRequest } from "next/server";
import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

type RouteContext = { params: Promise<{ path: string[] }> };

async function handle(request: NextRequest, context: RouteContext) {
  try {
    const { path: segments } = await context.params;
    const path = `/${segments.join("/")}`;
    const method = request.method.toUpperCase();

    const params: Record<string, string | undefined> = {};
    request.nextUrl.searchParams.forEach((v, k) => {
      params[k] = v;
    });

    let body: unknown;
    if (method !== "GET" && method !== "HEAD" && method !== "DELETE") {
      const text = await request.text();
      body = text ? (JSON.parse(text) as unknown) : undefined;
    }

    const forwardHeaders: Record<string, string> = {};
    const idempotency = request.headers.get("Idempotency-Key");
    if (idempotency) forwardHeaders["Idempotency-Key"] = idempotency;
    const acceptLanguage = request.headers.get("Accept-Language");
    if (acceptLanguage) forwardHeaders["Accept-Language"] = acceptLanguage;

    const requireAuth = !path.startsWith("/auth/");

    return await proxyGoWithJWT(path, {
      method,
      body,
      params,
      headers: forwardHeaders,
      requireAuth,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}

export const GET = handle;
export const POST = handle;
export const PUT = handle;
export const PATCH = handle;
export const DELETE = handle;
