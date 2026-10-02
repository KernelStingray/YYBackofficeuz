package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/telematics/domain"
	"github.com/hextec/yolyolakay-backend/internal/telematics/repository/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresTelematicsRepository struct {
	db        *pgxpool.Pool
	ridesPool *pgxpool.Pool
	queries   *postgres.Queries
}

func NewPostgresTelematicsRepository(db *pgxpool.Pool, ridesPool *pgxpool.Pool) *postgresTelematicsRepository {
	return &postgresTelematicsRepository{
		db:        db,
		ridesPool: ridesPool,
		queries:   postgres.New(db),
	}
}

func (r *postgresTelematicsRepository) IsConfirmedPassenger(ctx context.Context, driverID, passengerID uuid.UUID) (bool, error) {
	pool := r.ridesPool
	if pool == nil {
		pool = r.db
	}
	var exists int
	err := pool.QueryRow(ctx, `
		SELECT 1 FROM rides r
		JOIN bookings b ON b.ride_id = r.id
		WHERE r.driver_id = $1
		  AND b.passenger_id = $2
		  AND b.status IN ('confirmed', 'accepted')
		  AND r.status IN ('active', 'started')
		LIMIT 1
	`, driverID, passengerID).Scan(&exists)
	if err != nil {
		return false, nil
	}
	return exists == 1, nil
}

func (r *postgresTelematicsRepository) SaveLocationsBatch(ctx context.Context, locations []domain.Location) error {
	if len(locations) == 0 {
		return nil
	}

	driverIDs := make([]pgtype.UUID, len(locations))
	longitudes := make([]float64, len(locations))
	latitudes := make([]float64, len(locations))
	speeds := make([]pgtype.Numeric, len(locations))
	headings := make([]pgtype.Numeric, len(locations))
	createdAts := make([]pgtype.Timestamptz, len(locations))

	for i, loc := range locations {
		driverIDs[i] = pgtype.UUID{Bytes: loc.DriverID, Valid: true}
		longitudes[i] = loc.Longitude
		latitudes[i] = loc.Latitude

		// Simple mock conversion for speed and heading to pgtype.Numeric
		sp := pgtype.Numeric{}
		sp.Scan(loc.Speed)
		speeds[i] = sp

		hd := pgtype.Numeric{}
		hd.Scan(loc.Heading)
		headings[i] = hd

		createdAts[i] = pgtype.Timestamptz{Time: loc.Timestamp, Valid: true}
	}

	return r.queries.BulkInsertLocations(ctx, postgres.BulkInsertLocationsParams{
		DriverIds:  driverIDs,
		Longitudes: longitudes,
		Latitudes:  latitudes,
		Speeds:     speeds,
		Headings:   headings,
		CreatedAts: createdAts,
	})
}
