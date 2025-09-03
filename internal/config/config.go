// Package config
package config

import (
	"log"
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App `yaml:"app"`
	Logger `yaml:"logger"`
}

type App struct {
	Address string `yaml:"address" env-default:":8000"`
	BaseURL string `yaml:"baseURL" env-default:"http://localhost:8080"`
	FileStoragePath  string `yaml:"fileStoragePath" env:"FILE_STORAGE_PATH" env-default:"short_urls.json"`
}

type Logger struct {
	LogLevel int `yaml:"logLevel" env-default:"0"`
}

func (c *Config) parseArgs() {
	flag.StringVar(&c.Address, "a", c.Address, "Host")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base url")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "File storage path")	
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

func GetConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	
	if configPath == "" {
		log.Println("CONFIG_PATH is not declared and is set to default")
		configPath = "config.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file doesn't exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}
	
	cfg.parseArgs()

	return &cfg
}
