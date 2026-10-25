"""Step 12: removes the remaining duplicates (FR/EN versions of the same dish, near-identical variants) among the new
recipes and against the current catalog: same main protein, close French titles and overlapping main ingredients.

When two new recipes are duplicates, the one with the best nutrition coverage and the most ingredients is kept.
Writes data/12_finalDedup/recipes.json and duplicates.json.
"""

import json
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/12_finalDedup")]

from config import CATALOG_PATH
from is_duplicate import is_duplicate
from main_ingredients import main_ingredients
from read_json import read_json
from write_json import write_json

MIN_MAIN_INGREDIENTS = 2


def main() -> None:
    recipes = read_json("11_nutrition/recipes.json")
    catalog = json.loads(CATALOG_PATH.read_text())["recipes"]
    existing = [(raw[1], raw[3], {name for name, _, _ in raw[9]}) for raw in catalog]
    ranked = sorted(recipes, key=lambda r: (-r["nutrition_coverage"], -len(r["ingredients"]), r["lang"] != "fr", r["title"]))

    kept, duplicates = [], []
    for recipe in ranked:
        mains = main_ingredients(recipe["ingredients"])
        if len(mains) < MIN_MAIN_INGREDIENTS:
            kept.append(recipe)
            continue
        candidate = (recipe["title_fr"], recipe["protein"], mains)
        twin = next((other[0] for other in existing if is_duplicate(candidate, other)), None)
        if twin is None:
            twin = next((k["title_fr"] for k in kept if is_duplicate(candidate, (k["title_fr"], k["protein"], main_ingredients(k["ingredients"])))), None)
        if twin:
            duplicates.append({"title": recipe["title"], "title_fr": recipe["title_fr"], "duplicate_of": twin})
        else:
            kept.append(recipe)
    write_json("12_finalDedup/recipes.json", kept, indent=1)
    write_json("12_finalDedup/duplicates.json", duplicates, indent=1)
    print(f"recipes {len(recipes)} -> kept {len(kept)} | duplicates {len(duplicates)}")
    for d in duplicates[:12]:
        print(f"  {d['title_fr'][:45]:45} ≈ {d['duplicate_of'][:45]}")


if __name__ == "__main__":
    main()
