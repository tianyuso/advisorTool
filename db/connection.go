// Package db provides database connection utilities for the advisor tool.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/microsoft/go-mssqldb"
	// Oracle driver - uncomment if needed
	// _ "github.com/godror/godror"
)

// ConnectionConfig holds database connection configuration.
type ConnectionConfig struct {
	DbType      string // mysql, postgres, mssql, oracle
	Host        string
	Port        int
	User        string
	Password    string
	DbName      string
	Charset     string
	ServiceName string // For Oracle
	Sid         string // For Oracle
	SSLMode     string // For PostgreSQL
	Timeout     int    // Connection timeout in seconds
}

// OpenConnection opens a database connection based on the configuration.
func OpenConnection(ctx context.Context, config *ConnectionConfig) (*sql.DB, error) {
	if config.Timeout == 0 {
		config.Timeout = 5
	}

	dsn, driverName, err := buildDSN(config)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Duration(config.Timeout*2) * time.Second)

	// Test the connection
	pingCtx, cancel := context.WithTimeout(ctx, time.Duration(config.Timeout)*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// buildDSN builds the data source name based on database type.
func buildDSN(config *ConnectionConfig) (dsn string, driverName string, err error) {
	switch config.DbType {
	case "mysql", "mariadb", "tidb", "oceanbase":
		driverName = "mysql"
		charset := config.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&timeout=%ds",
			config.User, config.Password, config.Host, config.Port, config.DbName, charset, config.Timeout)

	case "postgres":
		driverName = "postgres"
		sslMode := config.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
			config.Host, config.Port, config.User, config.Password, config.DbName, sslMode, config.Timeout)

	case "mssql", "sqlserver":
		driverName = "sqlserver"
		dsn = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;encrypt=disable;dial timeout=%d",
			config.Host, config.User, config.Password, config.Port, config.DbName, config.Timeout)

	case "oracle":
		driverName = "godror"
		if config.ServiceName != "" {
			dsn = fmt.Sprintf("user=\"%s\" password=\"%s\" connectString=\"%s:%d/%s\"",
				config.User, config.Password, config.Host, config.Port, config.ServiceName)
		} else if config.Sid != "" {
			dsn = fmt.Sprintf("user=\"%s\" password=\"%s\" connectString=\"%s:%d/%s\"",
				config.User, config.Password, config.Host, config.Port, config.Sid)
		} else {
			return "", "", fmt.Errorf("oracle requires either serviceName or sid")
		}

	default:
		return "", "", fmt.Errorf("unsupported database type: %s", config.DbType)
	}

	return dsn, driverName, nil
}

// TestConnection tests if the database connection is valid.
func TestConnection(ctx context.Context, config *ConnectionConfig) error {
	db, err := OpenConnection(ctx, config)
	if err != nil {
		return err
	}
	defer db.Close()
	return nil
}
