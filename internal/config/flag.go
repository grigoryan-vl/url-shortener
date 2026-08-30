package config

import "flag"

type Config struct {
	Addr    string
	BaseUrl string
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "a", ":8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseUrl, "b", "", "базовый адрес сокращённого URL")

	flag.Parse()
	return cfg
}
