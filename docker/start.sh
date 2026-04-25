#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  echo "Error: OPENAI_API_KEY is required." >&2
  exit 1
fi

BACKEND_PORT="${PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"

export PORT="$BACKEND_PORT"

cleanup() {
  if [[ -n "${BACKEND_PID:-}" ]]; then
    kill "$BACKEND_PID" 2>/dev/null || true
  fi
  if [[ -n "${FRONTEND_PID:-}" ]]; then
    kill "$FRONTEND_PID" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

/app/bin/server &
BACKEND_PID=$!

cd /app/frontend
npm run start -- --hostname 0.0.0.0 --port "$FRONTEND_PORT" &
FRONTEND_PID=$!

wait -n "$BACKEND_PID" "$FRONTEND_PID"
