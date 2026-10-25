from recipe_sections import recipe_sections
from short_title import short_title

MAX_CHARS = 20000


def user_message(recipe: dict, page_text: str) -> str:
    """Renders one recipe for the rewriter: source title and procedure, frozen ingredients for 2, missing-side flags."""
    source = recipe_sections(page_text)
    if len(source) > MAX_CHARS:
        raise ValueError(f"{recipe['title']}: {len(source)} chars, above {MAX_CHARS}; refusing to truncate silently")
    ingredients = "\n".join(f"- {i['qty']:g} {i['unit']} {i['name']}" for i in recipe["ingredients"])
    flags = [f for f, on in (("missing a starch", recipe["needs_starch"]), ("missing vegetables", recipe["needs_vegetables"])) if on]
    return (
        f"Source title: {short_title(recipe['title'])}\nSource language: {recipe['lang']}\n"
        f"Flags: {', '.join(flags) or 'none'}\n\nFinal ingredients for 2 people:\n{ingredients}\n\nSource recipe:\n{source}"
    )
