import re

EN_HEAD_NOUNS = (
    "breast|breasts|thigh|thighs|wing|wings|leg|legs|drumsticks?|broth|stock|paste|powder|sauce|oil|juice|milk|cream|flour"
    "|seeds?|leaves|leaf|zest|vinegar|butter|cheese|meal|flakes|noodles|bread|fillets?|steak|mince|sausages?|syrup|water"
)
FR_LINK = r"(?:de la\b|de l'|de\b|d'|du\b|des\b|à la\b|à l'|au\b|aux\b)\s*"


def compound_key(key: str, name: str, lang: str) -> str:
    """Keeps the full name when the wiki markup only tagged part of a compound product ("fumet de {{i|poisson}}",
    "[[Chicken|chicken]] broth"), so the vocabulary step sees "fumet de poisson" rather than "poisson"."""
    if not key or not name or key == name or key not in name:
        return key
    k = re.escape(key)
    if lang == "fr":
        before = re.search(rf"\w\s+{FR_LINK}{k}", name)
        after = re.search(rf"{k}s?\s+{FR_LINK}\w", name)
        return name if before or after else key
    head = re.search(rf"\b{k}s?\s+({EN_HEAD_NOUNS})\b", name, re.I)
    return f"{key} {head.group(1).lower()}" if head else key
