// Package main provides the CLI entry point for the SQL advisor tool.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"advisorTool/db"
	"advisorTool/pkg/advisor"
)

var (
	configFile     = flag.String("config", "", "Path to the review config file (YAML or JSON)")
	engine         = flag.String("engine", "", "Database engine: mysql, postgres, tidb, oracle, mssql, snowflake, mariadb, oceanbase")
	sqlFile        = flag.String("file", "", "Path to the SQL file to review")
	sql            = flag.String("sql", "", "SQL statement to review (use - to read from stdin)")
	outputFormat   = flag.String("format", "text", "Output format: text, json, yaml")
	listRules      = flag.Bool("list-rules", false, "List all available rules")
	generateConfig = flag.Bool("generate-config", false, "Generate a sample config file for the specified engine")
	version        = flag.Bool("version", false, "Print version information")

	// Database connection parameters
	dbHost        = flag.String("host", "", "Database host address")
	dbPort        = flag.Int("port", 0, "Database port")
	dbUser        = flag.String("user", "", "Database username")
	dbPassword    = flag.String("password", "", "Database password")
	dbName        = flag.String("dbname", "", "Database name")
	dbCharset     = flag.String("charset", "", "Database charset (default: utf8mb4 for MySQL)")
	dbServiceName = flag.String("service-name", "", "Oracle service name")
	dbSid         = flag.String("sid", "", "Oracle SID")
	dbSSLMode     = flag.String("sslmode", "disable", "PostgreSQL SSL mode")
	dbTimeout     = flag.Int("timeout", 5, "Database connection timeout in seconds")
)

const toolVersion = "1.0.0"

// ReviewConfig represents a review configuration file.
type ReviewConfig struct {
	Name  string             `json:"name" yaml:"name"`
	Rules []*ReviewRuleEntry `json:"rules" yaml:"rules"`
}

// ReviewRuleEntry represents a single rule entry in the config file.
type ReviewRuleEntry struct {
	Type    string `json:"type" yaml:"type"`
	Level   string `json:"level" yaml:"level"`
	Payload string `json:"payload,omitempty" yaml:"payload,omitempty"`
	Engine  string `json:"engine,omitempty" yaml:"engine,omitempty"`
	Comment string `json:"comment,omitempty" yaml:"comment,omitempty"`
}

// ReviewResult represents the review result in Inception-compatible format.
type ReviewResult struct {
	OrderID      int    `json:"order_id"`
	Stage        string `json:"stage"`
	ErrorLevel   string `json:"error_level"`
	StageStatus  string `json:"stage_status"`
	ErrorMessage string `json:"error_message"`
	SQL          string `json:"sql"`
	AffectedRows int    `json:"affected_rows"`
	Sequence     string `json:"sequence"`
	BackupDBName string `json:"backup_dbname"`
	ExecuteTime  string `json:"execute_time"`
	SQLSha1      string `json:"sqlsha1"`
	BackupTime   string `json:"backup_time"`
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("SQL Advisor Tool v%s\n", toolVersion)
		fmt.Println("Based on Bytebase SQL Review Engine")
		fmt.Println("Supported engines: mysql, postgres, tidb, oracle, mssql, snowflake, mariadb, oceanbase")
		os.Exit(0)
	}

	if *listRules {
		listAvailableRules()
		os.Exit(0)
	}

	if *engine == "" {
		fmt.Fprintln(os.Stderr, "Error: -engine flag is required")
		flag.Usage()
		os.Exit(1)
	}

	engineType := advisor.EngineFromString(*engine)
	if engineType == 0 {
		fmt.Fprintf(os.Stderr, "Error: unsupported engine: %s\n", *engine)
		fmt.Fprintln(os.Stderr, "Supported engines: mysql, postgres, tidb, oracle, mssql, snowflake, mariadb, oceanbase")
		os.Exit(1)
	}

	if *generateConfig {
		generateSampleConfig(engineType)
		os.Exit(0)
	}

	// Get SQL statement
	statement, err := getStatement()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SQL: %v\n", err)
		os.Exit(1)
	}

	if statement == "" {
		fmt.Fprintln(os.Stderr, "Error: no SQL statement provided. Use -sql or -file flag")
		flag.Usage()
		os.Exit(1)
	}

	// Prepare review request
	req := &advisor.ReviewRequest{
		Engine:          engineType,
		Statement:       statement,
		CurrentDatabase: *dbName,
	}

	// Check if database connection parameters are provided
	hasMetadata := false
	if *dbHost != "" && *dbPort > 0 {
		metadata, err := fetchDatabaseMetadata(engineType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to fetch database metadata: %v\n", err)
			fmt.Fprintf(os.Stderr, "Some rules that require metadata will be skipped.\n")
		} else {
			req.DBSchema = metadata
			hasMetadata = true
		}
	}

	// Load review rules (pass hasMetadata to enable/disable metadata-dependent rules)
	rules, err := loadRules(engineType, hasMetadata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading rules: %v\n", err)
		os.Exit(1)
	}
	req.Rules = rules

	// Perform review
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during review: %v\n", err)
		os.Exit(1)
	}

	// Output results
	if err := outputResults(resp, statement); err != nil {
		fmt.Fprintf(os.Stderr, "Error outputting results: %v\n", err)
		os.Exit(1)
	}

	// Exit with error code if there are errors
	if resp.HasError {
		os.Exit(2)
	}
	if resp.HasWarning {
		os.Exit(1)
	}
}

func getStatement() (string, error) {
	if *sql != "" {
		if *sql == "-" {
			// Read from stdin
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("failed to read from stdin: %w", err)
			}
			return string(data), nil
		}
		return *sql, nil
	}

	if *sqlFile != "" {
		data, err := os.ReadFile(*sqlFile)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", *sqlFile, err)
		}
		return string(data), nil
	}

	return "", nil
}

func loadRules(engineType advisor.Engine, hasMetadata bool) ([]*advisor.SQLReviewRule, error) {
	if *configFile == "" {
		// Use default rules
		return getDefaultRules(engineType, hasMetadata), nil
	}

	data, err := os.ReadFile(*configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ReviewConfig
	if strings.HasSuffix(*configFile, ".yaml") || strings.HasSuffix(*configFile, ".yml") {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	}

	// Convert to SQLReviewRule
	var rules []*advisor.SQLReviewRule
	for _, entry := range config.Rules {
		level := advisor.RuleLevelFromString(entry.Level)
		if level == 0 {
			level = advisor.RuleLevelWarning
		}

		rule := &advisor.SQLReviewRule{
			Type:    entry.Type,
			Level:   level,
			Payload: entry.Payload,
			Comment: entry.Comment,
		}

		if entry.Engine != "" {
			rule.Engine = advisor.EngineFromString(entry.Engine)
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// getDefaultRules returns default rules based on engine type and whether metadata is available.
// hasMetadata indicates if database metadata is provided (some rules require it).
func getDefaultRules(engineType advisor.Engine, hasMetadata bool) []*advisor.SQLReviewRule {
	// 根据数据库类型返回该数据库支持的规则
	// 不同数据库实现的规则不同，这里只启用已确认支持的规则

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

func outputResults(resp *advisor.ReviewResponse, statement string) error {
	switch *outputFormat {
	case "json":
		results := convertToReviewResults(resp, statement)
		data, err := json.Marshal(results)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(resp)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		// Text format
		if len(resp.Advices) == 0 {
			fmt.Println("✅ No issues found!")
			return nil
		}

		fmt.Printf("Found %d issue(s):\n\n", len(resp.Advices))
		for i, advice := range resp.Advices {
			icon := "⚠️"
			statusStr := "WARNING"
			if advice.Status == advisor.AdviceStatusError {
				icon = "❌"
				statusStr = "ERROR"
			}
			fmt.Printf("%d. %s [%s] %s\n", i+1, icon, statusStr, advice.Title)
			fmt.Printf("   %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("   Location: line %d, column %d\n", advice.StartPosition.Line, advice.StartPosition.Column)
			}
			fmt.Println()
		}
	}
	return nil
}

// convertToReviewResults converts advisor response to Inception-compatible format.
func convertToReviewResults(resp *advisor.ReviewResponse, statement string) []ReviewResult {
	// Split SQL statements by semicolon
	sqlStatements := splitSQL(statement)

	// If no issues found, return success for each statement
	if len(resp.Advices) == 0 {
		var results []ReviewResult
		for i, sql := range sqlStatements {
			results = append(results, ReviewResult{
				OrderID:      i + 1,
				Stage:        "CHECKED",
				ErrorLevel:   "0",
				StageStatus:  "Audit Completed",
				ErrorMessage: "",
				SQL:          strings.TrimSpace(sql),
				AffectedRows: 0,
				Sequence:     fmt.Sprintf("0_0_%08d", i),
				BackupDBName: "",
				ExecuteTime:  "0",
				SQLSha1:      "",
				BackupTime:   "0",
			})
		}
		return results
	}

	// Group advices by SQL statement (using line number)
	advicesBySQL := make(map[int][]*advisor.Advice)
	for _, advice := range resp.Advices {
		line := 1
		if advice.StartPosition != nil {
			line = int(advice.StartPosition.Line)
		}
		// Find which SQL statement this line belongs to
		sqlIndex := findSQLIndexByLine(sqlStatements, statement, line)
		advicesBySQL[sqlIndex] = append(advicesBySQL[sqlIndex], advice)
	}

	var results []ReviewResult
	for i, sql := range sqlStatements {
		advices := advicesBySQL[i]

		errorLevel := "0"
		stageStatus := "Audit Completed"
		var errorMessages []string

		for _, advice := range advices {
			switch advice.Status {
			case advisor.AdviceStatusError:
				errorLevel = "2"
			case advisor.AdviceStatusWarning:
				if errorLevel != "2" {
					errorLevel = "1"
				}
			}
			// Collect error messages: [rule_type] message
			errorMessages = append(errorMessages, fmt.Sprintf("[%s] %s", advice.Title, advice.Content))
		}

		results = append(results, ReviewResult{
			OrderID:      i + 1,
			Stage:        "CHECKED",
			ErrorLevel:   errorLevel,
			StageStatus:  stageStatus,
			ErrorMessage: strings.Join(errorMessages, "\n"),
			SQL:          strings.TrimSpace(sql),
			AffectedRows: 0,
			Sequence:     fmt.Sprintf("0_0_%08d", i),
			BackupDBName: "",
			ExecuteTime:  "0",
			SQLSha1:      "",
			BackupTime:   "0",
		})
	}

	return results
}

// splitSQL splits SQL statements by semicolon.
func splitSQL(statement string) []string {
	// Simple split by semicolon - handles most cases
	parts := strings.Split(statement, ";")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		// If no semicolon found, treat entire statement as one SQL
		result = append(result, strings.TrimSpace(statement))
	}
	return result
}

// findSQLIndexByLine finds which SQL statement a line belongs to.
func findSQLIndexByLine(sqlStatements []string, fullStatement string, line int) int {
	if len(sqlStatements) <= 1 {
		return 0
	}

	// Count lines to find which statement the line belongs to
	lines := strings.Split(fullStatement, "\n")
	currentLine := 1
	currentSQLIndex := 0
	currentSQLStart := 0

	for i, l := range lines {
		if i > 0 {
			currentLine++
		}

		// Check if this line contains a semicolon (end of statement)
		if strings.Contains(l, ";") {
			if line >= currentSQLStart && line <= currentLine {
				return currentSQLIndex
			}
			currentSQLIndex++
			currentSQLStart = currentLine + 1
			if currentSQLIndex >= len(sqlStatements) {
				currentSQLIndex = len(sqlStatements) - 1
			}
		}
	}

	// If not found, return the last index or 0
	if line >= currentSQLStart {
		return currentSQLIndex
	}
	return 0
}

func listAvailableRules() {
	fmt.Println("Available SQL Review Rules:")
	fmt.Println("===========================\n")

	rules := advisor.AllRules()
	for _, ruleType := range rules {
		desc := advisor.GetRuleDescription(ruleType)
		fmt.Printf("  %s\n    %s\n\n", ruleType, desc)
	}

	fmt.Printf("\nTotal: %d rules\n", len(rules))
}

func generateSampleConfig(engineType advisor.Engine) {
	// Build a comprehensive config based on engine
	config := ReviewConfig{
		Name: fmt.Sprintf("%s-review-config", strings.ToLower(engineType.String())),
	}

	// Common rules for all engines
	commonRules := []ReviewRuleEntry{
		{Type: advisor.RuleStatementNoSelectAll, Level: "WARNING", Comment: "Disallow SELECT *"},
		{Type: advisor.RuleStatementRequireWhereForSelect, Level: "WARNING", Comment: "Require WHERE for SELECT"},
		{Type: advisor.RuleStatementRequireWhereForUpdateDelete, Level: "ERROR", Comment: "Require WHERE for UPDATE/DELETE"},
		{Type: advisor.RuleTableRequirePK, Level: "ERROR", Comment: "Require primary key"},
		{Type: advisor.RuleTableNoFK, Level: "WARNING", Comment: "Disallow foreign key"},
		{Type: advisor.RuleColumnNotNull, Level: "WARNING", Comment: "Columns no NULL value"},
		{Type: advisor.RuleRequiredColumn, Level: "ERROR", Payload: `{"list":["id","created_at","updated_at"]}`, Comment: "Required columns"},
	}

	config.Rules = make([]*ReviewRuleEntry, len(commonRules))
	for i := range commonRules {
		config.Rules[i] = &commonRules[i]
	}

	// Engine-specific rules
	switch engineType {
	case advisor.EngineMySQL, advisor.EngineMariaDB:
		config.Rules = append(config.Rules,
			&ReviewRuleEntry{Type: advisor.RuleMySQLEngine, Level: "ERROR", Comment: "Require InnoDB engine"},
			&ReviewRuleEntry{Type: advisor.RuleTableNaming, Level: "WARNING", Payload: `{"format":"^[a-z][a-z0-9_]*$","maxLength":64}`, Comment: "Table naming convention"},
			&ReviewRuleEntry{Type: advisor.RuleColumnNaming, Level: "WARNING", Payload: `{"format":"^[a-z][a-z0-9_]*$","maxLength":64}`, Comment: "Column naming convention"},
			&ReviewRuleEntry{Type: advisor.RuleIDXNaming, Level: "WARNING", Payload: `{"format":"^idx_{{table}}_{{column_list}}$","maxLength":64}`, Comment: "Index naming convention"},
			&ReviewRuleEntry{Type: advisor.RuleColumnAutoIncrementMustInteger, Level: "ERROR", Comment: "Auto-increment must be integer"},
			&ReviewRuleEntry{Type: advisor.RuleColumnAutoIncrementMustUnsigned, Level: "WARNING", Comment: "Auto-increment must be unsigned"},
			&ReviewRuleEntry{Type: advisor.RuleSchemaBackwardCompatibility, Level: "ERROR", Comment: "Backward compatible schema change"},
			&ReviewRuleEntry{Type: advisor.RuleCharsetAllowlist, Level: "WARNING", Payload: `{"list":["utf8mb4","utf8"]}`, Comment: "Charset allowlist"},
		)
	case advisor.EnginePostgres:
		config.Rules = append(config.Rules,
			&ReviewRuleEntry{Type: advisor.RuleFullyQualifiedObjectName, Level: "WARNING", Comment: "Require fully qualified names"},
			&ReviewRuleEntry{Type: advisor.RuleTableNaming, Level: "WARNING", Payload: `{"format":"^[a-z][a-z0-9_]*$","maxLength":63}`, Comment: "Table naming convention"},
			&ReviewRuleEntry{Type: advisor.RuleCreateIndexConcurrently, Level: "ERROR", Comment: "Create index concurrently"},
			&ReviewRuleEntry{Type: advisor.RuleStatementDisallowAddColumnWithDefault, Level: "WARNING", Comment: "Disallow ADD COLUMN with DEFAULT"},
			&ReviewRuleEntry{Type: advisor.RuleStatementAddCheckNotValid, Level: "WARNING", Comment: "Add CHECK with NOT VALID"},
			&ReviewRuleEntry{Type: advisor.RuleSchemaBackwardCompatibility, Level: "ERROR", Comment: "Backward compatible schema change"},
		)
	case advisor.EngineOracle:
		config.Rules = append(config.Rules,
			&ReviewRuleEntry{Type: advisor.RuleTableNaming, Level: "WARNING", Payload: `{"format":"^[A-Z][A-Z0-9_]*$","maxLength":30}`, Comment: "Table naming convention (uppercase)"},
			&ReviewRuleEntry{Type: advisor.RuleIdentifierNoKeyword, Level: "WARNING", Comment: "Disallow keywords as identifiers"},
			&ReviewRuleEntry{Type: advisor.RuleIdentifierCase, Level: "WARNING", Payload: `{"upper":true}`, Comment: "Identifier case (uppercase)"},
			&ReviewRuleEntry{Type: advisor.RuleColumnRequireDefault, Level: "WARNING", Comment: "Require column default value"},
		)
	case advisor.EngineMSSQL:
		config.Rules = append(config.Rules,
			&ReviewRuleEntry{Type: advisor.RuleTableNaming, Level: "WARNING", Payload: `{"format":"^[a-zA-Z][a-zA-Z0-9_]*$","maxLength":128}`, Comment: "Table naming convention"},
			&ReviewRuleEntry{Type: advisor.RuleIdentifierNoKeyword, Level: "WARNING", Comment: "Disallow keywords as identifiers"},
			&ReviewRuleEntry{Type: advisor.RuleSchemaBackwardCompatibility, Level: "ERROR", Comment: "Backward compatible schema change"},
			&ReviewRuleEntry{Type: advisor.RuleStatementDisallowCrossDBQueries, Level: "WARNING", Comment: "Disallow cross-database queries"},
		)
	case advisor.EngineTiDB:
		config.Rules = append(config.Rules,
			&ReviewRuleEntry{Type: advisor.RuleTableNaming, Level: "WARNING", Payload: `{"format":"^[a-z][a-z0-9_]*$","maxLength":64}`, Comment: "Table naming convention"},
			&ReviewRuleEntry{Type: advisor.RuleColumnNaming, Level: "WARNING", Payload: `{"format":"^[a-z][a-z0-9_]*$","maxLength":64}`, Comment: "Column naming convention"},
			&ReviewRuleEntry{Type: advisor.RuleSchemaBackwardCompatibility, Level: "ERROR", Comment: "Backward compatible schema change"},
		)
	case advisor.EngineSnowflake:
		config.Rules = append(config.Rules,
			&ReviewRuleEntry{Type: advisor.RuleTableNaming, Level: "WARNING", Payload: `{"format":"^[A-Z][A-Z0-9_]*$","maxLength":255}`, Comment: "Table naming convention (uppercase)"},
			&ReviewRuleEntry{Type: advisor.RuleIdentifierNoKeyword, Level: "WARNING", Comment: "Disallow keywords as identifiers"},
			&ReviewRuleEntry{Type: advisor.RuleIdentifierCase, Level: "WARNING", Payload: `{"upper":true}`, Comment: "Identifier case (uppercase)"},
		)
	case advisor.EngineOceanBase:
		config.Rules = append(config.Rules,
			&ReviewRuleEntry{Type: advisor.RuleTableNaming, Level: "WARNING", Payload: `{"format":"^[a-z][a-z0-9_]*$","maxLength":64}`, Comment: "Table naming convention"},
			&ReviewRuleEntry{Type: advisor.RuleStatementDisallowOfflineDDL, Level: "ERROR", Comment: "Disallow offline DDL"},
		)
	}

	// Output as YAML
	data, _ := yaml.Marshal(config)
	fmt.Println(string(data))
}

// fetchDatabaseMetadata fetches database schema metadata from the connected database.
func fetchDatabaseMetadata(engineType advisor.Engine) (*advisor.DatabaseSchemaMetadata, error) {
	// Build connection config
	config := &db.ConnectionConfig{
		DbType:      getDbTypeString(engineType),
		Host:        *dbHost,
		Port:        *dbPort,
		User:        *dbUser,
		Password:    *dbPassword,
		DbName:      *dbName,
		Charset:     *dbCharset,
		ServiceName: *dbServiceName,
		Sid:         *dbSid,
		SSLMode:     *dbSSLMode,
		Timeout:     *dbTimeout,
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

// getDbTypeString converts Engine type to database type string.
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
