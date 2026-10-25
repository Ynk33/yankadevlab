import re

CATEGORY_RX = re.compile(r"\[\[\s*(?:Cat[ée]gorie|Category)\s*:\s*([^\]|]+)", re.I)


def explicit_categories(text: str) -> list[str]:
    """Returns the [[Catégorie:...]] / [[Category:...]] names written in a page."""
    return [c.strip() for c in CATEGORY_RX.findall(text)]
