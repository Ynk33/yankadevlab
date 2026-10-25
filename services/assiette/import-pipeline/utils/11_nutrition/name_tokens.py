import re
import unicodedata

STOP_WORDS = {"de", "du", "des", "la", "le", "les", "et", "en", "au", "aux", "a", "of", "the", "and", "with", "or", "ou", "d", "l"}


def name_tokens(name: str) -> list[str]:
    """Ordered lowercase ASCII word stems of a food name (œ/æ ligatures expanded, accents, plural s/x and stop words removed)."""
    text = unicodedata.normalize("NFKD", name.replace("œ", "oe").replace("Œ", "Oe").replace("æ", "ae")).encode("ascii", "ignore").decode().lower()
    words = re.findall(r"[a-z]+", text)
    return [w[:-1] if len(w) > 3 and w[-1] in "sx" else w for w in words if w not in STOP_WORDS]
