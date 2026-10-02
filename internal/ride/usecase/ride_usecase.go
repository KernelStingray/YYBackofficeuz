package usecase

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/ride/domain"
	"github.com/hextec/yolyolakay-backend/internal/ride/integration"
	"github.com/hextec/yolyolakay-backend/pkg/sanitize"
)

type rideUseCase struct {
	repo      domain.RideRepository
	subRepo   domain.RouteSubscriptionRepository
	cache     domain.RideCache
	grpc      integration.GRPCClient
	publisher domain.EventPublisher
}

func NewRideUseCase(repo domain.RideRepository, subRepo domain.RouteSubscriptionRepository, cache domain.RideCache, grpcClient integration.GRPCClient, publisher domain.EventPublisher) domain.RideUseCase {
	return &rideUseCase{
		repo:      repo,
		subRepo:   subRepo,
		cache:     cache,
		grpc:      grpcClient,
		publisher: publisher,
	}
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in kilometers
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func (u *rideUseCase) GetAdminStats(ctx context.Context) (*domain.AdminStats, error) {
	return u.repo.GetAdminStats(ctx)
}

func (u *rideUseCase) PublishRide(ctx context.Context, ride *domain.Ride) error {
	if ride.TotalSeats <= 0 {
		return fmt.Errorf("%w: total seats must be greater than 0", domain.ErrInvalidRideData)
	}
	if ride.PricePerSeat <= 0 {
		return fmt.Errorf("%w: price must be greater than 0", domain.ErrInvalidRideData)
	}
	if ride.DepartureTime.Before(time.Now().Add(-5 * time.Minute)) {
		return domain.ErrPastDepartureTime
	}

	ride.ID = uuid.New()
	ride.Status = domain.RideStatusDraft // Changed from active to draft
	ride.AvailableSeats = ride.TotalSeats
	ride.CreatedAt = time.Now()
	ride.UpdatedAt = time.Now()

	ride.OriginAddress = sanitize.String(ride.OriginAddress)
	ride.DestinationAddress = sanitize.String(ride.DestinationAddress)

	for i := range ride.Stops {
		ride.Stops[i].ID = uuid.New()
		ride.Stops[i].RideID = ride.ID
		ride.Stops[i].Address = sanitize.String(ride.Stops[i].Address)
	}

	distanceKm := haversine(ride.OriginLat, ride.OriginLon, ride.DestinationLat, ride.DestinationLon)
	recommendedPrice := distanceKm * 500.0 // 500 UZS per KM

	if ride.PricePerSeat <= recommendedPrice*1.15 {
		ride.HasFairPriceBadge = true
	} else {
		ride.HasFairPriceBadge = false
	}

	if u.grpc != nil {
		vehicleRes, err := u.grpc.GetDriverVehicle(ctx, ride.DriverID)
		if err == nil && vehicleRes != nil {
			vid, _ := uuid.Parse(vehicleRes.Id)
			ride.VehicleData = &domain.Vehicle{
				ID:               vid,
				Brand:            vehicleRes.Brand,
				Model:            vehicleRes.Model,
				Color:            vehicleRes.Color,
				LicensePlate:     vehicleRes.LicensePlate,
				ModerationStatus: vehicleRes.ModerationStatus,
			}
			if vehicleRes.Year > 0 {
				ride.VehicleData.Year = vehicleRes.Year
			}
		}
	}

	// 1. Сохраняем поездку в PostgreSQL (в статусе draft)
	err := u.repo.CreateRide(ctx, ride)
	if err != nil {
		if strings.Contains(err.Error(), "foreign key constraint") {
			return fmt.Errorf("%w: %s", domain.ErrDriverNotFound, err.Error())
		}
		return err
	}

	// Не добавляем в Redis GEO пока поездка не переведена в active
	if u.cache != nil {
		_ = u.cache.SetRideDriver(ctx, ride.ID, ride.DriverID)
	}

	return nil
}

func (u *rideUseCase) ActivateRide(ctx context.Context, driverID, rideID uuid.UUID) error {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("failed to get ride: %w", err)
	}
	if ride.DriverID != driverID {
		return errors.New("unauthorized: you are not the driver of this ride")
	}

	if ride.Status != domain.RideStatusDraft {
		return errors.New("ride is not in draft status")
	}

	if u.grpc != nil {
		// 1. RPC check for approved car
		isApproved, err := u.grpc.CheckCarStatus(ctx, ride.DriverID)
		if err != nil {
			return fmt.Errorf("car check failed: %w", err)
		}
		if !isApproved {
			return errors.New("cannot activate ride: driver does not have an approved car")
		}

		// 2. RPC check for billing limits
		isValid, _, err := u.grpc.CheckRideLimit(ctx, ride.DriverID)
		if err != nil {
			return fmt.Errorf("billing check failed: %w", err)
		}
		if !isValid {
			// Driver has no active paid subscription — check if they use premium filters
			hasCustomSpeed := ride.DrivingSpeed != "" && ride.DrivingSpeed != "normal" && ride.DrivingSpeed != "standard"
			if ride.HasAc || ride.NoSmoking || ride.MaxTwoBack || hasCustomSpeed {
				return errors.New("filters (A/C, no smoking, 2 in back seat, driving speed) are only available on Pro, Plus and Premium plans")
			}
			return domain.ErrInsufficientRideLimits
		}
		// isValid=true means driver has an active paid subscription (plus/pro/premium) — filters allowed
	}

	// 3. Update status to active
	err = u.repo.UpdateRideStatus(ctx, rideID, domain.RideStatusActive)
	if err != nil {
		return fmt.Errorf("failed to update ride status: %w", err)
	}

	// 4. Add to Redis GEO for search and invalidate search cache
	if u.cache != nil {
		_ = u.cache.SetRideLocation(ctx, ride.ID, ride.OriginLon, ride.OriginLat)
		_ = u.cache.InvalidateSearchResults(ctx)
	}

	// 5. Publish ride.created event
	if u.publisher != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"ride_id":   rideID.String(),
			"driver_id": ride.DriverID.String(),
		})
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.created", payload); err != nil {
			slog.Error("Failed to publish", "error", err)
		}
	}

	// 6. Find matching route subscriptions
	if u.subRepo != nil {
		subs, err := u.subRepo.FindMatchingSubscriptions(ctx, ride.OriginLat, ride.OriginLon, ride.DestinationLat, ride.DestinationLon, ride.DepartureTime, 15.0)
		if err == nil {
			for _, sub := range subs {
				if u.publisher != nil {
					matchPayload, _ := json.Marshal(map[string]interface{}{
						"ride_id":      rideID.String(),
						"passenger_id": sub.PassengerID.String(),
						"driver_id":    ride.DriverID.String(),
					})
					if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.route_match", matchPayload); err != nil {
						slog.Error("Failed to publish", "error", err)
					}
				}
			}
		} else {
			slog.Error("Failed to find matching route subscriptions", "error", err)
		}
	}

	return nil
}

func (u *rideUseCase) StartRide(ctx context.Context, driverID, rideID uuid.UUID) error {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("failed to get ride: %w", err)
	}
	if ride.DriverID != driverID {
		return errors.New("unauthorized: you are not the driver of this ride")
	}

	if ride.Status != domain.RideStatusActive {
		return errors.New("ride is not in active status")
	}

	// Update status to in_progress
	err = u.repo.UpdateRideStatus(ctx, rideID, domain.RideStatusInProgress)
	if err != nil {
		return fmt.Errorf("failed to update ride status: %w", err)
	}

	// We could also remove it from Redis GEO so it's no longer searchable
	// e.g. u.cache.RemoveRideLocation(ctx, ride.ID) if such method exists.

	u.publishStatusChanged(ctx, rideID, string(domain.RideStatusInProgress))

	return nil
}

func (u *rideUseCase) publishStatusChanged(ctx context.Context, rideID uuid.UUID, status string) {
	if u.publisher == nil {
		return
	}
	bookings, err := u.repo.GetBookingsByRideID(ctx, rideID)
	if err != nil {
		slog.Error("failed to get bookings for ride status change", "error", err)
		return
	}
	var passengerIDs []string
	for _, b := range bookings {
		if b.Status == domain.BookingStatusConfirmed || b.Status == "accepted" {
			passengerIDs = append(passengerIDs, b.PassengerID.String())
		}
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"ride_id":       rideID.String(),
		"new_status":    status,
		"passenger_ids": passengerIDs,
	})
	if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.status_changed", payload); err != nil {
		slog.Error("Failed to publish", "error", err)
	}
}

func (u *rideUseCase) FindRides(ctx context.Context, params domain.RideSearchParams) ([]domain.Ride, error) {
	if params.MinSeats <= 0 {
		params.MinSeats = 1
	}

	// 0. Генерируем хэш параметров поиска для кэша
	// Округляем координаты до 2 знаков (~1.1 км точность), чтобы увеличить Cache Hit
	roundedOriginLat := math.Round(params.OriginLat*100) / 100
	roundedOriginLon := math.Round(params.OriginLon*100) / 100
	roundedDestLat := math.Round(params.DestLat*100) / 100
	roundedDestLon := math.Round(params.DestLon*100) / 100

	hashParams := fmt.Sprintf("%.2f:%.2f:%.2f:%.2f:%f:%d:%s", roundedOriginLat, roundedOriginLon, roundedDestLat, roundedDestLon, params.RadiusMeters, params.MinSeats, params.Gender)
	if params.PassengerTier != "" {
		hashParams += fmt.Sprintf(":tier%s", params.PassengerTier)
	}
	if params.PriceMin != nil {
		hashParams += fmt.Sprintf(":pmin%d", *params.PriceMin)
	}
	if params.PriceMax != nil {
		hashParams += fmt.Sprintf(":pmax%d", *params.PriceMax)
	}
	if params.DateFrom != nil {
		hashParams += fmt.Sprintf(":dfrom%d", params.DateFrom.Unix())
	}
	if params.DateTo != nil {
		hashParams += fmt.Sprintf(":dto%d", params.DateTo.Unix())
	}
	if params.City != "" {
		hashParams += fmt.Sprintf(":city%s", params.City)
	}
	if params.IncludeTest {
		hashParams += ":include_test=true"
	} else {
		hashParams += ":include_test=false"
	}

	hasher := sha256.New()
	hasher.Write([]byte(hashParams))
	cacheKey := hex.EncodeToString(hasher.Sum(nil))

	// Проверяем кэш
	if u.cache != nil {
		if cachedRides, err := u.cache.GetSearchResults(ctx, cacheKey); err == nil && len(cachedRides) > 0 {
			slog.Debug("returning search results from cache", "hash", cacheKey)
			return cachedRides, nil
		}
	}

	// 1. Ищем поездки в PostgreSQL через сложный CTE-запрос с учетом направления и остановок
	rides, err := u.repo.SearchRides(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rides from db: %w", err)
	}

	if len(rides) == 0 {
		return []domain.Ride{}, nil
	}

	// 2. Добавим профили водителей из кэша/gRPC и применим фильтры
	var finalRides []domain.Ride
	for _, r := range rides {
		r.Driver = u.resolveDriverProfile(ctx, r.DriverID)

		// Применяем строгий фильтр по полу водителя (если пассажир запросил конкретный пол)
		if params.Gender != "" && params.Gender != "any" {
			if r.Driver != nil && r.Driver.Gender != params.Gender {
				continue
			}
		}

		finalRides = append(finalRides, r)
	}

	// 3. Сохраняем полностью сформированный результат (с профилями водителей) в кэш на 2 минуты
	if u.cache != nil && len(finalRides) > 0 {
		_ = u.cache.SetSearchResults(ctx, cacheKey, finalRides, 2*time.Minute)
	}

	return finalRides, nil
}

func (u *rideUseCase) BookRide(ctx context.Context, rideID, passengerID uuid.UUID, seats int32, isFullCarBuyout bool, offeredPrice *float64, pickupPreference, dropoffType string) (*domain.Booking, error) {
	if seats <= 0 && !isFullCarBuyout {
		return nil, errors.New("must book at least 1 seat")
	}

	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride: %w", err)
	}
	if ride == nil {
		return nil, errors.New("ride not found")
	}
	if ride.Status != domain.RideStatusActive {
		return nil, domain.ErrRideNotActive
	}
	if !ride.DepartureTime.IsZero() && time.Now().After(ride.DepartureTime.Add(3*time.Hour)) {
		return nil, domain.ErrRideDeparted
	}
	if ride.DriverID == passengerID {
		return nil, errors.New("cannot book your own ride")
	}

	if !isFullCarBuyout && seats > ride.TotalSeats {
		return nil, errors.New("cannot book more seats than total vehicle capacity")
	}
	if !isFullCarBuyout && seats > ride.AvailableSeats {
		return nil, errors.New("not enough available seats")
	}

	if ride.GenderRestriction != "" && ride.GenderRestriction != "any" {
		if u.grpc != nil {
			profile, err := u.grpc.GetDriverProfile(ctx, passengerID)
			if err == nil && profile != nil && profile.Gender != "" {
				if !strings.EqualFold(profile.Gender, ride.GenderRestriction) {
					return nil, fmt.Errorf("ride is restricted to %s passengers only", ride.GenderRestriction)
				}
			}
		}
	}

	booking, err := u.repo.CreateBookingTx(ctx, rideID, passengerID, seats, isFullCarBuyout, offeredPrice, pickupPreference, dropoffType)
	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	// Обогащаем booking информацией о поездке (чтобы вернуть driver_id клиенту)
	booking.Ride = ride

	// Отправляем событие о создании бронирования водителю
	if u.publisher != nil {
		bookingEvent, _ := json.Marshal(map[string]interface{}{
			"booking_id":   booking.ID.String(),
			"ride_id":      rideID.String(),
			"driver_id":    ride.DriverID.String(),
			"passenger_id": passengerID.String(),
			"seats_booked": seats,
		})
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.booking.created", bookingEvent); err != nil {
			slog.Error("Failed to publish ride.booking.created", "error", err)
		}
	}

	// Если есть торг, дополнительно отправляем уведомление о торге
	if offeredPrice != nil && u.publisher != nil {
		event := map[string]interface{}{
			"ride_id":       rideID.String(),
			"driver_id":     ride.DriverID.String(),
			"passenger_id":  passengerID.String(),
			"offered_price": *offeredPrice,
		}
		body, _ := json.Marshal(event)
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.negotiation_requested", body); err != nil {
			slog.Error("Failed to publish ride.negotiation_requested", "error", err)
		}
	}

	return booking, nil
}

func (u *rideUseCase) AcceptBooking(ctx context.Context, driverID, rideID, bookingID uuid.UUID) error {
	// First, check if the ride belongs to this driver
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("failed to get ride: %w", err)
	}
	if ride.DriverID != driverID {
		return errors.New("unauthorized: you are not the driver of this ride")
	}

	booking, err := u.repo.AcceptBookingTx(ctx, rideID, bookingID)
	if err != nil {
		return err
	}

	if u.cache != nil {
		_ = u.cache.AddRidePassenger(ctx, rideID, booking.PassengerID)
	}

	if u.publisher != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"ride_id":      rideID.String(),
			"driver_id":    driverID.String(),
			"passenger_id": booking.PassengerID.String(),
			"booking_id":   bookingID.String(),
		})
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.booking.confirmed", payload); err != nil {
			slog.Error("Failed to publish ride.booking.confirmed", "error", err)
		}
	}

	return nil
}

func (u *rideUseCase) RejectBooking(ctx context.Context, driverID, rideID, bookingID uuid.UUID) error {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("failed to get ride: %w", err)
	}
	if ride.DriverID != driverID {
		return errors.New("unauthorized: you are not the driver of this ride")
	}

	booking, err := u.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}

	if err := u.repo.RejectBooking(ctx, bookingID); err != nil {
		return err
	}

	if u.publisher != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"ride_id":      rideID.String(),
			"driver_id":    driverID.String(),
			"passenger_id": booking.PassengerID.String(),
			"booking_id":   bookingID.String(),
		})
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.booking.rejected", payload); err != nil {
			slog.Error("Failed to publish ride.booking.rejected", "error", err)
		}
	}

	return nil
}

func (u *rideUseCase) CounterBooking(ctx context.Context, driverID, rideID, bookingID uuid.UUID, newPrice float64) error {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("failed to get ride: %w", err)
	}
	if ride.DriverID != driverID {
		return errors.New("unauthorized: you are not the driver of this ride")
	}

	booking, err := u.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}

	if booking.RideID != rideID {
		return errors.New("booking does not belong to this ride")
	}

	if booking.Status != domain.BookingStatusPending && booking.Status != domain.BookingStatusPendingNegotiation {
		return domain.ErrBookingAlreadyAccepted
	}

	err = u.repo.UpdateBookingPriceAndStatus(ctx, bookingID, newPrice, string(domain.BookingStatusPendingNegotiation))
	if err != nil {
		return fmt.Errorf("failed to update booking price and status: %w", err)
	}

	if u.publisher != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"booking_id": bookingID.String(),
			"ride_id":    rideID.String(),
			"new_price":  newPrice,
		})
		_ = u.publisher.Publish(ctx, "yolyolakay.events", "booking.countered", payload)
	}

	return nil
}

func (u *rideUseCase) TimeoutPendingNegotiations(ctx context.Context) error {
	rejectedBookings, err := u.repo.RejectExpiredNegotiations(ctx)
	if err != nil {
		return fmt.Errorf("failed to reject expired negotiations: %w", err)
	}

	if u.publisher != nil {
		for _, b := range rejectedBookings {
			title := "Предложение отклонено"
			message := "Водитель не ответил на ваше предложение цены в течение 30 минут. Бронирование отклонено."
			_ = u.publisher.PublishNotification(ctx, b.PassengerID, title, message, "booking.rejected")
		}
	}

	return nil
}

func (u *rideUseCase) CancelRide(ctx context.Context, driverID, rideID uuid.UUID, reason string) error {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("failed to get ride: %w", err)
	}

	if ride.DriverID != driverID {
		return errors.New("unauthorized: you are not the driver of this ride")
	}

	previousStatus := ride.Status
	wasActivated := previousStatus == domain.RideStatusActive || previousStatus == domain.RideStatusInProgress

	err = u.repo.UpdateRideStatus(ctx, rideID, domain.RideStatusCancelled)
	if err != nil {
		return fmt.Errorf("failed to update ride status: %w", err)
	}

	// Каскадная отмена бронирований и сбор всех затронутых пассажиров
	passengerMap := make(map[string]bool)
	var passengerIDs []string

	bookings, err := u.repo.GetBookingsByRideID(ctx, rideID)
	if err == nil {
		for _, b := range bookings {
			if b.Status == domain.BookingStatusPending || b.Status == domain.BookingStatusConfirmed || b.Status == "accepted" || b.Status == domain.BookingStatusPendingNegotiation {
				_ = u.repo.UpdateBookingStatus(ctx, b.ID, string(domain.BookingStatusCanceled))
				
				pIDStr := b.PassengerID.String()
				if !passengerMap[pIDStr] {
					passengerMap[pIDStr] = true
					passengerIDs = append(passengerIDs, pIDStr)
				}
			}
		}
	}

	if u.publisher != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"ride_id":       rideID.String(),
			"driver_id":     driverID.String(),
			"reason":        reason,
			"passenger_ids": passengerIDs,
			"was_activated": wasActivated,
		})
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.cancelled", payload); err != nil {
			slog.Error("Failed to publish ride.cancelled", "error", err)
		}
		u.publishStatusChanged(ctx, rideID, string(domain.RideStatusCancelled))
	}

	if u.cache != nil {
		_ = u.cache.InvalidateSearchResults(ctx)
	}

	return nil
}

func (u *rideUseCase) CancelBookingByPassenger(ctx context.Context, passengerID, bookingID uuid.UUID, reason string) error {
	booking, err := u.repo.CancelBookingTx(ctx, passengerID, bookingID)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	ride, err := u.repo.GetRideByID(ctx, booking.RideID)
	if err == nil && ride != nil && u.publisher != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"ride_id":      booking.RideID.String(),
			"driver_id":    ride.DriverID.String(),
			"passenger_id": passengerID.String(),
			"booking_id":   bookingID.String(),
			"reason":       reason,
		})
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.booking.cancelled", payload); err != nil {
			slog.Error("Failed to publish ride.booking.cancelled", "error", err)
		}

		u.publishStatusChanged(ctx, booking.RideID, "booking_cancelled")
	}

	return nil
}

func (u *rideUseCase) CompleteRide(ctx context.Context, driverID, rideID uuid.UUID) error {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("failed to get ride: %w", err)
	}

	if ride.DriverID != driverID {
		return errors.New("unauthorized: you are not the driver of this ride")
	}

	err = u.repo.UpdateRideStatus(ctx, rideID, domain.RideStatusCompleted)
	if err != nil {
		return fmt.Errorf("failed to update ride status: %w", err)
	}

	if u.publisher != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"ride_id":   rideID.String(),
			"driver_id": driverID.String(),
		})
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "ride.completed", payload); err != nil {
			slog.Error("Failed to publish", "error", err)
		}
		u.publishStatusChanged(ctx, rideID, string(domain.RideStatusCompleted))
	}

	return nil
}

func (u *rideUseCase) ReportPassengerNoShow(ctx context.Context, driverID, rideID, bookingID uuid.UUID) error {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("failed to get ride: %w", err)
	}

	if ride.DriverID != driverID {
		return errors.New("unauthorized: you are not the driver of this ride")
	}

	booking, err := u.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}

	if booking.RideID != rideID {
		return errors.New("booking does not belong to this ride")
	}

	if string(booking.Status) != string(domain.BookingStatusConfirmed) && string(booking.Status) != "accepted" {
		return errors.New("only confirmed bookings can be marked as no-show")
	}

	err = u.repo.ReportPassengerNoShowTx(ctx, rideID, bookingID)
	if err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	// Check if passenger should be blocked (3+ no-shows in last 30 days)
	since := time.Now().AddDate(0, 0, -30)
	count, err := u.repo.CountPassengerNoShows(ctx, booking.PassengerID, since)
	if err != nil {
		slog.Error("failed to count passenger no shows", "error", err)
	} else if count >= 3 && u.publisher != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"user_id": booking.PassengerID.String(),
			"reason":  "3_no_show",
		})
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "user.block_requested", payload); err != nil {
			slog.Error("Failed to publish", "error", err)
		}
	}

	return nil
}

func (u *rideUseCase) HandleLimitExceeded(ctx context.Context, rideID uuid.UUID, bookingID *uuid.UUID, userID uuid.UUID) error {
	if bookingID != nil {
		if err := u.repo.UpdateBookingStatus(ctx, *bookingID, string(domain.BookingStatusCanceled)); err != nil {
			return fmt.Errorf("failed to cancel booking: %w", err)
		}
	} else {
		if err := u.repo.UpdateRideStatus(ctx, rideID, domain.RideStatusDraft); err != nil {
			return fmt.Errorf("failed to revert ride to draft: %w", err)
		}
	}

	if u.publisher != nil {
		_ = u.publisher.PublishNotification(ctx, userID, "Limit tugadi", "Safar paketingiz tugadi. Tarifni yangilang", "limit_exceeded")
	}

	return nil
}

func (u *rideUseCase) GetMyRidesAsDriver(ctx context.Context, driverID uuid.UUID) ([]domain.Ride, error) {
	rides, err := u.repo.GetRidesByDriverID(ctx, driverID)
	if err != nil {
		return nil, err
	}

	for i := range rides {
		bookings, err := u.repo.GetBookingsByRideID(ctx, rides[i].ID)
		if err == nil && len(bookings) > 0 {
			for j := range bookings {
				passengerProfile, err := u.grpc.GetPassengerProfile(ctx, bookings[j].PassengerID)
				if err == nil && passengerProfile != nil {
					// Mask PII if the ride is in the past (completed or cancelled)
					if rides[i].Status == domain.RideStatusCompleted || rides[i].Status == domain.RideStatusCancelled {
						if len(passengerProfile.PhoneNumber) > 6 {
							// Mask middle digits, e.g. +99890***2233
							phone := passengerProfile.PhoneNumber
							passengerProfile.PhoneNumber = phone[:6] + "***" + phone[len(phone)-4:]
						} else {
							passengerProfile.PhoneNumber = "***"
						}
						// Can also mask the name partially if needed, but usually phone is the main PII.
					}
					bookings[j].Passenger = passengerProfile
				}
			}
			rides[i].Bookings = bookings
		}
	}

	return rides, nil
}

func (u *rideUseCase) GetMyBookingsAsPassenger(ctx context.Context, passengerID uuid.UUID) ([]domain.Booking, error) {
	bookings, err := u.repo.GetBookingsByPassengerID(ctx, passengerID)
	if err != nil {
		return nil, err
	}

	if len(bookings) == 0 {
		return bookings, nil
	}

	rideIDs := make([]uuid.UUID, 0, len(bookings))
	rideIDMap := make(map[uuid.UUID]bool)
	for _, b := range bookings {
		if !rideIDMap[b.RideID] {
			rideIDMap[b.RideID] = true
			rideIDs = append(rideIDs, b.RideID)
		}
	}

	rides, err := u.repo.GetRidesByIDs(ctx, rideIDs)
	if err != nil {
		return nil, err
	}

	ridesMap := make(map[uuid.UUID]*domain.Ride)
	for i := range rides {
		ridesMap[rides[i].ID] = &rides[i]
	}

	for i := range bookings {
		if r, ok := ridesMap[bookings[i].RideID]; ok {
			driverProfile := u.resolveDriverProfile(ctx, r.DriverID)
			if bookings[i].Status != domain.BookingStatusConfirmed && bookings[i].Status != "accepted" {
				dpCopy := *driverProfile
				dpCopy.Phone = ""
				driverProfile = &dpCopy
			}
			rCopy := *r
			rCopy.Driver = driverProfile
			bookings[i].Ride = &rCopy
		}
	}

	return bookings, nil
}

func (u *rideUseCase) GetBookingByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	return u.repo.GetBookingByID(ctx, id)
}

func (u *rideUseCase) GetDriverStatistics(ctx context.Context, driverID uuid.UUID) (*domain.DriverStatistics, error) {
	return u.repo.GetDriverStatisticsAggregated(ctx, driverID)
}

func (u *rideUseCase) GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserRideStats, error) {
	return u.repo.GetUserStats(ctx, userID)
}

func (u *rideUseCase) GetUserActiveRide(ctx context.Context, userID uuid.UUID) (*domain.Ride, error) {
	// Check as driver
	driverRides, err := u.repo.GetRidesByDriverID(ctx, userID)
	if err == nil {
		for _, r := range driverRides {
			if r.Status == domain.RideStatusActive || r.Status == domain.RideStatusInProgress {
				return &r, nil
			}
		}
	}

	// Check as passenger
	passengerBookings, err := u.repo.GetBookingsByPassengerID(ctx, userID)
	if err == nil {
		for _, b := range passengerBookings {
			if b.Status == domain.BookingStatusConfirmed {
				// Need to check the ride status
				ride, err := u.repo.GetRideByID(ctx, b.RideID)
				if err == nil && ride != nil {
					if ride.Status == domain.RideStatusActive || ride.Status == domain.RideStatusInProgress {
						return ride, nil
					}
				}
			}
		}
	}

	return nil, nil // No active ride found
}

func (u *rideUseCase) resolveDriverProfile(ctx context.Context, driverID uuid.UUID) *domain.DriverProfile {
	var profile *domain.DriverProfile
	var err error

	if u.cache != nil {
		profile, err = u.cache.GetDriverProfile(ctx, driverID)
	} else {
		err = errors.New("cache is disabled")
	}

	if err == nil && profile != nil {
		profile.ID = driverID
		return profile
	}

	// Fallback to User Service via gRPC
	if u.grpc != nil {
		grpcProfile, grpcErr := u.grpc.GetDriverProfile(ctx, driverID)
		if grpcErr == nil && grpcProfile != nil {
			grpcProfile.ID = driverID
			if u.cache != nil {
				_ = u.cache.SetDriverProfile(ctx, driverID, grpcProfile)
			}
			return grpcProfile
		}
	}

	// Fallback default
	return &domain.DriverProfile{
		ID:     driverID,
		Name:   "Водитель",
		Rating: 5.0,
	}
}

func (u *rideUseCase) resolvePassengerProfile(ctx context.Context, passengerID uuid.UUID) *domain.PassengerProfile {
	// Fallback to User Service via gRPC
	if u.grpc != nil {
		grpcProfile, grpcErr := u.grpc.GetDriverProfile(ctx, passengerID)
		if grpcErr == nil && grpcProfile != nil {
			return &domain.PassengerProfile{
				ID:          passengerID,
				Name:        grpcProfile.Name,
				Avatar:      grpcProfile.Avatar,
				Rating:      grpcProfile.Rating,
				PhoneNumber: grpcProfile.Phone,
				Gender:      grpcProfile.Gender,
				KycStatus:   grpcProfile.KycStatus,
				CreatedAt:   grpcProfile.CreatedAt,
			}
		}
	}

	return &domain.PassengerProfile{
		ID:     passengerID,
		Name:   "Yo'lovchi",
		Rating: 5.0,
	}
}

func (u *rideUseCase) GetRideByID(ctx context.Context, rideID uuid.UUID) (*domain.Ride, error) {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride: %w", err)
	}
	ride.Driver = u.resolveDriverProfile(ctx, ride.DriverID)
	if (ride.VehicleData == nil || ride.VehicleData.Color == "" || ride.VehicleData.Year == 0) && u.grpc != nil {
		vehicleRes, err := u.grpc.GetDriverVehicle(ctx, ride.DriverID)
		if err == nil && vehicleRes != nil {
			vid, _ := uuid.Parse(vehicleRes.Id)
			if ride.VehicleData == nil {
				ride.VehicleData = &domain.Vehicle{}
			}
			if ride.VehicleData.ID == uuid.Nil {
				ride.VehicleData.ID = vid
			}
			if ride.VehicleData.Brand == "" {
				ride.VehicleData.Brand = vehicleRes.Brand
			}
			if ride.VehicleData.Model == "" {
				ride.VehicleData.Model = vehicleRes.Model
			}
			if ride.VehicleData.Color == "" {
				ride.VehicleData.Color = vehicleRes.Color
			}
			if ride.VehicleData.LicensePlate == "" {
				ride.VehicleData.LicensePlate = vehicleRes.LicensePlate
			}
			if ride.VehicleData.ModerationStatus == "" {
				ride.VehicleData.ModerationStatus = vehicleRes.ModerationStatus
			}
			if ride.VehicleData.Year == 0 && vehicleRes.Year > 0 {
				ride.VehicleData.Year = vehicleRes.Year
			}
		}
	}

	bookings, err := u.repo.GetBookingsByRideID(ctx, rideID)
	if err == nil {
		for i := range bookings {
			if bookings[i].Status == domain.BookingStatusConfirmed || bookings[i].Status == "accepted" {
				bookings[i].Passenger = u.resolvePassengerProfile(ctx, bookings[i].PassengerID)
			}
		}
		ride.Bookings = bookings
	}

	return ride, nil
}

func (u *rideUseCase) CreateRouteSubscription(ctx context.Context, sub *domain.RouteSubscription) error {
	sub.ID = uuid.New()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	sub.IsActive = true
	return u.subRepo.CreateSubscription(ctx, sub)
}

func (u *rideUseCase) GetRouteSubscriptions(ctx context.Context, userID uuid.UUID) ([]domain.RouteSubscription, error) {
	return u.subRepo.GetSubscriptionsByUserID(ctx, userID)
}

func (u *rideUseCase) DeleteRouteSubscription(ctx context.Context, subID, userID uuid.UUID) error {
	return u.subRepo.DeleteSubscription(ctx, subID, userID)
}

func (u *rideUseCase) GetRideRecommendations(ctx context.Context, userID uuid.UUID) (*domain.RideRecommendations, error) {
	recs := &domain.RideRecommendations{
		HistoryBased: []domain.RouteStats{},
		TopGlobal:    []domain.RouteStats{},
		Unvisited:    []domain.CityStats{},
	}

	// 1. User's most visited destinations (max 3)
	historyStats, err := u.repo.GetUserMostVisitedDestinations(ctx, userID, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to get history destinations: %w", err)
	}
	if historyStats != nil {
		recs.HistoryBased = historyStats
	}

	// 2. Global top destinations (max 10)
	globalStats, err := u.repo.GetTopGlobalDestinations(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get global destinations: %w", err)
	}

	// 3. Find unvisited from global (max 3)
	historyMap := make(map[string]bool)
	for _, st := range historyStats {
		historyMap[st.To] = true
	}

	var unvisitedDests []domain.RouteStats
	var top3Global []domain.RouteStats

	for _, st := range globalStats {
		if len(top3Global) < 3 {
			top3Global = append(top3Global, st)
		}
		if !historyMap[st.To] && len(unvisitedDests) < 3 {
			unvisitedDests = append(unvisitedDests, st)
		}
	}
	
	if top3Global != nil {
		recs.TopGlobal = top3Global
	}

	// Create CityStats for unvisited
	for _, unv := range unvisitedDests {
		// Try to find an active ride to this destination to get coordinates
		// In a real scenario, you'd use a geocoder, but since it's a known destination, 
		// we can query an active ride going there.
		rides, _ := u.repo.GetActiveRidesToDestinations(ctx, []string{unv.To}, 1)
		var lat, lon float64
		if len(rides) > 0 {
			lat = rides[0].DestinationLat
			lon = rides[0].DestinationLon
		}
		recs.Unvisited = append(recs.Unvisited, domain.CityStats{
			City: unv.To,
			Lat:  lat,
			Lon:  lon,
		})
	}

	return recs, nil
}

func (u *rideUseCase) GetAdminActiveRides(ctx context.Context) ([]domain.Ride, error) {
	return u.repo.GetAdminActiveRides(ctx)
}

func (u *rideUseCase) GetRideDriverLocation(ctx context.Context, rideID, callerID uuid.UUID) (*domain.DriverLocation, error) {
	// 1. Fetch ride
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride: %w", err)
	}
	if ride == nil {
		return nil, errors.New("ride not found")
	}

	// 2. RBAC check: caller must be the driver OR a confirmed passenger
	isAuthorized := ride.DriverID == callerID
	if !isAuthorized {
		bookings, err := u.repo.GetBookingsByRideID(ctx, rideID)
		if err == nil {
			for _, b := range bookings {
				if b.PassengerID == callerID && (b.Status == domain.BookingStatusConfirmed || b.Status == "accepted") {
					isAuthorized = true
					break
				}
			}
		}
	}

	if !isAuthorized {
		return nil, errors.New("unauthorized: you must be the driver or a confirmed passenger of this ride")
	}

	// 3. Check ride status (must not be draft or cancelled)
	if ride.Status == domain.RideStatusDraft || ride.Status == domain.RideStatusCancelled {
		return nil, errors.New("ride is not active or in progress")
	}

	// 4. Retrieve live location from cache
	if u.cache == nil {
		return nil, errors.New("cache service unavailable")
	}

	loc, err := u.cache.GetDriverLiveLocation(ctx, ride.DriverID)
	if err != nil || loc == nil {
		return nil, errors.New("driver location not available or offline")
	}

	// 5. Check if location timestamp is not excessively stale (> 30 minutes)
	if time.Since(loc.UpdatedAt) > 30*time.Minute {
		return nil, errors.New("driver location is outdated or driver is offline")
	}

	return loc, nil
}

func (u *rideUseCase) CreateWatchToken(ctx context.Context, rideID, callerID uuid.UUID) (*domain.WatchTokenResponse, error) {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride: %w", err)
	}
	if ride == nil {
		return nil, errors.New("ride not found")
	}

	isAuthorized := ride.DriverID == callerID
	if !isAuthorized {
		bookings, err := u.repo.GetBookingsByRideID(ctx, rideID)
		if err == nil {
			for _, b := range bookings {
				if b.PassengerID == callerID && (b.Status == domain.BookingStatusConfirmed || b.Status == "accepted") {
					isAuthorized = true
					break
				}
			}
		}
	}

	if !isAuthorized {
		return nil, errors.New("unauthorized: you must be the driver or a confirmed passenger of this ride")
	}

	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}
	token := "wt_" + hex.EncodeToString(b)
	ttl := 24 * time.Hour
	expiresAt := time.Now().Add(ttl)

	if u.cache != nil {
		if err := u.cache.SetWatchToken(ctx, token, rideID, callerID, ttl); err != nil {
			return nil, fmt.Errorf("failed to store watch token: %w", err)
		}
	}

	return &domain.WatchTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (u *rideUseCase) RevokeWatchToken(ctx context.Context, rideID, callerID uuid.UUID, token string) error {
	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return fmt.Errorf("failed to get ride: %w", err)
	}
	if ride == nil {
		return errors.New("ride not found")
	}

	isAuthorized := ride.DriverID == callerID
	if !isAuthorized {
		bookings, err := u.repo.GetBookingsByRideID(ctx, rideID)
		if err == nil {
			for _, b := range bookings {
				if b.PassengerID == callerID && (b.Status == domain.BookingStatusConfirmed || b.Status == "accepted") {
					isAuthorized = true
					break
				}
			}
		}
	}

	if !isAuthorized {
		return errors.New("unauthorized: you must be the driver or a confirmed passenger of this ride")
	}

	if u.cache != nil {
		_ = u.cache.DeleteWatchToken(ctx, token)
	}
	return nil
}

func (u *rideUseCase) GetRideDriverLocationByWatchToken(ctx context.Context, rideID uuid.UUID, token string) (*domain.WatchLocationResponse, error) {
	if u.cache == nil {
		return nil, errors.New("cache service unavailable")
	}

	tokenRideID, exists, err := u.cache.GetWatchToken(ctx, token)
	if err != nil || !exists || tokenRideID != rideID {
		return nil, errors.New("invalid or expired watch token")
	}

	ride, err := u.repo.GetRideByID(ctx, rideID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride: %w", err)
	}
	if ride == nil {
		return nil, errors.New("ride not found")
	}

	if ride.Status == domain.RideStatusCancelled || ride.Status == domain.RideStatusCompleted {
		return nil, errors.New("ride is not active or in progress")
	}

	loc, err := u.cache.GetDriverLiveLocation(ctx, ride.DriverID)
	if err != nil || loc == nil {
		return nil, errors.New("driver location not available or offline")
	}

	if time.Since(loc.UpdatedAt) > 30*time.Minute {
		return nil, errors.New("driver location is outdated or driver is offline")
	}

	return &domain.WatchLocationResponse{
		Latitude:  loc.Latitude,
		Longitude: loc.Longitude,
		Heading:   loc.Heading,
		UpdatedAt: loc.UpdatedAt,
	}, nil
}

func (u *rideUseCase) GetDriverPublicRides(ctx context.Context, driverID uuid.UUID, limit int) ([]domain.DriverPublicRide, error) {
	return u.repo.GetPublicRidesByDriverID(ctx, driverID, limit)
}
