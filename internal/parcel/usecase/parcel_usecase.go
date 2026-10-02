package usecase

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/parcel/domain"
	billingPb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/billing"
)

type parcelUseCase struct {
	repo          domain.ParcelRepository
	billingClient billingPb.BillingServiceClient
	publisher     domain.EventPublisher
}

func NewParcelUseCase(repo domain.ParcelRepository, billingClient billingPb.BillingServiceClient, publisher domain.EventPublisher) domain.ParcelUseCase {
	return &parcelUseCase{
		repo:          repo,
		billingClient: billingClient,
		publisher:     publisher,
	}
}

func (u *parcelUseCase) GetAdminStats(ctx context.Context) (*domain.AdminStats, error) {
	return u.repo.GetAdminStats(ctx)
}

func (u *parcelUseCase) GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserParcelStats, error) {
	return u.repo.GetUserStats(ctx, userID)
}

func getEncryptionKey() []byte {
	key := os.Getenv("PARCEL_ENCRYPTION_KEY")
	if len(key) == 32 {
		return []byte(key)
	}
	if len(key) > 0 {
		h := sha256.Sum256([]byte(key))
		return h[:]
	}
	secret := os.Getenv("INTERNAL_API_SECRET")
	if secret != "" {
		h := sha256.Sum256([]byte("parcel-secret:" + secret))
		return h[:]
	}
	return []byte("0123456789abcdef0123456789abcdef")
}

func encryptPhone(phone string) string {
	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return ""
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(phone), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func decryptPhone(encryptedBase64 string) string {
	data, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return ""
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return decryptPhoneLegacy(data)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return decryptPhoneLegacy(data)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return decryptPhoneLegacy(data)
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return decryptPhoneLegacy(data)
	}

	return string(plaintext)
}

func decryptPhoneLegacy(decoded []byte) string {
	key := "blind-logistics-secret-key"
	out := make([]byte, len(decoded))
	for i := 0; i < len(decoded); i++ {
		out[i] = decoded[i] ^ key[i%len(key)]
	}
	return string(out)
}

func (u *parcelUseCase) CreateParcel(ctx context.Context, senderID uuid.UUID, origLon, origLat, destLon, destLat float64, receiverPhone, receiverName, dropoffType string, price float64, originAddress, destinationAddress, description, size string, deliverBy *time.Time) (*domain.Parcel, error) {
	encryptedPhone := encryptPhone(receiverPhone)

	if dropoffType == "" {
		dropoffType = "highway"
	}

	p := &domain.Parcel{
		ID:                     uuid.New(),
		SenderID:               senderID,
		OriginLon:              origLon,
		OriginLat:              origLat,
		DestinationLon:         destLon,
		DestinationLat:         destLat,
		ReceiverPhoneEncrypted: encryptedPhone,
		ReceiverName:           receiverName,
		DropoffType:            dropoffType,
		Price:                  price,
		Status:                 domain.ParcelStatusPending,
		OriginAddress:          originAddress,
		DestinationAddress:     destinationAddress,
		Description:            description,
		Size:                   size,
		DeliverBy:              deliverBy,
	}

	err := u.repo.CreateParcel(ctx, p)
	if err != nil {
		return nil, err
	}

	// For the sender, we can decrypt it back so they see it
	p.DecryptedReceiverPhone = receiverPhone
	return p, nil
}

func (u *parcelUseCase) SearchParcels(ctx context.Context, driverID uuid.UUID, params domain.ParcelSearchParams) ([]domain.Parcel, error) {
	// Drivers can search. But they only see encrypted phones (or no phone at all).
	parcels, err := u.repo.SearchParcels(ctx, params)
	if err != nil {
		return nil, err
	}

	// We ensure DecryptedReceiverPhone is empty, so they can't call before accepting.
	for i := range parcels {
		parcels[i].DecryptedReceiverPhone = "" // Explicitly hidden
	}

	return parcels, nil
}

func (u *parcelUseCase) AcceptParcel(ctx context.Context, parcelID uuid.UUID, driverID uuid.UUID) (*domain.Parcel, error) {
	// 1. Check parcel limit via Billing Service gRPC
	resp, err := u.billingClient.DeductParcelLimit(ctx, &billingPb.DeductParcelLimitRequest{
		DriverId: driverID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check parcel limit: %w", err)
	}

	if !resp.Success {
		return nil, errors.New("access denied: insufficient parcel quota")
	}

	// 2. Fetch Parcel
	p, err := u.repo.GetParcelByID(ctx, parcelID)
	if err != nil {
		return nil, err
	}

	if p.Status != domain.ParcelStatusPending {
		return nil, errors.New("parcel is no longer pending")
	}

	// 3. Update status
	err = u.repo.UpdateParcelStatus(ctx, parcelID, domain.ParcelStatusAccepted, &driverID)
	if err != nil {
		return nil, err
	}

	// 4. Decrypt phone ONLY after successful acceptance
	p.Status = domain.ParcelStatusAccepted
	p.DriverID = &driverID
	p.DecryptedReceiverPhone = decryptPhone(p.ReceiverPhoneEncrypted)

	return p, nil
}

func (u *parcelUseCase) GetAllParcels(ctx context.Context) ([]domain.Parcel, error) {
	parcels, err := u.repo.GetAllParcels(ctx)
	if err != nil {
		return nil, err
	}

	// For admin panel, we can decrypt the phone numbers
	for i := range parcels {
		parcels[i].DecryptedReceiverPhone = decryptPhone(parcels[i].ReceiverPhoneEncrypted)
	}

	return parcels, nil
}

func (u *parcelUseCase) GetMyParcels(ctx context.Context, userID uuid.UUID, role string) ([]domain.Parcel, error) {
	var parcels []domain.Parcel
	var err error
	if role == "driver" {
		parcels, err = u.repo.GetParcelsByDriverID(ctx, userID)
	} else {
		parcels, err = u.repo.GetParcelsBySenderID(ctx, userID)
	}

	if err != nil {
		return nil, err
	}

	for i := range parcels {
		parcels[i].DecryptedReceiverPhone = decryptPhone(parcels[i].ReceiverPhoneEncrypted)
	}
	return parcels, nil
}

func (u *parcelUseCase) RespondToParcel(ctx context.Context, parcelID uuid.UUID, driverID uuid.UUID, offeredPrice float64) (*domain.ParcelResponse, error) {
	parcel, err := u.repo.GetParcelByID(ctx, parcelID)
	if err != nil {
		return nil, fmt.Errorf("parcel not found: %w", err)
	}

	if parcel.SenderID == driverID {
		return nil, errors.New("cannot respond to your own parcel")
	}

	if parcel.Status != domain.ParcelStatusPending {
		return nil, errors.New("parcel is not in pending status")
	}

	resp := &domain.ParcelResponse{
		ID:           uuid.New(),
		ParcelID:     parcelID,
		DriverID:     driverID,
		Status:       domain.ParcelResponseStatusPending,
		OfferedPrice: offeredPrice,
	}

	err = u.repo.CreateParcelResponse(ctx, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to create parcel response: %w", err)
	}

	if u.publisher != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"parcel_id": parcelID.String(),
			"driver_id": driverID.String(),
			"sender_id": parcel.SenderID.String(),
			"status":    string(resp.Status),
		})
		if err := u.publisher.Publish(ctx, "yolyolakay.events", "parcel.response", payload); err != nil {
			slog.Error("Failed to publish parcel.response event", "error", err)
		}
	}

	return resp, nil
}

func (u *parcelUseCase) GetParcelResponses(ctx context.Context, parcelID uuid.UUID, senderID uuid.UUID) ([]domain.ParcelResponse, error) {
	parcel, err := u.repo.GetParcelByID(ctx, parcelID)
	if err != nil {
		return nil, fmt.Errorf("parcel not found: %w", err)
	}

	if parcel.SenderID != senderID {
		return nil, errors.New("unauthorized: you are not the sender of this parcel")
	}

	return u.repo.GetParcelResponsesByParcelID(ctx, parcelID)
}

func (u *parcelUseCase) GetParcelByID(ctx context.Context, parcelID uuid.UUID) (*domain.Parcel, error) {
	parcel, err := u.repo.GetParcelByID(ctx, parcelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parcel: %w", err)
	}
	return parcel, nil
}

func (u *parcelUseCase) AcceptParcelResponse(ctx context.Context, parcelID uuid.UUID, responseID uuid.UUID, senderID uuid.UUID) error {
	// 1. Get parcel and verify sender
	parcel, err := u.repo.GetParcelByID(ctx, parcelID)
	if err != nil {
		return fmt.Errorf("parcel not found: %w", err)
	}
	if parcel.SenderID != senderID {
		return errors.New("unauthorized: you are not the sender of this parcel")
	}

	responses, err := u.repo.GetParcelResponsesByParcelID(ctx, parcelID)
	if err != nil {
		return err
	}

	var targetResponse *domain.ParcelResponse
	for _, r := range responses {
		if r.ID == responseID {
			targetResponse = &r
			break
		}
	}

	if targetResponse == nil {
		return errors.New("response not found")
	}

	// Deduct driver parcel quota
	if u.billingClient != nil {
		resp, err := u.billingClient.DeductParcelLimit(ctx, &billingPb.DeductParcelLimitRequest{
			DriverId: targetResponse.DriverID.String(),
		})
		if err == nil && !resp.Success {
			return errors.New("access denied: driver has insufficient parcel quota")
		}
	}

	// Update parcel
	err = u.repo.UpdateParcelStatus(ctx, parcelID, domain.ParcelStatusAccepted, &targetResponse.DriverID)
	if err != nil {
		return fmt.Errorf("failed to update parcel status: %w", err)
	}

	// Update the accepted response status to accepted
	_ = u.repo.UpdateParcelResponseStatus(ctx, responseID, domain.ParcelResponseStatusAccepted)
	// Reject all other pending responses for this parcel
	_ = u.repo.RejectOtherParcelResponses(ctx, parcelID, responseID)

	return nil
}

func (u *parcelUseCase) RejectParcelResponse(ctx context.Context, parcelID uuid.UUID, responseID uuid.UUID, senderID uuid.UUID) error {
	parcel, err := u.repo.GetParcelByID(ctx, parcelID)
	if err != nil {
		return fmt.Errorf("parcel not found: %w", err)
	}
	if parcel.SenderID != senderID {
		return errors.New("unauthorized: you are not the sender of this parcel")
	}

	responses, err := u.repo.GetParcelResponsesByParcelID(ctx, parcelID)
	if err != nil {
		return err
	}

	var targetResponse *domain.ParcelResponse
	for _, r := range responses {
		if r.ID == responseID {
			targetResponse = &r
			break
		}
	}

	if targetResponse == nil {
		return errors.New("response not found for this parcel")
	}

	// Mark response as rejected
	_ = u.repo.UpdateParcelResponseStatus(ctx, responseID, domain.ParcelResponseStatusRejected)

	// If the rejected response was the accepted driver, reset parcel status back to pending
	if parcel.DriverID != nil && *parcel.DriverID == targetResponse.DriverID {
		if err := u.repo.UpdateParcelStatus(ctx, parcelID, domain.ParcelStatusPending, nil); err != nil {
			return err
		}
	}

	return nil
}


func (u *parcelUseCase) CancelParcel(ctx context.Context, parcelID uuid.UUID, senderID uuid.UUID) error {
	parcel, err := u.repo.GetParcelByID(ctx, parcelID)
	if err != nil {
		return fmt.Errorf("parcel not found: %w", err)
	}

	if parcel.SenderID != senderID {
		return errors.New("unauthorized: you are not the sender of this parcel")
	}

	if parcel.Status != domain.ParcelStatusPending && parcel.Status != domain.ParcelStatusAccepted {
		return errors.New("cannot cancel a parcel that is not pending or accepted")
	}

	return u.repo.UpdateParcelStatus(ctx, parcelID, domain.ParcelStatusCanceled, nil)
}

func (u *parcelUseCase) PickupParcel(ctx context.Context, parcelID uuid.UUID, driverID uuid.UUID) error {
	parcel, err := u.repo.GetParcelByID(ctx, parcelID)
	if err != nil {
		return fmt.Errorf("parcel not found: %w", err)
	}

	if parcel.DriverID == nil || *parcel.DriverID != driverID {
		return errors.New("only the assigned driver can pick up this parcel")
	}

	if parcel.Status != domain.ParcelStatusAccepted {
		return errors.New("parcel must be in accepted status to be picked up")
	}

	return u.repo.UpdateParcelStatus(ctx, parcelID, domain.ParcelStatusPickedUp, &driverID)
}

func (u *parcelUseCase) DeliverParcel(ctx context.Context, parcelID uuid.UUID, userID uuid.UUID) error {
	parcel, err := u.repo.GetParcelByID(ctx, parcelID)
	if err != nil {
		return fmt.Errorf("parcel not found: %w", err)
	}

	isDriver := parcel.DriverID != nil && *parcel.DriverID == userID
	isSender := parcel.SenderID == userID

	if !isDriver && !isSender {
		return errors.New("only the assigned driver or the parcel sender can confirm delivery")
	}

	if parcel.Status != domain.ParcelStatusPickedUp && parcel.Status != domain.ParcelStatusAccepted {
		return errors.New("parcel must be in picked up status to be delivered")
	}

	return u.repo.UpdateParcelStatus(ctx, parcelID, domain.ParcelStatusDelivered, parcel.DriverID)
}
