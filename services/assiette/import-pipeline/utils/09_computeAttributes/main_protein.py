import collections

PROTEIN_ROLES = {"protein", "legume", "cheese"}


def main_protein(ingredients: list[dict], portions: int = 2) -> tuple[str, float]:
    """Returns the heaviest protein category and its grams per portion, or ("none", 0)."""
    totals: dict[str, float] = collections.defaultdict(float)
    for item in ingredients:
        if item["role"] in PROTEIN_ROLES and item["protein"] != "none":
            totals[item["protein"]] += item["grams"] / portions
    if not totals:
        return "none", 0.0
    category = max(totals, key=totals.get)
    return category, round(totals[category], 1)
