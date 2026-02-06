package extractobject

import "fmt"

// TableInfo 表示从SQL语句中提取的表信息
type TableInfo struct {
	DBName string // 数据库名
	Schema string // 模式名
	TBName string // 表名
	Alias  string // 别名
	IsCTE  bool   // 是否是CTE临时表
}

// DBType 数据库类型
type DBType string

const (
	MySQL      DBType = "mysql"
	PostgreSQL DBType = "postgres"
	Oracle     DBType = "oracle"
	SQLServer  DBType = "sqlserver"
	TiDB       DBType = "tidb"
	MariaDB    DBType = "mariadb"
	OceanBase  DBType = "oceanbase"
	Snowflake  DBType = "snowflake"
)

// ParseDBType 解析数据库类型字符串，支持大小写不敏感
func ParseDBType(s string) (DBType, error) {
	switch s {
	case "mysql", "MYSQL", "MySQL":
		return MySQL, nil
	case "postgres", "postgresql", "POSTGRES", "POSTGRESQL", "PostgreSQL":
		return PostgreSQL, nil
	case "oracle", "ORACLE", "Oracle":
		return Oracle, nil
	case "sqlserver", "SQLSERVER", "SQLServer", "mssql", "MSSQL":
		return SQLServer, nil
	case "tidb", "TIDB", "TiDB":
		return TiDB, nil
	case "mariadb", "MARIADB", "MariaDB":
		return MariaDB, nil
	case "oceanbase", "OCEANBASE", "OceanBase":
		return OceanBase, nil
	case "snowflake", "SNOWFLAKE", "Snowflake":
		return Snowflake, nil
	default:
		return "", fmt.Errorf("不支持的数据库类型: %s (支持: mysql, postgres, oracle, sqlserver, tidb, mariadb, oceanbase, snowflake)", s)
	}
}

func (d DBType) String() string {
	return string(d)
}
