def short_title(title: str) -> str:
    """Drops the 'Cookbook:' namespace and the 'Livre de cuisine/' path from a page title."""
    return title.split(":", 1)[-1].split("/")[-1]
