package config

import (
	"os"

	"github.com/joho/godotenv"

	"BACKEND-UAS/database"
)

type Config struct {
	Connection *database.Connection
	Port       string
	JWTSecret  string
}

func NewConfig() *Config {
	godotenv.Load() // Load .env

	cfg := &Config{
		Connection: database.NewConnection(), // koneksi Postgres + Mongo
		Port:       os.Getenv("APP_PORT"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
	}

	if cfg.Port == "" {
		cfg.Port = "3000"
	}

	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "default-secret-ubah-sekarang"
	}

	return cfg
}
