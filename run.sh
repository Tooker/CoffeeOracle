#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_ENV="$ROOT_DIR/backend/.env"

cleanup() {
  printf "\nShutting down services...\n"
  if [[ -n "${BACKEND_PID:-}" ]] && ps -p "$BACKEND_PID" >/dev/null 2>&1; then
    kill "$BACKEND_PID" 2>/dev/null || true
  fi
  if [[ -n "${FRONTEND_PID:-}" ]] && ps -p "$FRONTEND_PID" >/dev/null 2>&1; then
    kill "$FRONTEND_PID" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

if [[ -f "$BACKEND_ENV" ]]; then
  echo "Loading backend environment from $BACKEND_ENV"
  set -a
  # shellcheck disable=SC1090
  source "$BACKEND_ENV"
  set +a
else
  echo "Warning: $BACKEND_ENV not found. Ensure PORT and OPENAI_API_KEY are exported." >&2
fi

PORT=${PORT:-8080}
if [[ -z "${OPENAI_API_KEY:-}" ]]; then
  echo "Error: OPENAI_API_KEY is not set. Please set it in backend/.env or your shell." >&2
  exit 1
fi

echo "Starting CoffeeOracle backend..."
(
  cd "$ROOT_DIR/backend"
  PORT="$PORT" OPENAI_API_KEY="$OPENAI_API_KEY" make run
) &
BACKEND_PID=$!

sleep 2

if ! ps -p "$BACKEND_PID" >/dev/null 2>&1; then
  echo "Backend process exited early. See output above." >&2
  wait "$BACKEND_PID"
  exit 1
fi

echo "Starting CoffeeOracle frontend dev server..."
(cd "$ROOT_DIR/frontend" && npm run dev) &
FRONTEND_PID=$!

echo "Backend PID: $BACKEND_PID"
echo "Frontend PID: $FRONTEND_PID"
echo "Press Ctrl+C to stop both services."

wait
