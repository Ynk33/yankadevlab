import re

from match_unit import SIZE_WORDS

LEADING_RX = re.compile(rf"^\s*(?:(?:(?:de la|de|du|des|of|the|some|a little|a few|quelques|un peu de)\b|de l['’]|d['’]|un peu d['’])\s*|{SIZE_WORDS}\s+|[-–,:.]\s*)+", re.I)
PARENTHESES_RX = re.compile(r"\([^()]*\)")
NOTE_RX = re.compile(r"\s*(?:\(.*|,.*|;.*|\s-\s.*|\s–\s.*|\s+(?:or|ou)\s+.*)$")


def ingredient_name(rest: str) -> str:
    """Extracts the ingredient name from what follows the quantity and unit, dropping articles, sizes, trailing notes and alternatives ("x or y" keeps x)."""
    name = NOTE_RX.sub("", LEADING_RX.sub("", PARENTHESES_RX.sub(" ", rest))).strip(" .,:;*")
    return re.sub(r"\s+", " ", name.replace("’", "'")).lower()
