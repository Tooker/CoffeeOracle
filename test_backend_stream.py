#!/usr/bin/env python3
from __future__ import annotations

import argparse
import mimetypes
import sys
from pathlib import Path

import requests


DEFAULT_IMAGE = Path("uploaded_files/WhatsApp Image 2026-01-29 at 18.11.52.jpeg")


def iter_sse_events(response: requests.Response):
    event_type = "message"
    data_lines: list[str] = []

    for raw_line in response.iter_lines(decode_unicode=True):
        if raw_line is None:
            continue

        line = raw_line.rstrip("\r")

        if line == "":
            if data_lines or event_type != "message":
                yield event_type, "\n".join(data_lines)
            event_type = "message"
            data_lines = []
            continue

        if line.startswith(":"):
            continue

        if line.startswith("event:"):
            event_type = line[6:].strip() or "message"
        elif line.startswith("data:"):
            value = line[5:]
            if value.startswith(" "):
                value = value[1:]
            data_lines.append(value)

    if data_lines or event_type != "message":
        yield event_type, "\n".join(data_lines)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Test the oracle backend stream and print full combined output.",
    )
    parser.add_argument(
        "--api-base",
        default="http://localhost:8080",
        help="Backend base URL (default: %(default)s)",
    )
    parser.add_argument(
        "--name",
        default="Tobias",
        help="Name sent to backend (default: %(default)s)",
    )
    parser.add_argument(
        "--creativity",
        type=int,
        default=5,
        help="Esoterik-Stufe 0..10 (default: %(default)s)",
    )
    parser.add_argument(
        "--image",
        type=Path,
        default=DEFAULT_IMAGE,
        help=f"Path to image file (default: {DEFAULT_IMAGE})",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Print each SSE event type while streaming.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()

    if not 0 <= args.creativity <= 10:
        print("Error: --creativity must be between 0 and 10", file=sys.stderr)
        return 2

    if not args.image.exists() or not args.image.is_file():
        print(f"Error: image not found: {args.image}", file=sys.stderr)
        return 2

    endpoint = args.api_base.rstrip("/") + "/api/oracle"
    mime = mimetypes.guess_type(str(args.image))[0] or "application/octet-stream"

    print(f"POST {endpoint}")
    print(f"Image: {args.image}")
    print()

    try:
        with args.image.open("rb") as f:
            response = requests.post(
                endpoint,
                data={"name": args.name, "creativity": str(args.creativity)},
                files={"file": (args.image.name, f, mime)},
                stream=True,
                timeout=(10, 300),
            )

        response.raise_for_status()
    except requests.RequestException as exc:
        print(f"Request failed: {exc}", file=sys.stderr)
        return 1

    chunks: list[str] = []
    completed = False

    for event_type, data in iter_sse_events(response):
        if args.verbose:
            print(f"[event] {event_type}")

        if event_type == "response.output_text.delta" or event_type == "message":
            chunks.append(data)
        elif event_type in {"response.error", "error"}:
            print(f"Stream error: {data}", file=sys.stderr)
            return 1
        elif event_type in {"response.completed", "complete"}:
            completed = True
            break

    full_text = "".join(chunks)

    print("=" * 80)
    print("COMPLETE ORACLE OUTPUT")
    print("=" * 80)
    print(full_text if full_text else "<empty output>")
    print("=" * 80)
    print(f"chunks={len(chunks)} completed={completed}")

    return 0 if full_text else 1


if __name__ == "__main__":
    raise SystemExit(main())
