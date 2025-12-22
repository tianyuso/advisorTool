// Package main demonstrates using advisorTool as a Go library with all rules supported by the engine.
// This example shows how to load all rules for a specific database engine (common + engine-specific rules).
//
// 使用方式:
//
//  1. 不连接数据库（只测试不需要元数据的规则）:
//     go run library_usage_all_rules.go -engine mysql
//
//  2. 连接数据库（测试所有规则，包括需要元数据的规则）:
//     go run library_usage_all_rules.go -engine mysql -host 127.0.0.1 -port 3306 -user root -password root -dbname mydata
//     go run library_usage_all_rules.go -engine postgres -host 127.0.0.1 -port 5432 -user postgres -password secret -dbname mydb
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/tianyuso/advisorTool/db"
	"github.com/tianyuso/advisorTool/pkg/advisor"
)

var (
	// 数据库引擎
	engineFlag = flag.String("engine", "mysql", "数据库引擎: mysql, postgres, tidb, oracle, mssql, snowflake, mariadb, oceanbase")

	// 数据库连接参数（可选，用于获取元数据以测试需要元数据的规则）
	dbHost     = flag.String("host", "", "数据库主机地址")
	dbPort     = flag.Int("port", 0, "数据库端口")
	dbUser     = flag.String("user", "", "数据库用户名")
	dbPassword = flag.String("password", "", "数据库密码")
	dbName     = flag.String("dbname", "", "数据库名称")
	dbCharset  = flag.String("charset", "utf8mb4", "字符集（MySQL）")
	dbSSLMode  = flag.String("sslmode", "disable", "SSL 模式（PostgreSQL）")
	dbTimeout  = flag.Int("timeout", 5, "连接超时时间（秒）")
)

func main() {
	flag.Parse()

	// 解析数据库引擎
	engine := advisor.EngineFromString(*engineFlag)
	if engine == 0 {
		fmt.Fprintf(os.Stderr, "不支持的数据库引擎: %s\n", *engineFlag)
		fmt.Fprintln(os.Stderr, "支持的引擎: mysql, postgres, tidb, oracle, mssql, snowflake, mariadb, oceanbase")
		os.Exit(1)
	}

	fmt.Printf("=== SQL Advisor 库使用示例 - 全规则验证 ===\n")
	fmt.Printf("数据库引擎: %s\n", engine)

	// 检查是否提供了数据库连接参数
	hasMetadata := false
	var metadata *advisor.DatabaseSchemaMetadata

	if *dbHost != "" && *dbPort > 0 {
		fmt.Printf("数据库连接: %s@%s:%d/%s\n", *dbUser, *dbHost, *dbPort, *dbName)
		fmt.Println("正在获取数据库元数据...")

		// 获取元数据
		var err error
		metadata, err = fetchDatabaseMetadata(engine, *dbHost, *dbPort, *dbUser, *dbPassword, *dbName, *dbCharset, *dbSSLMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  获取数据库元数据失败: %v\n", err)
			fmt.Fprintln(os.Stderr, "将只测试不需要元数据的规则")
		} else {
			hasMetadata = true
			fmt.Printf("✅ 成功获取数据库元数据（%d 个 schema）\n", len(metadata.Schemas))
		}
	} else {
		fmt.Println("未提供数据库连接参数，将只测试不需要元数据的规则")
		fmt.Println("提示: 使用 -host, -port, -user, -password, -dbname 参数连接数据库以测试更多规则")
	}
	fmt.Println()

	// 测试 SQL 语句（包含多种违规情况）
	testSQL := getTestSQL(engine)

	// 获取该引擎支持的所有规则（通用规则 + 引擎专属规则）
	rules := getAllRulesForEngine(engine, hasMetadata)

	fmt.Printf("该引擎支持的规则数: %d\n", len(rules))
	if hasMetadata {
		// 统计不需要和需要元数据的规则
		noMetadataCount := len(getAllRulesForEngine(engine, false))
		metadataCount := len(rules) - noMetadataCount
		fmt.Printf("  - 不需要元数据: %d 条\n", noMetadataCount)
		fmt.Printf("  - 需要元数据: %d 条\n", metadataCount)
	}

	fmt.Println("\n=== 规则列表 ===")
	for i, rule := range rules {
		desc := advisor.GetRuleDescription(rule.Type)
		fmt.Printf("%2d. [%7s] %s\n", i+1, rule.Level, desc)
	}
	fmt.Println()

	// 创建审核请求
	req := &advisor.ReviewRequest{
		Engine:          engine,
		Statement:       testSQL,
		Rules:           rules,
		CurrentDatabase: *dbName,
		DBSchema:        metadata, // 如果有元数据就传入
	}

	// 执行审核
	fmt.Println("=== 执行审核 ===")
	ctx := context.Background()
	resp, err := advisor.SQLReviewCheck(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 审核失败: %v\n", err)
		os.Exit(1)
	}

	// 统计结果
	errorCount := 0
	warningCount := 0
	successCount := 0

	for _, advice := range resp.Advices {
		switch advice.Status {
		case advisor.AdviceStatusError:
			errorCount++
		case advisor.AdviceStatusWarning:
			warningCount++
		case advisor.AdviceStatusSuccess:
			successCount++
		}
	}

	// 输出统计信息
	fmt.Printf("发现问题数: %d\n", len(resp.Advices))
	fmt.Printf("  - 错误 (ERROR): %d\n", errorCount)
	fmt.Printf("  - 警告 (WARNING): %d\n", warningCount)
	fmt.Printf("  - 成功 (SUCCESS): %d\n", successCount)
	fmt.Println()

	// 输出详细问题列表（最多显示前 20 个）
	if len(resp.Advices) > 0 {
		fmt.Println("=== 详细问题列表（前 20 个）===")
		maxDisplay := 20
		if len(resp.Advices) < maxDisplay {
			maxDisplay = len(resp.Advices)
		}

		for i := 0; i < maxDisplay; i++ {
			advice := resp.Advices[i]
			statusIcon := getStatusIcon(advice.Status)

			fmt.Printf("\n%d. %s [%s] %s\n", i+1, statusIcon, advice.Status, advice.Title)
			fmt.Printf("   内容: %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("   位置: 行 %d, 列 %d\n",
					advice.StartPosition.Line,
					advice.StartPosition.Column)
			}
		}

		if len(resp.Advices) > maxDisplay {
			fmt.Printf("\n... 还有 %d 个问题未显示\n", len(resp.Advices)-maxDisplay)
		}
		fmt.Println()
	}

	// 按规则类型分组显示
	fmt.Println("=== 问题分类统计（前 10 类）===")
	ruleTypeCount := make(map[string]int)
	for _, advice := range resp.Advices {
		ruleTypeCount[advice.Title]++
	}

	// 转换为切片并排序
	type ruleCount struct {
		name  string
		count int
	}
	var ruleCounts []ruleCount
	for name, count := range ruleTypeCount {
		ruleCounts = append(ruleCounts, ruleCount{name, count})
	}

	// 简单排序（按数量降序）
	for i := 0; i < len(ruleCounts); i++ {
		for j := i + 1; j < len(ruleCounts); j++ {
			if ruleCounts[j].count > ruleCounts[i].count {
				ruleCounts[i], ruleCounts[j] = ruleCounts[j], ruleCounts[i]
			}
		}
	}

	maxShow := 10
	if len(ruleCounts) < maxShow {
		maxShow = len(ruleCounts)
	}
	for i := 0; i < maxShow; i++ {
		fmt.Printf("  %d. %s: %d 次\n", i+1, ruleCounts[i].name, ruleCounts[i].count)
	}
	fmt.Println()

	// 最终结论
	fmt.Println("=== 审核结论 ===")
	if resp.HasError {
		fmt.Println("❌ 审核失败：发现错误级别的问题")
		os.Exit(2)
	} else if resp.HasWarning {
		fmt.Println("⚠️  审核通过但有警告")
		os.Exit(1)
	} else {
		fmt.Println("✅ 审核完全通过")
		os.Exit(0)
	}
}

// getAllRulesForEngine 获取指定数据库引擎支持的所有规则
// 这个函数模仿了 cmd/advisor/internal/config.go 中的 GetDefaultRules 逻辑
func getAllRulesForEngine(engine advisor.Engine, hasMetadata bool) []*advisor.SQLReviewRule {
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
	switch engine {
	case advisor.EngineMySQL, advisor.EngineMariaDB, advisor.EngineOceanBase, advisor.EngineTiDB:
		// MySQL 系列特有规则（不需要 metadata）
		mysqlErrorRules := []string{
			advisor.RuleMySQLEngine,                    // 要求使用 InnoDB
			advisor.RuleColumnAutoIncrementMustInteger, // 自增列必须是整数
			advisor.RuleIndexNoDuplicateColumn,         // 索引不能有重复列
		}
		mysqlWarningRules := []string{
			advisor.RuleStatementNoLeadingWildcardLike,     // LIKE 不能以 % 开头
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
			advisor.RuleProcedureDisallowCreate,            // 禁止创建存储过程
			advisor.RuleEventDisallowCreate,                // 禁止创建事件
			advisor.RuleViewDisallowCreate,                 // 禁止创建视图
			advisor.RuleFunctionDisallowCreate,             // 禁止创建函数
		}

		// MySQL 系列需要 metadata 的规则
		mysqlMetadataWarningRules := []string{
			advisor.RuleIndexTypeNoBlob,             // 索引不能包含 BLOB (需要元数据)
			advisor.RuleColumnNotNull,               // 列不能为 NULL
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
		// SQL Server 特有规则
		mssqlWarningRules := []string{
			advisor.RuleStatementDisallowCrossDBQueries, // 禁止跨库查询
		}
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
		// Oracle 特有规则
		oracleWarningRules := []string{
			advisor.RuleStatementNoLeadingWildcardLike,   // LIKE 不能以 % 开头
			advisor.RuleStatementInsertMustSpecifyColumn, // INSERT 必须指定列
		}
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
		// Snowflake 特有规则
		snowflakeWarningRules := []string{
			advisor.RuleStatementNoLeadingWildcardLike, // LIKE 不能以 % 开头
		}
		snowflakeMetadataWarningRules := []string{
			advisor.RuleColumnNotNull, // 列不能为 NULL
		}
		for _, r := range snowflakeWarningRules {
			ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
		}
		if hasMetadata {
			for _, r := range snowflakeMetadataWarningRules {
				ruleConfigs = append(ruleConfigs, ruleConfig{r, advisor.RuleLevelWarning})
			}
		}
	}

	// 转换为 SQLReviewRule
	var rules []*advisor.SQLReviewRule
	for _, rc := range ruleConfigs {
		rules = append(rules, &advisor.SQLReviewRule{
			Type:   rc.ruleType,
			Level:  rc.level,
			Engine: engine,
		})
	}

	return rules
}

// getTestSQL 根据数据库引擎返回测试 SQL
func getTestSQL(engine advisor.Engine) string {
	switch engine {
	case advisor.EngineMySQL, advisor.EngineMariaDB, advisor.EngineTiDB, advisor.EngineOceanBase:
		return `
-- 创建表：缺少主键、使用不规范的引擎
CREATE TABLE test_users (
	id INT AUTO_INCREMENT,
	name VARCHAR(255),
	email TEXT,
	status VARCHAR(50)
) ENGINE=MyISAM;

-- SELECT * 问题
SELECT * FROM test_users;

-- 缺少 WHERE 子句
DELETE FROM test_users;
UPDATE test_users SET status = 'active';

-- INSERT 问题：未指定列名
INSERT INTO test_users VALUES (1, 'test', 'test@example.com', 'active');

-- 创建外键（可能违反 no-foreign-key 规则）
ALTER TABLE test_users ADD id2 INT;
ALTER TABLE test_users ADD CONSTRAINT fk_test 
	FOREIGN KEY (id2) REFERENCES test_users(id);

-- 创建索引
CREATE INDEX idx_name ON test_users(name);

-- CHANGE COLUMN
ALTER TABLE test_users CHANGE COLUMN name user_name VARCHAR(255);

-- WHERE = NULL (错误)
SELECT * FROM test_users WHERE name = NULL;
`

	case advisor.EnginePostgres:
		return `
-- 创建表：缺少主键
CREATE TABLE users (
	id INTEGER,
	name VARCHAR(255),
	email TEXT,
	status VARCHAR(50)
);

-- SELECT * 问题
SELECT * FROM users;

-- 缺少 WHERE 子句
DELETE FROM users;
UPDATE users SET status = 'active';

-- INSERT 问题：未指定列名
INSERT INTO users VALUES (1, 'test', 'test@example.com', 'active');

-- 创建外键（可能违反 no-foreign-key 规则）
ALTER TABLE orders ADD CONSTRAINT fk_user 
	FOREIGN KEY (user_id) REFERENCES users(id);

-- 创建索引（应该使用 CONCURRENTLY）
CREATE INDEX idx_name ON users(name);

-- 添加列带默认值（可能在 PostgreSQL 中不推荐）
ALTER TABLE users ADD COLUMN created_at TIMESTAMP DEFAULT NOW();
`

	default:
		return `
-- 创建表：缺少主键
CREATE TABLE users (
	id INT,
	name VARCHAR(255),
	status VARCHAR(50)
);

-- SELECT * 问题
SELECT * FROM users;

-- 缺少 WHERE 子句
DELETE FROM users;
UPDATE users SET status = 'active';

-- INSERT 问题：未指定列名
INSERT INTO users VALUES (1, 'test', 'active');
`
	}
}

// fetchDatabaseMetadata 获取数据库元数据
func fetchDatabaseMetadata(engine advisor.Engine, host string, port int, user, password, dbName, charset, sslmode string) (*advisor.DatabaseSchemaMetadata, error) {
	var dbType string

	switch engine {
	case advisor.EngineMySQL:
		dbType = "mysql"
	case advisor.EngineMariaDB:
		dbType = "mariadb"
	case advisor.EngineTiDB:
		dbType = "tidb"
	case advisor.EngineOceanBase:
		dbType = "oceanbase"
	case advisor.EnginePostgres:
		dbType = "postgres"
	case advisor.EngineMSSQL:
		dbType = "mssql"
	case advisor.EngineOracle:
		dbType = "oracle"
	default:
		return nil, fmt.Errorf("不支持获取 %s 数据库的元数据", engine)
	}

	// 创建连接配置
	config := &db.ConnectionConfig{
		DbType:   dbType,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DbName:   dbName,
		Charset:  charset,
		SSLMode:  sslmode,
	}

	// 连接数据库
	ctx := context.Background()
	sqlDB, err := db.OpenConnection(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("无法连接数据库: %w", err)
	}
	defer sqlDB.Close()

	// 获取元数据
	metadata, err := db.GetDatabaseMetadata(ctx, sqlDB, config)
	if err != nil {
		return nil, fmt.Errorf("获取数据库元数据失败: %w", err)
	}

	return metadata, nil
}

// getStatusIcon 返回状态对应的图标
func getStatusIcon(status advisor.AdviceStatus) string {
	switch status {
	case advisor.AdviceStatusError:
		return "❌"
	case advisor.AdviceStatusWarning:
		return "⚠️"
	case advisor.AdviceStatusSuccess:
		return "✅"
	default:
		return "ℹ️"
	}
}
