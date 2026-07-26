import { NextResponse } from "next/server";
import { readSessionTokens } from "@/lib/api/proxy-go";

/**
 * GET /api/auth/session — presence check for httpOnly JWT.
 * Does not return the token; browser only learns whether a session exists.
 */
export async function GET() {
  const session = await readSessionTokens();
  return NextResponse.json({
    authenticated: Boolean(session?.accessToken),
    workspace_id: session?.workspaceId ?? null,
  });
}
