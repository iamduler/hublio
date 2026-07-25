/** Active workspace id, shared between client, proxy and BFF route handlers. */
export const WORKSPACE_COOKIE = "hublio_workspace";

const WORKSPACE_MAX_AGE = 30 * 24 * 60 * 60; // 30 days

export function readWorkspaceCookie(): string | undefined {
  if (typeof document === "undefined") return undefined;
  const match = document.cookie.match(
    new RegExp(`(?:^|; )${WORKSPACE_COOKIE}=([^;]*)`),
  );
  return match ? decodeURIComponent(match[1]!) : undefined;
}

export function writeWorkspaceCookie(workspaceId: string): void {
  if (typeof document === "undefined") return;
  document.cookie = `${WORKSPACE_COOKIE}=${encodeURIComponent(
    workspaceId,
  )}; path=/; max-age=${WORKSPACE_MAX_AGE}; SameSite=Lax`;
}

export function clearWorkspaceCookie(): void {
  if (typeof document === "undefined") return;
  document.cookie = `${WORKSPACE_COOKIE}=; path=/; max-age=0; SameSite=Lax`;
}
