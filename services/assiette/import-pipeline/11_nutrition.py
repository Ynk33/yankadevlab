"""Step 11 (LLM-assisted): matches every product to a Ciqual 2025 food, then computes kcal, proteins, fibres and
highlights per portion deterministically.

Candidates come from a word-overlap search; the LLM only picks one of them or "none". Manual fixes go in
utils/11_nutrition/overrides.json ({product: ciqual code or "none"}, versioned). Recipes whose matched ingredients
cover less than 85% of their weight, or whose energy per portion falls outside 250-1200 kcal (a sign of wrong
quantities or servings in the source), are dropped. Writes data/11_nutrition/matches.tsv and recipes.json.
Run: uv run --env-file ../.env python 11_nutrition.py
"""

import collections
import json
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/11_nutrition")]

from candidate_foods import candidate_foods
from config import CATALOG_PATH, DATA_DIR
from highlights import highlights
from load_ciqual import load_ciqual
from message_json import message_json
from name_tokens import name_tokens
from output_schema import output_schema
from read_json import read_json
from recipe_nutrients import recipe_nutrients
from run_batch import run_batch
from structured_params import structured_params
from write_json import write_json

CHUNK = 25
MEAT_AND_FISH = {"poisson", "fruits de mer", "volaille", "viande rouge", "porc"}
MIN_COVERAGE = 0.85
MIN_KCAL, MAX_KCAL = 250, 1200
OVERRIDES_PATH = Path(__file__).parent / "utils/11_nutrition/overrides.json"


def main() -> None:
    recipes = read_json("10_rewrite/recipes.json")
    products = read_json("06_mapVocabulary/products.json")
    english = json.loads(CATALOG_PATH.read_text())["en"]["ing"]
    foods = load_ciqual()
    tokens = {code: (name_tokens(f["name_fr"]), name_tokens(f["name_en"])) for code, f in foods.items()}

    names = sorted({i["name"] for r in recipes for i in r["ingredients"]})
    name_en = {n: products.get(n, {}).get("name_en") or english.get(n, "") for n in names}
    candidates = {n: candidate_foods(n, name_en[n], foods, tokens) for n in names}
    system, schema = (Path(__file__).parent / "utils/11_nutrition/prompt.md").read_text(), output_schema()
    askable = [n for n in names if candidates[n]]
    requests = {}
    for index in range(0, len(askable), CHUNK):
        chunk = askable[index:index + CHUNK]
        body = "\n\n".join(
            f"product: {n}\nenglish: {name_en[n]}\ncandidates:\n" + "\n".join(f"  {c}: {foods[c]['name_fr']}" for c in candidates[n]) for n in chunk
        )
        requests[f"chunk-{index // CHUNK:03d}"] = structured_params(system, body, schema, max_tokens=4000)
    results = run_batch("11_nutrition", requests)
    if results is None:
        sys.exit(2)

    matches = {n: "none" for n in names}
    for message in results.values():
        for item in (message_json(message) or {"items": []})["items"]:
            if item["product"] in matches and item["ciqual_code"] in candidates[item["product"]]:
                matches[item["product"]] = item["ciqual_code"]
    overrides = json.loads(OVERRIDES_PATH.read_text()) if OVERRIDES_PATH.exists() else {}
    matches.update({n: c for n, c in overrides.items() if n in matches})

    usage = collections.Counter(i["name"] for r in recipes for i in r["ingredients"])
    rows = ["product\tuses\tciqual_code\tciqual_name"] + [
        f"{n}\t{usage[n]}\t{matches[n]}\t{foods.get(matches[n], {}).get('name_fr', '')}" for n in sorted(names, key=lambda n: -usage[n])
    ]
    (DATA_DIR / "11_nutrition").mkdir(parents=True, exist_ok=True)
    (DATA_DIR / "11_nutrition/matches.tsv").write_text("\n".join(rows) + "\n")

    kept, dropped, implausible = [], [], []
    for recipe in recipes:
        nutrients, coverage = recipe_nutrients(recipe["ingredients"], matches, foods)
        if coverage < MIN_COVERAGE:
            dropped.append((recipe["title"], round(coverage, 2)))
            continue
        if not MIN_KCAL <= nutrients.get("kcal", 0) <= MAX_KCAL:
            implausible.append((recipe["title"], round(nutrients.get("kcal", 0))))
            continue
        kept.append({
            **recipe,
            "kcal": round(nutrients.get("kcal", 0)), "proteines": round(nutrients.get("proteines", 0)), "fibres": round(nutrients.get("fibres", 0)),
            "highlights": highlights(nutrients, any(i["protein"] in MEAT_AND_FISH for i in recipe["ingredients"])), "nutrition_coverage": round(coverage, 2),
        })
    write_json("11_nutrition/recipes.json", kept, indent=1)
    matched = sum(1 for c in matches.values() if c != "none")
    print(f"products {len(names)} -> matched {matched}, none {len(names) - matched} | overrides {len(overrides)}")
    print(f"recipes {len(recipes)} -> kept {len(kept)} | dropped for coverage < {MIN_COVERAGE:.0%}: {len(dropped)} | implausible energy: {len(implausible)}")
    kcal = sorted(r["kcal"] for r in kept)
    if kcal:
        print(f"kcal per portion: median {kcal[len(kcal) // 2]}, p10 {kcal[len(kcal) // 10]}, p90 {kcal[9 * len(kcal) // 10]}")
    print("highlights:", dict(collections.Counter(h for r in kept for h in r["highlights"]).most_common()))


if __name__ == "__main__":
    main()
