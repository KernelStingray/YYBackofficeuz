package click

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

type ClickGateway struct {
	MerchantID string
	ServiceID  string
	SecretKey  string
}

func NewClickGateway(merchantID, serviceID, secretKey string) *ClickGateway {
	return &ClickGateway{
		MerchantID: merchantID,
		ServiceID:  serviceID,
		SecretKey:  secretKey,
	}
}

func (c *ClickGateway) GeneratePaymentLink(txID uuid.UUID, amount float64, returnURL string) (string, error) {
	if strings.TrimSpace(returnURL) == "" {
		returnURL = "uz.yolyolakay.mobile://payment/success"
	}
	url := fmt.Sprintf("https://my.click.uz/services/pay?service_id=%s&merchant_id=%s&amount=%.2f&transaction_param=%s&return_url=%s",
		c.ServiceID, c.MerchantID, amount, txID.String(), url.QueryEscape(returnURL))
	return url, nil
}

func (c *ClickGateway) VerifySignature(payload []byte, signature string) bool {
	h := hmac.New(sha1.New, []byte(c.SecretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func (c *ClickGateway) ValidateClickSign(clickTransID, serviceID, merchantTransID, merchantPrepareID, amount, action, signTime, signString string) bool {
	var rawStr string
	if action == "1" {
		rawStr = fmt.Sprintf("%s%s%s%s%s%s%s%s", clickTransID, serviceID, c.SecretKey, merchantTransID, merchantPrepareID, amount, action, signTime)
	} else {
		rawStr = fmt.Sprintf("%s%s%s%s%s%s%s", clickTransID, serviceID, c.SecretKey, merchantTransID, amount, action, signTime)
	}
	hash := md5.Sum([]byte(rawStr))
	expectedSign := hex.EncodeToString(hash[:])
	return strings.EqualFold(expectedSign, signString)
}

type MockClickGateway struct {
	SecretKey string
}

func NewMockClickGateway(secret string) *MockClickGateway {
	return &MockClickGateway{SecretKey: secret}
}

func (m *MockClickGateway) GeneratePaymentLink(txID uuid.UUID, amount float64, returnURL string) (string, error) {
	return fmt.Sprintf("https://mock.click.uz/pay?tx_id=%s&amount=%.2f", txID.String(), amount), nil
}

func (m *MockClickGateway) VerifySignature(payload []byte, signature string) bool {
	h := hmac.New(sha1.New, []byte(m.SecretKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
