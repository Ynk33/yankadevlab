import re
import unicodedata


def slug(text: str) -> str:
    """Python port of web/src/lib/utils.ts slug(): lowercase, accents stripped, non-alphanumerics to '-', 80 chars max."""
    text = unicodedata.normalize("NFD", text.lower())
    text = re.sub(r"[̀-ͯ]", "", text)
    return re.sub(r"[^a-z0-9]+", "-", text).strip("-")[:80]
