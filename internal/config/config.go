package config

import (
	"encoding/json"
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
	Address         string `yaml:"address" env:"SERVER_ADDRESS" env-default:":8000"`
	GRPCAddress     string `yaml:"grpcAddress" env:"GRPC_ADDRESS" env-default:":9000"`
	BaseURL         string `yaml:"baseURL" env:"BASE_URL" env-default:"http://localhost:8000"`
	FileStoragePath string `yaml:"fileStoragePath" env:"FILE_STORAGE_PATH" env-default:"short_urls.json"`

	EnableHTTPS   bool   `yaml:"enableHTTPS" env:"ENABLE_HTTPS" env-default:"false"`
	CertFile      string `yaml:"certFile" env:"TLS_CERT_FILE" env-default:"server.crt"`
	KeyFile       string `yaml:"keyFile" env:"TLS_KEY_FILE" env-default:"server.key"`
	TrustedSubnet string `yaml:"trustedSubnet" env:"TRUSTED_SUBNET"`
}

type Logger struct {
	LogLevel int `yaml:"logLevel" env:"LOG_LEVEL" env-default:"0"`
}

type DB struct {
	DSN string `yaml:"dsn" env:"DATABASE_DSN"`
}

type Audit struct {
	AuditFile string `yaml:"auditFile" env:"AUDIT_FILE"`
	AuditURL  string `yaml:"auditURL" env:"AUDIT_URL"`
}

type Auth struct {
	SecretKey string `yaml:"secretKey" env:"SECRET_KEY"`
}

type JSONConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     *bool  `json:"enable_https"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

func GetConfig() *Config {
	var cfg Config

	loadFromYAML(&cfg)
	jsonPath := parseFlags(&cfg)
	loadFromJSON(&cfg, jsonPath)
	loadFromEnv(&cfg)
	validateSecretKey(&cfg)

	return &cfg
}

func loadFromYAML(cfg *Config) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	if _, err := os.Stat(configPath); err == nil {
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			log.Fatalf("cannot read YAML config: %v", err)
		}
	}
}

func loadFromEnv(cfg *Config) {
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("cannot read environment variables: %v", err)
	}
}

func loadFromJSON(cfg *Config, path string) {
	if path == "" {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read JSON config: %v", err)
	}

	var jc JSONConfig
	if err := json.Unmarshal(data, &jc); err != nil {
		log.Fatalf("invalid JSON config: %v", err)
	}

	if jc.ServerAddress == "" &&
		jc.BaseURL == "" &&
		jc.FileStoragePath == "" &&
		jc.DatabaseDSN == "" &&
		jc.EnableHTTPS == nil {
		log.Fatalf("JSON config is empty or invalid")
	}

	if jc.ServerAddress != "" {
		cfg.Address = jc.ServerAddress
	}
	if jc.BaseURL != "" {
		cfg.BaseURL = jc.BaseURL
	}
	if jc.FileStoragePath != "" {
		cfg.FileStoragePath = jc.FileStoragePath
	}
	if jc.DatabaseDSN != "" {
		cfg.DSN = jc.DatabaseDSN
	}
	if jc.EnableHTTPS != nil {
		cfg.EnableHTTPS = *jc.EnableHTTPS
	}
	if jc.TrustedSubnet != "" {
		cfg.TrustedSubnet = jc.TrustedSubnet
	}
}

func parseFlags(cfg *Config) string {
	var jsonConfigPath string

	flag.StringVar(&jsonConfigPath, "c", "", "Path to JSON config file")
	flag.StringVar(&jsonConfigPath, "config", "", "Path to JSON config file")

	flag.StringVar(&cfg.Address, "a", cfg.Address, "Server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path")
	flag.StringVar(&cfg.DSN, "d", cfg.DSN, "Database DSN")

	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Enable HTTPS")
	flag.StringVar(&cfg.CertFile, "tls-cert", cfg.CertFile, "TLS certificate file")
	flag.StringVar(&cfg.KeyFile, "tls-key", cfg.KeyFile, "TLS key file")

	flag.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "Audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "Audit server URL")
	flag.StringVar(&cfg.SecretKey, "secret-key", cfg.SecretKey, "Secret key")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "Trusted subnet in CIDR format")

	flag.Parse()

	if jsonConfigPath == "" {
		jsonConfigPath = os.Getenv("CONFIG")
	}

	return jsonConfigPath
}

func validateSecretKey(cfg *Config) {
	if cfg.SecretKey == "" {
		log.Println("WARNING: SECRET_KEY is not set, using development default")
		cfg.SecretKey = "guess_whos_back"
	}
}

func (c *Config) GetAddress() string         { return c.Address }
func (c *Config) GetBaseURL() string         { return c.BaseURL }
func (c *Config) GetFileStoragePath() string { return c.FileStoragePath }
func (c *Config) GetDSN() string             { return c.DSN }
func (c *Config) GetAuditFile() string       { return c.AuditFile }
func (c *Config) GetAuditURL() string        { return c.AuditURL }
func (c *Config) HasAudit() bool             { return c.AuditFile != "" || c.AuditURL != "" }
func (c *Config) GetSecretKey() string       { return c.SecretKey }
func (c *Config) IsHTTPSEnabled() bool       { return c.EnableHTTPS }
func (c *Config) GetTLSCertFile() string     { return c.CertFile }
func (c *Config) GetTLSKeyFile() string      { return c.KeyFile }
func (c *Config) GetTrustedSubnet() string   { return c.TrustedSubnet }
func (c *Config) GetGRPCAddress() string     { return c.GRPCAddress }
