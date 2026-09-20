package licensebin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var payloadLicensePaths []string

func init() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	dir := filepath.Dir(thisFile)
	payloadLicensePaths = []string{
		filepath.Join(dir, "..", "binaries", "sensors-payload-license"),
		filepath.Join(dir, "bin", "sensors-payload-license"),
	}
}

func SetPayloadLicensePath(path string) {
	clean := strings.TrimSpace(path)
	if clean != "" {
		payloadLicensePaths = []string{clean}
	}
}

func ProcessPayload(encodedPayload string) ([]byte, error) {
	licensePath, err := ensurePayloadLicense()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(licensePath)
	cmd.Stdin = strings.NewReader(encodedPayload)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return nil, fmt.Errorf("payload license processing failed: %s", errMsg)
	}
	result := bytes.TrimSpace(stdout.Bytes())
	if len(result) == 0 {
		return nil, fmt.Errorf("payload license processing returned empty payload")
	}
	return result, nil
}

func ensurePayloadLicense() (string, error) {
	var checked []string
	for _, licensePath := range payloadLicensePaths {
		info, err := os.Stat(licensePath)
		if err == nil {
			if info.Mode()&0111 == 0 {
				if chmodErr := os.Chmod(licensePath, 0755); chmodErr != nil {
					return "", fmt.Errorf("license chmod failed: %w", chmodErr)
				}
			}
			return licensePath, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("license stat failed: %w", err)
		}
		checked = append(checked, licensePath)
	}
	return "", fmt.Errorf("license not found, checked: %s", strings.Join(checked, ", "))
}
