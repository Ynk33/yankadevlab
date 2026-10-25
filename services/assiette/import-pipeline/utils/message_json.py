import json


def message_json(message: dict) -> dict | None:
    """Returns the JSON payload of a structured-output message, or None when it was refused or truncated."""
    if message.get("stop_reason") != "end_turn":
        return None
    text = next((b["text"] for b in message.get("content", []) if b.get("type") == "text"), "")
    return json.loads(text) if text else None
