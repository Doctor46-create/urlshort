package config

import (
	"log"
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App `yaml:"app"`
}

type App struct {
	Address string `yaml:"address" env-default:":8000"`
	BaseURL string `yaml:"baseURL" env-default:"http://localhost:8080"`
}

func (c *Config) parseArgs() {
	flag.StringVar(&c.Address, "a", c.Address, "Host")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base url")
	flag.Parse()
}

func (c *Config) GetAddress() string {
    return c.Address
}

func (c *Config) GetBaseURL() string {
    return c.BaseURL
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
	cfg.parseArgs()

	return &cfg
}
