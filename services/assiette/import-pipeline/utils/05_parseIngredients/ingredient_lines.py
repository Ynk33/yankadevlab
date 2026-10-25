import re

SECTION_RX = re.compile(r"^(=+)\s*(.*?)\s*=+\s*$")
INGREDIENTS_HEADING_RX = re.compile(r"ingr", re.I)
PROCEDURE_HEADING_RX = re.compile(r"(pr[ée]p|[ée]tape|recette|r[ée]alis|instr|proced|method|direct|steps|cuisson)", re.I)


def ingredient_lines(text: str) -> list[str]:
    """Returns the raw bullet lines of the ingredient section(s), sub-sections included, bullets stripped."""
    lines, on, level = [], False, 0
    for line in text.splitlines():
        heading = SECTION_RX.match(line)
        if heading and on and len(heading.group(1)) > level and PROCEDURE_HEADING_RX.match(heading.group(2)):
            on = False
            continue
        if heading and not (on and len(heading.group(1)) > level):
            on = bool(INGREDIENTS_HEADING_RX.match(heading.group(2)))
            level = len(heading.group(1))
            continue
        if on and line.strip().startswith(("*", "#")):
            lines.append(line.strip().lstrip("*#: ").strip())
    return [l for l in lines if l]
