// Package common provides shared utilities for demo examples
package common

import (
	"context"
	"fmt"

	"advisorTool/db"
	"advisorTool/pkg/advisor"
)

// DBConfig 数据库连接配置
type DBConfig struct {
	Host        string // 数据库主机地址
	Port        int    // 数据库端口
	User        string // 数据库用户名
	Password    string // 数据库密码
	DBName      string // 数据库名称
	Charset     string // 字符集（MySQL）
	ServiceName string // Oracle 服务名
	Sid         string // Oracle SID
	SSLMode     string // PostgreSQL SSL 模式
	Timeout     int    // 连接超时时间（秒）
}

// FetchDatabaseMetadata 从数据库获取元数据
func FetchDatabaseMetadata(engineType advisor.Engine, dbConfig *DBConfig) (*advisor.DatabaseSchemaMetadata, error) {
	if dbConfig == nil || dbConfig.Host == "" {
		return nil, nil // 没有配置数据库连接，返回 nil（静态分析模式）
	}

	// 构建连接配置
	config := &db.ConnectionConfig{
		DbType:      getDbTypeString(engineType),
		Host:        dbConfig.Host,
		Port:        dbConfig.Port,
		User:        dbConfig.User,
		Password:    dbConfig.Password,
		DbName:      dbConfig.DBName,
		Charset:     dbConfig.Charset,
		ServiceName: dbConfig.ServiceName,
		Sid:         dbConfig.Sid,
		SSLMode:     dbConfig.SSLMode,
		Timeout:     dbConfig.Timeout,
	}

	ctx := context.Background()

	// 打开数据库连接
	conn, err := db.OpenConnection(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	defer conn.Close()

	// 获取元数据
	metadata, err := db.GetDatabaseMetadata(ctx, conn, config)
	if err != nil {
		return nil, fmt.Errorf("获取数据库元数据失败: %w", err)
	}

	fmt.Printf("✅ 成功连接数据库并获取元数据 (%s@%s:%d/%s)\n",
		dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	return metadata, nil
}

// GetDefaultRules 返回指定数据库引擎的默认规则集
// hasMetadata 表示是否有数据库元数据（会影响部分规则的启用）
func GetDefaultRules(engineType advisor.Engine, hasMetadata bool) []*advisor.SQLReviewRule {
	type ruleConfig struct {
		ruleType string
		level    advisor.RuleLevel
	}

	var ruleConfigs []ruleConfig

	// 通用规则（所有数据库都支持，不需要 metadata）
	commonErrorRules := []string{
		advisor.RuleStatementRequireWhereForUpdateDelete, // UPDATE/DELETE 必须有 WHERE
		advisor.RuleTableRequirePK,                       // 表必须有主键
	}

	commonWarningRules := []string{
		advisor.RuleStatementNoSelectAll,           // 禁止 SELECT *
		advisor.RuleStatementRequireWhereForSelect, // SELECT 需要 WHERE
		advisor.RuleTableNoFK,                      // 表禁止外键
	}

	// 添加通用规则
	for _, r := range commonErrorRules {
		ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelError})
	}
	for _, r := range commonWarningRules {
		ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
	}

	// 根据数据库类型添加特定规则
	switch engineType {
	case advisor.EngineMySQL, advisor.EngineMariaDB, advisor.EngineOceanBase, advisor.EngineTiDB:
		// MySQL 系列特有规则（不需要 metadata）
		mysqlErrorRules := []string{
			advisor.RuleColumnAutoIncrementMustInteger, // 自增列必须是整数
			advisor.RuleIndexNoDuplicateColumn,         // 索引不能有重复列
		}
		mysqlWarningRules := []string{
			advisor.RuleStatementNoLeadingWildcardLike,     // LIKE 不能以 % 开头
			advisor.RuleStatementDisallowCommit,            // 禁止 COMMIT 语句
			advisor.RuleStatementDisallowLimit,             // 禁止 LIMIT 子句
			advisor.RuleStatementDisallowOrderBy,           // 禁止 ORDER BY 子句
			advisor.RuleStatementMergeAlterTable,           // 合并 ALTER TABLE 语句
			advisor.RuleStatementInsertMustSpecifyColumn,   // INSERT 必须指定列
			advisor.RuleStatementInsertDisallowOrderByRand, // INSERT 禁止 ORDER BY RAND
			advisor.RuleStatementWhereNoEqualNull,          // WHERE 不能使用 = NULL
			advisor.RuleTableDisallowPartition,             // 禁止分区表
			advisor.RuleTableNoDuplicateIndex,              // 禁止重复索引
			advisor.RuleColumnDisallowChange,               // 禁止 CHANGE COLUMN
			advisor.RuleColumnAutoIncrementMustUnsigned,    // 自增列必须无符号
			advisor.RuleIndexTypeNoBlob,                    // 索引不能包含 BLOB
			advisor.RuleProcedureDisallowCreate,            // 禁止创建存储过程
			advisor.RuleEventDisallowCreate,                // 禁止创建事件
			advisor.RuleViewDisallowCreate,                 // 禁止创建视图
			advisor.RuleFunctionDisallowCreate,             // 禁止创建函数
		}

		// MySQL 系列需要 metadata 的规则
		mysqlMetadataWarningRules := []string{
			advisor.RuleColumnNotNull,               // 列不能为 NULL（需要检查现有表结构）
			advisor.RuleColumnSetDefaultForNotNull,  // NOT NULL 列需要默认值
			advisor.RuleColumnRequireDefault,        // 列需要默认值
			advisor.RuleSchemaBackwardCompatibility, // 向后兼容
		}

		for _, r := range mysqlErrorRules {
			ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelError})
		}
		for _, r := range mysqlWarningRules {
			ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
		}
		// 只有在有 metadata 时才添加需要 metadata 的规则
		if hasMetadata {
			for _, r := range mysqlMetadataWarningRules {
				ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
			}
		}

	case advisor.EnginePostgres:
		// PostgreSQL 特有规则（不需要 metadata）
		pgWarningRules := []string{
			advisor.RuleStatementNoLeadingWildcardLike,        // LIKE 不能以 % 开头
			advisor.RuleStatementDisallowAddColumnWithDefault, // 禁止添加带默认值的列
			advisor.RuleStatementAddCheckNotValid,             // ADD CHECK 使用 NOT VALID
			advisor.RuleStatementAddFKNotValid,                // ADD FK 使用 NOT VALID
			advisor.RuleStatementDisallowAddNotNull,           // 禁止添加 NOT NULL
			advisor.RuleStatementNonTransactional,             // 检查非事务语句
			advisor.RuleStatementCreateSpecifySchema,          // 创建时指定 schema
			advisor.RuleStatementDisallowCommit,               // 禁止 COMMIT
			advisor.RuleStatementMergeAlterTable,              // 合并 ALTER TABLE
			advisor.RuleStatementInsertMustSpecifyColumn,      // INSERT 必须指定列
			advisor.RuleColumnDefaultDisallowVolatile,         // 默认值不能是易变函数
			advisor.RuleCreateIndexConcurrently,               // 并发创建索引
			advisor.RuleTableDisallowPartition,                // 禁止分区表
		}

		// PostgreSQL 需要 metadata 的规则
		pgMetadataWarningRules := []string{
			advisor.RuleColumnNotNull,               // 列不能为 NULL
			advisor.RuleColumnRequireDefault,        // 列需要默认值
			advisor.RuleSchemaBackwardCompatibility, // 向后兼容
		}

		for _, r := range pgWarningRules {
			ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
		}
		if hasMetadata {
			for _, r := range pgMetadataWarningRules {
				ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
			}
		}

	case advisor.EngineMSSQL:
		// SQL Server 特有规则（不需要 metadata）
		mssqlWarningRules := []string{
			advisor.RuleStatementDisallowCrossDBQueries, // 禁止跨库查询
		}

		// SQL Server 需要 metadata 的规则
		mssqlMetadataWarningRules := []string{
			advisor.RuleColumnNotNull,     // 列不能为 NULL
			advisor.RuleIndexNotRedundant, // 索引不能冗余
		}
		for _, r := range mssqlWarningRules {
			ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
		}
		if hasMetadata {
			for _, r := range mssqlMetadataWarningRules {
				ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
			}
		}

	case advisor.EngineOracle:
		// Oracle 特有规则（不需要 metadata）
		oracleWarningRules := []string{
			advisor.RuleStatementNoLeadingWildcardLike,   // LIKE 不能以 % 开头
			advisor.RuleStatementInsertMustSpecifyColumn, // INSERT 必须指定列
		}
		// Oracle 需要 metadata 的规则
		oracleMetadataWarningRules := []string{
			advisor.RuleColumnNotNull,        // 列不能为 NULL
			advisor.RuleColumnRequireDefault, // 列需要默认值
		}
		for _, r := range oracleWarningRules {
			ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
		}
		if hasMetadata {
			for _, r := range oracleMetadataWarningRules {
				ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
			}
		}

	case advisor.EngineSnowflake:
		// Snowflake 需要 metadata 的规则
		snowflakeMetadataWarningRules := []string{
			advisor.RuleColumnNotNull, // 列不能为 NULL
		}
		if hasMetadata {
			for _, r := range snowflakeMetadataWarningRules {
				ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
			}
		}
	}

	var rules []*advisor.SQLReviewRule
	for _, rc := range ruleConfigs {
		rules = append(rules, &advisor.SQLReviewRule{
			Type:   rc.ruleType,
			Level:  rc.level,
			Engine: engineType,
		})
	}

	return rules
}

// getDbTypeString 将引擎类型转换为数据库类型字符串
func getDbTypeString(engineType advisor.Engine) string {
	switch engineType {
	case advisor.EngineMySQL:
		return "mysql"
	case advisor.EngineMariaDB:
		return "mariadb"
	case advisor.EngineTiDB:
		return "tidb"
	case advisor.EngineOceanBase:
		return "oceanbase"
	case advisor.EnginePostgres:
		return "postgres"
	case advisor.EngineMSSQL:
		return "mssql"
	case advisor.EngineOracle:
		return "oracle"
	case advisor.EngineSnowflake:
		return "snowflake"
	default:
		return "mysql"
	}
}
