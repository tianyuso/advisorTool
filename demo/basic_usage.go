// Package main demonstrates basic usage of SQL Advisor Tool as a Go library
package main

import (
	"context"
	"fmt"
	"os"

	"advisorTool/pkg/advisor"
	"demo/common"
)

func main() {
	fmt.Println("=== SQL Advisor Tool - 基础用法示例 ===")
	fmt.Println("本示例使用完整的默认规则集进行审核\n")

	// 示例 1: 静态分析（无需数据库连接）
	example1()

	fmt.Println("\n" + "="*60 + "\n")

	// 示例 2: 动态分析（连接数据库获取元数据）
	example2()

	fmt.Println("\n" + "="*60 + "\n")

	// 示例 3: 批量 SQL 语句审核
	example3()

	fmt.Println("\n" + "="*60 + "\n")

	// 示例 4: 不同数据库引擎
	example4()
}

// example1 演示静态分析模式（不需要数据库连接）
func example1() {
	fmt.Println("示例 1: 静态分析模式")
	fmt.Println("不连接数据库，使用完整的静态规则集\n")

	// 待审核的 SQL 语句
	sql := `
SELECT * FROM users WHERE id = 1;
DELETE FROM orders;
UPDATE products SET price = 100;
CREATE TABLE test (name VARCHAR(50));
`

	fmt.Println("待审核 SQL:")
	fmt.Println(sql)

	// 获取默认规则（hasMetadata = false）
	rules := common.GetDefaultRules(advisor.EngineMySQL, false)
	fmt.Printf("已加载 %d 条审核规则（静态分析）\n\n", len(rules))

	// 创建审核请求
	req := &advisor.ReviewRequest{
		Engine:    advisor.EngineMySQL,
		Statement: sql,
		Rules:     rules,
	}

	// 执行审核
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ 审核失败: %v\n", err)
		return
	}

	// 输出结果
	printReviewResult(resp, "静态分析")
}

// example2 演示动态分析模式（需要数据库连接）
func example2() {
	fmt.Println("示例 2: 动态分析模式（连接数据库）")
	fmt.Println("提示: 如需测试，请修改数据库连接参数\n")

	// 数据库连接配置（默认为空，演示静态模式）
	// 如需测试动态分析，请取消注释并填写实际的数据库连接信息
	var dbConfig *common.DBConfig = nil

	/*
		// MySQL 示例配置
		dbConfig = &common.DBConfig{
			Host:     "127.0.0.1",
			Port:     3306,
			User:     "root",
			Password: "your_password",
			DBName:   "test_db",
			Charset:  "utf8mb4",
			Timeout:  5,
		}
	*/

	// 尝试获取数据库元数据
	metadata, err := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
	hasMetadata := (metadata != nil && err == nil)

	if !hasMetadata {
		fmt.Println("⚠️ 未配置数据库连接，将使用静态分析模式")
		fmt.Println("部分需要元数据的规则将被跳过\n")
	}

	// 获取规则（根据是否有元数据决定规则集）
	rules := common.GetDefaultRules(advisor.EngineMySQL, hasMetadata)
	fmt.Printf("已加载 %d 条审核规则 (hasMetadata=%v)\n\n", len(rules), hasMetadata)

	// 测试 SQL
	sql := `
CREATE TABLE user_orders (
    id BIGINT AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    total_amount DECIMAL(10,2),
    status VARCHAR(20)
);
`

	fmt.Println("待审核 SQL:")
	fmt.Println(sql)

	// 创建审核请求
	req := &advisor.ReviewRequest{
		Engine:    advisor.EngineMySQL,
		Statement: sql,
		Rules:     rules,
		DBSchema:  metadata,
	}

	// 执行审核
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ 审核失败: %v\n", err)
		return
	}

	// 输出结果
	mode := "静态分析"
	if hasMetadata {
		mode = "动态分析（含元数据）"
	}
	printReviewResult(resp, mode)
}

// example3 演示批量 SQL 语句审核
func example3() {
	fmt.Println("示例 3: 批量 SQL 语句审核")
	fmt.Println("使用完整规则集检查多个 SQL 语句\n")

	sqlStatements := []struct {
		desc string
		sql  string
	}{
		{
			desc: "符合规范的查询",
			sql:  "SELECT id, name, email FROM users WHERE id = 1;",
		},
		{
			desc: "使用 SELECT * (警告)",
			sql:  "SELECT * FROM users WHERE status = 'active';",
		},
		{
			desc: "DELETE 缺少 WHERE (错误)",
			sql:  "DELETE FROM orders;",
		},
		{
			desc: "UPDATE 有 WHERE (通过)",
			sql:  "UPDATE products SET price = 100 WHERE id = 1;",
		},
		{
			desc: "建表缺少主键 (错误)",
			sql:  "CREATE TABLE test (name VARCHAR(50));",
		},
		{
			desc: "正确的建表语句",
			sql:  "CREATE TABLE products (id BIGINT PRIMARY KEY, name VARCHAR(100)) ENGINE=InnoDB;",
		},
	}

	// 获取规则
	rules := common.GetDefaultRules(advisor.EngineMySQL, false)

	passCount := 0
	warnCount := 0
	errorCount := 0

	for i, tc := range sqlStatements {
		fmt.Printf("[%d] %s\n", i+1, tc.desc)
		fmt.Printf("SQL: %s\n", tc.sql)

		req := &advisor.ReviewRequest{
			Engine:    advisor.EngineMySQL,
			Statement: tc.sql,
			Rules:     rules,
		}

		resp, err := advisor.SQLReviewCheck(context.Background(), req)
		if err != nil {
			fmt.Printf("  ❌ 审核失败: %v\n\n", err)
			continue
		}

		if len(resp.Advices) == 0 {
			fmt.Println("  ✅ 通过审核")
			passCount++
		} else {
			if resp.HasError {
				errorCount++
			} else if resp.HasWarning {
				warnCount++
			}
			for _, advice := range resp.Advices {
				icon := "⚠️"
				statusText := "WARNING"
				if advice.Status == advisor.AdviceStatusError {
					icon = "❌"
					statusText = "ERROR"
				}
				fmt.Printf("  %s [%s] %s\n", icon, statusText, advice.Content)
			}
		}
		fmt.Println()
	}

	// 统计汇总
	fmt.Println("--- 批量审核汇总 ---")
	fmt.Printf("总语句数: %d\n", len(sqlStatements))
	fmt.Printf("✅ 通过: %d\n", passCount)
	fmt.Printf("⚠️ 警告: %d\n", warnCount)
	fmt.Printf("❌ 错误: %d\n", errorCount)
}

// example4 演示不同数据库引擎的完整规则集
func example4() {
	fmt.Println("示例 4: 不同数据库引擎的规则集")

	engines := []struct {
		name   string
		engine advisor.Engine
		sql    string
	}{
		{
			name:   "MySQL",
			engine: advisor.EngineMySQL,
			sql:    "SELECT * FROM users; DELETE FROM orders;",
		},
		{
			name:   "PostgreSQL",
			engine: advisor.EnginePostgres,
			sql:    "SELECT * FROM users; DELETE FROM orders;",
		},
		{
			name:   "Oracle",
			engine: advisor.EngineOracle,
			sql:    "SELECT * FROM users; DELETE FROM orders;",
		},
	}

	for _, e := range engines {
		fmt.Printf("\n--- %s 数据库 ---\n", e.name)

		// 获取该引擎的默认规则
		rules := common.GetDefaultRules(e.engine, false)
		fmt.Printf("规则数量: %d 条\n", len(rules))

		req := &advisor.ReviewRequest{
			Engine:    e.engine,
			Statement: e.sql,
			Rules:     rules,
		}

		resp, err := advisor.SQLReviewCheck(context.Background(), req)
		if err != nil {
			fmt.Printf("❌ 审核失败: %v\n", err)
			continue
		}

		if len(resp.Advices) > 0 {
			fmt.Printf("发现 %d 个问题:\n", len(resp.Advices))
			for _, advice := range resp.Advices {
				icon := "⚠️"
				if advice.Status == advisor.AdviceStatusError {
					icon = "❌"
				}
				fmt.Printf("  %s %s\n", icon, advice.Content)
			}
		} else {
			fmt.Println("✅ 通过审核")
		}
	}
}

// printReviewResult 打印审核结果
func printReviewResult(resp *advisor.ReviewResponse, mode string) {
	fmt.Printf("审核模式: %s\n", mode)
	fmt.Println("=" * 60)

	if len(resp.Advices) == 0 {
		fmt.Println("✅ 审核通过！没有发现问题。")
		return
	}

	fmt.Printf("发现 %d 个问题:\n\n", len(resp.Advices))

	// 分组统计
	errorAdvices := []*advisor.Advice{}
	warningAdvices := []*advisor.Advice{}

	for _, advice := range resp.Advices {
		if advice.Status == advisor.AdviceStatusError {
			errorAdvices = append(errorAdvices, advice)
		} else {
			warningAdvices = append(warningAdvices, advice)
		}
	}

	// 先输出错误
	if len(errorAdvices) > 0 {
		fmt.Printf("❌ 错误 (%d 个):\n", len(errorAdvices))
		for i, advice := range errorAdvices {
			fmt.Printf("%d. [%s]\n", i+1, advice.Title)
			fmt.Printf("   内容: %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("   位置: 行 %d, 列 %d\n",
					advice.StartPosition.Line,
					advice.StartPosition.Column)
			}
			fmt.Println()
		}
	}

	// 再输出警告
	if len(warningAdvices) > 0 {
		fmt.Printf("⚠️ 警告 (%d 个):\n", len(warningAdvices))
		for i, advice := range warningAdvices {
			fmt.Printf("%d. [%s]\n", i+1, advice.Title)
			fmt.Printf("   内容: %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("   位置: 行 %d, 列 %d\n",
					advice.StartPosition.Line,
					advice.StartPosition.Column)
			}
			fmt.Println()
		}
	}

	// 汇总统计
	fmt.Println("--- 审核汇总 ---")
	if resp.HasError {
		fmt.Println("❌ 状态: 存在错误级别问题，必须修复")
		os.Exit(2)
	} else if resp.HasWarning {
		fmt.Println("⚠️ 状态: 存在警告级别问题，建议修复")
		os.Exit(1)
	} else {
		fmt.Println("✅ 状态: 通过审核")
	}
}
