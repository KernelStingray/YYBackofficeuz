package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/user/domain"
	"github.com/hextec/yolyolakay-backend/internal/user/usecase"
	mocks "github.com/hextec/yolyolakay-backend/pkg/mocks/user"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserUseCase_UpgradeToDriver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	carRepo := mocks.NewMockCarRepository(ctrl)
	var rmq *rabbitmq.Client // nil for testing

	uc := usecase.NewUserUseCase(userRepo, carRepo, rmq)
	ctx := context.Background()

	t.Run("success with avatar and car photos", func(t *testing.T) {
		userID := uuid.New()
		techFront := "https://api.yolyolakay.com/media/tp_front.jpg"
		techBack := "https://api.yolyolakay.com/media/tp_back.jpg"
		dlFront := "https://api.yolyolakay.com/media/dl_front.jpg"
		dlBack := "https://api.yolyolakay.com/media/dl_back.jpg"
		carFront := "https://api.yolyolakay.com/media/car_front.jpg"
		carBack := "https://api.yolyolakay.com/media/car_back.jpg"
		carSide := "https://api.yolyolakay.com/media/car_side.jpg"
		avatar := "https://api.yolyolakay.com/media/driver_avatar.jpg"

		car := &domain.Car{
			GovNumber:          "01A123AA",
			Brand:              "Chevrolet",
			Model:              "Malibu",
			TechPassportFront:  &techFront,
			TechPassportBack:   &techBack,
			DriverLicenseFront: &dlFront,
			DriverLicenseBack:  &dlBack,
			CarFrontPhoto:      &carFront,
			CarBackPhoto:       &carBack,
			CarSidePhoto:       &carSide,
		}

		user := &domain.User{
			ID:        userID,
			FirstName: "Alisher",
			LastName:  "Navoiy",
			Locale:    "ru",
		}

		userRepo.EXPECT().GetByID(ctx, userID).Return(user, nil).Times(1)
		userRepo.EXPECT().UpdateProfile(ctx, userID, "Alisher", "Navoiy", user.Gender, &user.Locale, &avatar).Return(user, nil).Times(1)

		carRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, c *domain.Car) (*domain.Car, error) {
			assert.Equal(t, userID, c.UserID)
			assert.Equal(t, "01A123AA", c.GovNumber)
			assert.Equal(t, &techFront, c.TechPassportFront)
			assert.Equal(t, &carFront, c.CarFrontPhoto)
			return c, nil
		}).Times(1)

		updatedUser := &domain.User{
			ID:        userID,
			Role:      domain.RolePassenger,
			KycStatus: "pending",
			FirstName: "Alisher",
			LastName:  "Navoiy",
		}
		userRepo.EXPECT().UpdateKYC(ctx, userID, "pending", "Alisher", "Navoiy").Return(updatedUser, nil).Times(1)

		err := uc.UpgradeToDriver(ctx, userID, car, &avatar)
		assert.NoError(t, err)
	})
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	carRepo := mocks.NewMockCarRepository(ctrl)
	var rmq *rabbitmq.Client

	uc := usecase.NewUserUseCase(userRepo, carRepo, rmq)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()

		userRepo.EXPECT().Delete(ctx, userID).Return(nil).Times(1)

		err := uc.DeleteUser(ctx, userID)
		assert.NoError(t, err)
	})
}

func TestUserUseCase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	carRepo := mocks.NewMockCarRepository(ctrl)
	var rmq *rabbitmq.Client

	uc := usecase.NewUserUseCase(userRepo, carRepo, rmq)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("partial update - only locale", func(t *testing.T) {
		existing := &domain.User{
			ID:        userID,
			FirstName: "Aziz",
			LastName:  "Karimov",
			Locale:    "ru",
		}
		newLocale := "uz"

		userRepo.EXPECT().GetByID(ctx, userID).Return(existing, nil).Times(1)
		userRepo.EXPECT().UpdateProfile(ctx, userID, "Aziz", "Karimov", nil, &newLocale, nil).Return(&domain.User{
			ID:        userID,
			FirstName: "Aziz",
			LastName:  "Karimov",
			Locale:    newLocale,
		}, nil).Times(1)

		res, err := uc.UpdateProfile(ctx, userID, "", "", nil, &newLocale, nil)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "uz", res.Locale)
		assert.Equal(t, "Aziz", res.FirstName)
	})

	t.Run("full update", func(t *testing.T) {
		newLocale := "uz"
		gender := "male"
		avatar := "https://example.com/avatar.png"

		userRepo.EXPECT().UpdateProfile(ctx, userID, "Bobur", "Mirzo", &gender, &newLocale, &avatar).Return(&domain.User{
			ID:        userID,
			FirstName: "Bobur",
			LastName:  "Mirzo",
			Locale:    newLocale,
		}, nil).Times(1)

		res, err := uc.UpdateProfile(ctx, userID, "Bobur", "Mirzo", &gender, &newLocale, &avatar)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Bobur", res.FirstName)
	})
}

func TestUserUseCase_ApproveDriverKYC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	carRepo := mocks.NewMockCarRepository(ctrl)
	uc := usecase.NewUserUseCase(userRepo, carRepo, nil)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success approve", func(t *testing.T) {
		approvedUser := &domain.User{
			ID:        userID,
			Role:      domain.RoleDriver,
			KycStatus: "approved",
			FirstName: "Anvar",
			LastName:  "Karimov",
		}
		userRepo.EXPECT().UpdateKYCStatusAndRole(ctx, userID, "approved", domain.RoleDriver, (*string)(nil)).Return(approvedUser, nil).Times(1)
		userRepo.EXPECT().UpdateCarsModerationStatusByUserID(ctx, userID, "approved").Return(nil).Times(1)

		err := uc.ApproveDriverKYC(ctx, userID)
		assert.NoError(t, err)
	})
}

func TestUserUseCase_RejectDriverKYC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	carRepo := mocks.NewMockCarRepository(ctrl)
	uc := usecase.NewUserUseCase(userRepo, carRepo, nil)
	ctx := context.Background()
	userID := uuid.New()

	t.Run("success reject with reason", func(t *testing.T) {
		pendingUser := &domain.User{
			ID:        userID,
			Role:      domain.RolePassenger,
			KycStatus: "pending",
		}
		reason := "Плохое качество фото водительского удостоверения"
		rejectedUser := &domain.User{
			ID:                 userID,
			Role:               domain.RolePassenger,
			KycStatus:          "rejected",
			KycRejectionReason: &reason,
		}

		userRepo.EXPECT().GetByID(ctx, userID).Return(pendingUser, nil).Times(1)
		userRepo.EXPECT().UpdateKYCStatusAndRole(ctx, userID, "rejected", domain.RolePassenger, &reason).Return(rejectedUser, nil).Times(1)
		userRepo.EXPECT().UpdateCarsModerationStatusByUserID(ctx, userID, "rejected").Return(nil).Times(1)

		err := uc.RejectDriverKYC(ctx, userID, reason)
		assert.NoError(t, err)
	})
}

