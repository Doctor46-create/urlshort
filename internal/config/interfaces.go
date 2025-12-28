package config

type ServiceConfig interface {
	GetAddress() string
	GetBaseURL() string
	GetFileStoragePath() string
	GetDSN() string
	GetAuditFile() string
	GetAuditURL() string
	HasAudit() bool
	GetSecretKey() string
	IsHTTPSEnabled() bool
	GetTLSCertFile() string
	GetTLSKeyFile() string
	GetTrustedSubnet() string
	GetGRPCAddress() string
}
