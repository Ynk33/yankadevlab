from config import PROTEINS, RAYONS, UNITS

ROLES = ["protein", "legume", "starch", "vegetable", "fruit", "fat", "dairy", "cheese", "herb_spice", "condiment", "sweetener", "nut_seed", "liquid", "other"]


def output_schema() -> dict:
    """JSON schema of one vocabulary-mapping answer: {items: [one entry per input key]}."""
    item = {
        "type": "object",
        "properties": {
            "key": {"type": "string"},
            "name_fr": {"type": "string"},
            "name_en": {"type": "string"},
            "rayon": {"type": "string", "enum": RAYONS},
            "role": {"type": "string", "enum": ROLES},
            "protein": {"type": "string", "enum": PROTEINS + ["none"]},
            "processed": {"type": "boolean"},
            "hard_to_find": {"type": "boolean"},
            "unit": {"type": "string", "enum": UNITS},
            "density_g_per_ml": {"type": "number"},
            "piece_g": {"type": "number"},
        },
        "required": ["key", "name_fr", "name_en", "rayon", "role", "protein", "processed", "hard_to_find", "unit", "density_g_per_ml", "piece_g"],
        "additionalProperties": False,
    }
    return {"type": "object", "properties": {"items": {"type": "array", "items": item}}, "required": ["items"], "additionalProperties": False}
