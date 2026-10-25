def output_schema() -> dict:
    """JSON schema of one Ciqual matching answer: {items: [{product, ciqual_code}]}."""
    item = {
        "type": "object",
        "properties": {"product": {"type": "string"}, "ciqual_code": {"type": "string"}},
        "required": ["product", "ciqual_code"],
        "additionalProperties": False,
    }
    return {"type": "object", "properties": {"items": {"type": "array", "items": item}}, "required": ["items"], "additionalProperties": False}
