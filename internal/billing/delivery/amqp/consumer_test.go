package amqp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	billingMocks "github.com/hextec/yolyolakay-backend/pkg/mocks/billing"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	amqp091 "github.com/rabbitmq/amqp091-go"
	"go.uber.org/mock/gomock"
)

type dummyAcknowledger struct{}

func (d dummyAcknowledger) Ack(tag uint64, multiple bool) error                { return nil }
func (d dummyAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error { return nil }
func (d dummyAcknowledger) Reject(tag uint64, requeue bool) error              { return nil }

func TestHandleBookingConfirmed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := billingMocks.NewMockBillingRepository(ctrl)
	useCase := billingMocks.NewMockBillingUseCase(ctrl)

	// Create a client pointer but don't connect
	rmq := &rabbitmq.Client{}

	consumer := NewBillingEventsConsumer(rmq, repo, useCase)

	passengerID := uuid.New()
	rideID := uuid.New().String()
	bookingID := uuid.New().String()

	payload := map[string]interface{}{
		"ride_id":      rideID,
		"passenger_id": passengerID.String(),
		"booking_id":   bookingID,
	}
	body, _ := json.Marshal(payload)

	d := amqp091.Delivery{
		RoutingKey:   "ride.booking.confirmed",
		Body:         body,
		Acknowledger: dummyAcknowledger{},
	}

	// Passengers do not have ride limits, so DeductRideLimit must NOT be called.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Direct call to handler
	consumer.handleBookingConfirmed(ctx, d)
}

func TestHandleRideCompleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := billingMocks.NewMockBillingRepository(ctrl)
	useCase := billingMocks.NewMockBillingUseCase(ctrl)

	// Create a dummy client
	rmq := &rabbitmq.Client{}
	consumer := NewBillingEventsConsumer(rmq, repo, useCase)

	driverID := uuid.New()
	rideID := uuid.New().String()

	payload := map[string]interface{}{
		"ride_id":   rideID,
		"driver_id": driverID.String(),
	}
	body, _ := json.Marshal(payload)

	d := amqp091.Delivery{
		RoutingKey:   "ride.completed",
		Body:         body,
		Acknowledger: dummyAcknowledger{},
	}

	// EXPECT ExpireTrialSubscription to be called with driverID
	useCase.EXPECT().ExpireTrialSubscription(gomock.Any(), driverID).Return(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Call the handler directly
	consumer.handleRideCompleted(ctx, d)
}
