package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	dsn, err := ensureMultiStatements(dsn)
	if err != nil {
		log.Fatalf("invalid DATABASE_URL: %v", err)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	files, err := filepath.Glob("seeds/*.sql")
	if err != nil {
		log.Fatalf("find seed files: %v", err)
	}

	sort.Strings(files)

	if len(files) == 0 {
		log.Println("no seed files found")
		return
	}

	log.Println("seed started")

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("read seed file %s: %v", file, err)
		}

		sqlText := strings.TrimSpace(string(content))
		if sqlText == "" {
			log.Printf("skip empty seed file: %s", file)
			continue
		}

		log.Printf("running seed: %s", file)

		if _, err := db.Exec(sqlText); err != nil {
			log.Fatalf("seed failed %s: %v", file, err)
		}
	}

	fmt.Println("seed applied")
}

func ensureMultiStatements(dsn string) (string, error) {
	cfg, err := mysql.ParseDSN(strings.TrimSpace(dsn))
	if err != nil {
		return "", err
	}
	cfg.MultiStatements = true
	return cfg.FormatDSN(), nil
}
