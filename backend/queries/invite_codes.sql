-- name: GetInviteCode :one
SELECT code, expiry_date, CAST(json(roles) AS TEXT) as roles FROM invite_codes
WHERE code = sqlc.arg(code);

-- name: CreateInviteCode :one
INSERT INTO invite_codes (code, expiry_date, roles)
VALUES (hex(randomblob(16)), unixepoch() + sqlc.arg(hours) * 3600, jsonb(sqlc.arg(rolesJson)))
RETURNING code, expiry_date;

-- name: CheckInviteCode :one
SELECT CAST(EXISTS(SELECT 1 FROM invite_codes WHERE code = sqlc.arg(code)) as BOOLEAN);

-- name: RemoveInviteCode :exec
DELETE FROM invite_codes WHERE code = sqlc.arg(code);