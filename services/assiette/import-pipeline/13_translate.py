"""Step 13 (LLM): translates each recipe's French title and steps into English.

Ingredient names come from the vocabulary (step 6) and highlights from the catalog's existing FR→EN labels, so the LLM
only translates free text. Writes data/13_translate/recipes.json. Run: uv run --env-file ../.env python 13_translate.py
"""

import hashlib
import json
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/13_translate")]

from config import CATALOG_PATH
from highlight_translations import highlight_translations
from message_json import message_json
from output_schema import output_schema
from read_json import read_json
from run_batch import run_batch
from structured_params import structured_params
from write_json import write_json


def main() -> None:
    recipes = read_json("12_finalDedup/recipes.json")
    catalog = json.loads(CATALOG_PATH.read_text())
    labels = highlight_translations(catalog)
    system, schema = (Path(__file__).parent / "utils/13_translate/prompt.md").read_text(), output_schema()
    ids = {f"r-{hashlib.sha1(r['title'].encode()).hexdigest()[:12]}": r for r in recipes}
    requests = {
        cid: structured_params(system, json.dumps({"title": r["title_fr"], "steps": r["steps"]}, ensure_ascii=False), schema, max_tokens=2000)
        for cid, r in ids.items()
    }
    results = run_batch("13_translate", requests)
    if results is None:
        sys.exit(2)
    out, invalid = [], []
    for cid, recipe in ids.items():
        english = message_json(results.get(cid, {}))
        if english is None or len(english["steps"]) != len(recipe["steps"]):
            invalid.append(recipe["title"])
            continue
        out.append({**recipe, "en": {"title": english["title"], "steps": english["steps"], "highlights": [labels.get(h, h) for h in recipe["highlights"]]}})
    write_json("13_translate/recipes.json", out, indent=1)
    print(f"translated {len(out)}/{len(ids)} | invalid {len(invalid)}")
    if invalid:
        print("invalid:", invalid[:10])


if __name__ == "__main__":
    main()
