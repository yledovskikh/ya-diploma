package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

const (
	runAddressDefault    = ":8081"
	databaseURIDefault   = "postgres://gophermart:Passw0rd@localhost:5432/database_name"
	accrualSystemAddress = "localhost:8080"
)

//адрес и порт запуска сервиса: переменная окружения ОС RUN_ADDRESS или флаг -a
//адрес подключения к базе данных: переменная окружения ОС DATABASE_URI или флаг -d
//адрес системы расчёта начислений: переменная окружения ОС ACCRUAL_SYSTEM_ADDRESS или флаг -r

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

var cfgF Config

func init() {
	flag.StringVar(&cfgF.RunAddress, "a", runAddressDefault, "Address and Port listener to server")
	flag.StringVar(&cfgF.DatabaseURI, "d", databaseURIDefault, "Database connect string")
	flag.StringVar(&cfgF.AccrualSystemAddress, "r", accrualSystemAddress, "Address and Port for connect to Accrual system")
}

func GetConfig() Config {
	var cfg Config

	flag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		log.Error().Err(err)
	}

	log.Debug().Msgf("cfg - %s", cfg)
	log.Debug().Msgf("cfgF - %s", cfgF)
	preferedVars(&cfg, &cfgF)
	return cfg
}

func preferedVars(cfg, cfgF *Config) *Config {
	if cfg.RunAddress == "" {
		cfg.RunAddress = cfgF.RunAddress
	}
	if cfg.DatabaseURI == "" {
		cfg.DatabaseURI = cfgF.DatabaseURI
	}
	if cfg.AccrualSystemAddress == "" {
		cfg.AccrualSystemAddress = cfgF.AccrualSystemAddress
	}

	return cfg
}
