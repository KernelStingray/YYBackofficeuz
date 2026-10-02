package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/billing/domain"
	"github.com/hextec/yolyolakay-backend/pkg/integration/click"
)

func (u *billingUseCase) ClickPrepare(ctx context.Context, clickTransID, serviceID, merchantTransID, amount, action, signTime, signString string) (map[string]interface{}, error) {
	gatewayInter, ok := u.gateways["click"]
	if !ok {
		slog.Error("Click gateway not found in usecase gateways")
		return map[string]interface{}{"error": -1, "error_note": "System Error"}, fmt.Errorf("click gateway not found")
	}

	clickGateway, ok := gatewayInter.(*click.ClickGateway)
	if !ok {
		slog.Error("Click gateway type assertion failed")
		return map[string]interface{}{"error": -1, "error_note": "System Error"}, fmt.Errorf("click gateway type assertion failed")
	}

	clickTransIDInt, _ := strconv.ParseInt(clickTransID, 10, 64)
	var clickTransIDVal interface{} = clickTransID
	if clickTransIDInt > 0 {
		clickTransIDVal = clickTransIDInt
	}

	// 1. Verify Signature
	if !clickGateway.ValidateClickSign(clickTransID, serviceID, merchantTransID, "", amount, action, signTime, signString) {
		slog.Warn("ClickPrepare: signature check failed",
			"click_trans_id", clickTransID,
			"service_id", serviceID,
			"merchant_trans_id", merchantTransID,
			"amount", amount,
			"action", action,
			"sign_time", signTime,
			"sign_string", signString,
		)
		return map[string]interface{}{"error": -1, "error_note": "SIGN CHECK FAILED"}, fmt.Errorf("invalid signature")
	}

	txID, err := uuid.Parse(merchantTransID)
	if err != nil {
		slog.Warn("ClickPrepare: invalid merchant_trans_id UUID", "merchant_trans_id", merchantTransID, "err", err)
		return map[string]interface{}{"error": -5, "error_note": "User does not exist"}, fmt.Errorf("invalid merchant_trans_id: %w", err)
	}

	tx, err := u.repo.GetTransactionForUpdate(ctx, txID)
	if err != nil || tx == nil {
		slog.Warn("ClickPrepare: transaction not found in database", "merchant_trans_id", merchantTransID, "err", err)
		return map[string]interface{}{"error": -5, "error_note": "User does not exist"}, fmt.Errorf("transaction not found")
	}

	if tx.Status == domain.TransactionStatusSuccess {
		slog.Info("ClickPrepare: transaction already paid", "tx_id", tx.ID, "status", tx.Status)
		return map[string]interface{}{"error": -4, "error_note": "Already paid"}, fmt.Errorf("already paid")
	}

	if tx.Status == domain.TransactionStatusFailed || tx.Status == domain.TransactionStatusCancelled {
		slog.Warn("ClickPrepare: transaction was cancelled or failed", "tx_id", tx.ID, "status", tx.Status)
		return map[string]interface{}{"error": -9, "error_note": "Transaction cancelled"}, fmt.Errorf("transaction cancelled")
	}

	amtFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil || math.Abs(tx.Amount-amtFloat) > 0.01 {
		slog.Warn("ClickPrepare: incorrect amount", "expected", tx.Amount, "received", amount, "click_trans_id", clickTransID)
		return map[string]interface{}{"error": -2, "error_note": "Incorrect parameter amount"}, fmt.Errorf("incorrect amount: expected %f, got %s", tx.Amount, amount)
	}

	slog.Info("ClickPrepare: successfully validated",
		"click_trans_id", clickTransID,
		"merchant_trans_id", merchantTransID,
		"amount", amtFloat,
		"user_id", tx.UserID,
	)

	return map[string]interface{}{
		"click_trans_id":      clickTransIDVal,
		"merchant_trans_id":   merchantTransID,
		"merchant_prepare_id": clickTransIDVal,
		"error":               0,
		"error_note":          "Success",
	}, nil
}

func (u *billingUseCase) ClickComplete(ctx context.Context, clickTransID, serviceID, merchantTransID, merchantPrepareID, amount, action, signTime, signString string) (map[string]interface{}, error) {
	gateway, ok := u.gateways["click"]
	if !ok {
		slog.Error("Click gateway not found in usecase gateways")
		return map[string]interface{}{"error": -1, "error_note": "System Error"}, fmt.Errorf("click gateway not found")
	}

	clickGateway, ok := gateway.(*click.ClickGateway)
	if !ok {
		slog.Error("Click gateway type assertion failed")
		return map[string]interface{}{"error": -1, "error_note": "System Error"}, fmt.Errorf("click gateway type assertion failed")
	}

	clickTransIDInt, _ := strconv.ParseInt(clickTransID, 10, 64)
	var clickTransIDVal interface{} = clickTransID
	if clickTransIDInt > 0 {
		clickTransIDVal = clickTransIDInt
	}

	if !clickGateway.ValidateClickSign(clickTransID, serviceID, merchantTransID, merchantPrepareID, amount, action, signTime, signString) {
		slog.Warn("ClickComplete: signature check failed",
			"click_trans_id", clickTransID,
			"service_id", serviceID,
			"merchant_trans_id", merchantTransID,
			"merchant_prepare_id", merchantPrepareID,
			"amount", amount,
			"action", action,
			"sign_time", signTime,
			"sign_string", signString,
		)
		return map[string]interface{}{"error": -1, "error_note": "SIGN CHECK FAILED"}, fmt.Errorf("invalid signature")
	}

	txID, err := uuid.Parse(merchantTransID)
	if err != nil {
		slog.Warn("ClickComplete: invalid merchant_trans_id UUID", "merchant_trans_id", merchantTransID, "err", err)
		return map[string]interface{}{"error": -5, "error_note": "User does not exist"}, fmt.Errorf("invalid merchant_trans_id")
	}

	tx, err := u.repo.GetTransactionForUpdate(ctx, txID)
	if err != nil || tx == nil {
		slog.Warn("ClickComplete: transaction not found in database", "merchant_trans_id", merchantTransID, "err", err)
		return map[string]interface{}{"error": -5, "error_note": "User does not exist"}, fmt.Errorf("transaction not found")
	}

	if tx.Status == domain.TransactionStatusSuccess {
		slog.Info("ClickComplete: transaction already completed earlier", "tx_id", tx.ID)
		return map[string]interface{}{
			"click_trans_id":      clickTransIDVal,
			"merchant_trans_id":   merchantTransID,
			"merchant_confirm_id": clickTransIDVal,
			"error":               -4,
			"error_note":          "Already paid",
		}, nil
	}

	amtFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil || math.Abs(tx.Amount-amtFloat) > 0.01 {
		slog.Warn("ClickComplete: incorrect amount", "expected", tx.Amount, "received", amount, "click_trans_id", clickTransID)
		return map[string]interface{}{"error": -2, "error_note": "Incorrect parameter amount"}, fmt.Errorf("incorrect amount")
	}

	// Determine plan and quotas
	planName := "plus"
	if strings.HasPrefix(tx.Type, "subscription:") {
		planName = strings.TrimPrefix(tx.Type, "subscription:")
	}

	ridesLeft := int32(2)
	freeParcels := int32(2)
	switch planName {
	case "plus":
		ridesLeft = 2
		freeParcels = 2
	case "pro":
		ridesLeft = 5
		freeParcels = 10
	case "premium":
		ridesLeft = -1
		freeParcels = -1
	}

	err = u.repo.ProcessWebhookTx(ctx, txID, clickTransID, amtFloat, planName, ridesLeft, freeParcels)
	if err != nil {
		slog.Error("ClickComplete: ProcessWebhookTx failed", "tx_id", txID, "err", err)
		return map[string]interface{}{"error": -1, "error_note": "System Error"}, err
	}

	slog.Info("ClickComplete: successfully processed and subscription activated",
		"driver_id", tx.UserID,
		"plan", planName,
		"amount", amtFloat,
		"click_trans_id", clickTransID,
		"merchant_trans_id", merchantTransID,
	)

	// Publish event to RabbitMQ
	eventPayload, _ := json.Marshal(map[string]interface{}{
		"driver_id": tx.UserID.String(),
		"plan":      planName,
		"amount":    amtFloat,
		"event":     "subscription_activated",
	})
	if u.rmqClient != nil {
		if err := u.rmqClient.Publish(ctx, "billing.events", "billing.subscription.activated", eventPayload); err != nil {
			slog.Error("Failed to publish billing.subscription.activated event", "error", err)
		}
	}

	return map[string]interface{}{
		"click_trans_id":      clickTransIDVal,
		"merchant_trans_id":   merchantTransID,
		"merchant_confirm_id": clickTransIDVal,
		"error":               0,
		"error_note":          "Success",
	}, nil
}
