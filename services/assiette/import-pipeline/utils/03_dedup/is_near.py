import difflib

CUTOFF = 0.85


def is_near(normalized: str, pool: list[str]) -> bool:
    """True when a normalized title is at least 85% similar to one in the pool."""
    return bool(difflib.get_close_matches(normalized, pool, n=1, cutoff=CUTOFF))
