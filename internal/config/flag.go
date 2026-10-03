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
	Host     string
	Port     string
	DBName   string
	Login    string
	Password string
	DBDSN    string
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "a", ":8080", "адрес запуска HTTP-сервера")
	flag.StringVar(&cfg.BaseURL, "b", "", "базовый адрес сокращённого UÎRL")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "file storage path")

	flag.StringVar(&cfg.DBConfig.Host, "hs", "localhost", "Хост БД")
	flag.StringVar(&cfg.DBConfig.Port, "port", "5432", "Порт БД")
	flag.StringVar(&cfg.DBConfig.DBName, "DB", "postgres", "Наименование БД")
	flag.StringVar(&cfg.DBConfig.Login, "u", "vladimir", "Имя пользователя БД")
	flag.StringVar(&cfg.DBConfig.Password, "p", "1111", "Пароль пользователя БД")
	flag.StringVar(&cfg.DBConfig.DBDSN, "d", "", "Строка подключения к БД")

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
	if envDBHost := os.Getenv("DB_HOST"); envDBHost != "" {
		cfg.DBConfig.Host = envDBHost
	}
	if envDBPort := os.Getenv("DB_PORT"); envDBPort != "" {
		cfg.DBConfig.Port = envDBPort
	}
	if envDBName := os.Getenv("DB_NAME"); envDBName != "" {
		cfg.DBConfig.DBName = envDBName
	}
	if envDBLogin := os.Getenv("DB_LOGIN"); envDBLogin != "" {
		cfg.DBConfig.Login = envDBLogin
	}
	if envDBPassword := os.Getenv("DB_PASSWORD"); envDBPassword != "" {
		cfg.DBConfig.Password = envDBPassword
	}
	if envDBDSN := os.Getenv("DB_DSN"); envDBDSN != "" {
		cfg.DBConfig.DBDSN = envDBDSN
	}

	return cfg
}
