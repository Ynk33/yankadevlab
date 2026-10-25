import re

from explicit_categories import explicit_categories
from summary_param import summary_param

INGREDIENTS_RX = re.compile(r"^=+\s*Ingredients?[^\n]*\n(.*?)(?=^==[^=]|\Z)", re.M | re.I | re.S)
PROCEDURE_RX = re.compile(r"^=+\s*(Procedure|Preparation|Method|Directions|Instructions|Steps)", re.M | re.I)
SWEET_CATEGORY_RX = re.compile(
    r"dessert|pastr|cake|cookie|pie\b|tart|cocktail|beverage|drink|wine|beer|liqueur|spirit|candy|confection|bread|jam"
    r"|jelly|preserve|sauce|condiment|dressing|spice mix|seasoning|ice cream|sorbet|chocolate|syrup|tea\b|coffee|smoothie"
    r"|frosting|icing|biscuit|muffin|pancake|baby food|pet food",
    re.I,
)
SWEET_SUMMARY_RX = re.compile(
    r"dessert|pastr|cake|cookie|pie\b|tart|cocktail|beverage|drink|wine|beer|liqueur|candy|confection|bread|jam|jelly"
    r"|preserve|sauce|condiment|dressing|spice|seasoning|ice cream|chocolate|syrup|smoothie|biscuit|muffin|pancake"
    r"|breakfast|snack|appetizer|baby",
    re.I,
)
QUANTITY_RX = re.compile(r"\d|½|¼|¾|⅓|\bone\b|\ba\b|\ban\b|half", re.I)
MIN_LINES = 3
MIN_QUANTITY_RATIO = 0.7


def en_candidate_stage(text: str) -> str:
    """Returns the first funnel stage an EN Cookbook page fails ('sections', 'sweet', 'lines', 'quantities', 'summary'), or 'ok'."""
    ingredients = INGREDIENTS_RX.search(text)
    if not ingredients or not PROCEDURE_RX.search(text):
        return "sections"
    if any(SWEET_CATEGORY_RX.search(c) for c in explicit_categories(text)):
        return "sweet"
    lines = [l for l in ingredients.group(1).splitlines() if l.strip().startswith(("*", "#", "|"))]
    if len(lines) < MIN_LINES:
        return "lines"
    if sum(1 for l in lines if QUANTITY_RX.search(l)) / len(lines) < MIN_QUANTITY_RATIO:
        return "quantities"
    if SWEET_SUMMARY_RX.search(summary_param(text, "category")):
        return "summary"
    return "ok"
