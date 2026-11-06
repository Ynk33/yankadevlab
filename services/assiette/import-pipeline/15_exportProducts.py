"""Step 15: exports the products the recipe generator may use to services/assiette/data/products.json (versioned,
embedded by the Go backend).

Each product carries what the backend needs to keep generated recipes consistent with the catalog: aisle, English name,
role, protein category, catalog unit, conversion factors (density, piece weight) and Ciqual 2025 nutrients per 100 g,
so nutrition is computed rather than written by the LLM. Products without a Ciqual match are left out.
"""

import json
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils")]

from config import CATALOG_PATH, PRODUCTS_EXPORT_PATH
from generator_products import generator_products
from load_ciqual import load_ciqual
from read_json import read_json

FIELDS = ("name_en", "rayon", "role", "protein", "unit", "density_g_per_ml", "piece_g")


def main() -> None:
    products = read_json("06_mapVocabulary/products.json")
    vocabulary = read_json("06_mapVocabulary/vocabulary.json")
    matches = read_json("11_nutrition/matches.json")
    catalog = json.loads(CATALOG_PATH.read_text())
    foods = load_ciqual()

    exported, unmatched = {}, []
    for name in generator_products(products, vocabulary, set(catalog["rayons"])):
        code = matches.get(name, "none")
        if code not in foods:
            unmatched.append(name)
            continue
        entry = {field: products[name][field] for field in FIELDS}
        entry["rayon"] = catalog["rayons"].get(name, entry["rayon"])
        entry["name_en"] = catalog["en"]["ing"].get(name, entry["name_en"])
        entry["ciqual_code"] = code
        entry["nutrients"] = {k: round(v, 3) for k, v in sorted(foods[code]["nutrients"].items())}
        exported[name] = entry

    PRODUCTS_EXPORT_PATH.parent.mkdir(parents=True, exist_ok=True)
    PRODUCTS_EXPORT_PATH.write_text(json.dumps(exported, ensure_ascii=False, indent=1, sort_keys=True) + "\n")
    missing_catalog = [n for n in unmatched if n in catalog["rayons"]]
    print(f"exported {len(exported)} products to {PRODUCTS_EXPORT_PATH} | unmatched {len(unmatched)} (catalog ones: {missing_catalog})")


if __name__ == "__main__":
    main()
