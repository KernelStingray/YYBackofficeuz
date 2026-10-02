package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/billing/domain"
	"github.com/hextec/yolyolakay-backend/internal/billing/integration"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
)

type billingUseCase struct {
	repo      domain.BillingRepository
	rmqClient *rabbitmq.Client
	gateways  map[string]integration.PaymentGateway
}

func NewBillingUseCase(repo domain.BillingRepository, rmqClient *rabbitmq.Client, gateways map[string]integration.PaymentGateway) domain.BillingUseCase {
	if rmqClient != nil {
		_ = rmqClient.SetupExchangeQueue("billing.events", "topic", "billing.notifications", "billing.promo.*")
		_ = rmqClient.SetupExchangeQueue("billing.events", "topic", "billing.notifications", "billing.subscription.*")
	}
	return &billingUseCase{
		repo:      repo,
		rmqClient: rmqClient,
		gateways:  gateways,
	}
}

func (u *billingUseCase) GetAdminStats(ctx context.Context) (*domain.AdminStats, error) {
	return u.repo.GetAdminStats(ctx)
}

func (u *billingUseCase) SubscribeDriver(ctx context.Context, driverID uuid.UUID, planName string, paymentTokenID *uuid.UUID, promoCode *string, returnURL *string) (string, uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var price float64
	switch strings.ToLower(planName) {
	case "plus", "driver_plus":
		price = 5000
	case "pro", "driver_pro":
		price = 7000
	case "premium", "driver_premium":
		price = 10000
	default:
		return "", uuid.Nil, errors.New("invalid plan name")
	}

	provider := "click" // Default for now
	gateway, ok := u.gateways[provider]
	if !ok {
		return "", uuid.Nil, errors.New("unsupported payment provider")
	}

	if promoCode != nil && *promoCode != "" {
		promo, err := u.repo.ApplyPromoCodeTx(ctx, driverID, *promoCode)
		if err != nil {
			return "", uuid.Nil, fmt.Errorf("promo code error: %w", err)
		}

		// Apply discount
		if promo.DiscountType == "fixed_amount" && promo.DiscountAmount != nil {
			price -= *promo.DiscountAmount
		} else if promo.DiscountType == "percentage" && promo.DiscountPercent != nil {
			price -= price * (*promo.DiscountPercent / 100.0)
		}
		if price < 0 {
			price = 0
		}

		// Send event
		eventPayload, _ := json.Marshal(map[string]interface{}{
			"user_id":       driverID.String(),
			"promo_code_id": promo.ID.String(),
			"code":          promo.Code,
			"event":         "promo_code_applied",
		})
		if u.rmqClient != nil {
			err = u.rmqClient.Publish(ctx, "billing.events", "billing.promo.applied", eventPayload)
			if err != nil {
				_ = err // log error
			}
		}
	}

	tx := &domain.Transaction{
		ID:             uuid.New(),
		UserID:         driverID,
		Amount:         price,
		Currency:       "UZS",
		Type:           fmt.Sprintf("subscription:%s", planName),
		Status:         domain.TransactionStatusPending,
		Provider:       provider,
		PaymentTokenID: paymentTokenID,
	}

	if price == 0 {
		tx.Status = domain.TransactionStatusSuccess
	}

	err := u.repo.CreateTransaction(ctx, tx)
	if err != nil {
		return "", uuid.Nil, err
	}

	if price == 0 {
		// Activate immediately
		err = u.repo.UpdateTransactionStatus(ctx, tx.ID, domain.TransactionStatusSuccess, nil)
		if err != nil {
			return "", uuid.Nil, err
		}
		err = u.activateSubscription(ctx, tx.UserID, tx.Amount, planName)
		if err != nil {
			return "", uuid.Nil, err
		}
		return "success", tx.ID, nil
	}

	var retURL string
	if returnURL != nil {
		retURL = *returnURL
	}
	paymentLink, err := gateway.GeneratePaymentLink(tx.ID, price, retURL)
	if err != nil {
		return "", uuid.Nil, err
	}

	// [Mock Mode] Auto-simulate a successful webhook callback immediately
	if strings.Contains(paymentLink, "mock.") {
		go func(txID uuid.UUID, provider string, amount float64) {
			// We skip HTTP and directly update DB to simulate the webhook's end result
			ctx := context.Background()
			err1 := u.repo.ProcessWebhookTx(ctx, txID, "mock-ext-id", amount, planName, -1, 10)
			if err1 != nil {
				slog.Error("Mock webhook ProcessWebhookTx failed", "error", err1)
			}
		}(tx.ID, provider, price)
	}

	return paymentLink, tx.ID, nil
}

func (u *billingUseCase) GetTransaction(ctx context.Context, userID, txID uuid.UUID) (*domain.Transaction, error) {
	tx, err := u.repo.GetTransactionForUpdate(ctx, txID)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, errors.New("transaction not found")
	}
	if tx.UserID != userID {
		return nil, errors.New("unauthorized: transaction belongs to another user")
	}
	return tx, nil
}

func (u *billingUseCase) ProcessWebhook(ctx context.Context, provider string, payload []byte, signature string, externalTxID string, txID uuid.UUID, amount float64) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gateway, ok := u.gateways[provider]
	if !ok {
		return errors.New("unsupported payment provider")
	}

	if !gateway.VerifySignature(payload, signature) {
		return errors.New("invalid signature")
	}

	tx, err := u.repo.GetTransactionForUpdate(ctx, txID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	planName := "plus"
	if strings.HasPrefix(tx.Type, "subscription:") {
		planName = strings.TrimPrefix(tx.Type, "subscription:")
	}

	ridesLeft := int32(2)
	freeParcels := int32(2)

	switch strings.ToLower(planName) {
	case "plus", "driver_plus":
		ridesLeft = 2
		freeParcels = 2
	case "pro", "driver_pro":
		ridesLeft = 5
		freeParcels = 5
	case "premium", "driver_premium":
		ridesLeft = 60
		freeParcels = -1
	}

	err = u.repo.ProcessWebhookTx(ctx, txID, externalTxID, amount, planName, ridesLeft, freeParcels)
	if err != nil {
		return err
	}

	// Publish event to RabbitMQ
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"driver_id": tx.UserID.String(),
		"plan":      planName,
		"amount":    amount,
		"event":     "subscription_activated",
	})
	if u.rmqClient != nil {
		if err := u.rmqClient.Publish(ctx, "billing.events", "billing.subscription.activated", eventPayload); err != nil {
			slog.Error("Failed to publish", "error", err)
		}
	}

	return nil
}

func (u *billingUseCase) activateSubscription(ctx context.Context, userID uuid.UUID, amount float64, planName string) error {
	ridesLeft := int32(2)
	freeParcels := int32(2)

	// Base defaults on plan name instead of amount to support discounted amounts
	switch strings.ToLower(planName) {
	case "plus", "driver_plus":
		ridesLeft = 2
		freeParcels = 2
	case "pro", "driver_pro":
		ridesLeft = 5
		freeParcels = 5
	case "premium", "driver_premium":
		ridesLeft = 60
		freeParcels = -1
	}

	// Check for active subscription
	activeSub, _ := u.repo.GetActiveSubscription(ctx, userID)

	var err error
	if activeSub != nil {
		// Cancel existing subscription
		err = u.repo.CancelActiveSubscription(ctx, userID)
		if err != nil {
			return err
		}
	}

	// Create new subscription
	sub := &domain.Subscription{
		ID:               uuid.New(),
		UserID:           userID,
		PlanName:         planName,
		Status:           domain.SubscriptionStatusActive,
		StartDate:        time.Now(),
		EndDate:          time.Now().AddDate(0, 1, 0),
		RidesLeft:        ridesLeft,
		FreeParcelsQuota: freeParcels,
	}

	err = u.repo.CreateSubscription(ctx, sub)
	if err != nil {
		return err
	}

	// Publish event to RabbitMQ
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"driver_id": userID.String(),
		"plan":      planName,
		"amount":    amount,
		"event":     "subscription_activated",
	})
	if u.rmqClient != nil {
		err = u.rmqClient.Publish(ctx, "billing.events", "billing.subscription.activated", eventPayload)
		if err != nil {
			slog.Error("Failed to publish event", "error", err)
		}
	}

	return nil
}

func (u *billingUseCase) StartTrialSubscription(ctx context.Context, driverID uuid.UUID, trialType string) error {
	// Check if already used trial
	hasUsed, err := u.repo.CheckHasUsedTrial(ctx, driverID)
	if err != nil {
		return err
	}
	if hasUsed {
		return errors.New("trial already used") // silently ignore or return error, but domain logic says they can't have it again
	}

	// Check if active subscription exists
	activeSub, _ := u.repo.GetActiveSubscription(ctx, driverID)
	if activeSub != nil {
		return errors.New("active subscription already exists") // already has an active subscription
	}

	ridesLeft := int32(0)
	freeParcels := int32(0)

	switch trialType {
	case "ride":
		ridesLeft = 1
		freeParcels = 1
	case "parcel":
		ridesLeft = 1
		freeParcels = 1
	default:
		return errors.New("invalid trial type, must be 'ride' or 'parcel'")
	}

	sub := &domain.Subscription{
		ID:               uuid.New(),
		UserID:           driverID,
		PlanKey:          "driver_trial",
		PlanName:         fmt.Sprintf("trial_%s", trialType),
		Status:           domain.SubscriptionStatusActive,
		StartDate:        time.Now(),
		EndDate:          time.Now().AddDate(0, 1, 0), // 30 days trial period
		HasUsedTrial:     true,
		RidesLeft:        ridesLeft,
		FreeParcelsQuota: freeParcels,
	}

	err = u.repo.CreateSubscription(ctx, sub)
	if err != nil {
		return err
	}

	// Publish event
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"driver_id": driverID.String(),
		"plan":      "trial",
		"event":     "trial_started",
	})
	if u.rmqClient != nil {
		if err := u.rmqClient.Publish(ctx, "billing.events", "billing.trial.started", eventPayload); err != nil {
			slog.Error("Failed to publish", "error", err)
		}
	}

	return nil
}

func (u *billingUseCase) ExpireTrialSubscription(ctx context.Context, driverID uuid.UUID) error {
	activeSub, err := u.repo.GetActiveSubscription(ctx, driverID)
	if err != nil {
		return err
	}
	if activeSub == nil || !strings.HasPrefix(activeSub.PlanName, "trial") {
		return nil
	}

	err = u.repo.UpdateActiveSubscriptionLimits(ctx, activeSub.ID, "none", 0, 0, time.Now())
	if err != nil {
		return fmt.Errorf("failed to expire trial subscription: %w", err)
	}

	eventPayload, _ := json.Marshal(map[string]interface{}{
		"driver_id": driverID.String(),
		"plan":      "trial",
		"event":     "subscription_expired",
	})
	if u.rmqClient != nil {
		if err := u.rmqClient.Publish(ctx, "billing.events", "billing.subscription.expired", eventPayload); err != nil {
			slog.Error("Failed to publish", "error", err)
		}
	}

	return nil
}

func (u *billingUseCase) GetDriverSubscription(ctx context.Context, driverID uuid.UUID) (*domain.Subscription, error) {
	sub, err := u.repo.GetLatestSubscription(ctx, driverID)
	if err != nil {
		return nil, err
	}

	if sub != nil {
		if sub.EndDate.Before(time.Now()) && sub.Status == domain.SubscriptionStatusActive {
			sub.Status = domain.SubscriptionStatusExpired
		}
		// Unified plan_key and display plan_name
		switch {
		case strings.HasPrefix(sub.PlanName, "trial") || sub.PlanName == "free_trial" || sub.PlanName == "Trial ride" || sub.PlanName == "driver_trial":
			sub.PlanKey = "driver_trial"
			sub.PlanName = "Free Trial"
		case strings.EqualFold(sub.PlanName, "plus") || sub.PlanName == "driver_plus":
			sub.PlanKey = "driver_plus"
			sub.PlanName = "Plus"
		case strings.EqualFold(sub.PlanName, "pro") || sub.PlanName == "driver_pro":
			sub.PlanKey = "driver_pro"
			sub.PlanName = "Pro"
		case strings.EqualFold(sub.PlanName, "premium") || sub.PlanName == "driver_premium":
			sub.PlanKey = "driver_premium"
			sub.PlanName = "Premium"
		default:
			sub.PlanKey = strings.ToLower(sub.PlanName)
		}
	}
	return sub, nil
}

func (u *billingUseCase) AddPaymentMethod(ctx context.Context, userID uuid.UUID, provider string, token string, maskedPan string) (*domain.PaymentToken, error) {
	pt := &domain.PaymentToken{
		ID:        uuid.New(),
		UserID:    userID,
		Provider:  provider,
		Token:     token,
		MaskedPan: maskedPan,
		Status:    "active",
	}

	err := u.repo.SavePaymentToken(ctx, pt)
	if err != nil {
		return nil, err
	}
	return pt, nil
}

func (u *billingUseCase) DeductParcelLimit(ctx context.Context, driverID uuid.UUID) (bool, error) {
	err := u.repo.DeductParcelLimit(ctx, driverID)
	if err != nil {
		return false, err // or return false, nil if we want to handle it as "no quota"
	}
	return true, nil
}

func (u *billingUseCase) GetPaymentMethods(ctx context.Context, userID uuid.UUID) ([]domain.PaymentToken, error) {
	return u.repo.GetPaymentTokens(ctx, userID)
}

// CheckRideLimit checks if a driver has an active subscription with remaining rides
func (u *billingUseCase) CheckRideLimit(ctx context.Context, driverID uuid.UUID) (bool, string, error) {
	// Let's use the repository to get the active subscription
	sub, err := u.repo.GetActiveSubscription(ctx, driverID)
	if err != nil {
		return false, "", err
	}

	if sub == nil || sub.Status != domain.SubscriptionStatusActive {
		return false, "free", nil
	}

	if sub.RidesLeft == 0 {
		return false, sub.PlanName, nil
	}

	// For simplicity in the mock logic, we return true if sub != nil, active, and has rides left (or unlimited).
	return true, sub.PlanName, nil
}
