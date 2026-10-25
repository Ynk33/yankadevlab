import collections


def recipe_nutrients(ingredients: list[dict], matches: dict[str, str], foods: dict[str, dict], portions: int = 2) -> tuple[dict[str, float], float]:
    """Sums Ciqual nutrients per portion over matched ingredients and returns (nutrients, share of grams matched)."""
    totals: dict[str, float] = collections.defaultdict(float)
    matched = total = 0.0
    for item in ingredients:
        total += item["grams"]
        code = matches.get(item["name"], "none")
        if code == "none" or code not in foods:
            continue
        matched += item["grams"]
        for nutrient, per_100g in foods[code]["nutrients"].items():
            totals[nutrient] += per_100g * item["grams"] / 100 / portions
    return dict(totals), (matched / total if total else 0.0)
