-- name: GetInviteCode :one
SELECT code, expiry_date FROM invite_codes
WHERE code = sqlc.arg(code);

-- name: CreateInviteCode :one
INSERT INTO invite_codes (code, expiry_date)
VALUES (hex(randomblob(16)), unixepoch() + sqlc.arg(hours) * 3600)
RETURNING code, expiry_date;

-- name: CheckInviteCode :one
SELECT CAST(EXISTS(SELECT 1 FROM invite_codes WHERE code = sqlc.arg(code)) as BOOLEAN);

-- name: RemoveInviteCode :exec
DELETE FROM invite_codes WHERE code = sqlc.arg(code);