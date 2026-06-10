package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	const (
		maxOpenConns    = 20
		maxIdleConns    = 10
		connMaxIdleTime = 5 * time.Minute
		connMaxLifetime = 30 * time.Minute
	)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(connMaxIdleTime)
	db.SetConnMaxLifetime(connMaxLifetime)
	log.Printf("database pool max_open=%d max_idle=%d max_idle_time=%s max_lifetime=%s", maxOpenConns, maxIdleConns, connMaxIdleTime, connMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db, db.PingContext(ctx)
}

func ApplyFirstExistingSQLFile(db *sql.DB, paths ...string) error {
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return ApplySQL(db, string(content))
		}
		if !os.IsNotExist(err) {
			return err
		}
	}
	return fmt.Errorf("migration file not found: %s", strings.Join(paths, ", "))
}

func ApplySQL(db *sql.DB, script string) error {
	for _, statement := range splitSQLStatements(script) {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("migration statement failed: %w", err)
		}
	}
	return nil
}

func splitSQLStatements(script string) []string {
	statements := []string{}
	var current strings.Builder
	inSingleQuote := false

	for i := 0; i < len(script); i++ {
		ch := script[i]
		if ch == '\'' {
			inSingleQuote = !inSingleQuote
		}
		if ch == ';' && !inSingleQuote {
			statements = append(statements, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(ch)
	}
	if strings.TrimSpace(current.String()) != "" {
		statements = append(statements, current.String())
	}
	return statements
}

func WithTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
