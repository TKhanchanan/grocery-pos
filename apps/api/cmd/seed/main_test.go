package main

import (
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestEnsureMultiStatements(t *testing.T) {
	dsn, err := ensureMultiStatements("user:pass@tcp(db:3306)/grocery_pos?parseTime=true&multiStatements=false")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.MultiStatements {
		t.Fatal("multiStatements is false")
	}
	if !cfg.ParseTime {
		t.Fatal("parseTime was not preserved")
	}
	if !strings.Contains(dsn, "multiStatements=true") {
		t.Fatalf("DSN does not enable multiStatements: %q", dsn)
	}
}
