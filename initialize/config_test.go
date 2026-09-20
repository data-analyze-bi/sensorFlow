package initialize

import "testing"

func TestConfigValueUsesFallbackWhenUnset(t *testing.T) {
	const name = "SENSORFLOW_TEST_CONFIG_UNSET"
	t.Setenv(name, "")
	if got := configValue(name, "fallback"); got != "fallback" {
		t.Fatalf("configValue() = %q, want fallback", got)
	}
}

func TestConfigValueUsesEnvironment(t *testing.T) {
	const name = "SENSORFLOW_TEST_CONFIG_VALUE"
	t.Setenv(name, "configured")
	if got := configValue(name, "fallback"); got != "configured" {
		t.Fatalf("configValue() = %q, want configured", got)
	}
}
