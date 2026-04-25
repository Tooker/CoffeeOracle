# CoffeeOracle Go + Next.js Stack Guide

## Overview

This document explains how to run and extend the new Go backend (`backend/`) and
Next.js frontend (`frontend/`). The backend accepts image uploads, sanitizes
user input, calls OpenAI, and streams fortunes over Server-Sent Events (SSE).
The frontend provides the upload interface, theme shell, and SSE consumption
logic.

## Repository layout

- `backend/` – Go service with `cmd/server`, `internal/config`, `internal/oracle`,
  `internal/server`, and `tests/`
- `frontend/` – Next.js 16 App Router workspace (TypeScript, Tailwind, Flowbite)
- `docs/` – This guide and future stack notes

## Prerequisites

- Go 1.22+
- Node.js 20+ / npm 10+
- OpenAI API key with Responses API access

## Backend setup

1. Copy `backend/env.example.txt` to `.env` (or set env vars another way):

   ```bash
   cp backend/env.example.txt backend/.env
   # edit PORT / OPENAI_API_KEY as needed
   ```

2. From repository root, run backend tooling via Makefile:

   ```bash
   make -C backend tidy   # optional: keeps go.mod/go.sum fresh
   make -C backend lint   # go vet
   make -C backend test   # unit + integration tests (mocked OpenAI)
   make -C backend run    # starts server on PORT
   ```

3. The server exposes:

   - `GET /healthz` – health probe
   - `POST /api/oracle` – accepts multipart or JSON payloads, resizes images to
     ≤1024px, sanitizes names (scripts/URLs stripped), and streams fortunes using
     SSE (`text/event-stream`).

## Frontend setup

1. Copy `frontend/env.local.example.txt` to `.env.local` (or set via shell):

   ```bash
   cp frontend/env.local.example.txt frontend/.env.local
   # ensure NEXT_PUBLIC_API_BASE matches backend origin, e.g. http://localhost:8080
   ```

2. Install dependencies and run scripts from `frontend/`:

   ```bash
   cd frontend
   npm install
   npm run lint     # ESLint (Next.js config)
   npm run test     # Jest + Testing Library
   npm run dev      # Next.js dev server on http://localhost:3000
   npm run build    # production build check
   ```

3. The UI exposes:

   - Hero + backend target summary (server component)
   - `UploadForm` (client) → sanitized name entry, creativity slider (0–10), and
     Flowbite file input
   - `useOracleStream` hook + SSE client (`lib/api/oracleClient.ts`) bridging to
     `/api/oracle`
   - `ResultsPanel` with live streaming output & error states

## Manual integration test

1. Start backend: `make -C backend run` (defaults to `:8080`).
2. Start frontend: `cd frontend && npm run dev`.
3. Navigate to `http://localhost:3000`, upload a coffee image (PNG/JPEG/WEBP),
   enter a name and creativity level, then submit. Watch the streaming fortune
   populate the results panel. Check backend logs for sanitized input entries.

## Automated tests summary

- Backend: `go test ./...` (unit + handler/integration tests with mocked OpenAI)
- Frontend: `npm run test` (Jest ensures SSE parser + layout renders)

## Prompt-safety & next steps

- Name fields are sanitized on both frontend (before submission) and backend
  (validator) to remove scripts, URLs, and disallowed characters.
- Backend enforces MIME + file-size limits, resizes images, and uses SSE error
  events to avoid leaking stack traces to clients.
- Future work: deployment automation, auth hardening, additional UI tests, and
  real OpenAI streaming smoke tests in staging.
