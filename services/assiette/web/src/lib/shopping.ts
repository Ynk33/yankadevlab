import { RAYONS, rayonOf, type Recipe } from "@/lib/catalog"
import type { Dict, Lang } from "@/lib/i18n"
import type { AppState } from "@/lib/state"
import { slug } from "@/lib/utils"

export interface Total {
  /** `slug(name)|unit`, also the key of AppState.bought. */
  key: string
  name: string
  unit: string
  qty: number
  rayon: string
}

export interface ShoppingItem {
  key: string
  name: string
  rayon: string
  unit?: string
  qty?: number
  /** Quantity still to buy. */
  rem?: number
  partial?: boolean
  extra?: boolean
}

export interface ShoppingList {
  groups: { rayon: string; items: ShoppingItem[] }[]
  done: ShoppingItem[]
  count: number
}

/** Factor to apply to a recipe's quantities for `persons` servings. */
export const scale = (r: Recipe, persons: number) => persons / (r.base || 2)

/** Sums the ingredients of every recipe in `mine`, scaled to its servings, by name and unit. */
export function totals(
  mine: AppState["mine"],
  catalog: Record<string, Recipe>,
): Map<string, Total> {
  const map = new Map<string, Total>()
  for (const [id, m] of Object.entries(mine)) {
    const r = catalog[id]
    if (!r) continue
    const k = scale(r, m.persons)
    for (const i of r.ingredients) {
      const key = slug(i.name) + "|" + i.unit
      const cur = map.get(key) ?? {
        key,
        name: i.name,
        unit: i.unit,
        qty: 0,
        rayon: rayonOf(i),
      }
      cur.qty += (Number(i.qty) || 0) * k
      map.set(key, cur)
    }
  }
  return map
}

/** Drops bought entries that are no longer on the list and caps the others at the needed quantity. */
export function clampBought(
  bought: AppState["bought"],
  t: Map<string, Total>,
): AppState["bought"] {
  const out: AppState["bought"] = {}
  for (const [k, v] of Object.entries(bought)) {
    const tot = t.get(k)
    if (tot) out[k] = Math.min(v, tot.qty)
  }
  return out
}

/**
 * Builds the shopping list: items left to buy grouped by aisle (in RAYONS order), and items already in the
 * basket. Both are sorted by their displayed name.
 */
export function shoppingGroups(
  st: AppState,
  catalog: Record<string, Recipe>,
  ingName: (name: string) => string,
  lang: Lang,
): ShoppingList {
  const t = totals(st.mine, catalog)
  const todo: Record<string, ShoppingItem[]> = {}
  const done: ShoppingItem[] = []
  const cmp = (a: ShoppingItem, b: ShoppingItem) =>
    ingName(a.name).localeCompare(ingName(b.name), lang)
  for (const it of t.values()) {
    const b = st.bought[it.key] || 0
    const rem = it.qty - b
    if (rem > it.qty * 0.001 + 1e-9) {
      ;(todo[it.rayon] ??= []).push({ ...it, rem, partial: b > 0 })
    } else {
      done.push(it)
    }
  }
  for (const [id, x] of Object.entries(st.extras)) {
    const it: ShoppingItem = {
      key: "x:" + id,
      name: x.name,
      rayon: x.rayon,
      extra: true,
    }
    if (x.done) done.push(it)
    else (todo[x.rayon] ??= []).push(it)
  }
  return {
    groups: RAYONS.filter((r) => todo[r]).map((r) => ({
      rayon: r,
      items: todo[r].sort(cmp),
    })),
    done: done.sort(cmp),
    count: t.size + Object.keys(st.extras).length,
  }
}

/** Formats a quantity for display: grams/ml rounded, kg/L above 1000, other units to the nearest half. */
export function fmtQty(q: number, unit: string, lang: Lang, t: Dict): string {
  q = Number(q) || 0
  if (q <= 0) return ""
  const dec = lang === "en" ? "." : ","
  const half = (x: number) =>
    (Math.round(x * 2) / 2).toString().replace(".", dec)
  const tenth = (x: number) =>
    (Math.round(x / 100) / 10).toString().replace(".", dec)
  if (unit === "g") {
    if (q >= 1000) return tenth(q) + " kg"
    return (
      (q >= 100 ? Math.round(q / 10) * 10 : Math.max(1, Math.round(q))) + " g"
    )
  }
  if (unit === "ml") {
    if (q >= 1000) return tenth(q) + " L"
    return Math.max(1, Math.round(q)) + " ml"
  }
  if (unit === "pièce") return half(Math.max(0.5, q))
  if (unit === "botte")
    return half(Math.max(0.5, q)) + " " + t.bunch[q > 1.25 ? 1 : 0]
  return half(Math.max(0.5, q)) + " " + (t.unit[unit] || unit)
}
