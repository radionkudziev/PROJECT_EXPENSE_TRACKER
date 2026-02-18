package config

import (
	"fmt"
	"search-job/pkg/postgres"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	IsProd   bool
	Web      *WebParams
	Postgres *postgres.ConnectionData
}

type WebParams struct {
	Port uint16
}

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	viper.AddConfigPath("./configs/") // добавим ещё один путь

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Проверка обязательных полей
	port := viper.GetUint16("server.port")
	if port == 0 {
		return nil, fmt.Errorf("server.port is required in config")
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
			Port:     viper.GetUint16("server.pg.port"),
			DBName:   viper.GetString("server.pg.database"),
			SSLMode:  viper.GetString("server.pg.sslmode"),
		},
	}

	// Проверка PostgreSQL конфига
	if cfg.Postgres.User == "" || cfg.Postgres.Host == "" || cfg.Postgres.DBName == "" {
		return nil, fmt.Errorf("incomplete postgres configuration")
	}

	return cfg, nil
}

func (cfg *Config) GetWebPort() string {
	if cfg == nil || cfg.Web == nil {
		return ""
	}
	return ":" + strconv.Itoa(int(cfg.Web.Port))
}
