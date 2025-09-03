package config

type ServiceConfig interface {
	GetAddress() string
	GetBaseURL() string
	GetFileStoragePath() string
}
