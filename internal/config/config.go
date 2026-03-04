package config

import (
	"fmt"
	"search-job/internal/pkg/postgres"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	IsProd      bool
	Web         *WebParams
	Postgres    *postgres.ConnectionData
	ExternalAPI *ExternalAPIConfig `mapstructure:"external_api"`
	JWT         *JWTConfig         `mapstructure:"jwt"`
}

type WebParams struct {
	Port uint16
}

type ExternalAPIConfig struct {
	CurrencyURL string        `mapstructure:"currency_url"`
	Timeout     time.Duration `mapstructure:"timeout"`
	APIKey      string        `mapstructure:"api_key"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./cmd")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	port := viper.GetUint16("server.port")
	if port == 0 {
		return nil, fmt.Errorf("server.port is required")
	}

	cfg := &Config{
		IsProd: viper.GetBool("server.isProd"),
		Web: &WebParams{
			Port: port,
		},
		Postgres: &postgres.ConnectionData{
			User:     viper.GetString("server.pg.user"),
			Password: viper.GetString("server.pg.password"),
			Host:     viper.GetString("server.pg.host"),
			Port:     uint16(viper.GetInt("server.pg.port")),
			DBName:   viper.GetString("server.pg.database"),
			SSLMode:  viper.GetString("server.pg.sslmode"),
		},
		ExternalAPI: &ExternalAPIConfig{
			CurrencyURL: viper.GetString("external_api.currency_url"),
			Timeout:     viper.GetDuration("external_api.timeout"),
			APIKey:      viper.GetString("external_api.api_key"),
		},
		JWT: &JWTConfig{
			Secret: viper.GetString("jwt.secret"),
		},
	}

	if cfg.Postgres.User == "" || cfg.Postgres.Host == "" || cfg.Postgres.DBName == "" {
		return nil, fmt.Errorf("incomplete postgres configuration")
	}
	if cfg.ExternalAPI.CurrencyURL == "" {
		return nil, fmt.Errorf("external_api.currency_url is required")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("jwt.secret is required")
	}

	return cfg, nil
}

func (cfg *Config) GetWebPort() string {
	if cfg == nil || cfg.Web == nil {
		return ":8585"
	}
	return ":" + strconv.Itoa(int(cfg.Web.Port))
}
