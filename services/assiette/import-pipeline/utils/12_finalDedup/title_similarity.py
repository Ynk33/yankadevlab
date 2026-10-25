import difflib
import re

STOP_WORDS = {"de", "du", "des", "la", "le", "les", "à", "au", "aux", "et", "en", "d", "l"}


def title_similarity(a: str, b: str) -> float:
    """Highest of the character ratio and the share of the shorter title's words found in the other title."""
    words_a = {w for w in re.findall(r"\w+", a.lower()) if w not in STOP_WORDS}
    words_b = {w for w in re.findall(r"\w+", b.lower()) if w not in STOP_WORDS}
    containment = len(words_a & words_b) / min(len(words_a), len(words_b)) if words_a and words_b else 0.0
    return max(difflib.SequenceMatcher(None, a.lower(), b.lower()).ratio(), containment)
