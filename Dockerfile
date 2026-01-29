# Use a Python image with uv pre-installed
FROM ghcr.io/astral-sh/uv:debian-slim

# Setup a non-root user
#RUN groupadd --system --gid 999 nonroot \
# && useradd --system --gid 999 --uid 999 --create-home nonroot
run apt update && apt install unzip curl -y
# Install the project into `/app`
WORKDIR /app

# Enable bytecode compilation
ENV UV_COMPILE_BYTECODE=1

COPY pyproject.toml pyproject.toml

# Ensure installed tools can be executed out of the box
ENV UV_TOOL_BIN_DIR=/usr/local/bin
COPY . .
RUN uv sync



EXPOSE 3000

ENTRYPOINT [ "uv", "run","reflex" ,"run"]