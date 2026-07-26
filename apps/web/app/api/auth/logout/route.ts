import { type NextRequest, NextResponse } from "next/server";
import {
  clearAuthCookiesOnStore,
  goBaseUrl,
  proxyErrorResponse,
  readSessionTokens,
} from "@/lib/api/proxy-go";

/** POST /api/auth/logout — revoke refresh token server-side and clear httpOnly cookies. */
export async function POST(_request: NextRequest) {
  try {
    const session = await readSessionTokens();
    if (session?.refreshToken) {
      await fetch(`${goBaseUrl()}/auth/logout`, {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ refresh_token: session.refreshToken }),
        cache: "no-store",
      }).catch(() => undefined);
    }
    await clearAuthCookiesOnStore();
    return NextResponse.json({
      status: "success",
      message: "logged out",
      data: null,
    });
  } catch (err) {
    await clearAuthCookiesOnStore().catch(() => undefined);
    return proxyErrorResponse(err);
  }
}
