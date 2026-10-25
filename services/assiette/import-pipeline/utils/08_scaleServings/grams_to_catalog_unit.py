from round_quantity import round_quantity
from unit_to_grams import DEFAULT_BUNCH_GRAMS

SPOON_SOURCES = {"tsp", "tbsp", "pinch", "g", "ml"}


def grams_to_catalog_unit(grams: float, product: dict, source_unit: str, source_qty: float) -> tuple[float, str]:
    """Expresses an amount in the product's catalog unit (g, ml, pièce, botte, c. à s., c. à c.), rounded; herbs measured by spoon or weight stay in c. à c."""
    unit = product.get("unit", "g")
    if unit == "botte" and source_unit in SPOON_SOURCES:
        unit = "c. à c."
    density = product.get("density_g_per_ml") or 1
    piece = product.get("piece_g") or 0
    if unit == "g":
        value = grams
    elif unit == "ml":
        value = grams / density
    elif unit == "pièce":
        if not piece:
            return round_quantity(grams, "g"), "g"
        value = grams / piece
    elif unit == "botte":
        value = source_qty if source_unit == "bunch" else grams / DEFAULT_BUNCH_GRAMS
    elif unit == "c. à s.":
        value = grams / density / 15
    else:
        value = grams / density / 5
    return round_quantity(value, unit), unit
