package usecase

import (
	"strings"

	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/auth/domain"
	userDomain "github.com/hextec/yolyolakay-backend/internal/user/domain"
	"github.com/hextec/yolyolakay-backend/pkg/jwt"
	"github.com/hextec/yolyolakay-backend/pkg/rabbitmq"
)

type AuthUseCase struct {
	otpRepo     domain.OTPRepository
	sessionRepo domain.SessionRepository
	smsProvider domain.SMSProvider
	userRepo    userDomain.UserRepository
	jwtManager  *jwt.JWTManager
	rmqClient   *rabbitmq.Client
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewAuthUseCase(
	otpRepo domain.OTPRepository,
	sessionRepo domain.SessionRepository,
	smsProvider domain.SMSProvider,
	userRepo userDomain.UserRepository,
	jwtManager *jwt.JWTManager,
	rmqClient *rabbitmq.Client,
) *AuthUseCase {
	accessTTL := 1 * time.Hour
	if ttlStr := os.Getenv("ACCESS_TOKEN_TTL"); ttlStr != "" {
		if d, err := time.ParseDuration(ttlStr); err == nil && d > 0 {
			accessTTL = d
		}
	}

	return &AuthUseCase{
		otpRepo:     otpRepo,
		sessionRepo: sessionRepo,
		smsProvider: smsProvider,
		userRepo:    userRepo,
		jwtManager:  jwtManager,
		rmqClient:   rmqClient,
		accessTTL:   accessTTL,
		refreshTTL:  30 * 24 * time.Hour,
	}
}

// generateRandomCode creates a secure 5-digit code
func generateRandomCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(90000))
	return fmt.Sprintf("%05d", n.Int64()+10000)
}

func (u *AuthUseCase) SendOTP(ctx context.Context, phone string, ip string, reqType string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	testPhone := os.Getenv("TEST_PHONE")
	if testPhone == "" {
		testPhone = "+998991234567"
	}
	testOTP := os.Getenv("TEST_OTP")
	if testOTP == "" {
		testOTP = "99887"
	}

	isTestAccount := phone == testPhone || phone == "+998991234567" || strings.HasPrefix(phone, "+99800")

	// 0. Check User Existence based on reqType
	if reqType == "register" || reqType == "login" || reqType == "frgt_pswd" {
		_, err := u.userRepo.GetByPhone(ctx, phone)
		if !isTestAccount {
			if reqType == "register" && err == nil {
				return errors.New("user already exists")
			}
			if (reqType == "login" || reqType == "frgt_pswd") && err != nil {
				return errors.New("user not found")
			}
		}
	} else {
		return errors.New("invalid request type")
	}

	// 1. Check IP Rate Limit (Max 10 requests per hour)
	ipCount, err := u.otpRepo.CheckIPRateLimit(ctx, ip)
	if err != nil {
		return err
	}
	if ipCount >= 10 && !isTestAccount {
		return errors.New("rate limit exceeded: max 10 requests per hour from this IP")
	}

	// 2. Check SMS Rate Limit (Max 3 SMS per 5 minutes)
	count, err := u.otpRepo.CheckRateLimit(ctx, phone)
	if err != nil {
		return err
	}
	if count >= 3 && !isTestAccount {
		return errors.New("rate limit exceeded: max 3 requests per 5 minutes")
	}

	// 3. Generate 5-digit OTP
	code := generateRandomCode()
	if isTestAccount {
		code = testOTP
	}

	// 4. Save OTP in Redis
	if err := u.otpRepo.SaveOTP(ctx, phone, code); err != nil {
		return err
	}

	// 5. Increment Rate Limit Counters
	if !isTestAccount {
		if err := u.otpRepo.IncrementRateLimit(ctx, phone); err != nil {
			slog.Error("Failed to increment rate limit", "error", err)
		}
		_ = u.otpRepo.IncrementIPRateLimit(ctx, ip)
	}

	if isTestAccount {
		return nil
	}

	// 6. Select and Format SMS Template based on reqType
	var template string
	switch reqType {
	case "register":
		template = os.Getenv("SMS_TEMPLATE_REGISTER")
		if template == "" {
			template = "Yolyolakay platformasiga ro‘yxatdan o‘tish uchun tasdiqlash kodi: %s"
		}
	case "frgt_pswd":
		template = os.Getenv("SMS_TEMPLATE_FRGT_PSWD")
		if template == "" {
			template = "Yolyolakay platformasining parolini tiklash uchun tasdiqlash kodi: %s"
		}
	case "login":
		fallthrough
	default:
		template = os.Getenv("SMS_TEMPLATE_LOGIN")
		if template == "" {
			template = os.Getenv("SMS_TEMPLATE")
		}
		if template == "" {
			template = "Kodni hech kimga bermang! Yolyolakay mobil ilovasiga kirish uchun tasdiqlash kodi: %s"
		}
	}

	msg := fmt.Sprintf(template, code)
	return u.smsProvider.SendSMS(ctx, phone, msg)
}



func (u *AuthUseCase) VerifyOTP(ctx context.Context, phone string, code string, deviceIP, deviceID, userAgent, osType string, password ...string) (*domain.TokenResponse, error) {
	// 1. Check if phone is locked
	locked, err := u.otpRepo.IsPhoneLocked(ctx, phone)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, errors.New("account is temporarily locked due to too many failed attempts")
	}

	testPhone := os.Getenv("TEST_PHONE")
	if testPhone == "" {
		testPhone = "+998991234567"
	}
	testOTP := os.Getenv("TEST_OTP")
	if testOTP == "" {
		testOTP = "99887"
	}

	isTestAccount := phone == testPhone || phone == "+998991234567" || strings.HasPrefix(phone, "+99800")

	// 2. Verify OTP in Redis or static bypass for test account
	var valid bool
	if isTestAccount && code == testOTP {
		valid = true
	} else {
		valid, err = u.otpRepo.VerifyOTP(ctx, phone, code)
		if err != nil {
			return nil, err
		}
	}

	if !valid {
		// Increment failed attempts
		fails, err := u.otpRepo.IncrementFailedAttempts(ctx, phone)
		if err == nil && fails >= 5 {
			// Lock phone for 15 minutes
			_ = u.otpRepo.LockPhone(ctx, phone, 15*time.Minute)
			_ = u.otpRepo.ResetFailedAttempts(ctx, phone) // clean up counter
			return nil, errors.New("account is temporarily locked due to too many failed attempts")
		}
		return nil, errors.New("invalid or expired OTP")
	}

	// Reset failed attempts upon successful login
	_ = u.otpRepo.ResetFailedAttempts(ctx, phone)

	// 2. Get User from Database (Cluster A)
	user, err := u.userRepo.GetByPhone(ctx, phone)
	isNewUser := false

	if err != nil {
		// Assume user not found, so we create a new one (Passwordless Registration)
		// Usually we'd check if err == ErrNotFound, but sqlc returns sql.ErrNoRows which is caught in repo or here
		newUser, createErr := u.userRepo.Create(ctx, "Новый", "Пользователь", phone, userDomain.RolePassenger, nil)
		if createErr != nil {
			return nil, fmt.Errorf("failed to create new user: %w", createErr)
		}
		user = newUser
		isNewUser = true

		// Publish user.created event
		if u.rmqClient != nil {
			type userCreatedEvent struct {
				UserID string `json:"user_id"`
				Role   string `json:"role"`
			}
			ev := userCreatedEvent{UserID: user.ID.String(), Role: string(user.Role)}
			payload, _ := json.Marshal(ev)
			go func() {
				if err := u.rmqClient.Publish(context.Background(), "yolyolakay.events", "user.created", payload); err != nil {
					slog.Error("Failed to publish", "error", err)
				}
			}()
		}
	}

	// 2.1. Set password if provided
	if len(password) > 0 && password[0] != "" {
		pwd := password[0]
		if len(pwd) >= 6 {
			hashed, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
			if err == nil {
				if err := u.userRepo.UpdatePassword(ctx, user.ID, string(hashed)); err != nil {
					slog.Error("Failed to set user password during OTP verification", "error", err)
				}
			}
		}
	}

	// Generate Session ID early so we can embed it in JWT
	sessionID := uuid.New()

	// 3. Generate JWT Pair
	access, err := u.jwtManager.GenerateAccessToken(user.ID, string(user.Role), sessionID, u.accessTTL)
	if err != nil {
		return nil, err
	}

	refresh, err := u.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 4. Store Session in Redis
	session := &domain.Session{
		SessionID:    sessionID,
		UserID:       user.ID,
		DeviceID:     deviceID,
		UserAgent:    userAgent,
		OSType:       osType,
		RefreshToken: refresh,
		DeviceIP:     deviceIP,
		ExpiresAt:    time.Now().Add(u.refreshTTL),
		CreatedAt:    time.Now(),
	}
	if err := u.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	// Add device to trusted list in Postgres (ignoring errors to not fail login)
	_ = u.userRepo.AddTrustedDevice(ctx, user.ID, deviceID)

	return &domain.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		IsNewUser:    isNewUser,
	}, nil
}

// SetupPassword sets or updates the user's password
func (u *AuthUseCase) SetupPassword(ctx context.Context, userID uuid.UUID, password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return u.userRepo.UpdatePassword(ctx, userID, string(hashed))
}

// LoginWithPassword authenticates the user using their phone and password
func (u *AuthUseCase) LoginWithPassword(ctx context.Context, phone string, password string, deviceIP, deviceID, userAgent, osType string) (*domain.TokenResponse, error) {
	// 1. Check if phone is locked (reuse existing check from OTP)
	locked, err := u.otpRepo.IsPhoneLocked(ctx, phone)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, errors.New("account is temporarily locked due to too many failed attempts")
	}

	user, err := u.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, errors.New("invalid phone or password")
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, errors.New("password is not set for this account, please use OTP login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		// Increment failed attempts
		fails, err := u.otpRepo.IncrementFailedAttempts(ctx, phone)
		if err == nil && fails >= 5 {
			// Lock phone for 15 minutes
			_ = u.otpRepo.LockPhone(ctx, phone, 15*time.Minute)
			_ = u.otpRepo.ResetFailedAttempts(ctx, phone)
			return nil, errors.New("account is temporarily locked due to too many failed attempts")
		}
		return nil, errors.New("invalid phone or password")
	}

	// Reset failed attempts upon successful login
	_ = u.otpRepo.ResetFailedAttempts(ctx, phone)

	// Check if this device is trusted
	deviceKnown := false
	for _, td := range user.TrustedDevices {
		if td == deviceID {
			deviceKnown = true
			break
		}
	}
	if !deviceKnown {
		err := u.SendOTP(ctx, phone, deviceIP, "login")
		if err != nil {
			return nil, fmt.Errorf("new_device_sms_failed:%w", err)
		}
		return nil, errors.New("new_device_otp_sent")
	}

	// Generate Session ID early so we can embed it in JWT
	sessionID := uuid.New()

	// Generate JWT Pair
	access, err := u.jwtManager.GenerateAccessToken(user.ID, string(user.Role), sessionID, u.accessTTL)
	if err != nil {
		return nil, err
	}

	refresh, err := u.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store Session in Redis
	session := &domain.Session{
		SessionID:    sessionID,
		UserID:       user.ID,
		DeviceID:     deviceID,
		UserAgent:    userAgent,
		OSType:       osType,
		RefreshToken: refresh,
		DeviceIP:     deviceIP,
		ExpiresAt:    time.Now().Add(u.refreshTTL),
		CreatedAt:    time.Now(),
	}
	if err := u.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	return &domain.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		IsNewUser:    false, // They already have a password, so they are not new
	}, nil
}

func (u *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string, deviceIP string) (*domain.TokenResponse, error) {
	// 0. Check Grace Period (if another concurrent request already rotated this token in the last 60s)
	if graceRes, err := u.sessionRepo.GetGraceToken(ctx, refreshToken); err == nil && graceRes != nil {
		return graceRes, nil
	}

	// 1. Get Session from Redis using the refresh token
	session, err := u.sessionRepo.GetByToken(ctx, refreshToken)
	if err != nil || session == nil {
		return nil, errors.New("invalid or expired session")
	}

	// 2. Prevent token reuse / theft
	if session.RefreshToken != refreshToken {
		_ = u.sessionRepo.DeleteByToken(ctx, refreshToken) // Revoke session
		return nil, errors.New("token mismatch - possible reuse detected")
	}

	// 3. Delete old session mapping
	_ = u.sessionRepo.DeleteByToken(ctx, refreshToken)

	// 4. Generate New JWT Pair
	// We need the user's role, so we must fetch the user.
	user, err := u.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	access, err := u.jwtManager.GenerateAccessToken(user.ID, string(user.Role), session.SessionID, u.accessTTL)
	if err != nil {
		return nil, err
	}

	newRefresh, err := u.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// 5. Store New Session
	session.RefreshToken = newRefresh
	session.DeviceIP = deviceIP
	session.ExpiresAt = time.Now().Add(u.refreshTTL)

	if err := u.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	resp := &domain.TokenResponse{
		AccessToken:  access,
		RefreshToken: newRefresh,
		IsNewUser:    false,
	}

	// 6. Save in Grace Period cache for 60 seconds to support concurrent race conditions (RFC 6749)
	_ = u.sessionRepo.SaveGraceToken(ctx, refreshToken, resp, 60*time.Second)

	return resp, nil
}

func (u *AuthUseCase) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	return u.sessionRepo.GetAllByUserID(ctx, userID)
}

func (u *AuthUseCase) TerminateAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return u.sessionRepo.DeleteAllByUserID(ctx, userID)
}

func (u *AuthUseCase) TerminateSession(ctx context.Context, userID uuid.UUID, deviceID string) error {
	// Let's get all sessions to find the exact one being deleted (to get the session_id)
	sessions, err := u.sessionRepo.GetAllByUserID(ctx, userID)
	found := false
	if err == nil {
		for _, s := range sessions {
			if s.DeviceID == deviceID {
				found = true
				// Broadcast event
				type logoutEvent struct {
					UserID    uuid.UUID `json:"user_id"`
					SessionID uuid.UUID `json:"session_id"`
				}
				ev := logoutEvent{UserID: userID, SessionID: s.SessionID}
				if payload, err := json.Marshal(ev); err == nil {
					if u.rmqClient != nil {
						if err := u.rmqClient.Publish(ctx, "yolyolakay.events", "user.logged_out", payload); err != nil {
							slog.Error("Failed to publish", "error", err)
						}
					}
				}
				break
			}
		}
	}
	if !found {
		return errors.New("session not found")
	}
	return u.sessionRepo.DeleteByDeviceID(ctx, userID, deviceID)
}

func (u *AuthUseCase) SaveMockOTP(ctx context.Context, phone, code string) error {
	return u.otpRepo.SaveOTP(ctx, phone, code)
}
