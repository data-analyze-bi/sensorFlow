package decoderbin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var payloadDecoderBinaryPaths []string

func init() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	dir := filepath.Dir(thisFile)
	payloadDecoderBinaryPaths = []string{
		filepath.Join(dir, "..", "binaries", "sensors-payload-decoder"),
		filepath.Join(dir, "bin", "sensors-payload-decoder"),
	}
}

// SetPayloadDecoderBinaryPath sets the primary decoder binary path.
// It is intended for host applications to override binary location by config.
func SetPayloadDecoderBinaryPath(path string) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return
	}
	payloadDecoderBinaryPaths = []string{clean}
}

// DecodePayload restores payload by invoking downloaded decoder binary.
func DecodePayload(encodedPayload string) ([]byte, error) {
	decoderBinaryPath, err := ensurePayloadDecoderBinary()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(decoderBinaryPath)
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
		return nil, fmt.Errorf("payload decoder binary failed: %s", errMsg)
	}
	result := bytes.TrimSpace(stdout.Bytes())
	if len(result) == 0 {
		return nil, fmt.Errorf("payload decoder binary returned empty payload")
	}
	return result, nil
}

func ensurePayloadDecoderBinary() (string, error) {
	var checked []string
	for _, payloadDecoderBinaryPath := range payloadDecoderBinaryPaths {
		info, err := os.Stat(payloadDecoderBinaryPath)
		if err == nil {
			if info.Mode()&0111 == 0 {
				if chmodErr := os.Chmod(payloadDecoderBinaryPath, 0755); chmodErr != nil {
					return "", fmt.Errorf("decoder binary chmod failed: %w", chmodErr)
				}
			}
			return payloadDecoderBinaryPath, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("decoder binary stat failed: %w", err)
		}
		checked = append(checked, payloadDecoderBinaryPath)
	}
	return "", fmt.Errorf("decoder binary not found, checked: %s", strings.Join(checked, ", "))
}
