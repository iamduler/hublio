import { NextResponse } from "next/server";
import {
  readSessionTokens,
  refreshSessionAccessToken,
} from "@/lib/api/proxy-go";

/**
 * GET /api/auth/session — presence check for httpOnly JWT.
 * If access expired but refresh is present, rotates once so soft-gate stays alive.
 */
export async function GET() {
  let session = await readSessionTokens();
  if (!session?.accessToken && session?.refreshToken) {
    const access = await refreshSessionAccessToken(session.refreshToken);
    if (access) {
      session = await readSessionTokens();
    }
  }
  return NextResponse.json({
    authenticated: Boolean(session?.accessToken),
  });
}
