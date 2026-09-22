package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var ErrDatabaseURLRequired = errors.New("La variable DATABASE_URL es obligatoria")

type Config struct {
	DatabaseURL string
	Port        string
	BcryptCost  int

	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CookieSecure    bool
	CookieSameSite  string

	CorsAllowedOrigins []string
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, errors.New("La variable JWT_SECRET es obligatoria")
	}

	accessTokenTTL, err := parseDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}

	refreshTokenTTL, err := parseDurationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	cookieSecure := false
	if value := os.Getenv("COOKIE_SECURE"); value != "" {
		parsedSecure, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, errors.New("COOKIE_SECURE debe ser true o false")
		}
		cookieSecure = parsedSecure
	}

	cookieSameSite := os.Getenv("COOKIE_SAMESITE")
	if cookieSameSite == "" {
		cookieSameSite = "lax"
	}

	corsAllowedOrigins := parseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(corsAllowedOrigins) == 0 {
		corsAllowedOrigins = []string{"http://localhost:4200"}
	}

	return Config{
		DatabaseURL:        databaseURL,
		Port:               port,
		BcryptCost:         bcryptCost,
		JWTSecret:          jwtSecret,
		AccessTokenTTL:     accessTokenTTL,
		RefreshTokenTTL:    refreshTokenTTL,
		CookieSecure:       cookieSecure,
		CookieSameSite:     cookieSameSite,
		CorsAllowedOrigins: corsAllowedOrigins,
	}, nil
}

// parseOrigins splits a comma-separated list of origins into a trimmed slice.
// It drops empty entries and returns nil when no origin is provided.
func parseOrigins(value string) []string {
	var origins []string
	for _, origin := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func parseDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s debe ser una duracion valida (ej: 15m, 168h): %w", key, err)
	}
	return duration, nil
}
