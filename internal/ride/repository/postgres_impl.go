package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hextec/yolyolakay-backend/internal/ride/domain"
	sqlc "github.com/hextec/yolyolakay-backend/internal/ride/repository/postgres"
	"github.com/hextec/yolyolakay-backend/pkg/database"
)

type postgresRideRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewPostgresRideRepository(db *pgxpool.Pool) domain.RideRepository {
	return &postgresRideRepository{
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

func float64ToNumeric(f float64) pgtype.Numeric {
	return database.Float64ToNumeric(f)
}

func float64PtrToNumeric(f *float64) pgtype.Numeric {
	return database.Float64PtrToNumeric(f)
}

func numericToFloat64Ptr(n pgtype.Numeric) *float64 {
	return database.NumericToFloat64Ptr(n)
}

func (r *postgresRideRepository) GetAdminStats(ctx context.Context) (*domain.AdminStats, error) {
	totalRides, err := r.queries.GetTotalRidesCount(ctx)
	if err != nil {
		return nil, err
	}

	activeRides, err := r.queries.GetActiveRidesCount(ctx)
	if err != nil {
		return nil, err
	}

	popularPG, err := r.queries.GetPopularRoutes(ctx)
	if err != nil {
		return nil, err
	}

	var popularRoutes []domain.PopularRoute
	for _, p := range popularPG {
		popularRoutes = append(popularRoutes, domain.PopularRoute{
			Origin:      p.OriginAddress,
			Destination: p.DestinationAddress,
			Count:       p.Count,
		})
	}

	chartPG, err := r.queries.GetRidesChartData(ctx)
	if err != nil {
		return nil, err
	}

	var chartData []domain.ChartDataPoint
	for _, c := range chartPG {
		chartData = append(chartData, domain.ChartDataPoint{
			Date:  c.Date,
			Count: c.Count,
		})
	}

	return &domain.AdminStats{
		TotalRides:    totalRides,
		ActiveRides:   activeRides,
		PopularRoutes: popularRoutes,
		ChartData:     chartData,
	}, nil
}

func (r *postgresRideRepository) CreateRide(ctx context.Context, ride *domain.Ride) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		slog.Error("Failed to begin tx for create ride", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	num := float64ToNumeric(ride.PricePerSeat)

	var vehicleDataBytes []byte
	if ride.VehicleData != nil {
		vehicleDataBytes, _ = json.Marshal(ride.VehicleData)
	}

	var musicTasteText pgtype.Text
	if ride.MusicTaste != nil && *ride.MusicTaste != "" {
		musicTasteText = pgtype.Text{String: *ride.MusicTaste, Valid: true}
	}

	_, err = qtx.CreateRide(ctx, sqlc.CreateRideParams{
		ID:                 uuidToPg(ride.ID),
		DriverID:           uuidToPg(ride.DriverID),
		VehicleID:          uuidToPg(ride.VehicleID),
		VehicleData:        vehicleDataBytes,
		StMakepoint:        ride.OriginLon,
		StMakepoint_2:      ride.OriginLat,
		StMakepoint_3:      ride.DestinationLon,
		StMakepoint_4:      ride.DestinationLat,
		OriginAddress:      ride.OriginAddress,
		DestinationAddress: ride.DestinationAddress,
		DepartureTime:      pgtype.Timestamptz{Time: ride.DepartureTime, Valid: true},
		TotalSeats:         ride.TotalSeats,
		AvailableSeats:     ride.AvailableSeats,
		PricePerSeat:       num,
		GenderRestriction:  pgtype.Text{String: ride.GenderRestriction, Valid: true},
		HasAc:              ride.HasAc,
		HasLuggage:         ride.HasLuggage,
		TakesParcels:       ride.TakesParcels,
		PickupPoint:        ride.PickupPoint,
		PickupLat:          float64PtrToFloat8(ride.PickupLat),
		PickupLon:          float64PtrToFloat8(ride.PickupLon),
		MusicTaste:         musicTasteText,
		NoSmoking:          ride.NoSmoking,
		NoPets:             ride.NoPets,
		DrivingSpeed:       pgtype.Text{String: ride.DrivingSpeed, Valid: ride.DrivingSpeed != ""},
		MaxTwoBack:         ride.MaxTwoBack,
		Status:             string(ride.Status),
		IsNegotiable:       ride.IsNegotiable,
		HasFairPriceBadge:  ride.HasFairPriceBadge,
		AllowsCityDetour:   pgtype.Bool{Bool: ride.AllowsCityDetour, Valid: true},
		CityDetourFee:      float64ToNumeric(ride.CityDetourFee),
	})
	if err != nil {
		slog.Error("Error CreateRide", "error", err)
		return err
	}

	for _, stop := range ride.Stops {
		err = qtx.InsertRideStop(ctx, sqlc.InsertRideStopParams{
			ID:            uuidToPg(stop.ID),
			RideID:        uuidToPg(ride.ID),
			StMakepoint:   stop.Lon,
			StMakepoint_2: stop.Lat,
			Address:       pgtype.Text{String: stop.Address, Valid: true},
			OrderIndex:    stop.OrderIndex,
		})
		if err != nil {
			slog.Error("Error InsertRideStop", "error", err)
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

func float64PtrToFloat8(f *float64) pgtype.Float8 {
	if f == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *f, Valid: true}
}

func mapFieldsToRide(
	id, driverID, vehicleID pgtype.UUID,
	vehicleData []byte,
	originLat, originLon, destLat, destLon interface{},
	originAddress, destAddress string,
	departureTime pgtype.Timestamptz,
	totalSeats, availableSeats int32,
	pricePerSeat pgtype.Numeric,
	status string,
	isNegotiable, hasFairPriceBadge bool,
	allowsCityDetour pgtype.Bool,
	cityDetourFee pgtype.Numeric,
	genderRestriction pgtype.Text,
	hasAc, hasLuggage, takesParcels bool,
	pickupPoint string,
	pickupLat, pickupLon pgtype.Float8,
	musicTaste pgtype.Text,
	noSmoking, noPets bool,
	drivingSpeed string,
	maxTwoBack bool,
	cancelReason pgtype.Text,
	cancelledAt pgtype.Timestamptz,
	createdAt, updatedAt pgtype.Timestamptz,
) domain.Ride {
	var price float64
	if pricePerSeat.Valid {
		val, _ := pricePerSeat.Float64Value()
		price = val.Float64
	}

	var detourFee float64
	if cityDetourFee.Valid {
		val, _ := cityDetourFee.Float64Value()
		detourFee = val.Float64
	}

	var cancelledAtTime *time.Time
	if cancelledAt.Valid {
		t := cancelledAt.Time
		cancelledAtTime = &t
	}

	var musicTastePtr *string
	if musicTaste.Valid && musicTaste.String != "" {
		musicTastePtr = &musicTaste.String
	}

	var pLat *float64
	if pickupLat.Valid {
		pLat = &pickupLat.Float64
	}
	var pLon *float64
	if pickupLon.Valid {
		pLon = &pickupLon.Float64
	}

	return domain.Ride{
		ID:                 pgToUUID(id),
		DriverID:           pgToUUID(driverID),
		VehicleID:          pgToUUID(vehicleID),
		VehicleData:        parseVehicleData(vehicleData),
		OriginLat:          parseFloatFromInterface(originLat),
		OriginLon:          parseFloatFromInterface(originLon),
		DestinationLat:     parseFloatFromInterface(destLat),
		DestinationLon:     parseFloatFromInterface(destLon),
		OriginAddress:      originAddress,
		DestinationAddress: destAddress,
		DepartureTime:      departureTime.Time,
		TotalSeats:         totalSeats,
		AvailableSeats:     availableSeats,
		PricePerSeat:       price,
		Status:             domain.RideStatus(status),
		IsNegotiable:       isNegotiable,
		HasFairPriceBadge:  hasFairPriceBadge,
		AllowsCityDetour:   allowsCityDetour.Bool,
		CityDetourFee:      detourFee,
		GenderRestriction:  genderRestriction.String,
		HasAc:              hasAc,
		HasLuggage:         hasLuggage,
		TakesParcels:       takesParcels,
		PickupPoint:        pickupPoint,
		PickupLat:          pLat,
		PickupLon:          pLon,
		MusicTaste:         musicTastePtr,
		NoSmoking:          noSmoking,
		NoPets:             noPets,
		DrivingSpeed:       drivingSpeed,
		MaxTwoBack:         maxTwoBack,
		CancelReason:       cancelReason.String,
		CancelledAt:        cancelledAtTime,
		CreatedAt:          createdAt.Time,
		UpdatedAt:          updatedAt.Time,
	}
}

func parseVehicleData(data []byte) *domain.Vehicle {
	if len(data) == 0 {
		return nil
	}
	var v domain.Vehicle
	if err := json.Unmarshal(data, &v); err != nil {
		return nil
	}
	return &v
}

func mapSearchRide(pgRide sqlc.SearchRidesRow) domain.Ride {
	return mapFieldsToRide(
		pgRide.ID, pgRide.DriverID, pgRide.VehicleID, pgRide.VehicleData,
		pgRide.OriginLat, pgRide.OriginLon, pgRide.DestLat, pgRide.DestLon,
		pgRide.OriginAddress, pgRide.DestinationAddress, pgRide.DepartureTime,
		pgRide.TotalSeats, pgRide.AvailableSeats, pgRide.PricePerSeat,
		pgRide.Status, pgRide.IsNegotiable, pgRide.HasFairPriceBadge,
		pgRide.AllowsCityDetour, pgRide.CityDetourFee,
		pgRide.GenderRestriction, pgRide.HasAc, pgRide.HasLuggage, pgRide.TakesParcels,
		pgRide.PickupPoint, pgRide.PickupLat, pgRide.PickupLon, pgRide.MusicTaste,
		pgRide.NoSmoking, pgRide.NoPets, pgRide.DrivingSpeed.String, pgRide.MaxTwoBack,
		pgRide.CancelReason, pgRide.CancelledAt, pgRide.CreatedAt, pgRide.UpdatedAt,
	)
}

func mapUpdateRide(pgRide sqlc.GetRideForUpdateRow) domain.Ride {
	return mapFieldsToRide(
		pgRide.ID, pgRide.DriverID, pgRide.VehicleID, pgRide.VehicleData,
		pgRide.OriginLat, pgRide.OriginLon, pgRide.DestLat, pgRide.DestLon,
		pgRide.OriginAddress, pgRide.DestinationAddress, pgRide.DepartureTime,
		pgRide.TotalSeats, pgRide.AvailableSeats, pgRide.PricePerSeat,
		pgRide.Status, pgRide.IsNegotiable, pgRide.HasFairPriceBadge,
		pgRide.AllowsCityDetour, pgRide.CityDetourFee,
		pgRide.GenderRestriction, pgRide.HasAc, pgRide.HasLuggage, pgRide.TakesParcels,
		pgRide.PickupPoint, pgRide.PickupLat, pgRide.PickupLon, pgRide.MusicTaste,
		pgRide.NoSmoking, pgRide.NoPets, pgRide.DrivingSpeed.String, pgRide.MaxTwoBack,
		pgRide.CancelReason, pgRide.CancelledAt, pgRide.CreatedAt, pgRide.UpdatedAt,
	)
}

func parseFloatFromInterface(i interface{}) float64 {
	if i == nil {
		return 0.0
	}
	switch v := i.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0.0
	}
}

func mapGetRideByID(pgRide sqlc.GetRideByIDRow) domain.Ride {
	return mapFieldsToRide(
		pgRide.ID, pgRide.DriverID, pgRide.VehicleID, pgRide.VehicleData,
		pgRide.OriginLat, pgRide.OriginLon, pgRide.DestLat, pgRide.DestLon,
		pgRide.OriginAddress, pgRide.DestinationAddress, pgRide.DepartureTime,
		pgRide.TotalSeats, pgRide.AvailableSeats, pgRide.PricePerSeat,
		pgRide.Status, pgRide.IsNegotiable, pgRide.HasFairPriceBadge,
		pgRide.AllowsCityDetour, pgRide.CityDetourFee,
		pgRide.GenderRestriction, pgRide.HasAc, pgRide.HasLuggage, pgRide.TakesParcels,
		pgRide.PickupPoint, pgRide.PickupLat, pgRide.PickupLon, pgRide.MusicTaste,
		pgRide.NoSmoking, pgRide.NoPets, pgRide.DrivingSpeed.String, pgRide.MaxTwoBack,
		pgRide.CancelReason, pgRide.CancelledAt, pgRide.CreatedAt, pgRide.UpdatedAt,
	)
}

func (r *postgresRideRepository) loadStopsForRide(ctx context.Context, ride *domain.Ride) error {
	pgStops, err := r.queries.GetRideStopsByRideID(ctx, uuidToPg(ride.ID))
	if err != nil {
		return err
	}
	var stops []domain.RideStop
	for _, pgs := range pgStops {
		stops = append(stops, domain.RideStop{
			ID:         pgToUUID(pgs.ID),
			RideID:     pgToUUID(pgs.RideID),
			Lat:        parseFloatFromInterface(pgs.Lat),
			Lon:        parseFloatFromInterface(pgs.Lon),
			Address:    pgs.Address.String,
			OrderIndex: pgs.OrderIndex,
		})
	}
	ride.Stops = stops
	return nil
}

func (r *postgresRideRepository) GetRideByID(ctx context.Context, rideID uuid.UUID) (*domain.Ride, error) {
	pgRide, err := r.queries.GetRideByID(ctx, uuidToPg(rideID))
	if err != nil {
		return nil, err
	}
	ride := mapGetRideByID(pgRide)

	if err := r.loadStopsForRide(ctx, &ride); err != nil {
		return nil, err
	}

	bookings, err := r.GetBookingsByRideID(ctx, rideID)
	if err == nil {
		ride.Bookings = bookings
	}

	return &ride, nil
}

func (r *postgresRideRepository) SearchRides(ctx context.Context, params domain.RideSearchParams) ([]domain.Ride, error) {
	sqlcParams := sqlc.SearchRidesParams{
		AvailableSeats:  params.MinSeats,
		StMakepoint:     params.OriginLon,
		StMakepoint_2:   params.OriginLat,
		StDwithin:       params.RadiusMeters,
		StMakepoint_3:   params.DestLon,
		StMakepoint_4:   params.DestLat,
		Limit:            params.Limit,
		Offset:           params.Offset,
		PassengerGender:  pgtype.Text{String: params.Gender, Valid: params.Gender != ""},
		PickupPreference: pgtype.Text{String: params.PickupPreference, Valid: params.PickupPreference != ""},
		IncludeTest:      pgtype.Bool{Bool: params.IncludeTest, Valid: true},
	}

	if params.PriceMin != nil {
		sqlcParams.PriceMin = pgtype.Numeric{Int: big.NewInt(int64(*params.PriceMin)), Valid: true}
	}
	if params.PriceMax != nil {
		sqlcParams.PriceMax = pgtype.Numeric{Int: big.NewInt(int64(*params.PriceMax)), Valid: true}
	}
	if params.DateFrom != nil {
		sqlcParams.DateFrom = pgtype.Timestamptz{Time: *params.DateFrom, Valid: true}
	}
	if params.DateTo != nil {
		sqlcParams.DateTo = pgtype.Timestamptz{Time: *params.DateTo, Valid: true}
	}

	pgRides, err := r.queries.SearchRides(ctx, sqlcParams)
	if err != nil {
		return nil, err
	}

	if len(pgRides) == 0 {
		return []domain.Ride{}, nil
	}

	var rideIDs []pgtype.UUID
	for _, p := range pgRides {
		rideIDs = append(rideIDs, p.ID)
	}

	pgStops, err := r.queries.GetRideStopsByRideIDs(ctx, rideIDs)
	if err != nil {
		return nil, err
	}

	stopsByRide := make(map[uuid.UUID][]domain.RideStop)
	for _, pgs := range pgStops {
		rID := pgToUUID(pgs.RideID)
		stopsByRide[rID] = append(stopsByRide[rID], domain.RideStop{
			ID:         pgToUUID(pgs.ID),
			RideID:     rID,
			Lat:        parseFloatFromInterface(pgs.Lat),
			Lon:        parseFloatFromInterface(pgs.Lon),
			Address:    pgs.Address.String,
			OrderIndex: pgs.OrderIndex,
		})
	}

	var rides []domain.Ride
	for _, p := range pgRides {
		ride := mapSearchRide(p)
		if stops, ok := stopsByRide[ride.ID]; ok {
			ride.Stops = stops
		} else {
			ride.Stops = []domain.RideStop{}
		}
		rides = append(rides, ride)
	}

	return rides, nil
}

func mapActiveDestRide(pgRide sqlc.GetActiveRidesToDestinationsRow) domain.Ride {
	return mapFieldsToRide(
		pgRide.ID, pgRide.DriverID, pgRide.VehicleID, pgRide.VehicleData,
		pgRide.OriginLat, pgRide.OriginLon, pgRide.DestLat, pgRide.DestLon,
		pgRide.OriginAddress, pgRide.DestinationAddress, pgRide.DepartureTime,
		pgRide.TotalSeats, pgRide.AvailableSeats, pgRide.PricePerSeat,
		pgRide.Status, pgRide.IsNegotiable, pgRide.HasFairPriceBadge,
		pgRide.AllowsCityDetour, pgRide.CityDetourFee,
		pgRide.GenderRestriction, pgRide.HasAc, pgRide.HasLuggage, pgRide.TakesParcels,
		pgRide.PickupPoint, pgRide.PickupLat, pgRide.PickupLon, pgRide.MusicTaste,
		pgRide.NoSmoking, pgRide.NoPets, pgRide.DrivingSpeed.String, pgRide.MaxTwoBack,
		pgRide.CancelReason, pgRide.CancelledAt, pgRide.CreatedAt, pgRide.UpdatedAt,
	)
}

func mapGetRidesByIDs(pgRide sqlc.GetRidesByIDsRow) domain.Ride {
	return mapFieldsToRide(
		pgRide.ID, pgRide.DriverID, pgRide.VehicleID, pgRide.VehicleData,
		pgRide.OriginLat, pgRide.OriginLon, pgRide.DestLat, pgRide.DestLon,
		pgRide.OriginAddress, pgRide.DestinationAddress, pgRide.DepartureTime,
		pgRide.TotalSeats, pgRide.AvailableSeats, pgRide.PricePerSeat,
		pgRide.Status, pgRide.IsNegotiable, pgRide.HasFairPriceBadge,
		pgRide.AllowsCityDetour, pgRide.CityDetourFee,
		pgRide.GenderRestriction, pgRide.HasAc, pgRide.HasLuggage, pgRide.TakesParcels,
		pgRide.PickupPoint, pgRide.PickupLat, pgRide.PickupLon, pgRide.MusicTaste,
		pgRide.NoSmoking, pgRide.NoPets, pgRide.DrivingSpeed.String, pgRide.MaxTwoBack,
		pgRide.CancelReason, pgRide.CancelledAt, pgRide.CreatedAt, pgRide.UpdatedAt,
	)
}

func (r *postgresRideRepository) GetRidesByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Ride, error) {
	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pgIDs[i] = uuidToPg(id)
	}

	pgRides, err := r.queries.GetRidesByIDs(ctx, pgIDs)
	if err != nil {
		return nil, err
	}

	var rides []domain.Ride
	for _, p := range pgRides {
		rides = append(rides, mapGetRidesByIDs(p))
	}

	return rides, nil
}

func (r *postgresRideRepository) CreateBookingTx(ctx context.Context, rideID, passengerID uuid.UUID, seats int32, isFullCarBuyout bool, offeredPrice *float64, pickupPreference, dropoffType string) (*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgRide, err := qtx.GetRideForUpdate(ctx, uuidToPg(rideID))
	if err != nil {
		return nil, fmt.Errorf("failed to get ride or ride not active: %w", err)
	}
	ride := mapUpdateRide(pgRide)

	existingBookings, err := qtx.GetBookingsByRideID(ctx, uuidToPg(rideID))
	if err != nil {
		return nil, fmt.Errorf("failed to check existing bookings: %w", err)
	}
	for _, b := range existingBookings {
		if pgToUUID(b.PassengerID) == passengerID {
			if b.Status == string(domain.BookingStatusPending) ||
				b.Status == string(domain.BookingStatusPendingNegotiation) ||
				b.Status == string(domain.BookingStatusConfirmed) ||
				b.Status == "accepted" {
				return nil, errors.New("you already have an active booking for this ride")
			}
		}
	}

	finalSeats := seats
	if isFullCarBuyout {
		finalSeats = ride.AvailableSeats
		if finalSeats <= 0 {
			return nil, errors.New("no seats available for buyout")
		}
	} else {
		if finalSeats > ride.AvailableSeats {
			return nil, errors.New("not enough available seats")
		}
	}

	if dropoffType == "" {
		dropoffType = "highway"
	}
	if dropoffType == "door" {
		dropoffType = "city_detour"
	}

	status := domain.BookingStatusPending
	totalPrice := float64(finalSeats) * ride.PricePerSeat
	if offeredPrice != nil {
		if !ride.IsNegotiable {
			return nil, errors.New("ride is not negotiable")
		}
		status = domain.BookingStatusPendingNegotiation
		totalPrice = *offeredPrice
	} else if dropoffType == "city_detour" && ride.AllowsCityDetour {
		totalPrice += ride.CityDetourFee
	}

	// NOTE: We do NOT subtract seats at booking creation time!
	// Seats will be subtracted only when the driver accepts the booking in AcceptBookingTx.

	bookingID := uuid.New()
	pgBooking, err := qtx.CreateBooking(ctx, sqlc.CreateBookingParams{
		ID:               uuidToPg(bookingID),
		RideID:           uuidToPg(rideID),
		PassengerID:      uuidToPg(passengerID),
		SeatsBooked:      finalSeats,
		TotalPrice:       float64ToNumeric(totalPrice),
		IsFullCarBuyout:  isFullCarBuyout,
		OfferedPrice:     float64PtrToNumeric(offeredPrice),
		Status:           string(status),
		PickupPreference: pgtype.Text{String: pickupPreference, Valid: pickupPreference != ""},
		DropoffType:      pgtype.Text{String: dropoffType, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save booking: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	val, _ := pgBooking.TotalPrice.Float64Value()

	return &domain.Booking{
		ID:               bookingID,
		RideID:           rideID,
		PassengerID:      passengerID,
		SeatsBooked:      finalSeats,
		TotalPrice:       val.Float64,
		IsFullCarBuyout:  isFullCarBuyout,
		OfferedPrice:     numericToFloat64Ptr(pgBooking.OfferedPrice),
		Status:           domain.BookingStatus(pgBooking.Status),
		PickupPreference: pgBooking.PickupPreference.String,
		DropoffType:      pgBooking.DropoffType.String,
		CreatedAt:        pgBooking.CreatedAt.Time,
		UpdatedAt:        pgBooking.UpdatedAt.Time,
	}, nil
}

func (r *postgresRideRepository) AcceptBookingTx(ctx context.Context, rideID, bookingID uuid.UUID) (*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgRide, err := qtx.GetRideForUpdate(ctx, uuidToPg(rideID))
	if err != nil {
		return nil, fmt.Errorf("failed to get ride or ride not active: %w", err)
	}
	ride := mapUpdateRide(pgRide)

	pgBooking, err := qtx.GetBookingForUpdate(ctx, uuidToPg(bookingID))
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	if pgBooking.Status != string(domain.BookingStatusPendingNegotiation) && pgBooking.Status != string(domain.BookingStatusPending) {
		return nil, errors.New("booking is not in a valid state to be accepted")
	}

	// Seats are deducted at AcceptBookingTx time for both standard and negotiated bookings
	if ride.AvailableSeats < pgBooking.SeatsBooked {
		return nil, errors.New("not enough available seats to accept this booking")
	}

	err = qtx.UpdateRideSeats(ctx, sqlc.UpdateRideSeatsParams{
		AvailableSeats: pgBooking.SeatsBooked,
		ID:             uuidToPg(rideID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update seats: %w", err)
	}

	err = qtx.UpdateBookingStatus(ctx, sqlc.UpdateBookingStatusParams{
		Status: string(domain.BookingStatusConfirmed),
		ID:     uuidToPg(bookingID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update booking status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	val, _ := pgBooking.TotalPrice.Float64Value()

	return &domain.Booking{
		ID:              pgToUUID(pgBooking.ID),
		RideID:          pgToUUID(pgBooking.RideID),
		PassengerID:     pgToUUID(pgBooking.PassengerID),
		SeatsBooked:     pgBooking.SeatsBooked,
		TotalPrice:      val.Float64,
		IsFullCarBuyout: pgBooking.IsFullCarBuyout,
		OfferedPrice:    numericToFloat64Ptr(pgBooking.OfferedPrice),
		Status:          domain.BookingStatusConfirmed,
		CreatedAt:       pgBooking.CreatedAt.Time,
		UpdatedAt:       time.Now(),
	}, nil
}

func (r *postgresRideRepository) CancelBookingTx(ctx context.Context, passengerID, bookingID uuid.UUID) (*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgBooking, err := qtx.GetBookingForUpdate(ctx, uuidToPg(bookingID))
	if err != nil {
		return nil, fmt.Errorf("failed to get booking for update: %w", err)
	}

	if pgToUUID(pgBooking.PassengerID) != passengerID {
		return nil, errors.New("unauthorized: not your booking")
	}

	if pgBooking.Status == string(domain.BookingStatusCanceled) || pgBooking.Status == string(domain.BookingStatusPassengerNoShow) {
		return nil, errors.New("booking is already canceled or no-show")
	}

	// Passenger cannot cancel booking once the ride has started (in_progress) or is completed
	pgRide, err := qtx.GetRideByID(ctx, pgBooking.RideID)
	if err == nil {
		if pgRide.Status == string(domain.RideStatusInProgress) {
			return nil, errors.New("cannot cancel booking: ride is already in progress")
		}
		if pgRide.Status == string(domain.RideStatusCompleted) {
			return nil, errors.New("cannot cancel booking: ride is already completed")
		}
	}

	// Seats are restored ONLY if the booking was confirmed (seats were actually deducted).
	// For pending/pending_negotiation, no seats were deducted, so no seats to return.
	if pgBooking.Status == string(domain.BookingStatusConfirmed) || pgBooking.Status == "accepted" {
		err = qtx.UpdateRideSeats(ctx, sqlc.UpdateRideSeatsParams{
			AvailableSeats: -pgBooking.SeatsBooked, 
			ID:             pgBooking.RideID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to return seats to ride: %w", err)
		}
	}

	// Обновляем статус брони
	err = qtx.UpdateBookingStatus(ctx, sqlc.UpdateBookingStatusParams{
		Status: string(domain.BookingStatusCanceled),
		ID:     uuidToPg(bookingID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update booking status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	return r.GetBookingByID(ctx, bookingID)
}

func (r *postgresRideRepository) UpdateBookingStatus(ctx context.Context, id uuid.UUID, status string) error {
	pgID := pgtype.UUID{Bytes: id, Valid: true}
	err := r.queries.UpdateBookingStatus(ctx, sqlc.UpdateBookingStatusParams{
		ID:     pgID,
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}
	return nil
}

func (r *postgresRideRepository) ReportPassengerNoShowTx(ctx context.Context, rideID, bookingID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgBooking, err := qtx.GetBookingForUpdate(ctx, uuidToPg(bookingID))
	if err != nil {
		return fmt.Errorf("failed to get booking for update: %w", err)
	}

	// Restore seats ONLY if the booking was confirmed (seats were actually deducted).
	if pgBooking.Status == string(domain.BookingStatusConfirmed) || pgBooking.Status == "accepted" {
		err = qtx.UpdateRideSeats(ctx, sqlc.UpdateRideSeatsParams{
			AvailableSeats: -pgBooking.SeatsBooked,
			ID:             pgBooking.RideID,
		})
		if err != nil {
			return fmt.Errorf("failed to return seats to ride: %w", err)
		}
	}

	err = qtx.UpdateBookingStatus(ctx, sqlc.UpdateBookingStatusParams{
		Status: string(domain.BookingStatusPassengerNoShow),
		ID:     uuidToPg(bookingID),
	})
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *postgresRideRepository) GetBookingByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	pgID := pgtype.UUID{Bytes: id, Valid: true}
	row, err := r.queries.GetBookingByID(ctx, pgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	var totalPrice float64
	if row.TotalPrice.Valid {
		val, _ := row.TotalPrice.Float64Value()
		totalPrice = val.Float64
	}

	var offeredPricePtr *float64
	if row.OfferedPrice.Valid {
		v, _ := row.OfferedPrice.Float64Value()
		offeredPricePtr = &v.Float64
	}

	return &domain.Booking{
		ID:               row.ID.Bytes,
		RideID:           row.RideID.Bytes,
		PassengerID:      row.PassengerID.Bytes,
		SeatsBooked:      row.SeatsBooked,
		TotalPrice:       totalPrice,
		IsFullCarBuyout:  row.IsFullCarBuyout,
		OfferedPrice:     offeredPricePtr,
		Status:           domain.BookingStatus(row.Status),
		PickupPreference: row.PickupPreference.String,
		DropoffType:      row.DropoffType.String,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}, nil
}

func (r *postgresRideRepository) GetBookingsByRideID(ctx context.Context, rideID uuid.UUID) ([]domain.Booking, error) {
	pgID := pgtype.UUID{Bytes: rideID, Valid: true}
	rows, err := r.queries.GetBookingsByRideID(ctx, pgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings by ride id: %w", err)
	}

	var bookings []domain.Booking
	for _, row := range rows {
		var offeredPricePtr *float64
		if row.OfferedPrice.Valid {
			v, _ := row.OfferedPrice.Float64Value()
			offeredPricePtr = &v.Float64
		}
		val, _ := row.TotalPrice.Float64Value()

		bookings = append(bookings, domain.Booking{
			ID:               row.ID.Bytes,
			RideID:           row.RideID.Bytes,
			PassengerID:      row.PassengerID.Bytes,
			SeatsBooked:      row.SeatsBooked,
			TotalPrice:       val.Float64,
			IsFullCarBuyout:  row.IsFullCarBuyout,
			OfferedPrice:     offeredPricePtr,
			Status:           domain.BookingStatus(row.Status),
			PickupPreference: row.PickupPreference.String,
			DropoffType:      row.DropoffType.String,
			CreatedAt:        row.CreatedAt.Time,
			UpdatedAt:        row.UpdatedAt.Time,
		})
	}
	return bookings, nil
}

func (r *postgresRideRepository) RejectBooking(ctx context.Context, bookingID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	pgBooking, err := qtx.GetBookingForUpdate(ctx, uuidToPg(bookingID))
	if err != nil {
		return fmt.Errorf("failed to get booking for update: %w", err)
	}

	// Seats are restored ONLY if the booking was confirmed (seats were actually deducted).
	if pgBooking.Status == string(domain.BookingStatusConfirmed) || pgBooking.Status == "accepted" {
		err = qtx.UpdateRideSeats(ctx, sqlc.UpdateRideSeatsParams{
			AvailableSeats: -pgBooking.SeatsBooked,
			ID:             pgBooking.RideID,
		})
		if err != nil {
			return fmt.Errorf("failed to return seats to ride: %w", err)
		}
	}

	err = qtx.UpdateBookingStatus(ctx, sqlc.UpdateBookingStatusParams{
		Status: string(domain.BookingStatusRejected),
		ID:     uuidToPg(bookingID),
	})
	if err != nil {
		return fmt.Errorf("failed to reject booking: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *postgresRideRepository) DeleteRidesByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteRidesByDriverID(ctx, uuidToPg(userID))
}

func (r *postgresRideRepository) UpdateRideStatus(ctx context.Context, rideID uuid.UUID, status domain.RideStatus) error {
	params := sqlc.UpdateRideStatusParams{
		ID:     uuidToPg(rideID),
		Status: string(status),
	}
	err := r.queries.UpdateRideStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update ride status: %w", err)
	}
	return nil
}

func (r *postgresRideRepository) RejectExpiredNegotiations(ctx context.Context) ([]domain.RejectedBooking, error) {
	rows, err := r.queries.RejectExpiredNegotiations(ctx)
	if err != nil {
		return nil, err
	}

	var results []domain.RejectedBooking
	for _, row := range rows {
		results = append(results, domain.RejectedBooking{
			ID:          row.ID.Bytes,
			PassengerID: row.PassengerID.Bytes,
			RideID:      row.RideID.Bytes,
		})
	}
	return results, nil
}

func (r *postgresRideRepository) CountPassengerNoShows(ctx context.Context, passengerID uuid.UUID, since time.Time) (int, error) {
	pgPassengerID := pgtype.UUID{Bytes: passengerID, Valid: true}
	pgSince := pgtype.Timestamptz{Time: since, Valid: true}

	count, err := r.queries.CountPassengerNoShows(ctx, sqlc.CountPassengerNoShowsParams{
		PassengerID: pgPassengerID,
		CreatedAt:   pgSince,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count passenger no shows: %w", err)
	}
	return int(count), nil
}

func (r *postgresRideRepository) GetRidesByDriverID(ctx context.Context, driverID uuid.UUID) ([]domain.Ride, error) {
	pgRides, err := r.queries.GetRidesByDriverID(ctx, uuidToPg(driverID))
	if err != nil {
		return nil, fmt.Errorf("failed to get rides by driver id: %w", err)
	}

	var rides []domain.Ride
	for _, p := range pgRides {
		var price float64
		if p.PricePerSeat.Valid {
			val, _ := p.PricePerSeat.Float64Value()
			price = val.Float64
		}

		var detourFee float64
		if p.CityDetourFee.Valid {
			val, _ := p.CityDetourFee.Float64Value()
			detourFee = val.Float64
		}

		lat1 := parseFloatFromInterface(p.OriginLat)
		lon1 := parseFloatFromInterface(p.OriginLon)
		lat2 := parseFloatFromInterface(p.DestLat)
		lon2 := parseFloatFromInterface(p.DestLon)

		ride := domain.Ride{
			ID:                 pgToUUID(p.ID),
			DriverID:           pgToUUID(p.DriverID),
			VehicleID:          pgToUUID(p.VehicleID),
			VehicleData:        parseVehicleData(p.VehicleData),
			OriginLat:          lat1,
			OriginLon:          lon1,
			DestinationLat:     lat2,
			DestinationLon:     lon2,
			OriginAddress:      p.OriginAddress,
			DestinationAddress: p.DestinationAddress,
			DepartureTime:      p.DepartureTime.Time,
			TotalSeats:         p.TotalSeats,
			AvailableSeats:     p.AvailableSeats,
			PricePerSeat:       price,
			Status:             domain.RideStatus(p.Status),
			IsNegotiable:       p.IsNegotiable,
			HasFairPriceBadge:  p.HasFairPriceBadge,
			AllowsCityDetour:   p.AllowsCityDetour.Bool,
			CityDetourFee:      detourFee,
			GenderRestriction:  p.GenderRestriction.String,
			HasAc:              p.HasAc,
			HasLuggage:         p.HasLuggage,
			TakesParcels:       p.TakesParcels,
			NoSmoking:          p.NoSmoking,
			NoPets:             p.NoPets,
			DrivingSpeed:       p.DrivingSpeed.String,
			MaxTwoBack:         p.MaxTwoBack,
			PickupPoint:        p.PickupPoint,
			CancelReason:       p.CancelReason.String,
			CreatedAt:          p.CreatedAt.Time,
			UpdatedAt:          p.UpdatedAt.Time,
		}

		if p.PickupLat.Valid {
			lat := p.PickupLat.Float64
			ride.PickupLat = &lat
		}
		if p.PickupLon.Valid {
			lon := p.PickupLon.Float64
			ride.PickupLon = &lon
		}

		if p.MusicTaste.Valid && p.MusicTaste.String != "" {
			ride.MusicTaste = &p.MusicTaste.String
		}

		if p.CancelledAt.Valid {
			t := p.CancelledAt.Time
			ride.CancelledAt = &t
		}

		if err := r.loadStopsForRide(ctx, &ride); err != nil {
			return nil, err
		}

		rides = append(rides, ride)
	}

	return rides, nil
}

func (r *postgresRideRepository) GetPublicRidesByDriverID(ctx context.Context, driverID uuid.UUID, limit int) ([]domain.DriverPublicRide, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	pgRides, err := r.queries.GetPublicRidesByDriverID(ctx, sqlc.GetPublicRidesByDriverIDParams{
		DriverID: uuidToPg(driverID),
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get public rides by driver id: %w", err)
	}

	res := make([]domain.DriverPublicRide, len(pgRides))
	for i, p := range pgRides {
		var price float64
		if p.PricePerSeat.Valid {
			val, _ := p.PricePerSeat.Float64Value()
			price = val.Float64
		}
		res[i] = domain.DriverPublicRide{
			ID:                 pgToUUID(p.ID),
			DriverID:           pgToUUID(p.DriverID),
			OriginAddress:      p.OriginAddress,
			DestinationAddress: p.DestinationAddress,
			OriginLat:          parseFloatFromInterface(p.OriginLat),
			OriginLon:          parseFloatFromInterface(p.OriginLon),
			DestLat:            parseFloatFromInterface(p.DestLat),
			DestLon:            parseFloatFromInterface(p.DestLon),
			DepartureTime:      p.DepartureTime.Time,
			PricePerSeat:       price,
			Status:             p.Status,
		}
	}
	return res, nil
}

func (r *postgresRideRepository) GetBookingsByPassengerID(ctx context.Context, passengerID uuid.UUID) ([]domain.Booking, error) {
	pgBookings, err := r.queries.GetBookingsByPassengerID(ctx, uuidToPg(passengerID))
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings by passenger id: %w", err)
	}

	var bookings []domain.Booking
	for _, row := range pgBookings {
		var offeredPricePtr *float64
		if row.OfferedPrice.Valid {
			v, _ := row.OfferedPrice.Float64Value()
			offeredPricePtr = &v.Float64
		}
		val, _ := row.TotalPrice.Float64Value()

		bookings = append(bookings, domain.Booking{
			ID:               row.ID.Bytes,
			RideID:           row.RideID.Bytes,
			PassengerID:      row.PassengerID.Bytes,
			SeatsBooked:      row.SeatsBooked,
			TotalPrice:       val.Float64,
			IsFullCarBuyout:  row.IsFullCarBuyout,
			OfferedPrice:     offeredPricePtr,
			Status:           domain.BookingStatus(row.Status),
			PickupPreference: row.PickupPreference.String,
			DropoffType:      row.DropoffType.String,
			CreatedAt:        row.CreatedAt.Time,
			UpdatedAt:        row.UpdatedAt.Time,
		})
	}
	return bookings, nil
}

// --- Route Subscriptions ---

type postgresRouteSubscriptionRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRouteSubscriptionRepository(db *pgxpool.Pool) domain.RouteSubscriptionRepository {
	return &postgresRouteSubscriptionRepo{db: db}
}

func (r *postgresRouteSubscriptionRepo) CreateSubscription(ctx context.Context, sub *domain.RouteSubscription) error {
	query := `
		INSERT INTO route_subscriptions (
			id, passenger_id, origin_lat, origin_lon, dest_lat, dest_lon, departure_date, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`
	_, err := r.db.Exec(ctx, query,
		sub.ID, sub.PassengerID, sub.OriginLat, sub.OriginLon, sub.DestLat, sub.DestLon,
		sub.DepartureDate, sub.IsActive, sub.CreatedAt, sub.UpdatedAt,
	)
	return err
}

func (r *postgresRideRepository) UpdateBookingPriceAndStatus(ctx context.Context, id uuid.UUID, price float64, status string) error {
	err := r.queries.UpdateBookingPriceAndStatus(ctx, sqlc.UpdateBookingPriceAndStatusParams{
		Status:       status,
		OfferedPrice: float64ToNumeric(price),
		ID:           uuidToPg(id),
	})
	if err != nil {
		if err.Error() == "no rows in result set" {
			return errors.New("not found")
		}
		return err
	}
	return nil
}

func (r *postgresRouteSubscriptionRepo) GetSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.RouteSubscription, error) {
	query := `
		SELECT id, passenger_id, origin_lat, origin_lon, dest_lat, dest_lon, departure_date, is_active, created_at, updated_at
		FROM route_subscriptions
		WHERE passenger_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.RouteSubscription
	for rows.Next() {
		var sub domain.RouteSubscription
		err := rows.Scan(
			&sub.ID, &sub.PassengerID, &sub.OriginLat, &sub.OriginLon,
			&sub.DestLat, &sub.DestLon, &sub.DepartureDate, &sub.IsActive,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (r *postgresRouteSubscriptionRepo) DeleteSubscription(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM route_subscriptions WHERE id = $1 AND passenger_id = $2`
	res, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("subscription not found")
	}
	return nil
}

func (r *postgresRouteSubscriptionRepo) FindMatchingSubscriptions(ctx context.Context, originLat, originLon, destLat, destLon float64, departureTime time.Time, radiusKm float64) ([]domain.RouteSubscription, error) {
	query := `
		SELECT id, passenger_id, origin_lat, origin_lon, dest_lat, dest_lon, departure_date, is_active, created_at, updated_at
		FROM route_subscriptions
		WHERE is_active = true
		  AND (
			6371 * acos(
				cos(radians($1)) * cos(radians(origin_lat)) * cos(radians(origin_lon) - radians($2)) +
				sin(radians($1)) * sin(radians(origin_lat))
			)
		  ) <= $3
		  AND (
			6371 * acos(
				cos(radians($4)) * cos(radians(dest_lat)) * cos(radians(dest_lon) - radians($5)) +
				sin(radians($4)) * sin(radians(dest_lat))
			)
		  ) <= $3
		  AND (departure_date IS NULL OR departure_date = $6::DATE)
	`
	rows, err := r.db.Query(ctx, query, originLat, originLon, radiusKm, destLat, destLon, departureTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.RouteSubscription
	for rows.Next() {
		var sub domain.RouteSubscription
		err := rows.Scan(
			&sub.ID, &sub.PassengerID, &sub.OriginLat, &sub.OriginLon,
			&sub.DestLat, &sub.DestLon, &sub.DepartureDate, &sub.IsActive,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (r *postgresRideRepository) GetDriverStatisticsAggregated(ctx context.Context, driverID uuid.UUID) (*domain.DriverStatistics, error) {
	stats, err := r.queries.GetDriverStatisticsAggregated(ctx, uuidToPg(driverID))
	if err != nil {
		return nil, err
	}
	return &domain.DriverStatistics{
		TotalRides:     int(stats.TotalRides),
		CompletedRides: int(stats.CompletedRides),
		CancelledRides: int(stats.CancelledRides),
		TotalEarnings:  stats.TotalEarnings,
	}, nil
}

func (r *postgresRideRepository) GetUserMostVisitedDestinations(ctx context.Context, userID uuid.UUID, limit int32) ([]domain.RouteStats, error) {
	pgDests, err := r.queries.GetUserMostVisitedDestinations(ctx, sqlc.GetUserMostVisitedDestinationsParams{
		PassengerID: uuidToPg(userID),
		Limit:       limit,
	})
	if err != nil {
		return nil, err
	}

	var stats []domain.RouteStats
	for _, d := range pgDests {
		stats = append(stats, domain.RouteStats{
			From:       d.OriginAddress,
			To:         d.DestinationAddress,
			TripsCount: int(d.VisitCount),
		})
	}
	return stats, nil
}

func (r *postgresRideRepository) GetTopGlobalDestinations(ctx context.Context, limit int32) ([]domain.RouteStats, error) {
	pgDests, err := r.queries.GetTopGlobalDestinations(ctx, limit)
	if err != nil {
		return nil, err
	}

	var stats []domain.RouteStats
	for _, d := range pgDests {
		stats = append(stats, domain.RouteStats{
			From:       d.OriginAddress,
			To:         d.DestinationAddress,
			TripsCount: int(d.VisitCount),
		})
	}
	return stats, nil
}

func (r *postgresRideRepository) GetActiveRidesToDestinations(ctx context.Context, destinations []string, limit int32) ([]domain.Ride, error) {
	if len(destinations) == 0 {
		return []domain.Ride{}, nil
	}

	pgRides, err := r.queries.GetActiveRidesToDestinations(ctx, sqlc.GetActiveRidesToDestinationsParams{
		Column1: destinations,
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}

	var res []domain.Ride
	for _, pgR := range pgRides {
		var vehicleID uuid.UUID
		if pgR.VehicleID.Valid {
			vehicleID = pgToUUID(pgR.VehicleID)
		}

		val, _ := pgR.PricePerSeat.Float64Value()
		price := val.Float64

		ride := domain.Ride{
			ID:                 pgToUUID(pgR.ID),
			DriverID:           pgToUUID(pgR.DriverID),
			VehicleID:          vehicleID,
			OriginLon:          parseFloatFromInterface(pgR.OriginLon),
			OriginLat:          parseFloatFromInterface(pgR.OriginLat),
			DestinationLon:     parseFloatFromInterface(pgR.DestLon),
			DestinationLat:     parseFloatFromInterface(pgR.DestLat),
			OriginAddress:      pgR.OriginAddress,
			DestinationAddress: pgR.DestinationAddress,
			DepartureTime:      pgR.DepartureTime.Time,
			TotalSeats:         pgR.TotalSeats,
			AvailableSeats:     pgR.AvailableSeats,
			PricePerSeat:       price,
			GenderRestriction:  pgR.GenderRestriction.String,
			HasAc:              pgR.HasAc,
			HasLuggage:         pgR.HasLuggage,
			TakesParcels:       pgR.TakesParcels,
			NoSmoking:          pgR.NoSmoking,
			NoPets:             pgR.NoPets,
			DrivingSpeed:       pgR.DrivingSpeed.String,
			MaxTwoBack:         pgR.MaxTwoBack,
			Status:             domain.RideStatus(pgR.Status),
			IsNegotiable:       pgR.IsNegotiable,
			HasFairPriceBadge:  pgR.HasFairPriceBadge,
			CancelReason:       pgR.CancelReason.String,
			CreatedAt:          pgR.CreatedAt.Time,
			UpdatedAt:          pgR.UpdatedAt.Time,
		}

		if pgR.CancelledAt.Valid {
			t := pgR.CancelledAt.Time
			ride.CancelledAt = &t
		}

		res = append(res, ride)
	}
	return res, nil
}

func (r *postgresRideRepository) UpdateRideRoutePolygon(ctx context.Context, rideID uuid.UUID, wkt string, radius float64) error {
	err := r.queries.UpdateRideRoutePolygon(ctx, sqlc.UpdateRideRoutePolygonParams{
		ID:             uuidToPg(rideID),
		StGeomfromtext: wkt,
		Column3:        radius,
	})
	if err != nil {
		slog.Error("Failed to update route polygon", "error", err, "rideID", rideID)
		return err
	}
	return nil
}

func (r *postgresRideRepository) GetAdminActiveRides(ctx context.Context) ([]domain.Ride, error) {
	query := `
		SELECT 
			id, 
			driver_id, 
			vehicle_id, 
			ST_Y(origin::geometry) as origin_lat, 
			ST_X(origin::geometry) as origin_lon, 
			ST_Y(destination::geometry) as destination_lat, 
			ST_X(destination::geometry) as destination_lon, 
			status
		FROM rides
		WHERE status IN ('created', 'draft', 'active', 'in_progress', 'completed') AND deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rides []domain.Ride
	for rows.Next() {
		var ride domain.Ride
		var id, driverID, vehicleID pgtype.UUID
		var status string
		var oLat, oLon, dLat, dLon pgtype.Float8
		
		if err := rows.Scan(&id, &driverID, &vehicleID, &oLat, &oLon, &dLat, &dLon, &status); err != nil {
			return nil, err
		}
		
		ride.ID = pgToUUID(id)
		ride.DriverID = pgToUUID(driverID)
		ride.VehicleID = pgToUUID(vehicleID)
		ride.Status = domain.RideStatus(status)
		if oLat.Valid { ride.OriginLat = oLat.Float64 }
		if oLon.Valid { ride.OriginLon = oLon.Float64 }
		if dLat.Valid { ride.DestinationLat = dLat.Float64 }
		if dLon.Valid { ride.DestinationLon = dLon.Float64 }
		
		rides = append(rides, ride)
	}
	return rides, nil
}

func (r *postgresRideRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserRideStats, error) {
	pgID := uuidToPg(userID)
	ridesAsDriver, err := r.queries.CountUserRidesAsDriver(ctx, pgID)
	if err != nil {
		ridesAsDriver = 0
	}

	ridesAsPassenger, err := r.queries.CountUserRidesAsPassenger(ctx, pgID)
	if err != nil {
		ridesAsPassenger = 0
	}

	return &domain.UserRideStats{
		TotalRidesAsDriver:    ridesAsDriver,
		TotalRidesAsPassenger: ridesAsPassenger,
	}, nil
}
