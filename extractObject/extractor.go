package extractobject

import (
	"context"
	"fmt"

	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
)

// Extractor SQL表名提取器接口
type Extractor interface {
	Extract(ctx context.Context, sql string) ([]TableInfo, error)
}

// ExtractTables 从SQL语句中提取表名信息
// dbType: 数据库类型
// sql: SQL语句
// 返回: 表信息列表和错误
func ExtractTables(dbType DBType, sql string) ([]TableInfo, error) {
	return ExtractTablesWithContext(context.Background(), dbType, sql)
}

// ExtractTablesWithContext 从SQL语句中提取表名信息(带上下文)
func ExtractTablesWithContext(ctx context.Context, dbType DBType, sql string) ([]TableInfo, error) {
	engine, err := dbTypeToEngine(dbType)
	if err != nil {
		return nil, err
	}

	extractor, err := getExtractor(engine)
	if err != nil {
		return nil, err
	}

	return extractor.Extract(ctx, sql)
}

// dbTypeToEngine 将DBType转换为storepb.Engine
func dbTypeToEngine(dbType DBType) (storepb.Engine, error) {
	switch dbType {
	case MySQL:
		return storepb.Engine_MYSQL, nil
	case PostgreSQL:
		return storepb.Engine_POSTGRES, nil
	case Oracle:
		return storepb.Engine_ORACLE, nil
	case SQLServer:
		return storepb.Engine_MSSQL, nil
	case TiDB:
		return storepb.Engine_TIDB, nil
	case MariaDB:
		return storepb.Engine_MARIADB, nil
	case OceanBase:
		return storepb.Engine_OCEANBASE, nil
	case Snowflake:
		return storepb.Engine_SNOWFLAKE, nil
	default:
		return storepb.Engine_ENGINE_UNSPECIFIED, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
}

// getExtractor 根据数据库引擎获取对应的提取器
func getExtractor(engine storepb.Engine) (Extractor, error) {
	switch engine {
	case storepb.Engine_MYSQL, storepb.Engine_MARIADB, storepb.Engine_OCEANBASE:
		return &MySQLExtractor{}, nil
	case storepb.Engine_POSTGRES:
		return &PostgreSQLExtractor{}, nil
	case storepb.Engine_ORACLE:
		return &OracleExtractor{}, nil
	case storepb.Engine_MSSQL:
		return &SQLServerExtractor{}, nil
	case storepb.Engine_TIDB:
		return &TiDBExtractor{}, nil
	case storepb.Engine_SNOWFLAKE:
		return &SnowflakeExtractor{}, nil
	default:
		return nil, fmt.Errorf("不支持的数据库引擎: %s", engine)
	}
}

// parseSQL 通用SQL解析函数
func parseSQL(engine storepb.Engine, sql string) ([]base.AST, error) {
	asts, err := base.Parse(engine, sql)
	if err != nil {
		return nil, fmt.Errorf("解析SQL失败: %w", err)
	}
	return asts, nil
}


