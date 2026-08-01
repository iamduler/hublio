import { type NextRequest, NextResponse } from "next/server";
import {
  clearAuthCookiesOnStore,
  proxyErrorResponse,
  ProxyError,
  readSessionTokens,
  refreshSessionAccessToken,
} from "@/lib/api/proxy-go";

/**
 * POST /api/auth/refresh — rotate httpOnly session using hublio_refresh.
 * Browser never sends or receives token plaintext.
 */
export async function POST(_request: NextRequest) {
  try {
    const session = await readSessionTokens();
    if (!session?.refreshToken) {
      await clearAuthCookiesOnStore().catch(() => undefined);
      throw new ProxyError(401, "UNAUTHORIZED", "No refresh token.");
    }

    const access = await refreshSessionAccessToken(session.refreshToken);
    if (!access) {
      throw new ProxyError(401, "UNAUTHORIZED", "Session expired.");
    }

    return NextResponse.json({
      status: "success",
      message: "token refreshed",
      data: null,
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
