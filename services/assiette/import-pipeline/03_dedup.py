"""Step 3: removes duplicates between FR and EN candidates, inside EN, and against the current catalog.

Writes data/03_dedup/pool.json ({fr: [titles], en: [titles]}).
"""

import collections
import json
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/03_dedup")]

from config import CATALOG_PATH
from is_near import is_near
from linked_pairs import linked_pairs
from load_pages import load_pages
from normalize_title import normalize_title
from read_json import read_json
from write_json import write_json


def main() -> None:
    candidates = read_json("02_extract/candidates.json")
    pages = load_pages()
    fr_titles = {p["title"] for p in read_json("02_extract/pages_fr.json")}
    pairs = linked_pairs(read_json("01_fetch/wikidata_pairs.json"), candidates["fr"], pages, fr_titles)

    linked = {en for _, en in pairs}
    en_unlinked = [t for t in candidates["en"] if t not in linked]
    fr_normalized = {normalize_title(t) for t in candidates["fr"]}
    en_distinct = [t for t in en_unlinked if not (normalize_title(t) and normalize_title(t) in fr_normalized)]
    variants = collections.defaultdict(list)
    for title in en_distinct:
        variants[normalize_title(title)].append(title)
    en_representatives = [group[0] for group in variants.values()]

    catalog = json.loads(CATALOG_PATH.read_text())
    catalog_fr = [normalize_title(r[1]) for r in catalog["recipes"]]
    catalog_en = [normalize_title(r[0]) for r in catalog["en"]["rec"].values()]
    fr_pool = [t for t in candidates["fr"] if not is_near(normalize_title(t), catalog_fr)]
    en_pool = [t for t in en_representatives if not is_near(normalize_title(t), catalog_en)]

    write_json("03_dedup/pool.json", {"fr": fr_pool, "en": en_pool}, indent=1)
    print(f"FR {len(candidates['fr'])} -> {len(fr_pool)} (catalog duplicates {len(candidates['fr']) - len(fr_pool)})")
    print(
        f"EN {len(candidates['en'])} -> {len(en_pool)} (linked {len(candidates['en']) - len(en_unlinked)}, "
        f"same title as FR {len(en_unlinked) - len(en_distinct)}, EN variants {len(en_distinct) - len(en_representatives)}, "
        f"catalog duplicates {len(en_representatives) - len(en_pool)})"
    )
    print(f"pool: {len(fr_pool) + len(en_pool)}")


if __name__ == "__main__":
    main()
