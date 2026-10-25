import collections


def highlight_translations(catalog: dict) -> dict[str, str]:
    """FR→EN highlight labels as already used by the catalog (most frequent English label per French one)."""
    pairs: dict[str, collections.Counter] = collections.defaultdict(collections.Counter)
    for raw in catalog["recipes"]:
        english = catalog["en"]["rec"].get(raw[0])
        if english and len(english[1]) == len(raw[8]):
            for fr, en in zip(raw[8], english[1]):
                pairs[fr][en] += 1
    return {fr: counts.most_common(1)[0][0] for fr, counts in pairs.items()}
