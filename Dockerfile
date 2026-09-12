# ---- Frontend builder ----
FROM node:20-slim AS frontend-builder

WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --ignore-scripts
COPY frontend/ ./
RUN npm run build

# ---- Python builder ----
FROM python:3.12-slim AS builder

RUN pip install --no-cache-dir poetry==1.8.* \
    && poetry config virtualenvs.in-project true

WORKDIR /app
COPY pyproject.toml poetry.lock* README.md ./
RUN poetry install --no-interaction --no-ansi --extras api --without dev,docs --no-root

COPY src/ src/
COPY --from=frontend-builder /src/devon/ui/static src/devon/ui/static
RUN poetry install --no-interaction --no-ansi --extras api --without dev,docs

# ---- Runtime ----
FROM python:3.12-slim

RUN groupadd --gid 1000 devon \
    && useradd --uid 1000 --gid devon --create-home devon

WORKDIR /app
COPY --from=builder /app/.venv .venv
COPY --from=builder /app/src src
COPY pyproject.toml ./

ENV PATH="/app/.venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    DEVON_STORAGE_PATH=/data/models \
    DEVON_CONFIG_PATH=/data/config.yaml

RUN mkdir -p /data/models && chown -R devon:devon /data

USER devon

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health')" || exit 1

ENTRYPOINT ["devon", "serve", "--host", "0.0.0.0", "--port", "8000"]
