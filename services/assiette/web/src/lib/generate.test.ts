import { describe, expect, it } from "vitest"
import { generationError } from "@/lib/generate"
import { I18N } from "@/lib/i18n"

describe("generationError", () => {
  const t = I18N.fr
  it("maps the generate endpoint statuses to messages", () => {
    expect(generationError(422, t)).toBe(t.genDeclined)
    expect(generationError(429, t)).toBe(t.genTooMany)
    expect(generationError(503, t)).toBe(t.genOff)
  })
  it("falls back to a generic failure", () => {
    expect(generationError(502, t)).toBe(t.genFail)
    expect(generationError(0, t)).toBe(t.genFail)
  })
})
