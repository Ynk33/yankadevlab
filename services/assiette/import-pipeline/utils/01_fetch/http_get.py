import urllib.request

from config import USER_AGENT


def http_get(url: str, accept: str = "*/*") -> bytes:
    """GETs a URL with the pipeline user agent (required by Wikimedia)."""
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT, "Accept": accept})
    with urllib.request.urlopen(request, timeout=600) as response:
        return response.read()
