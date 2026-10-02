package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"time"

	golangjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager handles generation and validation of JWT access tokens and secure refresh tokens.
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewJWTManager creates a manager for both signing (needs private key) and validation (needs public key).
// For services that only validate, privateKey can be nil.
func NewJWTManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *JWTManager {
	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

type AccessClaims struct {
	UserID    uuid.UUID `json:"sub"`
	Role      string    `json:"role"`
	SessionID uuid.UUID `json:"session_id"`
	golangjwt.RegisteredClaims
}

// GenerateAccessToken creates a short-lived RS256 token (e.g. 15 minutes)
func (m *JWTManager) GenerateAccessToken(userID uuid.UUID, role string, sessionID uuid.UUID, ttl time.Duration) (string, error) {
	if m.privateKey == nil {
		return "", errors.New("private key is required to generate access tokens")
	}

	claims := AccessClaims{
		UserID:    userID,
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: golangjwt.RegisteredClaims{
			ExpiresAt: golangjwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  golangjwt.NewNumericDate(time.Now()),
		},
	}
	token := golangjwt.NewWithClaims(golangjwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// GenerateRefreshToken creates a long-lived token (e.g. 30 days) as a secure random string (not JWT)
func (m *JWTManager) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateAccessToken parses and validates an access token using the public key
func (m *JWTManager) ValidateAccessToken(tokenStr string) (*AccessClaims, error) {
	if m.publicKey == nil {
		return nil, errors.New("public key is required to validate access tokens")
	}

	token, err := golangjwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(token *golangjwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*golangjwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	return claims, nil
}
