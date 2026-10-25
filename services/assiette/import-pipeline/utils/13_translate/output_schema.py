def output_schema() -> dict:
    """JSON schema of one translation answer."""
    return {
        "type": "object",
        "properties": {"title": {"type": "string"}, "steps": {"type": "array", "items": {"type": "string"}}},
        "required": ["title", "steps"],
        "additionalProperties": False,
    }
