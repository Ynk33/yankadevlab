import xml.etree.ElementTree as ET

from config import DATA_DIR
from parse_value import parse_value

NUTRIENTS = {
    "328": "kcal", "25000": "proteines", "34100": "fibres", "40000": "lipides", "55100": "vitamine_c", "10260": "fer",
    "10200": "calcium", "10120": "magnesium", "10190": "potassium", "10300": "zinc", "10530": "iode", "10340": "selenium",
    "56700": "folates", "56600": "vitamine_b12", "52100": "vitamine_d", "54101": "vitamine_k", "51330": "beta_carotene",
    "41833": "ala", "42053": "epa", "42263": "dha",
}


def load_ciqual() -> dict[str, dict]:
    """Reads the Ciqual XML files into {alim_code: {name_fr, name_en, nutrients: {nutrient: value per 100 g}}}."""
    foods = {}
    for alim in ET.parse(DATA_DIR / "01_fetch/ciqual_alim.xml").getroot():
        code = alim.findtext("alim_code").strip()
        foods[code] = {"name_fr": alim.findtext("alim_nom_fr").strip(), "name_en": (alim.findtext("alim_nom_eng") or "").strip(), "nutrients": {}}
    for _, element in ET.iterparse(DATA_DIR / "01_fetch/ciqual_compo.xml", events=("end",)):
        if element.tag != "COMPO":
            continue
        const = element.findtext("const_code").strip()
        code = element.findtext("alim_code").strip()
        if const in NUTRIENTS and code in foods:
            foods[code]["nutrients"][NUTRIENTS[const]] = parse_value(element.findtext("teneur"))
        element.clear()
    return foods
