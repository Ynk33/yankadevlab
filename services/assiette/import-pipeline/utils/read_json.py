import json

from config import DATA_DIR


def read_json(relative_path: str):
    """Loads a JSON file from the data directory."""
    return json.loads((DATA_DIR / relative_path).read_text())
