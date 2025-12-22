package services

import (
	"context"
	"fmt"

	"github.com/tianyuso/advisorTool/db"
	"github.com/tianyuso/advisorTool/pkg/advisor"
)

// FetchDatabaseMetadata fetches database schema metadata from the connected database.
func FetchDatabaseMetadata(engineType advisor.Engine, dbParams *DBConnectionParams) (*advisor.DatabaseSchemaMetadata, error) {
	if dbParams == nil {
		return nil, fmt.Errorf("database connection parameters are nil")
	}

	// Build connection config
	config := &db.ConnectionConfig{
		DbType:        GetDbTypeString(engineType),
		Host:          dbParams.Host,
		Port:          dbParams.Port,
		User:          dbParams.User,
		Password:      dbParams.Password,
		DbName:        dbParams.DbName,
		Charset:       dbParams.Charset,
		ServiceName:   dbParams.ServiceName,
		Sid:           dbParams.Sid,
		SSLMode:       dbParams.SSLMode,
		Timeout:       dbParams.Timeout,
		Schema:        dbParams.Schema,
		SetSearchPath: false, // 获取元数据时不设置 search_path，避免影响审核
	}

	ctx := context.Background()

	// Open database connection
	conn, err := db.OpenConnection(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	// Fetch metadata
	metadata, err := db.GetDatabaseMetadata(ctx, conn, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get database metadata: %w", err)
	}

	return metadata, nil
}
