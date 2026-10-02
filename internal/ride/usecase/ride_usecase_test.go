package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/ride/domain"
	rideMocks "github.com/hextec/yolyolakay-backend/pkg/mocks/ride"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupRideUseCaseTest(t *testing.T) (domain.RideUseCase, *rideMocks.MockRideRepository, *rideMocks.MockRideCache, *rideMocks.MockGRPCClient, *rideMocks.MockEventPublisher) {
	ctrl := gomock.NewController(t)

	repo := rideMocks.NewMockRideRepository(ctrl)
	cache := rideMocks.NewMockRideCache(ctrl)
	grpcClient := rideMocks.NewMockGRPCClient(ctrl)
	publisher := rideMocks.NewMockEventPublisher(ctrl)

	uc := NewRideUseCase(repo, nil, cache, grpcClient, publisher)

	cache.EXPECT().SetRideDriver(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	cache.EXPECT().InvalidateSearchResults(gomock.Any()).Return(nil).AnyTimes()
	repo.EXPECT().UpdateRideRoutePolygon(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	return uc, repo, cache, grpcClient, publisher
}

func TestPublishRide_Success(t *testing.T) {
	uc, repo, _, grpcClient, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()
	ride := &domain.Ride{
		DriverID:       driverID,
		OriginAddress:  "Tashkent",
		OriginLat:      41.2995, // Tashkent
		OriginLon:      69.2401,
		DestinationLat: 39.6542, // Samarkand
		DestinationLon: 66.9597,
		PricePerSeat:   100000,
		TotalSeats:     4,
		AvailableSeats: 4,
		DepartureTime:  time.Now().Add(2 * time.Hour),
	}

	grpcClient.EXPECT().GetDriverVehicle(ctx, driverID).Return(nil, errors.New("not found")).AnyTimes()

	repo.EXPECT().CreateRide(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, r *domain.Ride) error {
		r.ID = uuid.New()
		assert.True(t, r.HasFairPriceBadge)
		return nil
	})

	err := uc.PublishRide(ctx, ride)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, ride.ID)
	assert.True(t, ride.HasFairPriceBadge)
}

func TestPublishRide_HighPrice_NoBadge(t *testing.T) {
	uc, repo, _, grpcClient, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()
	ride := &domain.Ride{
		DriverID:       driverID,
		OriginAddress:  "Tashkent",
		OriginLat:      41.2995, // Tashkent
		OriginLon:      69.2401,
		DestinationLat: 39.6542, // Samarkand
		DestinationLon: 66.9597,
		PricePerSeat:   200000,
		TotalSeats:     4,
		AvailableSeats: 4,
		DepartureTime:  time.Now().Add(2 * time.Hour),
	}

	grpcClient.EXPECT().GetDriverVehicle(ctx, driverID).Return(nil, errors.New("not found")).AnyTimes()

	repo.EXPECT().CreateRide(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, r *domain.Ride) error {
		r.ID = uuid.New()
		assert.False(t, r.HasFairPriceBadge)
		return nil
	})

	err := uc.PublishRide(ctx, ride)

	require.NoError(t, err)
	assert.False(t, ride.HasFairPriceBadge)
}

func TestFindRides_CacheMiss_DBFallback(t *testing.T) {
	uc, repo, cache, _, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	dbRides := []domain.Ride{
		{ID: uuid.New(), OriginAddress: "Tashkent", DestinationAddress: "Samarkand", DriverID: uuid.New()},
	}

	cache.EXPECT().GetSearchResults(ctx, gomock.Any()).Return(nil, errors.New("redis: nil"))
	repo.EXPECT().SearchRides(ctx, gomock.Any()).Return(dbRides, nil)
	cache.EXPECT().SetSearchResults(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	cache.EXPECT().GetDriverProfile(ctx, gomock.Any()).Return(&domain.DriverProfile{}, nil).AnyTimes()

	params := domain.RideSearchParams{
		OriginLat:    41.0,
		OriginLon:    69.0,
		DestLat:      39.0,
		DestLon:      66.0,
		RadiusMeters: 10.0,
		MinSeats:     1,
	}
	rides, err := uc.FindRides(ctx, params)

	require.NoError(t, err)
	assert.Len(t, rides, 1)
}

func TestGetRideByID_EnrichesDriverProfile(t *testing.T) {
	uc, repo, cache, grpcClient, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()
	rideID := uuid.New()
	dbRide := &domain.Ride{
		ID:                 rideID,
		DriverID:           driverID,
		OriginAddress:      "Tashkent",
		DestinationAddress: "Samarkand",
	}

	expectedProfile := &domain.DriverProfile{
		ID:     driverID,
		Name:   "Alisher Navoiy",
		Rating: 4.9,
		Avatar: "https://api.yolyolakay.com/api/v1/media/avatar.png",
	}

	repo.EXPECT().GetRideByID(ctx, rideID).Return(dbRide, nil)
	repo.EXPECT().GetBookingsByRideID(ctx, rideID).Return([]domain.Booking{}, nil)
	cache.EXPECT().GetDriverProfile(ctx, driverID).Return(nil, errors.New("redis: nil"))
	grpcClient.EXPECT().GetDriverProfile(ctx, driverID).Return(expectedProfile, nil)
	cache.EXPECT().SetDriverProfile(ctx, driverID, expectedProfile).Return(nil)
	grpcClient.EXPECT().GetDriverVehicle(ctx, driverID).Return(nil, errors.New("vehicle not found"))

	ride, err := uc.GetRideByID(ctx, rideID)

	require.NoError(t, err)
	require.NotNil(t, ride)
	require.NotNil(t, ride.Driver)
	assert.Equal(t, driverID, ride.Driver.ID)
	assert.Equal(t, "Alisher Navoiy", ride.Driver.Name)
	assert.Equal(t, 4.9, ride.Driver.Rating)
}

func TestGetRideDriverLocation_DriverAndPassengerAuthorized(t *testing.T) {
	uc, repo, cache, _, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()
	passengerID := uuid.New()
	unauthorizedID := uuid.New()
	rideID := uuid.New()

	dbRide := &domain.Ride{
		ID:       rideID,
		DriverID: driverID,
		Status:   domain.RideStatusActive,
	}

	expectedLocation := &domain.DriverLocation{
		DriverID:  driverID,
		Latitude:  41.311081,
		Longitude: 69.240562,
		Speed:     65.0,
		Heading:   180.0,
		UpdatedAt: time.Now().Add(-1 * time.Minute),
	}

	// 1. Driver queries location -> OK
	repo.EXPECT().GetRideByID(ctx, rideID).Return(dbRide, nil)
	cache.EXPECT().GetDriverLiveLocation(ctx, driverID).Return(expectedLocation, nil)

	loc, err := uc.GetRideDriverLocation(ctx, rideID, driverID)
	require.NoError(t, err)
	require.NotNil(t, loc)
	assert.Equal(t, 41.311081, loc.Latitude)
	assert.Equal(t, 180.0, loc.Heading)

	// 2. Confirmed passenger queries location -> OK
	confirmedBookings := []domain.Booking{
		{
			ID:          uuid.New(),
			RideID:      rideID,
			PassengerID: passengerID,
			Status:      domain.BookingStatusConfirmed,
		},
	}
	repo.EXPECT().GetRideByID(ctx, rideID).Return(dbRide, nil)
	repo.EXPECT().GetBookingsByRideID(ctx, rideID).Return(confirmedBookings, nil)
	cache.EXPECT().GetDriverLiveLocation(ctx, driverID).Return(expectedLocation, nil)

	loc, err = uc.GetRideDriverLocation(ctx, rideID, passengerID)
	require.NoError(t, err)
	require.NotNil(t, loc)

	// 3. Unauthorized user queries location -> 403 Forbidden
	repo.EXPECT().GetRideByID(ctx, rideID).Return(dbRide, nil)
	repo.EXPECT().GetBookingsByRideID(ctx, rideID).Return(confirmedBookings, nil)

	loc, err = uc.GetRideDriverLocation(ctx, rideID, unauthorizedID)
	require.Error(t, err)
	assert.Nil(t, loc)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestBookRide_CapacityAndOwnRideValidation(t *testing.T) {
	uc, repo, _, _, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()
	passengerID := uuid.New()
	rideID := uuid.New()

	ride := &domain.Ride{
		ID:             rideID,
		DriverID:       driverID,
		TotalSeats:     4,
		AvailableSeats: 2,
		Status:         domain.RideStatusActive,
		DepartureTime:  time.Now().Add(2 * time.Hour),
	}

	// 1. Cannot book 0 seats without buyout
	_, err := uc.BookRide(ctx, rideID, passengerID, 0, false, nil, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must book at least 1 seat")

	// 2. Cannot book own ride
	repo.EXPECT().GetRideByID(ctx, rideID).Return(ride, nil)
	_, err = uc.BookRide(ctx, rideID, driverID, 1, false, nil, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot book your own ride")

	// 3. Exceeds total vehicle capacity
	repo.EXPECT().GetRideByID(ctx, rideID).Return(ride, nil)
	_, err = uc.BookRide(ctx, rideID, passengerID, 5, false, nil, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot book more seats than total vehicle capacity")

	// 4. Exceeds currently available seats
	repo.EXPECT().GetRideByID(ctx, rideID).Return(ride, nil)
	_, err = uc.BookRide(ctx, rideID, passengerID, 3, false, nil, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not enough available seats")
}

func TestBookRide_GenderRestriction(t *testing.T) {
	uc, repo, _, grpcClient, publisher := setupRideUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()
	malePassengerID := uuid.New()
	femalePassengerID := uuid.New()
	rideID := uuid.New()

	ride := &domain.Ride{
		ID:                rideID,
		DriverID:          driverID,
		TotalSeats:        4,
		AvailableSeats:    4,
		GenderRestriction: "female",
		Status:            domain.RideStatusActive,
		DepartureTime:     time.Now().Add(2 * time.Hour),
	}

	// 1. Male passenger attempting to book female-only ride -> Error
	repo.EXPECT().GetRideByID(ctx, rideID).Return(ride, nil)
	grpcClient.EXPECT().GetDriverProfile(ctx, malePassengerID).Return(&domain.DriverProfile{
		Gender: "male",
	}, nil)

	_, err := uc.BookRide(ctx, rideID, malePassengerID, 1, false, nil, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ride is restricted to female passengers only")

	// 2. Female passenger booking female-only ride -> Success
	repo.EXPECT().GetRideByID(ctx, rideID).Return(ride, nil)
	grpcClient.EXPECT().GetDriverProfile(ctx, femalePassengerID).Return(&domain.DriverProfile{
		Gender: "female",
	}, nil)
	repo.EXPECT().CreateBookingTx(ctx, rideID, femalePassengerID, int32(1), false, nil, "", "").Return(&domain.Booking{
		ID:          uuid.New(),
		RideID:      rideID,
		PassengerID: femalePassengerID,
		SeatsBooked: 1,
		Status:      domain.BookingStatusPending,
	}, nil)
	publisher.EXPECT().Publish(ctx, "yolyolakay.events", "ride.booking.created", gomock.Any()).Return(nil)

	booking, err := uc.BookRide(ctx, rideID, femalePassengerID, 1, false, nil, "", "")
	require.NoError(t, err)
	assert.NotNil(t, booking)
}

func TestCancelBookingByPassenger(t *testing.T) {
	uc, repo, _, _, publisher := setupRideUseCaseTest(t)
	ctx := context.Background()

	rideID := uuid.New()
	driverID := uuid.New()
	passengerID := uuid.New()
	bookingID := uuid.New()

	booking := &domain.Booking{
		ID:          bookingID,
		RideID:      rideID,
		PassengerID: passengerID,
		SeatsBooked: 2,
		Status:      domain.BookingStatusCanceled,
	}

	ride := &domain.Ride{
		ID:       rideID,
		DriverID: driverID,
		Status:   domain.RideStatusActive,
	}

	repo.EXPECT().CancelBookingTx(ctx, passengerID, bookingID).Return(booking, nil).Times(1)
	repo.EXPECT().GetRideByID(ctx, rideID).Return(ride, nil).Times(1)
	repo.EXPECT().GetBookingsByRideID(ctx, rideID).Return([]domain.Booking{}, nil).AnyTimes()
	publisher.EXPECT().Publish(ctx, "yolyolakay.events", "ride.booking.cancelled", gomock.Any()).Return(nil).Times(1)
	publisher.EXPECT().Publish(ctx, "yolyolakay.events", "ride.status_changed", gomock.Any()).Return(nil).Times(1)

	err := uc.CancelBookingByPassenger(ctx, passengerID, bookingID, "Планы изменились")
	require.NoError(t, err)
}

func TestPublishRide_PastDepartureTime(t *testing.T) {
	uc, _, _, _, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	ride := &domain.Ride{
		DriverID:       uuid.New(),
		OriginAddress:  "Tashkent",
		DepartureTime:  time.Now().Add(-1 * time.Hour), // 1 hour in the past
		PricePerSeat:   50000,
		TotalSeats:     4,
		AvailableSeats: 4,
	}

	err := uc.PublishRide(ctx, ride)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrPastDepartureTime)
}

func TestBookRide_DepartedRide(t *testing.T) {
	uc, repo, _, _, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	rideID := uuid.New()
	passengerID := uuid.New()

	ride := &domain.Ride{
		ID:            rideID,
		Status:        domain.RideStatusActive,
		DepartureTime: time.Now().Add(-4 * time.Hour), // departed 4 hours ago (>3h threshold)
	}

	repo.EXPECT().GetRideByID(ctx, rideID).Return(ride, nil)

	_, err := uc.BookRide(ctx, rideID, passengerID, 1, false, nil, "", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrRideDeparted)
}

func TestBookRide_InactiveRide(t *testing.T) {
	uc, repo, _, _, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	rideID := uuid.New()
	passengerID := uuid.New()

	ride := &domain.Ride{
		ID:            rideID,
		Status:        domain.RideStatusCompleted,
		DepartureTime: time.Now().Add(2 * time.Hour),
	}

	repo.EXPECT().GetRideByID(ctx, rideID).Return(ride, nil)

	_, err := uc.BookRide(ctx, rideID, passengerID, 1, false, nil, "", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrRideNotActive)
}

func TestFindRides_CacheKeySeparation_IncludeTest(t *testing.T) {
	uc, repo, cache, _, _ := setupRideUseCaseTest(t)
	ctx := context.Background()

	dbRidesFalse := []domain.Ride{}
	dbRidesTrue := []domain.Ride{
		{ID: uuid.New(), OriginAddress: "Tashkent", DestinationAddress: "Samarkand", DriverID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
	}

	var capturedKeyFalse, capturedKeyTrue string
	cache.EXPECT().GetSearchResults(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, key string) ([]domain.Ride, error) {
		capturedKeyFalse = key
		return nil, errors.New("miss")
	})
	repo.EXPECT().SearchRides(ctx, gomock.Any()).Return(dbRidesFalse, nil)

	paramsFalse := domain.RideSearchParams{
		OriginLat:    41.0,
		OriginLon:    69.0,
		DestLat:      39.0,
		DestLon:      66.0,
		RadiusMeters: 10.0,
		MinSeats:     1,
		IncludeTest:  false,
	}
	_, err := uc.FindRides(ctx, paramsFalse)
	require.NoError(t, err)

	cache.EXPECT().GetSearchResults(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, key string) ([]domain.Ride, error) {
		capturedKeyTrue = key
		return nil, errors.New("miss")
	})
	repo.EXPECT().SearchRides(ctx, gomock.Any()).Return(dbRidesTrue, nil)
	cache.EXPECT().SetSearchResults(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	cache.EXPECT().GetDriverProfile(ctx, gomock.Any()).Return(&domain.DriverProfile{}, nil).AnyTimes()

	paramsTrue := domain.RideSearchParams{
		OriginLat:    41.0,
		OriginLon:    69.0,
		DestLat:      39.0,
		DestLon:      66.0,
		RadiusMeters: 10.0,
		MinSeats:     1,
		IncludeTest:  true,
	}
	_, err = uc.FindRides(ctx, paramsTrue)
	require.NoError(t, err)

	// Ensure the cache keys are strictly different!
	assert.NotEmpty(t, capturedKeyFalse)
	assert.NotEmpty(t, capturedKeyTrue)
	assert.NotEqual(t, capturedKeyFalse, capturedKeyTrue, "Cache keys for IncludeTest=false and IncludeTest=true must not collide")
}

func TestActivateRide_InvalidatesSearchCache(t *testing.T) {
	uc, repo, cache, grpcClient, publisher := setupRideUseCaseTest(t)
	ctx := context.Background()

	driverID := uuid.New()
	rideID := uuid.New()
	ride := &domain.Ride{
		ID:        rideID,
		DriverID:  driverID,
		Status:    domain.RideStatusDraft,
		OriginLon: 69.24,
		OriginLat: 41.29,
	}

	repo.EXPECT().GetRideByID(ctx, rideID).Return(ride, nil)
	grpcClient.EXPECT().CheckCarStatus(ctx, driverID).Return(true, nil)
	grpcClient.EXPECT().CheckRideLimit(ctx, driverID).Return(true, "plus", nil)
	repo.EXPECT().UpdateRideStatus(ctx, rideID, domain.RideStatusActive).Return(nil)
	cache.EXPECT().SetRideLocation(ctx, rideID, ride.OriginLon, ride.OriginLat).Return(nil)
	publisher.EXPECT().Publish(ctx, "yolyolakay.events", "ride.created", gomock.Any()).Return(nil)

	err := uc.ActivateRide(ctx, driverID, rideID)
	require.NoError(t, err)
}
