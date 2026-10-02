package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/notification/domain"
	"github.com/hextec/yolyolakay-backend/internal/notification/i18n"
	rmqPkg "github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
)

type Consumer struct {
	client     *rmqPkg.Client
	uc         domain.NotificationUseCase
	translator *i18n.Translator
}

func NewConsumer(client *rmqPkg.Client, uc domain.NotificationUseCase) *Consumer {
	return &Consumer{
		client:     client,
		uc:         uc,
		translator: i18n.NewTranslator(),
	}
}

func (c *Consumer) Start() {
	// Let's give RabbitMQ client some time to establish connection
	time.Sleep(2 * time.Second)

	// Declare exchange and queue with DLQ for billing
	err := c.client.SetupExchangeQueueWithDLQ(
		"billing.events",
		"topic",
		"notification.billing.events",
		"billing.#",
		"billing.events.dlx",
		"notification.billing.events.dlq",
	)
	if err != nil {
		slog.Error("Failed to setup exchange/queue for billing", "error", err)
		return
	}

	// Declare exchange and queue with DLQ for chat messages
	err = c.client.SetupExchangeQueueWithDLQ(
		"yolyolakay.events",
		"topic",
		"notification.chat.events",
		"chat.message.saved",
		"yolyolakay.events.dlx",
		"notification.chat.events.dlq",
	)
	if err != nil {
		slog.Error("Failed to setup exchange/queue for chat messages", "error", err)
		return
	}

	// Declare exchange and queue with DLQ for auth events (logout)
	err = c.client.SetupExchangeQueueWithDLQ(
		"yolyolakay.events",
		"topic",
		"notification.auth.events",
		"user.logged_out",
		"yolyolakay.events.dlx",
		"notification.auth.events.dlq",
	)
	if err != nil {
		slog.Error("Failed to setup exchange/queue for auth messages", "error", err)
		return
	}
	// Declare exchange and queue with DLQ for ride events
	err = c.client.SetupExchangeQueueWithDLQ(
		"yolyolakay.events",
		"topic",
		"notification.ride.events",
		"ride.#",
		"yolyolakay.events.dlx",
		"notification.ride.events.dlq",
	)
	if err != nil {
		slog.Error("Failed to setup exchange/queue for ride messages", "error", err)
		return
	}

	// Declare exchange and queue with DLQ for general push notifications
	err = c.client.SetupExchangeQueueWithDLQ(
		"notification_exchange",
		"topic",
		"notification.general.events",
		"notification.#",
		"notification_exchange.dlx",
		"notification.general.events.dlq",
	)
	if err != nil {
		slog.Error("Failed to setup exchange/queue for general notifications", "error", err)
		return
	}

	// Declare exchange and queue with DLQ for parcel events
	err = c.client.SetupExchangeQueueWithDLQ(
		"yolyolakay.events",
		"topic",
		"notification.parcel.events",
		"parcel.#",
		"yolyolakay.events.dlx",
		"notification.parcel.events.dlq",
	)
	if err != nil {
		slog.Error("Failed to setup exchange/queue for parcel messages", "error", err)
		return
	}

	go c.consumeBillingLoop()
	go c.consumeChatLoop()
	go c.consumeAuthLoop()
	go c.consumeRideLoop()
	go c.consumeSystemPushLoop()
	go c.consumeParcelLoop()
}

func (c *Consumer) consumeBillingLoop() {
	for {
		msgs, err := c.client.Consume("notification.billing.events")
		if err != nil {
			slog.Error("Consumer failed to start: %v. Retrying in 5s...", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		slog.Info("Consumer started successfully for notification.billing.events")

		for msg := range msgs {
			slog.Info("Received message", "body", string(msg.Body))

			// Parse the event (for billing.subscription.activated)
			var payload map[string]interface{}
			if err := json.Unmarshal(msg.Body, &payload); err == nil {
				if driverIDStr, ok := payload["driver_id"].(string); ok {
					if driverID, err := uuid.Parse(driverIDStr); err == nil {

						eventStr, _ := payload["event"].(string)
						plan, _ := payload["plan"].(string)

						// 1. Determine Language
						lang := c.uc.GetUserLocale(context.Background(), driverID)

						// 2. Translate content
						content := c.translator.Translate(lang, "new_message")
						title := "Yolyolakay"
						if eventStr == "subscription_activated" {
							content = c.translator.Translate(lang, "subscription_activated", plan)
						} else if eventStr == "trial_started" {
							title = c.translator.Translate(lang, "trial_started_title")
							content = c.translator.Translate(lang, "trial_started_body")
						} else if eventStr == "limit_exceeded" {
							title = c.translator.Translate(lang, "limit_exceeded_title")
							content = c.translator.Translate(lang, "limit_exceeded_body")
						}

						err := c.uc.SendSystemPush(context.Background(), driverID, title, content, eventStr)
						if err != nil {
							slog.Error("Failed to push system notification", "error", err)
						} else {
							slog.Info("System notification sent to driver", "driver_id", driverIDStr)
						}
					}
				}
			}

			msg.Ack(false)
		}

		slog.Info("Consumer channel closed for billing events, attempting to reconnect...")
		time.Sleep(5 * time.Second)
	}
}

func (c *Consumer) consumeChatLoop() {
	for {
		msgs, err := c.client.Consume("notification.chat.events")
		if err != nil {
			slog.Error("Chat Consumer failed to start: %v. Retrying in 5s...", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		slog.Info("Consumer started successfully for notification.chat.events")

		for msg := range msgs {
			var chatMsg domain.Message
			if err := json.Unmarshal(msg.Body, &chatMsg); err == nil {
				// Async save to db
				if err := c.uc.SaveMessageToDB(context.Background(), &chatMsg); err != nil {
					slog.Error("Failed to save chat message to DB", "error", err)
					msg.Nack(false, true)
					continue
				}
			} else {
				slog.Error("Failed to unmarshal chat message", "error", err)
			}
			msg.Ack(false)
		}

		slog.Info("Consumer channel closed for chat events, attempting to reconnect...")
		time.Sleep(5 * time.Second)
	}
}

func (c *Consumer) consumeAuthLoop() {
	for {
		msgs, err := c.client.Consume("notification.auth.events")
		if err != nil {
			slog.Error("Auth Consumer failed to start: %v. Retrying in 5s...", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		slog.Info("Consumer started successfully for notification.auth.events")

		for msg := range msgs {
			// Handle user logout: clean up push tokens
			var ev struct {
				UserID    string `json:"user_id"`
				SessionID string `json:"session_id"`
			}
			if err := json.Unmarshal(msg.Body, &ev); err == nil {
				if uID, err := uuid.Parse(ev.UserID); err == nil {
					var sID uuid.UUID
					if ev.SessionID != "" {
						sID, _ = uuid.Parse(ev.SessionID)
					}
					_ = c.uc.DeleteDeviceToken(context.Background(), uID, sID)
				}
			}
			msg.Ack(false)
		}

		slog.Info("Consumer channel closed for auth events, attempting to reconnect...")
		time.Sleep(5 * time.Second)
	}
}

func (c *Consumer) consumeRideLoop() {
	for {
		msgs, err := c.client.Consume("notification.ride.events")
		if err != nil {
			slog.Error("Ride Consumer failed to start: %v. Retrying in 5s...", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		slog.Info("Consumer started successfully for notification.ride.events")

		for msg := range msgs {
			slog.Info("Received ride message", "routingKey", msg.RoutingKey, "body", string(msg.Body))

			if msg.RoutingKey == "ride.negotiation_requested" {
				var ev struct {
					RideID       string  `json:"ride_id"`
					DriverID     string  `json:"driver_id"`
					PassengerID  string  `json:"passenger_id"`
					OfferedPrice float64 `json:"offered_price"`
				}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					if driverID, err := uuid.Parse(ev.DriverID); err == nil {
						lang := c.uc.GetUserLocale(context.Background(), driverID)
						title := c.translator.Translate(lang, "negotiation_requested")
						body := fmt.Sprintf("Yo'lovchi yangi narx taklif qildi: %.0f", ev.OfferedPrice)
						if lang == "ru" {
							body = fmt.Sprintf("Пассажир предложил новую цену: %.0f", ev.OfferedPrice)
						} else if lang == "en" {
							body = fmt.Sprintf("Passenger offered a new price: %.0f", ev.OfferedPrice)
						}
						data := map[string]string{
							"type":          "negotiation_requested",
							"title":         title,
							"body":          body,
							"ride_id":       ev.RideID,
							"passenger_id":  ev.PassengerID,
							"offered_price": fmt.Sprintf("%.2f", ev.OfferedPrice),
						}
						_ = c.uc.SendSystemPushWithData(context.Background(), driverID, title, body, data)
					}
				} else {
					slog.Error("Failed to unmarshal ride event", "error", err)
				}
			} else if msg.RoutingKey == "ride.status_changed" {
				var ev struct {
					RideID       string   `json:"ride_id"`
					Status       string   `json:"status"`
					PassengerIDs []string `json:"passenger_ids"`
				}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					for _, pIDStr := range ev.PassengerIDs {
						if pID, err := uuid.Parse(pIDStr); err == nil {
							if err := c.uc.PushWSEvent(context.Background(), pID, "ride_status_changed", ev); err != nil {
								slog.Error("Failed to push ride status event", "error", err)
							}
						}
					}
				} else {
					slog.Error("Failed to unmarshal ride event", "error", err)
				}
			} else if msg.RoutingKey == "ride.booking.created" {
				var ev struct {
					BookingID   string `json:"booking_id"`
					RideID      string `json:"ride_id"`
					DriverID    string `json:"driver_id"`
					PassengerID string `json:"passenger_id"`
					SeatsBooked int32  `json:"seats_booked"`
				}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					if driverID, err := uuid.Parse(ev.DriverID); err == nil {
						lang := c.uc.GetUserLocale(context.Background(), driverID)
						title := c.translator.Translate(lang, "booking_created_title")
						body := c.translator.Translate(lang, "booking_created_body")
						data := map[string]string{
							"type":       "booking_created",
							"title":      title,
							"body":       body,
							"ride_id":    ev.RideID,
							"booking_id": ev.BookingID,
						}
						_ = c.uc.SendSystemPushWithData(context.Background(), driverID, title, body, data)
					}
				}
			} else if msg.RoutingKey == "ride.booking.confirmed" {
				var ev map[string]interface{}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					pIDStr, _ := ev["passenger_id"].(string)
					rideIDStr, _ := ev["ride_id"].(string)
					bookingIDStr, _ := ev["booking_id"].(string)
					if pID, err := uuid.Parse(pIDStr); err == nil {
						lang := c.uc.GetUserLocale(context.Background(), pID)
						title := c.translator.Translate(lang, "booking_confirmed_title")
						body := c.translator.Translate(lang, "booking_confirmed_body")
						data := map[string]string{
							"type":       "booking_confirmed",
							"title":      title,
							"body":       body,
							"ride_id":    rideIDStr,
							"booking_id": bookingIDStr,
						}
						_ = c.uc.SendSystemPushWithData(context.Background(), pID, title, body, data)
					}
				}
			} else if msg.RoutingKey == "ride.booking.rejected" {
				var ev struct {
					BookingID   string `json:"booking_id"`
					RideID      string `json:"ride_id"`
					PassengerID string `json:"passenger_id"`
				}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					if pID, err := uuid.Parse(ev.PassengerID); err == nil {
						lang := c.uc.GetUserLocale(context.Background(), pID)
						title := c.translator.Translate(lang, "booking_rejected_title")
						body := c.translator.Translate(lang, "booking_rejected_body")
						data := map[string]string{
							"type":       "booking_rejected",
							"title":      title,
							"body":       body,
							"ride_id":    ev.RideID,
							"booking_id": ev.BookingID,
						}
						_ = c.uc.SendSystemPushWithData(context.Background(), pID, title, body, data)
					}
				}
			} else if msg.RoutingKey == "ride.booking.cancelled" {
				var ev struct {
					BookingID   string `json:"booking_id"`
					RideID      string `json:"ride_id"`
					DriverID    string `json:"driver_id"`
					PassengerID string `json:"passenger_id"`
					Reason      string `json:"reason"`
				}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					if driverID, err := uuid.Parse(ev.DriverID); err == nil {
						lang := c.uc.GetUserLocale(context.Background(), driverID)
						title := c.translator.Translate(lang, "booking_cancelled_title")
						body := c.translator.Translate(lang, "booking_cancelled_body")
						data := map[string]string{
							"type":       "booking_cancelled",
							"title":      title,
							"body":       body,
							"ride_id":    ev.RideID,
							"booking_id": ev.BookingID,
						}
						_ = c.uc.SendSystemPushWithData(context.Background(), driverID, title, body, data)
					}
				}
			} else if msg.RoutingKey == "ride.cancelled" {
				var ev struct {
					RideID       string   `json:"ride_id"`
					DriverID     string   `json:"driver_id"`
					Reason       string   `json:"reason"`
					PassengerIDs []string `json:"passenger_ids"`
				}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					for _, pIDStr := range ev.PassengerIDs {
						if pID, err := uuid.Parse(pIDStr); err == nil {
							lang := c.uc.GetUserLocale(context.Background(), pID)
							title := c.translator.Translate(lang, "ride_cancelled_title")
							body := c.translator.Translate(lang, "ride_cancelled_body")
							data := map[string]string{
								"type":    "ride_cancelled",
								"title":   title,
								"body":    body,
								"ride_id": ev.RideID,
							}
							if pushErr := c.uc.SendSystemPushWithData(context.Background(), pID, title, body, data); pushErr != nil {
								slog.Error("Failed to send system push for ride.cancelled", "error", pushErr)
							}
						}
					}
				} else {
					slog.Error("Failed to unmarshal ride.cancelled event", "error", err)
				}
			} else if msg.RoutingKey == "ride.reminder" {
				var ev struct {
					Kind   string `json:"kind"`
					RideID string `json:"ride_id"`
					UserID string `json:"user_id"`
				}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					if uID, err := uuid.Parse(ev.UserID); err == nil {
						lang := c.uc.GetUserLocale(context.Background(), uID)
						defTitle := "Поездка через час"
						defBody := "Ваша поездка начнется через 1 час"
						if ev.Kind == "ride_reminder_15m" {
							defTitle = "Поездка через 15 минут"
							defBody = "Ваша поездка начнется через 15 минут"
						}
						title := c.translator.Translate(lang, ev.Kind+"_title")
						if title == ev.Kind+"_title" {
							title = defTitle
						}
						body := c.translator.Translate(lang, ev.Kind+"_body")
						if body == ev.Kind+"_body" {
							body = defBody
						}

						data := map[string]string{
							"type":    ev.Kind,
							"title":   title,
							"body":    body,
							"ride_id": ev.RideID,
						}
						_ = c.uc.SendSystemPushWithData(context.Background(), uID, title, body, data)
					}
				}
			} else if msg.RoutingKey == "ride.completed" {
				var ev struct {
					RideID   string `json:"ride_id"`
					DriverID string `json:"driver_id"`
				}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					// Broadcast to all passengers
					// Since we don't have the passenger list in the payload, we might need it.
					// But for now, we just log or try to send system push to the driver.
				}
			} else if msg.RoutingKey == "ride.route_match" {
				var ev struct {
					RideID      string `json:"ride_id"`
					PassengerID string `json:"passenger_id"`
					DriverID    string `json:"driver_id"`
				}
				if err := json.Unmarshal(msg.Body, &ev); err == nil {
					if pID, err := uuid.Parse(ev.PassengerID); err == nil {
						lang := c.uc.GetUserLocale(context.Background(), pID)
						title := c.translator.Translate(lang, "system_notification")
						if title == "system_notification" {
							title = "System Notification"
						}
						content := c.translator.Translate(lang, "route_match")
						if pushErr := c.uc.SendSystemPush(context.Background(), pID, title, content, "route_match"); pushErr != nil {
							slog.Error("Failed to send system push for route_match", "error", pushErr)
						}
					}
				} else {
					slog.Error("Failed to unmarshal route_match event", "error", err)
				}
			}

			msg.Ack(false)
		}

		slog.Info("Consumer channel closed for ride events, attempting to reconnect...")
		time.Sleep(5 * time.Second)
	}
}

type NotificationPayload struct {
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	EventType string `json:"event_type"`
}

func (c *Consumer) consumeSystemPushLoop() {
	for {
		msgs, err := c.client.Consume("notification.general.events")
		if err != nil {
			slog.Error("SystemPush Consumer failed to start: %v. Retrying in 5s...", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		slog.Info("Consumer started successfully for notification.general.events")

		for msg := range msgs {
			var payload NotificationPayload
			if err := json.Unmarshal(msg.Body, &payload); err == nil {
				if uID, parseErr := uuid.Parse(payload.UserID); parseErr == nil {
					// 1. WebSocket/FCM Push (it also saves to DB internally)
					if pushErr := c.uc.SendSystemPush(context.Background(), uID, payload.Title, payload.Message, payload.EventType); pushErr != nil {
						slog.Error("Failed to send system push", "error", pushErr)
					}
				} else {
					slog.Error("Invalid user_id in system push event", "user_id", payload.UserID)
				}
			} else {
				// Try to unmarshal as array of NotificationPayload (for send_batch)
				var batchPayloads []NotificationPayload
				if errBatch := json.Unmarshal(msg.Body, &batchPayloads); errBatch == nil {
					for _, p := range batchPayloads {
						if uID, parseErr := uuid.Parse(p.UserID); parseErr == nil {
							_ = c.uc.SendSystemPush(context.Background(), uID, p.Title, p.Message, p.EventType)
						}
					}
				} else {
					slog.Error("Failed to unmarshal system push event", "error", err)
				}
			}
			msg.Ack(false)
		}

		slog.Info("Consumer channel closed for general events, attempting to reconnect...")
		time.Sleep(5 * time.Second)
	}
}

func (c *Consumer) consumeParcelLoop() {
	for {
		msgs, err := c.client.Consume("notification.parcel.events")
		if err != nil {
			slog.Error("Consumer failed to start for parcel events. Retrying in 5s...", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}

		slog.Info("Consumer started successfully for notification.parcel.events")

		for msg := range msgs {
			var payload map[string]interface{}
			if err := json.Unmarshal(msg.Body, &payload); err == nil {
				routingKey := msg.RoutingKey
				if routingKey == "parcel.response" {
					if senderIDStr, ok := payload["sender_id"].(string); ok {
						if senderID, err := uuid.Parse(senderIDStr); err == nil {
							lang := c.uc.GetUserLocale(context.Background(), senderID)
							title := c.translator.Translate(lang, "parcel_response_title")
							content := c.translator.Translate(lang, "parcel_response_body")

							_ = c.uc.SendSystemPush(context.Background(), senderID, title, content, "parcel_response")
						}
					}
				}
			} else {
				slog.Error("Failed to unmarshal parcel event payload", "error", err)
			}

			msg.Ack(false)
		}
	}
}
