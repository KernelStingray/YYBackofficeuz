package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/user/domain"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	"github.com/hextec/yolyolakay-backend/pkg/sanitize"
)

type UserUseCase struct {
	userRepo domain.UserRepository
	carRepo  domain.CarRepository
	rmq      *rabbitmq.Client
}

func NewUserUseCase(
	ur domain.UserRepository,
	cr domain.CarRepository,
	rmq *rabbitmq.Client,
) *UserUseCase {
	return &UserUseCase{
		userRepo: ur,
		carRepo:  cr,
		rmq:      rmq,
	}
}

func (u *UserUseCase) GetAdminStats(ctx context.Context) (*domain.AdminStats, error) {
	return u.userRepo.GetAdminStats(ctx)
}

func (u *UserUseCase) SearchAdminUsers(ctx context.Context, page, limit int32, search, role, status, kycStatus string) ([]*domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.userRepo.SearchAdminUsers(ctx, page, limit, search, role, status, kycStatus)
}

func (u *UserUseCase) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	err := u.userRepo.Delete(ctx, userID)
	if err != nil {
		return err
	}

	eventPayload, _ := json.Marshal(map[string]interface{}{
		"user_id": userID.String(),
	})

	if u.rmq != nil {
		if err := u.rmq.Publish(ctx, "yolyolakay.events", "user.deleted", eventPayload); err != nil {
			slog.Error("Failed to publish", "error", err)
		}
	}

	return nil
}

// RegisterPassenger registers a new user with passenger role
func (u *UserUseCase) RegisterPassenger(ctx context.Context, firstName, lastName, phone string, gender *string) (*domain.User, error) {
	if firstName == "" || lastName == "" || phone == "" {
		return nil, errors.New("first name, last name, and phone are required")
	}

	// XSS Sanitization
	sanitizedFirstName := sanitize.String(firstName)
	sanitizedLastName := sanitize.String(lastName)

	// Check if user already exists
	existing, _ := u.userRepo.GetByPhone(ctx, phone)
	if existing != nil {
		return nil, errors.New("user with this phone number already exists")
	}

	return u.userRepo.Create(ctx, sanitizedFirstName, sanitizedLastName, phone, domain.RolePassenger, gender)
}

// UpgradeToDriver upgrades a passenger to a driver by adding a car, updating driver photo (if provided) and setting KYC to pending
func (u *UserUseCase) UpgradeToDriver(ctx context.Context, userID uuid.UUID, car *domain.Car, avatarURL *string) error {
	// 1. Check user exists
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// 2. Update driver photo/avatar if provided
	if avatarURL != nil && *avatarURL != "" {
		_, err = u.userRepo.UpdateProfile(ctx, userID, user.FirstName, user.LastName, user.Gender, &user.Locale, avatarURL)
		if err != nil {
			return fmt.Errorf("failed to update driver avatar: %w", err)
		}
	}

	// 3. Create Car (with XSS Sanitization)
	car.UserID = userID
	car.GovNumber = sanitize.String(car.GovNumber)
	car.Brand = sanitize.String(car.Brand)
	car.Model = sanitize.String(car.Model)
	_, err = u.carRepo.Create(ctx, car)
	if err != nil {
		return err
	}

	// 4. Set kyc_status to 'pending' (or 'approved' for test accounts)
	testPhone := os.Getenv("TEST_PHONE")
	if testPhone == "" {
		testPhone = "+998991234567"
	}
	if user.PhoneNumber == testPhone || user.PhoneNumber == "+998991234567" || strings.HasPrefix(user.PhoneNumber, "+99800") {
		_, _ = u.userRepo.UpdateKYCStatusAndRole(ctx, userID, "approved", domain.RoleDriver, nil)
		_ = u.userRepo.UpdateCarsModerationStatusByUserID(ctx, userID, "approved")
	} else {
		_, err = u.userRepo.UpdateKYC(ctx, userID, "pending", user.FirstName, user.LastName)
		if err != nil {
			return err
		}
	}

	return nil
}

// ApproveDriverKYC approves the driver registration application
func (u *UserUseCase) ApproveDriverKYC(ctx context.Context, userID uuid.UUID) error {
	updatedUser, err := u.userRepo.UpdateKYCStatusAndRole(ctx, userID, "approved", domain.RoleDriver, nil)
	if err != nil {
		return err
	}

	// Update the moderation status of the driver's cars
	if err := u.userRepo.UpdateCarsModerationStatusByUserID(ctx, userID, "approved"); err != nil {
		slog.Error("Failed to update cars moderation status", "user_id", userID, "error", err)
	}

	// Publish user.verified event to RabbitMQ
	if u.rmq != nil {
		eventPayload, _ := json.Marshal(map[string]interface{}{
			"user_id":          updatedUser.ID.String(),
			"role":             string(updatedUser.Role),
			"first_name":       updatedUser.FirstName,
			"last_name":        updatedUser.LastName,
		})
		if err := u.rmq.Publish(ctx, "yolyolakay.events", "user.verified", eventPayload); err != nil {
			slog.Error("Failed to publish user.verified", "error", err)
		}

		title, msg := getCarApprovedNotification(updatedUser.Locale)
		carApprovedPayload, _ := json.Marshal(map[string]string{
			"user_id":    userID.String(),
			"title":      title,
			"message":    msg,
			"event_type": "car_approved",
		})
		if err := u.rmq.Publish(ctx, "notification_exchange", "notification.car", carApprovedPayload); err != nil {
			slog.Error("Failed to publish car_approved notification", "error", err)
		}
	}
	return nil
}

func getCarApprovedNotification(locale string) (string, string) {
	switch locale {
	case "ru":
		return "Автомобиль одобрен", "Ваш автомобиль и профиль водителя успешно прошли модерацию!"
	case "en":
		return "Car approved", "Your car and driver profile have been successfully approved!"
	default: // "uz"
		return "Avtomobil tasdiqlandi", "Sizning avtomobilingiz va haydovchi profilingiz muvaffaqiyatli tasdiqlandi!"
	}
}

func getCarRejectedNotification(locale, reason string) (string, string) {
	switch locale {
	case "ru":
		title := "Заявка отклонена"
		msg := "Ваша заявка на модерацию автомобиля/профиля была отклонена."
		if reason != "" {
			msg += " Причина: " + reason
		}
		return title, msg
	case "en":
		title := "Application rejected"
		msg := "Your car/profile application has been rejected."
		if reason != "" {
			msg += " Reason: " + reason
		}
		return title, msg
	default: // "uz"
		title := "Ariza rad etildi"
		msg := "Sizning avtomobil/profil arizangiz rad etildi."
		if reason != "" {
			msg += " Sabab: " + reason
		}
		return title, msg
	}
}

// RejectDriverKYC rejects the driver registration application
func (u *UserUseCase) RejectDriverKYC(ctx context.Context, userID uuid.UUID, reason string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}
	if user.KycStatus != "pending" {
		return errors.New("kyc is not pending")
	}

	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}
	_, err = u.userRepo.UpdateKYCStatusAndRole(ctx, userID, "rejected", domain.RolePassenger, reasonPtr)
	if err != nil {
		return err
	}

	// Update all cars of this user to rejected
	err = u.userRepo.UpdateCarsModerationStatusByUserID(ctx, userID, "rejected")
	if err != nil {
		// Log the error but don't fail the entire process
		slog.Error("failed to update cars moderation status", "error", err, "user_id", userID)
	}

	slog.Info("driver KYC rejected", "user_id", userID, "reason", reason)

	if u.rmq != nil {
		title, msg := getCarRejectedNotification(user.Locale, reason)
		carRejectedPayload, _ := json.Marshal(map[string]string{
			"user_id":    userID.String(),
			"title":      title,
			"message":    msg,
			"event_type": "car_rejected",
		})
		if err := u.rmq.Publish(ctx, "notification_exchange", "notification.car", carRejectedPayload); err != nil {
			slog.Error("Failed to publish car_rejected notification", "error", err)
		}
	}

	return nil
}

// GetUserProfile returns user details and associated cars
func (u *UserUseCase) GetUserProfile(ctx context.Context, userID uuid.UUID) (*domain.User, []domain.Car, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	cars, err := u.carRepo.GetByUserID(ctx, userID)
	if err != nil {
		// If error occurs, just return empty list
		cars = []domain.Car{}
	}

	return user, cars, nil
}

// GetCarByID returns a car by its ID
func (u *UserUseCase) GetCarByID(ctx context.Context, id uuid.UUID) (*domain.Car, error) {
	return u.carRepo.GetByID(ctx, id)
}

// GetAllUsers returns all users for admin panel
func (u *UserUseCase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return u.userRepo.GetAllUsers(ctx)
}

// UpdateProfile updates the user's basic profile details
func (u *UserUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string, gender *string, locale *string, avatarURL *string) (*domain.User, error) {
	sanitizedFirstName := sanitize.String(firstName)
	sanitizedLastName := sanitize.String(lastName)

	// If first name or last name is omitted in a partial update, retain existing values from database
	if sanitizedFirstName == "" || sanitizedLastName == "" {
		existingUser, err := u.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if existingUser == nil {
			return nil, domain.ErrUserNotFound
		}
		if sanitizedFirstName == "" {
			sanitizedFirstName = existingUser.FirstName
		}
		if sanitizedLastName == "" {
			sanitizedLastName = existingUser.LastName
		}
	}

	return u.userRepo.UpdateProfile(ctx, userID, sanitizedFirstName, sanitizedLastName, gender, locale, avatarURL)
}

// CheckCarApproved checks if the given user has at least one approved car
func (u *UserUseCase) CheckCarApproved(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, _ := u.userRepo.GetByID(ctx, userID)
	if user != nil {
		testPhone := os.Getenv("TEST_PHONE")
		if testPhone == "" {
			testPhone = "+998991234567"
		}
		if user.PhoneNumber == testPhone || user.PhoneNumber == "+998991234567" || strings.HasPrefix(user.PhoneNumber, "+99800") {
			return true, nil
		}
	}

	cars, err := u.carRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, car := range cars {
		if car.ModerationStatus == "approved" {
			return true, nil
		}
	}

	return false, nil
}

// GetActiveCar returns the first approved car for a user
func (u *UserUseCase) GetActiveCar(ctx context.Context, userID uuid.UUID) (*domain.Car, error) {
	cars, err := u.carRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, car := range cars {
		if car.ModerationStatus == "approved" {
			return &car, nil
		}
	}
	return nil, errors.New("no approved car found")
}

// CreateCar creates a car for a user
func (u *UserUseCase) CreateCar(ctx context.Context, userID uuid.UUID, car *domain.Car) (*domain.Car, error) {
	car.UserID = userID
	car.GovNumber = sanitize.String(car.GovNumber)
	car.Brand = sanitize.String(car.Brand)
	car.Model = sanitize.String(car.Model)

	createdCar, err := u.carRepo.Create(ctx, car)
	if err != nil {
		return nil, err
	}

	// For test accounts, auto-approve the car and upgrade to driver immediately
	user, err := u.userRepo.GetByID(ctx, userID)
	if err == nil && user != nil {
		testPhone := os.Getenv("TEST_PHONE")
		if testPhone == "" {
			testPhone = "+998991234567"
		}
		if user.PhoneNumber == testPhone || user.PhoneNumber == "+998991234567" || strings.HasPrefix(user.PhoneNumber, "+99800") {
			_ = u.userRepo.UpdateCarsModerationStatusByUserID(ctx, userID, "approved")
			_, _ = u.userRepo.UpdateKYCStatusAndRole(ctx, userID, "approved", domain.RoleDriver, nil)
			createdCar.ModerationStatus = "approved"
		}
	}

	return createdCar, nil
}

// UpdateCar updates car details for the current user
func (u *UserUseCase) UpdateCar(ctx context.Context, userID, carID uuid.UUID, car *domain.Car) (*domain.Car, error) {
	existing, err := u.carRepo.GetByID(ctx, carID)
	if err != nil || existing == nil {
		return nil, errors.New("car not found")
	}
	if existing.UserID != userID {
		return nil, errors.New("unauthorized: you do not own this car")
	}

	car.ID = carID
	car.UserID = userID
	car.GovNumber = sanitize.String(car.GovNumber)
	car.Brand = sanitize.String(car.Brand)
	car.Model = sanitize.String(car.Model)

	return u.carRepo.Update(ctx, car)
}

// DeleteCar deletes a car belonging to the current user
func (u *UserUseCase) DeleteCar(ctx context.Context, userID, carID uuid.UUID) error {
	existing, err := u.carRepo.GetByID(ctx, carID)
	if err != nil || existing == nil {
		return errors.New("car not found")
	}
	if existing.UserID != userID {
		return errors.New("unauthorized: you do not own this car")
	}

	return u.carRepo.Delete(ctx, carID)
}

// GetCars returns all cars for a user
func (u *UserUseCase) GetCars(ctx context.Context, userID uuid.UUID) ([]domain.Car, error) {
	return u.carRepo.GetByUserID(ctx, userID)
}

func (u *UserUseCase) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	if status != "active" && status != "blocked" {
		return errors.New("invalid status, must be active or blocked")
	}
	return u.userRepo.UpdateUserStatus(ctx, userID, status)
}

func (u *UserUseCase) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return u.userRepo.BlockUser(ctx, blockerID, blockedID)
}

func (u *UserUseCase) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	return u.userRepo.UnblockUser(ctx, blockerID, blockedID)
}

func (u *UserUseCase) GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]domain.User, error) {
	blockedIDs, err := u.userRepo.GetBlockedUsers(ctx, blockerID)
	if err != nil {
		return nil, err
	}
	var users []domain.User
	for _, id := range blockedIDs {
		user, _, err := u.GetUserProfile(ctx, id)
		if err == nil && user != nil {
			users = append(users, *user)
		}
	}
	return users, nil
}

func (u *UserUseCase) CreateReport(ctx context.Context, reporterID, reportedID uuid.UUID, reason string) error {
	return u.userRepo.CreateReport(ctx, uuid.New(), reporterID, reportedID, reason)
}

func (u *UserUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return u.DeleteUser(ctx, userID)
}
