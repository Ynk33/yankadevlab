import re

CONVERT_RX = re.compile(r"\{\{\s*convert\s*\|\s*([\d.,/½¼¾⅓⅔ ]+?)\s*\|\s*([^|}]+?)\s*(?:\|[^}]*)?\}\}", re.I)


def expand_convert(text: str) -> str:
    """Replaces {{convert|240|ml|...}} templates with their first value and unit ("240 ml")."""
    return CONVERT_RX.sub(lambda m: f"{m.group(1).strip()} {m.group(2).strip()}", text)
