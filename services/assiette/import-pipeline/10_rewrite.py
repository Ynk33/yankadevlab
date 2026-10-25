"""Step 10 (LLM): writes the French card of each kept recipe (title, 3-5 steps, total time) and completes the service.

Ingredients are frozen: the LLM only picks a starch or vegetable side from a closed list (asked when step 9 flagged one
as missing); any side it picks is added here with its amount, since the steps mention it, and recorded in `adaptations`.
Writes data/10_rewrite/recipes.json. Run: uv run --env-file ../.env python 10_rewrite.py
"""

import hashlib
import json
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/10_rewrite")]

from config import CATALOG_PATH
from load_pages import load_pages
from message_json import message_json
from output_schema import output_schema
from read_json import read_json
from run_batch import run_batch
from side_ingredient import side_ingredient
from structured_params import structured_params
from user_message import user_message
from write_json import write_json

MIN_STEPS, MAX_STEPS = 3, 5


def main() -> None:
    recipes = read_json("09_computeAttributes/recipes.json")["kept"]
    pages = load_pages()
    rayons = json.loads(CATALOG_PATH.read_text())["rayons"]
    system = (Path(__file__).parent / "utils/10_rewrite/prompt.md").read_text()
    schema = output_schema()
    ids = {f"r-{hashlib.sha1(r['title'].encode()).hexdigest()[:12]}": r for r in recipes}
    requests = {cid: structured_params(system, user_message(r, pages[r["title"]]), schema, max_tokens=2000) for cid, r in ids.items()}
    results = run_batch("10_rewrite", requests)
    if results is None:
        sys.exit(2)

    out, invalid = [], []
    for cid, recipe in ids.items():
        card = message_json(results.get(cid, {}))
        if card is None or not MIN_STEPS <= len(card["steps"]) <= MAX_STEPS or card["minutes"] <= 0:
            invalid.append(recipe["title"])
            continue
        ingredients, adaptations = list(recipe["ingredients"]), []
        for field, role in (("side_starch", "starch"), ("side_vegetable", "vegetable")):
            if card[field] != "none":
                ingredients.append(side_ingredient(card[field], role, rayons[card[field]]))
                adaptations.append(f"Service complété avec : {card[field]}")
        out.append({**recipe, "title_fr": card["title"], "steps": card["steps"], "minutes": card["minutes"], "ingredients": ingredients, "adaptations": adaptations})
    write_json("10_rewrite/recipes.json", out, indent=1)
    print(f"rewritten {len(out)}/{len(ids)} | invalid cards {len(invalid)} | adapted {sum(1 for r in out if r['adaptations'])}")
    if invalid:
        print("invalid:", invalid[:10])


if __name__ == "__main__":
    main()
