package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hextec/yolyolakay-backend/internal/billing/domain"
	sqlc "github.com/hextec/yolyolakay-backend/internal/billing/repository/postgres"
	"github.com/hextec/yolyolakay-backend/pkg/database"
)

type postgresBillingRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewPostgresBillingRepository(db *pgxpool.Pool) domain.BillingRepository {
	return &postgresBillingRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return database.UUIDToPg(id)
}

func pgToUUID(id pgtype.UUID) uuid.UUID {
	return database.PgToUUID(id)
}

func (r *postgresBillingRepository) GetAdminStats(ctx context.Context) (*domain.AdminStats, error) {
	totalRevenue, err := r.queries.GetTotalRevenue(ctx)
	if err != nil {
		return nil, err
	}

	chartPG, err := r.queries.GetRevenueChartData(ctx)
	if err != nil {
		return nil, err
	}

	var chartData []domain.ChartDataPoint
	for _, c := range chartPG {
		chartData = append(chartData, domain.ChartDataPoint{
			Date:   c.Date,
			Amount: c.Amount,
		})
	}

	return &domain.AdminStats{
		TotalRevenue: totalRevenue,
		ChartData:    chartData,
	}, nil
}

func (r *postgresBillingRepository) CreateSubscription(ctx context.Context, sub *domain.Subscription) error {
	pgSub, err := r.queries.CreateSubscription(ctx, sqlc.CreateSubscriptionParams{
		ID:               uuidToPg(sub.ID),
		UserID:           uuidToPg(sub.UserID),
		PlanName:         sub.PlanName,
		Status:           string(sub.Status),
		StartDate:        pgtype.Timestamptz{Time: sub.StartDate, Valid: true},
		EndDate:          pgtype.Timestamptz{Time: sub.EndDate, Valid: true},
		HasUsedTrial:     sub.HasUsedTrial,
		RidesLeft:        sub.RidesLeft,
		FreeParcelsQuota: sub.FreeParcelsQuota,
	})
	if err != nil {
		return err
	}

	sub.CreatedAt = pgSub.CreatedAt.Time
	sub.UpdatedAt = pgSub.UpdatedAt.Time
	return nil
}

func (r *postgresBillingRepository) GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	pgSub, err := r.queries.GetActiveSubscription(ctx, uuidToPg(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.Subscription{
		ID:               pgToUUID(pgSub.ID),
		UserID:           pgToUUID(pgSub.UserID),
		PlanName:         pgSub.PlanName,
		Status:           domain.SubscriptionStatus(pgSub.Status),
		StartDate:        pgSub.StartDate.Time,
		EndDate:          pgSub.EndDate.Time,
		HasUsedTrial:     pgSub.HasUsedTrial,
		RidesLeft:        pgSub.RidesLeft,
		FreeParcelsQuota: pgSub.FreeParcelsQuota,
		CreatedAt:        pgSub.CreatedAt.Time,
		UpdatedAt:        pgSub.UpdatedAt.Time,
	}, nil
}

func (r *postgresBillingRepository) GetLatestSubscription(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	pgSub, err := r.queries.GetLatestSubscription(ctx, uuidToPg(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &domain.Subscription{
		ID:               pgToUUID(pgSub.ID),
		UserID:           pgToUUID(pgSub.UserID),
		PlanName:         pgSub.PlanName,
		Status:           domain.SubscriptionStatus(pgSub.Status),
		StartDate:        pgSub.StartDate.Time,
		EndDate:          pgSub.EndDate.Time,
		HasUsedTrial:     pgSub.HasUsedTrial,
		RidesLeft:        pgSub.RidesLeft,
		FreeParcelsQuota: pgSub.FreeParcelsQuota,
		CreatedAt:        pgSub.CreatedAt.Time,
		UpdatedAt:        pgSub.UpdatedAt.Time,
	}, nil
}

func (r *postgresBillingRepository) CheckHasUsedTrial(ctx context.Context, userID uuid.UUID) (bool, error) {
	exists, err := r.queries.CheckHasUsedTrial(ctx, uuidToPg(userID))
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *postgresBillingRepository) SavePaymentToken(ctx context.Context, pt *domain.PaymentToken) error {
	pgToken, err := r.queries.SavePaymentToken(ctx, sqlc.SavePaymentTokenParams{
		ID:        uuidToPg(pt.ID),
		UserID:    uuidToPg(pt.UserID),
		Provider:  pt.Provider,
		Token:     pt.Token,
		MaskedPan: pt.MaskedPan,
		Status:    pt.Status,
	})
	if err != nil {
		return err
	}

	pt.CreatedAt = pgToken.CreatedAt.Time
	pt.UpdatedAt = pgToken.UpdatedAt.Time
	return nil
}

func (r *postgresBillingRepository) GetPaymentTokens(ctx context.Context, userID uuid.UUID) ([]domain.PaymentToken, error) {
	pgTokens, err := r.queries.GetPaymentTokensByUserID(ctx, uuidToPg(userID))
	if err != nil {
		return nil, err
	}

	var tokens []domain.PaymentToken
	for _, pt := range pgTokens {
		tokens = append(tokens, domain.PaymentToken{
			ID:        pgToUUID(pt.ID),
			UserID:    pgToUUID(pt.UserID),
			Provider:  pt.Provider,
			Token:     pt.Token,
			MaskedPan: pt.MaskedPan,
			Status:    pt.Status,
			CreatedAt: pt.CreatedAt.Time,
			UpdatedAt: pt.UpdatedAt.Time,
		})
	}
	return tokens, nil
}

func (r *postgresBillingRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	var num pgtype.Numeric
	num.Scan(fmt.Sprintf("%f", tx.Amount))

	var pgTokenID pgtype.UUID
	if tx.PaymentTokenID != nil {
		pgTokenID = uuidToPg(*tx.PaymentTokenID)
	}

	var extTxID pgtype.Text
	if tx.ExternalTxID != nil {
		extTxID = pgtype.Text{String: *tx.ExternalTxID, Valid: true}
	}

	pgTx, err := r.queries.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		ID:             uuidToPg(tx.ID),
		UserID:         uuidToPg(tx.UserID),
		Amount:         num,
		Currency:       tx.Currency,
		Type:           tx.Type,
		Status:         string(tx.Status),
		Provider:       tx.Provider,
		PaymentTokenID: pgTokenID,
		ExternalTxID:   extTxID,
	})
	if err != nil {
		return err
	}

	tx.CreatedAt = pgTx.CreatedAt.Time
	tx.UpdatedAt = pgTx.UpdatedAt.Time
	return nil
}

func (r *postgresBillingRepository) UpdateTransactionStatus(ctx context.Context, txID uuid.UUID, status domain.TransactionStatus, externalTxID *string) error {
	var extTxID pgtype.Text
	if externalTxID != nil {
		extTxID = pgtype.Text{String: *externalTxID, Valid: true}
	}
	return r.queries.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
		Status:       string(status),
		ExternalTxID: extTxID,
		ID:           uuidToPg(txID),
	})
}

func (r *postgresBillingRepository) GetTransactionForUpdate(ctx context.Context, txID uuid.UUID) (*domain.Transaction, error) {
	pgTx, err := r.queries.GetTransactionForUpdate(ctx, uuidToPg(txID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var amount float64
	val, _ := pgTx.Amount.Float64Value()
	amount = val.Float64

	var paymentTokenID *uuid.UUID
	if pgTx.PaymentTokenID.Valid {
		id := pgToUUID(pgTx.PaymentTokenID)
		paymentTokenID = &id
	}

	var externalTxID *string
	if pgTx.ExternalTxID.Valid {
		externalTxID = &pgTx.ExternalTxID.String
	}

	return &domain.Transaction{
		ID:             pgToUUID(pgTx.ID),
		UserID:         pgToUUID(pgTx.UserID),
		Amount:         amount,
		Currency:       pgTx.Currency,
		Type:           pgTx.Type,
		Status:         domain.TransactionStatus(pgTx.Status),
		Provider:       pgTx.Provider,
		PaymentTokenID: paymentTokenID,
		ExternalTxID:   externalTxID,
		CreatedAt:      pgTx.CreatedAt.Time,
		UpdatedAt:      pgTx.UpdatedAt.Time,
	}, nil
}

func (r *postgresBillingRepository) GetSubscriptionsForBilling(ctx context.Context) ([]domain.Subscription, error) {
	pgSubs, err := r.queries.GetSubscriptionsForBilling(ctx)
	if err != nil {
		return nil, err
	}

	subs := make([]domain.Subscription, len(pgSubs))
	for i, s := range pgSubs {
		subs[i] = domain.Subscription{
			ID:               pgToUUID(s.ID),
			UserID:           pgToUUID(s.UserID),
			PlanName:         s.PlanName,
			Status:           domain.SubscriptionStatus(s.Status),
			StartDate:        s.StartDate.Time,
			EndDate:          s.EndDate.Time,
			HasUsedTrial:     s.HasUsedTrial,
			RidesLeft:        s.RidesLeft,
			FreeParcelsQuota: s.FreeParcelsQuota,
			CreatedAt:        s.CreatedAt.Time,
			UpdatedAt:        s.UpdatedAt.Time,
		}
	}
	return subs, nil
}

func (r *postgresBillingRepository) UpdateSubscriptionStatus(ctx context.Context, subID uuid.UUID, status domain.SubscriptionStatus, endDate time.Time) error {
	return r.queries.UpdateSubscriptionStatus(ctx, sqlc.UpdateSubscriptionStatusParams{
		Status:  string(status),
		EndDate: pgtype.Timestamptz{Time: endDate, Valid: true},
		ID:      uuidToPg(subID),
	})
}

func (r *postgresBillingRepository) ApplyPromoCodeTx(ctx context.Context, userID uuid.UUID, code string) (*domain.PromoCode, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgPromo, err := qtx.GetPromoCodeForUpdate(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get promo code or promo code not active: %w", err)
	}

	if pgPromo.ExpiresAt.Valid && pgPromo.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("promo code has expired")
	}

	if pgPromo.StartsAt.Valid && pgPromo.StartsAt.Time.After(time.Now()) {
		return nil, fmt.Errorf("promo code is not yet active")
	}

	if pgPromo.CurrentUses >= pgPromo.MaxUses {
		return nil, fmt.Errorf("promo code usage limit reached")
	}

	hasUsed, err := qtx.CheckPromoUsage(ctx, sqlc.CheckPromoUsageParams{
		UserID:      uuidToPg(userID),
		PromoCodeID: pgPromo.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check promo usage: %w", err)
	}
	if hasUsed {
		return nil, fmt.Errorf("user has already used this promo code")
	}

	err = qtx.UpdatePromoCodeUses(ctx, pgPromo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update promo code uses: %w", err)
	}

	err = qtx.RecordPromoUsage(ctx, sqlc.RecordPromoUsageParams{
		ID:          uuidToPg(uuid.New()),
		UserID:      uuidToPg(userID),
		PromoCodeID: pgPromo.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to record promo usage: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	var discAmt *float64
	if pgPromo.DiscountAmount.Valid {
		val, _ := pgPromo.DiscountAmount.Float64Value()
		v := val.Float64
		discAmt = &v
	}

	var discPct *float64
	if pgPromo.DiscountPercent.Valid {
		val, _ := pgPromo.DiscountPercent.Float64Value()
		v := val.Float64
		discPct = &v
	}

	var startsAt *time.Time
	if pgPromo.StartsAt.Valid {
		startsAt = &pgPromo.StartsAt.Time
	}

	var expiresAt *time.Time
	if pgPromo.ExpiresAt.Valid {
		expiresAt = &pgPromo.ExpiresAt.Time
	}

	return &domain.PromoCode{
		ID:              pgToUUID(pgPromo.ID),
		Code:            pgPromo.Code,
		DiscountAmount:  discAmt,
		DiscountPercent: discPct,
		MaxUses:         pgPromo.MaxUses,
		CurrentUses:     pgPromo.CurrentUses + 1,
		IsActive:        pgPromo.IsActive,
		DiscountType:    pgPromo.DiscountType.String,
		StartsAt:        startsAt,
		ExpiresAt:       expiresAt,
		CreatedAt:       pgPromo.CreatedAt.Time,
		UpdatedAt:       pgPromo.UpdatedAt.Time,
	}, nil
}

func (r *postgresBillingRepository) AddFreeParcelsToPremium(ctx context.Context, quota int32) error {
	return r.queries.AddFreeParcelsToPremium(ctx, quota)
}

func (r *postgresBillingRepository) DeductRideLimit(ctx context.Context, driverID uuid.UUID) error {
	rows, err := r.queries.DeductRideLimit(ctx, uuidToPg(driverID))
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrLimitExceeded
	}
	return nil
}

func (r *postgresBillingRepository) RefundRideLimit(ctx context.Context, driverID uuid.UUID) error {
	return r.queries.RefundRideLimit(ctx, uuidToPg(driverID))
}

func (r *postgresBillingRepository) CancelStaleTransactions(ctx context.Context) error {
	return r.queries.CancelStaleTransactions(ctx)
}

func (r *postgresBillingRepository) ResetExpiredSubscriptions(ctx context.Context) (int64, error) {
	return r.queries.ResetExpiredSubscriptions(ctx)
}

func (r *postgresBillingRepository) DeductParcelLimit(ctx context.Context, driverID uuid.UUID) error {
	count, err := r.queries.DeductParcelLimit(ctx, uuidToPg(driverID))
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("no parcel quota left or active subscription not found")
	}
	return nil
}

func (r *postgresBillingRepository) RefundParcelLimit(ctx context.Context, driverID uuid.UUID) error {
	return r.queries.RefundParcelLimit(ctx, uuidToPg(driverID))
}

func (r *postgresBillingRepository) UpdateActiveSubscriptionLimits(ctx context.Context, subID uuid.UUID, planName string, rides int32, parcels int32, endDate time.Time) error {
	return r.queries.UpdateActiveSubscriptionLimits(ctx, sqlc.UpdateActiveSubscriptionLimitsParams{
		PlanName:         planName,
		RidesLeft:        rides,
		FreeParcelsQuota: parcels,
		EndDate:          pgtype.Timestamptz{Time: endDate, Valid: true},
		ID:               uuidToPg(subID),
	})
}

func (r *postgresBillingRepository) CancelActiveSubscription(ctx context.Context, driverID uuid.UUID) error {
	return r.queries.CancelActiveSubscription(ctx, uuidToPg(driverID))
}

func (r *postgresBillingRepository) ProcessWebhookTx(ctx context.Context, txID uuid.UUID, externalTxID string, amount float64, planName string, ridesLeft int32, freeParcels int32) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// 1. Lock transaction
	pgTx, err := qtx.GetTransactionForUpdate(ctx, uuidToPg(txID))
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	if domain.TransactionStatus(pgTx.Status) == domain.TransactionStatusSuccess {
		return nil // Idempotency
	}

	val, _ := pgTx.Amount.Float64Value()
	if val.Float64 != amount {
		return fmt.Errorf("amount mismatch")
	}

	// 2. Update transaction status
	err = qtx.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
		Status:       string(domain.TransactionStatusSuccess),
		ExternalTxID: pgtype.Text{String: externalTxID, Valid: true},
		ID:           uuidToPg(txID),
	})
	if err != nil {
		return fmt.Errorf("failed to update tx status: %w", err)
	}

	userID := pgToUUID(pgTx.UserID)

	// 3. Cancel existing active subscription if any
	pgSub, err := qtx.GetActiveSubscription(ctx, uuidToPg(userID))
	if err == nil {
		err = qtx.UpdateSubscriptionStatus(ctx, sqlc.UpdateSubscriptionStatusParams{
			Status:  string(domain.SubscriptionStatusCanceled),
			EndDate: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ID:      pgSub.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to cancel existing sub: %w", err)
		}
	}

	// 4. Create new subscription
	newSubID := uuid.New()
	_, err = qtx.CreateSubscription(ctx, sqlc.CreateSubscriptionParams{
		ID:               uuidToPg(newSubID),
		UserID:           uuidToPg(userID),
		PlanName:         planName,
		Status:           string(domain.SubscriptionStatusActive),
		StartDate:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
		EndDate:          pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
		HasUsedTrial:     false,
		RidesLeft:        ridesLeft,
		FreeParcelsQuota: freeParcels,
	})
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}
