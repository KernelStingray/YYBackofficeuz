package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

	if vaultAddr == "" || vaultToken == "" {
		log.Fatal("VAULT_ADDR and VAULT_TOKEN must be set")
	}

	// Wait for Vault to be ready
	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; i < 10; i++ {
		resp, err := client.Get(vaultAddr + "/v1/sys/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			break
		}
		log.Println("Waiting for Vault to be ready...")
		time.Sleep(2 * time.Second)
	}

	// Check if JWT keys already exist
	req, _ := http.NewRequest("GET", vaultAddr+"/v1/secret/data/jwt", nil)
	req.Header.Set("X-Vault-Token", vaultToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to check Vault: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("JWT keys already exist in Vault.")
		return
	}

	log.Println("Generating new 4096-bit RSA key pair...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("Failed to generate RSA key: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatalf("Failed to marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	// Store in Vault KV v2 format
	payload := map[string]interface{}{
		"data": map[string]string{
			"private_key": string(privPEM),
			"public_key":  string(pubPEM),
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	req, _ = http.NewRequest("POST", vaultAddr+"/v1/secret/data/jwt", bytes.NewBuffer(payloadBytes))
	req.Header.Set("X-Vault-Token", vaultToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("Failed to write keys to Vault: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Vault returned error: %s - %s", resp.Status, string(body))
	}

	log.Println("Successfully generated and stored JWT keys in Vault!")
}
