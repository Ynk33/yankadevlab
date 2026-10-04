import { isObj } from "@/lib/utils"

/**
 * JSON merge patch (RFC 7386) turning `a` into `b`: removed keys become null, unchanged keys are omitted.
 * Mirrors merge.go on the server.
 * @returns the patch, or undefined when nothing changed
 */
export function mergeDiff(a: unknown, b: unknown): unknown {
  if (!isObj(a) || !isObj(b)) {
    return JSON.stringify(a) === JSON.stringify(b) ? undefined : b
  }
  const patch: Record<string, unknown> = {}
  let n = 0
  for (const k of Object.keys(a)) {
    if (!(k in b)) {
      patch[k] = null
      n++
    }
  }
  for (const k of Object.keys(b)) {
    const d = mergeDiff(a[k], b[k])
    if (d !== undefined) {
      patch[k] = d
      n++
    }
  }
  return n ? patch : undefined
}
