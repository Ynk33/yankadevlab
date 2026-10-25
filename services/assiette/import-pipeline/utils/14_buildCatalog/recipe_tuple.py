def recipe_tuple(recipe_id: str, recipe: dict) -> list:
    """Converts a pipeline recipe into the catalog's RawRecipe tuple (see web/src/lib/catalog.ts)."""
    return [
        recipe_id, recipe["title_fr"], recipe["meal"], recipe["protein"], recipe["minutes"],
        recipe["kcal"], recipe["proteines"], recipe["fibres"], recipe["highlights"],
        [[i["name"], i["qty"], i["unit"]] for i in recipe["ingredients"]],
        recipe["steps"], int(recipe["world"]), int(recipe["batch"]),
    ]
