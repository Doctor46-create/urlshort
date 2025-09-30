package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App    `yaml:"app"`
	Logger `yaml:"logger"`
	DB     `yaml:"db"`
}

type App struct {
	Address         string `yaml:"address" env-default:":8000"`
	BaseURL         string `yaml:"baseURL" env-default:"http://localhost:8080"`
	FileStoragePath string `yaml:"fileStoragePath" env:"FILE_STORAGE_PATH" env-default:"short_urls.json"`
}

type Logger struct {
	LogLevel int `yaml:"logLevel" env-default:"0"`
}

type DB struct {
	DSN string `yaml:"dsn" env:"DATABASE_DSN"`
}

func (c *Config) parseArgs() {
	flag.StringVar(&c.Address, "a", c.Address, "Host")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base url")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "File storage path")
	flag.StringVar(&c.DSN, "d", c.DSN, "Database DSN")
	flag.Parse()
}

func (c *Config) GetAddress() string {
	return c.Address
}

func (c *Config) GetBaseURL() string {
	return c.BaseURL
}

func (c *Config) GetFileStoragePath() string {
	return c.FileStoragePath
}

func (c *Config) GetDSN() string {
	return c.DSN
}

func GetConfig() *Config {
	var cfg Config

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DSN = envDSN
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Println("CONFIG_PATH is not declared and is set to default")
		configPath = "config.yaml"
	}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("cannot read config: %s", err)
		}
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DSN = envDSN
	}

	cfg.parseArgs()

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DSN = envDSN
	}

	log.Printf("Final DSN: %s", cfg.DSN)

	return &cfg
}
