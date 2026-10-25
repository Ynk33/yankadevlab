import json
from pathlib import Path

from config import CATALOG_PATH


def system_prompt() -> str:
    """Instructions from prompt.md followed by the current catalog vocabulary (name | aisle | English name)."""
    catalog = json.loads(CATALOG_PATH.read_text())
    english = catalog["en"]["ing"]
    vocabulary = "\n".join(f"{name} | {rayon} | {english.get(name, '')}" for name, rayon in sorted(catalog["rayons"].items()))
    instructions = (Path(__file__).parent / "prompt.md").read_text()
    return f"{instructions}\n## Current vocabulary (name_fr | rayon | name_en)\n{vocabulary}\n"
