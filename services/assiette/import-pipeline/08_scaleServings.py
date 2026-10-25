"""Step 8: scales every recipe to 2 people and converts amounts to catalog units (g, ml, pièce, botte, c. à s., c. à c.).

Servings come from the page when stated, else from step 7's estimate. Liquid stock becomes cubes (1 per 500 ml), as in the catalog. Unquantified seasonings get a default amount;
unquantified main ingredients are listed in `unquantified_main` for step 9 to reject.
Writes data/08_scaleServings/recipes.json.
"""

import collections
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/08_scaleServings")]

from default_amount import default_amount
from explicit_servings import explicit_servings
from grams_to_catalog_unit import grams_to_catalog_unit
from round_quantity import round_quantity
from stock_cubes import stock_cubes
from load_pages import load_pages
from read_json import read_json
from unit_to_grams import unit_to_grams
from write_json import write_json

TARGET_SERVINGS = 2
MAX_SERVINGS = 20
PRODUCT_FIELDS = ("rayon", "role", "protein")


def scale_ingredients(rows: list[dict], factor: float, vocabulary: dict, products: dict) -> tuple[list[dict], list[str]]:
    """Converts the readable rows of one recipe to grams for 2 people, merges duplicate products and expresses them in catalog units."""
    merged: dict[str, dict] = {}
    unquantified_main = []
    for row in rows:
        if row["status"] not in ("parsed", "unquantified"):
            continue
        entry = vocabulary.get(row["key"] or row["name"], {})
        name = entry.get("name_fr")
        if not name:
            continue
        product = products[name]
        item = merged.setdefault(name, {
            "name": name, "grams": 0.0, "source_unit": None, "source_qty": 0.0, "defaulted": False, "cubes": 0.0,
            "processed": False, "hard_to_find": False, **{f: product[f] for f in PRODUCT_FIELDS},
        })
        item["processed"] |= bool(entry.get("processed"))
        item["hard_to_find"] |= bool(entry.get("hard_to_find"))
        cubes = stock_cubes(name, row["unit"], row["qty"]) if row["status"] == "parsed" else None
        if cubes is not None:
            item["cubes"] += cubes * factor
            item["grams"] += cubes * factor * (product.get("piece_g") or 10)
        elif row["status"] == "parsed":
            item["grams"] += unit_to_grams(row["qty"], row["unit"], product) * factor
            item["source_unit"] = item["source_unit"] or row["unit"]
            item["source_qty"] += row["qty"] * factor
        elif item["grams"] == 0:
            default = default_amount(product)
            if default is None:
                unquantified_main.append(name)
            else:
                item.update(qty=default[0], unit=default[1], grams=default[2], defaulted=True)
    ingredients = []
    for name, item in merged.items():
        if name in unquantified_main and item["grams"] == 0:
            continue
        if item["cubes"]:
            item["qty"], item["unit"] = round_quantity(item["cubes"], "pièce"), "pièce"
        elif not item["defaulted"]:
            item["qty"], item["unit"] = grams_to_catalog_unit(item["grams"], products[name], item["source_unit"], item["source_qty"])
        item["grams"] = round(item["grams"], 1)
        ingredients.append({k: item[k] for k in ("name", "qty", "unit", "grams", *PRODUCT_FIELDS, "processed", "hard_to_find", "defaulted")})
    return ingredients, sorted(set(unquantified_main) - {i["name"] for i in ingredients})


def main() -> None:
    recipes = [r for r in read_json("05_parseIngredients/recipes.json") if r["kept"]]
    classes = read_json("07_classify/classes.json")
    vocabulary = read_json("06_mapVocabulary/vocabulary.json")
    products = read_json("06_mapVocabulary/products.json")
    pages = load_pages()
    out, sources = [], collections.Counter()
    for recipe in recipes:
        stated = explicit_servings(pages[recipe["title"]])
        servings = stated or min(MAX_SERVINGS, max(1, classes[recipe["title"]]["servings_llm"]))
        sources["page" if stated else "llm"] += 1
        ingredients, unquantified_main = scale_ingredients(recipe["ingredients"], TARGET_SERVINGS / servings, vocabulary, products)
        out.append({
            "title": recipe["title"], "lang": recipe["lang"], "servings": servings, "servings_source": "page" if stated else "llm",
            **classes[recipe["title"]], "ingredients": ingredients, "unquantified_main": unquantified_main,
        })
    write_json("08_scaleServings/recipes.json", out, indent=1)
    units = collections.Counter(i["unit"] for r in out for i in r["ingredients"])
    print(f"scaled {len(out)} recipes | servings from {dict(sources)} | with unquantified main ingredient {sum(1 for r in out if r['unquantified_main'])}")
    print("catalog units:", dict(units.most_common()), "| defaulted seasonings", sum(i["defaulted"] for r in out for i in r["ingredients"]))


if __name__ == "__main__":
    main()
