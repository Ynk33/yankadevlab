import catalogText from "@/data/catalog.json?raw"
import { slug } from "@/lib/utils"

export interface Ingredient {
  name: string
  qty: number
  unit: string
  rayon?: string
}

export interface RecipeText {
  title: string
  highlights: string[]
  steps: string[]
}

export interface Recipe extends RecipeText {
  id: string
  meal: string
  protein: string
  minutes: number
  kcal?: number
  proteines?: number
  fibres?: number
  ingredients: Ingredient[]
  world: boolean
  batch: boolean
  base: number
  custom?: boolean
  en?: RecipeText
}

/** Custom recipe as stored by the API (written by the former Claude generation, nullable fields included). */
export interface CustomRecipeJSON {
  id: string
  title: string
  meal?: string
  protein: string
  minutes?: number
  kcal?: number
  proteines?: number
  fibres?: number
  highlights?: string[]
  ingredients: (Ingredient & { name_en?: string })[]
  steps: string[]
  world?: boolean
  batch?: boolean
  base?: number
  en?: RecipeText
}

type RawRecipe = [
  id: string,
  title: string,
  meal: string,
  protein: string,
  minutes: number,
  kcal: number,
  proteines: number,
  fibres: number,
  highlights: string[],
  ingredients: [name: string, qty: number, unit: string][],
  steps: string[],
  world: number,
  batch: number,
]

interface RawCatalog {
  season: Record<string, number[]>
  world: string[]
  rayons: Record<string, string>
  recipes: RawRecipe[]
  en: {
    ing: Record<string, string>
    rec: Record<string, [string, string[], string[]]>
  }
}

const raw: RawCatalog = JSON.parse(catalogText)

/** Protein → category. */
export const PROT: Record<string, Category> = {
  poisson: "fish",
  "fruits de mer": "fish",
  volaille: "poultry",
  "viande rouge": "meat",
  porc: "meat",
  légumineuses: "veg",
  œufs: "veg",
  "tofu/tempeh": "veg",
  fromage: "veg",
}
export const CATEGORIES = ["fish", "poultry", "meat", "veg"] as const
export type Category = (typeof CATEGORIES)[number]
export const CAT_KEYS = ["all", ...CATEGORIES] as const
export type CatKey = (typeof CAT_KEYS)[number]
export const TAB_KEYS = ["all", "quick", "batch", "dinner"] as const
export type TabKey = (typeof TAB_KEYS)[number]

export const RAYONS = [
  "Fruits & légumes",
  "Boucherie & poisson",
  "Crèmerie & œufs",
  "Boulangerie",
  "Épicerie",
  "Surgelés",
  "Herbes & épices",
]
export const UNITS = ["g", "ml", "pièce", "botte", "c. à s.", "c. à c."]

/** Ingredient name → months (1-12) when it's in season. Ingredients absent from the map are always in season. */
export const SEASON = raw.season
const RAYON_OF = raw.rayons
/** French ingredient name → English name. Extended by custom recipes. */
export const ING_EN: Record<string, string> = { ...raw.en.ing }

/** Every known recipe by id: the static catalog plus the team's custom recipes. */
export const CATALOG: Record<string, Recipe> = {}

for (const a of raw.recipes) {
  const e = raw.en.rec[a[0]]
  CATALOG[a[0]] = {
    id: a[0],
    title: a[1],
    meal: a[2],
    protein: a[3],
    minutes: a[4],
    kcal: a[5],
    proteines: a[6],
    fibres: a[7],
    highlights: a[8],
    ingredients: a[9].map(([name, qty, unit]) => ({ name, qty, unit })),
    steps: a[10],
    world: !!a[11],
    batch: !!a[12],
    base: 2,
    en: e ? { title: e[0], highlights: e[1], steps: e[2] } : undefined,
  }
}

/** Adds a custom recipe from the API to CATALOG. Invalid entries are ignored. */
export function registerCustom(d: CustomRecipeJSON | undefined) {
  if (!d || !d.id || !Array.isArray(d.ingredients)) return
  for (const i of d.ingredients) {
    if (i.name_en && !ING_EN[i.name]) ING_EN[i.name] = i.name_en
  }
  CATALOG[d.id] = {
    id: d.id,
    title: d.title,
    meal: d.meal || "both",
    protein: d.protein,
    minutes: d.minutes || 30,
    kcal: d.kcal || undefined,
    proteines: d.proteines || undefined,
    fibres: d.fibres || undefined,
    highlights: d.highlights || [],
    ingredients: d.ingredients.map((i) => ({
      name: i.name,
      qty: i.qty,
      unit: i.unit,
      rayon: i.rayon || undefined,
    })),
    steps: d.steps || [],
    world: !!d.world,
    batch: !!d.batch,
    base: d.base || 2,
    custom: true,
    en: d.en || undefined,
  }
}

export const inTab = (r: Recipe, t: TabKey) =>
  t === "quick"
    ? r.minutes <= 20
    : t === "batch"
      ? r.batch
      : t === "dinner"
        ? r.minutes > 20
        : true

export const inCat = (r: Recipe, c: CatKey) =>
  c === "all" || PROT[r.protein] === c

export const rayonOf = (i: Ingredient) =>
  i.rayon || RAYON_OF[i.name] || "Épicerie"

/** Ingredients of `r` that are out of season in `month` (1-12). */
export const offSeason = (r: Recipe, month: number) =>
  r.ingredients
    .filter((i) => SEASON[i.name] && !SEASON[i.name].includes(month))
    .map((i) => i.name)

export const isLocal = (r: Recipe, month: number) =>
  !r.world && offSeason(r, month).length === 0

/** True when the slug `q` appears in the recipe's title or ingredients, in French or English. */
export const matches = (r: Recipe, q: string) =>
  r.ingredients
    .flatMap((i) => [slug(i.name), slug(ING_EN[i.name])])
    .concat([slug(r.title), slug(r.en?.title)])
    .some((h) => !!h && h.includes(q))
