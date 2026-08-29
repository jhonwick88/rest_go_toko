import sys

path = r'H:\FlutterProject\rest_go_toko\services\license_service.go'
with open(path, 'r', encoding='utf-8') as f:
    content = f.read()

target = """import (
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
}"""

replacement = """import (
	"bytes"
	"encoding/json"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
)

const (
	licenseFile   = "license.token"
	activateURL   = "http://localhost:8080/api/v1/license/activate"
)

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
		MachineFingerprint string json:"machine_fingerprint"
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return false
	}

	hardwareID, err := GetMachineFingerprint()
	if err != nil {
		return false
	}

	// The magic of Hardware Locking:
	return claims.MachineFingerprint == hardwareID
}"""

content = content.replace(target, replacement)

with open(path, 'w', encoding='utf-8') as f:
    f.write(content)
