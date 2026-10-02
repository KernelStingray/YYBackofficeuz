package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLimitExceeded = errors.New("limit exceeded")
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
)

type Subscription struct {
	ID               uuid.UUID          `json:"id"`
	UserID           uuid.UUID          `json:"user_id"`
	PlanKey          string             `json:"plan_key"`
	PlanName         string             `json:"plan_name"`
	Status           SubscriptionStatus `json:"status" swaggertype:"string"`
	StartDate        time.Time          `json:"start_date"`
	EndDate          time.Time          `json:"end_date"`
	HasUsedTrial     bool               `json:"has_used_trial"`
	RidesLeft        int32              `json:"rides_left"`
	FreeParcelsQuota int32              `json:"free_parcels_quota"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

type PaymentToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Provider  string    `json:"provider"`
	Token     string    `json:"token"`
	MaskedPan string    `json:"masked_pan"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusSuccess   TransactionStatus = "success"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

type Transaction struct {
	ID             uuid.UUID         `json:"id"`
	UserID         uuid.UUID         `json:"user_id"`
	Amount         float64           `json:"amount"`
	Currency       string            `json:"currency"`
	Type           string            `json:"type"` // "subscription", "ride_fee"
	Status         TransactionStatus `json:"status" swaggertype:"string"`
	Provider       string            `json:"provider"`
	PaymentTokenID *uuid.UUID        `json:"payment_token_id,omitempty"`
	ExternalTxID   *string           `json:"external_tx_id,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type PromoCode struct {
	ID              uuid.UUID  `json:"id"`
	Code            string     `json:"code"`
	DiscountAmount  *float64   `json:"discount_amount,omitempty"`
	DiscountPercent *float64   `json:"discount_percent,omitempty"`
	MaxUses         int32      `json:"max_uses"`
	CurrentUses     int32      `json:"current_uses"`
	IsActive        bool       `json:"is_active"`
	DiscountType    string     `json:"discount_type"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type UserPromoUsage struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	PromoCodeID uuid.UUID `json:"promo_code_id"`
	UsedAt      time.Time `json:"used_at"`
}

type AdminStats struct {
	TotalRevenue      float64          `json:"total_revenue"`
	ChartData         []ChartDataPoint `json:"chart_data"`
}

type ChartDataPoint struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}
