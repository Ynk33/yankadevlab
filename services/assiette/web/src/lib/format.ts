import { ING_EN, type Recipe, type RecipeText } from "@/lib/catalog"
import type { Dict, Lang } from "@/lib/i18n"

/** Recipe title, highlights and steps in the given language, falling back to French. */
export const recipeText = (r: Recipe, lang: Lang): RecipeText =>
  lang === "en" && r.en ? r.en : r

export const ingName = (name: string, lang: Lang) =>
  lang === "en" ? ING_EN[name] || name : name

export const protName = (p: string, t: Dict) => t.prot[p] || p

export const rayonName = (r: string, t: Dict) => t.rayon[r] || r

/** "poisson, 30 min, 620 kcal", plus a mention for custom recipes. */
export const metaLine = (r: Recipe, t: Dict) =>
  [
    protName(r.protein, t),
    r.minutes ? r.minutes + " min" : "",
    r.kcal ? r.kcal + " kcal" : "",
  ]
    .filter(Boolean)
    .join(", ") + (r.custom ? t.byClaude : "")

export const fmtDate = (iso: string, lang: Lang) =>
  new Date(iso).toLocaleDateString(lang === "en" ? "en-GB" : "fr-FR", {
    day: "numeric",
    month: "short",
    year: "numeric",
  })
