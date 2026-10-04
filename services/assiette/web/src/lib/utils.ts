import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Lowercase ASCII slug: accents stripped, non-alphanumerics collapsed to "-", max 80 chars. */
export function slug(s: string | undefined): string {
  return String(s || "")
    .toLowerCase()
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 80)
}

/** Fisher-Yates shuffle into a new array. */
export function shuffle<T>(a: T[], rand: () => number = Math.random): T[] {
  const out = a.slice()
  for (let i = out.length - 1; i > 0; i--) {
    const j = Math.floor(rand() * (i + 1))
    ;[out[i], out[j]] = [out[j], out[i]]
  }
  return out
}

export function isObj(o: unknown): o is Record<string, unknown> {
  return !!o && typeof o === "object" && !Array.isArray(o)
}
