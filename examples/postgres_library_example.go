// Package main demonstrates how to use the SQL Advisor library for PostgreSQL.
// This example connects to a PostgreSQL database and performs SQL review.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"advisorTool/db"
	"advisorTool/pkg/advisor"
	"advisorTool/services"
)

func main() {
	fmt.Println("=== PostgreSQL SQL Advisor 库使用示例 ===\n")

	// // 示例 1: 基础用法（不连接数据库）
	// fmt.Println("【示例 1】基础审核（无数据库连接）")
	// basicExample()

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// 示例 2: 高级用法（连接数据库获取元数据）
	fmt.Println("【示例 2】高级审核（带数据库元数据）")
	advancedExample()

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// // 示例 3: 自定义规则配置
	// fmt.Println("【示例 3】自定义规则配置")
	// customRulesExample()
}

// advancedExample 演示连接数据库进行高级审核
func advancedExample() {
	// 1. 数据库连接配置（使用环境变量或配置文件）
	dbConfig := &db.ConnectionConfig{
		DbType:   "postgres",
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DbName:   "mydb",
		SSLMode:  "disable",
		Timeout:  10,
	}

	// 2. 连接数据库
	ctx := context.Background()
	conn, err := db.OpenConnection(ctx, dbConfig)
	if err != nil {
		log.Printf("连接数据库失败: %v\n", err)
		log.Println("提示: 请确保 PostgreSQL 正在运行，且连接参数正确")
		log.Println("     docker run --name postgres-test -e POSTGRES_PASSWORD=secret -p 5432:5432 -d postgres")
		return
	}
	defer conn.Close()
	// 4. 要审核的 SQL（修改现有表结构）
	sql := `
-- 在 mydata schema 中创建新表
CREATE TABLE mydata.products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引（PostgreSQL 推荐使用 CONCURRENTLY）
CREATE INDEX idx_products_name ON mydata.products(name);
`

	fmt.Println("✅ 数据库连接成功")
	// Prepare review request
	req := &advisor.ReviewRequest{
		Engine:          advisor.EnginePostgres,
		Statement:       sql,
		CurrentDatabase: "mydb",
	}

	hasMetadata := false
	// 3. 获取数据库 schema 元数据
	metadata, err := db.GetDatabaseMetadata(ctx, conn, dbConfig)
	if err != nil {
		log.Printf("获取数据库元数据失败: %v\n", err)
		return
	} else {
		req.DBSchema = metadata
		hasMetadata = true
	}

	fmt.Printf("✅ 获取元数据成功，Schema 数量: %d\n\n", len(metadata.Schemas))

	// Load review rules
	rules, err := services.LoadRules("", advisor.EnginePostgres, hasMetadata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading rules: %v\n", err)
		os.Exit(1)
	}
	req.Rules = rules

	// 6. 构建带元数据的审核请求

	// 7. 执行审核
	resp, err := advisor.SQLReviewCheck(ctx, req)
	if err != nil {
		log.Printf("SQL 审核失败: %v\n", err)
		return
	}

	// 8. 输出详细的审核结果
	fmt.Printf("审核完成，发现 %d 个问题\n\n", len(resp.Advices))

	if len(resp.Advices) == 0 {
		fmt.Println("✅ 未发现任何问题，SQL 完全符合规范！")
		return
	}

	// 按严重程度分组显示
	errors := []*advisor.Advice{}
	warnings := []*advisor.Advice{}

	for _, advice := range resp.Advices {
		switch advice.Status {
		case advisor.AdviceStatusError:
			errors = append(errors, advice)
		case advisor.AdviceStatusWarning:
			warnings = append(warnings, advice)
		}

	}

	if len(errors) > 0 {
		fmt.Printf("❌ 错误 (%d):\n", len(errors))
		for i, advice := range errors {
			fmt.Printf("  %d. %s\n", i+1, advice.Title)
			fmt.Printf("     %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("     位置: Line %d, Column %d\n",
					advice.StartPosition.Line, advice.StartPosition.Column)
			}
			fmt.Println()
		}
	}

	if len(warnings) > 0 {
		fmt.Printf("⚠️  警告 (%d):\n", len(warnings))
		for i, advice := range warnings {
			fmt.Printf("  %d. %s\n", i+1, advice.Title)
			fmt.Printf("     %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("     位置: Line %d, Column %d\n",
					advice.StartPosition.Line, advice.StartPosition.Column)
			}
			fmt.Println()
		}
	}

	// 9. 决策建议
	fmt.Println("=== 决策建议 ===")
	if resp.HasError {
		fmt.Println("❌ 存在错误级别问题，强烈建议修复后再执行")
	} else if resp.HasWarning {
		fmt.Println("⚠️  存在警告级别问题，建议评估风险")
	}
}

// customRulesExample 演示如何使用自定义规则配置
func customRulesExample() {
	// 1. 使用 Payload 配置规则参数

	// 表命名规范：必须是小写字母和下划线，最大长度 63
	tableNamingRule, err := advisor.NewRuleWithPayload(
		advisor.RuleTableNaming,
		advisor.RuleLevelWarning,
		advisor.NamingRulePayload{
			Format:    "^[a-z][a-z0-9_]*$",
			MaxLength: 63,
		},
	)
	if err != nil {
		log.Printf("创建规则失败: %v\n", err)
		return
	}

	// 列命名规范
	columnNamingRule, err := advisor.NewRuleWithPayload(
		advisor.RuleColumnNaming,
		advisor.RuleLevelWarning,
		advisor.NamingRulePayload{
			Format:    "^[a-z][a-z0-9_]*$",
			MaxLength: 63,
		},
	)
	if err != nil {
		log.Printf("创建规则失败: %v\n", err)
		return
	}

	// 必需列：每个表必须包含这些列
	requiredColumnsRule, err := advisor.NewRuleWithPayload(
		advisor.RuleRequiredColumn,
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"id", "created_at", "updated_at"},
		},
	)
	if err != nil {
		log.Printf("创建规则失败: %v\n", err)
		return
	}

	// 2. 组合所有规则
	rules := []*advisor.SQLReviewRule{
		tableNamingRule,
		columnNamingRule,
		requiredColumnsRule,

		// 其他基础规则（不需要元数据）
		advisor.NewRule(
			advisor.RuleStatementNoSelectAll,
			advisor.RuleLevelWarning,
		),
		advisor.NewRule(
			advisor.RuleTableNoFK,
			advisor.RuleLevelWarning,
		),
	}

	// 3. 测试 SQL（故意包含一些不符合规范的语句）
	sql := `
-- 表名不符合命名规范（应该是 user_profile）
CREATE TABLE UserProfile (
    user_id SERIAL PRIMARY KEY,
    UserName VARCHAR(3000),  -- 列名不符合规范，VARCHAR 长度超限
    email VARCHAR(100)
    -- 缺少 created_at 和 updated_at 列
);

-- 禁止 SELECT *
SELECT * FROM UserProfile;
`

	// 4. 执行审核
	req := &advisor.ReviewRequest{
		Engine:          advisor.EnginePostgres,
		Statement:       sql,
		CurrentDatabase: "mydb",
		Rules:           rules,
	}

	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		log.Printf("SQL 审核失败: %v\n", err)
		return
	}

	// 5. 输出结果
	fmt.Printf("审核完成，发现 %d 个问题\n\n", len(resp.Advices))

	for i, advice := range resp.Advices {
		severity := "INFO"
		icon := "ℹ️ "
		if advice.Status == advisor.AdviceStatusError {
			severity = "ERROR"
			icon = "❌"
		} else if advice.Status == advisor.AdviceStatusWarning {
			severity = "WARNING"
			icon = "⚠️ "
		}

		fmt.Printf("%d. %s [%s] %s\n", i+1, icon, severity, advice.Title)
		fmt.Printf("   %s\n", advice.Content)
		if advice.StartPosition != nil {
			fmt.Printf("   位置: Line %d, Column %d\n",
				advice.StartPosition.Line, advice.StartPosition.Column)
		}
		fmt.Println()
	}

	if resp.HasError {
		fmt.Println("❌ 存在错误级别的规范问题")
	} else if resp.HasWarning {
		fmt.Println("⚠️  存在警告级别的规范问题")
	} else {
		fmt.Println("✅ 所有检查通过")
	}
}
