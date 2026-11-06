MIN_USES = 2
EXCLUDED_ROLES = {"other"}


def generator_products(products: dict[str, dict], vocabulary: dict[str, dict], catalog_names: set[str]) -> list[str]:
    """Products the recipe generator may use: every catalog ingredient, plus products seen in at least 2 recipes that
    are real ingredients and not only known as processed or hard to find."""
    flags: dict[str, list[int]] = {}
    for entry in vocabulary.values():
        name = entry.get("name_fr")
        if name:
            seen = flags.setdefault(name, [0, 0, 0])
            seen[0] += 1
            seen[1] += bool(entry.get("processed"))
            seen[2] += bool(entry.get("hard_to_find"))
    names = []
    for name, product in products.items():
        total, processed, hard = flags.get(name, [0, 0, 0])
        common = product["count"] >= MIN_USES and product["role"] not in EXCLUDED_ROLES and processed < total and hard < total
        if name in catalog_names or common:
            names.append(name)
    return sorted(names)
