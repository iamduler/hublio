import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

/** GET /api/auth/me — JWT from httpOnly cookie; returns user + organization. */
export async function GET() {
  try {
    return await proxyGoWithJWT("/auth/me", {
      method: "GET",
      requireAuth: true,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
