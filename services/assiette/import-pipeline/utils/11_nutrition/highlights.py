REFERENCE_INTAKES = {
    "Vitamine C": ("vitamine_c", 80), "Fer": ("fer", 14), "Calcium": ("calcium", 800), "Magnésium": ("magnesium", 375),
    "Potassium": ("potassium", 2000), "Zinc": ("zinc", 10), "Iode": ("iode", 150), "Sélénium": ("selenium", 55),
    "Folates": ("folates", 200), "Vitamine B12": ("vitamine_b12", 2.5), "Vitamine D": ("vitamine_d", 5),
    "Vitamine K": ("vitamine_k", 75), "Bêta-carotène": ("beta_carotene", 4800),
}
MICRONUTRIENT_SHARE = 0.5
MACRO_THRESHOLDS = {"Protéines": ("proteines", 30), "Fibres": ("fibres", 8)}
OMEGA3_LONG_CHAIN_G = 0.25
OMEGA3_ALA_G = 1.1
MAX_HIGHLIGHTS = 3


def highlights(nutrients: dict[str, float], has_meat_or_fish: bool) -> list[str]:
    """Picks up to 3 nutritional benefits of a portion, strongest first, with the catalog's labels: micronutrients reaching
    50% of the EU reference intake (beta-carotene as vitamin A equivalents), proteins >= 30 g, fibres >= 8 g, omega-3
    from EPA+DHA >= 0.25 g or ALA >= 1.1 g. Vitamin B12 only counts without meat or fish, where nearly any portion reaches it."""
    ratios = {label: nutrients.get(key, 0) / (intake * MICRONUTRIENT_SHARE) for label, (key, intake) in REFERENCE_INTAKES.items()}
    if has_meat_or_fish:
        ratios.pop("Vitamine B12")
    ratios.update({label: nutrients.get(key, 0) / threshold for label, (key, threshold) in MACRO_THRESHOLDS.items()})
    ratios["Oméga-3"] = max(
        (nutrients.get("epa", 0) + nutrients.get("dha", 0)) / OMEGA3_LONG_CHAIN_G,
        nutrients.get("ala", 0) / OMEGA3_ALA_G,
    )
    reached = sorted((r, label) for label, r in ratios.items() if r >= 1)
    return [label for _, label in reversed(reached)][:MAX_HIGHLIGHTS]
