import re
import unicodedata

from short_title import short_title

NOISE_WORDS_RX = re.compile(
    r"\b(ii|iii|iv|2|3|easy|simple|quick|basic|classic|traditional|my|best|homemade|recette|facile|rapide|maison"
    r"|de|du|des|la|le|les|a|au|aux|the|with|and|et|of)\b"
)


def normalize_title(title: str) -> str:
    """Reduces a title to lowercase ASCII words, without parentheses, variant markers or stop words."""
    text = re.sub(r"\(.*?\)", "", short_title(title))
    text = unicodedata.normalize("NFKD", text).encode("ascii", "ignore").decode().lower()
    text = NOISE_WORDS_RX.sub(" ", text)
    return re.sub(r"[^a-z]+", " ", text).strip()
