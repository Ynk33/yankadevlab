import re

STARTS_WITH_QUANTITY_RX = re.compile(r"^\s*(\d|½|¼|¾|⅓|⅔|une?\b|deux\b|trois\b|one\b|two\b|three\b|a\b|an\b)", re.I)
SEPARATOR_BEFORE_QUANTITY_RX = re.compile(r"\s*(?:,\s|;|\bet\b|\band\b|\+)\s*(?=(?:\d|½|¼|¾|une?\s|deux\s|trois\s))", re.I)


def split_line(raw_line: str) -> list[str]:
    """Splits a raw line holding several ingredients ("1 noix de beurre, 2 c. à s. d'huile" or "huile, sel, poivre")."""
    if not STARTS_WITH_QUANTITY_RX.match(re.sub(r"\{\{[^{}]*\}\}|\[\[[^\]]*\]\]", "x", raw_line)):
        if re.search(r"\d", raw_line):
            return [raw_line]
        return [part for part in re.split(r"\s*,\s*|\s+et\s+|\s+and\s+", raw_line) if part.strip()]
    parts = SEPARATOR_BEFORE_QUANTITY_RX.split(raw_line)
    return [part for part in parts if part.strip()]
