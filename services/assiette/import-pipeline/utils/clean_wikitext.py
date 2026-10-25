import re

INGREDIENT_TEMPLATE_RX = re.compile(r"\{\{\s*(?:i\'?|ingr[ée]dient)\s*\|(?:[^|}]*=[^|}]*\|)*([^|}]*)(?:\|([^|}]*))?\}\}")
LINK_RX = re.compile(r"\[\[(?:[^|\]]*\|)?([^\]]*)\]\]")
TEMPLATE_RX = re.compile(r"\{\{[^{}]*\}\}")
MARKUP_RX = re.compile(r"'''?|<[^>]+>")


def clean_wikitext(text: str) -> str:
    """Strips wiki markup: {{i|...}} ingredient templates, links, other templates, bold/italic and HTML tags."""
    text = INGREDIENT_TEMPLATE_RX.sub(lambda m: m.group(2) or m.group(1), text)
    text = LINK_RX.sub(r"\1", text)
    text = TEMPLATE_RX.sub("", text)
    return MARKUP_RX.sub("", text)
