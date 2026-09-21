package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL, S3Endpoint, S3AccessKey, S3SecretKey, S3Bucket, HTTPAddress, Environment string
	Timeout                                                                               time.Duration
	RateLimit                                                                             float64
	RateBurst                                                                             int
}

func Load(lookup func(string) (string, bool)) (Config, error) {
	get := func(key, fallback string) string {
		if value, ok := lookup("CORRIDOR_" + key); ok {
			return value
		}
		return fallback
	}
	c := Config{DatabaseURL: get("DATABASE_URL", ""), S3Endpoint: get("S3_ENDPOINT", ""), S3AccessKey: get("S3_ACCESS_KEY", ""), S3SecretKey: get("S3_SECRET_KEY", ""), S3Bucket: get("S3_BUCKET", ""), HTTPAddress: get("HTTP_ADDRESS", ":8080"), Environment: get("ENV", "development")}
	for key, value := range map[string]string{"DATABASE_URL": c.DatabaseURL, "S3_ENDPOINT": c.S3Endpoint, "S3_ACCESS_KEY": c.S3AccessKey, "S3_SECRET_KEY": c.S3SecretKey, "S3_BUCKET": c.S3Bucket} {
		if strings.TrimSpace(value) == "" {
			return Config{}, fmt.Errorf("CORRIDOR_%s is required", key)
		}
	}
	if err := validate(c); err != nil {
		return Config{}, err
	}
	timeout, err := time.ParseDuration(get("TIMEOUT", "3s"))
	if err != nil || timeout <= 0 || timeout > 10*time.Second {
		return Config{}, errors.New("CORRIDOR_TIMEOUT must be between 0 and 10s")
	}
	rate, err := strconv.ParseFloat(get("RATE_LIMIT", "100"), 64)
	if err != nil || !(rate > 0 && rate <= 10000) {
		return Config{}, errors.New("CORRIDOR_RATE_LIMIT must be 0–10000")
	}
	burst, err := strconv.Atoi(get("RATE_BURST", "200"))
	if err != nil || burst < 1 || burst > 10000 {
		return Config{}, errors.New("CORRIDOR_RATE_BURST must be 1–10000")
	}
	return Config{DatabaseURL: c.DatabaseURL, S3Endpoint: c.S3Endpoint, S3AccessKey: c.S3AccessKey, S3SecretKey: c.S3SecretKey, S3Bucket: c.S3Bucket, HTTPAddress: c.HTTPAddress, Environment: c.Environment, Timeout: timeout, RateLimit: rate, RateBurst: burst}, nil
}
func validate(c Config) error {
	db, err := url.Parse(c.DatabaseURL)
	if err != nil || db.Hostname() == "" || (db.Scheme != "postgres" && db.Scheme != "postgresql") {
		return errors.New("CORRIDOR_DATABASE_URL must be a PostgreSQL URL")
	}
	s3, err := url.Parse(c.S3Endpoint)
	if err != nil || s3.Hostname() == "" || (s3.Scheme != "http" && s3.Scheme != "https") || s3.User != nil || s3.RawQuery != "" || s3.Fragment != "" || (s3.Path != "" && s3.Path != "/") {
		return errors.New("CORRIDOR_S3_ENDPOINT must be an HTTP(S) origin without credentials")
	}
	_, port, err := net.SplitHostPort(c.HTTPAddress)
	if err != nil {
		return errors.New("CORRIDOR_HTTP_ADDRESS must contain host and port")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return errors.New("CORRIDOR_HTTP_ADDRESS port must be 1–65535")
	}
	if c.Environment != "development" && c.Environment != "production" && c.Environment != "test" {
		return errors.New("CORRIDOR_ENV is invalid")
	}
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`).MatchString(c.S3Bucket) {
		return errors.New("CORRIDOR_S3_BUCKET is invalid")
	}
	return nil
}
