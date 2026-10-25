"""Step 5: parses every ingredient line into quantity, unit and name; rejects recipes with too many unreadable lines.

Writes data/05_parseIngredients/recipes.json: [{title, lang, ingredients: [...], unreadable_ratio, kept, reason}].
Only failed lines count as unreadable; unquantified lines are judged in step 7, once the vocabulary tells seasoning from main ingredients.
Units are normalised to g, ml, tbsp, tsp or a countable unit (piece, clove, slice, bunch, ...); step 7 converts them to catalog units.
"""

import collections
import random
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/05_parseIngredients")]

from ingredient_lines import ingredient_lines
from load_pages import load_pages
from parse_ingredient import parse_ingredient
from read_json import read_json
from short_title import short_title
from split_line import split_line
from write_json import write_json

MAX_UNREADABLE_RATIO = 0.2
MIN_PARSED = 3


def main() -> None:
    kept = read_json("04_prefilter/result.json")["kept"]
    pages = load_pages()
    recipes = []
    for lang in ("fr", "en"):
        for title in kept[lang]:
            rows = [parse_ingredient(segment, lang) for line in ingredient_lines(pages[title]) for segment in split_line(line)]
            counted = [r for r in rows if r["status"] in ("parsed", "unquantified", "failed")]
            parsed = sum(1 for r in counted if r["status"] == "parsed")
            failed = sum(1 for r in counted if r["status"] == "failed")
            ratio = round(failed / len(counted), 3) if counted else 1.0
            reason = None
            if parsed < MIN_PARSED:
                reason = f"fewer than {MIN_PARSED} parsed ingredients"
            elif ratio > MAX_UNREADABLE_RATIO:
                reason = f"{ratio:.0%} unreadable lines"
            recipes.append({"title": title, "lang": lang, "ingredients": rows, "unreadable_ratio": ratio, "kept": reason is None, "reason": reason})
    write_json("05_parseIngredients/recipes.json", recipes, indent=1)

    statuses = collections.Counter(r["status"] for recipe in recipes for r in recipe["ingredients"])
    units = collections.Counter(r["unit"] for recipe in recipes for r in recipe["ingredients"] if r["status"] == "parsed")
    kept_count = collections.Counter(r["lang"] for r in recipes if r["kept"])
    print(f"recipes {len(recipes)} -> kept {kept_count['fr']} FR + {kept_count['en']} EN = {sum(kept_count.values())}")
    print("line statuses:", dict(statuses))
    print("units:", dict(units.most_common()))
    print("rejections:", collections.Counter(r["reason"].split(" ", 1)[1] if r["reason"] and "%" in r["reason"] else r["reason"] for r in recipes if not r["kept"]))
    random.seed(5)
    for status in ("failed", "unquantified"):
        rows = [(short_title(recipe["title"]), r["text"]) for recipe in recipes for r in recipe["ingredients"] if r["status"] == status]
        print(f"\n{status} sample ({len(rows)}):")
        for title, text in random.sample(rows, min(15, len(rows))):
            print(f"  [{title[:25]}] {text[:110]}")


if __name__ == "__main__":
    main()
