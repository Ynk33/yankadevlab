from name_tokens import name_tokens

RAW_WORDS = {"cru", "crue", "raw", "frai", "fraiche", "fresh"}
HEAD_WEIGHTS = (1.0, 0.4)
RAW_BONUS = 0.3
LENGTH_PENALTY = 0.02


def candidate_foods(product_name: str, product_name_en: str, foods: dict[str, dict], tokens: dict[str, tuple[list[str], list[str]]], limit: int = 15) -> list[str]:
    """Ranks Ciqual foods for a product: same head noun first (Ciqual names start with it; French counts more than
    English), then share of the product's
    words found (French or English), then raw forms, then shorter (more generic) names."""
    wanted = (name_tokens(product_name), name_tokens(product_name_en))
    if not any(wanted):
        return []
    scored = []
    for code, food in tokens.items():
        coverage = max(len(set(w) & set(f)) / len(set(w)) for w, f in zip(wanted, food) if w) if any(wanted) else 0
        if not coverage:
            continue
        head = sum(weight for weight, w, f in zip(HEAD_WEIGHTS, wanted, food) if w and f and w[0] == f[0])
        raw = bool(RAW_WORDS & set(food[0] + food[1]))
        score = head + coverage + RAW_BONUS * raw - LENGTH_PENALTY * len(food[0])
        scored.append((score, code))
    return [code for _, code in sorted(scored, reverse=True)[:limit]]
