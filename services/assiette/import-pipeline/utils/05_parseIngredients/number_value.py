import re

UNICODE_FRACTIONS = {"½": 0.5, "¼": 0.25, "¾": 0.75, "⅓": 1 / 3, "⅔": 2 / 3, "⅛": 0.125, "⅜": 0.375, "⅝": 0.625, "⅞": 0.875}
WORDS = {
    "un": 1, "une": 1, "a": 1, "an": 1, "one": 1, "deux": 2, "two": 2, "trois": 3, "three": 3, "quatre": 4, "four": 4,
    "cinq": 5, "five": 5, "six": 6, "huit": 8, "eight": 8, "dix": 10, "ten": 10, "douze": 12, "twelve": 12,
    "demi": 0.5, "half": 0.5, "une demi": 0.5, "un demi": 0.5, "a half": 0.5,
}


def number_value(token: str) -> float:
    """Converts one quantity token (\"1,5\", \"1 1/2\", \"2½\", \"¾\", \"une\", \"half\") into a float."""
    token = token.strip().lower()
    if token in WORDS:
        return float(WORDS[token])
    mixed = re.match(r"^(\d+)\s*([½¼¾⅓⅔⅛⅜⅝⅞])$", token)
    if mixed:
        return int(mixed.group(1)) + UNICODE_FRACTIONS[mixed.group(2)]
    if token in UNICODE_FRACTIONS:
        return UNICODE_FRACTIONS[token]
    whole_fraction = re.match(r"^(\d+)\s+(\d+)/(\d+)$", token)
    if whole_fraction:
        return int(whole_fraction.group(1)) + int(whole_fraction.group(2)) / int(whole_fraction.group(3))
    fraction = re.match(r"^(\d+)/(\d+)$", token)
    if fraction:
        return int(fraction.group(1)) / int(fraction.group(2)) if int(fraction.group(2)) else 0.0
    return float(token.replace(",", "."))
