#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_IMAGE="$ROOT_DIR/uploaded_files/WhatsApp Image 2026-01-29 at 18.11.52.jpeg"

API_BASE="${API_BASE:-http://localhost:8080}"
NAME="${NAME:-Tobias}"
CREATIVITY="${CREATIVITY:-5}"
IMAGE_PATH="${1:-$DEFAULT_IMAGE}"

if [[ ! -f "$IMAGE_PATH" ]]; then
  echo "Error: image not found: $IMAGE_PATH" >&2
  echo "Usage: $0 [image-path]" >&2
  exit 1
fi

if [[ ! "$CREATIVITY" =~ ^[0-9]+$ ]] || (( CREATIVITY < 0 || CREATIVITY > 10 )); then
  echo "Error: CREATIVITY must be an integer between 0 and 10 (current: $CREATIVITY)" >&2
  exit 1
fi

echo "Testing backend stream against: $API_BASE/api/oracle"
echo "Using image: $IMAGE_PATH"
echo "Name: $NAME | Creativity: $CREATIVITY"
echo

curl --no-buffer --show-error --fail \
  -X POST "$API_BASE/api/oracle" \
  -F "name=$NAME" \
  -F "creativity=$CREATIVITY" \
  -F "file=@$IMAGE_PATH;type=image/jpeg"
