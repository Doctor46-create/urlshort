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
	Audit  `yaml:"audit"`
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

type Audit struct {
	AuditFile string `yaml:"auditFile" env:"AUDIT_FILE" env-default:""`
	AuditURL  string `yaml:"auditURL" env:"AUDIT_URL" env-default:""`
}

func (c *Config) parseArgs() {
	flag.StringVar(&c.Address, "a", c.Address, "Host")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base url")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "File storage path")
	flag.StringVar(&c.DSN, "d", c.DSN, "Database DSN")
	flag.StringVar(&c.AuditFile, "audit-file", c.AuditFile, "Audit file path")
	flag.StringVar(&c.AuditURL, "audit-url", c.AuditURL, "Audit server URL")
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

func (c *Config) GetAuditFile() string {
	return c.AuditFile
}

func (c *Config) GetAuditURL() string {
	return c.AuditURL
}

func (c *Config) HasAudit() bool {
	return c.AuditFile != "" || c.AuditURL != ""
}

func GetConfig() *Config {
	var cfg Config

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

	cfg.parseArgs()

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		cfg.DSN = envDSN
	}

	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile != "" {
		cfg.AuditFile = envAuditFile
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		cfg.AuditURL = envAuditURL
	}

	log.Printf("Final DSN: %s", cfg.DSN)
	if cfg.AuditFile != "" {
		log.Printf("File audit enabled: %s", cfg.AuditFile)
	} else {
		log.Printf("File audit disabled")
	}
	if cfg.AuditURL != "" {
		log.Printf("HTTP audit enabled: %s", cfg.AuditURL)
	} else {
		log.Printf("HTTP audit disabled")
	}

	return &cfg
}
