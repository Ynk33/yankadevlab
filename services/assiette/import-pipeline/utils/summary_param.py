import re

SUMMARY_RX = re.compile(r"\{\{\s*recipe summary(.*?)\}\}", re.I | re.S)


def summary_param(text: str, key: str) -> str:
    """Returns a field of the EN {{recipe summary}} template, or an empty string."""
    summary = SUMMARY_RX.search(text)
    if not summary:
        return ""
    field = re.search(r"\|\s*" + re.escape(key) + r"\s*=\s*([^|}\n]*)", summary.group(1), re.I)
    return field.group(1).strip() if field else ""
