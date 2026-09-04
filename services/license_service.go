package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
)

const licenseFile = "license.token"

// activateURL can be overridden at build time using -ldflags "-X rest_go_toko/services.activateURL=http://172.16.0.137/api/v1/license/activate"
var activateURL = "http://localhost:8080/api/v1/license/activate"

// GetMachineFingerprint retrieves a permanent Hardware ID for this machine.
func GetMachineFingerprint() (string, error) {
	id, err := machineid.ProtectedID("TokoPintar")
	if err != nil {
		return "", fmt.Errorf("failed to get hardware ID: %v", err)
	}
	return id, nil
}

// HasValidLicense checks if the license token exists AND is bound to this exact Hardware ID.
func HasValidLicense() bool {
	data, err := os.ReadFile(licenseFile)
	if err != nil {
		return false
	}

	tokenStr := strings.TrimSpace(string(data))
	if tokenStr == "" {
		return false
	}

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return false
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	var claims struct {
		MachineFingerprint string `json:"machine_fingerprint"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return false
	}

	hardwareID, err := GetMachineFingerprint()
	if err != nil {
		return false
	}

	return claims.MachineFingerprint == hardwareID
}

// ActivateLicense performs the activation request to the PintarLabs License server.
func ActivateLicense(licenseKey string) error {
	fingerprint, err := GetMachineFingerprint()
	if err != nil {
		return err
	}

	payload := map[string]string{
		"license_key":         licenseKey,
		"machine_fingerprint": fingerprint,
		"app_version":         "1.0.0",
		"hostname":            "TokoPintar-Server",
		"platform":            "windows-server",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(activateURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to connect to license server: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var result struct {
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("invalid response format: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(result.Message)
	}

	if result.Data.Token == "" {
		return errors.New("received empty token")
	}

	// Save token locally
	err = os.WriteFile(licenseFile, []byte(result.Data.Token), 0644)
	if err != nil {
		return fmt.Errorf("failed to save license token: %v", err)
	}

	return nil
}

// GetLicenseFeatures decodes the JWT and returns the features map
func GetLicenseFeatures() map[string]interface{} {
	data, err := os.ReadFile(licenseFile)
	if err != nil {
		return nil
	}

	tokenStr := strings.TrimSpace(string(data))
	if tokenStr == "" {
		return nil
	}

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return nil
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil
	}

	var claims struct {
		Features map[string]interface{} `json:"features"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil
	}

	return claims.Features
}


type LicenseClaims struct {
	LicenseID          string                 `json:"license_id"`
	ProductID          string                 `json:"product_id"`
	CustomerID         string                 `json:"customer_id"`
	PlanID             string                 `json:"plan_id"`
	InstallationID     string                 `json:"installation_id"`
	MachineFingerprint string                 `json:"machine_fingerprint"`
	Features           map[string]interface{} `json:"features"`
}

// GetLicenseClaims decodes the license.token and returns the full claims
func GetLicenseClaims() (*LicenseClaims, error) {
	data, err := os.ReadFile(licenseFile)
	if err != nil {
		return nil, errors.New("license not found")
	}

	tokenStr := strings.TrimSpace(string(data))
	if tokenStr == "" {
		return nil, errors.New("license is empty")
	}

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 && len(parts) != 3 {
		return nil, errors.New("invalid license token format")
	}

	var payloadIdx int
	if len(parts) == 3 {
		payloadIdx = 1 // Standard JWT: header.payload.signature
	} else if len(parts) == 2 {
		payloadIdx = 0 // Custom token: payload.signature
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[payloadIdx])
	if err != nil {
		return nil, errors.New("failed to decode license payload")
	}

	var claims LicenseClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errors.New("failed to parse license claims")
	}

	return &claims, nil
}
