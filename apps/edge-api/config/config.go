package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type Config struct {
	Server   Server  `mapstructure:"server"`
	Identity Service `mapstructure:"identity"`
	Order    Service `mapstructure:"order"`
}
type Server struct {
	Port     string `mapstructure:"port"`
	LogLevel string `mapstructure:"log_level"`
}
type Service struct {
	GRPCTarget    string `mapstructure:"grpc_target"`
	TLSServerName string `mapstructure:"tls_server_name"`
	TLSCertFile   string `mapstructure:"tls_cert_file"`
	TLSKeyFile    string `mapstructure:"tls_key_file"`
	TLSCAFile     string `mapstructure:"tls_ca_file"`
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath())
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.log_level", "info")
	for _, name := range []string{"identity", "order"} {
		v.SetDefault(name+".grpc_target", "localhost:"+map[string]string{"identity": "9081", "order": "9082"}[name])
		v.SetDefault(name+".tls_server_name", name+"-service.internal")
		v.SetDefault(name+".tls_cert_file", "../../deployments/compose/certs/edge-api.crt")
		v.SetDefault(name+".tls_key_file", "../../deployments/compose/certs/edge-api.key")
		v.SetDefault(name+".tls_ca_file", "../../deployments/compose/certs/ca.crt")
	}
	var cfg Config
	if err := v.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}
func configPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return "properties.yaml"
}
