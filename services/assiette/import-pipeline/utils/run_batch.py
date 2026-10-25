import hashlib
import json
import time

from anthropic.types.message_create_params import MessageCreateParamsNonStreaming
from anthropic.types.messages.batch_create_params import Request

from claude_client import claude_client
from config import DATA_DIR

POLL_SECONDS = 30


def run_batch(step: str, requests: dict[str, dict], wait_minutes: float = 5) -> dict[str, dict] | None:
    """Runs `requests` ({custom_id: params}) through the Message Batches API with a per-request cache.

    Each request is fingerprinted (id + params). Cached answers are reused; only new or changed requests are submitted,
    as one batch that a rerun resumes. Waits up to `wait_minutes`; returns {custom_id: message dict} for every request
    answered so far once nothing is pending, or None while a batch is still running (the step then exits with code 2; rerun it later to collect).
    State lives in data/<step>/cache.jsonl and pending_batch.json.
    """
    folder = DATA_DIR / step
    folder.mkdir(parents=True, exist_ok=True)
    cache_path, pending_path = folder / "cache.jsonl", folder / "pending_batch.json"
    digests = {cid: hashlib.sha256(json.dumps([cid, params], sort_keys=True, ensure_ascii=False).encode()).hexdigest() for cid, params in requests.items()}
    cache = {row["digest"]: row for row in map(json.loads, cache_path.read_text().splitlines())} if cache_path.exists() else {}
    pending = {cid: requests[cid] for cid, digest in digests.items() if digest not in cache}

    if pending:
        client = claude_client()
        state = json.loads(pending_path.read_text()) if pending_path.exists() else None
        if not state or sorted(state["digests"]) != sorted(digests[cid] for cid in pending):
            batch = client.messages.batches.create(
                requests=[Request(custom_id=cid, params=MessageCreateParamsNonStreaming(**params)) for cid, params in pending.items()]
            )
            state = {"id": batch.id, "digests": sorted(digests[cid] for cid in pending)}
            pending_path.write_text(json.dumps(state))
            print(f"batch {batch.id} submitted ({len(pending)} new or changed requests, {len(requests) - len(pending)} cached)")
        deadline = time.time() + wait_minutes * 60
        while True:
            batch = client.messages.batches.retrieve(state["id"])
            if batch.processing_status == "ended":
                break
            if time.time() > deadline:
                print(f"batch {state['id']} still {batch.processing_status} ({batch.request_counts.succeeded}/{len(pending)} done): rerun the step later to collect")
                return None
            print(f"  waiting... {batch.request_counts.succeeded}/{len(pending)} done", flush=True)
            time.sleep(POLL_SECONDS)
        with cache_path.open("a") as f:
            for result in client.messages.batches.results(state["id"]):
                if result.result.type != "succeeded":
                    print(f"  {result.custom_id}: {result.result.type}")
                    continue
                row = {"custom_id": result.custom_id, "digest": digests[result.custom_id], "message": result.result.message.model_dump(mode="json")}
                cache[row["digest"]] = row
                f.write(json.dumps(row, ensure_ascii=False) + "\n")
        pending_path.unlink()
        print(f"batch {state['id']} ended")
    return {cid: cache[digest]["message"] for cid, digest in digests.items() if digest in cache}
