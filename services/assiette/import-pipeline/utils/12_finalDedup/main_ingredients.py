MAIN_ROLES = {"protein", "legume", "starch", "vegetable", "fruit", "cheese", "nut_seed"}


def main_ingredients(ingredients: list[dict]) -> set[str]:
    """Names of the ingredients that define a dish (no seasonings, fats, liquids or added sides)."""
    return {i["name"] for i in ingredients if i["role"] in MAIN_ROLES and not i.get("side")}
