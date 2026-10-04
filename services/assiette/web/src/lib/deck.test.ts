import { describe, expect, it } from "vitest"
import { CATALOG, inCat, inTab, isLocal, type Recipe } from "@/lib/catalog"
import {
  DECK,
  DONE_COOLDOWN,
  deckKey,
  eligible,
  excluded,
  newDeck,
  pickSpread,
} from "@/lib/deck"
import { DEFAULT_STATE, type AppState } from "@/lib/state"

const recipes = Object.values(CATALOG)
const NOW = Date.UTC(2026, 9, 4)

function seeded(seed: number) {
  return () => {
    seed = (seed * 1664525 + 1013904223) % 2 ** 32
    return seed / 2 ** 32
  }
}

function state(patch: Partial<AppState> = {}): AppState {
  return { ...structuredClone(DEFAULT_STATE), ...patch }
}

const salmon = CATALOG["saumon-roti-lentilles-vertes-et-epinards"]

describe("deckKey", () => {
  it("omits the category when it is all", () => {
    expect(deckKey("quick", "all")).toBe("quick")
    expect(deckKey("quick", "fish")).toBe("quick|fish")
  })
})

describe("excluded", () => {
  it("matches French and English ingredient names, ignoring accents", () => {
    expect(excluded(salmon, "")).toBe(false)
    expect(excluded(salmon, "porc, Épinard")).toBe(true)
    expect(excluded(salmon, "salmon")).toBe(true)
    expect(excluded(salmon, "boeuf")).toBe(false)
  })
})

describe("eligible", () => {
  it("rejects disliked and recently done recipes", () => {
    expect(eligible(salmon, state(), NOW)).toBe(true)
    expect(eligible(salmon, state({ dislikes: { [salmon.id]: 1 } }), NOW)).toBe(
      false,
    )
    expect(
      eligible(salmon, state({ done: { [salmon.id]: NOW - 1 } }), NOW),
    ).toBe(false)
    expect(
      eligible(
        salmon,
        state({ done: { [salmon.id]: NOW - DONE_COOLDOWN } }),
        NOW,
      ),
    ).toBe(true)
  })
})

describe("pickSpread", () => {
  it("alternates proteins before repeating one", () => {
    const pool: Recipe[] = [
      ...recipes.filter((r) => r.protein === "poisson").slice(0, 5),
      ...recipes.filter((r) => r.protein === "porc").slice(0, 1),
    ]
    const picked = pickSpread(pool, 2, seeded(1))
    expect(new Set(picked.map((r) => r.protein))).toEqual(
      new Set(["poisson", "porc"]),
    )
  })

  it("returns at most the pool size", () => {
    expect(pickSpread(recipes.slice(0, 3), 10, seeded(2))).toHaveLength(3)
  })
})

describe("newDeck", () => {
  it("draws DECK eligible recipes for the tab and category", () => {
    const st = state({
      tab: "quick",
      cat: "veg",
      mine: { [recipes[0].id]: { persons: 2, addedAt: 1 } },
    })
    const ids = newDeck(recipes, st, 10, NOW, seeded(3))
    expect(ids).toHaveLength(DECK)
    expect(new Set(ids).size).toBe(DECK)
    for (const id of ids) {
      const r = CATALOG[id]
      expect(inTab(r, "quick") && inCat(r, "veg")).toBe(true)
      expect(id).not.toBe(recipes[0].id)
    }
  })

  it("keeps at least two slots for non-local recipes when there are some", () => {
    const ids = newDeck(recipes, state(), 1, NOW, seeded(4))
    expect(
      ids.filter((id) => isLocal(CATALOG[id], 1)).length,
    ).toBeLessThanOrEqual(DECK - 2)
  })
})
