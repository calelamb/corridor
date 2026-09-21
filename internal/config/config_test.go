package config

import (
	"strings"
	"testing"
)

func valid() map[string]string {
	return map[string]string{"CORRIDOR_DATABASE_URL": "postgres://app:test@localhost/corridor", "CORRIDOR_S3_ENDPOINT": "http://localhost:9000", "CORRIDOR_S3_ACCESS_KEY": "test-only-key", "CORRIDOR_S3_SECRET_KEY": "test-only-secret", "CORRIDOR_S3_BUCKET": "corridor-raw"}
}
func TestLoad(t *testing.T) {
	cfg, err := Load(func(k string) (string, bool) { v, ok := valid()[k]; return v, ok })
	if err != nil || cfg.HTTPAddress != ":8080" {
		t.Fatalf("%+v %v", cfg, err)
	}
	for _, tt := range []struct{ key, value string }{{"CORRIDOR_DATABASE_URL", ""}, {"CORRIDOR_DATABASE_URL", "postgres://app:private-secret@/db"}, {"CORRIDOR_HTTP_ADDRESS", ":99999"}, {"CORRIDOR_HTTP_ADDRESS", "broken"}, {"CORRIDOR_S3_ENDPOINT", "https://u:private-secret@host"}, {"CORRIDOR_S3_ENDPOINT", "ftp://host"}, {"CORRIDOR_TIMEOUT", "0s"}, {"CORRIDOR_TIMEOUT", "no"}, {"CORRIDOR_ENV", "bogus"}, {"CORRIDOR_RATE_LIMIT", "0"}, {"CORRIDOR_RATE_BURST", "-1"}, {"CORRIDOR_S3_BUCKET", "../oops"}} {
		t.Run(tt.key+tt.value, func(t *testing.T) {
			values := valid()
			values[tt.key] = tt.value
			_, err := Load(func(k string) (string, bool) { v, ok := values[k]; return v, ok })
			if err == nil {
				t.Fatal("accepted invalid input")
			}
			if strings.Contains(err.Error(), "private-secret") {
				t.Fatal("secret leaked")
			}
		})
	}
}
