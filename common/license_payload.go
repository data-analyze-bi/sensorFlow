package common

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	licensebin "github.com/data-analyze-bi/license-bin-sdk"
)

const MaxEncodedPayloadBytes = 8 << 20

var licenseSlots = make(chan struct{}, 8)

func ProcessPayloadByLicense(encodedPayload string) ([]byte, error) {
	if len(encodedPayload) > MaxEncodedPayloadBytes {
		return nil, fmt.Errorf("encoded payload exceeds %d bytes", MaxEncodedPayloadBytes)
	}
	licensePath := strings.TrimSpace(beego.AppConfig.DefaultString("license_binary_path", ""))
	if licensePath == "" {
		licensePath = "./binaries/sensors-payload-license"
	}
	if licensePath == "" {
		return licensebin.ProcessPayload(encodedPayload)
	}
	if _, err := os.Stat(licensePath); err == nil {
		select {
		case licenseSlots <- struct{}{}:
			defer func() { <-licenseSlots }()
		default:
			return nil, fmt.Errorf("payload license processor is busy")
		}
		timeout := beego.AppConfig.DefaultInt("license_timeout_seconds", 0)
		if timeout == 0 {
			timeout = 10
		}
		if timeout < 1 {
			timeout = 1
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, licensePath)
		cmd.Stdin = strings.NewReader(encodedPayload)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("payload license processing timed out")
			}
			errMsg := strings.TrimSpace(stderr.String())
			if errMsg == "" {
				errMsg = err.Error()
			}
			return nil, fmt.Errorf("payload license processing failed: %s", errMsg)
		}
		out := bytes.TrimSpace(stdout.Bytes())
		if len(out) == 0 {
			return nil, fmt.Errorf("payload license processing returned empty payload")
		}
		return out, nil
	}
	return licensebin.ProcessPayload(encodedPayload)
}
