package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/notification/domain"
	"github.com/hextec/yolyolakay-backend/internal/notification/repository/postgres"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresNotificationRepository struct {
	db      *pgxpool.Pool
	queries *postgres.Queries
}

func NewPostgresNotificationRepository(db *pgxpool.Pool) domain.NotificationRepository {
	return &postgresNotificationRepository{
		db:      db,
		queries: postgres.New(db),
	}
}

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidPtrToPg(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgToUUID(pgID pgtype.UUID) uuid.UUID {
	if !pgID.Valid {
		return uuid.Nil
	}
	return pgID.Bytes
}

func pgToUUIDPtr(pgID pgtype.UUID) *uuid.UUID {
	if !pgID.Valid {
		return nil
	}
	u := uuid.UUID(pgID.Bytes)
	return &u
}

func (r *postgresNotificationRepository) CreateMessage(ctx context.Context, msg *domain.Message) error {
	var rideID, replyToID pgtype.UUID
	if msg.RideID != nil {
		rideID = uuidToPg(*msg.RideID)
	}
	if msg.ReplyToID != nil {
		replyToID = uuidToPg(*msg.ReplyToID)
	}

	pgMsg, err := r.queries.CreateMessage(ctx, postgres.CreateMessageParams{
		ID:         uuidToPg(msg.ID),
		SenderID:   uuidToPg(msg.SenderID),
		ReceiverID: uuidToPg(msg.ReceiverID),
		RideID:     rideID,
		ReplyToID:  replyToID,
		Content:    msg.Content,
	})
	if err != nil {
		return err
	}

	msg.Status = domain.MessageStatus(pgMsg.Status.String)
	if pgMsg.CreatedAt.Valid && !pgMsg.CreatedAt.Time.IsZero() {
		msg.CreatedAt = pgMsg.CreatedAt.Time
	} else if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now().UTC()
	}
	if pgMsg.UpdatedAt.Valid && !pgMsg.UpdatedAt.Time.IsZero() {
		msg.UpdatedAt = pgMsg.UpdatedAt.Time
	} else if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = msg.CreatedAt
	}
	return nil
}

func (r *postgresNotificationRepository) GetChatHistory(ctx context.Context, userID1, userID2 uuid.UUID, rideID *uuid.UUID) ([]domain.Message, error) {
	var pgRideID pgtype.UUID
	if rideID != nil {
		pgRideID = uuidToPg(*rideID)
	}

	pgMsgs, err := r.queries.GetChatHistory(ctx, postgres.GetChatHistoryParams{
		SenderID:   uuidToPg(userID1),
		ReceiverID: uuidToPg(userID2),
		Column3:    pgRideID,
	})
	if err != nil {
		return nil, err
	}

	msgs := make([]domain.Message, len(pgMsgs))
	for i, pgMsg := range pgMsgs {
		var rID, replyID *uuid.UUID
		if pgMsg.RideID.Valid {
			id := pgToUUID(pgMsg.RideID)
			rID = &id
		}
		if pgMsg.ReplyToID.Valid {
			id := pgToUUID(pgMsg.ReplyToID)
			replyID = &id
		}
		createdAt := pgMsg.CreatedAt.Time
		if !pgMsg.CreatedAt.Valid || createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		updatedAt := pgMsg.UpdatedAt.Time
		if !pgMsg.UpdatedAt.Valid || updatedAt.IsZero() {
			updatedAt = createdAt
		}

		msgs[i] = domain.Message{
			ID:         pgToUUID(pgMsg.ID),
			SenderID:   pgToUUID(pgMsg.SenderID),
			ReceiverID: pgToUUID(pgMsg.ReceiverID),
			RideID:     rID,
			ReplyToID:  replyID,
			Content:    pgMsg.Content,
			Status:     domain.MessageStatus(pgMsg.Status.String),
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}
	}
	return msgs, nil
}

func (r *postgresNotificationRepository) MarkMessageAsRead(ctx context.Context, messageID, receiverID uuid.UUID) error {
	return r.queries.MarkMessageAsRead(ctx, postgres.MarkMessageAsReadParams{
		ID:         uuidToPg(messageID),
		ReceiverID: uuidToPg(receiverID),
	})
}

func (r *postgresNotificationRepository) GetMessageByID(ctx context.Context, messageID uuid.UUID) (*domain.Message, error) {
	query := `SELECT id, sender_id, receiver_id, ride_id, reply_to_id, content, status, created_at, updated_at FROM chats_messages WHERE id = $1 AND deleted_at IS NULL`
	var msgID, senderID, receiverID pgtype.UUID
	var rideID, replyToID pgtype.UUID
	var content, status string
	var createdAt, updatedAt pgtype.Timestamptz

	err := r.db.QueryRow(ctx, query, uuidToPg(messageID)).Scan(&msgID, &senderID, &receiverID, &rideID, &replyToID, &content, &status, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	var rID, replyID *uuid.UUID
	if rideID.Valid {
		id := pgToUUID(rideID)
		rID = &id
	}
	if replyToID.Valid {
		id := pgToUUID(replyToID)
		replyID = &id
	}

	createdAtTime := createdAt.Time
	if !createdAt.Valid || createdAtTime.IsZero() {
		createdAtTime = time.Now().UTC()
	}
	updatedAtTime := updatedAt.Time
	if !updatedAt.Valid || updatedAtTime.IsZero() {
		updatedAtTime = createdAtTime
	}

	return &domain.Message{
		ID:         pgToUUID(msgID),
		SenderID:   pgToUUID(senderID),
		ReceiverID: pgToUUID(receiverID),
		RideID:     rID,
		ReplyToID:  replyID,
		Content:    content,
		Status:     domain.MessageStatus(status),
		CreatedAt:  createdAtTime,
		UpdatedAt:  updatedAtTime,
	}, nil
}

func (r *postgresNotificationRepository) SaveSystemNotification(ctx context.Context, notif *domain.NotificationHistoryItem) error {
	var dataBytes []byte
	if notif.Data != nil {
		dataBytes, _ = json.Marshal(notif.Data)
	}

	pgNotif, err := r.queries.CreateNotification(ctx, postgres.CreateNotificationParams{
		ID:        uuidToPg(notif.ID),
		UserID:    uuidToPg(notif.UserID),
		Title:     notif.Title,
		Body:      notif.Body,
		Type:      notif.Type,
		RideID:    uuidPtrToPg(notif.RideID),
		BookingID: uuidPtrToPg(notif.BookingID),
		ParcelID:  uuidPtrToPg(notif.ParcelID),
		Data:      dataBytes,
	})
	if err != nil {
		return err
	}
	notif.IsRead = pgNotif.IsRead.Bool
	notif.CreatedAt = pgNotif.CreatedAt.Time
	return nil
}

func (r *postgresNotificationRepository) GetNotificationHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.NotificationHistoryItem, error) {
	pgNotifs, err := r.queries.GetNotificationsByUser(ctx, postgres.GetNotificationsByUserParams{
		UserID: uuidToPg(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	notifs := make([]domain.NotificationHistoryItem, len(pgNotifs))
	for i, pgNotif := range pgNotifs {
		var dataMap map[string]string
		if len(pgNotif.Data) > 0 {
			_ = json.Unmarshal(pgNotif.Data, &dataMap)
		}

		notifs[i] = domain.NotificationHistoryItem{
			ID:        pgToUUID(pgNotif.ID),
			UserID:    pgToUUID(pgNotif.UserID),
			Title:     pgNotif.Title,
			Body:      pgNotif.Body,
			Type:      pgNotif.Type,
			IsRead:    pgNotif.IsRead.Bool,
			RideID:    pgToUUIDPtr(pgNotif.RideID),
			BookingID: pgToUUIDPtr(pgNotif.BookingID),
			ParcelID:  pgToUUIDPtr(pgNotif.ParcelID),
			Data:      dataMap,
			CreatedAt: pgNotif.CreatedAt.Time,
		}
	}
	return notifs, nil
}

func (r *postgresNotificationRepository) MarkNotificationAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	return r.queries.MarkNotificationAsRead(ctx, postgres.MarkNotificationAsReadParams{
		ID:     uuidToPg(notificationID),
		UserID: uuidToPg(userID),
	})
}

func (r *postgresNotificationRepository) GetUserChats(ctx context.Context, userID uuid.UUID) ([]domain.ChatListItem, error) {
	query := `
		WITH partner_messages AS (
			SELECT 
				CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END AS partner_id,
				id, sender_id, receiver_id, ride_id, content, status, created_at, updated_at,
				ROW_NUMBER() OVER (
					PARTITION BY CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END 
					ORDER BY created_at DESC
				) as rn
			FROM chats_messages
			WHERE (sender_id = $1 OR receiver_id = $1) AND deleted_at IS NULL
		)
		SELECT partner_id, id, sender_id, receiver_id, ride_id, content, status, created_at, updated_at,
			(SELECT COUNT(*)::int FROM chats_messages WHERE receiver_id = $1 AND sender_id = partner_messages.partner_id AND status != 'read' AND deleted_at IS NULL) AS unread_count
		FROM partner_messages
		WHERE rn = 1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, uuidToPg(userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ChatListItem
	for rows.Next() {
		var partnerID, msgID, senderID, receiverID pgtype.UUID
		var rideID pgtype.UUID
		var content, status string
		var createdAt, updatedAt pgtype.Timestamptz
		var unreadCount int32

		if err := rows.Scan(&partnerID, &msgID, &senderID, &receiverID, &rideID, &content, &status, &createdAt, &updatedAt, &unreadCount); err != nil {
			return nil, err
		}

		var rID *uuid.UUID
		if rideID.Valid {
			id := pgToUUID(rideID)
			rID = &id
		}

		msg := &domain.Message{
			ID:         pgToUUID(msgID),
			SenderID:   pgToUUID(senderID),
			ReceiverID: pgToUUID(receiverID),
			RideID:     rID,
			Content:    content,
			Status:     domain.MessageStatus(status),
			CreatedAt:  createdAt.Time,
			UpdatedAt:  updatedAt.Time,
		}

		items = append(items, domain.ChatListItem{
			PartnerID:   pgToUUID(partnerID),
			LastMessage: msg,
			UnreadCount: unreadCount,
		})
	}

	return items, nil
}
