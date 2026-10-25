from read_json import read_json


def load_pages() -> dict[str, str]:
    """Returns every extracted FR and EN page as {title: wikitext}."""
    pages = {}
    for lang in ("fr", "en"):
        pages.update({p["title"]: p["text"] for p in read_json(f"02_extract/pages_{lang}.json")})
    return pages
