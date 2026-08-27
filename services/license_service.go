package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	machineIDFile = "machine.id"
	licenseFile   = "license.token"
	activateURL   = "http://localhost:8080/api/v1/license/activate"
)

// GetMachineFingerprint retrieves or creates a permanent UUID for this machine.
func GetMachineFingerprint() (string, error) {
	data, err := os.ReadFile(machineIDFile)
	if err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	newID := uuid.New().String()
	err = os.WriteFile(machineIDFile, []byte(newID), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save machine ID: %v", err)
	}

	return newID, nil
}

// HasValidLicense checks if the license token exists.
// In a full implementation, this could also decode the JWT to check expiration offline.
func HasValidLicense() bool {
	data, err := os.ReadFile(licenseFile)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) != ""
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
