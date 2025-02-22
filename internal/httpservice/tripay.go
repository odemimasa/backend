package httpservice

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/odemimasa/backend/internal/config"
	"github.com/odemimasa/backend/internal/task"
	"github.com/odemimasa/backend/repository"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const QRIS_PAYMENT_METHOD = "QRIS"

type tripayAPIResp struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type tripayOrderItem struct {
	SubscriptionPlanID string `json:"subscription_plan_id"`
	Name               string `json:"name"`
	Price              int    `json:"price"`
	Quantity           int    `json:"quantity"`
}

type createTripayTxParams struct {
	Method        string            `json:"method"`
	MerchantRef   string            `json:"merchant_ref"`
	Amount        int               `json:"amount"`
	CustomerName  string            `json:"customer_name"`
	CustomerEmail string            `json:"customer_email"`
	CustomerPhone string            `json:"customer_phone"`
	OrderItems    []tripayOrderItem `json:"order_items"`
	Signature     string            `json:"signature"`
	ExpiredTime   int               `json:"expired_time"`
}

type tripayTxData struct {
	Reference     string            `json:"reference"`
	MerchantRef   string            `json:"merchant_ref"`
	PaymentMethod string            `json:"payment_method"`
	CustomerName  string            `json:"customer_name"`
	CustomerEmail string            `json:"customer_email"`
	CustomerPhone string            `json:"customer_phone"`
	Amount        int               `json:"amount"`
	PayCode       string            `json:"pay_code"`
	PayURL        string            `json:"pay_url"`
	CheckoutURL   string            `json:"checkout_url"`
	Status        string            `json:"status"`
	ExpiredTime   int               `json:"expired_time"`
	OrderItems    []tripayOrderItem `json:"order_items"`
	QrString      string            `json:"qr_string"`
	QrURL         string            `json:"qr_url"`
}

func createTripayTxSig(merchantRef string, amount int) string {
	key := []byte(config.Env.TRIPAY_PRIVATE_KEY)
	message := fmt.Sprintf("%s%s%d", config.Env.TRIPAY_MERCHANT_CODE, merchantRef, amount)

	hash := hmac.New(sha256.New, key)
	hash.Write([]byte(message))

	return hex.EncodeToString(hash.Sum(nil))
}

func createTripayTx(params *createTripayTxParams) (*tripayAPIResp, error) {
	URL := "https://tripay.co.id/api-sandbox/transaction/create"
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode request body")
	}

	req, err := http.NewRequest(http.MethodPost, URL, &buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new http request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Env.TRIPAY_API_KEY))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make http post request")
	}
	defer res.Body.Close()

	var payload tripayAPIResp
	if err = json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}

	return &payload, nil
}

type tripayWebhookRequest struct {
	Reference         string `json:"reference"`
	MerchantRef       string `json:"merchant_ref"`
	PaymentMethod     string `json:"payment_method"`
	PaymentMethodCode string `json:"payment_method_code"`
	TotalAmount       int    `json:"total_amount"`
	FeeMerchant       int    `json:"fee_merchant"`
	FeeCustomer       int    `json:"fee_customer"`
	TotalFee          int    `json:"total_fee"`
	AmountReceived    int    `json:"amount_received"`
	IsClosedPayment   int    `json:"is_closed_payment"`
	Status            string `json:"status"`
	PaidAt            int    `json:"paid_at"`
	Note              string `json:"note"`
}

type updateTxAndUserParams struct {
	txID         [16]byte
	userID       string
	subsDuration int64
	paidAt       int
}

// this function do three things:
// 1. update transaction status
// 2. update user subscription to PREMIUM
// 3. create task queue to downgrade user
func updateTxAndUser(ctx context.Context, params *updateTxAndUserParams) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start db transaction to update transaction status and user subscription to PREMIUM")
	}

	qtx := queries.WithTx(tx)
	err = qtx.UpdateTxStatus(
		ctx,
		repository.UpdateTxStatusParams{
			ID:     pgtype.UUID{Bytes: params.txID, Valid: true},
			Status: repository.TransactionStatusPAID,
			PaidAt: pgtype.Timestamptz{Time: time.Unix(int64(params.paidAt), 0), Valid: true},
		},
	)

	if err != nil {
		return errors.Wrap(err, "failed to update transaction status")
	}

	err = qtx.UpdateUserSubs(ctx, repository.UpdateUserSubsParams{
		ID:          params.userID,
		AccountType: repository.AccountTypePREMIUM,
	})

	if err != nil {
		return errors.Wrap(err, "failed to update user subscription to PREMIUM")
	}

	asynqTask, err := task.NewUserDowngradeTask(task.UserDowngradePayload{UserID: params.userID})
	if err != nil {
		return errors.Wrap(err, "failed to create user downgrade task")
	}

	_, err = asynqClient.Enqueue(asynqTask, asynq.ProcessIn(time.Duration(params.subsDuration)*time.Second))
	if err != nil {
		return errors.Wrap(err, "failed to enqueue user downgrade task")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to commit db transaction to update transaction status and user subscription to PREMIUM")
	}

	return nil
}

type updateTxAndRollbackCouponParams struct {
	txID       [16]byte
	txStatus   string
	couponCode pgtype.Text
}

func updateTxAndRollbackCoupon(ctx context.Context, params *updateTxAndRollbackCouponParams) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start db transaction to update transaction status and/or rollback coupon quota")
	}

	qtx := queries.WithTx(tx)
	err = qtx.UpdateTxStatus(
		ctx,
		repository.UpdateTxStatusParams{
			ID:     pgtype.UUID{Bytes: params.txID, Valid: true},
			Status: repository.TransactionStatus(params.txStatus),
		},
	)

	if err != nil {
		return errors.Wrap(err, "failed to update transaction status")
	}

	if params.couponCode.Valid {
		err = qtx.IncrementCouponQuota(ctx, params.couponCode.String)
		if err != nil {
			return errors.Wrap(err, "failed to increment coupon quota")
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to commit db transaction to update transaction status and/or rollback coupon quota")
	}

	return nil
}

func tripayWebhookHandler(res http.ResponseWriter, req *http.Request) {
	start := time.Now()
	ctx := req.Context()
	logWithCtx := log.Ctx(ctx).With().Logger()

	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		logWithCtx.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to read tripay webhook request")
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	key := []byte(config.Env.TRIPAY_PRIVATE_KEY)
	hash := hmac.New(sha256.New, key)
	hash.Write(bytes)

	signature := hex.EncodeToString(hash.Sum(nil))
	tripaySignature := req.Header.Get("X-Callback-Signature")
	if signature != tripaySignature {
		logWithCtx.Error().Err(err).Caller().Int("status_code", http.StatusForbidden).Msg("invalid signature")
		http.Error(res, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	var body tripayWebhookRequest
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		logWithCtx.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to unmarshal tripay webhook request")
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	merchantRefBytes, err := uuid.Parse(body.MerchantRef)
	if err != nil {
		logWithCtx.
			Error().
			Err(err).
			Caller().
			Int("status_code", http.StatusInternalServerError).
			Str("transaction_id", body.MerchantRef).
			Msg("failed to parse merchant ref uuid string to bytes")

		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// update transaction status and user subscription
	if body.Status == string(repository.TransactionStatusPAID) {
		tx, err := queries.GetTxWithSubsPlanByID(ctx, pgtype.UUID{Bytes: merchantRefBytes, Valid: true})
		if err != nil {
			logWithCtx.
				Error().
				Err(err).
				Caller().
				Int("status_code", http.StatusInternalServerError).
				Str("transaction_id", body.MerchantRef).
				Msg("failed to get transaction with subscription plan by id")

			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		monthInSecs := time.Hour.Seconds() * 24 * 30
		err = updateTxAndUser(ctx, &updateTxAndUserParams{
			txID:         merchantRefBytes,
			userID:       tx.UserID,
			subsDuration: int64(monthInSecs) * int64(tx.DurationInMonths),
			paidAt:       body.PaidAt,
		})

		if err != nil {
			logWithCtx.
				Error().
				Err(err).
				Caller().
				Int("status_code", http.StatusInternalServerError).
				Str("transaction_id", body.MerchantRef).
				Str("user_id", tx.UserID).
				Msg("failed to update transaction status and user subscription to PREMIUM")

			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	// update transaction status and rollback coupon quota
	if body.Status != string(repository.TransactionStatusPAID) {
		tx, err := queries.GetTxByID(ctx, pgtype.UUID{Bytes: merchantRefBytes, Valid: true})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				logWithCtx.
					Error().
					Err(err).
					Caller().
					Int("status_code", http.StatusNotFound).
					Str("merchant_ref", body.MerchantRef).
					Msg("transaction not found")

				http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			} else {
				logWithCtx.
					Error().
					Err(err).
					Caller().
					Int("status_code", http.StatusInternalServerError).
					Str("transaction_id", body.MerchantRef).
					Msg("failed to get transaction by id")

				http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		// check for unknown status
		txStatus := body.Status
		isNotFailed := body.Status != string(repository.TransactionStatusFAILED)
		isNotExpired := body.Status != string(repository.TransactionStatusEXPIRED)
		isNotRefund := body.Status != string(repository.TransactionStatusREFUND)

		if isNotFailed && isNotExpired && isNotRefund {
			txStatus = string(repository.TransactionStatusFAILED)
			err := errors.New("unknown tripay transaction status")

			logWithCtx.
				Error().
				Err(err).
				Caller().
				Int("status_code", http.StatusBadRequest).
				Str("transaction_id", body.MerchantRef).
				Str("transaction_status", body.Status).
				Send()
		}

		err = updateTxAndRollbackCoupon(ctx, &updateTxAndRollbackCouponParams{
			txID:       merchantRefBytes,
			txStatus:   txStatus,
			couponCode: tx.CouponCode,
		})

		if err != nil {
			logWithCtx.
				Error().
				Err(err).
				Caller().
				Int("status_code", http.StatusInternalServerError).
				Str("transaction_id", body.MerchantRef).
				Str("transaction_status", body.Status).
				Str("coupon_code", tx.CouponCode.String).
				Msg("failed to update transaction status and/or rollback coupon quota")

			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	respBody := struct {
		Status bool `json:"status"`
	}{
		Status: true,
	}

	err = sendJSONSuccessResponse(res, successResponseParams{StatusCode: http.StatusOK, Data: respBody})
	if err != nil {
		logWithCtx.Error().Err(err).Caller().Int("status_code", http.StatusInternalServerError).Msg("failed to send successful response body")
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	logWithCtx.Info().Int("status_code", http.StatusOK).Dur("response_time", time.Since(start)).Msg("request completed")
}
