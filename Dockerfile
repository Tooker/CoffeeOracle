# syntax=docker/dockerfile:1.7

FROM golang:1.25.5-alpine AS backend-builder
WORKDIR /build/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/. .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM node:22-alpine AS frontend-builder
WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/. .
ARG NEXT_PUBLIC_API_BASE=http://localhost:8080
ENV NEXT_PUBLIC_API_BASE=$NEXT_PUBLIC_API_BASE
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build && npm prune --omit=dev

FROM node:22-alpine AS runtime
RUN apk add --no-cache bash tini
WORKDIR /app

COPY --from=backend-builder /out/server /app/bin/server
COPY --from=frontend-builder /build/frontend /app/frontend
COPY docker/start.sh /app/docker/start.sh

RUN chmod +x /app/docker/start.sh /app/bin/server

ENV PORT=8080
ENV FRONTEND_PORT=3000
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

EXPOSE 3000

ENTRYPOINT ["/sbin/tini", "--", "/app/docker/start.sh"]
