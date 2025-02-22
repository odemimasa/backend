-- name: GetUserByID :one
SELECT * FROM "user" WHERE id = $1;

-- name: GetUserPrayerByID :one
SELECT
  u.phone_number,
  u.account_type,
  u.time_zone
FROM "user" u WHERE u.id = $1;

-- name: GetUserByPhoneNumber :one
SELECT * FROM "user" WHERE phone_number = $1;

-- name: GetUsersByTimeZone :many
SELECT
  u.id,
  u.phone_number,
  u.account_type,
  u.time_zone
FROM "user" u WHERE u.time_zone = $1;

-- name: GetUserPhoneByID :one
SELECT u.phone_number FROM "user" u WHERE u.id = $1;

-- name: GetUserTimeZoneByID :one
SELECT u.time_zone FROM "user" u WHERE u.id = $1;

-- name: GetUserSubsByID :one
SELECT u.account_type FROM "user" u WHERE u.id = $1;

-- name: UpdateUserPhoneNumber :exec
UPDATE "user" SET phone_number = $2, phone_verified = $3 WHERE id = $1;

-- name: UpdateUserSubs :exec
UPDATE "user" SET account_type = $2 WHERE id = $1;

-- name: UpdateUserTimeZone :exec
UPDATE "user" SET time_zone = $2 WHERE id = $1;

-- name: CreateUser :one
INSERT INTO "user" (id, name, email) VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteUserByID :one
DELETE FROM "user" WHERE id = $1 RETURNING id;

-- name: GetSubsPlans :many
SELECT * FROM subscription_plan WHERE deleted_at IS NULL;

-- name: DecrementCouponQuota :one
UPDATE coupon SET quota = quota - 1
WHERE code = $1 AND quota > 0 AND deleted_at IS NULL RETURNING quota;

-- name: IncrementCouponQuota :exec
UPDATE coupon SET quota = quota + 1 WHERE code = $1;

-- name: GetTxByUserID :many
SELECT
  t.id AS transaction_id,
  t.coupon_code,
  t.status,
  t.qr_url,
  t.paid_at,
  t.expired_at,
  s.price,
  s.duration_in_months
FROM transaction t JOIN subscription_plan s ON t.subscription_plan_id = s.id
WHERE t.user_id = $1 AND (status = 'PAID' OR (status = 'UNPAID' AND expired_at > NOW()));

-- name: GetTxByID :one
SELECT * FROM transaction WHERE id = $1;

-- name: GetTxWithSubsPlanByID :one
SELECT 
  t.id AS transaction_id,
  t.user_id,
  s.duration_in_months
FROM transaction t JOIN subscription_plan s ON t.subscription_plan_id = s.id WHERE t.id = $1;

-- name: CreateTx :exec
INSERT INTO transaction (id, user_id, subscription_plan_id, ref_id, coupon_code, payment_method, qr_url, expired_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpdateTxStatus :exec
UPDATE transaction SET status = $2, paid_at = $3 WHERE id = $1;

-- name: GetTasksByUserID :many
SELECT 
  t.id,
  t.name,
  t.description,
  t.checked
 FROM task t WHERE t.user_id = $1;

-- name: CreateTask :one
INSERT INTO task (user_id, name, description) VALUES ($1, $2, $3) RETURNING id, name, description, checked;

-- name: UpdateTaskByID :exec
UPDATE task SET name = $2, description = $3, checked = $4 WHERE id = $1;

-- name: RemoveCheckedTask :exec
DELETE FROM task WHERE checked = TRUE;

-- name: DeleteTaskByID :exec
DELETE FROM task WHERE id = $1;

-- name: GetTodayPrayers :many
SELECT
  p.id,
  p.name,
  p.status
FROM prayer p WHERE p.user_id = $1 AND p.year = $2 AND p.month = $3 AND p.day = $4;

-- name: GetThisMonthPrayers :many
SELECT
  p.id,
  p.name,
  p.status
FROM prayer p WHERE p.user_id = $1 AND p.year = $2 AND p.month = $3 AND p.status IS NOT NULL;

-- name: CreatePrayers :copyfrom
INSERT INTO prayer (id, user_id, name, year, month, day)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdatePrayerStatus :exec
UPDATE prayer SET status = $2 WHERE id = $1;

-- name: UpdatePrayersToMissed :exec
UPDATE prayer SET status = 'MISSED' WHERE status IS NULL AND (day < $1 OR month < $2 OR year < $3);