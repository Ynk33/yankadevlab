import re

from explicit_categories import explicit_categories

INGREDIENTS_RX = re.compile(r"^=+\s*Ingr[ée]dients?[^\n]*\n(.*?)(?=^=+)", re.M | re.I | re.S)
PROCEDURE_RX = re.compile(r"^=+\s*(Pr[ée]paration|[ÉE]tapes|Recette|R[ée]alisation|Instructions)", re.M | re.I)
SWEET_RX = re.compile(
    r"dessert|pâtisserie|gâteau|tarte.? sucrée|cocktail|boisson|vodka|rhum|whisky|gin\b|liqueur|confiserie|biscuit"
    r"|confiture|crêpe|glace|sorbet|chocolat|sauce|condiment|pain\b|vin\b|bière|sirop|apéritif|tisane|thé\b|café|sans alcool",
    re.I,
)
QUANTITY_RX = re.compile(r"\d|½|¼|une?\b|demi")
MIN_LINES = 3
MIN_QUANTITY_RATIO = 0.7


def fr_candidate_stage(title: str, text: str) -> str:
    """Returns the first funnel stage a FR page fails ('sections', 'sweet', 'lines', 'quantities'), or 'ok'."""
    ingredients = INGREDIENTS_RX.search(text)
    if not ingredients or not PROCEDURE_RX.search(text) or "/Boissons/" in title:
        return "sections"
    if any(SWEET_RX.search(c) for c in explicit_categories(text)):
        return "sweet"
    lines = [l for l in ingredients.group(1).splitlines() if l.strip().startswith(("*", "#"))]
    if len(lines) < MIN_LINES:
        return "lines"
    if sum(1 for l in lines if QUANTITY_RX.search(l)) / len(lines) < MIN_QUANTITY_RATIO:
        return "quantities"
    return "ok"
