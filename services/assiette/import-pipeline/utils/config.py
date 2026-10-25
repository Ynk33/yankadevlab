"""Pipeline-wide constants: paths, Wikimedia sources and the Assiette catalog vocabularies."""

from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DATA_DIR = ROOT / "data"
CATALOG_PATH = ROOT.parent / "web/src/data/catalog.json"
IMPORTED_PATH = ROOT.parent / "web/src/data/imported.json"
ENV_PATH = ROOT.parent / ".env"

USER_AGENT = "assiette-import-pipeline/0.1 (https://github.com/Ynk33/yankadevlab)"
FR_PREFIX = "Livre de cuisine/"
EN_NAMESPACE = "102"
WIKI_URLS = {"fr": "https://fr.wikibooks.org/wiki/", "en": "https://en.wikibooks.org/wiki/"}

MODEL = "claude-haiku-5-5"

RAYONS = ["Fruits & légumes", "Boucherie & poisson", "Crèmerie & œufs", "Boulangerie", "Épicerie", "Surgelés", "Herbes & épices"]
UNITS = ["g", "ml", "pièce", "botte", "c. à s.", "c. à c."]
PROTEINS = ["poisson", "fruits de mer", "volaille", "viande rouge", "porc", "légumineuses", "œufs", "tofu/tempeh", "fromage"]
MEALS = ["midi", "soir", "both"]
