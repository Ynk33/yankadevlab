"""Step 7 (LLM): classifies each recipe (dish type, meal, world, batch) and estimates its servings.

Writes data/07_classify/classes.json ({title: {dish_type, meal, world, batch, servings_llm}}).
Run: uv run --env-file ../.env python 07_classify.py   (rerun to collect if the batch is still processing)
"""

import collections
import hashlib
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/07_classify")]

from load_pages import load_pages
from message_json import message_json
from output_schema import output_schema
from read_json import read_json
from run_batch import run_batch
from structured_params import structured_params
from user_message import user_message
from write_json import write_json


def main() -> None:
    recipes = [r for r in read_json("05_parseIngredients/recipes.json") if r["kept"]]
    pages = load_pages()
    system = (Path(__file__).parent / "utils/07_classify/prompt.md").read_text()
    schema = output_schema()
    ids = {f"r-{hashlib.sha1(r['title'].encode()).hexdigest()[:12]}": r["title"] for r in recipes}
    langs = {r["title"]: r["lang"] for r in recipes}
    requests = {cid: structured_params(system, user_message(title, langs[title], pages[title]), schema, max_tokens=500) for cid, title in ids.items()}

    results = run_batch("07_classify", requests)
    if results is None:
        sys.exit(2)
    classes, missing = {}, []
    for cid, title in ids.items():
        payload = message_json(results.get(cid, {}))
        if payload is None:
            missing.append(title)
            continue
        payload["servings_llm"] = payload.pop("servings")
        classes[title] = payload
    write_json("07_classify/classes.json", classes, indent=1)
    print(f"classified {len(classes)}/{len(ids)} | missing {len(missing)}")
    print("dish types:", dict(collections.Counter(c["dish_type"] for c in classes.values()).most_common()))
    print("meal:", dict(collections.Counter(c["meal"] for c in classes.values())), "| world", sum(c["world"] for c in classes.values()), "| batch", sum(c["batch"] for c in classes.values()))
    if missing:
        print("missing:", missing[:10])


if __name__ == "__main__":
    main()
