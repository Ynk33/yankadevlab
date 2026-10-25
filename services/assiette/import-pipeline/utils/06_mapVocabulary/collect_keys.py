import collections


def collect_keys(recipes: list[dict]) -> dict[str, dict]:
    """Groups the readable ingredient rows of kept recipes by key: {key: {count, langs, names (most frequent first)}}."""
    keys: dict[str, dict] = {}
    for recipe in recipes:
        if not recipe["kept"]:
            continue
        for row in recipe["ingredients"]:
            if row["status"] not in ("parsed", "unquantified"):
                continue
            key = row["key"] or row["name"]
            entry = keys.setdefault(key, {"count": 0, "langs": set(), "names": collections.Counter()})
            entry["count"] += 1
            entry["langs"].add(recipe["lang"])
            entry["names"][row["name"]] += 1
    return {k: {"count": v["count"], "langs": sorted(v["langs"]), "names": [n for n, _ in v["names"].most_common(3)]} for k, v in keys.items()}
