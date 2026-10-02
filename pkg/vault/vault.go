package vault

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

func fetchJWTKeysFromVault() (string, string, error) {
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

	if vaultAddr == "" || vaultToken == "" {
		return "", "", errors.New("VAULT_ADDR and VAULT_TOKEN environment variables must be set")
	}

	req, err := http.NewRequest("GET", vaultAddr+"/v1/secret/data/jwt", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("X-Vault-Token", vaultToken)

	client := &http.Client{Timeout: 10 * time.Second}
	
	// Retry up to 5 times if Vault is not ready yet
	var resp *http.Response
	for i := 0; i < 5; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	
	if err != nil {
		return "", "", fmt.Errorf("failed to connect to vault: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("vault returned non-200 status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Data struct {
				PublicKey  string `json:"public_key"`
				PrivateKey string `json:"private_key"`
			} `json:"data"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	if result.Data.Data.PublicKey == "" {
		return "", "", errors.New("public key not found in vault")
	}

	return result.Data.Data.PrivateKey, result.Data.Data.PublicKey, nil
}

func LoadPublicKey() (*rsa.PublicKey, error) {
	_, pubPEM, err := fetchJWTKeysFromVault()
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("key type is not RSA")
	}
}

func LoadPrivateKey() (*rsa.PrivateKey, error) {
	privPEM, _, err := fetchJWTKeysFromVault()
	if err != nil {
		return nil, err
	}
	
	if privPEM == "" {
		return nil, errors.New("private key not found in vault")
	}

	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the private key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}
