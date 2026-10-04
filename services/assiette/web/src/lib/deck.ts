import {
  isLocal,
  matches,
  inCat,
  inTab,
  type CatKey,
  type Recipe,
  type TabKey,
} from "@/lib/catalog"
import type { AppState } from "@/lib/state"
import { shuffle, slug } from "@/lib/utils"

/** Number of suggestions in a deck. */
export const DECK = 12
/** A recipe marked as done isn't suggested again for 21 days (ms). */
export const DONE_COOLDOWN = 21 * 864e5

/** Decks are stored per tab and category, e.g. "quick" or "quick|fish". */
export const deckKey = (tab: TabKey, cat: CatKey) =>
  tab + (cat === "all" ? "" : "|" + cat)

/** True when the recipe matches one of the comma-separated exclusion terms. */
export function excluded(r: Recipe, exclusions: string): boolean {
  const terms = exclusions
    .split(",")
    .map((s) => slug(s.trim()))
    .filter(Boolean)
  return terms.some((t) => matches(r, t))
}

export function eligible(r: Recipe, st: AppState, now: number): boolean {
  if (st.dislikes[r.id] || excluded(r, st.prefs.exclusions)) return false
  const d = st.done[r.id]
  return !(d && now - d < DONE_COOLDOWN)
}

/** Picks `n` recipes, round-robin across proteins so a deck isn't all fish. */
export function pickSpread(
  pool: Recipe[],
  n: number,
  rand: () => number = Math.random,
): Recipe[] {
  const groups: Record<string, Recipe[]> = {}
  for (const r of shuffle(pool, rand)) (groups[r.protein] ??= []).push(r)
  const keys = shuffle(Object.keys(groups), rand)
  const out: Recipe[] = []
  while (out.length < n && keys.some((k) => groups[k].length)) {
    for (const k of keys) {
      const r =
        groups[k].length && out.length < n ? groups[k].shift() : undefined
      if (r) out.push(r)
    }
  }
  return out
}

/**
 * Draws a new deck for the current tab and category: up to DECK - 2 local, in-season recipes, topped up with
 * the others at random positions.
 * @returns recipe ids
 */
export function newDeck(
  recipes: Recipe[],
  st: AppState,
  month: number,
  now: number,
  rand: () => number = Math.random,
): string[] {
  const pool = recipes.filter(
    (r) =>
      !st.mine[r.id] &&
      eligible(r, st, now) &&
      inTab(r, st.tab) &&
      inCat(r, st.cat),
  )
  const local = pickSpread(
    pool.filter((r) => isLocal(r, month)),
    DECK - 2,
    rand,
  )
  const rest = pickSpread(
    pool.filter((r) => !isLocal(r, month)),
    DECK - local.length,
    rand,
  )
  const out = local.slice()
  for (const r of rest) {
    out.splice(Math.floor(rand() * (out.length + 1)), 0, r)
  }
  return out.map((r) => r.id)
}

/** Recipes shown for the current tab and category, minus disliked and excluded ones. */
export const visible = (r: Recipe, st: AppState) =>
  !st.dislikes[r.id] &&
  !excluded(r, st.prefs.exclusions) &&
  inTab(r, st.tab) &&
  inCat(r, st.cat)
