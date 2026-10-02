package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	domain "github.com/hextec/yolyolakay-backend/internal/ride/domain"
	"github.com/hextec/yolyolakay-backend/pkg/middleware"
)

type Handler struct {
	useCase domain.RideUseCase
}

func NewHandler(useCase domain.RideUseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) InitRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// Public
	public := router.Group("/api/v1/rides")
	{
		public.GET("/search", h.SearchRides)
		public.GET("/driver/:driver_id", h.GetDriverPublicRides)
		public.GET("/:id/location", h.GetRideLocation)
		public.GET("/:id", h.GetRideByID)
	}

	// Protected
	protected := router.Group("/api/v1/rides")
	protected.Use(authMiddleware)
	{
		protected.GET("/:id/passengers", h.GetRidePassengers)
		protected.POST("", h.PublishRide)
		protected.POST("/", h.PublishRide)
		protected.POST("/:id/book", h.BookRide)
		protected.POST("/:id/bookings/:booking_id/accept", h.AcceptBooking)
		protected.POST("/:id/bookings/:booking_id/reject", h.RejectBooking)
		protected.POST("/:id/bookings/:booking_id/counter", h.CounterBooking)
		protected.POST("/:id/bookings/:booking_id/no-show", h.ReportNoShow)
		protected.POST("/:id/bookings/:booking_id/cancel", h.CancelRideBooking)
		protected.POST("/:id/activate", h.ActivateRide)
		protected.POST("/:id/start", h.StartRide)
		protected.POST("/:id/cancel", h.CancelRide)
		protected.DELETE("/:id", h.CancelRide)
		protected.DELETE("/:id/cancel", h.CancelRide)
		protected.PUT("/:id/cancel", h.CancelRide)
		protected.PATCH("/:id/cancel", h.CancelRide)
		protected.POST("/:id/complete", h.CompleteRide)
		protected.GET("/:id/driver-location", h.GetDriverLocation)
		protected.POST("/:id/watch-token", h.CreateWatchToken)
		protected.DELETE("/:id/watch-token", h.RevokeWatchToken)
		protected.GET("/my", h.GetMyRides)
		protected.GET("/recommendations", h.GetRideRecommendations)
		protected.GET("/earnings", h.GetDriverStatistics)

		// Route subscriptions
		protected.POST("/route-subscriptions", h.CreateRouteSubscription)
		protected.GET("/route-subscriptions", h.GetRouteSubscriptions)
		protected.DELETE("/route-subscriptions/:id", h.DeleteRouteSubscription)
	}

	bookings := router.Group("/api/v1/bookings")
	bookings.Use(authMiddleware)
	{
		bookings.POST("/:id/no-show", h.ReportBookingNoShow)
		bookings.POST("/:id/cancel", h.CancelBooking)
	}

	// Internal
	internal := router.Group("/internal/v1/rides")
	internal.Use(middleware.InternalAuth())
	{
		internal.GET("/:id", h.GetRideByID)
	}

	// Admin
	admin := router.Group("/api/v1/rides/admin")
	admin.Use(authMiddleware)
	{
		admin.GET("/stats", h.GetAdminStats)
		admin.GET("/user/:id/stats", h.GetUserStats)
		admin.GET("/user/:id/active", h.GetUserActiveRide)
		admin.GET("/active", h.GetAdminActiveRides)
	}
}

// PublishRide godoc
// @Summary Publish a new ride
// @Tags rides
// @Accept json
// @Produce json
// @Param request body PublishRideRequest true "Ride details"
// @Success 201 {object} map[string]string
// @Router /api/v1/rides/ [post]
// @Security BearerAuth
func (h *Handler) PublishRide(c *gin.Context) {
	var req PublishRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driverIDStr := c.GetString("user_id")
	if driverIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	departureTime, err := time.Parse(time.RFC3339, req.DepartureTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid departure time format"})
		return
	}
	if departureTime.Before(time.Now().Add(-5 * time.Minute)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "departure time cannot be in the past"})
		return
	}

	var stops []domain.RideStop
	for i, s := range req.Stops {
		stops = append(stops, domain.RideStop{
			Lat:        s.Lat,
			Lon:        s.Lon,
			Address:    s.Address,
			OrderIndex: int32(i),
		})
	}

	var musicTastePtr *string
	if req.MusicTaste != "" {
		musicTastePtr = &req.MusicTaste
	}

	pickupLat := req.PickupLat
	if pickupLat == nil && req.PickupPointLat != nil {
		pickupLat = req.PickupPointLat
	}
	pickupLon := req.PickupLon
	if pickupLon == nil && req.PickupPointLon != nil {
		pickupLon = req.PickupPointLon
	}

	ride := &domain.Ride{
		DriverID:           driverID,
		VehicleID:          req.VehicleID,
		OriginLat:          req.OriginLat,
		OriginLon:          req.OriginLon,
		DestinationLat:     req.DestinationLat,
		DestinationLon:     req.DestinationLon,
		OriginAddress:      req.OriginAddress,
		DestinationAddress: req.DestinationAddress,
		DepartureTime:      departureTime,
		TotalSeats:         req.TotalSeats,
		AvailableSeats:     req.TotalSeats,
		PricePerSeat:       req.PricePerSeat,
		IsNegotiable:       req.IsNegotiable,
		GenderRestriction:  req.GenderRestriction,
		HasAc:              req.HasAc,
		HasLuggage:         req.HasLuggage,
		TakesParcels:       req.TakesParcels,
		NoSmoking:          req.NoSmoking,
		NoPets:             req.NoPets,
		DrivingSpeed:       req.DrivingSpeed,
		MaxTwoBack:         req.MaxTwoBack,
		PickupPoint:        req.PickupPoint,
		PickupLat:          pickupLat,
		PickupLon:          pickupLon,
		MusicTaste:         musicTastePtr,
		AllowsCityDetour:   req.AllowsCityDetour,
		CityDetourFee:      req.CityDetourFee,
		Stops:              stops,
	}

	if err := h.useCase.PublishRide(c.Request.Context(), ride); err != nil {
		if errors.Is(err, domain.ErrInvalidRideData) || errors.Is(err, domain.ErrDriverNotFound) ||
			errors.Is(err, domain.ErrRideNotEditable) || errors.Is(err, domain.ErrSeatLimitExceeded) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "ride created successfully", "ride_id": ride.ID})
}

// SearchRides godoc
// @Summary Search rides by origin and destination
// @Tags rides
// @Produce json
// @Param origin_lat query float64 true "Origin Latitude"
// @Param origin_lon query float64 true "Origin Longitude"
// @Param dest_lat query float64 true "Destination Latitude"
// @Param dest_lon query float64 true "Destination Longitude"
// @Param radius_km query float64 true "Radius in KM"
// @Param min_seats query int false "Minimum available seats"
// @Param price_min query number false "Minimum price"
// @Param price_max query number false "Maximum price"
// @Param date_from query string false "Date from (RFC3339 or flexible like '2006-01-02')"
// @Param date_to query string false "Date to (RFC3339 or flexible like '2006-01-02')"
// @Param gender_restriction query string false "Gender restriction (male, female, any)"
// @Param city query string false "Filter by city name"
// @Param pickup_preference query string false "Pickup preference (door_to_door, main_road, bus_station)"
// @Param passenger_tier query string false "Passenger tier (standard, premium)"
// @Param include_test query bool false "Include test rides (default false)"
// @Success 200 {object} SearchRidesResponse
// @Router /api/v1/rides/search [get]
func (h *Handler) SearchRides(c *gin.Context) {
	var req SearchRidesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gender := req.Gender
	if req.GenderRestriction != "" {
		gender = req.GenderRestriction
	}
	
	if gender != "" && gender != "male" && gender != "female" && gender != "any" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid gender, must be male, female or any"})
		return
	}

	parseFlexibleDate := func(d *string) *time.Time {
		if d == nil || *d == "" {
			return nil
		}
		t, err := time.Parse(time.RFC3339, *d)
		if err == nil {
			return &t
		}
		t, err = time.Parse("2006-01-02", *d)
		if err == nil {
			return &t
		}
		return nil
	}

	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 10.0
	}

	params := domain.RideSearchParams{
		OriginLat:        req.OriginLat,
		OriginLon:        req.OriginLon,
		DestLat:          req.DestLat,
		DestLon:          req.DestLon,
		RadiusMeters:     radiusKm * 1000,
		MinSeats:         req.MinSeats,
		Gender:           gender,
		PriceMin:         req.PriceMin,
		PriceMax:         req.PriceMax,
		DateFrom:         parseFlexibleDate(req.DateFrom),
		DateTo:           parseFlexibleDate(req.DateTo),
		City:             req.City,
		PickupPreference: req.PickupPreference,
		PassengerTier:    req.PassengerTier,
		IncludeTest:      req.IncludeTest,
		Limit:            50, // default limit
		Offset:           0,
	}

	rides, err := h.useCase.FindRides(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRideData) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	response := make([]RideSearchResponse, 0)
	for _, r := range rides {
		var stops []RideStop
		for _, s := range r.Stops {
			stops = append(stops, RideStop{
				ID:         s.ID,
				RideID:     s.RideID,
				Lat:        s.Lat,
				Lon:        s.Lon,
				Address:    s.Address,
				OrderIndex: s.OrderIndex,
			})
		}
		var driver *DriverProfile
		if r.Driver != nil {
			driver = &DriverProfile{
				ID:          r.Driver.ID,
				Name:        r.Driver.Name,
				Phone:       "", // Privacy: Never leak driver phone number in public search results
				Rating:      r.Driver.Rating,
				RatingCount: r.Driver.RatingCount,
				Avatar:      r.Driver.Avatar,
				Gender:      r.Driver.Gender,
				KycStatus:   r.Driver.KycStatus,
				CreatedAt:   r.Driver.CreatedAt,
			}
		}
		var car *Vehicle
		if r.VehicleData != nil {
			car = &Vehicle{
				ID:               r.VehicleData.ID,
				Brand:            r.VehicleData.Brand,
				Model:            r.VehicleData.Model,
				Color:            r.VehicleData.Color,
				LicensePlate:     r.VehicleData.LicensePlate,
				Year:             r.VehicleData.Year,
				ModerationStatus: r.VehicleData.ModerationStatus,
			}
		}
		var musicTaste string
		if r.MusicTaste != nil {
			musicTaste = *r.MusicTaste
		}
		response = append(response, RideSearchResponse{
			ID:                 r.ID,
			OriginLat:          r.OriginLat,
			OriginLon:          r.OriginLon,
			DestinationLat:     r.DestinationLat,
			DestinationLon:     r.DestinationLon,
			OriginAddress:      r.OriginAddress,
			DestinationAddress: r.DestinationAddress,
			DepartureTime:      r.DepartureTime.Format(time.RFC3339),
			TotalSeats:         r.TotalSeats,
			AvailableSeats:     r.AvailableSeats,
			PricePerSeat:       r.PricePerSeat,
			Status:             string(r.Status),
			IsNegotiable:       r.IsNegotiable,
			HasFairPriceBadge:  r.HasFairPriceBadge,
			GenderRestriction:  r.GenderRestriction,
			HasAc:              r.HasAc,
			HasLuggage:         r.HasLuggage,
			TakesParcels:       r.TakesParcels,
			NoSmoking:          r.NoSmoking,
			NoPets:             r.NoPets,
			DrivingSpeed:       r.DrivingSpeed,
			MaxTwoBack:         r.MaxTwoBack,
			PickupPoint:        r.PickupPoint,
			PickupLat:          r.PickupLat,
			PickupLon:          r.PickupLon,
			MusicTaste:         musicTaste,
			AllowsCityDetour:   r.AllowsCityDetour,
			CityDetourFee:      r.CityDetourFee,
			Stops:              stops,
			Driver:             driver,
			Car:                car,
		})
	}

	c.JSON(http.StatusOK, gin.H{"rides": response})
}

// GetRideRecommendations godoc
// @Summary Get personalized ride recommendations
// @Description Returns 3 lists of rides: based on user history, global top destinations, and unvisited destinations
// @Tags rides
// @Accept json
// @Produce json
// @Success 200 {object} RecommendationsResponse
// @Router /api/v1/rides/recommendations [get]
// @Security BearerAuth
func (h *Handler) GetRideRecommendations(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	recs, err := h.useCase.GetRideRecommendations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var byRoutes []RouteStats
	for _, s := range recs.HistoryBased {
		byRoutes = append(byRoutes, RouteStats{From: s.From, To: s.To, TripsCount: s.TripsCount})
	}
	var popular []RouteStats
	for _, s := range recs.TopGlobal {
		popular = append(popular, RouteStats{From: s.From, To: s.To, TripsCount: s.TripsCount})
	}
	var unvisited []CityStats
	for _, c := range recs.Unvisited {
		unvisited = append(unvisited, CityStats{City: c.City, Lat: c.Lat, Lon: c.Lon})
	}

	c.JSON(http.StatusOK, RecommendationsResponse{
		ByYourRoutes:  byRoutes,
		Popular:       popular,
		NewDirections: unvisited,
	})
}

// BookRide godoc
// @Summary Book seats in a ride
// @Tags rides
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body BookRideRequest true "Booking details"
// @Success 200 {object} BookingResponse
// @Router /api/v1/rides/{id}/book [post]
// @Security BearerAuth
func (h *Handler) BookRide(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	passengerIDStr := c.GetString("user_id")
	if passengerIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	passengerID, err := uuid.Parse(passengerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	var req BookRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	booking, err := h.useCase.BookRide(c.Request.Context(), rideID, passengerID, req.Seats, req.IsFullCarBuyout, req.OfferedPrice, req.PickupPreference, req.DropoffType)
	if err != nil {
		if errors.Is(err, domain.ErrRideDeparted) || errors.Is(err, domain.ErrRideNotActive) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ride booked successfully", "booking": booking})
}

// AcceptBooking godoc
// @Summary Accept a pending negotiation
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Param booking_id path string true "Booking ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/rides/{id}/bookings/{booking_id}/accept [post]
// @Security BearerAuth
func (h *Handler) AcceptBooking(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	bookingIDStr := c.Param("booking_id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id format"})
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, _ := uuid.Parse(driverIDStr)

	err = h.useCase.AcceptBooking(c.Request.Context(), driverID, rideID, bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking accepted successfully"})
}

// RejectBooking godoc
// @Summary Reject a pending negotiation
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Param booking_id path string true "Booking ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/rides/{id}/bookings/{booking_id}/reject [post]
// @Security BearerAuth
func (h *Handler) RejectBooking(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	bookingIDStr := c.Param("booking_id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id format"})
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, _ := uuid.Parse(driverIDStr)

	err = h.useCase.RejectBooking(c.Request.Context(), driverID, rideID, bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking rejected successfully"})
}

// CounterBooking godoc
// @Summary Counter a passenger's booking price
// @Tags rides
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param booking_id path string true "Booking ID"
// @Param request body CounterBookingRequest true "New Price"
// @Success 200 {object} map[string]string
// @Router /api/v1/rides/{id}/bookings/{booking_id}/counter [post]
// @Security BearerAuth
func (h *Handler) CounterBooking(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	bookingIDStr := c.Param("booking_id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id format"})
		return
	}

	var req CounterBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, _ := uuid.Parse(driverIDStr)

	err = h.useCase.CounterBooking(c.Request.Context(), driverID, rideID, bookingID, req.NewPrice)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "counter price sent successfully"})
}

// ReportNoShow godoc
// @Summary Report a passenger no-show
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Param booking_id path string true "Booking ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/rides/{id}/bookings/{booking_id}/no-show [post]
// @Security BearerAuth
func (h *Handler) ReportNoShow(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	bookingIDStr := c.Param("booking_id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id format"})
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, _ := uuid.Parse(driverIDStr)

	err = h.useCase.ReportPassengerNoShow(c.Request.Context(), driverID, rideID, bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "passenger no-show reported successfully"})
}

// ReportBookingNoShow godoc
// @Summary Report a passenger no-show via booking ID
// @Tags bookings
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/bookings/{id}/no-show [post]
// @Security BearerAuth
func (h *Handler) ReportBookingNoShow(c *gin.Context) {
	bookingIDStr := c.Param("id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id format"})
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, _ := uuid.Parse(driverIDStr)

	// We need to fetch the booking to get the ride ID, since useCase requires it.
	booking, err := h.useCase.GetBookingByID(c.Request.Context(), bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	err = h.useCase.ReportPassengerNoShow(c.Request.Context(), driverID, booking.RideID, bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "passenger no-show reported successfully"})
}

type CancelBookingRequest struct {
	Reason string `json:"reason"`
}

// CancelBooking godoc
// @Summary Cancel a booking (for passenger)
// @Tags bookings
// @Produce json
// @Param id path string true "Booking ID"
// @Param request body CancelBookingRequest false "Cancel reason"
// @Success 200 {object} map[string]string
// @Router /api/v1/bookings/{id}/cancel [post]
// @Security BearerAuth
func (h *Handler) CancelBooking(c *gin.Context) {
	bookingIDStr := c.Param("id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id format"})
		return
	}

	passengerIDStr := c.GetString("user_id")
	passengerID, _ := uuid.Parse(passengerIDStr)

	var req CancelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		// Ignore EOF since body is optional
	}

	err = h.useCase.CancelBookingByPassenger(c.Request.Context(), passengerID, bookingID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled successfully"})
}

// CancelRideBooking godoc
// @Summary Cancel a booking in a ride (for passenger)
// @Description Allows a passenger to cancel their own booking within a specific ride
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Param booking_id path string true "Booking ID"
// @Param request body CancelBookingRequest false "Cancel reason"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/rides/{id}/bookings/{booking_id}/cancel [post]
// @Security BearerAuth
func (h *Handler) CancelRideBooking(c *gin.Context) {
	bookingIDStr := c.Param("booking_id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking_id format"})
		return
	}

	passengerIDStr := c.GetString("user_id")
	passengerID, _ := uuid.Parse(passengerIDStr)

	var req CancelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		// Ignore EOF since body is optional
	}

	err = h.useCase.CancelBookingByPassenger(c.Request.Context(), passengerID, bookingID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled successfully"})
}


// ActivateRide godoc
// @Summary Activate a draft ride
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string "Error message (e.g., filters not available on Plus plan or insufficient limits)"
// @Router /api/v1/rides/{id}/activate [post]
// @Security BearerAuth
func (h *Handler) ActivateRide(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, _ := uuid.Parse(driverIDStr)

	err = h.useCase.ActivateRide(c.Request.Context(), driverID, rideID)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientRideLimits) {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "payment required: " + err.Error()})
			return
		}
		errMsg := err.Error()
		if strings.Contains(errMsg, "Unavailable") || strings.Contains(errMsg, "connection error") || strings.Contains(errMsg, "i/o timeout") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable: " + errMsg})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ride activated successfully"})
}

// StartRide godoc
// @Summary Start an active ride
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/rides/{id}/start [post]
// @Security BearerAuth
func (h *Handler) StartRide(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, _ := uuid.Parse(driverIDStr)

	err = h.useCase.StartRide(c.Request.Context(), driverID, rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ride started successfully"})
}

type CancelRideRequest struct {
	Reason string `json:"reason"`
}

// CancelRide godoc
// @Summary Cancel a ride (driver)
// @Description Allows the driver to cancel their ride. Supports POST /:id/cancel, DELETE /:id, DELETE /:id/cancel.
// @Tags rides
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body CancelRideRequest false "Cancel reason"
// @Param reason query string false "Cancel reason (alternative via query parameter)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/rides/{id}/cancel [post]
// @Security BearerAuth
func (h *Handler) CancelRide(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	driverIDStr := c.GetString("user_id")
	if driverIDStr == "" {
		driverIDStr = c.GetHeader("X-User-Id")
	}
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CancelRideRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		// allow empty body
	}

	reason := req.Reason
	if reason == "" {
		reason = c.Query("reason")
	}

	err = h.useCase.CancelRide(c.Request.Context(), driverID, rideID, reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ride cancelled successfully"})
}

// CompleteRide godoc
// @Summary Complete a ride
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/rides/{id}/complete [post]
// @Security BearerAuth
func (h *Handler) CompleteRide(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	driverIDStr := c.GetString("user_id")
	driverID, _ := uuid.Parse(driverIDStr)

	err = h.useCase.CompleteRide(c.Request.Context(), driverID, rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ride completed successfully"})
}

// GetMyRides godoc
// @Summary Get user's rides or bookings
// @Tags rides
// @Produce json
// @Param role query string true "Role (driver or passenger)" Enums(driver, passenger)
// @Param status query string false "Comma-separated statuses to filter by"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/rides/my [get]
// @Security BearerAuth
func (h *Handler) GetMyRides(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	statusQuery := c.Query("status")
	statusFilters := make(map[string]bool)
	if statusQuery != "" {
		for _, s := range strings.Split(statusQuery, ",") {
			statusFilters[strings.TrimSpace(s)] = true
		}
	}

	role := c.Query("role")
	if role == "" {
		role = c.GetString("role")
	}
	if role == "" {
		role = "driver"
	}
	switch role {
	case "driver":
		rides, err := h.useCase.GetMyRidesAsDriver(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if rides == nil {
			rides = []domain.Ride{}
		}

		if len(statusFilters) > 0 {
			var filtered []domain.Ride
			for _, r := range rides {
				if statusFilters[string(r.Status)] {
					filtered = append(filtered, r)
				}
			}
			rides = filtered
		}

		c.JSON(http.StatusOK, gin.H{"rides": rides})
		return
	case "passenger":
		bookings, err := h.useCase.GetMyBookingsAsPassenger(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if bookings == nil {
			bookings = []domain.Booking{}
		}

		if len(statusFilters) > 0 {
			var filtered []domain.Booking
			for _, b := range bookings {
				if statusFilters[string(b.Status)] {
					filtered = append(filtered, b)
				}
			}
			bookings = filtered
		}
		c.JSON(http.StatusOK, gin.H{"bookings": bookings})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role parameter, must be 'driver' or 'passenger'"})
		return
	}
}

// GetDriverStatistics godoc
// @Summary Get driver's earnings and statistics
// @Description Returns total rides, completed, cancelled, and total earnings
// @Tags rides
// @Produce json
// @Success 200 {object} DriverStatisticsResponse
// @Router /api/v1/rides/earnings [get]
// @Security BearerAuth
func (h *Handler) GetDriverStatistics(c *gin.Context) {
	driverIDStr := c.GetString("user_id")
	if driverIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
		return
	}

	groupBy := c.Query("group_by")
	if groupBy != "" && groupBy != "total" && groupBy != "all" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "group_by is not supported yet for earnings, only total aggregated statistics are available",
		})
		return
	}

	stats, err := h.useCase.GetDriverStatistics(c.Request.Context(), driverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserStats godoc
// @Summary Получить статистику поездок пользователя (Админ)
// @Description Получает количество поездок пользователя как водителя и пассажира
// @Tags rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} UserRideStatsResponse
// @Router /api/v1/rides/admin/user/{id}/stats [get]
func (h *Handler) GetUserStats(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" && role != "system" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	stats, err := h.useCase.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRideByID godoc
// @Summary Get a single ride by ID
// @Description Returns full details of a specific ride
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Success 200 {object} RideResponse
// @Router /api/v1/rides/{id} [get]
func (h *Handler) GetRideByID(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ride not found"})
		return
	}

	ride, err := h.useCase.GetRideByID(c.Request.Context(), rideID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ride not found"})
		return
	}

	// Check if requesting user has a confirmed booking for this ride
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		userIDStr = c.GetHeader("X-User-Id")
	}

	hasConfirmedBooking := false
	if userIDStr != "" {
		if uID, err := uuid.Parse(userIDStr); err == nil {
			if uID == ride.DriverID {
				hasConfirmedBooking = true
			} else {
				for _, b := range ride.Bookings {
					if b.PassengerID == uID && (b.Status == domain.BookingStatusConfirmed || b.Status == "accepted") {
						hasConfirmedBooking = true
						break
					}
				}
			}
		}
	}

	if !hasConfirmedBooking && ride.Driver != nil {
		driverCopy := *ride.Driver
		driverCopy.Phone = ""
		ride.Driver = &driverCopy
	}

	if userIDStr == "" {
		ride.Bookings = nil // Hide passengers for unauthenticated users
		ride.PickupPoint = ""
	}

	var driver *DriverProfile
	if ride.Driver != nil {
		driver = &DriverProfile{
			ID:          ride.Driver.ID,
			Name:        ride.Driver.Name,
			Phone:       ride.Driver.Phone,
			PhoneNumber: ride.Driver.Phone,
			Rating:      ride.Driver.Rating,
			RatingCount: ride.Driver.RatingCount,
			Avatar:      ride.Driver.Avatar,
			Gender:      ride.Driver.Gender,
			KycStatus:   ride.Driver.KycStatus,
			CreatedAt:   ride.Driver.CreatedAt,
		}
	}

	var car *Vehicle
	if ride.VehicleData != nil {
		car = &Vehicle{
			ID:               ride.VehicleData.ID,
			Brand:            ride.VehicleData.Brand,
			Model:            ride.VehicleData.Model,
			Color:            ride.VehicleData.Color,
			LicensePlate:     ride.VehicleData.LicensePlate,
			Year:             ride.VehicleData.Year,
			ModerationStatus: ride.VehicleData.ModerationStatus,
		}
	}
	
	var stops []RideStop
	for _, s := range ride.Stops {
		stops = append(stops, RideStop{
			ID:         s.ID,
			RideID:     s.RideID,
			Lat:        s.Lat,
			Lon:        s.Lon,
			Address:    s.Address,
			OrderIndex: s.OrderIndex,
		})
	}

	var musicTaste string
	if ride.MusicTaste != nil {
		musicTaste = *ride.MusicTaste
	}

	passengers := make([]domain.PassengerProfile, 0)
	for _, b := range ride.Bookings {
		if (b.Status == domain.BookingStatusConfirmed || b.Status == "accepted") && b.Passenger != nil {
			phone := b.Passenger.PhoneNumber
			if !hasConfirmedBooking {
				phone = ""
			}
			passengers = append(passengers, domain.PassengerProfile{
				ID:          b.Passenger.ID,
				Name:        b.Passenger.Name,
				Avatar:      b.Passenger.Avatar,
				Rating:      b.Passenger.Rating,
				PhoneNumber: phone,
				Gender:      b.Passenger.Gender,
				KycStatus:   b.Passenger.KycStatus,
				CreatedAt:   b.Passenger.CreatedAt,
			})
		}
	}

	resp := RideResponse{
		ID:                 ride.ID,
		DriverID:           ride.DriverID,
		VehicleID:          ride.VehicleID,
		OriginLat:          ride.OriginLat,
		OriginLon:          ride.OriginLon,
		DestinationLat:     ride.DestinationLat,
		DestinationLon:     ride.DestinationLon,
		OriginAddress:      ride.OriginAddress,
		DestinationAddress: ride.DestinationAddress,
		DepartureTime:      ride.DepartureTime.Format(time.RFC3339),
		TotalSeats:         ride.TotalSeats,
		AvailableSeats:     ride.AvailableSeats,
		PricePerSeat:       ride.PricePerSeat,
		Status:             string(ride.Status),
		IsNegotiable:       ride.IsNegotiable,
		HasFairPriceBadge:  ride.HasFairPriceBadge,
		GenderRestriction:  ride.GenderRestriction,
		HasAc:              ride.HasAc,
		HasLuggage:         ride.HasLuggage,
		TakesParcels:       ride.TakesParcels,
		NoSmoking:          ride.NoSmoking,
		NoPets:             ride.NoPets,
		DrivingSpeed:       ride.DrivingSpeed,
		MaxTwoBack:         ride.MaxTwoBack,
		PickupPoint:        ride.PickupPoint,
		PickupLat:          ride.PickupLat,
		PickupLon:          ride.PickupLon,
		MusicTaste:         musicTaste,
		AllowsCityDetour:   ride.AllowsCityDetour,
		CityDetourFee:      ride.CityDetourFee,
		CreatedAt:          ride.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          ride.UpdatedAt.Format(time.RFC3339),
		Stops:              stops,
		Driver:             driver,
		Car:                car,
		Passengers:         passengers,
	}

	c.JSON(http.StatusOK, resp)
}

// GetRidePassengers godoc
// @Summary Get confirmed passengers for a ride
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/rides/{id}/passengers [get]
func (h *Handler) GetRidePassengers(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id format"})
		return
	}

	callerIDStr := c.GetString("user_id")
	callerID, err := uuid.Parse(callerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	role := c.GetString("role")

	ride, err := h.useCase.GetRideByID(c.Request.Context(), rideID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ride not found"})
		return
	}

	// RBAC check: caller must be the driver OR a confirmed passenger (or admin/moderator/system)
	isAuthorized := (ride.DriverID == callerID) || role == "admin" || role == "moderator" || role == "system"
	if !isAuthorized {
		for _, b := range ride.Bookings {
			if b.PassengerID == callerID && (b.Status == domain.BookingStatusConfirmed || b.Status == "accepted") {
				isAuthorized = true
				break
			}
		}
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: you must be the driver or a confirmed passenger of this ride"})
		return
	}

	passengers := make([]domain.PassengerProfile, 0)
	for _, b := range ride.Bookings {
		if (b.Status == domain.BookingStatusConfirmed || b.Status == "accepted") && b.Passenger != nil {
			passengers = append(passengers, domain.PassengerProfile{
				ID:          b.Passenger.ID,
				Name:        b.Passenger.Name,
				Avatar:      b.Passenger.Avatar,
				Rating:      b.Passenger.Rating,
				PhoneNumber: b.Passenger.PhoneNumber,
				Gender:      b.Passenger.Gender,
				KycStatus:   b.Passenger.KycStatus,
				CreatedAt:   b.Passenger.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"passengers": passengers, "count": len(passengers)})
}

// --- Route Subscriptions ---

type CreateRouteSubscriptionRequest struct {
	OriginLat     float64    `json:"origin_lat" binding:"required"`
	OriginLon     float64    `json:"origin_lon" binding:"required"`
	DestLat       float64    `json:"dest_lat" binding:"required"`
	DestLon       float64    `json:"dest_lon" binding:"required"`
	DepartureDate *time.Time `json:"departure_date,omitempty"`
}

// CreateRouteSubscription godoc
// @Summary Create route subscription
// @Description Subscribe to notifications for rides along a specific route
// @Tags rides
// @Accept json
// @Produce json
// @Param request body CreateRouteSubscriptionRequest true "Route Subscription Request"
// @Success 201 {object} RouteSubscriptionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/rides/route-subscriptions [post]
// @Security BearerAuth
func (h *Handler) CreateRouteSubscription(c *gin.Context) {
	passengerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	pID, err := uuid.Parse(passengerID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req CreateRouteSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub := &domain.RouteSubscription{
		PassengerID:   pID,
		OriginLat:     req.OriginLat,
		OriginLon:     req.OriginLon,
		DestLat:       req.DestLat,
		DestLon:       req.DestLon,
		DepartureDate: req.DepartureDate,
	}

	if err := h.useCase.CreateRouteSubscription(c.Request.Context(), sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// GetRouteSubscriptions godoc
// @Summary List route subscriptions
// @Description Returns list of active route subscriptions for current passenger
// @Tags rides
// @Produce json
// @Success 200 {array} RouteSubscriptionResponse
// @Failure 401 {object} map[string]string
// @Router /api/v1/rides/route-subscriptions [get]
// @Security BearerAuth
func (h *Handler) GetRouteSubscriptions(c *gin.Context) {
	passengerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	pID, err := uuid.Parse(passengerID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	subs, err := h.useCase.GetRouteSubscriptions(c.Request.Context(), pID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if subs == nil {
		subs = []domain.RouteSubscription{}
	}
	c.JSON(http.StatusOK, subs)
}

// DeleteRouteSubscription godoc
// @Summary Delete route subscription
// @Description Removes a route subscription by ID
// @Tags rides
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/rides/route-subscriptions/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeleteRouteSubscription(c *gin.Context) {
	passengerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	pID, err := uuid.Parse(passengerID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	subIDStr := c.Param("id")
	subID, err := uuid.Parse(subIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
		return
	}

	if err := h.useCase.DeleteRouteSubscription(c.Request.Context(), subID, pID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription deleted"})
}


// GetUserActiveRide godoc
// @Summary Fetch user active ride
// @Description Gets the active ride for a specific user (as driver or passenger)
// @Tags rides
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} RideResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rides/admin/user/{id}/active [get]
func (h *Handler) GetUserActiveRide(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	ride, err := h.useCase.GetUserActiveRide(c.Request.Context(), targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ride == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active ride found"})
		return
	}

	c.JSON(http.StatusOK, ride)
}

// GetAdminActiveRides godoc
// @Summary Fetch all active rides for admin map
// @Tags rides
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/rides/admin/active [get]
func (h *Handler) GetAdminActiveRides(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" && role != "system" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	rides, err := h.useCase.GetAdminActiveRides(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	statusFilter := c.Query("status")
	
	var filtered []domain.Ride
	for _, r := range rides {
		if statusFilter == "" || statusFilter == "active" {
			if r.Status != domain.RideStatusCompleted && r.Status != domain.RideStatusCancelled {
				filtered = append(filtered, r)
			}
		} else if statusFilter == string(r.Status) {
			filtered = append(filtered, r)
		}
	}
	
	if filtered == nil {
		filtered = []domain.Ride{}
	}

	c.JSON(http.StatusOK, gin.H{"rides": filtered})
}

// GetAdminStats godoc
// @Summary Fetch ride statistics
// @Description Fetch aggregated statistics for rides for admin dashboard
// @Tags rides
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AdminStatsResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rides/admin/stats [get]
func (h *Handler) GetAdminStats(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "moderator" && role != "system" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: insufficient permissions"})
		return
	}

	stats, err := h.useCase.GetAdminStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDriverLocation godoc
// DriverLocationResponse represents driver's telemetry for a ride
type DriverLocationResponse struct {
	DriverID  uuid.UUID `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Heading   float64   `json:"heading"`
	Speed     float64   `json:"speed"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetDriverLocation godoc
// @Summary Получение актуального местоположения водителя поездки
// @Description Возвращает последние GPS координаты, скорость, угол направления и время обновления водителя. Доступно только водителю и подтвержденным пассажирам поездки.
// @Tags rides
// @Produce json
// @Security BearerAuth
// @Param id path string true "Ride ID"
// @Success 200 {object} DriverLocationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rides/{id}/driver-location [get]
func (h *Handler) GetDriverLocation(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id"})
		return
	}

	callerIDStr := c.GetString("user_id")
	callerID, err := uuid.Parse(callerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	loc, err := h.useCase.GetRideDriverLocation(c.Request.Context(), rideID, callerID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
			return
		}
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "offline") || strings.Contains(errMsg, "not available") || strings.Contains(errMsg, "outdated") || strings.Contains(errMsg, "not active") {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, loc)
}

// GetRideLocation godoc
// @Summary Получение координат водителя по поездке (через Bearer токен или Watch Token)
// @Description Возвращает текущие координаты водителя для участника поездки или доверенного наблюдателя с watch_token.
// @Tags rides
// @Produce json
// @Param id path string true "Ride ID"
// @Param watch_token query string false "Временный токен наблюдения"
// @Success 200 {object} domain.WatchLocationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rides/{id}/location [get]
func (h *Handler) GetRideLocation(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id"})
		return
	}

	// 1. Check if watch_token is provided (query param or header)
	watchToken := c.Query("watch_token")
	if watchToken == "" {
		watchToken = c.GetHeader("X-Watch-Token")
	}
	if watchToken != "" {
		loc, err := h.useCase.GetRideDriverLocationByWatchToken(c.Request.Context(), rideID, watchToken)
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "expired") {
				c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
				return
			}
			if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "offline") || strings.Contains(errMsg, "not available") || strings.Contains(errMsg, "outdated") || strings.Contains(errMsg, "not active") {
				c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusOK, loc)
		return
	}

	// 2. Otherwise require authenticated caller
	callerIDStr := c.GetString("user_id")
	if callerIDStr == "" {
		callerIDStr = c.GetHeader("X-User-Id")
	}
	if callerIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header or token query parameter is required"})
		return
	}

	callerID, err := uuid.Parse(callerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	loc, err := h.useCase.GetRideDriverLocation(c.Request.Context(), rideID, callerID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
			return
		}
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "offline") || strings.Contains(errMsg, "not available") || strings.Contains(errMsg, "outdated") || strings.Contains(errMsg, "not active") {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, loc)
}

func (h *Handler) CreateWatchToken(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id"})
		return
	}

	callerIDStr := c.GetString("user_id")
	if callerIDStr == "" {
		callerIDStr = c.GetHeader("X-User-Id")
	}
	callerID, err := uuid.Parse(callerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.useCase.CreateWatchToken(c.Request.Context(), rideID, callerID)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "unauthorized") {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
			return
		}
		if strings.Contains(errMsg, "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) RevokeWatchToken(c *gin.Context) {
	rideIDStr := c.Param("id")
	rideID, err := uuid.Parse(rideIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ride id"})
		return
	}

	callerIDStr := c.GetString("user_id")
	if callerIDStr == "" {
		callerIDStr = c.GetHeader("X-User-Id")
	}
	callerID, err := uuid.Parse(callerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	token := c.Query("token")
	if token == "" {
		var req struct {
			Token string `json:"token"`
		}
		_ = c.ShouldBindJSON(&req)
		token = req.Token
	}
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	if err := h.useCase.RevokeWatchToken(c.Request.Context(), rideID, callerID, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) GetDriverPublicRides(c *gin.Context) {
	driverIDStr := c.Param("driver_id")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver id"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rides, err := h.useCase.GetDriverPublicRides(c.Request.Context(), driverID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rides": rides})
}
