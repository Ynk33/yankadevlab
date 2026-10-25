from config import MEALS

DISH_TYPES = ["main", "soup", "salad", "sandwich", "side", "starter", "breakfast", "dessert", "sauce", "base", "drink", "snack", "other"]


def output_schema() -> dict:
    """JSON schema of one classification answer."""
    return {
        "type": "object",
        "properties": {
            "dish_type": {"type": "string", "enum": DISH_TYPES},
            "meal": {"type": "string", "enum": MEALS},
            "world": {"type": "boolean"},
            "batch": {"type": "boolean"},
            "servings": {"type": "integer"},
        },
        "required": ["dish_type", "meal", "world", "batch", "servings"],
        "additionalProperties": False,
    }
