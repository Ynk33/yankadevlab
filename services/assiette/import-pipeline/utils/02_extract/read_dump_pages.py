import bz2
import xml.etree.ElementTree as ET
from pathlib import Path
from typing import Callable


def read_dump_pages(dump: Path, keep: Callable[[str, str], bool]) -> list[dict]:
    """Streams a MediaWiki XML dump and returns the non-redirect pages for which keep(namespace, title) is true."""
    pages = []
    for _, element in ET.iterparse(bz2.open(dump), events=("end",)):
        if not element.tag.endswith("}page"):
            continue
        ns = element.tag[:-4]
        title = element.findtext(ns + "title") or ""
        if element.find(ns + "redirect") is None and keep(element.findtext(ns + "ns") or "", title):
            pages.append({"title": title, "text": element.findtext(f"{ns}revision/{ns}text") or ""})
        element.clear()
    return pages
