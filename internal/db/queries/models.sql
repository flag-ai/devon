-- name: ListModels :many
SELECT * FROM devon_models ORDER BY source, model_id;

-- name: GetModel :one
SELECT * FROM devon_models WHERE id = $1;

-- name: GetModelByIdentity :one
SELECT * FROM devon_models WHERE source = $1 AND model_id = $2;

-- name: UpsertModel :one
INSERT INTO devon_models (source, model_id, metadata)
VALUES ($1, $2, $3)
ON CONFLICT (source, model_id) DO UPDATE
    SET metadata = EXCLUDED.metadata,
        updated_at = now()
RETURNING *;

-- name: MarkModelDownloaded :exec
UPDATE devon_models
SET downloaded_at = now(), updated_at = now()
WHERE id = $1;

-- name: TouchModelUsed :exec
UPDATE devon_models
SET last_used_at = now(), updated_at = now()
WHERE id = $1;

-- name: DeleteModel :exec
DELETE FROM devon_models WHERE id = $1;
