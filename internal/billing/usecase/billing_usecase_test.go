package usecase

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/billing/domain"
	"github.com/hextec/yolyolakay-backend/internal/billing/integration"
	"github.com/hextec/yolyolakay-backend/pkg/integration/click"
	billingMocks "github.com/hextec/yolyolakay-backend/pkg/mocks/billing"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupBillingUseCaseTest(t *testing.T) (domain.BillingUseCase, *billingMocks.MockBillingRepository, *click.MockClickGateway) {
	ctrl := gomock.NewController(t)

	repo := billingMocks.NewMockBillingRepository(ctrl)
	gw := click.NewMockClickGateway("secret")
	gateways := map[string]integration.PaymentGateway{
		"click": gw,
	}

	var rmqClient *rabbitmq.Client = nil

	uc := NewBillingUseCase(repo, rmqClient, gateways)

	return uc, repo, gw
}

func TestStartTrialSubscription_Success(t *testing.T) {
	uc, repo, _ := setupBillingUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()

	repo.EXPECT().CheckHasUsedTrial(ctx, driverID).Return(false, nil)
	repo.EXPECT().GetActiveSubscription(ctx, driverID).Return(nil, nil)
	repo.EXPECT().CreateSubscription(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, sub *domain.Subscription) error {
		assert.Equal(t, "trial_ride", sub.PlanName)
		assert.Equal(t, domain.SubscriptionStatusActive, sub.Status)
		// Trial is 30 days
		assert.True(t, sub.EndDate.After(time.Now().AddDate(0, 0, 20)))
		return nil
	})

	err := uc.StartTrialSubscription(ctx, driverID, "ride")

	require.NoError(t, err)
}

func TestStartTrialSubscription_AlreadyUsed(t *testing.T) {
	uc, repo, _ := setupBillingUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()

	repo.EXPECT().CheckHasUsedTrial(ctx, driverID).Return(true, nil)

	err := uc.StartTrialSubscription(ctx, driverID, "ride")

	require.EqualError(t, err, "trial already used")
}

func TestSubscribeDriver_Paid_WithClick(t *testing.T) {
	uc, repo, _ := setupBillingUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()

	repo.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, tx *domain.Transaction) error {
		tx.ID = uuid.New()
		assert.Equal(t, 10000.0, tx.Amount) // premium
		return nil
	})
	repo.EXPECT().ProcessWebhookTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	result, txID, err := uc.SubscribeDriver(ctx, driverID, "premium", nil, nil, nil)

	require.NoError(t, err)
	assert.Contains(t, result, "https://mock.click.uz/pay?tx_id=")
	assert.NotEqual(t, uuid.Nil, txID)
}

func TestSubscribeDriver_DriverPlusKey(t *testing.T) {
	uc, repo, _ := setupBillingUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()

	repo.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, tx *domain.Transaction) error {
		tx.ID = uuid.New()
		assert.Equal(t, 5000.0, tx.Amount) // plus
		return nil
	})
	repo.EXPECT().ProcessWebhookTx(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	result, txID, err := uc.SubscribeDriver(ctx, driverID, "driver_plus", nil, nil, nil)

	require.NoError(t, err)
	assert.Contains(t, result, "https://mock.click.uz/pay?tx_id=")
	assert.NotEqual(t, uuid.Nil, txID)
}

func TestSubscribeDriver_CustomReturnURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := billingMocks.NewMockBillingRepository(ctrl)
	realGw := click.NewClickGateway("63396", "108681", "secret")
	gateways := map[string]integration.PaymentGateway{
		"click": realGw,
	}
	uc := NewBillingUseCase(repo, nil, gateways)
	ctx := context.Background()
	driverID := uuid.New()

	repo.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Return(nil)

	customReturnURL := "https://yolyolakay.com/uz/payment"
	result, txID, err := uc.SubscribeDriver(ctx, driverID, "plus", nil, nil, &customReturnURL)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, txID)
	assert.Contains(t, result, "return_url=https%3A%2F%2Fyolyolakay.com%2Fuz%2Fpayment")
}

func TestGetTransaction(t *testing.T) {
	uc, repo, _ := setupBillingUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()
	txID := uuid.New()
	tx := &domain.Transaction{
		ID:       txID,
		UserID:   driverID,
		Amount:   5000,
		Currency: "UZS",
		Status:   domain.TransactionStatusSuccess,
	}

	// 1. Success case
	repo.EXPECT().GetTransactionForUpdate(ctx, txID).Return(tx, nil)
	fetchedTx, err := uc.GetTransaction(ctx, driverID, txID)
	require.NoError(t, err)
	assert.Equal(t, txID, fetchedTx.ID)

	// 2. Unauthorized case (other user tries to view)
	otherUser := uuid.New()
	repo.EXPECT().GetTransactionForUpdate(ctx, txID).Return(tx, nil)
	_, errUnauthorized := uc.GetTransaction(ctx, otherUser, txID)
	require.Error(t, errUnauthorized)
	assert.Contains(t, errUnauthorized.Error(), "unauthorized")

	// 3. Not found case
	missingTxID := uuid.New()
	repo.EXPECT().GetTransactionForUpdate(ctx, missingTxID).Return(nil, nil)
	_, errNotFound := uc.GetTransaction(ctx, driverID, missingTxID)
	require.Error(t, errNotFound)
	assert.Contains(t, errNotFound.Error(), "not found")
}

func TestClickPrepareAndComplete(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := billingMocks.NewMockBillingRepository(ctrl)
	secret := "test-secret"
	serviceID := "108681"
	realGw := click.NewClickGateway("63396", serviceID, secret)
	gateways := map[string]integration.PaymentGateway{
		"click": realGw,
	}
	uc := NewBillingUseCase(repo, nil, gateways)
	ctx := context.Background()

	txID := uuid.New()
	driverID := uuid.New()
	tx := &domain.Transaction{
		ID:       txID,
		UserID:   driverID,
		Amount:   5000,
		Currency: "UZS",
		Type:     "subscription:plus",
		Status:   domain.TransactionStatusPending,
		Provider: "click",
	}

	clickTransID := "999888777"
	amount := "5000.00"
	signTime := "2026-09-14 12:00:00"

	// Prepare signature: click_trans_id + service_id + secret + merchant_trans_id + amount + action + sign_time
	prepareRaw := fmt.Sprintf("%s%s%s%s%s%s%s", clickTransID, serviceID, secret, txID.String(), amount, "0", signTime)
	h0 := md5.Sum([]byte(prepareRaw))
	sign0 := hex.EncodeToString(h0[:])

	// 1. Prepare with valid sign
	repo.EXPECT().GetTransactionForUpdate(gomock.Any(), txID).Return(tx, nil).Times(1)
	resp, err := uc.ClickPrepare(ctx, clickTransID, serviceID, txID.String(), amount, "0", signTime, sign0)
	require.NoError(t, err)
	assert.Equal(t, 0, resp["error"])
	assert.Equal(t, int64(999888777), resp["click_trans_id"])

	// 2. Prepare with invalid sign
	respBad, errBad := uc.ClickPrepare(ctx, clickTransID, serviceID, txID.String(), amount, "0", signTime, "invalid_sign")
	require.Error(t, errBad)
	assert.Equal(t, -1, respBad["error"])

	// 3. Complete signature: click_trans_id + service_id + secret + merchant_trans_id + merchant_prepare_id + amount + action + sign_time
	merchantPrepareID := clickTransID
	completeRaw := fmt.Sprintf("%s%s%s%s%s%s%s%s", clickTransID, serviceID, secret, txID.String(), merchantPrepareID, amount, "1", signTime)
	h1 := md5.Sum([]byte(completeRaw))
	sign1 := hex.EncodeToString(h1[:])

	repo.EXPECT().GetTransactionForUpdate(gomock.Any(), txID).Return(tx, nil).Times(1)
	repo.EXPECT().ProcessWebhookTx(gomock.Any(), txID, clickTransID, 5000.0, "plus", int32(2), int32(2)).Return(nil).Times(1)

	respComp, errComp := uc.ClickComplete(ctx, clickTransID, serviceID, txID.String(), merchantPrepareID, amount, "1", signTime, sign1)
	require.NoError(t, errComp)
	assert.Equal(t, 0, respComp["error"])
	assert.Equal(t, int64(999888777), respComp["click_trans_id"])
}

