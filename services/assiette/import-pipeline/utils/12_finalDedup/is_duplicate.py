from jaccard import jaccard
from title_similarity import title_similarity

TITLE_SIMILARITY = 0.75
INGREDIENT_SIMILARITY = 0.5


def is_duplicate(a: tuple[str, str, set[str]], b: tuple[str, str, set[str]]) -> bool:
    """Two recipes, given as (French title, main protein category, main ingredients), are duplicates when they share
    the main protein, their titles are close and their main ingredients overlap."""
    return a[1] == b[1] and title_similarity(a[0], b[0]) >= TITLE_SIMILARITY and jaccard(a[2], b[2]) >= INGREDIENT_SIMILARITY
