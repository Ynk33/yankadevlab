DEFAULT_SPOONS = {"fat": ("c. à s.", 1), "herb_spice": ("c. à c.", 0.5), "condiment": ("c. à c.", 1), "sweetener": ("c. à c.", 1)}
SPOON_ML = {"c. à s.": 15, "c. à c.": 5}
HERB_BUNCH_FRACTION = 0.5
BUNCH_GRAMS = 50


def default_amount(product: dict) -> tuple[float, str, float] | None:
    """Amount for 2 people given to an unquantified seasoning (oil, herb, spice, condiment) as (qty, unit, grams); None for main ingredients."""
    role = product.get("role")
    if role not in DEFAULT_SPOONS:
        return None
    if role == "herb_spice" and product.get("unit") == "botte":
        return HERB_BUNCH_FRACTION, "botte", HERB_BUNCH_FRACTION * BUNCH_GRAMS
    unit, qty = DEFAULT_SPOONS[role]
    return qty, unit, qty * SPOON_ML[unit] * (product.get("density_g_per_ml") or 1)
