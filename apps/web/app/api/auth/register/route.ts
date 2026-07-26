import { type NextRequest, NextResponse } from "next/server";
import {
  goBaseUrl,
  proxyErrorResponse,
  ProxyError,
} from "@/lib/api/proxy-go";

/** POST /api/auth/register — creates org + owner; no tokens issued by Go. */
export async function POST(request: NextRequest) {
  try {
    const body = (await request.json()) as unknown;
    const res = await fetch(`${goBaseUrl()}/auth/register`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
      cache: "no-store",
    });
    const text = await res.text();
    if (!res.ok) {
      return new NextResponse(text, {
        status: res.status,
        headers: { "Content-Type": "application/json" },
      });
    }
    if (!text) {
      throw new ProxyError(502, "AUTH_FAILED", "Empty register response.");
    }
    return new NextResponse(text, {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
