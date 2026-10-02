package repository_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/ride/repository"
	"github.com/hextec/yolyolakay-backend/pkg/testutils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCreateBookingTx_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()

	migrations := []string{
		"../../../deployments/migrations/user",
		"../../../deployments/migrations/ride",
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

	repo := repository.NewPostgresRideRepository(pool)

	driverID := uuid.New()
	carID := uuid.New()
	rideID := uuid.New()

	driverPhone := fmt.Sprintf("+999333%06d", time.Now().Unix()%1000000)
	_, err = pool.Exec(ctx, "INSERT INTO users (id, first_name, last_name, phone_number) VALUES ($1, 'Test', 'Driver', $2)", driverID, driverPhone)
	if err != nil {
		t.Fatalf("Failed to insert driver: %v", err)
	}
	govNumber := fmt.Sprintf("A%03dBC", time.Now().Unix()%1000)
	_, err = pool.Exec(ctx, "INSERT INTO cars (id, user_id, gov_number, brand, model) VALUES ($1, $2, $3, 'Toyota', 'Camry')", carID, driverID, govNumber)
	if err != nil {
		t.Fatalf("Failed to insert car: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO rides (
			id, driver_id, vehicle_id, origin, destination, 
			origin_address, destination_address, departure_time, 
			total_seats, available_seats, price_per_seat, status
		) VALUES (
			$1, $2, $3, ST_MakePoint(69.24, 41.31), ST_MakePoint(69.25, 41.32),
			'Origin', 'Dest', NOW() + INTERVAL '1 day',
			2, 2, 50000, 'active'
		)`, rideID, driverID, carID)
	if err != nil {
		t.Fatalf("Failed to insert ride: %v", err)
	}

	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	bookingIDs := make([]uuid.UUID, concurrency)
	var createSuccessCount int32

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()

			passengerID := uuid.New()
			passengerPhone := fmt.Sprintf("+999444%02d%04d", idx, time.Now().Unix()%10000)
			_, _ = pool.Exec(ctx, "INSERT INTO users (id, first_name, last_name, phone_number) VALUES ($1, 'Pass', 'enger', $2)", passengerID, passengerPhone)

			b, err := repo.CreateBookingTx(ctx, rideID, passengerID, 1, false, nil, "door_to_door", "highway")
			if err == nil && b != nil {
				bookingIDs[idx] = b.ID
				atomic.AddInt32(&createSuccessCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// 1. All 10 requests can be submitted (booking creation doesn't block)
	if createSuccessCount != int32(concurrency) {
		t.Errorf("Expected all %d bookings to be created as pending, got %d", concurrency, createSuccessCount)
	}

	var initialSeats int
	err = pool.QueryRow(ctx, "SELECT available_seats FROM rides WHERE id = $1", rideID).Scan(&initialSeats)
	if err != nil || initialSeats != 2 {
		t.Errorf("Expected available_seats to remain 2 after creation, got %d (err: %v)", initialSeats, err)
	}

	// 2. Concurrently accept bookings (only 2 seats should be accepted)
	var acceptSuccessCount int32
	var acceptFailCount int32
	var acceptedBookingID uuid.UUID
	var mu sync.Mutex

	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			bID := bookingIDs[idx]
			if bID == uuid.Nil {
				return
			}
			_, err := repo.AcceptBookingTx(ctx, rideID, bID)
			if err != nil {
				atomic.AddInt32(&acceptFailCount, 1)
			} else {
				atomic.AddInt32(&acceptSuccessCount, 1)
				mu.Lock()
				acceptedBookingID = bID
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if acceptSuccessCount != 2 {
		t.Errorf("Expected exactly 2 bookings to be accepted, got %d", acceptSuccessCount)
	}
	if acceptFailCount != 8 {
		t.Errorf("Expected 8 accept attempts to fail due to seat exhaustion, got %d", acceptFailCount)
	}

	var finalSeats int
	err = pool.QueryRow(ctx, "SELECT available_seats FROM rides WHERE id = $1", rideID).Scan(&finalSeats)
	if err != nil || finalSeats != 0 {
		t.Errorf("Expected available_seats to be 0 after accepting 2 seats, got %d", finalSeats)
	}

	// 3. Driver rejects one confirmed booking -> seat must be restored
	if acceptedBookingID != uuid.Nil {
		err = repo.RejectBooking(ctx, acceptedBookingID)
		if err != nil {
			t.Errorf("Failed to reject confirmed booking: %v", err)
		}

		var restoredSeats int
		_ = pool.QueryRow(ctx, "SELECT available_seats FROM rides WHERE id = $1", rideID).Scan(&restoredSeats)
		if restoredSeats != 1 {
			t.Errorf("Expected available_seats to be restored to 1, got %d", restoredSeats)
		}
	}
}
