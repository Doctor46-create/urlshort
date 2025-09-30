package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Doctor46-create/urlshort/internal/config"
	"github.com/jackc/pgx/v5"
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

func (d *DBConfig) PingDB() error {
	if d.DSN == "" {
		return nil
	}

	conn, err := pgx.Connect(context.Background(), d.DSN)
	if err != nil {
		log.Printf("Error creating connection: %v", err)
		return err
	}
	defer conn.Close(context.Background())

	err = conn.Ping(context.Background())
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		return err
	}

	fmt.Println("Database connection successful")
	return nil
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
