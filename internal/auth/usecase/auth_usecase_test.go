package usecase

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	authMocks "github.com/hextec/yolyolakay-backend/pkg/mocks/auth"
	userDomain "github.com/hextec/yolyolakay-backend/internal/user/domain"
	userMocks "github.com/hextec/yolyolakay-backend/pkg/mocks/user"
	"github.com/hextec/yolyolakay-backend/pkg/jwt"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupAuthUseCaseTest(t *testing.T) (*AuthUseCase, *authMocks.MockOTPRepository, *userMocks.MockUserRepository, *authMocks.MockSessionRepository) {
	ctrl := gomock.NewController(t)

	otpRepo := authMocks.NewMockOTPRepository(ctrl)
	sessionRepo := authMocks.NewMockSessionRepository(ctrl)
	smsProvider := authMocks.NewMockSMSProvider(ctrl)
	userRepo := userMocks.NewMockUserRepository(ctrl)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwtManager := jwt.NewJWTManager(privateKey, &privateKey.PublicKey)

	var rmqClient *rabbitmq.Client = nil

	uc := NewAuthUseCase(
		otpRepo,
		sessionRepo,
		smsProvider,
		userRepo,
		jwtManager,
		rmqClient,
	)

	return uc, otpRepo, userRepo, sessionRepo
}

func TestVerifyOTP_Success(t *testing.T) {
	uc, otpRepo, userRepo, sessionRepo := setupAuthUseCaseTest(t)
	ctx := context.Background()

	phone := "+998901234567"
	code := "12345"
	userID := uuid.New()

	otpRepo.EXPECT().IsPhoneLocked(ctx, phone).Return(false, nil)
	otpRepo.EXPECT().VerifyOTP(ctx, phone, code).Return(true, nil)
	otpRepo.EXPECT().ResetFailedAttempts(ctx, phone).Return(nil)

	user := &userDomain.User{
		ID:          userID,
		PhoneNumber: phone,
		Role:        userDomain.RolePassenger,
	}
	userRepo.EXPECT().GetByPhone(ctx, phone).Return(user, nil)
	userRepo.EXPECT().AddTrustedDevice(ctx, userID, "device123").Return(nil).AnyTimes()
	sessionRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)

	tokens, err := uc.VerifyOTP(ctx, phone, code, "127.0.0.1", "device123", "UserAgent", "iOS")

	require.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.False(t, tokens.IsNewUser)
}

func TestVerifyOTP_WithPassword(t *testing.T) {
	uc, otpRepo, userRepo, sessionRepo := setupAuthUseCaseTest(t)
	ctx := context.Background()

	phone := "+998901234567"
	code := "12345"
	password := "Secret123!"
	userID := uuid.New()

	otpRepo.EXPECT().IsPhoneLocked(ctx, phone).Return(false, nil)
	otpRepo.EXPECT().VerifyOTP(ctx, phone, code).Return(true, nil)
	otpRepo.EXPECT().ResetFailedAttempts(ctx, phone).Return(nil)

	user := &userDomain.User{
		ID:          userID,
		PhoneNumber: phone,
		Role:        userDomain.RolePassenger,
	}
	userRepo.EXPECT().GetByPhone(ctx, phone).Return(user, nil)
	userRepo.EXPECT().UpdatePassword(ctx, userID, gomock.Any()).Return(nil)
	userRepo.EXPECT().AddTrustedDevice(ctx, userID, "device123").Return(nil).AnyTimes()
	sessionRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)

	tokens, err := uc.VerifyOTP(ctx, phone, code, "127.0.0.1", "device123", "UserAgent", "iOS", password)

	require.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.False(t, tokens.IsNewUser)
}

func TestVerifyOTP_PhoneLocked(t *testing.T) {
	uc, otpRepo, _, _ := setupAuthUseCaseTest(t)
	ctx := context.Background()

	phone := "+998901234567"
	code := "12345"

	otpRepo.EXPECT().IsPhoneLocked(ctx, phone).Return(true, nil)

	tokens, err := uc.VerifyOTP(ctx, phone, code, "127.0.0.1", "device123", "UserAgent", "iOS")

	require.Error(t, err)
	assert.Nil(t, tokens)
	assert.Equal(t, "account is temporarily locked due to too many failed attempts", err.Error())
}

func TestVerifyOTP_InvalidCode_BruteForceLock(t *testing.T) {
	uc, otpRepo, _, _ := setupAuthUseCaseTest(t)
	ctx := context.Background()

	phone := "+998901234567"
	code := "00000"

	otpRepo.EXPECT().IsPhoneLocked(ctx, phone).Return(false, nil)
	otpRepo.EXPECT().VerifyOTP(ctx, phone, code).Return(false, nil)

	otpRepo.EXPECT().IncrementFailedAttempts(ctx, phone).Return(5, nil)
	otpRepo.EXPECT().LockPhone(ctx, phone, 15*time.Minute).Return(nil)
	otpRepo.EXPECT().ResetFailedAttempts(ctx, phone).Return(nil)

	tokens, err := uc.VerifyOTP(ctx, phone, code, "127.0.0.1", "device123", "UserAgent", "iOS")

	require.Error(t, err)
	assert.Nil(t, tokens)
	assert.Equal(t, "account is temporarily locked due to too many failed attempts", err.Error())
}

func TestVerifyOTP_UserNotFound_CreatesNewUser(t *testing.T) {
	uc, otpRepo, userRepo, sessionRepo := setupAuthUseCaseTest(t)
	ctx := context.Background()

	phone := "+998901234567"
	code := "12345"
	newUserID := uuid.New()

	otpRepo.EXPECT().IsPhoneLocked(ctx, phone).Return(false, nil)
	otpRepo.EXPECT().VerifyOTP(ctx, phone, code).Return(true, nil)
	otpRepo.EXPECT().ResetFailedAttempts(ctx, phone).Return(nil)

	userRepo.EXPECT().GetByPhone(ctx, phone).Return(nil, errors.New("not found"))

	newUser := &userDomain.User{
		ID:          newUserID,
		PhoneNumber: phone,
		Role:        userDomain.RolePassenger,
	}
	userRepo.EXPECT().Create(ctx, "Новый", "Пользователь", phone, userDomain.RolePassenger, nil).Return(newUser, nil)
	userRepo.EXPECT().AddTrustedDevice(ctx, newUserID, "device123").Return(nil).AnyTimes()
	sessionRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)

	tokens, err := uc.VerifyOTP(ctx, phone, code, "127.0.0.1", "device123", "UserAgent", "iOS")

	require.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.True(t, tokens.IsNewUser)
}

func setupAuthWithSMSTest(t *testing.T) (*AuthUseCase, *authMocks.MockOTPRepository, *userMocks.MockUserRepository, *authMocks.MockSMSProvider) {
	ctrl := gomock.NewController(t)

	otpRepo := authMocks.NewMockOTPRepository(ctrl)
	sessionRepo := authMocks.NewMockSessionRepository(ctrl)
	smsProvider := authMocks.NewMockSMSProvider(ctrl)
	userRepo := userMocks.NewMockUserRepository(ctrl)

	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	jwtManager := jwt.NewJWTManager(privateKey, &privateKey.PublicKey)

	uc := NewAuthUseCase(
		otpRepo,
		sessionRepo,
		smsProvider,
		userRepo,
		jwtManager,
		nil,
	)

	return uc, otpRepo, userRepo, smsProvider
}

func TestSendOTP_Templates(t *testing.T) {
	tests := []struct {
		name            string
		reqType         string
		userExists      bool
		expectedSnippet string
	}{
		{
			name:            "Register uses ro'yxatdan o'tish template",
			reqType:         "register",
			userExists:      false,
			expectedSnippet: "Yolyolakay platformasiga ro‘yxatdan o‘tish uchun tasdiqlash kodi:",
		},
		{
			name:            "Login uses kirish template",
			reqType:         "login",
			userExists:      true,
			expectedSnippet: "Kodni hech kimga bermang! Yolyolakay mobil ilovasiga kirish uchun tasdiqlash kodi:",
		},
		{
			name:            "Forgot password uses parolini tiklash template",
			reqType:         "frgt_pswd",
			userExists:      true,
			expectedSnippet: "Yolyolakay platformasining parolini tiklash uchun tasdiqlash kodi:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, otpRepo, userRepo, smsProvider := setupAuthWithSMSTest(t)
			ctx := context.Background()
			phone := "+998919998877"
			ip := "127.0.0.1"

			if tt.userExists {
				userRepo.EXPECT().GetByPhone(gomock.Any(), phone).Return(&userDomain.User{ID: uuid.New(), PhoneNumber: phone}, nil)
			} else {
				userRepo.EXPECT().GetByPhone(gomock.Any(), phone).Return(nil, errors.New("not found"))
			}

			otpRepo.EXPECT().CheckIPRateLimit(gomock.Any(), ip).Return(0, nil)
			otpRepo.EXPECT().CheckRateLimit(gomock.Any(), phone).Return(0, nil)
			otpRepo.EXPECT().SaveOTP(gomock.Any(), phone, gomock.Any()).Return(nil)
			otpRepo.EXPECT().IncrementRateLimit(gomock.Any(), phone).Return(nil)
			otpRepo.EXPECT().IncrementIPRateLimit(gomock.Any(), ip).Return(nil)

			smsProvider.EXPECT().SendSMS(gomock.Any(), phone, gomock.Any()).DoAndReturn(func(_ context.Context, _ string, msg string) error {
				assert.Contains(t, msg, tt.expectedSnippet)
				return nil
			})

			err := uc.SendOTP(ctx, phone, ip, tt.reqType)
			require.NoError(t, err)
		})
	}
}

