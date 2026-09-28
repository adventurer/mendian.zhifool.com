package database

import (
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"mendian-backend/internal/config"
)

func TestBuildDSN(t *testing.T) {
	dsn, err := buildDSN(config.DatabaseConfig{
		Host:      "db.example.test",
		Port:      3306,
		Username:  "app-user",
		Password:  "p@ssword",
		Name:      "mendian",
		Charset:   "utf8mb4",
		ParseTime: true,
		Location:  "UTC",
	})
	if err != nil {
		t.Fatalf("build DSN: %v", err)
	}

	parsed, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}
	if parsed.User != "app-user" || parsed.Passwd != "p@ssword" || parsed.DBName != "mendian" {
		t.Fatalf("unexpected DSN credentials or database: %+v", parsed)
	}
	if parsed.Addr != "db.example.test:3306" || !parsed.ParseTime || parsed.Loc.String() != time.UTC.String() {
		t.Fatalf("unexpected DSN connection options: %+v", parsed)
	}
}

func TestCreateDatabaseStatement(t *testing.T) {
	statement, err := createDatabaseStatement("mendian", "utf8mb4")
	if err != nil {
		t.Fatalf("build create database statement: %v", err)
	}
	if statement != "CREATE DATABASE IF NOT EXISTS `mendian` CHARACTER SET utf8mb4" {
		t.Fatalf("unexpected create database statement: %s", statement)
	}

	if _, err := createDatabaseStatement("mendian", "utf8mb4; DROP DATABASE other"); err == nil {
		t.Fatal("expected invalid character set to be rejected")
	}
}
