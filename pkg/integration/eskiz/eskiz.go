package eskiz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type SMSProvider interface {
	SendSMS(ctx context.Context, phone string, message string) error
}

type EskizProvider struct {
	email    string
	password string
	apiURL   string
	client   *http.Client

	mu              sync.RWMutex
	token           string
	tokenExpiration time.Time
}

func NewEskizProvider(email, password, apiURL string) SMSProvider {
	if apiURL == "" {
		apiURL = "https://notify.eskiz.uz/api"
	}
	apiURL = strings.TrimRight(apiURL, "/")

	return &EskizProvider{
		email:    email,
		password: password,
		apiURL:   apiURL,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *EskizProvider) SendSMS(ctx context.Context, phone string, message string) error {
	phone = strings.TrimPrefix(phone, "+")

	token, err := e.getToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get eskiz token: %w", err)
	}

	err = e.sendRequest(ctx, token, phone, message)

	if err != nil && isUnauthorizedErr(err) {
		e.resetToken()

		token, err = e.getToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to get eskiz token on retry: %w", err)
		}

		err = e.sendRequest(ctx, token, phone, message)
	}

	if err != nil {
		if os.Getenv("APP_ENV") == "development" || os.Getenv("ENVIRONMENT") == "development" {
			fmt.Printf("[ESKIZ API ERROR] %v\n", err)
			fmt.Printf("[MOCK SMS FALLBACK] To: %s, Message: %s\n", phone, message)
			return nil
		}
		return fmt.Errorf("failed to send SMS via eskiz to %s: %w", phone, err)
	}

	return nil
}

func (e *EskizProvider) sendRequest(ctx context.Context, token, phone, message string) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	_ = w.WriteField("mobile_phone", phone)
	_ = w.WriteField("message", message)

	senderID := os.Getenv("ESKIZ_SENDER_ID")
	if senderID == "" {
		senderID = "4546"
	}
	_ = w.WriteField("from", senderID)

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to create multipart form: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.apiURL+"/message/sms/send", &b)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return errUnauthorized
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("eskiz api returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (e *EskizProvider) getToken(ctx context.Context) (string, error) {
	e.mu.RLock()
	if e.token != "" && time.Now().Add(time.Hour).Before(e.tokenExpiration) {
		tok := e.token
		e.mu.RUnlock()
		return tok, nil
	}
	e.mu.RUnlock()

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.token != "" && time.Now().Add(time.Hour).Before(e.tokenExpiration) {
		return e.token, nil
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	_ = w.WriteField("email", e.email)
	_ = w.WriteField("password", e.password)
	_ = w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", e.apiURL+"/auth/login", &b)
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	if result.Data.Token == "" {
		return "", errors.New("empty token returned from eskiz")
	}

	e.token = result.Data.Token
	e.tokenExpiration = time.Now().Add(29 * 24 * time.Hour)

	return e.token, nil
}

func (e *EskizProvider) resetToken() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.token = ""
	e.tokenExpiration = time.Time{}
}

var errUnauthorized = errors.New("eskiz returned 401 unauthorized")

func isUnauthorizedErr(err error) bool {
	return errors.Is(err, errUnauthorized)
}

type MockSMSProvider struct{}

func NewMockSMSProvider() SMSProvider {
	return &MockSMSProvider{}
}

func (m *MockSMSProvider) SendSMS(ctx context.Context, phone string, message string) error {
	time.Sleep(500 * time.Millisecond)
	slog.Info("[MOCK SMS GATEWAY] To: %s | Message: %s\n", phone, message)
	return nil
}
