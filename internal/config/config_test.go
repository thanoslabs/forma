package config

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestProductionConfigurationFailsClosed(t *testing.T) {
	values := map[string]string{"FORMA_ENV": "production", "DATABASE_URL": "postgres://db/forma"}
	lookup := func(key string) string { return values[key] }
	if _, err := Load(lookup); err == nil {
		t.Fatal("expected missing production security settings")
	}
	values["FORMA_API_TOKEN"] = strings.Repeat("t", 32)
	values["FORMA_CREDENTIAL_KEY"] = base64.StdEncoding.EncodeToString(make([]byte, 32))
	values["FORMA_TRUSTED_ORIGINS"] = "https://forma.example"
	config, err := Load(lookup)
	if err != nil || !config.Production || len(config.TrustedOrigins) != 1 {
		t.Fatalf("config=%#v err=%v", config, err)
	}
	values["FORMA_ENABLE_DEMO_API"] = "true"
	if _, err := Load(lookup); err == nil {
		t.Fatal("expected production demo rejection")
	}
}

func TestDevelopmentConfigurationHasSafeDefaults(t *testing.T) {
	config, err := Load(func(key string) string {
		if key == "DATABASE_URL" {
			return "postgres://db/forma"
		}
		return ""
	})
	if err != nil || config.Production || config.ListenAddress != ":8080" {
		t.Fatalf("config=%#v err=%v", config, err)
	}
}
