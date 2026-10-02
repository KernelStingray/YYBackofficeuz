package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/ride/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReminderWorker struct {
	pool      *pgxpool.Pool
	publisher domain.EventPublisher
	stopChan  chan struct{}
}

func NewReminderWorker(pool *pgxpool.Pool, publisher domain.EventPublisher) *ReminderWorker {
	return &ReminderWorker{
		pool:      pool,
		publisher: publisher,
		stopChan:  make(chan struct{}),
	}
}

func (w *ReminderWorker) Start() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				w.processReminders(context.Background())
			case <-w.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
	slog.Info("Ride Reminder Worker started")
}

func (w *ReminderWorker) Stop() {
	close(w.stopChan)
}

func (w *ReminderWorker) processReminders(ctx context.Context) {
	if w.pool == nil || w.publisher == nil {
		return
	}

	// 1. Reminders for 1 hour window (55-65 mins from now)
	w.checkAndSend(ctx, "ride_reminder_1h", 55*time.Minute, 65*time.Minute)

	// 2. Reminders for 15 minutes window (10-20 mins from now)
	w.checkAndSend(ctx, "ride_reminder_15m", 10*time.Minute, 20*time.Minute)

	// 3. Auto-complete departed rides (> 24 hours past departure)
	w.expireOldActiveRides(ctx)
}

func (w *ReminderWorker) expireOldActiveRides(ctx context.Context) {
	query := `
		UPDATE rides
		SET status = 'completed', updated_at = NOW()
		WHERE status = 'active' AND departure_time < NOW() - INTERVAL '24 hours';
	`
	_, err := w.pool.Exec(ctx, query)
	if err != nil {
		slog.Warn("failed to expire old active rides", "error", err)
	}
}

func (w *ReminderWorker) checkAndSend(ctx context.Context, kind string, minOffset, maxOffset time.Duration) {
	now := time.Now().UTC()
	startTime := now.Add(minOffset)
	endTime := now.Add(maxOffset)

	queryRides := `
		SELECT id, driver_id
		FROM rides
		WHERE status = 'published' AND departure_time BETWEEN $1 AND $2
	`
	rows, err := w.pool.Query(ctx, queryRides, startTime, endTime)
	if err != nil {
		slog.Error("Failed to query rides for reminders", "kind", kind, "error", err)
		return
	}
	defer rows.Close()

	type rideEntry struct {
		rideID   uuid.UUID
		driverID uuid.UUID
	}
	var rides []rideEntry
	for rows.Next() {
		var r rideEntry
		if err := rows.Scan(&r.rideID, &r.driverID); err == nil {
			rides = append(rides, r)
		}
	}

	for _, r := range rides {
		// Send to Driver
		w.sendIfNotSent(ctx, r.rideID, r.driverID, kind)

		// Send to Confirmed Passengers
		queryBookings := `
			SELECT passenger_id
			FROM bookings
			WHERE ride_id = $1 AND status = 'confirmed'
		`
		bRows, err := w.pool.Query(ctx, queryBookings, r.rideID)
		if err == nil {
			for bRows.Next() {
				var pID uuid.UUID
				if err := bRows.Scan(&pID); err == nil {
					w.sendIfNotSent(ctx, r.rideID, pID, kind)
				}
			}
			bRows.Close()
		}
	}
}

func (w *ReminderWorker) sendIfNotSent(ctx context.Context, rideID, userID uuid.UUID, kind string) {
	// Deduplication insert
	query := `
		INSERT INTO sent_reminders (ride_id, user_id, kind)
		VALUES ($1, $2, $3)
		ON CONFLICT (ride_id, user_id, kind) DO NOTHING
		RETURNING id
	`
	var insertedID uuid.UUID
	err := w.pool.QueryRow(ctx, query, rideID, userID, kind).Scan(&insertedID)
	if err != nil {
		// If no row was inserted due to ON CONFLICT, QueryRow returns pgx.ErrNoRows -> already sent!
		return
	}

	// Successfully inserted -> publish event to RabbitMQ
	payload, _ := json.Marshal(map[string]interface{}{
		"kind":    kind,
		"ride_id": rideID.String(),
		"user_id": userID.String(),
	})

	if err := w.publisher.Publish(ctx, "yolyolakay.events", "ride.reminder", payload); err != nil {
		slog.Error("Failed to publish ride.reminder event", "kind", kind, "user_id", userID, "error", err)
	} else {
		slog.Info("Successfully published reminder event", "kind", kind, "ride_id", rideID, "user_id", userID)
	}
}
