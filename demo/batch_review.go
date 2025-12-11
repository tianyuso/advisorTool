// Package main demonstrates batch SQL review from files with full rule sets and database metadata
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"advisorTool/pkg/advisor"
	"demo/common"
)

func main() {
	fmt.Println("=== SQL Advisor Tool - 批量审核示例 ===")
	fmt.Println("使用完整规则集批量审核多个 SQL 文件\n")

	// 示例 1: 从文件读取并审核（完整规则集）
	example1()

	fmt.Println("\n" + "="*60 + "\n")

	// 示例 2: 批量审核多个文件（支持元数据）
	example2()

	fmt.Println("\n" + "="*60 + "\n")

	// 示例 3: 生成详细审核报告
	example3()
}

// example1 演示从文件读取 SQL 进行审核
func example1() {
	fmt.Println("示例 1: 从文件读取 SQL 进行审核")
	fmt.Println("使用完整规则集进行审核\n")

	// 创建临时测试文件
	testSQL := `-- 用户订单表
CREATE TABLE user_orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    order_amount DECIMAL(10,2),
    created_at TIMESTAMP
);

-- 查询订单
SELECT * FROM user_orders WHERE user_id = 1;

-- 更新订单金额（缺少 WHERE）
UPDATE user_orders SET order_amount = 100;

-- 删除订单
DELETE FROM user_orders WHERE id = 1;

-- 插入订单
INSERT INTO user_orders VALUES (1, 100, 99.99, NOW());
`

	// 写入临时文件
	tmpFile := "/tmp/test_review.sql"
	if err := ioutil.WriteFile(tmpFile, []byte(testSQL), 0644); err != nil {
		fmt.Printf("❌ 创建测试文件失败: %v\n", err)
		return
	}
	defer os.Remove(tmpFile)

	fmt.Printf("测试文件: %s\n", tmpFile)
	fmt.Println("文件内容:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(testSQL)
	fmt.Println(strings.Repeat("-", 60))

	// 读取文件内容
	content, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		fmt.Printf("❌ 读取文件失败: %v\n", err)
		return
	}

	// 获取完整规则集
	rules := common.GetDefaultRules(advisor.EngineMySQL, false)
	fmt.Printf("\n已加载 %d 条审核规则\n\n", len(rules))

	// 执行审核
	req := &advisor.ReviewRequest{
		Engine:    advisor.EngineMySQL,
		Statement: string(content),
		Rules:     rules,
	}

	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ 审核失败: %v\n", err)
		return
	}

	// 输出结果
	fmt.Println("审核结果:")
	fmt.Println("=" * 60)
	printDetailedResult(resp)
}

// example2 演示批量审核多个文件（支持元数据）
func example2() {
	fmt.Println("示例 2: 批量审核多个 SQL 文件")
	fmt.Println("支持数据库元数据的完整规则集\n")

	// 数据库配置（可选）
	var dbConfig *common.DBConfig = nil

	/*
		// 如需测试元数据功能，取消注释并填写真实配置
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

	// 获取元数据
	metadata, _ := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
	hasMetadata := (metadata != nil)

	if !hasMetadata {
		fmt.Println("⚠️ 未配置数据库连接，使用静态分析模式\n")
	}

	// 创建临时目录和多个测试文件
	tmpDir := "/tmp/sql_reviews"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// 文件 1: 建表语句
	file1 := filepath.Join(tmpDir, "01_create_tables.sql")
	sql1 := `-- 用户表
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 产品表（缺少主键）
CREATE TABLE products (
    product_id INT,
    name VARCHAR(100),
    price DECIMAL(10,2)
) ENGINE=InnoDB;
`
	ioutil.WriteFile(file1, []byte(sql1), 0644)

	// 文件 2: DML 语句
	file2 := filepath.Join(tmpDir, "02_dml_operations.sql")
	sql2 := `-- 插入用户
INSERT INTO users (id, username, email) VALUES (1, 'alice', 'alice@example.com');

-- 更新用户（有 WHERE）
UPDATE users SET email = 'newemail@example.com' WHERE id = 1;

-- 删除用户（有 WHERE）
DELETE FROM users WHERE id = 1;

-- 危险操作：无 WHERE 的 UPDATE
UPDATE users SET status = 'inactive';
`
	ioutil.WriteFile(file2, []byte(sql2), 0644)

	// 文件 3: 查询语句
	file3 := filepath.Join(tmpDir, "03_queries.sql")
	sql3 := `-- 查询所有用户（SELECT *）
SELECT * FROM users;

-- 正确的查询
SELECT id, username, email FROM users WHERE status = 'active';

-- 前导通配符 LIKE
SELECT * FROM products WHERE name LIKE '%phone%';
`
	ioutil.WriteFile(file3, []byte(sql3), 0644)

	// 获取规则
	rules := common.GetDefaultRules(advisor.EngineMySQL, hasMetadata)
	fmt.Printf("已加载 %d 条审核规则 (hasMetadata=%v)\n\n", len(rules), hasMetadata)

	// 遍历目录中的所有 .sql 文件
	files, err := filepath.Glob(filepath.Join(tmpDir, "*.sql"))
	if err != nil {
		fmt.Printf("❌ 读取目录失败: %v\n", err)
		return
	}

	fmt.Printf("找到 %d 个 SQL 文件\n", len(files))

	type fileResult struct {
		filename string
		passed   bool
		hasError bool
		hasWarn  bool
		advices  []*advisor.Advice
	}

	results := []fileResult{}

	for _, file := range files {
		fmt.Printf("\n" + strings.Repeat("=", 60))
		fmt.Printf("\n审核文件: %s\n", filepath.Base(file))
		fmt.Println(strings.Repeat("=", 60))

		// 读取文件内容
		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("❌ 读取文件失败: %v\n", err)
			continue
		}

		fmt.Printf("\n文件内容:\n%s\n", string(content))

		// 执行审核
		req := &advisor.ReviewRequest{
			Engine:    advisor.EngineMySQL,
			Statement: string(content),
			Rules:     rules,
			DBSchema:  metadata,
		}

		resp, err := advisor.SQLReviewCheck(context.Background(), req)
		if err != nil {
			fmt.Printf("❌ 审核失败: %v\n", err)
			continue
		}

		// 记录结果
		result := fileResult{
			filename: filepath.Base(file),
			passed:   len(resp.Advices) == 0,
			hasError: resp.HasError,
			hasWarn:  resp.HasWarning,
			advices:  resp.Advices,
		}
		results = append(results, result)

		// 输出结果
		fmt.Println("\n审核结果:")
		if len(resp.Advices) == 0 {
			fmt.Println("✅ 通过审核")
		} else {
			for i, advice := range resp.Advices {
				icon := "⚠️"
				statusText := "WARNING"
				if advice.Status == advisor.AdviceStatusError {
					icon = "❌"
					statusText = "ERROR"
				}
				fmt.Printf("%d. %s [%s] %s\n", i+1, icon, statusText, advice.Title)
				fmt.Printf("   %s\n", advice.Content)
				if advice.StartPosition != nil {
					fmt.Printf("   位置: 行 %d\n", advice.StartPosition.Line)
				}
			}
		}
	}

	// 汇总报告
	printBatchSummary(results)
}

// example3 演示生成详细审核报告
func example3() {
	fmt.Println("示例 3: 生成详细审核报告")
	fmt.Println("包含问题分类、严重程度统计和修复建议\n")

	// 准备测试数据
	tmpDir := "/tmp/sql_report_demo"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// 创建包含各种问题的 SQL 文件
	testSQL := `-- 测试文件：包含多种问题

-- 问题1: 缺少主键
CREATE TABLE test1 (
    name VARCHAR(50)
);

-- 问题2: SELECT *
SELECT * FROM users;

-- 问题3: UPDATE 缺少 WHERE
UPDATE products SET price = 100;

-- 问题4: 使用禁止的类型
CREATE TABLE test2 (
    id INT PRIMARY KEY,
    content TEXT
);

-- 问题5: 插入不指定列名
INSERT INTO test1 VALUES ('test');

-- 正确的语句
CREATE TABLE correct_table (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL
) ENGINE=InnoDB;
`

	testFile := filepath.Join(tmpDir, "test_with_issues.sql")
	ioutil.WriteFile(testFile, []byte(testSQL), 0644)

	// 获取完整规则集
	rules := common.GetDefaultRules(advisor.EngineMySQL, false)

	// 添加类型黑名单规则
	typeRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleColumnTypeDisallowList,
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"TEXT", "BLOB"},
		},
	)
	rules = append(rules, typeRule)

	fmt.Printf("使用 %d 条审核规则\n", len(rules))
	fmt.Printf("测试文件: %s\n\n", testFile)

	// 读取并审核
	content, _ := ioutil.ReadFile(testFile)

	req := &advisor.ReviewRequest{
		Engine:    advisor.EngineMySQL,
		Statement: string(content),
		Rules:     rules,
	}

	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ 审核失败: %v\n", err)
		return
	}

	// 生成详细报告
	generateDetailedReport(resp, string(content))
}

// printDetailedResult 打印详细的审核结果
func printDetailedResult(resp *advisor.ReviewResponse) {
	if len(resp.Advices) == 0 {
		fmt.Println("✅ 通过所有审核规则！")
		return
	}

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

	fmt.Printf("发现 %d 个问题 (错误: %d, 警告: %d)\n\n",
		len(resp.Advices), len(errorAdvices), len(warningAdvices))

	// 先输出错误
	if len(errorAdvices) > 0 {
		fmt.Printf("❌ 错误 (%d 个) - 必须修复:\n", len(errorAdvices))
		for i, advice := range errorAdvices {
			fmt.Printf("%d. [%s]\n", i+1, advice.Title)
			fmt.Printf("   内容: %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("   位置: 行 %d, 列 %d\n",
					advice.StartPosition.Line,
					advice.StartPosition.Column)
			}
			fmt.Printf("   修复建议: %s\n", getSuggestion(advice.Title))
			fmt.Println()
		}
	}

	// 再输出警告
	if len(warningAdvices) > 0 {
		fmt.Printf("⚠️ 警告 (%d 个) - 建议修复:\n", len(warningAdvices))
		for i, advice := range warningAdvices {
			fmt.Printf("%d. [%s]\n", i+1, advice.Title)
			fmt.Printf("   内容: %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("   位置: 行 %d, 列 %d\n",
					advice.StartPosition.Line,
					advice.StartPosition.Column)
			}
			fmt.Printf("   修复建议: %s\n", getSuggestion(advice.Title))
			fmt.Println()
		}
	}
}

// printBatchSummary 打印批量审核汇总报告
func printBatchSummary(results []fileResult) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("批量审核汇总报告")
	fmt.Println(strings.Repeat("=", 60))

	passedCount := 0
	errorCount := 0
	warningCount := 0
	totalIssues := 0

	for _, r := range results {
		if r.passed {
			passedCount++
		}
		if r.hasError {
			errorCount++
		}
		if r.hasWarn {
			warningCount++
		}
		totalIssues += len(r.advices)
	}

	fmt.Printf("\n总文件数: %d\n", len(results))
	fmt.Printf("✅ 通过: %d\n", passedCount)
	fmt.Printf("⚠️ 有警告: %d\n", warningCount)
	fmt.Printf("❌ 有错误: %d\n", errorCount)
	fmt.Printf("总问题数: %d\n\n", totalIssues)

	// 详细列表
	fmt.Println("文件详情:")
	for i, r := range results {
		icon := "✅"
		status := "通过"
		if r.hasError {
			icon = "❌"
			status = fmt.Sprintf("错误 (%d 个问题)", len(r.advices))
		} else if r.hasWarn {
			icon = "⚠️"
			status = fmt.Sprintf("警告 (%d 个问题)", len(r.advices))
		}
		fmt.Printf("%d. %s %s - %s\n", i+1, icon, r.filename, status)
	}

	// 最终判定
	fmt.Println("\n" + strings.Repeat("-", 60))
	if errorCount > 0 {
		fmt.Println("❌ 状态: 存在错误，必须修复后才能部署")
	} else if warningCount > 0 {
		fmt.Println("⚠️ 状态: 存在警告，建议修复")
	} else {
		fmt.Println("✅ 状态: 所有文件通过审核")
	}
}

// generateDetailedReport 生成详细的审核报告
func generateDetailedReport(resp *advisor.ReviewResponse, sqlContent string) {
	fmt.Println("=" * 60)
	fmt.Println("详细审核报告")
	fmt.Println("=" * 60)

	if len(resp.Advices) == 0 {
		fmt.Println("\n✅ 通过所有审核规则！")
		return
	}

	// 统计信息
	errorCount := 0
	warningCount := 0
	ruleTypes := make(map[string]int)

	for _, advice := range resp.Advices {
		if advice.Status == advisor.AdviceStatusError {
			errorCount++
		} else {
			warningCount++
		}
		ruleTypes[advice.Title]++
	}

	// 1. 总体概况
	fmt.Println("\n【总体概况】")
	fmt.Printf("  • 总问题数: %d\n", len(resp.Advices))
	fmt.Printf("  • 错误级别: %d (必须修复)\n", errorCount)
	fmt.Printf("  • 警告级别: %d (建议修复)\n", warningCount)

	// 2. 问题分类统计
	fmt.Println("\n【问题分类统计】")
	for ruleType, count := range ruleTypes {
		category := getRuleCategory(ruleType)
		fmt.Printf("  • %s: %d 个\n", category, count)
	}

	// 3. 详细问题列表
	fmt.Println("\n【详细问题列表】")
	for i, advice := range resp.Advices {
		icon := "⚠️"
		level := "WARNING"
		if advice.Status == advisor.AdviceStatusError {
			icon = "❌"
			level = "ERROR"
		}

		fmt.Printf("\n问题 %d: %s [%s]\n", i+1, icon, level)
		fmt.Printf("  规则: %s\n", advice.Title)
		fmt.Printf("  描述: %s\n", advice.Content)
		if advice.StartPosition != nil {
			fmt.Printf("  位置: 行 %d\n", advice.StartPosition.Line)
		}
		fmt.Printf("  建议: %s\n", getSuggestion(advice.Title))
	}

	// 4. 修复优先级
	fmt.Println("\n【修复优先级】")
	if errorCount > 0 {
		fmt.Println("🔴 高优先级 (ERROR) - 必须立即修复:")
		priority := 1
		for _, advice := range resp.Advices {
			if advice.Status == advisor.AdviceStatusError {
				fmt.Printf("  %d. %s\n", priority, advice.Content)
				priority++
			}
		}
	}
	if warningCount > 0 {
		fmt.Println("\n🟡 中优先级 (WARNING) - 建议尽快修复:")
		priority := 1
		for _, advice := range resp.Advices {
			if advice.Status == advisor.AdviceStatusWarning {
				fmt.Printf("  %d. %s\n", priority, advice.Content)
				priority++
			}
		}
	}

	// 5. 最终评估
	fmt.Println("\n【最终评估】")
	if errorCount > 0 {
		fmt.Println("❌ 不通过 - 存在必须修复的错误")
		fmt.Println("建议: 修复所有 ERROR 级别问题后重新审核")
	} else if warningCount > 0 {
		fmt.Println("⚠️ 有风险 - 存在需要关注的警告")
		fmt.Println("建议: 评估警告影响，建议修复后部署")
	} else {
		fmt.Println("✅ 通过 - 符合规范要求")
	}
}

// getSuggestion 根据规则类型返回修复建议
func getSuggestion(ruleTitle string) string {
	suggestions := map[string]string{
		"statement.select.no-select-all":        "使用明确的列名代替 SELECT *",
		"statement.where.require.update-delete": "为 UPDATE/DELETE 语句添加 WHERE 条件",
		"statement.where.require.select":        "为 SELECT 语句添加 WHERE 条件以提高性能",
		"table.require-pk":                      "为表添加主键约束",
		"table.no-foreign-key":                  "考虑在应用层实现外键逻辑",
		"naming.table":                          "调整表名符合命名规范（小写+下划线）",
		"column.type-disallow-list":             "使用 VARCHAR 代替 TEXT，VARBINARY 代替 BLOB",
		"statement.insert.must-specify-column":  "INSERT 语句明确指定列名",
		"column.auto-increment-must-integer":    "自增列使用 INT 或 BIGINT 类型",
		"index.no-duplicate-column":             "移除索引中的重复列",
		"statement.no-leading-wildcard-like":    "避免 LIKE 前导 %，考虑使用全文索引",
		"column.auto-increment-must-unsigned":   "自增列使用 UNSIGNED 类型",
	}

	for key, suggestion := range suggestions {
		if strings.Contains(ruleTitle, key) {
			return suggestion
		}
	}

	return "请根据规则描述进行修复"
}

// getRuleCategory 获取规则分类
func getRuleCategory(ruleTitle string) string {
	if strings.Contains(ruleTitle, "statement") {
		return "语句规范"
	} else if strings.Contains(ruleTitle, "table") {
		return "表结构规范"
	} else if strings.Contains(ruleTitle, "column") {
		return "列规范"
	} else if strings.Contains(ruleTitle, "index") {
		return "索引规范"
	} else if strings.Contains(ruleTitle, "naming") {
		return "命名规范"
	}
	return "其他规范"
}

type fileResult struct {
	filename string
	passed   bool
	hasError bool
	hasWarn  bool
	advices  []*advisor.Advice
}
