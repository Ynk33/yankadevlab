"""Step 14: builds web/src/data/imported.json, the pipeline-owned part of the Assiette catalog, and validates it.

The hand-written catalog.json is only read, never written: imported.json holds the new recipes plus the aisles, English
names and seasons (utils/14_buildCatalog/seasons.json, versioned; [] means imported) of ingredients catalog.json lacks.
web/src/lib/catalog.ts merges both, catalog.json winning on conflicts. Writes data/14_buildCatalog/imported.json and
sources.json ({recipe id: Wikibooks URL, license, adaptations}); with --write, also replaces web/src/data/imported.json.
"""

import json
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/14_buildCatalog")]

from config import CATALOG_PATH, IMPORTED_PATH, WIKI_URLS
from read_json import read_json
from recipe_tuple import recipe_tuple
from slug import slug
from validate_catalog import validate_catalog
from write_json import write_json

LICENSE = "CC BY-SA 4.0"
SEASONS_PATH = Path(__file__).parent / "utils/14_buildCatalog/seasons.json"


def main() -> None:
    recipes = read_json("13_translate/recipes.json")
    products = read_json("06_mapVocabulary/products.json")
    seasons = json.loads(SEASONS_PATH.read_text())
    base = json.loads(CATALOG_PATH.read_text())
    imported = {"season": {}, "rayons": {}, "recipes": [], "en": {"ing": {}, "rec": {}}}
    ids = {raw[0] for raw in base["recipes"]}
    sources = {}
    for recipe in sorted(recipes, key=lambda r: r["title_fr"]):
        root = slug(recipe["title_fr"])
        rid, n = root, 2
        while rid in ids:
            rid, n = f"{root[:76]}-{n}", n + 1
        ids.add(rid)
        imported["recipes"].append(recipe_tuple(rid, recipe))
        imported["en"]["rec"][rid] = [recipe["en"]["title"], recipe["en"]["highlights"], recipe["en"]["steps"]]
        for item in recipe["ingredients"]:
            name = item["name"]
            if name not in base["rayons"]:
                imported["rayons"][name] = item["rayon"]
            english = products.get(name, {}).get("name_en")
            if english and name not in base["en"]["ing"]:
                imported["en"]["ing"][name] = english
            if name in seasons and name not in base["season"]:
                imported["season"][name] = seasons[name]
        sources[rid] = {"url": WIKI_URLS[recipe["lang"]] + recipe["title"].replace(" ", "_"), "source_title": recipe["title"], "license": LICENSE, "adaptations": recipe["adaptations"]}

    merged = {
        "season": {**imported["season"], **base["season"]},
        "rayons": {**imported["rayons"], **base["rayons"]},
        "recipes": base["recipes"] + imported["recipes"],
        "en": {"ing": {**imported["en"]["ing"], **base["en"]["ing"]}, "rec": {**imported["en"]["rec"], **base["en"]["rec"]}},
        "_new": {raw[0] for raw in imported["recipes"]},
    }
    errors = validate_catalog(merged)
    if errors:
        print(f"{len(errors)} validation errors, nothing written:")
        for error in errors[:20]:
            print(" ", error)
        sys.exit(1)
    write_json("14_buildCatalog/imported.json", imported)
    write_json("14_buildCatalog/sources.json", sources, indent=1)
    print(
        f"imported.json: {len(imported['recipes'])} recipes, {len(imported['rayons'])} new ingredients, {len(imported['season'])} seasons "
        f"| merged catalog {len(merged['recipes'])} recipes"
    )
    if "--write" in sys.argv:
        IMPORTED_PATH.write_text(json.dumps(imported, ensure_ascii=False, separators=(",", ":")))
        print(f"written to {IMPORTED_PATH}")


if __name__ == "__main__":
    main()
