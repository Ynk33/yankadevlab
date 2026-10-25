from explicit_categories import explicit_categories
from summary_param import summary_param


def page_categories(text: str) -> set[str]:
    """Explicit categories plus the `category` field of the EN recipe summary template."""
    cats = set(explicit_categories(text))
    summary_category = summary_param(text, "category")
    if summary_category:
        cats.add(summary_category)
    return cats
