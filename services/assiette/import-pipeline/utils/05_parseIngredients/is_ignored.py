import re

IGNORED_NAMES_RX = re.compile(
    r"^(sel|gros sel|fleur de sel|poivre|poivre noir|poivre du moulin|sel et poivre|sel, poivre|eau|eau froide|eau chaude"
    r"|glaçons?|salt|kosher salt|sea salt|table salt|pepper|black pepper|ground black pepper|salt and pepper|water|cold water"
    r"|hot water|warm water|boiling water|ice|ice cubes)$",
    re.I,
)
OPTIONAL_RX = re.compile(
    r"\b(to taste|as needed|as required|optional|for garnish|to garnish|garnish|au goût|selon (le|votre) goût|facultati(f|ve)"
    r"|pour (la )?décor(er|ation)|pour garnir)\b",
    re.I,
)


def is_ignored(name: str, text: str) -> bool:
    """True for salt, pepper, water and optional or to-taste items, which the catalog leaves out."""
    return bool(IGNORED_NAMES_RX.match(name.strip())) or bool(OPTIONAL_RX.search(text))
