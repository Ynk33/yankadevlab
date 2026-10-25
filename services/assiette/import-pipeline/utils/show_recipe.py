"""CLI: prints a recipe as the pipeline sees it. Usage: uv run python utils/show_recipe.py "<page title or part of it>"."""

import sys

from load_pages import load_pages
from recipe_sections import recipe_sections


def show_recipe(query: str) -> None:
    """Prints the cleaned sections of every page whose title contains `query`."""
    pages = load_pages()
    matches = [t for t in pages if query.lower() in t.lower()]
    if not matches:
        print(f"no page matches {query!r}")
    for title in matches[:5]:
        print(f"##### {title}\n{recipe_sections(pages[title])}\n")
    if len(matches) > 5:
        print(f"... {len(matches) - 5} more matches")


if __name__ == "__main__":
    show_recipe(" ".join(sys.argv[1:]))
