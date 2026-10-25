"""Step 6 (LLM): maps every ingredient key to a canonical French product with aisle, role, protein category and conversion factors.

Writes data/06_mapVocabulary/vocabulary.json ({key: {name_fr, processed, hard_to_find}}), products.json ({name_fr: consolidated
product}) and products.tsv for review.
Manual fixes go in utils/06_mapVocabulary/overrides.json (versioned), applied on top of the LLM output:
{"keys": {key: {field: value}}, "rename": {llm name_fr: name_fr to use}, "products": {name_fr: {field: value}}}.
Run: uv run --env-file ../.env python 06_mapVocabulary.py   (rerun to collect if the batch is still processing)
"""

import hashlib
import json
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/06_mapVocabulary")]

from collect_keys import collect_keys
from config import CATALOG_PATH, DATA_DIR
from consolidate_products import consolidate_products
from message_json import message_json
from output_schema import output_schema
from read_json import read_json
from run_batch import run_batch
from structured_params import structured_params
from system_prompt import system_prompt
from write_json import write_json

CHUNK = 80
BUCKETS = 32
RETRY_ROUNDS = 2
OVERRIDES_PATH = Path(__file__).parent / "utils/06_mapVocabulary/overrides.json"
TSV_FIELDS = ["name_fr", "count", "in_catalog", "name_en", "rayon", "role", "protein", "unit", "density_g_per_ml", "piece_g", "processed_keys", "hard_to_find_keys", "keys"]


def main() -> None:
    keys = collect_keys(read_json("05_parseIngredients/recipes.json"))
    ordered = sorted(keys, key=lambda k: (-keys[k]["count"], k))
    buckets: dict[int, list[str]] = {}
    for key in ordered:
        buckets.setdefault(int(hashlib.sha1(key.encode()).hexdigest(), 16) % BUCKETS, []).append(key)
    system, schema = system_prompt(), output_schema()
    chunk_ids = {f"bucket-{index:02d}": buckets[index] for index in sorted(buckets)}
    requests, vocabulary = {}, {}
    for round_index in range(RETRY_ROUNDS + 1):
        for cid, chunk in chunk_ids.items():
            lines = "\n".join(json.dumps({"key": k, "langs": keys[k]["langs"], "raw_names": keys[k]["names"]}, ensure_ascii=False) for k in chunk)
            requests[cid] = structured_params(system, f"Map these {len(chunk)} ingredient keys:\n{lines}", schema)
        results = run_batch("06_mapVocabulary", requests)
        if results is None:
            sys.exit(2)
        for cid in chunk_ids:
            payload = message_json(results.get(cid, {})) or {"items": []}
            for item in payload["items"]:
                if item["key"] in keys and item["key"] not in vocabulary:
                    vocabulary[item["key"]] = {k: v for k, v in item.items() if k != "key"}
        missing = [k for k in ordered if k not in vocabulary]
        if not missing or round_index == RETRY_ROUNDS:
            break
        chunk_ids = {f"retry{round_index + 1}-{i:03d}": missing[i:i + CHUNK] for i in range(0, len(missing), CHUNK)}
    overrides = json.loads(OVERRIDES_PATH.read_text()) if OVERRIDES_PATH.exists() else {}
    for key, fields in overrides.get("keys", {}).items():
        vocabulary.setdefault(key, {}).update(fields)
    for entry in vocabulary.values():
        entry["name_fr"] = overrides.get("rename", {}).get(entry.get("name_fr"), entry.get("name_fr"))
    product_overrides = overrides.get("products", {})
    for entry in vocabulary.values():
        entry.update({f: v for f, v in product_overrides.get(entry.get("name_fr"), {}).items() if f in ("processed", "hard_to_find")})

    products = consolidate_products(vocabulary, {k: v["count"] for k, v in keys.items()}, json.loads(CATALOG_PATH.read_text()))
    for name, fields in product_overrides.items():
        if name in products:
            products[name].update({f: v for f, v in fields.items() if f not in ("processed", "hard_to_find")})
    write_json("06_mapVocabulary/vocabulary.json", {k: {f: v.get(f) for f in ("name_fr", "processed", "hard_to_find")} for k, v in vocabulary.items()}, indent=1)
    write_json("06_mapVocabulary/products.json", products, indent=1)
    rows = ["\t".join(TSV_FIELDS)]
    for name, product in sorted(products.items(), key=lambda item: -item[1]["count"]):
        flagged = {
            "processed_keys": ", ".join(k for k in product["keys"] if vocabulary[k]["processed"]),
            "hard_to_find_keys": ", ".join(k for k in product["keys"] if vocabulary[k]["hard_to_find"]),
        }
        entry = {**product, **flagged, "name_fr": name, "keys": ", ".join(product["keys"])}
        rows.append("\t".join(str(entry.get(f, "")) for f in TSV_FIELDS))
    (DATA_DIR / "06_mapVocabulary/products.tsv").write_text("\n".join(rows) + "\n")

    print(
        f"keys {len(keys)} -> {len(products)} products ({sum(p['in_catalog'] for p in products.values())} already in catalog) "
        f"| non-ingredients {sum(1 for v in vocabulary.values() if not v.get('name_fr'))} | missing {len(missing)} | overrides {len(overrides.get('keys', {}))} keys, {len(overrides.get('rename', {}))} renames, {len(product_overrides)} products"
    )
    if missing:
        print(f"still missing after {RETRY_ROUNDS} retries:", missing[:20])


if __name__ == "__main__":
    main()
