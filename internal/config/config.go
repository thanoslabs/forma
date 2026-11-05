// Package config validates process configuration before Forma starts.
package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Config struct {
	Production     bool
	DatabaseURL    string
	ListenAddress  string
	APIToken       string
	CredentialKey  string
	TrustedOrigins []string
	EnableDemoAPI  bool
	PaseoURL       string
	PaseoPassword  string
}

type Lookup func(string) string

func Load(lookup Lookup) (Config, error) {
	environment := strings.TrimSpace(lookup("FORMA_ENV"))
	if environment == "" {
		environment = "development"
	}
	if environment != "development" && environment != "production" {
		return Config{}, fmt.Errorf("FORMA_ENV must be development or production")
	}
	config := Config{
		Production: environment == "production", DatabaseURL: strings.TrimSpace(lookup("DATABASE_URL")), ListenAddress: strings.TrimSpace(lookup("FORMA_LISTEN_ADDR")), APIToken: lookup("FORMA_API_TOKEN"), CredentialKey: lookup("FORMA_CREDENTIAL_KEY"), EnableDemoAPI: lookup("FORMA_ENABLE_DEMO_API") == "true", PaseoURL: strings.TrimSpace(lookup("FORMA_PASEO_URL")), PaseoPassword: lookup("FORMA_PASEO_PASSWORD"),
	}
	if config.ListenAddress == "" {
		config.ListenAddress = ":8080"
	}
	if raw := strings.TrimSpace(lookup("FORMA_TRUSTED_ORIGINS")); raw != "" {
		for _, origin := range strings.Split(raw, ",") {
			origin = strings.TrimSpace(origin)
			parsed, err := url.Parse(origin)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.Path != "" {
				return Config{}, fmt.Errorf("invalid trusted origin %q", origin)
			}
			config.TrustedOrigins = append(config.TrustedOrigins, origin)
		}
	}
	if config.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if config.CredentialKey != "" {
		key, err := base64.StdEncoding.DecodeString(config.CredentialKey)
		if err != nil || len(key) != 32 {
			return Config{}, errors.New("FORMA_CREDENTIAL_KEY must be a base64-encoded 32-byte key")
		}
	}
	if config.Production {
		if len(config.APIToken) < 32 {
			return Config{}, errors.New("FORMA_API_TOKEN must contain at least 32 bytes in production")
		}
		if config.CredentialKey == "" {
			return Config{}, errors.New("FORMA_CREDENTIAL_KEY is required in production")
		}
		if len(config.TrustedOrigins) == 0 {
			return Config{}, errors.New("FORMA_TRUSTED_ORIGINS is required in production")
		}
		if config.EnableDemoAPI {
			return Config{}, errors.New("FORMA_ENABLE_DEMO_API cannot be enabled in production")
		}
	}
	return config, nil
}
