package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type BillingRepository interface {
	GetAdminStats(ctx context.Context) (*AdminStats, error)
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetActiveSubscription(ctx context.Context, driverID uuid.UUID) (*Subscription, error)
	GetLatestSubscription(ctx context.Context, driverID uuid.UUID) (*Subscription, error)
	CheckHasUsedTrial(ctx context.Context, driverID uuid.UUID) (bool, error)

	SavePaymentToken(ctx context.Context, pt *PaymentToken) error
	GetPaymentTokens(ctx context.Context, userID uuid.UUID) ([]PaymentToken, error)

	CreateTransaction(ctx context.Context, tx *Transaction) error
	UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status TransactionStatus, externalTxID *string) error
	GetTransactionForUpdate(ctx context.Context, txID uuid.UUID) (*Transaction, error)

	GetSubscriptionsForBilling(ctx context.Context) ([]Subscription, error)
	UpdateSubscriptionStatus(ctx context.Context, subID uuid.UUID, status SubscriptionStatus, endDate time.Time) error

	ApplyPromoCodeTx(ctx context.Context, userID uuid.UUID, code string) (*PromoCode, error)
	DeductRideLimit(ctx context.Context, driverID uuid.UUID) error
	RefundRideLimit(ctx context.Context, driverID uuid.UUID) error
	AddFreeParcelsToPremium(ctx context.Context, quota int32) error
	CancelStaleTransactions(ctx context.Context) error
	ResetExpiredSubscriptions(ctx context.Context) (int64, error)
	DeductParcelLimit(ctx context.Context, driverID uuid.UUID) error
	RefundParcelLimit(ctx context.Context, driverID uuid.UUID) error
	UpdateActiveSubscriptionLimits(ctx context.Context, subID uuid.UUID, planName string, rides int32, parcels int32, endDate time.Time) error
	CancelActiveSubscription(ctx context.Context, driverID uuid.UUID) error
	ProcessWebhookTx(ctx context.Context, txID uuid.UUID, externalTxID string, amount float64, planName string, ridesLeft int32, freeParcels int32) error
}
