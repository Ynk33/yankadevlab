SPOON_ML = {"tbsp": 15, "tsp": 5}
COUNTABLE_GRAMS = {"slice": 30, "leaf": 0.5, "sprig": 2, "pinch": 0.5, "handful": 30, "knob": 10, "inch": 10, "cm": 4}
DEFAULT_CAN_GRAMS = 400
DEFAULT_PACKAGE_GRAMS = 250
DEFAULT_BUNCH_GRAMS = 50
DEFAULT_PIECE_GRAMS = 100
SEASONING_PIECE_GRAMS = {"fat": 15, "herb_spice": 2, "condiment": 10, "sweetener": 10}
BUTTER_STICK_GRAMS = 113


def unit_to_grams(qty: float, unit: str, product: dict) -> float:
    """Converts a parsed amount (g, ml, tbsp, tsp or a countable unit) to grams using the product's density and piece weight."""
    density = product.get("density_g_per_ml") or 1
    piece = product.get("piece_g") or 0
    if unit == "g":
        return qty
    if unit == "ml":
        return qty * density
    if unit in SPOON_ML:
        return qty * SPOON_ML[unit] * density
    if unit == "drizzle":
        return qty * 10 * density
    if unit in COUNTABLE_GRAMS:
        return qty * COUNTABLE_GRAMS[unit]
    if unit == "can":
        return qty * (piece if piece >= 150 else DEFAULT_CAN_GRAMS)
    if unit == "package":
        return qty * (piece if piece >= 50 else DEFAULT_PACKAGE_GRAMS)
    if unit == "bunch":
        return qty * DEFAULT_BUNCH_GRAMS
    if unit == "stick":
        return qty * (BUTTER_STICK_GRAMS if product.get("role") == "fat" else piece or DEFAULT_PIECE_GRAMS)
    return qty * (piece or SEASONING_PIECE_GRAMS.get(product.get("role"), DEFAULT_PIECE_GRAMS))
