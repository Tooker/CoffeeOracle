# CoffeeOracle Backend

Go service providing the CoffeeOracle API for the upcoming Next.js frontend.

## Prerequisites

- Go 1.22+
- OpenAI API key

## Configuration

Copy `env.example.txt` to `.env` (or inject the following environment variables):

```
PORT=8080
OPENAI_API_KEY=sk-your-openai-key
LOG_LEVEL=info       # error | info | debug
LOG_ENABLED=true      # set to false to silence logs
```

## Commands

```
make tidy   # ensure go.mod/go.sum are up to date
make lint   # go vet
make test   # unit and integration tests
make run    # start server on PORT
```

## Architecture

- `cmd/server`: entry point with graceful shutdown
- `internal/config`: env loading + validation
- `internal/logger`: leveled logging that can be disabled via env
- `internal/oracle`: request models, validation, OpenAI service (Responses API)
- `internal/server`: router + middleware
- `internal/server/handlers`: HTTP handlers (SSE streaming, uploads)
- `tests`: integration tests (mocked OpenAI)

## Prompt Safety

The backend enforces prompt-injection safeguards by sanitizing user names,
limiting creativity to 0-10, resizing images, and validating MIME types before
constructing OpenAI prompts. SSE responses avoid echoing raw errors to clients.
