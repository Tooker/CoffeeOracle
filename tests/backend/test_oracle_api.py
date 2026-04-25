"""Integration-style tests for the Go backend API using the real HTTP server.

To run these tests:

1. Ensure the Go backend is running locally (e.g. `PORT=8080 OPENAI_API_KEY=... make -C backend run`).
2. Export `COFFEE_ORACLE_BASE_URL=http://localhost:8080` if you changed the port.
3. Execute `pytest tests/backend -k oracle_api`.

The tests stream real responses, so they require an active OpenAI API key
configured on the running backend.
"""

from __future__ import annotations

import os
from pathlib import Path
from typing import Iterator

import pytest
import requests

BASE_URL = os.getenv("COFFEE_ORACLE_BASE_URL", "http://localhost:8080")
IMAGE_PATH = (
    Path(__file__).resolve().parents[2]
    / "uploaded_files"
    / "WhatsApp Image 2026-01-29 at 18.11.52.jpeg"
).resolve()


def _require_backend() -> None:
    try:
        resp = requests.get(f"{BASE_URL}/healthz", timeout=5)
        resp.raise_for_status()
    except Exception as exc:  # pragma: no cover - network guard
        pytest.skip(f"Backend not reachable at {BASE_URL}: {exc}")


def _iter_sse_chunks(response: requests.Response) -> Iterator[tuple[str, str]]:
    event = ""
    for raw_line in response.iter_lines(decode_unicode=True):
        if raw_line is None:
            continue
        line = raw_line.strip()
        if not line:
            continue
        if line.startswith("event:"):
            event = line.split(":", 1)[1].strip()
            continue
        if line.startswith("data:"):
            data_value = line.split(":", 1)[1].strip()
            yield event or "message", data_value


@pytest.fixture(scope="module")
def sample_image() -> tuple[bytes, str]:
    if not IMAGE_PATH.exists():
        pytest.skip(f"Sample image missing at {IMAGE_PATH}")
    return IMAGE_PATH.read_bytes(), IMAGE_PATH.name


def test_healthz_endpoint() -> None:
    _require_backend()
    resp = requests.get(f"{BASE_URL}/healthz", timeout=5)
    assert resp.status_code == 200
    payload = resp.json()
    assert payload.get("status") == "ok"


def test_oracle_streams_text(sample_image: tuple[bytes, str]) -> None:
    _require_backend()
    image_bytes, filename = sample_image

    data = {"name": "PyTest Runner", "creativity": "7"}
    files = {"file": (filename, image_bytes, "image/jpeg")}

    with requests.post(
        f"{BASE_URL}/api/oracle",
        data=data,
        files=files,
        stream=True,
        timeout=120,
        headers={"Accept": "text/event-stream"},
    ) as resp:
        assert resp.status_code == 200, resp.text
        chunks: list[str] = []
        for event, payload in _iter_sse_chunks(resp):
            if event == "response.output_text.delta":
                chunks.append(payload)
            elif event == "response.error":
                pytest.fail(f"Backend/OpenAI error: {payload}")
            elif event in {"response.completed", "complete"}:
                break

        assert chunks, "Expected at least one streamed chunk from oracle"


def test_oracle_rejects_missing_file() -> None:
    _require_backend()
    resp = requests.post(
        f"{BASE_URL}/api/oracle",
        data={"name": "NoFile", "creativity": "5"},
        timeout=10,
    )
    assert resp.status_code == 400
    payload = resp.json()
    assert "image" in payload.get("error", "").lower()
