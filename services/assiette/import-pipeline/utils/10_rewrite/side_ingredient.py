STARCH_SIDES = {
    "riz complet": (120, "g", 120), "pâtes complètes": (160, "g", 160), "quinoa": (120, "g", 120), "semoule complète": (120, "g", 120),
    "boulgour": (120, "g", 120), "pomme de terre": (400, "g", 400), "tranche de pain complet": (4, "pièce", 120), "pita complète": (2, "pièce", 120),
}
VEGETABLE_SIDES = {
    "salade verte": (0.5, "pièce", 150), "mâche": (100, "g", 100), "roquette": (60, "g", 60), "concombre": (1, "pièce", 300),
    "tomate": (2, "pièce", 240), "carotte": (2, "pièce", 200), "haricot vert": (250, "g", 250), "brocoli": (300, "g", 300),
    "pousse d'épinard": (150, "g", 150),
}
ROLES = {"starch": STARCH_SIDES, "vegetable": VEGETABLE_SIDES}


def side_ingredient(name: str, role: str, rayon: str) -> dict:
    """Builds the catalog ingredient (amount for 2 people) for a side chosen to complete the service."""
    qty, unit, grams = ROLES[role][name]
    return {"name": name, "qty": qty, "unit": unit, "grams": grams, "rayon": rayon, "role": role, "protein": "none",
            "processed": False, "hard_to_find": False, "defaulted": False, "side": True}
