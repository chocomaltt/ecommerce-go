// Package config loads service configuration from properties.yaml
// with environment variable overrides (dots -> underscores).
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server Server `mapstructure:"server"`
	Kratos Kratos `mapstructure:"kratos"`
	Hydra  Hydra  `mapstructure:"hydra"`
}

type Server struct {
	Port     string `mapstructure:"port"`
	LogLevel string `mapstructure:"log_level"`
}

type Kratos struct {
	PublicURL string `mapstructure:"public_url"`
	AdminURL  string `mapstructure:"admin_url"`
}

type Hydra struct {
	PublicURL    string   `mapstructure:"public_url"`
	AdminURL     string   `mapstructure:"admin_url"`
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURIs []string `mapstructure:"redirect_uris"`
}

// Load reads properties.yaml (path from CONFIG_PATH, default "properties.yaml")
// and overlays environment variables.
func Load() (Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath())
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("server.port", "8080")
	v.SetDefault("server.log_level", "info")
	v.SetDefault("kratos.public_url", "http://localhost:4433")
	v.SetDefault("kratos.admin_url", "http://localhost:4434")
	v.SetDefault("hydra.public_url", "http://localhost:4444")
	v.SetDefault("hydra.admin_url", "http://localhost:4445")
	v.SetDefault("hydra.client_id", "edge-api")
	v.SetDefault("hydra.client_secret", "edge-api-secret")
	v.SetDefault("hydra.redirect_uris", []string{"http://localhost:8080/auth/callback"})

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}

func configPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return "properties.yaml"
}
