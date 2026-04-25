# CoffeeOracle – Code- und Config-Referenz (für Nicht-Go/JS-Kenner)

Diese Datei zeigt dir, **wo** welche Logik liegt und verlinkt auf die Originaldateien.
Ziel: Du kannst schnell vom Konzept zur tatsächlichen Implementierung springen.

## 1) Backend (Go)

### Einstieg & Server-Lebenszyklus
- [`backend/cmd/server/main.go`](../backend/cmd/server/main.go)
  - Startpunkt der Anwendung.
  - Lädt Konfiguration, initialisiert Logger und Services, startet HTTP-Server, behandelt sauberes Herunterfahren.

### Konfiguration & Logging
- [`backend/internal/config/config.go`](../backend/internal/config/config.go)
  - Liest Umgebungsvariablen (`PORT`, `OPENAI_API_KEY`, Logging-Optionen).
  - Validiert Pflichtfelder.
- [`backend/internal/logger/logger.go`](../backend/internal/logger/logger.go)
  - Einfaches Logging mit Leveln (`error`, `info`, `debug`).

### Oracle-Domäne
- [`backend/internal/oracle/models.go`](../backend/internal/oracle/models.go)
  - Request-Datenmodell (`OracleRequest`).
- [`backend/internal/oracle/validator.go`](../backend/internal/oracle/validator.go)
  - Sicherheits- und Plausibilitätsprüfung für Eingaben.
- [`backend/internal/oracle/service.go`](../backend/internal/oracle/service.go)
  - Kommunikation mit OpenAI Responses API.
  - Verarbeitung von SSE-Streaming-Events.

### HTTP-Schicht
- [`backend/internal/server/router.go`](../backend/internal/server/router.go)
  - Definiert Routen (`/healthz`, `/api/oracle`) und Middleware-Kette.
- [`backend/internal/server/middleware.go`](../backend/internal/server/middleware.go)
  - Logging, CORS, Timeout.
- [`backend/internal/server/handlers/oracle.go`](../backend/internal/server/handlers/oracle.go)
  - Nimmt Uploads entgegen, verarbeitet Bilder, startet Stream-Antwort.

### Backend-Tests
- [`backend/internal/server/handlers/oracle_test.go`](../backend/internal/server/handlers/oracle_test.go)
- [`backend/internal/oracle/service_test.go`](../backend/internal/oracle/service_test.go)
- [`backend/internal/oracle/validator_test.go`](../backend/internal/oracle/validator_test.go)
- [`backend/internal/server/middleware_test.go`](../backend/internal/server/middleware_test.go)
- [`backend/tests/oracle_integration_test.go`](../backend/tests/oracle_integration_test.go)

## 2) Frontend (Next.js / TypeScript)

### Seitenstruktur
- [`frontend/app/layout.tsx`](../frontend/app/layout.tsx)
  - Globales Layout (HTML-Frame, Footer, Hauptcontainer).
- [`frontend/app/page.tsx`](../frontend/app/page.tsx)
  - Landing-Page und Einstieg in die Oracle-Interaktion.

### UI-Komponenten
- [`frontend/components/OracleExperience.tsx`](../frontend/components/OracleExperience.tsx)
  - Steuert den Gesamtfluss zwischen Formular und Ergebnisanzeige.
- [`frontend/components/UploadForm.tsx`](../frontend/components/UploadForm.tsx)
  - Formulareingaben (Name, Kreativität, Bild).
- [`frontend/components/ResultsPanel.tsx`](../frontend/components/ResultsPanel.tsx)
  - Stream-Ausgabe, Fehlerzustände, Markdown-Rendering.
- [`frontend/components/HeroActions.tsx`](../frontend/components/HeroActions.tsx)
  - Zusätzliche CTA-Buttons.

### Frontend-Logik
- [`frontend/hooks/useOracleStream.ts`](../frontend/hooks/useOracleStream.ts)
  - Hook für Upload-/Streaming-Status, Fehlerbehandlung, Reset/Abort.
- [`frontend/lib/api/oracleClient.ts`](../frontend/lib/api/oracleClient.ts)
  - API-Client inklusive SSE-Parser.
- [`frontend/lib/env.ts`](../frontend/lib/env.ts)
  - Aufbereitung der öffentlichen API-Basis-URL.

### Frontend-Tests
- [`frontend/tests/oracleClient.test.ts`](../frontend/tests/oracleClient.test.ts)
- [`frontend/__tests__/layout.test.tsx`](../frontend/__tests__/layout.test.tsx)

## 3) Config-Dateien (mit Bedeutung)

### Direkt kommentierbare Configs (TS/MJS)
- [`frontend/next.config.ts`](../frontend/next.config.ts) – lokale API-Rewrites
- [`frontend/tailwind.config.ts`](../frontend/tailwind.config.ts) – Design-Tokens/CSS-Scan
- [`frontend/eslint.config.mjs`](../frontend/eslint.config.mjs) – Lint-Regeln
- [`frontend/postcss.config.mjs`](../frontend/postcss.config.mjs) – CSS-Pipeline
- [`frontend/jest.config.mjs`](../frontend/jest.config.mjs) – Test-Setup

### Nicht direkt kommentierbare Configs (JSON)
JSON erlaubt offiziell keine Kommentare. Deshalb hier die Referenzen:
- [`frontend/package.json`](../frontend/package.json) – Skripte & Dependencies
- [`frontend/tsconfig.json`](../frontend/tsconfig.json) – TypeScript-Compileroptionen
- [`frontend/package-lock.json`](../frontend/package-lock.json) – gelockte Dependency-Versionen

### Weitere Projekt-Configs
- [`pyproject.toml`](../pyproject.toml) – Python-Tooling für das Repo
- [`backend/.env`](../backend/.env) und [`.env`](../.env) – lokale Umgebungsvariablen

## 4) Lesereihenfolge (empfohlen)
1. `backend/cmd/server/main.go`
2. `backend/internal/server/handlers/oracle.go`
3. `backend/internal/oracle/service.go`
4. `frontend/components/OracleExperience.tsx`
5. `frontend/hooks/useOracleStream.ts`
6. `frontend/lib/api/oracleClient.ts`

So verstehst du erst den Backend-Flow und dann den Frontend-Stream bis zur Anzeige.
