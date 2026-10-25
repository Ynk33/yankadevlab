import re

SIZE_WORDS = r"(?:(?:extra[- ])?large|medium(?:[- ]sized?)?|small|big|fair[- ]sized|heaping|heaped|level|rounded|scant|generous|whole|fresh|gros(?:se)?s?|petite?s?|moyen(?:ne)?s?|bonne?s?|belle?s?|grande?s?)"
UNITS = [
    (r"kilos?|kilogrammes?|kilograms?|kg", "g", 1000),
    (r"milligrammes?|mg", "g", 0.001),
    (r"grammes?|grams?|grs?|g", "g", 1),
    (r"livres?|lbs?|pounds?", "g", 453.6),
    (r"fl\.?\s?oz|fluid ounces?", "ml", 29.57),
    (r"onces?|ounces?|oz", "g", 28.35),
    (r"litres?|liters?|l", "ml", 1000),
    (r"décilitres?|decilitres?|dl", "ml", 100),
    (r"centilitres?|cl", "ml", 10),
    (r"millilitres?|milliliters?|ml", "ml", 1),
    (r"cups?|c\.", "ml", 240),
    (r"tasses?", "ml", 250),
    (r"verres?|glass(?:es)?", "ml", 200),
    (r"bols?|bowls?", "ml", 350),
    (r"pints?", "ml", 473),
    (r"quarts?|qt", "ml", 946),
    (r"gallons?", "ml", 3785),
    (r"c\.?\s?à\.?\s?s\.?|càs|cs|cuill?[eè]re?s?\s+à\s+soupe|cuillerées?\s+à\s+soupe|c\.\s?à\s+soupe|tablespoons?|tbsps?|tbls?|tbs", "tbsp", 1),
    (r"c\.?\s?à\.?\s?c\.?|càc|cc|cuill?[eè]re?s?\s+à\s+(?:café|thé)|cuillerées?\s+à\s+(?:café|thé)|teaspoons?|tsps?", "tsp", 1),
    (r"cuill?[eè]re?s?|cuillerées?|spoons?", "tbsp", 1),
    (r"gousses?|cloves?", "clove", 1),
    (r"tranches?|slices?", "slice", 1),
    (r"bottes?|bouquets?|bunch(?:es)?", "bunch", 1),
    (r"branches?|brins?|tiges?|sprigs?|stalks?|stems?|ribs?", "sprig", 1),
    (r"pincées?|prises?|pinch(?:es)?|dash(?:es)?", "pinch", 1),
    (r"boîtes?|boites?|conserves?|cans?|tins?|jars?|bocaux|bocal", "can", 1),
    (r"sticks?|bâtons?", "stick", 1),
    (r"feuilles?|leaves|leaf", "leaf", 1),
    (r"poignées?|handfuls?", "handful", 1),
    (r"noix|noisettes?|knobs?|dabs?", "knob", 1),
    (r"filets?(?=\s+(?:d['’]|de\s+)(?:huile|vinaigre|jus|citron|sauce))|drizzles?", "drizzle", 1),
    (r"sachets?|paquets?|packages?|packets?|pkgs?|bags?|boxes?|envelopes?", "package", 1),
    (r"inch(?:es)?|in\.|pouces?", "inch", 1),
    (r"cm", "cm", 1),
    (r"morceaux?|pieces?|pièces?|pcs?|ea\.?|heads?|têtes?|bulbs?|bulbes?", "piece", 1),
]
UNIT_RX = re.compile(
    rf"^\s*(?:{SIZE_WORDS}\s+)*(?:(?P<unit>" + "|".join(f"(?:{pattern})" for pattern, _, _ in UNITS) + r")(?=[\s.,;:)(’']|$|d['’]|de\s))?",
    re.I,
)


def match_unit(text: str) -> tuple[str, float, str]:
    """Reads the unit following a quantity and returns (unit code, factor to that unit, rest); 'piece' when none is written."""
    match = UNIT_RX.match(text)
    token = match.group("unit") if match else None
    if not token:
        return "piece", 1, text[match.end():] if match else text
    in_context = text[match.start("unit"):]
    for pattern, unit, factor in UNITS:
        candidate = re.match(f"(?:{pattern})", in_context, re.I)
        if candidate and candidate.end() == len(token):
            return unit, factor, text[match.end():].lstrip(".")
    return "piece", 1, text
