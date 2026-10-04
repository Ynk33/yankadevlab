import { CAT_KEYS, TAB_KEYS, type CatKey, type TabKey } from "@/lib/catalog"
import { isObj } from "@/lib/utils"

export interface MineEntry {
  persons: number
  addedAt: number
}

export interface Extra {
  name: string
  rayon: string
  done?: boolean
}

export interface Prefs {
  persons: number
  exclusions: string
  envies: string
}

/**
 * Team state, stored as is by the server (PATCH /api/state with a JSON merge patch).
 * Keys must not be renamed: existing teams' data depends on them. Timestamps are epoch ms.
 */
export interface AppState {
  tab: TabKey
  cat: CatKey
  /** Deck key (see deckKey) → recipe ids. */
  decks: Record<string, string[]>
  mine: Record<string, MineEntry>
  done: Record<string, number>
  favs: Record<string, number>
  dislikes: Record<string, number>
  /** Shopping item key (see totals) → quantity already bought. */
  bought: Record<string, number>
  extras: Record<string, Extra>
  prefs: Prefs
}

export const DEFAULT_STATE: AppState = {
  tab: "all",
  cat: "all",
  decks: {},
  mine: {},
  done: {},
  favs: {},
  dislikes: {},
  bought: {},
  extras: {},
  prefs: { persons: 2, exclusions: "", envies: "" },
}

/** State as returned by GET /api/state: unvalidated, possibly in a legacy shape. */
export interface StoredState extends Partial<Omit<AppState, "extras">> {
  extras?: Record<string, Extra> | (Extra & { id: string })[]
  deck?: string[]
  lang?: string
}

const MAPS = ["decks", "mine", "done", "favs", "dislikes", "bought"] as const

function isTab(t: unknown): t is TabKey {
  return TAB_KEYS.some((k) => k === t)
}

function isCat(c: unknown): c is CatKey {
  return CAT_KEYS.some((k) => k === c)
}

function extrasOf(x: StoredState["extras"]): Record<string, Extra> {
  if (Array.isArray(x)) {
    return Object.fromEntries(
      x.filter((e) => e && e.id).map(({ id, ...e }) => [id, e]),
    )
  }
  return isObj(x) ? { ...x } : {}
}

/**
 * Turns the state returned by GET /api/state into a valid AppState: fills defaults, repairs malformed maps and
 * migrates legacy shapes (`extras` array, single `deck`). Unknown keys are kept; `lang` is dropped since it now
 * lives in /api/me.
 */
export function normalizeState(st: StoredState | undefined): AppState {
  if (!st) return structuredClone(DEFAULT_STATE)
  const rest = { ...st }
  delete rest.lang
  delete rest.deck
  const base = structuredClone(DEFAULT_STATE)
  const out: AppState = {
    ...base,
    ...rest,
    tab: isTab(st.tab) ? st.tab : "all",
    cat: isCat(st.cat) ? st.cat : "all",
    prefs: { ...base.prefs, ...(isObj(st.prefs) ? st.prefs : {}) },
    extras: extrasOf(st.extras),
  }
  for (const k of MAPS) {
    if (!isObj(st[k])) out[k] = {}
  }
  if (Array.isArray(st.deck) && !out.decks.all) {
    out.decks = { ...out.decks, all: st.deck }
  }
  return out
}
