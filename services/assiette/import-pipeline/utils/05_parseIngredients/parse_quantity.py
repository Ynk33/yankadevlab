import re

from number_value import WORDS, number_value

NUMBER = r"(?:\d+\s+\d+/\d+|\d+\s*[½¼¾⅓⅔⅛⅜⅝⅞]|\d+/\d+|\d+(?:[.,]\d+)?|[½¼¾⅓⅔⅛⅜⅝⅞])"
WORD = r"(?:une demi|un demi|a half|" + "|".join(sorted((w for w in WORDS if " " not in w), key=len, reverse=True)) + r")"
QUANTITY_RX = re.compile(
    rf"^\s*(?:(?:about|approx\.?|approximately|around|roughly|environ|env\.|ca\.)\s+)?(?P<a>{NUMBER}|{WORD}\b)(?:\s*(?:-|–|—|à|to|or|ou)\s*(?P<b>{NUMBER}))?",
    re.I,
)


def parse_quantity(text: str) -> tuple[float | None, str]:
    """Reads a leading quantity (digits, fractions, unicode fractions, ranges as their mean, number words) and returns (value, rest)."""
    match = QUANTITY_RX.match(text)
    if not match:
        return None, text
    value = number_value(match.group("a"))
    if match.group("b"):
        value = (value + number_value(match.group("b"))) / 2
    return value, text[match.end():]
