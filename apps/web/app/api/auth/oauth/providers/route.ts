import { NextResponse } from "next/server";
import { goBaseUrl, proxyErrorResponse } from "@/lib/api/proxy-go";

/** GET /api/auth/oauth/providers */
export async function GET() {
  try {
    const res = await fetch(`${goBaseUrl()}/auth/oauth/providers`, {
      method: "GET",
      headers: { Accept: "application/json" },
      cache: "no-store",
    });
    const text = await res.text();
    return new NextResponse(text, {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
