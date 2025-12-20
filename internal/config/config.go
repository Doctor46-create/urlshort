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
	Auth   `yaml:"auth"`
}

type App struct {
	Address         string `yaml:"address" env-default:":8000"`
	BaseURL         string `yaml:"baseURL" env-default:"http://localhost:8000"`
	FileStoragePath string `yaml:"fileStoragePath" env:"FILE_STORAGE_PATH" env-default:"short_urls.json"`

	EnableHTTPS bool   `yaml:"enableHTTPS" env:"ENABLE_HTTPS" env-default:"false"`
	CertFile    string `yaml:"certFile" env:"TLS_CERT_FILE" env-default:"server.crt"`
	KeyFile     string `yaml:"keyFile" env:"TLS_KEY_FILE" env-default:"server.key"`
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

type Auth struct {
	SecretKey string `yaml:"secretKey" env:"SECRET_KEY" env-default:""`
}

func (c *Config) parseArgs() {
	flag.StringVar(&c.Address, "a", c.Address, "Server address")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base URL")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "File storage path")
	flag.StringVar(&c.DSN, "d", c.DSN, "Database DSN")

	flag.BoolVar(&c.EnableHTTPS, "s", c.EnableHTTPS, "Enable HTTPS")
	flag.StringVar(&c.CertFile, "tls-cert", c.CertFile, "TLS certificate file")
	flag.StringVar(&c.KeyFile, "tls-key", c.KeyFile, "TLS key file")

	flag.StringVar(&c.AuditFile, "audit-file", c.AuditFile, "Audit file path")
	flag.StringVar(&c.AuditURL, "audit-url", c.AuditURL, "Audit server URL")
	flag.StringVar(&c.SecretKey, "secret-key", c.SecretKey, "Secret key for cookies")

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

func (c *Config) GetSecretKey() string {
	return c.SecretKey
}

func (c *Config) IsHTTPSEnabled() bool {
	return c.EnableHTTPS
}

func (c *Config) GetTLSCertFile() string {
	return c.CertFile
}

func (c *Config) GetTLSKeyFile() string {
	return c.KeyFile
}

func GetConfig() *Config {
	var cfg Config

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Println("CONFIG_PATH is not set, using default config.yaml")
		configPath = "config.yaml"
	}

	if _, err := os.Stat(configPath); err == nil {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("cannot read config: %s", err)
		}
	}

	cfg.parseArgs()

	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		cfg.FileStoragePath = v
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.DSN = v
	}
	if v := os.Getenv("AUDIT_FILE"); v != "" {
		cfg.AuditFile = v
	}
	if v := os.Getenv("AUDIT_URL"); v != "" {
		cfg.AuditURL = v
	}
	if v := os.Getenv("SECRET_KEY"); v != "" {
		cfg.SecretKey = v
	}
	if v := os.Getenv("ENABLE_HTTPS"); v == "true" {
		cfg.EnableHTTPS = true
	}

	if cfg.SecretKey == "" {
		log.Println("WARNING: SECRET_KEY is not set, using development default")
		cfg.SecretKey = "guess_whos_back"
	}

	log.Printf("Server address: %s", cfg.Address)
	log.Printf("Base URL: %s", cfg.BaseURL)
	log.Printf("HTTPS enabled: %v", cfg.EnableHTTPS)

	if cfg.EnableHTTPS {
		log.Printf("TLS cert: %s", cfg.CertFile)
		log.Printf("TLS key: %s", cfg.KeyFile)
	}

	return &cfg
}
