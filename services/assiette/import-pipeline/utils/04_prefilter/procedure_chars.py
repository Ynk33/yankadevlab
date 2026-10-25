import re

PROCEDURE_HEADING_RX = re.compile(
    r"^(=+)\s*(pr[ée]parations?|[ée]tapes?|recettes?|r[ée]alisations?|instructions?|procedures?|methods?|directions?|steps?)\b[^=]*=+\s*$",
    re.I | re.M,
)
SUB_HEADING_RX = re.compile(r"^(=+)[^=].*?=+\s*$")
NOISE_RX = re.compile(r"\[\[(?:Image|Fichier|File):[^\]]*\]\]|\{\{[^{}]*\}\}|[*#:\s]", re.I)


def procedure_chars(text: str) -> int:
    """Counts the meaningful characters under the procedure heading, sub-sections included, up to the next heading of the same level."""
    heading = PROCEDURE_HEADING_RX.search(text)
    if not heading:
        return 0
    level = len(heading.group(1))
    body = []
    for line in text[heading.end():].splitlines():
        sub = SUB_HEADING_RX.match(line)
        if sub and len(sub.group(1)) <= level:
            break
        if not sub:
            body.append(line)
    return len(NOISE_RX.sub("", "\n".join(body)))
