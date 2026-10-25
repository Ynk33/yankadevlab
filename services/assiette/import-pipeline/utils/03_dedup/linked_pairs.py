import re

FR_TO_EN_RX = re.compile(r"\[\[\s*en\s*:\s*(Cookbook:[^\]|]+)", re.I)
EN_TO_FR_RX = re.compile(r"\[\[\s*fr\s*:\s*(Livre de cuisine/[^\]|]+)", re.I)


def linked_pairs(wikidata_pairs: list[dict], fr_candidates: list[str], pages: dict[str, str], fr_titles: set[str]) -> set[tuple[str, str]]:
    """Collects (FR title, EN title) pairs from Wikidata sitelinks and inline interwiki links in both directions."""
    pairs = {(p["fr"], p["en"]) for p in wikidata_pairs}
    for title in fr_candidates:
        pairs.update((title, en.strip()) for en in FR_TO_EN_RX.findall(pages[title]))
    for title, text in pages.items():
        if title.startswith("Cookbook:"):
            pairs.update((fr.strip(), title) for fr in EN_TO_FR_RX.findall(text) if fr.strip() in fr_titles)
    return pairs
