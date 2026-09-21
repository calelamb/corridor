-- name: EventCount :one
SELECT ingested_events FROM public_coverage;
-- name: Sources :many
SELECT id,name,url,license,status FROM public_sources ORDER BY id LIMIT $1 OFFSET $2;
-- name: SourceCount :one
SELECT count(*)::bigint FROM public_sources;
