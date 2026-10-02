package domain

import (
	"context"

	"github.com/google/uuid"
)

type BillingUseCase interface {
	GetAdminStats(ctx context.Context) (*AdminStats, error)
	SubscribeDriver(ctx context.Context, driverID uuid.UUID, planName string, paymentTokenID *uuid.UUID, promoCode *string, returnURL *string) (string, uuid.UUID, error)
	GetTransaction(ctx context.Context, userID, txID uuid.UUID) (*Transaction, error)
	ProcessWebhook(ctx context.Context, provider string, payload []byte, signature string, externalTxID string, txID uuid.UUID, amount float64) error
	
	// Click integration
	ClickPrepare(ctx context.Context, clickTransID, serviceID, merchantTransID, amount, action, signTime, signString string) (map[string]interface{}, error)
	ClickComplete(ctx context.Context, clickTransID, serviceID, merchantTransID, merchantPrepareID, amount, action, signTime, signString string) (map[string]interface{}, error)

	StartTrialSubscription(ctx context.Context, driverID uuid.UUID, trialType string) error
	GetDriverSubscription(ctx context.Context, driverID uuid.UUID) (*Subscription, error)

	AddPaymentMethod(ctx context.Context, userID uuid.UUID, provider string, token string, maskedPan string) (*PaymentToken, error)
	GetPaymentMethods(ctx context.Context, userID uuid.UUID) ([]PaymentToken, error)

	CheckRideLimit(ctx context.Context, driverID uuid.UUID) (bool, string, error)
	DeductParcelLimit(ctx context.Context, driverID uuid.UUID) (bool, error)
	ExpireTrialSubscription(ctx context.Context, driverID uuid.UUID) error
}
