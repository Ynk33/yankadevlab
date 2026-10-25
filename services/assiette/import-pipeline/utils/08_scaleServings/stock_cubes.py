ML_PER_CUBE = 500
VOLUME_UNITS = {"ml", "tbsp", "tsp", "drizzle"}
SPOON_ML = {"tbsp": 15, "tsp": 5, "drizzle": 10, "ml": 1}


def stock_cubes(name: str, unit: str, qty: float) -> float | None:
    """Following the catalog's convention (stock as cubes, water left out), turns a liquid stock amount into cubes
    (1 cube per 500 ml); None when the row is not a liquid stock."""
    if not name.startswith("bouillon") or unit not in VOLUME_UNITS:
        return None
    return qty * SPOON_ML[unit] / ML_PER_CUBE
