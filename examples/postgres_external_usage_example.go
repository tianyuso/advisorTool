// Package main 演示如何在外部程序中使用 advisorTool/services 包进行 PostgreSQL SQL 审核
// 本示例连接到真实的 PostgreSQL 数据库，获取元数据进行全面审核
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tianyuso/advisorTool/pkg/advisor"
	"github.com/tianyuso/advisorTool/services"
)

func main() {
	fmt.Println("=== PostgreSQL SQL 审核示例（带数据库连接） ===\n")

	//  1. 配置数据库连接参数
	dbParams := &services.DBConnectionParams{
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DbName:   "mydb",
		SSLMode:  "disable",
		Timeout:  10,
		// Schema:   "mydata", // 注释掉 Schema 参数，避免审核时的表名解析问题
	}

	engineType := advisor.EnginePostgres

	// 2. 获取数据库元数据（用于向后兼容性检查等高级规则）
	fmt.Println("📊 正在连接数据库并获取元数据...")
	metadata, err := services.FetchDatabaseMetadata(engineType, dbParams)
	if err != nil {
		log.Printf("⚠️  警告: 获取数据库元数据失败: %v", err)
		log.Println("将使用基础规则进行审核（跳过需要元数据的规则）\n")
		metadata = nil
	} else {
		fmt.Printf("✅ 成功连接到数据库 %s@%s:%d/%s\n",
			dbParams.User, dbParams.Host, dbParams.Port, dbParams.DbName)
		fmt.Printf("✅ 获取元数据成功，Schema 数量: %d\n\n", len(metadata.Schemas))
	}

	// 3. 使用 services 包加载规则（包括需要元数据的规则）
	hasMetadata := (metadata != nil)
	rules, err := services.LoadRules("", engineType, hasMetadata)
	if err != nil {
		log.Fatalf("❌ 加载规则失败: %v", err)
	}

	fmt.Printf("✅ 成功加载 %d 条 PostgreSQL 审核规则\n\n", len(rules))

	// 4. 准备要审核的 SQL（包含建表、索引、UPDATE、DELETE 等）
	sql := `
-- ===== 创建表 =====
-- 在 mydata schema 中创建测试表
CREATE TABLE mydata.test_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    age INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建订单表
CREATE TABLE mydata.test_orders (
    order_id SERIAL PRIMARY KEY,
    user_id INT,
    order_no VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2),
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- ===== UPDATE 语句 =====
-- 正常的 UPDATE（有 WHERE 条件）
UPDATE mydata.test_users 
SET status = 'inactive', updated_at = CURRENT_TIMESTAMP 
WHERE id <= 3;

`

	fmt.Println("📝 准备审核以下 SQL 语句:")
	fmt.Println("   - CREATE TABLE (2 个表)")
	fmt.Println("   - CREATE INDEX (3 个索引)")
	fmt.Println("   - SELECT (2 个查询)")
	fmt.Println("   - UPDATE (2 个更新)")
	fmt.Println("   - DELETE (2 个删除)")
	fmt.Println("   - ALTER TABLE (3 个变更)")
	fmt.Println()

	// 5. 创建审核请求
	req := &advisor.ReviewRequest{
		Engine:          engineType,
		Statement:       sql,
		CurrentDatabase: dbParams.DbName,
		Rules:           rules,
		DBSchema:        metadata, // 提供元数据以支持高级规则
	}

	// 6. 执行 SQL 审核
	fmt.Println("🔍 开始执行 SQL 审核...")
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		log.Fatalf("❌ SQL 审核失败: %v", err)
	}

	fmt.Printf("✅ 审核完成，发现 %d 个问题\n\n", len(resp.Advices))

	// 7. 使用 services 包转换结果为结构化格式
	affectedRowsMap := services.CalculateAffectedRowsForStatements(sql, engineType, dbParams)
	results := services.ConvertToReviewResults(resp, sql, engineType, affectedRowsMap)

	// 8. 输出详细结果
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("📋 审核结果详情")
	fmt.Println(strings.Repeat("=", 70))

	errorCount := 0
	warningCount := 0

	for _, result := range results {
		var level string
		var icon string

		switch result.ErrorLevel {
		case "0":
			level = "✓ OK"
			icon = "✅"
		case "1":
			level = "⚠ WARNING"
			icon = "⚠️ "
			warningCount++
		case "2":
			level = "✗ ERROR"
			icon = "❌"
			errorCount++
		}

		fmt.Printf("\n%s SQL #%d [%s]\n", icon, result.OrderID, level)

		// 格式化显示 SQL（限制长度）
		sqlPreview := result.SQL
		if len(sqlPreview) > 80 {
			sqlPreview = sqlPreview[:77] + "..."
		}
		fmt.Printf("   SQL: %s\n", sqlPreview)

		if result.AffectedRows > 0 {
			fmt.Printf("   影响行数: %d\n", result.AffectedRows)
		}

		if result.ErrorMessage != "" {
			fmt.Printf("   问题: %s\n", result.ErrorMessage)
		}
	}

	// 9. 输出统计信息
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📊 统计信息")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("总 SQL 语句数: %d\n", len(results))
	fmt.Printf("✅ 通过: %d\n", len(results)-errorCount-warningCount)
	if warningCount > 0 {
		fmt.Printf("⚠️  警告: %d\n", warningCount)
	}
	if errorCount > 0 {
		fmt.Printf("❌ 错误: %d\n", errorCount)
	}

	// 10. 也可以使用 services.OutputResults 输出 JSON 格式
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📄 JSON 格式输出（兼容 Inception 格式）")
	fmt.Println(strings.Repeat("=", 70))
	if err := services.OutputResults(resp, sql, engineType, "json", dbParams); err != nil {
		log.Printf("输出结果失败: %v", err)
	}

	// 11. 也可以输出表格格式
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📊 表格格式输出")
	fmt.Println(strings.Repeat("=", 70))
	if err := services.OutputResults(resp, sql, engineType, "table", dbParams); err != nil {
		log.Printf("输出结果失败: %v", err)
	}

	// 12. 决策建议
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("💡 决策建议")
	fmt.Println(strings.Repeat("=", 70))

	if resp.HasError {
		fmt.Println("❌ 存在错误级别问题，强烈建议修复后再执行")
		fmt.Println("   这些问题可能导致：数据丢失、服务中断、向后不兼容等严重后果")
		fmt.Println("\n   需要修复的错误：")
		for _, advice := range resp.Advices {
			if advice.Status == advisor.AdviceStatusError {
				fmt.Printf("   - %s\n", advice.Title)
			}
		}
	} else if resp.HasWarning {
		fmt.Println("⚠️  存在警告级别问题，建议评估风险")
		fmt.Println("   这些问题可能影响：性能、可维护性、最佳实践等")
		fmt.Println("\n   建议优化的警告：")
		for _, advice := range resp.Advices {
			if advice.Status == advisor.AdviceStatusWarning {
				fmt.Printf("   - %s\n", advice.Title)
			}
		}
	} else {
		fmt.Println("✅ 审核通过，可以安全执行")
	}

	// 13. PostgreSQL 特定建议
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🐘 PostgreSQL 特定建议")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("1. 创建索引时使用 CONCURRENTLY 关键字，避免锁表")
	fmt.Println("   示例: CREATE INDEX CONCURRENTLY idx_name ON table(column);")
	fmt.Println()
	fmt.Println("2. 添加带默认值的列可能会锁表，建议分两步：")
	fmt.Println("   a) ALTER TABLE ADD COLUMN without DEFAULT;")
	fmt.Println("   b) UPDATE TABLE SET column = value;")
	fmt.Println()
	fmt.Println("3. 添加约束时使用 NOT VALID，然后再 VALIDATE")
	fmt.Println("   示例: ALTER TABLE ADD CONSTRAINT ... CHECK (...) NOT VALID;")
	fmt.Println()
	fmt.Println("4. 始终在 schema 名称前加上完全限定名")
	fmt.Println("   示例: SELECT * FROM mydata.test_users;")

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎉 审核完成！")
	fmt.Println(strings.Repeat("=", 70))
}
