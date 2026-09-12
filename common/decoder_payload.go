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
	"github.com/data-analyze-bi/decoder-bin-sdk"
)

const MaxEncodedPayloadBytes = 8 << 20

var decoderSlots = make(chan struct{}, 8)

func DecodePayloadByConfig(encodedPayload string) ([]byte, error) {
	if len(encodedPayload) > MaxEncodedPayloadBytes {
		return nil, fmt.Errorf("encoded payload exceeds %d bytes", MaxEncodedPayloadBytes)
	}
	decoderPath := strings.TrimSpace(beego.AppConfig.DefaultString("decoder_binary_path", "./binaries/sensors-payload-decoder"))
	if decoderPath == "" {
		return decoderbin.DecodePayload(encodedPayload)
	}
	if _, err := os.Stat(decoderPath); err == nil {
		select {
		case decoderSlots <- struct{}{}:
			defer func() { <-decoderSlots }()
		default:
			return nil, fmt.Errorf("payload decoder is busy")
		}
		timeout := beego.AppConfig.DefaultInt("decoder_timeout_seconds", 10)
		if timeout < 1 {
			timeout = 1
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, decoderPath)
		cmd.Stdin = strings.NewReader(encodedPayload)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			if ctx.Err() != nil {
				return nil, fmt.Errorf("payload decoder timed out")
			}
			errMsg := strings.TrimSpace(stderr.String())
			if errMsg == "" {
				errMsg = err.Error()
			}
			return nil, fmt.Errorf("payload decoder binary failed: %s", errMsg)
		}
		out := bytes.TrimSpace(stdout.Bytes())
		if len(out) == 0 {
			return nil, fmt.Errorf("payload decoder binary returned empty payload")
		}
		return out, nil
	}
	return decoderbin.DecodePayload(encodedPayload)
}
