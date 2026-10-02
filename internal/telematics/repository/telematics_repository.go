package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/telematics/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	redis "github.com/redis/go-redis/v9"
)

type telematicsRepository struct {
	redisRepo    *redisTelematicsRepository
	postgresRepo *postgresTelematicsRepository
}

func NewTelematicsRepository(redisClient *redis.Client, pgPool *pgxpool.Pool, ridesPool *pgxpool.Pool) domain.TelematicsRepository {
	return &telematicsRepository{
		redisRepo:    NewRedisTelematicsRepository(redisClient),
		postgresRepo: NewPostgresTelematicsRepository(pgPool, ridesPool),
	}
}

// Hot Data methods pass to Redis
func (r *telematicsRepository) SaveLocationCache(ctx context.Context, loc *domain.Location) error {
	return r.redisRepo.SaveLocationCache(ctx, loc)
}

func (r *telematicsRepository) GetNearbyDrivers(ctx context.Context, lat, lon float64, radiusKm float64) ([]domain.NearbyDriver, error) {
	return r.redisRepo.GetNearbyDrivers(ctx, lat, lon, radiusKm)
}

func (r *telematicsRepository) GetLocationCache(ctx context.Context, driverID uuid.UUID) (*domain.Location, error) {
	return r.redisRepo.GetLocationCache(ctx, driverID)
}

func (r *telematicsRepository) GetAllActiveLocations(ctx context.Context) ([]domain.Location, error) {
	return r.redisRepo.GetAllActiveLocations(ctx)
}

func (r *telematicsRepository) PushLocationToQueue(ctx context.Context, loc *domain.Location) error {
	return r.redisRepo.PushLocationToQueue(ctx, loc)
}

func (r *telematicsRepository) PopLocationsFromQueue(ctx context.Context, count int) ([]domain.Location, error) {
	return r.redisRepo.PopLocationsFromQueue(ctx, count)
}

// Cold Data methods pass to Postgres
func (r *telematicsRepository) SaveLocationsBatch(ctx context.Context, locations []domain.Location) error {
	return r.postgresRepo.SaveLocationsBatch(ctx, locations)
}

func (r *telematicsRepository) IsConfirmedPassenger(ctx context.Context, driverID, passengerID uuid.UUID) (bool, error) {
	return r.postgresRepo.IsConfirmedPassenger(ctx, driverID, passengerID)
}
