import { describe, expect, it } from "vitest"
import { CATALOG } from "@/lib/catalog"
import { I18N } from "@/lib/i18n"
import { clampBought, fmtQty, shoppingGroups, totals } from "@/lib/shopping"
import { DEFAULT_STATE, type AppState } from "@/lib/state"

const salmon = CATALOG["saumon-roti-lentilles-vertes-et-epinards"]

function state(patch: Partial<AppState> = {}): AppState {
  return { ...structuredClone(DEFAULT_STATE), ...patch }
}

describe("fmtQty", () => {
  const fr = (q: number, u: string) => fmtQty(q, u, "fr", I18N.fr)
  const en = (q: number, u: string) => fmtQty(q, u, "en", I18N.en)

  it.each([
    [0, "g", ""],
    [0.4, "g", "1 g"],
    [43, "g", "43 g"],
    [147, "g", "150 g"],
    [1250, "g", "1,3 kg"],
    [1500, "ml", "1,5 L"],
    [0.2, "pièce", "0,5"],
    [1.3, "pièce", "1,5"],
    [1, "botte", "1 botte"],
    [1.5, "botte", "1,5 bottes"],
    [2, "c. à s.", "2 c. à s."],
  ])("fr %d %s → %s", (q, u, want) => {
    expect(fr(q, u)).toBe(want)
  })

  it("uses English units and decimal points", () => {
    expect(en(1250, "g")).toBe("1.3 kg")
    expect(en(1.5, "botte")).toBe("1.5 bunches")
    expect(en(2, "c. à s.")).toBe("2 tbsp")
  })
})

describe("totals", () => {
  it("scales quantities to the servings", () => {
    const t = totals({ [salmon.id]: { persons: 4, addedAt: 1 } }, CATALOG)
    expect(t.get("saumon|g")).toMatchObject({ qty: 520, unit: "g" })
  })

  it("ignores unknown recipes", () => {
    expect(totals({ nope: { persons: 2, addedAt: 1 } }, CATALOG).size).toBe(0)
  })
})

describe("clampBought", () => {
  it("drops items no longer needed and caps the others", () => {
    const t = totals({ [salmon.id]: { persons: 2, addedAt: 1 } }, CATALOG)
    expect(clampBought({ "saumon|g": 999, "nope|g": 1 }, t)).toEqual({
      "saumon|g": 260,
    })
  })
})

describe("shoppingGroups", () => {
  it("splits items between aisles and basket, extras included", () => {
    const st = state({
      mine: { [salmon.id]: { persons: 2, addedAt: 1 } },
      bought: { "saumon|g": 260, "lentille-verte|g": 50 },
      extras: {
        a: { name: "éponge", rayon: "Épicerie" },
        b: { name: "pain", rayon: "Boulangerie", done: true },
      },
    })
    const list = shoppingGroups(st, CATALOG, (n) => n, "fr")
    expect(list.count).toBe(salmon.ingredients.length + 2)
    expect(list.done.map((i) => i.name).sort()).toEqual(["pain", "saumon"])
    const items = list.groups.flatMap((g) => g.items)
    expect(items.find((i) => i.key === "lentille-verte|g")).toMatchObject({
      rem: 100,
      partial: true,
    })
    expect(items.some((i) => i.key === "x:a" && i.extra)).toBe(true)
    expect(list.groups.map((g) => g.rayon)).toEqual(
      [
        "Fruits & légumes",
        "Boucherie & poisson",
        "Crèmerie & œufs",
        "Boulangerie",
        "Épicerie",
        "Surgelés",
        "Herbes & épices",
      ].filter((r) => list.groups.some((g) => g.rayon === r)),
    )
  })
})
