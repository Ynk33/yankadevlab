import { describe, expect, it } from "vitest"
import { DEFAULT_STATE, normalizeState } from "@/lib/state"

describe("normalizeState", () => {
  it("returns the defaults without stored state", () => {
    expect(normalizeState(undefined)).toEqual(DEFAULT_STATE)
  })

  it("repairs invalid values and keeps unknown keys", () => {
    const st = normalizeState(
      JSON.parse(
        '{"tab":"nope","cat":"fish","favs":[],"prefs":{"persons":4,"exclusions":"porc","envies":""},"lang":"en","future":1}',
      ),
    )
    expect(st).toMatchObject({
      tab: "all",
      cat: "fish",
      favs: {},
      prefs: { persons: 4, exclusions: "porc", envies: "" },
      future: 1,
    })
    expect(st).not.toHaveProperty("lang")
  })

  it("migrates legacy extras and deck", () => {
    const st = normalizeState({
      extras: [{ id: "a", name: "pain", rayon: "Boulangerie" }],
      deck: ["x", "y"],
    })
    expect(st.extras).toEqual({ a: { name: "pain", rayon: "Boulangerie" } })
    expect(st.decks).toEqual({ all: ["x", "y"] })
    expect(st).not.toHaveProperty("deck")
  })
})
