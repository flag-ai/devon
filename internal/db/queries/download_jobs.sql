-- name: ListDownloadJobs :many
SELECT * FROM devon_download_jobs ORDER BY created_at DESC;

-- name: GetDownloadJob :one
SELECT * FROM devon_download_jobs WHERE id = $1;

-- name: ListPendingDownloadJobs :many
SELECT * FROM devon_download_jobs WHERE status = 'pending' ORDER BY created_at;

-- name: CreateDownloadJob :one
INSERT INTO devon_download_jobs (model_id, bonnie_agent_id, status, patterns)
VALUES ($1, $2, 'pending', $3)
RETURNING *;

-- name: MarkDownloadJobRunning :exec
UPDATE devon_download_jobs
SET status = 'running', started_at = now(), updated_at = now()
WHERE id = $1;

-- name: MarkDownloadJobSucceeded :exec
UPDATE devon_download_jobs
SET status = 'succeeded', finished_at = now(), updated_at = now(), error = ''
WHERE id = $1;

-- name: MarkDownloadJobFailed :exec
UPDATE devon_download_jobs
SET status = 'failed', finished_at = now(), updated_at = now(), error = $2
WHERE id = $1;

-- name: RestartDownloadJob :exec
UPDATE devon_download_jobs
SET status = 'pending', started_at = NULL, finished_at = NULL, error = '', updated_at = now()
WHERE id = $1;
