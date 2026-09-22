package db

import (
	"context"
	"strings"
	"testing"
)

func TestMigrateRejectsInvalidDSNWithoutLeakingCredentials(t *testing.T) {
	const secret = "synthetic-private-password"
	err := Migrate(context.Background(), "postgres://user:"+secret+"@%invalid/database")
	if err == nil {
		t.Fatal("invalid migration connection accepted")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("migration error exposes credentials")
	}
}
