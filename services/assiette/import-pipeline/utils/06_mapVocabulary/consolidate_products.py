import collections
import statistics

PRODUCT_FIELDS = ["name_en", "rayon", "role", "protein", "unit"]


def consolidate_products(vocabulary: dict[str, dict], key_counts: dict[str, int], catalog: dict) -> dict[str, dict]:
    """Merges the per-key LLM entries into one entry per French product.

    Categorical fields take the usage-weighted majority, conversion factors the median; products already in the
    catalog keep the catalog's aisle, English name and most used unit.
    """
    catalog_units: dict[str, collections.Counter] = collections.defaultdict(collections.Counter)
    for raw in catalog["recipes"]:
        for ingredient, _, unit in raw[9]:
            catalog_units[ingredient][unit] += 1
    groups: dict[str, list[str]] = collections.defaultdict(list)
    for key, entry in vocabulary.items():
        if entry.get("name_fr"):
            groups[entry["name_fr"]].append(key)
    products = {}
    for name, keys in groups.items():
        product = {"keys": sorted(keys, key=lambda k: -key_counts.get(k, 0)), "count": sum(key_counts.get(k, 0) for k in keys)}
        for field in PRODUCT_FIELDS:
            votes = collections.Counter()
            for key in keys:
                votes[vocabulary[key][field]] += key_counts.get(key, 1)
            product[field] = votes.most_common(1)[0][0]
        for field in ("density_g_per_ml", "piece_g"):
            values = [vocabulary[k][field] for k in keys if vocabulary[k][field] > 0]
            product[field] = round(statistics.median(values), 2) if values else 0
        if name in catalog["rayons"]:
            product["rayon"] = catalog["rayons"][name]
            product["unit"] = catalog_units.get(name, collections.Counter({product["unit"]: 1})).most_common(1)[0][0]
            product["name_en"] = catalog["en"]["ing"].get(name, product["name_en"])
            product["in_catalog"] = True
        else:
            product["in_catalog"] = False
        products[name] = product
    return products
