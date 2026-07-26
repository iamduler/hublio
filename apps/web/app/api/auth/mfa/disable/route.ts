import { type NextRequest } from "next/server";
import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

/** POST /api/auth/mfa/disable — turn off MFA after password confirm (JWT). */
export async function POST(request: NextRequest) {
  try {
    const body = (await request.json()) as unknown;
    return await proxyGoWithJWT("/auth/mfa/disable", {
      method: "POST",
      body,
      requireAuth: true,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
