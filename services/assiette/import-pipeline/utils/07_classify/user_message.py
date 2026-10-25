from explicit_categories import explicit_categories
from recipe_sections import recipe_sections
from short_title import short_title

MAX_CHARS = 20000


def user_message(title: str, lang: str, text: str) -> str:
    """Renders one recipe for the classifier: title, language, categories and cleaned ingredient + procedure sections."""
    body = recipe_sections(text)
    if len(body) > MAX_CHARS:
        raise ValueError(f"{title}: {len(body)} chars, above {MAX_CHARS}; refusing to truncate silently")
    categories = ", ".join(explicit_categories(text)) or "none"
    return f"Title: {short_title(title)}\nLanguage: {lang}\nCategories: {categories}\n\n{body}"
