"""Step 9: decides which recipes enter the catalog and computes their attributes (main protein, missing starch).

Rules (per portion, recipes scaled to 2 people): kept dish types only, no unquantified main ingredient, no hard-to-find
ingredient beyond a trace, no cured meat or ultra-processed product beyond a trace, a main protein, enough vegetables,
bounded fat and cheese. Missing starch or vegetables (< 80 g per portion) is flagged for step 10 to complete the service. Writes data/09_computeAttributes/recipes.json ({kept: [...], rejected: [{title, reasons}]}).
"""

import collections
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/09_computeAttributes")]

from grams_per_portion import grams_per_portion
from main_protein import main_protein
from read_json import read_json
from rejection_reasons import rejection_reasons
from short_title import short_title
from write_json import write_json

STARCH_ROLES = {"starch", "legume"}
MIN_VEGETABLES_G = 80


def main() -> None:
    recipes = read_json("08_scaleServings/recipes.json")
    kept, rejected = [], []
    for recipe in recipes:
        reasons = rejection_reasons(recipe)
        if reasons:
            rejected.append({"title": recipe["title"], "lang": recipe["lang"], "reasons": reasons})
            continue
        category, protein_g = main_protein(recipe["ingredients"])
        by_role = grams_per_portion(recipe["ingredients"])
        kept.append({
            **recipe,
            "protein": category,
            "protein_g_per_portion": protein_g,
            "needs_starch": not any(i["role"] in STARCH_ROLES for i in recipe["ingredients"]) and recipe["dish_type"] != "sandwich",
            "vegetables_g_per_portion": round(by_role.get("vegetable", 0) + by_role.get("fruit", 0), 1),
            "needs_vegetables": by_role.get("vegetable", 0) + by_role.get("fruit", 0) < MIN_VEGETABLES_G,
        })
    write_json("09_computeAttributes/recipes.json", {"kept": kept, "rejected": rejected}, indent=1)

    langs = collections.Counter(r["lang"] for r in kept)
    print(f"recipes {len(recipes)} -> kept {langs['fr']} FR + {langs['en']} EN = {len(kept)} | rejected {len(rejected)}")
    first = collections.Counter(r["reasons"][0].split(":")[0] for r in rejected)
    every = collections.Counter(reason.split(":")[0] for r in rejected for reason in r["reasons"])
    print("first failing rule:", dict(first.most_common()))
    print("all failing rules:", dict(every.most_common()))
    print("protein:", dict(collections.Counter(r["protein"] for r in kept).most_common()), "| needs starch", sum(r["needs_starch"] for r in kept), "| needs vegetables", sum(r["needs_vegetables"] for r in kept))
    reference = read_json("reference/sample20.json")
    verdicts = read_json("reference/manual_verdicts.json")["decisions"]
    outcome = {r["title"]: "kept" for r in kept} | {r["title"]: "; ".join(r["reasons"]) for r in rejected}
    print("\nreference sample (manual verdict -> pipeline):")
    for sample, verdict in zip(reference, verdicts):
        print(f"  {verdict['verdict']:15} {short_title(sample['title'])[:34]:34} -> {outcome.get(sample['title'], 'rejected before step 9')}")


if __name__ == "__main__":
    main()
