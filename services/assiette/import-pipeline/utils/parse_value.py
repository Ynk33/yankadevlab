import re


def parse_value(text: str) -> float:
    """Reads a Ciqual content ("1,5", "< 0,5", "traces", "-") as a float; bounds and traces count as 0."""
    text = (text or "").strip().replace(",", ".")
    if not text or text in ("-", "traces"):
        return 0.0
    if text.startswith("<"):
        return 0.0
    match = re.search(r"\d+(?:\.\d+)?", text)
    return float(match.group(0)) if match else 0.0
