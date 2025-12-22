// Package services provides utilities for the advisor tool that can be imported by external programs.
package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/tianyuso/advisorTool/pkg/advisor"
)

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

// LoadRules loads SQL review rules from a config file or returns default rules.
func LoadRules(configFile string, engineType advisor.Engine, hasMetadata bool) ([]*advisor.SQLReviewRule, error) {
	if configFile == "" {
		// Use default rules
		return GetDefaultRules(engineType, hasMetadata), nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ReviewConfig
	if strings.HasSuffix(configFile, ".yaml") || strings.HasSuffix(configFile, ".yml") {
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

// GetDefaultRules returns default rules based on engine type and whether metadata is available.
// hasMetadata indicates if database metadata is provided (some rules require it).
func GetDefaultRules(engineType advisor.Engine, hasMetadata bool) []*advisor.SQLReviewRule {
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

// GenerateSampleConfig generates a sample configuration file for the specified engine.
func GenerateSampleConfig(engineType advisor.Engine) string {
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
	return string(data)
}

