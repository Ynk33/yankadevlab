import collections


def grams_per_portion(ingredients: list[dict], portions: int = 2) -> dict[str, float]:
    """Sums ingredient grams per portion by role (protein, vegetable, fat, cheese, ...)."""
    totals: dict[str, float] = collections.defaultdict(float)
    for item in ingredients:
        totals[item["role"]] += item["grams"] / portions
    return dict(totals)
