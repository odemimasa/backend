// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPrayer = `-- name: CreatePrayer :exec
INSERT INTO prayer (user_id, name, status, year, month, day)
VALUES ($1, $2, $3, $4, $5, $6)
`

type CreatePrayerParams struct {
	UserID string       `json:"user_id"`
	Name   string       `json:"name"`
	Status PrayerStatus `json:"status"`
	Year   int16        `json:"year"`
	Month  int16        `json:"month"`
	Day    int16        `json:"day"`
}

func (q *Queries) CreatePrayer(ctx context.Context, arg CreatePrayerParams) error {
	_, err := q.db.Exec(ctx, createPrayer,
		arg.UserID,
		arg.Name,
		arg.Status,
		arg.Year,
		arg.Month,
		arg.Day,
	)
	return err
}

const createTask = `-- name: CreateTask :one
INSERT INTO task (user_id, name, description) VALUES ($1, $2, $3) RETURNING id, name, description, checked
`

type CreateTaskParams struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateTaskRow struct {
	ID          pgtype.UUID `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Checked     bool        `json:"checked"`
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) (CreateTaskRow, error) {
	row := q.db.QueryRow(ctx, createTask, arg.UserID, arg.Name, arg.Description)
	var i CreateTaskRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Checked,
	)
	return i, err
}

const createTx = `-- name: CreateTx :exec
INSERT INTO transaction (id, user_id, subscription_plan_id, ref_id, coupon_code, payment_method, qr_url, expired_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

type CreateTxParams struct {
	ID                 pgtype.UUID        `json:"id"`
	UserID             string             `json:"user_id"`
	SubscriptionPlanID pgtype.UUID        `json:"subscription_plan_id"`
	RefID              string             `json:"ref_id"`
	CouponCode         pgtype.Text        `json:"coupon_code"`
	PaymentMethod      string             `json:"payment_method"`
	QrUrl              string             `json:"qr_url"`
	ExpiredAt          pgtype.Timestamptz `json:"expired_at"`
}

func (q *Queries) CreateTx(ctx context.Context, arg CreateTxParams) error {
	_, err := q.db.Exec(ctx, createTx,
		arg.ID,
		arg.UserID,
		arg.SubscriptionPlanID,
		arg.RefID,
		arg.CouponCode,
		arg.PaymentMethod,
		arg.QrUrl,
		arg.ExpiredAt,
	)
	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO "user" (id, name, email) VALUES ($1, $2, $3) RETURNING id, name, email, phone_number, phone_verified, account_type, time_zone, created_at
`

type CreateUserParams struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.ID, arg.Name, arg.Email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.PhoneNumber,
		&i.PhoneVerified,
		&i.AccountType,
		&i.TimeZone,
		&i.CreatedAt,
	)
	return i, err
}

const decrementCouponQuota = `-- name: DecrementCouponQuota :one
UPDATE coupon SET quota = quota - 1
WHERE code = $1 AND quota > 0 AND deleted_at IS NULL RETURNING quota
`

func (q *Queries) DecrementCouponQuota(ctx context.Context, code string) (int16, error) {
	row := q.db.QueryRow(ctx, decrementCouponQuota, code)
	var quota int16
	err := row.Scan(&quota)
	return quota, err
}

const deleteTaskByID = `-- name: DeleteTaskByID :exec
DELETE FROM task WHERE id = $1
`

func (q *Queries) DeleteTaskByID(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteTaskByID, id)
	return err
}

const deleteUserByID = `-- name: DeleteUserByID :one
DELETE FROM "user" WHERE id = $1 RETURNING id
`

func (q *Queries) DeleteUserByID(ctx context.Context, id string) (string, error) {
	row := q.db.QueryRow(ctx, deleteUserByID, id)
	err := row.Scan(&id)
	return id, err
}

const getSubsPlanByID = `-- name: GetSubsPlanByID :one
SELECT id, name, price, duration_in_months, created_at, deleted_at FROM subscription_plan WHERE id = $1
`

func (q *Queries) GetSubsPlanByID(ctx context.Context, id pgtype.UUID) (SubscriptionPlan, error) {
	row := q.db.QueryRow(ctx, getSubsPlanByID, id)
	var i SubscriptionPlan
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Price,
		&i.DurationInMonths,
		&i.CreatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const getSubsPlans = `-- name: GetSubsPlans :many
SELECT id, name, price, duration_in_months, created_at, deleted_at FROM subscription_plan WHERE deleted_at IS NULL
`

func (q *Queries) GetSubsPlans(ctx context.Context) ([]SubscriptionPlan, error) {
	rows, err := q.db.Query(ctx, getSubsPlans)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubscriptionPlan
	for rows.Next() {
		var i SubscriptionPlan
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Price,
			&i.DurationInMonths,
			&i.CreatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTasksByUserID = `-- name: GetTasksByUserID :many
SELECT 
  t.id,
  t.name,
  t.description,
  t.checked
 FROM task t WHERE t.user_id = $1
`

type GetTasksByUserIDRow struct {
	ID          pgtype.UUID `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Checked     bool        `json:"checked"`
}

func (q *Queries) GetTasksByUserID(ctx context.Context, userID string) ([]GetTasksByUserIDRow, error) {
	rows, err := q.db.Query(ctx, getTasksByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTasksByUserIDRow
	for rows.Next() {
		var i GetTasksByUserIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Checked,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTxByID = `-- name: GetTxByID :one
SELECT id, user_id, subscription_plan_id, ref_id, coupon_code, payment_method, qr_url, status, created_at, paid_at, expired_at FROM transaction WHERE id = $1
`

func (q *Queries) GetTxByID(ctx context.Context, id pgtype.UUID) (Transaction, error) {
	row := q.db.QueryRow(ctx, getTxByID, id)
	var i Transaction
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.SubscriptionPlanID,
		&i.RefID,
		&i.CouponCode,
		&i.PaymentMethod,
		&i.QrUrl,
		&i.Status,
		&i.CreatedAt,
		&i.PaidAt,
		&i.ExpiredAt,
	)
	return i, err
}

const getTxByUserID = `-- name: GetTxByUserID :many
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
WHERE t.user_id = $1 AND (status = 'PAID' OR (status = 'UNPAID' AND expired_at > NOW()))
`

type GetTxByUserIDRow struct {
	TransactionID    pgtype.UUID        `json:"transaction_id"`
	CouponCode       pgtype.Text        `json:"coupon_code"`
	Status           TransactionStatus  `json:"status"`
	QrUrl            string             `json:"qr_url"`
	PaidAt           pgtype.Timestamptz `json:"paid_at"`
	ExpiredAt        pgtype.Timestamptz `json:"expired_at"`
	Price            int32              `json:"price"`
	DurationInMonths int16              `json:"duration_in_months"`
}

func (q *Queries) GetTxByUserID(ctx context.Context, userID string) ([]GetTxByUserIDRow, error) {
	rows, err := q.db.Query(ctx, getTxByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTxByUserIDRow
	for rows.Next() {
		var i GetTxByUserIDRow
		if err := rows.Scan(
			&i.TransactionID,
			&i.CouponCode,
			&i.Status,
			&i.QrUrl,
			&i.PaidAt,
			&i.ExpiredAt,
			&i.Price,
			&i.DurationInMonths,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTxWithSubsPlanByID = `-- name: GetTxWithSubsPlanByID :one
SELECT 
  t.id AS transaction_id,
  t.user_id,
  s.duration_in_months
FROM transaction t JOIN subscription_plan s ON t.subscription_plan_id = s.id WHERE t.id = $1
`

type GetTxWithSubsPlanByIDRow struct {
	TransactionID    pgtype.UUID `json:"transaction_id"`
	UserID           string      `json:"user_id"`
	DurationInMonths int16       `json:"duration_in_months"`
}

func (q *Queries) GetTxWithSubsPlanByID(ctx context.Context, id pgtype.UUID) (GetTxWithSubsPlanByIDRow, error) {
	row := q.db.QueryRow(ctx, getTxWithSubsPlanByID, id)
	var i GetTxWithSubsPlanByIDRow
	err := row.Scan(&i.TransactionID, &i.UserID, &i.DurationInMonths)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, name, email, phone_number, phone_verified, account_type, time_zone, created_at FROM "user" WHERE id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.PhoneNumber,
		&i.PhoneVerified,
		&i.AccountType,
		&i.TimeZone,
		&i.CreatedAt,
	)
	return i, err
}

const getUserByPhoneNumber = `-- name: GetUserByPhoneNumber :one
SELECT id, name, email, phone_number, phone_verified, account_type, time_zone, created_at FROM "user" WHERE phone_number = $1
`

func (q *Queries) GetUserByPhoneNumber(ctx context.Context, phoneNumber pgtype.Text) (User, error) {
	row := q.db.QueryRow(ctx, getUserByPhoneNumber, phoneNumber)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.PhoneNumber,
		&i.PhoneVerified,
		&i.AccountType,
		&i.TimeZone,
		&i.CreatedAt,
	)
	return i, err
}

const getUserPhoneByID = `-- name: GetUserPhoneByID :one
SELECT u.phone_number FROM "user" u WHERE u.id = $1
`

func (q *Queries) GetUserPhoneByID(ctx context.Context, id string) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, getUserPhoneByID, id)
	var phone_number pgtype.Text
	err := row.Scan(&phone_number)
	return phone_number, err
}

const getUserPrayerByID = `-- name: GetUserPrayerByID :one
SELECT
  u.phone_number,
  u.account_type,
  u.time_zone
FROM "user" u WHERE u.id = $1
`

type GetUserPrayerByIDRow struct {
	PhoneNumber pgtype.Text           `json:"phone_number"`
	AccountType AccountType           `json:"account_type"`
	TimeZone    NullIndonesiaTimeZone `json:"time_zone"`
}

func (q *Queries) GetUserPrayerByID(ctx context.Context, id string) (GetUserPrayerByIDRow, error) {
	row := q.db.QueryRow(ctx, getUserPrayerByID, id)
	var i GetUserPrayerByIDRow
	err := row.Scan(&i.PhoneNumber, &i.AccountType, &i.TimeZone)
	return i, err
}

const getUserSubsByID = `-- name: GetUserSubsByID :one
SELECT u.account_type FROM "user" u WHERE u.id = $1
`

func (q *Queries) GetUserSubsByID(ctx context.Context, id string) (AccountType, error) {
	row := q.db.QueryRow(ctx, getUserSubsByID, id)
	var account_type AccountType
	err := row.Scan(&account_type)
	return account_type, err
}

const getUserTimeZoneByID = `-- name: GetUserTimeZoneByID :one
SELECT u.time_zone FROM "user" u WHERE u.id = $1
`

func (q *Queries) GetUserTimeZoneByID(ctx context.Context, id string) (NullIndonesiaTimeZone, error) {
	row := q.db.QueryRow(ctx, getUserTimeZoneByID, id)
	var time_zone NullIndonesiaTimeZone
	err := row.Scan(&time_zone)
	return time_zone, err
}

const getUsersByTimeZone = `-- name: GetUsersByTimeZone :many
SELECT
  u.id,
  u.phone_number,
  u.account_type,
  u.time_zone
FROM "user" u WHERE u.time_zone = $1
`

type GetUsersByTimeZoneRow struct {
	ID          string                `json:"id"`
	PhoneNumber pgtype.Text           `json:"phone_number"`
	AccountType AccountType           `json:"account_type"`
	TimeZone    NullIndonesiaTimeZone `json:"time_zone"`
}

func (q *Queries) GetUsersByTimeZone(ctx context.Context, timeZone NullIndonesiaTimeZone) ([]GetUsersByTimeZoneRow, error) {
	rows, err := q.db.Query(ctx, getUsersByTimeZone, timeZone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUsersByTimeZoneRow
	for rows.Next() {
		var i GetUsersByTimeZoneRow
		if err := rows.Scan(
			&i.ID,
			&i.PhoneNumber,
			&i.AccountType,
			&i.TimeZone,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const incrementCouponQuota = `-- name: IncrementCouponQuota :exec
UPDATE coupon SET quota = quota + 1 WHERE code = $1
`

func (q *Queries) IncrementCouponQuota(ctx context.Context, code string) error {
	_, err := q.db.Exec(ctx, incrementCouponQuota, code)
	return err
}

const removeCheckedTask = `-- name: RemoveCheckedTask :exec
DELETE FROM task WHERE checked = TRUE
`

func (q *Queries) RemoveCheckedTask(ctx context.Context) error {
	_, err := q.db.Exec(ctx, removeCheckedTask)
	return err
}

const updateTaskByID = `-- name: UpdateTaskByID :exec
UPDATE task SET name = $2, description = $3, checked = $4 WHERE id = $1
`

type UpdateTaskByIDParams struct {
	ID          pgtype.UUID `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Checked     bool        `json:"checked"`
}

func (q *Queries) UpdateTaskByID(ctx context.Context, arg UpdateTaskByIDParams) error {
	_, err := q.db.Exec(ctx, updateTaskByID,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.Checked,
	)
	return err
}

const updateTxStatus = `-- name: UpdateTxStatus :exec
UPDATE transaction SET status = $2, paid_at = $3 WHERE id = $1
`

type UpdateTxStatusParams struct {
	ID     pgtype.UUID        `json:"id"`
	Status TransactionStatus  `json:"status"`
	PaidAt pgtype.Timestamptz `json:"paid_at"`
}

func (q *Queries) UpdateTxStatus(ctx context.Context, arg UpdateTxStatusParams) error {
	_, err := q.db.Exec(ctx, updateTxStatus, arg.ID, arg.Status, arg.PaidAt)
	return err
}

const updateUserPhoneNumber = `-- name: UpdateUserPhoneNumber :exec
UPDATE "user" SET phone_number = $2, phone_verified = $3 WHERE id = $1
`

type UpdateUserPhoneNumberParams struct {
	ID            string      `json:"id"`
	PhoneNumber   pgtype.Text `json:"phone_number"`
	PhoneVerified bool        `json:"phone_verified"`
}

func (q *Queries) UpdateUserPhoneNumber(ctx context.Context, arg UpdateUserPhoneNumberParams) error {
	_, err := q.db.Exec(ctx, updateUserPhoneNumber, arg.ID, arg.PhoneNumber, arg.PhoneVerified)
	return err
}

const updateUserSubs = `-- name: UpdateUserSubs :exec
UPDATE "user" SET account_type = $2 WHERE id = $1
`

type UpdateUserSubsParams struct {
	ID          string      `json:"id"`
	AccountType AccountType `json:"account_type"`
}

func (q *Queries) UpdateUserSubs(ctx context.Context, arg UpdateUserSubsParams) error {
	_, err := q.db.Exec(ctx, updateUserSubs, arg.ID, arg.AccountType)
	return err
}

const updateUserTimeZone = `-- name: UpdateUserTimeZone :exec
UPDATE "user" SET time_zone = $2 WHERE id = $1
`

type UpdateUserTimeZoneParams struct {
	ID       string                `json:"id"`
	TimeZone NullIndonesiaTimeZone `json:"time_zone"`
}

func (q *Queries) UpdateUserTimeZone(ctx context.Context, arg UpdateUserTimeZoneParams) error {
	_, err := q.db.Exec(ctx, updateUserTimeZone, arg.ID, arg.TimeZone)
	return err
}
