package repository_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/billing/repository"
	"github.com/hextec/yolyolakay-backend/pkg/testutils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestApplyPromoCodeTx_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()

	migrations := []string{
		"../../../deployments/migrations/user",
		"../../../deployments/migrations/billing",
	}

	connString, cleanup, err := testutils.SetupTestDatabase(ctx, migrations)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresBillingRepository(pool)

	promoCodeStr := fmt.Sprintf("PROMO_%d", time.Now().UnixNano())

	_, err = pool.Exec(ctx, `
		INSERT INTO promo_codes (
			id, code, discount_percent, max_uses, current_uses, is_active, created_at, updated_at
		) VALUES (
			$1, $2, 10.0, 3, 0, true, NOW(), NOW()
		)`, uuid.New(), promoCodeStr)
	if err != nil {
		t.Fatalf("Failed to insert promo code: %v", err)
	}

	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	var successCount int32
	var failCount int32

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()

			userID := uuid.New()
			userPhone := fmt.Sprintf("+999222%02d%04d", idx, time.Now().Unix()%10000)
			// Insert user to satisfy FK
			_, _ = pool.Exec(ctx, "INSERT INTO users (id, first_name, last_name, phone_number) VALUES ($1, 'User', 'Test', $2)", userID, userPhone)

			_, err := repo.ApplyPromoCodeTx(ctx, userID, promoCodeStr)
			if err != nil {
				atomic.AddInt32(&failCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if successCount != 3 {
		t.Errorf("Expected 3 successful applies, got %d", successCount)
	}
	if failCount != 7 {
		t.Errorf("Expected 7 failed applies, got %d", failCount)
	}

	var currentUses int
	err = pool.QueryRow(ctx, "SELECT current_uses FROM promo_codes WHERE code = $1", promoCodeStr).Scan(&currentUses)
	if err != nil {
		t.Fatalf("Failed to query promo code: %v", err)
	}

	if currentUses != 3 {
		t.Errorf("Expected current_uses to be 3, got %d", currentUses)
	}
}
