import re

from summary_param import summary_param

RANGE = r"(\d+)(?:\s*(?:à|-|–|to|ou|or)\s*(\d+))?"
FOR_RX = re.compile(rf"\b(?:pour|for)\s+{RANGE}\s*(?:personnes?|pers\b|parts?|couverts|portions?|people|persons|servings)", re.I)
SERVES_RX = re.compile(rf"\b(?:serves|servings?|makes|yields?)\s*:?\s*{RANGE}", re.I)
SUMMARY_RX = re.compile(r"(\d+)(?:\s*(?:-|–|to)\s*(\d+))?")
MAX_SERVINGS = 20


def explicit_servings(text: str) -> float | None:
    """Reads the servings a page states (EN summary `servings`, "Pour 4 personnes", "Serves 4-6" as 5), or None."""
    summary = summary_param(text, "servings")
    for source, rx in ((summary, SUMMARY_RX), (text, FOR_RX), (text, SERVES_RX)):
        match = rx.search(source) if source else None
        if match:
            low = int(match.group(1))
            value = (low + int(match.group(2))) / 2 if match.group(2) else low
            if 1 <= value <= MAX_SERVINGS:
                return value
    return None
