import re

from page_categories import page_categories
from procedure_chars import procedure_chars
from short_title import short_title

MIN_PROCEDURE_CHARS = 60
MAIN_COURSE_CATEGORIES = {"Plat principal", "Main course recipes"}
NON_MAIN_CATEGORY_RX = re.compile(
    r"^(accompagnements?|entrées?|tapas|amuse[- ]bouches?|apéritifs?|pains?|glacées?|recettes de galettes des rois"
    r"|side dish recipes|appetizer recipes|recipes for soda pop|marinade recipes|snack recipes|condiment recipes"
    r"|dip recipes|bread recipes|beverage recipes|dessert recipes|spice mix recipes|recipes for sauce)$",
    re.I,
)
NON_MAIN_TITLE_RX = re.compile(
    r"\b(fond (blanc|brun)|fumet|court-bouillon|mirepoix|roux|béchamel|vinaigrette|mayonnaise|marinade"
    r"|pâte (brisée|feuilletée|sablée|à pizza|à pain)|glace|sorbet|crème glacée|compote|confiture|smoothie|milk-?shake"
    r"|limonade|sirop|stock|broth base|dough|pie crust|seasoning|spice (mix|blend|rub)|dry rub|dressing|chutney|relish"
    r"|salsa|dip|ice cream|lemonade|soda|syrup|jam)\b",
    re.I,
)


def reject_reason(title: str, text: str) -> str | None:
    """Returns why a recipe is not a main dish (empty procedure, non-main category, base/sweet title), or None."""
    if procedure_chars(text) < MIN_PROCEDURE_CHARS:
        return "empty procedure"
    cats = page_categories(text)
    if not cats & MAIN_COURSE_CATEGORIES:
        non_main = sorted(c for c in cats if NON_MAIN_CATEGORY_RX.match(c))
        if non_main:
            return f"category: {non_main[0]}"
    match = NON_MAIN_TITLE_RX.search(short_title(title))
    if match:
        return f"title: {match.group(0).lower()}"
    return None
