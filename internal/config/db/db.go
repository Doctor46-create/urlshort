package db

import (
	"database/sql"
	"fmt"

	"github.com/Doctor46-create/urlshort/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBConfig struct {
	DSN string
}

type Database struct {
	*sql.DB
	config *DBConfig
}

func NewDBConfig(c *config.Config) *DBConfig {
	return &DBConfig{
		DSN: c.GetDSN(),
	}
}

func NewDatabase(dbConf *DBConfig) (*Database, error) {
	if dbConf.DSN == "" {
		return nil, fmt.Errorf("DSN is empty")
	}

	db, err := sql.Open("pgx", dbConf.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to verify connection: %w", err)
	}

	return &Database{
		DB:     db,
		config: dbConf,
	}, nil
}
