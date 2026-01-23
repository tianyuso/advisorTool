package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	storepb "github.com/tianyuso/advisorTool/generated-go/store"
)

// GetDatabaseMetadata retrieves the database schema metadata.
func GetDatabaseMetadata(ctx context.Context, db *sql.DB, config *ConnectionConfig) (*storepb.DatabaseSchemaMetadata, error) {
	switch config.DbType {
	case "mysql", "mariadb", "tidb", "oceanbase":
		return getMySQLMetadata(ctx, db, config.DbName)
	case "postgres":
		return getPostgresMetadata(ctx, db, config.DbName)
	case "mssql", "sqlserver":
		return getMSSQLMetadata(ctx, db, config.DbName)
	case "oracle":
		return getOracleMetadata(ctx, db, config.DbName)
	default:
		return nil, fmt.Errorf("unsupported database type for metadata: %s", config.DbType)
	}
}

// getMySQLMetadata retrieves MySQL/MariaDB/TiDB database metadata.
func getMySQLMetadata(ctx context.Context, db *sql.DB, dbName string) (*storepb.DatabaseSchemaMetadata, error) {
	metadata := &storepb.DatabaseSchemaMetadata{
		Name: dbName,
		Schemas: []*storepb.SchemaMetadata{
			{
				Name:   "", // MySQL doesn't have schema concept, use empty string
				Tables: []*storepb.TableMetadata{},
			},
		},
	}

	// Get tables
	tables, err := getMySQLTables(ctx, db, dbName)
	if err != nil {
		return nil, err
	}
	metadata.Schemas[0].Tables = tables

	return metadata, nil
}

func getMySQLTables(ctx context.Context, db *sql.DB, dbName string) ([]*storepb.TableMetadata, error) {
	query := `
		SELECT TABLE_NAME, TABLE_COMMENT, ENGINE
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`
	rows, err := db.QueryContext(ctx, query, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []*storepb.TableMetadata
	for rows.Next() {
		var tableName, comment, engine sql.NullString
		if err := rows.Scan(&tableName, &comment, &engine); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}

		table := &storepb.TableMetadata{
			Name:    tableName.String,
			Comment: comment.String,
			Engine:  engine.String,
		}

		// Get columns for this table
		columns, err := getMySQLColumns(ctx, db, dbName, tableName.String)
		if err != nil {
			return nil, err
		}
		table.Columns = columns

		// Get indexes for this table
		indexes, err := getMySQLIndexes(ctx, db, dbName, tableName.String)
		if err != nil {
			return nil, err
		}
		table.Indexes = indexes

		tables = append(tables, table)
	}

	return tables, nil
}

func getMySQLColumns(ctx context.Context, db *sql.DB, dbName, tableName string) ([]*storepb.ColumnMetadata, error) {
	query := `
		SELECT 
			COLUMN_NAME, 
			COLUMN_TYPE, 
			IS_NULLABLE, 
			COLUMN_DEFAULT, 
			COLUMN_COMMENT,
			EXTRA,
			ORDINAL_POSITION
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`
	rows, err := db.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []*storepb.ColumnMetadata
	for rows.Next() {
		var name, colType, nullable, extra sql.NullString
		var defaultVal, comment sql.NullString
		var position int

		if err := rows.Scan(&name, &colType, &nullable, &defaultVal, &comment, &extra, &position); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col := &storepb.ColumnMetadata{
			Name:     name.String,
			Type:     colType.String,
			Nullable: nullable.String == "YES",
			Default:  defaultVal.String,
			Comment:  comment.String,
			Position: int32(position),
		}

		columns = append(columns, col)
	}

	return columns, nil
}

func getMySQLIndexes(ctx context.Context, db *sql.DB, dbName, tableName string) ([]*storepb.IndexMetadata, error) {
	query := `
		SELECT 
			INDEX_NAME,
			COLUMN_NAME,
			NON_UNIQUE,
			SEQ_IN_INDEX
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`
	rows, err := db.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	indexMap := make(map[string]*storepb.IndexMetadata)
	var indexOrder []string

	for rows.Next() {
		var indexName, columnName sql.NullString
		var nonUnique int
		var seqInIndex int

		if err := rows.Scan(&indexName, &columnName, &nonUnique, &seqInIndex); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}

		idx, ok := indexMap[indexName.String]
		if !ok {
			idx = &storepb.IndexMetadata{
				Name:        indexName.String,
				Expressions: []string{},
				Unique:      nonUnique == 0,
				Primary:     indexName.String == "PRIMARY",
			}
			indexMap[indexName.String] = idx
			indexOrder = append(indexOrder, indexName.String)
		}
		idx.Expressions = append(idx.Expressions, columnName.String)
	}

	var indexes []*storepb.IndexMetadata
	for _, name := range indexOrder {
		indexes = append(indexes, indexMap[name])
	}

	return indexes, nil
}

// getPostgresMetadata retrieves PostgreSQL database metadata.
func getPostgresMetadata(ctx context.Context, db *sql.DB, dbName string) (*storepb.DatabaseSchemaMetadata, error) {
	metadata := &storepb.DatabaseSchemaMetadata{
		Name:    dbName,
		Schemas: []*storepb.SchemaMetadata{},
	}

	// Get schemas
	schemaQuery := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		ORDER BY schema_name
	`
	rows, err := db.QueryContext(ctx, schemaQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, err
		}

		schema := &storepb.SchemaMetadata{
			Name:   schemaName,
			Tables: []*storepb.TableMetadata{},
		}

		// Get tables for this schema
		tables, err := getPostgresTables(ctx, db, schemaName)
		if err != nil {
			return nil, err
		}
		schema.Tables = tables

		metadata.Schemas = append(metadata.Schemas, schema)
	}

	return metadata, nil
}

func getPostgresTables(ctx context.Context, db *sql.DB, schemaName string) ([]*storepb.TableMetadata, error) {
	query := `
		SELECT 
			t.table_name,
			COALESCE(obj_description((quote_ident(t.table_schema) || '.' || quote_ident(t.table_name))::regclass), '') as comment
		FROM information_schema.tables t
		WHERE t.table_schema = $1 AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name
	`
	rows, err := db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []*storepb.TableMetadata
	for rows.Next() {
		var tableName, comment string
		if err := rows.Scan(&tableName, &comment); err != nil {
			return nil, err
		}

		table := &storepb.TableMetadata{
			Name:    tableName,
			Comment: comment,
		}

		// Get columns
		columns, err := getPostgresColumns(ctx, db, schemaName, tableName)
		if err != nil {
			return nil, err
		}
		table.Columns = columns

		// Get indexes
		indexes, err := getPostgresIndexes(ctx, db, schemaName, tableName)
		if err != nil {
			return nil, err
		}
		table.Indexes = indexes

		tables = append(tables, table)
	}

	return tables, nil
}

func getPostgresColumns(ctx context.Context, db *sql.DB, schemaName, tableName string) ([]*storepb.ColumnMetadata, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			COALESCE(col_description((quote_ident($1) || '.' || quote_ident($2))::regclass, ordinal_position), ''),
			ordinal_position
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`
	rows, err := db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []*storepb.ColumnMetadata
	for rows.Next() {
		var name, dataType, nullable string
		var defaultVal sql.NullString
		var comment string
		var position int

		if err := rows.Scan(&name, &dataType, &nullable, &defaultVal, &comment, &position); err != nil {
			return nil, err
		}

		col := &storepb.ColumnMetadata{
			Name:     name,
			Type:     dataType,
			Nullable: nullable == "YES",
			Default:  defaultVal.String,
			Comment:  comment,
			Position: int32(position),
		}
		columns = append(columns, col)
	}

	return columns, nil
}

func getPostgresIndexes(ctx context.Context, db *sql.DB, schemaName, tableName string) ([]*storepb.IndexMetadata, error) {
	query := `
		SELECT 
			i.relname as index_name,
			array_to_string(array_agg(a.attname ORDER BY array_position(ix.indkey, a.attnum)), ',') as column_names,
			ix.indisunique as is_unique,
			ix.indisprimary as is_primary
		FROM pg_class t
		JOIN pg_index ix ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE n.nspname = $1 AND t.relname = $2
		GROUP BY i.relname, ix.indisunique, ix.indisprimary
		ORDER BY i.relname
	`
	rows, err := db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []*storepb.IndexMetadata
	for rows.Next() {
		var indexName, columnNames string
		var isUnique, isPrimary bool

		if err := rows.Scan(&indexName, &columnNames, &isUnique, &isPrimary); err != nil {
			return nil, err
		}

		idx := &storepb.IndexMetadata{
			Name:        indexName,
			Expressions: strings.Split(columnNames, ","),
			Unique:      isUnique,
			Primary:     isPrimary,
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// getMSSQLMetadata retrieves SQL Server database metadata.
func getMSSQLMetadata(ctx context.Context, db *sql.DB, dbName string) (*storepb.DatabaseSchemaMetadata, error) {
	metadata := &storepb.DatabaseSchemaMetadata{
		Name:    dbName,
		Schemas: []*storepb.SchemaMetadata{},
	}

	// Get schemas
	schemaQuery := `
		SELECT SCHEMA_NAME 
		FROM INFORMATION_SCHEMA.SCHEMATA 
		WHERE CATALOG_NAME = @p1
		AND SCHEMA_NAME NOT IN ('sys', 'INFORMATION_SCHEMA', 'guest')
		ORDER BY SCHEMA_NAME
	`
	rows, err := db.QueryContext(ctx, schemaQuery, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, err
		}

		schema := &storepb.SchemaMetadata{
			Name:   schemaName,
			Tables: []*storepb.TableMetadata{},
		}

		tables, err := getMSSQLTables(ctx, db, dbName, schemaName)
		if err != nil {
			return nil, err
		}
		schema.Tables = tables

		metadata.Schemas = append(metadata.Schemas, schema)
	}

	return metadata, nil
}

func getMSSQLTables(ctx context.Context, db *sql.DB, dbName, schemaName string) ([]*storepb.TableMetadata, error) {
	query := `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_CATALOG = @p1 AND TABLE_SCHEMA = @p2 AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`
	rows, err := db.QueryContext(ctx, query, dbName, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []*storepb.TableMetadata
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		table := &storepb.TableMetadata{
			Name: tableName,
		}

		columns, err := getMSSQLColumns(ctx, db, dbName, schemaName, tableName)
		if err != nil {
			return nil, err
		}
		table.Columns = columns

		indexes, err := getMSSQLIndexes(ctx, db, schemaName, tableName)
		if err != nil {
			return nil, err
		}
		table.Indexes = indexes

		tables = append(tables, table)
	}

	return tables, nil
}

func getMSSQLColumns(ctx context.Context, db *sql.DB, dbName, schemaName, tableName string) ([]*storepb.ColumnMetadata, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			ORDINAL_POSITION
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_CATALOG = @p1 AND TABLE_SCHEMA = @p2 AND TABLE_NAME = @p3
		ORDER BY ORDINAL_POSITION
	`
	rows, err := db.QueryContext(ctx, query, dbName, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []*storepb.ColumnMetadata
	for rows.Next() {
		var name, dataType, nullable string
		var defaultVal sql.NullString
		var position int

		if err := rows.Scan(&name, &dataType, &nullable, &defaultVal, &position); err != nil {
			return nil, err
		}

		col := &storepb.ColumnMetadata{
			Name:     name,
			Type:     dataType,
			Nullable: nullable == "YES",
			Default:  defaultVal.String,
			Position: int32(position),
		}
		columns = append(columns, col)
	}

	return columns, nil
}

func getMSSQLIndexes(ctx context.Context, db *sql.DB, schemaName, tableName string) ([]*storepb.IndexMetadata, error) {
	query := `
		SELECT 
			i.name as index_name,
			STUFF((
				SELECT ',' + c2.name
				FROM sys.index_columns ic2
				INNER JOIN sys.columns c2 ON ic2.object_id = c2.object_id AND ic2.column_id = c2.column_id
				WHERE ic2.object_id = i.object_id AND ic2.index_id = i.index_id
				ORDER BY ic2.key_ordinal
				FOR XML PATH(''), TYPE
			).value('.', 'NVARCHAR(MAX)'), 1, 1, '') as column_names,
			i.is_unique,
			i.is_primary_key
		FROM sys.indexes i
		INNER JOIN sys.tables t ON i.object_id = t.object_id
		INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
		WHERE s.name = @p1 AND t.name = @p2 AND i.name IS NOT NULL
		GROUP BY i.object_id, i.index_id, i.name, i.is_unique, i.is_primary_key
		ORDER BY i.name
	`
	rows, err := db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []*storepb.IndexMetadata
	for rows.Next() {
		var indexName, columnNames string
		var isUnique, isPrimary bool

		if err := rows.Scan(&indexName, &columnNames, &isUnique, &isPrimary); err != nil {
			return nil, err
		}

		idx := &storepb.IndexMetadata{
			Name:        indexName,
			Expressions: strings.Split(columnNames, ","),
			Unique:      isUnique,
			Primary:     isPrimary,
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// getOracleMetadata retrieves Oracle database metadata (placeholder).
func getOracleMetadata(ctx context.Context, db *sql.DB, dbName string) (*storepb.DatabaseSchemaMetadata, error) {
	metadata := &storepb.DatabaseSchemaMetadata{
		Name: dbName,
		Schemas: []*storepb.SchemaMetadata{
			{
				Name:   dbName,
				Tables: []*storepb.TableMetadata{},
			},
		},
	}

	// Get tables
	query := `
		SELECT TABLE_NAME, COMMENTS
		FROM USER_TAB_COMMENTS
		WHERE TABLE_TYPE = 'TABLE'
		ORDER BY TABLE_NAME
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		var comment sql.NullString
		if err := rows.Scan(&tableName, &comment); err != nil {
			return nil, err
		}

		table := &storepb.TableMetadata{
			Name:    tableName,
			Comment: comment.String,
		}

		// Get columns
		columns, err := getOracleColumns(ctx, db, tableName)
		if err != nil {
			return nil, err
		}
		table.Columns = columns

		// Get indexes
		indexes, err := getOracleIndexes(ctx, db, tableName)
		if err != nil {
			return nil, err
		}
		table.Indexes = indexes

		metadata.Schemas[0].Tables = append(metadata.Schemas[0].Tables, table)
	}

	return metadata, nil
}

func getOracleColumns(ctx context.Context, db *sql.DB, tableName string) ([]*storepb.ColumnMetadata, error) {
	query := `
		SELECT 
			c.COLUMN_NAME,
			c.DATA_TYPE,
			c.NULLABLE,
			c.DATA_DEFAULT,
			cc.COMMENTS,
			c.COLUMN_ID
		FROM USER_TAB_COLUMNS c
		LEFT JOIN USER_COL_COMMENTS cc ON c.TABLE_NAME = cc.TABLE_NAME AND c.COLUMN_NAME = cc.COLUMN_NAME
		WHERE c.TABLE_NAME = :1
		ORDER BY c.COLUMN_ID
	`
	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []*storepb.ColumnMetadata
	for rows.Next() {
		var name, dataType, nullable string
		var defaultVal, comment sql.NullString
		var position int

		if err := rows.Scan(&name, &dataType, &nullable, &defaultVal, &comment, &position); err != nil {
			return nil, err
		}

		col := &storepb.ColumnMetadata{
			Name:     name,
			Type:     dataType,
			Nullable: nullable == "Y",
			Default:  strings.TrimSpace(defaultVal.String),
			Comment:  comment.String,
			Position: int32(position),
		}
		columns = append(columns, col)
	}

	return columns, nil
}

func getOracleIndexes(ctx context.Context, db *sql.DB, tableName string) ([]*storepb.IndexMetadata, error) {
	query := `
		SELECT 
			i.INDEX_NAME,
			LISTAGG(ic.COLUMN_NAME, ',') WITHIN GROUP (ORDER BY ic.COLUMN_POSITION) as COLUMN_NAMES,
			i.UNIQUENESS,
			CASE WHEN c.CONSTRAINT_TYPE = 'P' THEN 'Y' ELSE 'N' END as IS_PRIMARY
		FROM USER_INDEXES i
		JOIN USER_IND_COLUMNS ic ON i.INDEX_NAME = ic.INDEX_NAME
		LEFT JOIN USER_CONSTRAINTS c ON i.INDEX_NAME = c.INDEX_NAME AND c.CONSTRAINT_TYPE = 'P'
		WHERE i.TABLE_NAME = :1
		GROUP BY i.INDEX_NAME, i.UNIQUENESS, c.CONSTRAINT_TYPE
		ORDER BY i.INDEX_NAME
	`
	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []*storepb.IndexMetadata
	for rows.Next() {
		var indexName, columnNames, uniqueness, isPrimaryStr string

		if err := rows.Scan(&indexName, &columnNames, &uniqueness, &isPrimaryStr); err != nil {
			return nil, err
		}

		idx := &storepb.IndexMetadata{
			Name:        indexName,
			Expressions: strings.Split(columnNames, ","),
			Unique:      uniqueness == "UNIQUE",
			Primary:     isPrimaryStr == "Y",
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}
