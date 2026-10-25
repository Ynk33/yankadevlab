import re

from clean_wikitext import clean_wikitext

SECTION_RX = re.compile(r"^(=+)\s*(.*?)\s*=+\s*$")
KEEP_RX = re.compile(r"(ingr|pr[ée]p|[ée]tape|recette|r[ée]alis|instr|proced|method|direct|steps|cuisson)", re.I)


def recipe_sections(text: str) -> str:
    """Returns the cleaned ingredient and procedure sections of a recipe page, sub-sections included."""
    kept, on, level = [], False, 0
    for line in text.splitlines():
        heading = SECTION_RX.match(line)
        if heading and not (on and len(heading.group(1)) > level):
            on = bool(KEEP_RX.match(heading.group(2)))
            level = len(heading.group(1))
        if on:
            kept.append(line)
    return clean_wikitext("\n".join(kept))
