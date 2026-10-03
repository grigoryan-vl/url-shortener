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
	DbConfig        DbConfig
}

type DbConfig struct {
	Host     string
	Port     string
	DbName   string
	Login    string
	Password string
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "a", ":8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "", "базовый адрес сокращённого URL")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "file storage path")

	flag.StringVar(&cfg.DbConfig.Host, "hs", "localhost", "Хост БД")
	flag.StringVar(&cfg.DbConfig.Port, "port", "5432", "Порт БД")
	flag.StringVar(&cfg.DbConfig.DbName, "db", "postgres", "Наименование БД")
	flag.StringVar(&cfg.DbConfig.Login, "u", "vladimir", "Имя пользователя БД")
	flag.StringVar(&cfg.DbConfig.Password, "p", "1111", "Пароль пользователя БД")

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

	if envDbHost := os.Getenv("DB_HOST"); envDbHost != "" {
		cfg.DbConfig.Host = envDbHost
	}
	if envDbPort := os.Getenv("DB_PORT"); envDbPort != "" {
		cfg.DbConfig.Port = envDbPort
	}
	if envDbName := os.Getenv("DB_NAME"); envDbName != "" {
		cfg.DbConfig.DbName = envDbName
	}
	if envDbLogin := os.Getenv("DB_LOGIN"); envDbLogin != "" {
		cfg.DbConfig.Login = envDbLogin
	}
	if envDbPassword := os.Getenv("DB_PASSWORD"); envDbPassword != "" {
		cfg.DbConfig.Password = envDbPassword
	}

	return cfg
}
