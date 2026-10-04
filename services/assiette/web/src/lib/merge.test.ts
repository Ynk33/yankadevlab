import { describe, expect, it } from "vitest"
import { mergeDiff } from "@/lib/merge"
import { isObj } from "@/lib/utils"

function applyPatch(target: unknown, patch: unknown): unknown {
  if (!isObj(patch)) return patch
  const out: Record<string, unknown> = isObj(target) ? { ...target } : {}
  for (const [k, v] of Object.entries(patch)) {
    if (v === null) delete out[k]
    else out[k] = applyPatch(out[k], v)
  }
  return out
}

describe("mergeDiff", () => {
  it("returns undefined when nothing changed", () => {
    expect(mergeDiff({ a: 1, b: { c: [1] } }, { a: 1, b: { c: [1] } })).toBe(
      undefined,
    )
  })

  it("only sends changed keys, nulls removed ones", () => {
    expect(
      mergeDiff(
        { a: 1, bought: { x: 1, y: 2 }, deck: [1, 2] },
        { a: 1, bought: { y: 2, z: 3 }, deck: [4] },
      ),
    ).toEqual({ bought: { x: null, z: 3 }, deck: [4] })
  })

  it.each([
    [{}, { a: { b: 1 } }],
    [{ a: 1 }, { a: { b: 1 } }],
    [{ a: { b: 1 } }, { a: 2 }],
    [{ a: 1, b: { c: 1, d: 2 } }, { b: { d: 3 } }],
    [{ mine: { x: { persons: 2 } } }, { mine: {} }],
  ])("round-trips through a merge patch: %j → %j", (a, b) => {
    expect(applyPatch(a, mergeDiff(a, b))).toEqual(b)
  })
})
