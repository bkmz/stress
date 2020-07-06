package config

import (
	"github.com/rs/zerolog/log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ListenAddress string `envconfig:"LISTEN_ADDRESS" default:"0.0.0.0"`
	ListenPort    string `envconfig:"LISTEN_PORT" default:"8080"`
}

func Load() *Config {
	var c Config

	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().
			Str("ERROR", err.Error()).
			Msg("")
	}

	return &c
}
