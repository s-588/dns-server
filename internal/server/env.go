package server

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// LoadEnvs loads environment variables from the "dns-server.env" and "postgres.env" files.
func LoadEnvs() {
	// skip error because it can run in the environment that don't need for loading envs
	_ = godotenv.Load("dns-server.env", "postgres.env")
}

// OptionsFromEnv builds functional options from environment variables.
// If an environment variable is unset or empty, the corresponding option is omitted
// so NewServer keeps its built-in default.
func OptionsFromEnv() []Option {
	var opts []Option

	if p := strings.TrimSpace(os.Getenv("DNS_PORT")); p != "" {
		if !strings.HasPrefix(p, ":") {
			p = ":" + p
		}
		opts = append(opts, WithDNSPort(p))
	}

	if p := strings.TrimSpace(os.Getenv("HTTP_PORT")); p != "" {
		if !strings.HasPrefix(p, ":") {
			p = ":" + p
		}
		opts = append(opts, WithHTTPPort(p))
	}

	if v := strings.TrimSpace(os.Getenv("HTTPS")); v != "" {
		enable, err := strconv.ParseBool(v)
		if err != nil {
			lower := strings.ToLower(v)
			enable = lower == "1" || lower == "yes" || lower == "on" || lower == "true"
		}
		opts = append(opts, WithHTTPS(enable))
	}

	return opts
}
