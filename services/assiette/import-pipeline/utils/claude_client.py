import os

import anthropic


def claude_client() -> anthropic.Anthropic:
    """Returns an Anthropic client; run steps with `uv run --env-file ../.env` so ANTHROPIC_API_KEY is set."""
    if not os.environ.get("ANTHROPIC_API_KEY"):
        raise SystemExit("ANTHROPIC_API_KEY is not set: run with `uv run --env-file ../.env python <step>.py`")
    return anthropic.Anthropic()
