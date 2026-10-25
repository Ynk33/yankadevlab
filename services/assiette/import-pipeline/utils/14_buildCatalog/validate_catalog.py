from config import MEALS, PROTEINS, RAYONS, UNITS

MIN_STEPS, MAX_STEPS = 3, 5


def validate_catalog(catalog: dict) -> list[str]:
    """Lists every violation of what web/src/lib/catalog.ts expects: unique ids, known ingredients with an aisle and an
    English name, catalog units, meal and protein values, step count, positive numbers."""
    errors, seen = [], set()
    for raw in catalog["recipes"]:
        rid = raw[0]
        if len(raw) != 13:
            errors.append(f"{rid}: {len(raw)} fields")
            continue
        if rid in seen:
            errors.append(f"{rid}: duplicate id")
        seen.add(rid)
        if raw[2] not in MEALS:
            errors.append(f"{rid}: meal {raw[2]!r}")
        if raw[3] not in PROTEINS:
            errors.append(f"{rid}: protein {raw[3]!r}")
        if not MIN_STEPS <= len(raw[10]) <= MAX_STEPS and rid in catalog.get("_new", set()):
            errors.append(f"{rid}: {len(raw[10])} steps")
        if raw[4] <= 0:
            errors.append(f"{rid}: minutes {raw[4]}")
        for name, qty, unit in raw[9]:
            if catalog["rayons"].get(name) not in RAYONS:
                errors.append(f"{rid}: ingredient {name!r} has no aisle")
            if unit not in UNITS or qty <= 0:
                errors.append(f"{rid}: {qty} {unit} {name}")
        if rid not in catalog["en"]["rec"]:
            errors.append(f"{rid}: no English text")
    return errors
