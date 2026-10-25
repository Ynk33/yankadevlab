import json

from config import DATA_DIR


def write_json(relative_path: str, payload, indent: int | None = None) -> None:
    """Writes a JSON file into the data directory, creating parent folders."""
    path = DATA_DIR / relative_path
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, ensure_ascii=False, indent=indent))
