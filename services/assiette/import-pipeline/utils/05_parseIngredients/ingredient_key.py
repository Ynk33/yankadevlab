import re

from match_unit import UNITS

FR_TEMPLATE_RX = re.compile(r"\{\{\s*(?:i\'?|ingr[ée]dient)\s*\|([^}]*)\}\}", re.I)
LINK_RX = re.compile(r"\[\[\s*([^\]|]+?)\s*(?:\|[^\]]*)?\]\]")
NON_INGREDIENT_LINK_RX = re.compile(
    r"^(w:|image:|fichier:|file:|category:|catégorie:)|^cookbook:\s*(cups?|tablespoons?|teaspoons?|tbsp|tsp|pounds?|ounces?"
    r"|grams?|kilograms?|milliliters?|millilitres?|liters?|litres?|pints?|quarts?|gallons?|fluid ounce|inch|centimet(er|re)s?"
    r"|slicing|chopping|dicing|mincing|grating|julienne|frying|boiling|baking|roasting|simmering|sautéing|saut[ée]|blanching"
    r"|peeling|whisk(ing)?|beat(ing)?|mashing|marinating|steaming|grilling|braising|kneading|caramelization|cuisine.*)$",
    re.I,
)


def ingredient_key(raw_line: str) -> str:
    """Returns the canonical ingredient name given by the wiki markup ({{i|...}} in FR, [[Cookbook:...]] in EN), or ''."""
    template = FR_TEMPLATE_RX.search(raw_line)
    if template:
        params = [p.strip() for p in template.group(1).split("|") if "=" not in p and p.strip()]
        if params:
            return params[0]
    for target in LINK_RX.findall(raw_line):
        bare = target.split(":", 1)[-1].strip()
        if any(re.fullmatch(pattern.split("(?=")[0], bare, re.I) for pattern, _, _ in UNITS):
            continue
        if not NON_INGREDIENT_LINK_RX.match(target.strip()):
            return target.split(":", 1)[-1].strip() if target.lower().startswith("cookbook:") else target.strip()
    return ""
