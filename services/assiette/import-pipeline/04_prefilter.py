"""Step 4: rejects pool recipes that are obviously not a main dish, before any parsing or LLM call.

Writes data/04_prefilter/result.json ({kept: {fr, en}, rejected: [{title, lang, reason}]}).
"""

import collections
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/04_prefilter")]

from load_pages import load_pages
from read_json import read_json
from reject_reason import reject_reason
from write_json import write_json


def main() -> None:
    pool = read_json("03_dedup/pool.json")
    pages = load_pages()
    kept, rejected = {"fr": [], "en": []}, []
    for lang in ("fr", "en"):
        for title in pool[lang]:
            reason = reject_reason(title, pages[title])
            if reason:
                rejected.append({"title": title, "lang": lang, "reason": reason})
            else:
                kept[lang].append(title)
    write_json("04_prefilter/result.json", {"kept": kept, "rejected": rejected}, indent=1)

    total = len(pool["fr"]) + len(pool["en"])
    print(f"pool {total} -> kept {len(kept['fr'])} FR + {len(kept['en'])} EN = {len(kept['fr']) + len(kept['en'])} | rejected {len(rejected)}")
    print("by rule:", dict(collections.Counter(r["reason"].split(":")[0] for r in rejected)))
    print("top reasons:", collections.Counter(r["reason"] for r in rejected).most_common(10))


if __name__ == "__main__":
    main()
