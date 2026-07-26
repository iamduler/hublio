import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

/** GET /api/auth/mfa/status — JWT from httpOnly cookie. */
export async function GET() {
  try {
    return await proxyGoWithJWT("/auth/mfa/status", {
      method: "GET",
      requireAuth: true,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
