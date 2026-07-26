const DEVICE_ID_KEY = "hublio_device_id";
const MFA_TOKEN_KEY = "hublio_mfa_token";

function randomId(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `dev_${Math.random().toString(36).slice(2)}_${Date.now()}`;
}

export function getOrCreateDeviceId(): string {
  if (typeof window === "undefined") return "";
  try {
    const existing = window.localStorage.getItem(DEVICE_ID_KEY);
    if (existing) return existing;
    const next = randomId();
    window.localStorage.setItem(DEVICE_ID_KEY, next);
    return next;
  } catch {
    return randomId();
  }
}

export function storeMFAToken(token: string): void {
  if (typeof window === "undefined") return;
  try {
    window.sessionStorage.setItem(MFA_TOKEN_KEY, token);
  } catch {
    // ignore
  }
}

export function readMFAToken(): string | null {
  if (typeof window === "undefined") return null;
  try {
    return window.sessionStorage.getItem(MFA_TOKEN_KEY);
  } catch {
    return null;
  }
}

export function clearMFAToken(): void {
  if (typeof window === "undefined") return;
  try {
    window.sessionStorage.removeItem(MFA_TOKEN_KEY);
  } catch {
    // ignore
  }
}
