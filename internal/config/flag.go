package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr            string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
	DBConfig        DBConfig
}

type DBConfig struct {
	DatabaseDSN string
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "a", ":8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "", "базовый адрес сокращённого UÎRL")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file storage path")
	flag.StringVar(&cfg.DBConfig.DatabaseDSN, "d", "", "database DSN")

	flag.Parse()

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	if envServAddr := os.Getenv("SERVER_ADDRESS"); envServAddr != "" {
		cfg.Addr = envServAddr
	}
	if envbaseURL := os.Getenv("BASE_URL"); envbaseURL != "" {
		cfg.BaseURL = envbaseURL
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	/////////////////////////////////////////////////////////////////////////////////////////
	if envDBDSN := os.Getenv("DATABASE_DSN"); envDBDSN != "" {
		cfg.DBConfig.DatabaseDSN = envDBDSN
	}

	return cfg
}
