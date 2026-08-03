// Package config loads identity-service configuration.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	GRPC   GRPC   `mapstructure:"grpc"`
	Kratos Kratos `mapstructure:"kratos"`
	Hydra  Hydra  `mapstructure:"hydra"`
	Actor  Actor  `mapstructure:"actor"`
}

type GRPC struct {
	Port          string `mapstructure:"port"`
	TLSCertFile   string `mapstructure:"tls_cert_file"`
	TLSKeyFile    string `mapstructure:"tls_key_file"`
	TLSCAFile     string `mapstructure:"tls_ca_file"`
	TrustedCaller string `mapstructure:"trusted_caller"`
}

type Kratos struct {
	PublicURL string `mapstructure:"public_url"`
}
type Hydra struct {
	AdminURL     string   `mapstructure:"admin_url"`
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURIs []string `mapstructure:"redirect_uris"`
}
type Actor struct {
	Issuer         string `mapstructure:"issuer"`
	PrivateKeyFile string `mapstructure:"private_key_file"`
	TTLSeconds     int    `mapstructure:"ttl_seconds"`
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath())
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("grpc.port", "9081")
	v.SetDefault("grpc.tls_cert_file", "../../deployments/compose/certs/identity-service.crt")
	v.SetDefault("grpc.tls_key_file", "../../deployments/compose/certs/identity-service.key")
	v.SetDefault("grpc.tls_ca_file", "../../deployments/compose/certs/ca.crt")
	v.SetDefault("grpc.trusted_caller", "edge-api.internal")
	v.SetDefault("kratos.public_url", "http://localhost:4433")
	v.SetDefault("hydra.admin_url", "http://localhost:4445")
	v.SetDefault("hydra.client_id", "edge-api")
	v.SetDefault("hydra.client_secret", "edge-api-secret")
	v.SetDefault("hydra.redirect_uris", []string{"http://localhost:8080/auth/callback"})
	v.SetDefault("actor.issuer", "identity-service")
	v.SetDefault("actor.private_key_file", "../../deployments/compose/certs/identity-actor.key")
	v.SetDefault("actor.ttl_seconds", 60)
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
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return "properties.yaml"
}
