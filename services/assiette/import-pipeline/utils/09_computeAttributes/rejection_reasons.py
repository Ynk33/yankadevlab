from grams_per_portion import grams_per_portion
from main_protein import main_protein

KEPT_DISH_TYPES = {"main", "soup", "salad", "sandwich"}
MIN_PROTEIN_G = 50
MAX_FAT_G = 25
MAX_CHEESE_G = 60
MAX_HARD_TO_FIND_G = 10
MAX_PROCESSED_G = 15


def rejection_reasons(recipe: dict, portions: int = 2) -> list[str]:
    """Lists every reason a scaled recipe fails the catalog rules (dish type, health thresholds per portion, sourcing)."""
    reasons = []
    if recipe["dish_type"] not in KEPT_DISH_TYPES:
        reasons.append(f"dish type: {recipe['dish_type']}")
    if recipe["unquantified_main"]:
        reasons.append(f"unquantified: {', '.join(recipe['unquantified_main'][:3])}")
    for item in recipe["ingredients"]:
        per_portion = item["grams"] / portions
        if item["hard_to_find"] and per_portion >= MAX_HARD_TO_FIND_G:
            reasons.append(f"hard to find: {item['name']}")
        if item["processed"] and (item["role"] == "protein" or per_portion >= MAX_PROCESSED_G):
            reasons.append(f"processed: {item['name']}")
    _, protein_g = main_protein(recipe["ingredients"], portions)
    if protein_g < MIN_PROTEIN_G:
        reasons.append("no main protein")
    by_role = grams_per_portion(recipe["ingredients"], portions)
    if by_role.get("fat", 0) > MAX_FAT_G:
        reasons.append("too much fat")
    if by_role.get("cheese", 0) > MAX_CHEESE_G:
        reasons.append("too much cheese")
    return reasons
