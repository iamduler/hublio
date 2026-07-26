import { type NextRequest } from "next/server";
import { proxyErrorResponse, proxyGoWithJWT } from "@/lib/api/proxy-go";

/** POST /api/auth/mfa/setup — start TOTP enrollment (JWT). */
export async function POST(request: NextRequest) {
  try {
    const text = await request.text();
    const body = text ? (JSON.parse(text) as unknown) : {};
    return await proxyGoWithJWT("/auth/mfa/setup", {
      method: "POST",
      body,
      requireAuth: true,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
