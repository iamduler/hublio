/** Success envelope from Hublio Go API. */
export interface SuccessEnvelope<T> {
  status: string;
  message?: string;
  data: T;
  pagination?: {
    next_cursor?: string;
    has_next?: boolean;
    limit?: number;
    page?: number;
    total?: number;
    total_pages?: number;
    has_previous?: boolean;
  };
}

/** Error envelope from Hublio Go API. */
export interface ErrorEnvelope {
  error?: string;
  code?: string;
  detail?: string;
}

export function unwrapData<T>(res: unknown): T {
  const body = res as { data?: T } & T;
  if (body && typeof body === "object" && "data" in body && body.data !== undefined) {
    return body.data;
  }
  return body as T;
}
