def round_quantity(qty: float, unit: str) -> float:
    """Rounds to kitchen-friendly amounts: 5 g/ml steps under 100, 10 above; halves for pieces, bunches and spoons."""
    if unit in ("g", "ml"):
        step = 5 if qty < 100 else 10
        return max(step, round(qty / step) * step)
    return max(0.5, round(qty * 2) / 2)
