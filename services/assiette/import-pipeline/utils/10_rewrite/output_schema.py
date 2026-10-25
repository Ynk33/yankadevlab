from side_ingredient import STARCH_SIDES, VEGETABLE_SIDES


def output_schema() -> dict:
    """JSON schema of one rewrite answer."""
    return {
        "type": "object",
        "properties": {
            "title": {"type": "string"},
            "steps": {"type": "array", "items": {"type": "string"}},
            "minutes": {"type": "integer"},
            "side_starch": {"type": "string", "enum": ["none", *STARCH_SIDES]},
            "side_vegetable": {"type": "string", "enum": ["none", *VEGETABLE_SIDES]},
        },
        "required": ["title", "steps", "minutes", "side_starch", "side_vegetable"],
        "additionalProperties": False,
    }
