package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr     string
	BaseURL  string
	LogLevel string
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "a", ":8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "", "базовый адрес сокращённого URL")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")

	flag.Parse()

	if envServAddr := os.Getenv("SERVER_ADDRESS"); envServAddr != "" {
		cfg.Addr = envServAddr
	}
	if envbaseURL := os.Getenv("BASE_URL"); envbaseURL != "" {
		cfg.BaseURL = envbaseURL
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}

	return cfg
}
