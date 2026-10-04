import type { Category } from "@/lib/catalog"

export const CAT_BORDER: Record<Category, string> = {
  fish: "border-l-fish",
  poultry: "border-l-poultry",
  meat: "border-l-meat",
  veg: "border-l-veg",
}

export const CAT_DOT: Record<Category, string> = {
  fish: "before:bg-fish",
  poultry: "before:bg-poultry",
  meat: "before:bg-meat",
  veg: "before:bg-veg",
}
