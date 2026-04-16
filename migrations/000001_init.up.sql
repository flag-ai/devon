-- DEVON initial schema.

CREATE TABLE devon_models (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source         TEXT NOT NULL,
    model_id       TEXT NOT NULL,
    metadata       JSONB NOT NULL DEFAULT '{}'::jsonb,
    downloaded_at  TIMESTAMPTZ,
    last_used_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source, model_id)
);

CREATE INDEX idx_devon_models_source ON devon_models (source);

CREATE TABLE devon_bonnie_agents (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL UNIQUE,
    url           TEXT NOT NULL,
    token         TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'offline',
    last_seen_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE devon_placements (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id         UUID NOT NULL REFERENCES devon_models(id) ON DELETE CASCADE,
    bonnie_agent_id  UUID NOT NULL REFERENCES devon_bonnie_agents(id) ON DELETE CASCADE,
    remote_entry_id  TEXT NOT NULL DEFAULT '',
    host_path        TEXT NOT NULL,
    size_bytes       BIGINT NOT NULL DEFAULT 0,
    fetched_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (model_id, bonnie_agent_id)
);

CREATE INDEX idx_devon_placements_agent ON devon_placements (bonnie_agent_id);

CREATE TABLE devon_download_jobs (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id          UUID NOT NULL REFERENCES devon_models(id) ON DELETE CASCADE,
    bonnie_agent_id   UUID NOT NULL REFERENCES devon_bonnie_agents(id) ON DELETE CASCADE,
    status            TEXT NOT NULL DEFAULT 'pending',
    patterns          JSONB NOT NULL DEFAULT '[]'::jsonb,
    error             TEXT NOT NULL DEFAULT '',
    started_at        TIMESTAMPTZ,
    finished_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_devon_download_jobs_status ON devon_download_jobs (status);
CREATE INDEX idx_devon_download_jobs_agent  ON devon_download_jobs (bonnie_agent_id);
