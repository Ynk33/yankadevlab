from config import MODEL


def structured_params(system: str, user: str, schema: dict, max_tokens: int = 16000) -> dict:
    """Builds Messages API params for a Haiku call that must answer with JSON matching `schema` (system prompt cached)."""
    return {
        "model": MODEL,
        "max_tokens": max_tokens,
        "system": [{"type": "text", "text": system, "cache_control": {"type": "ephemeral"}}],
        "thinking": {"type": "disabled"},
        "output_config": {"effort": "low", "format": {"type": "json_schema", "schema": schema}},
        "messages": [{"role": "user", "content": user}],
    }
