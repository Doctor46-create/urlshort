package config

type ServiceConfig interface {
	GetAddress() string
	GetBaseURL() string
	GetFileStoragePath() string
	GetAuditFile() string
	GetAuditURL() string
}
