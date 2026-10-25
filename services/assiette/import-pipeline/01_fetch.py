"""Step 1: downloads the FR/EN Wikibooks dumps, the FR↔EN Wikidata sitelink pairs and the ANSES Ciqual table into data/01_fetch/."""

import json
import re
import sys
import urllib.parse
from pathlib import Path

sys.path[:0] = [str(Path(__file__).parent / "utils"), str(Path(__file__).parent / "utils/01_fetch")]

from config import DATA_DIR
from http_get import http_get
from write_json import write_json

DUMPS = {
    "frwikibooks.xml.bz2": "https://dumps.wikimedia.org/frwikibooks/latest/frwikibooks-latest-pages-articles.xml.bz2",
    "enwikibooks.xml.bz2": "https://dumps.wikimedia.org/enwikibooks/latest/enwikibooks-latest-pages-articles.xml.bz2",
}
CIQUAL_DATASET = "https://entrepot.recherche.data.gouv.fr/api/datasets/:persistentId/?persistentId=doi:10.57745/RDMHWY"
CIQUAL_FILE_URL = "https://entrepot.recherche.data.gouv.fr/api/access/datafile/{id}"
CIQUAL_FILE_RX = re.compile(r"^(alim|const|compo)_\d{4}_\d{2}_\d{2}\.xml$")
SITELINKS_QUERY = """SELECT ?fr ?en WHERE {
  ?frl schema:about ?item; schema:isPartOf <https://fr.wikibooks.org/>; schema:name ?fr.
  FILTER(STRSTARTS(?fr, "Livre de cuisine/"))
  ?enl schema:about ?item; schema:isPartOf <https://en.wikibooks.org/>; schema:name ?en.
}"""


def main() -> None:
    target_dir = DATA_DIR / "01_fetch"
    target_dir.mkdir(parents=True, exist_ok=True)
    for name, url in DUMPS.items():
        target = target_dir / name
        if target.exists():
            print(f"{name} already present, skipped (delete it to refresh)")
            continue
        target.write_bytes(http_get(url))
        print(f"{name} downloaded ({target.stat().st_size // 1_000_000} MB)")
    for entry in json.loads(http_get(CIQUAL_DATASET))["data"]["latestVersion"]["files"]:
        data_file = entry["dataFile"]
        match = CIQUAL_FILE_RX.match(data_file["filename"])
        if not match:
            continue
        target = target_dir / f"ciqual_{match.group(1)}.xml"
        if target.exists():
            print(f"{target.name} already present, skipped (delete it to refresh)")
            continue
        target.write_bytes(http_get(CIQUAL_FILE_URL.format(id=data_file["id"])))
        print(f"{target.name} downloaded from {data_file['filename']} ({target.stat().st_size // 1_000_000} MB)")
    query = urllib.parse.urlencode({"query": SITELINKS_QUERY})
    bindings = json.loads(http_get(f"https://query.wikidata.org/sparql?{query}", accept="application/sparql-results+json"))["results"]["bindings"]
    pairs = [{"fr": b["fr"]["value"], "en": b["en"]["value"]} for b in bindings]
    write_json("01_fetch/wikidata_pairs.json", pairs, indent=1)
    print(f"wikidata FR↔EN pairs: {len(pairs)}")


if __name__ == "__main__":
    main()
