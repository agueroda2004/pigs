package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var ErrDatabaseURLRequired = errors.New("La variable DATABASE_URL es obligatoria")

type Config struct {
	DatabaseURL string
	Port        string
	BcryptCost  int
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("No se pudo cargar el archivo .env: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, ErrDatabaseURLRequired
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	bcryptCost := 0
	if value := os.Getenv("BCRYPT_COST"); value != "" {
		parsedCost, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, errors.New("BCRYPT_COST debe ser un numero entero")
		}
		bcryptCost = parsedCost
	}

	return Config{
		DatabaseURL: databaseURL,
		Port:        port,
		BcryptCost:  bcryptCost,
	}, nil
}
