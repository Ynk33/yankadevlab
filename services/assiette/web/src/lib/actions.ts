import { CATALOG, RAYONS } from "@/lib/catalog"
import { recipeText } from "@/lib/format"
import { I18N } from "@/lib/i18n"
import { clampBought, totals } from "@/lib/shopping"
import type { AppState, MineEntry, Prefs } from "@/lib/state"
import { closeSheet, getSnapshot, toast, update } from "@/lib/store"
import type { CatKey, TabKey } from "@/lib/catalog"

const t = () => I18N[getSnapshot().lang]
const st = () => getSnapshot().st

function without<T>(map: Record<string, T>, id: string): Record<string, T> {
  const out = { ...map }
  delete out[id]
  return out
}

/** Removes a recipe from "mine" and drops what it no longer needs from the bought quantities. */
function dropMine(s: AppState, id: string): AppState {
  const mine = without(s.mine, id)
  return { ...s, mine, bought: clampBought(s.bought, totals(mine, CATALOG)) }
}

function restoreMine(id: string, prev: MineEntry | undefined) {
  if (prev) update((s) => ({ ...s, mine: { ...s.mine, [id]: prev } }))
}

export function setTab(tab: TabKey) {
  update((s) => ({ ...s, tab }))
}

export function setCat(cat: CatKey) {
  update((s) => ({ ...s, cat }))
}

export function addMine(id: string, persons?: number) {
  update((s) => ({
    ...s,
    mine: {
      ...s.mine,
      [id]: { persons: persons || s.prefs.persons || 2, addedAt: Date.now() },
    },
  }))
  toast(t().tAddedMine)
}

export function removeMine(id: string) {
  const prev = st().mine[id]
  update((s) => dropMine(s, id))
  toast(t().tRemoved, () => restoreMine(id, prev))
}

export function markDone(id: string) {
  const prev = st().mine[id]
  update((s) => ({ ...dropMine(s, id), done: { ...s.done, [id]: Date.now() } }))
  if (getSnapshot().sheet === id) closeSheet()
  toast(t().tDone, () => {
    restoreMine(id, prev)
    update((s) => ({ ...s, done: without(s.done, id) }))
  })
}

export function setPersons(id: string, delta: number) {
  update((s) => {
    const m = s.mine[id]
    if (!m) return s
    const persons = Math.min(12, Math.max(1, m.persons + delta))
    return { ...s, mine: { ...s.mine, [id]: { ...m, persons } } }
  })
}

export function toggleFav(id: string) {
  const on = !st().favs[id]
  update((s) => ({
    ...s,
    favs: on ? { ...s.favs, [id]: Date.now() } : without(s.favs, id),
  }))
  toast(on ? t().tFavOn : t().tFavOff)
}

export function dislike(id: string) {
  const r = CATALOG[id]
  const { lang } = getSnapshot()
  if (!r || !confirm(t().confirmNever(recipeText(r, lang).title))) return
  update((s) => ({
    ...dropMine(s, id),
    dislikes: { ...s.dislikes, [id]: Date.now() },
    favs: without(s.favs, id),
  }))
  closeSheet()
  toast(t().tNever)
}

export function undislike(id: string) {
  update((s) => ({ ...s, dislikes: without(s.dislikes, id) }))
  toast(t().tReallow)
}

/** Marks a shopping item (`slug|unit`, or `x:<id>` for extras) as bought or not. */
export function setBought(key: string, bought: boolean) {
  update((s) => {
    if (key.startsWith("x:")) {
      const id = key.slice(2)
      const x = s.extras[id]
      if (!x) return s
      return { ...s, extras: { ...s.extras, [id]: { ...x, done: bought } } }
    }
    const tot = totals(s.mine, CATALOG).get(key)
    if (!tot) return s
    return {
      ...s,
      bought: bought ? { ...s.bought, [key]: tot.qty } : without(s.bought, key),
    }
  })
}

export function addExtra(name: string, rayon: string) {
  const id = Date.now().toString(36) + Math.random().toString(36).slice(2, 6)
  update((s) => ({
    ...s,
    extras: {
      ...s.extras,
      [id]: {
        name,
        rayon: RAYONS.includes(rayon) ? rayon : "Épicerie",
        done: false,
      },
    },
  }))
}

export function removeExtra(id: string) {
  const prev = st().extras[id]
  if (!prev) return
  update((s) => ({ ...s, extras: without(s.extras, id) }))
  toast(t().tExtraRm, () =>
    update((s) => ({ ...s, extras: { ...s.extras, [id]: prev } })),
  )
}

export function savePrefs(prefs: Prefs) {
  update((s) => ({ ...s, prefs }), { now: true })
  toast(t().settingsSaved)
}

export async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    toast(t().copied)
  } catch {
    toast(t().copyFail)
  }
}
