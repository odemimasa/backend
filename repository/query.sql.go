// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createTx = `-- name: CreateTx :exec
INSERT INTO transaction (id, user_id, subscription_plan_id, ref_id, coupon_code, payment_method, qr_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

type CreateTxParams struct {
	ID                 pgtype.UUID `json:"id"`
	UserID             string      `json:"user_id"`
	SubscriptionPlanID pgtype.UUID `json:"subscription_plan_id"`
	RefID              string      `json:"ref_id"`
	CouponCode         pgtype.Text `json:"coupon_code"`
	PaymentMethod      string      `json:"payment_method"`
	QrUrl              string      `json:"qr_url"`
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
	)
	return err
}

const createUser = `-- name: CreateUser :exec
INSERT INTO "user" (id, name, email) VALUES ($1, $2, $3)
`

type CreateUserParams struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.Exec(ctx, createUser, arg.ID, arg.Name, arg.Email)
	return err
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

const deleteUserByID = `-- name: DeleteUserByID :one
DELETE FROM "user" WHERE id = $1 RETURNING id
`

func (q *Queries) DeleteUserByID(ctx context.Context, id string) (string, error) {
	row := q.db.QueryRow(ctx, deleteUserByID, id)
	err := row.Scan(&id)
	return id, err
}

const getSubsPlanByID = `-- name: GetSubsPlanByID :one
SELECT id, name, price, duration_in_seconds, created_at, deleted_at FROM subscription_plan WHERE id = $1
`

func (q *Queries) GetSubsPlanByID(ctx context.Context, id pgtype.UUID) (SubscriptionPlan, error) {
	row := q.db.QueryRow(ctx, getSubsPlanByID, id)
	var i SubscriptionPlan
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Price,
		&i.DurationInSeconds,
		&i.CreatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const getTransactions = `-- name: GetTransactions :many
SELECT id, user_id, subscription_plan_id, ref_id, coupon_code, payment_method, qr_url, status, created_at, paid_at, expired_at FROM transaction
`

func (q *Queries) GetTransactions(ctx context.Context) ([]Transaction, error) {
	rows, err := q.db.Query(ctx, getTransactions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Transaction
	for rows.Next() {
		var i Transaction
		if err := rows.Scan(
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

const getTxWithSubsPlanByID = `-- name: GetTxWithSubsPlanByID :one
SELECT 
  t.id AS transaction_id,
  t.user_id,
  s.duration_in_seconds
FROM transaction t JOIN subscription_plan s ON t.subscription_plan_id = s.id WHERE t.id = $1
`

type GetTxWithSubsPlanByIDRow struct {
	TransactionID     pgtype.UUID `json:"transaction_id"`
	UserID            string      `json:"user_id"`
	DurationInSeconds int32       `json:"duration_in_seconds"`
}

func (q *Queries) GetTxWithSubsPlanByID(ctx context.Context, id pgtype.UUID) (GetTxWithSubsPlanByIDRow, error) {
	row := q.db.QueryRow(ctx, getTxWithSubsPlanByID, id)
	var i GetTxWithSubsPlanByIDRow
	err := row.Scan(&i.TransactionID, &i.UserID, &i.DurationInSeconds)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, name, email, phone_number, phone_verified, account_type, upgraded_at, expired_at, created_at FROM "user" WHERE id = $1
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
		&i.UpgradedAt,
		&i.ExpiredAt,
		&i.CreatedAt,
	)
	return i, err
}

const getUserByPhoneNumber = `-- name: GetUserByPhoneNumber :one
SELECT id, name, email, phone_number, phone_verified, account_type, upgraded_at, expired_at, created_at FROM "user" WHERE phone_number = $1
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
		&i.UpgradedAt,
		&i.ExpiredAt,
		&i.CreatedAt,
	)
	return i, err
}

const incrementCouponQuota = `-- name: IncrementCouponQuota :exec
UPDATE coupon SET quota = quota + 1 WHERE code = $1
`

func (q *Queries) IncrementCouponQuota(ctx context.Context, code string) error {
	_, err := q.db.Exec(ctx, incrementCouponQuota, code)
	return err
}

const updateTxStatus = `-- name: UpdateTxStatus :exec
UPDATE transaction SET status = $2 WHERE id = $1
`

type UpdateTxStatusParams struct {
	ID     pgtype.UUID       `json:"id"`
	Status TransactionStatus `json:"status"`
}

func (q *Queries) UpdateTxStatus(ctx context.Context, arg UpdateTxStatusParams) error {
	_, err := q.db.Exec(ctx, updateTxStatus, arg.ID, arg.Status)
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
UPDATE "user" SET account_type = $2, upgraded_at = $3, expired_at = $4
WHERE id = $1
`

type UpdateUserSubsParams struct {
	ID          string             `json:"id"`
	AccountType AccountType        `json:"account_type"`
	UpgradedAt  pgtype.Timestamptz `json:"upgraded_at"`
	ExpiredAt   pgtype.Timestamptz `json:"expired_at"`
}

func (q *Queries) UpdateUserSubs(ctx context.Context, arg UpdateUserSubsParams) error {
	_, err := q.db.Exec(ctx, updateUserSubs,
		arg.ID,
		arg.AccountType,
		arg.UpgradedAt,
		arg.ExpiredAt,
	)
	return err
}
