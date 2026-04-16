-- name: ListPlacements :many
SELECT * FROM devon_placements ORDER BY fetched_at DESC;

-- name: ListPlacementsByModel :many
SELECT * FROM devon_placements WHERE model_id = $1 ORDER BY fetched_at DESC;

-- name: ListPlacementsByAgent :many
SELECT * FROM devon_placements WHERE bonnie_agent_id = $1 ORDER BY fetched_at DESC;

-- name: GetPlacement :one
SELECT * FROM devon_placements WHERE id = $1;

-- name: GetPlacementByModelAgent :one
SELECT * FROM devon_placements WHERE model_id = $1 AND bonnie_agent_id = $2;

-- name: UpsertPlacement :one
INSERT INTO devon_placements (model_id, bonnie_agent_id, remote_entry_id, host_path, size_bytes)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (model_id, bonnie_agent_id) DO UPDATE
    SET remote_entry_id = EXCLUDED.remote_entry_id,
        host_path       = EXCLUDED.host_path,
        size_bytes      = EXCLUDED.size_bytes,
        fetched_at      = now()
RETURNING *;

-- name: DeletePlacement :exec
DELETE FROM devon_placements WHERE id = $1;

-- name: DeletePlacementsByModel :exec
DELETE FROM devon_placements WHERE model_id = $1;
