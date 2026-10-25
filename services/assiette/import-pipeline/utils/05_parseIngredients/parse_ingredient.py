import re

from clean_wikitext import clean_wikitext
from compound_key import compound_key
from expand_convert import expand_convert
from ingredient_key import ingredient_key
from ingredient_name import ingredient_name
from is_ignored import is_ignored
from match_unit import match_unit
from metric_in_parentheses import metric_in_parentheses
from parse_quantity import parse_quantity

MAX_NAME_WORDS = 6
HEADER_RX = re.compile(r"(:\s*$|^pour (la|le|les|l['’])\s.*:?$|^for the\s.*:?$)", re.I)
SERVINGS_RX = re.compile(r"^(pour|for|serves)\s+\d+(\s*(à|-|to)\s*\d+)?\s*(personnes?|pers\.?|people|persons|servings|portions)?\s*:?\s*$", re.I)


def parse_ingredient(segment: str, lang: str) -> dict:
    """Parses one raw ingredient segment into {raw, key, name, qty, unit, status}; status is parsed, unquantified, ignored, header or failed."""
    text = re.sub(r"\s+", " ", clean_wikitext(expand_convert(segment)).replace("’", "'")).strip()
    key = ingredient_key(segment).replace("’", "'").lower()
    row = {"raw": segment, "text": text, "key": key, "name": "", "qty": None, "unit": None, "status": "failed"}
    if SERVINGS_RX.match(text) or (HEADER_RX.search(text) and not re.search(r"\d", text)):
        row["status"] = "header"
        return row
    qty, rest = parse_quantity(text)
    unit, factor = None, 1
    if qty is not None:
        unit, factor, rest = match_unit(rest)
        metric = metric_in_parentheses(text)
        if metric and unit not in ("g", "ml"):
            qty, factor, unit = metric[0], 1, metric[1]
        qty = round(qty * factor, 2)
    row.update(name=ingredient_name(rest), qty=qty, unit=unit)
    row["key"] = compound_key(row["key"], row["name"], lang)
    if is_ignored(row["key"] or row["name"], text):
        row["status"] = "ignored"
    elif not (row["key"] or row["name"]) or (not row["key"] and len(row["name"].split()) > MAX_NAME_WORDS):
        row["status"] = "failed"
    else:
        row["status"] = "parsed" if qty else "unquantified"
    return row
