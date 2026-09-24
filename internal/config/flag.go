package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr    string
	BaseUrl string
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "a", ":8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseUrl, "b", "", "базовый адрес сокращённого URL")

	flag.Parse()

	if envServAddr := os.Getenv("SERVER_ADDRESS"); envServAddr != "" {
		cfg.Addr = envServAddr
	}

	if envBaseUrl := os.Getenv("BASE_URL"); envBaseUrl != "" {
		cfg.BaseUrl = envBaseUrl
	}

	return cfg
}
