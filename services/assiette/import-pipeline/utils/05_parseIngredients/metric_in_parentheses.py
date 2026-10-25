import re

from match_unit import match_unit
from parse_quantity import parse_quantity

PARENTHESES_RX = re.compile(r"\(([^()]*)\)")


def metric_in_parentheses(text: str) -> tuple[float, str] | None:
    """Finds a metric amount written in parentheses ("1 cup (240 ml)", "1 can (400 g)") and returns (value, 'g'|'ml')."""
    for inner in PARENTHESES_RX.findall(text):
        value, rest = parse_quantity(inner.strip())
        if value is None:
            continue
        unit, factor, _ = match_unit(rest)
        if unit in ("g", "ml"):
            return value * factor, unit
    return None
