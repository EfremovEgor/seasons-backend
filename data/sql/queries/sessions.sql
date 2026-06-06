-- name: CreateSession :one
INSERT INTO sessions (user_id, token, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE token = $1
  AND expires_at > NOW()
LIMIT 1;

-- name: UpdateSessionLastUsed :exec
UPDATE sessions
SET last_used_at = NOW()
WHERE token = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = $1;

-- name: DeleteUserSessions :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < NOW();
