import { api, HttpError } from "@/lib/api"
import { CATALOG, registerCustom, type CustomRecipeJSON } from "@/lib/catalog"
import type { Dict } from "@/lib/i18n"
import { I18N } from "@/lib/i18n"
import { getSnapshot, setGeneration, toast } from "@/lib/store"

/** Number of recipes a request may ask for. */
export const MAX_RECIPES = 5

const t = () => I18N[getSnapshot().lang]

export function isDraft(id: string): boolean {
  return !!CATALOG[id]?.draft
}

/** Message shown for a failed POST /api/generate, by HTTP status (0 when the request didn't reach the server). */
export function generationError(status: number, dict: Dict): string {
  if (status === 422) return dict.genDeclined
  if (status === 429) return dict.genTooMany
  if (status === 503) return dict.genOff
  return dict.genFail
}

/**
 * Asks the server for `n` recipes answering `prompt`: catalog recipes that fit, completed with generated drafts.
 * Drafts are registered in CATALOG so they can be opened, and kept only if added to a list or favourites.
 */
export async function generate(prompt: string, n: number) {
  setGeneration({ busy: true, matches: [], generated: [] })
  try {
    const res = await api<{ matches: string[]; generated: CustomRecipeJSON[] }>(
      "/api/generate",
      { method: "POST", body: { prompt, n } },
    )
    const generated = res?.generated ?? []
    for (const r of generated) {
      registerCustom(r)
      if (CATALOG[r.id]) CATALOG[r.id] = { ...CATALOG[r.id], draft: true }
    }
    const matches = (res?.matches ?? []).filter((id) => CATALOG[id])
    setGeneration({
      busy: false,
      matches,
      generated: generated.map((r) => r.id).filter((id) => CATALOG[id]),
      error: matches.length || generated.length ? undefined : t().genNone,
    })
  } catch (e) {
    console.error(e)
    const status = e instanceof HttpError ? e.status : 0
    setGeneration({
      busy: false,
      matches: [],
      generated: [],
      error: generationError(status, t()),
    })
  }
}

export function clearGeneration() {
  setGeneration({ busy: false, matches: [], generated: [] })
}

/**
 * Adds a generated draft to the global catalog.
 * @returns false (after a toast) when the server refused it
 */
export async function publishDraft(id: string): Promise<boolean> {
  try {
    await api(`/api/catalog/${encodeURIComponent(id)}`, { method: "POST" })
    if (CATALOG[id]) CATALOG[id] = { ...CATALOG[id], draft: false }
    return true
  } catch (e) {
    console.error(e)
    toast(t().genAddFail)
    return false
  }
}
