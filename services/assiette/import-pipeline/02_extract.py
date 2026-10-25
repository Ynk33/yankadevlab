"""Step 2: extracts the cookbook pages from the dumps and keeps the savoury recipe candidates.

Writes data/02_extract/pages_fr.json, pages_en.json (every cookbook page) and candidates.json ({fr: [titles], en: [titles]}).
"""

import collections
import sys
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/02_extract")]

from config import DATA_DIR, EN_NAMESPACE, FR_PREFIX
from en_candidate_stage import en_candidate_stage
from fr_candidate_stage import fr_candidate_stage
from read_dump_pages import read_dump_pages
from write_json import write_json


def main() -> None:
    fr_pages = read_dump_pages(DATA_DIR / "01_fetch/frwikibooks.xml.bz2", lambda ns, title: ns == "0" and title.startswith(FR_PREFIX))
    en_pages = read_dump_pages(DATA_DIR / "01_fetch/enwikibooks.xml.bz2", lambda ns, title: ns == EN_NAMESPACE)
    write_json("02_extract/pages_fr.json", fr_pages)
    write_json("02_extract/pages_en.json", en_pages)

    fr_stages = {p["title"]: fr_candidate_stage(p["title"], p["text"]) for p in fr_pages}
    en_stages = {p["title"]: en_candidate_stage(p["text"]) for p in en_pages}
    candidates = {
        "fr": [t for t, s in fr_stages.items() if s == "ok"],
        "en": [t for t, s in en_stages.items() if s == "ok"],
    }
    write_json("02_extract/candidates.json", candidates, indent=1)

    for lang, pages, stages in (("fr", fr_pages, fr_stages), ("en", en_pages, en_stages)):
        failed = collections.Counter(s for s in stages.values() if s != "ok")
        print(f"{lang}: {len(pages)} pages -> {len(candidates[lang])} candidates | rejected by stage {dict(failed)}")


if __name__ == "__main__":
    main()
