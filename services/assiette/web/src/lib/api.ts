export class HttpError extends Error {
  status: number

  constructor(path: string, status: number) {
    super(path + " " + status)
    this.status = status
  }
}

let onUnauthorized = () => {}

/** Registers the callback run when the session is missing or expired (401). */
export function setUnauthorizedHandler(fn: () => void) {
  onUnauthorized = fn
}

/**
 * Fetches a JSON endpoint of the service. Bodies are sent as JSON.
 * @throws {HttpError} on any non-2xx status; a 401 also triggers the unauthorized handler
 * @returns the parsed body, or undefined for 204 and null bodies
 */
export async function api<T>(
  path: string,
  opt: { method?: string; body?: unknown; keepalive?: boolean } = {},
): Promise<T | undefined> {
  const body = opt.body === undefined ? undefined : JSON.stringify(opt.body)
  const r = await fetch(path, {
    method: opt.method,
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body,
    keepalive: opt.keepalive && !!body && body.length < 65536,
  })
  if (r.status === 401 && path.startsWith("/api/")) onUnauthorized()
  if (!r.ok) throw new HttpError(path, r.status)
  if (r.status === 204) return undefined
  const data: T | null = await r.json()
  return data ?? undefined
}
