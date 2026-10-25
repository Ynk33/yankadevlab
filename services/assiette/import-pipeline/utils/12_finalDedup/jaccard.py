def jaccard(a: set[str], b: set[str]) -> float:
    """Size of the intersection over size of the union (0 when both are empty)."""
    return len(a & b) / len(a | b) if a | b else 0.0
