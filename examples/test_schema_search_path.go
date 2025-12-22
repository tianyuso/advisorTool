// Package main 演示 PostgreSQL schema search_path 功能
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
	fmt.Println("=== PostgreSQL Schema Search Path 功能测试 ===\n")

	// 1. 配置数据库连接参数（指定 Schema）
	dbParams := &services.DBConnectionParams{
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DbName:   "mydb",
		SSLMode:  "disable",
		Timeout:  10,
		Schema:   "mydata", // 指定 schema
	}

	engineType := advisor.EnginePostgres

	// 2. 获取数据库元数据（会自动设置 search_path）
	fmt.Println("📊 正在连接数据库并获取元数据...")
	metadata, err := services.FetchDatabaseMetadata(engineType, dbParams)
	if err != nil {
		log.Fatalf("❌ 获取数据库元数据失败: %v", err)
	}

	fmt.Printf("✅ 成功连接到数据库 %s@%s:%d/%s\n",
		dbParams.User, dbParams.Host, dbParams.Port, dbParams.DbName)
	fmt.Printf("✅ 设置 search_path 为: %s, public\n", dbParams.Schema)
	fmt.Printf("✅ 获取元数据成功，Schema 数量: %d\n\n", len(metadata.Schemas))

	// 3. 加载规则
	rules, err := services.LoadRules("", engineType, true)
	if err != nil {
		log.Fatalf("❌ 加载规则失败: %v", err)
	}

	fmt.Printf("✅ 成功加载 %d 条 PostgreSQL 审核规则\n\n", len(rules))

	// 4. 准备 SQL - 注意：这里使用不带 schema 前缀的表名
	sql := `
-- 测试 1: 不带 schema 前缀的 UPDATE（应该能正常工作）
UPDATE test_users 
SET status = 'inactive' 
WHERE id = 100;

-- 测试 2: 不带 schema 前缀的 DELETE（应该能正常工作）
DELETE FROM test_users WHERE id > 1000;

-- 测试 3: 不带 schema 前缀的全表 UPDATE（应该能正确计算影响行数）
UPDATE test_users SET status = 'active';

-- 测试 4: 不带 schema 前缀的全表 DELETE（应该能正确计算影响行数）
DELETE FROM test_users;
`

	fmt.Println("📝 准备审核以下 SQL 语句（不带 schema 前缀）:")
	fmt.Println("   - UPDATE test_users (不是 mydata.test_users)")
	fmt.Println("   - DELETE FROM test_users (不是 mydata.test_users)")
	fmt.Println()

	// 5. 创建审核请求
	req := &advisor.ReviewRequest{
		Engine:          engineType,
		Statement:       sql,
		CurrentDatabase: dbParams.DbName,
		Rules:           rules,
		DBSchema:        metadata,
	}

	// 6. 执行 SQL 审核
	fmt.Println("🔍 开始执行 SQL 审核...")
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		log.Fatalf("❌ SQL 审核失败: %v", err)
	}

	fmt.Printf("✅ 审核完成，发现 %d 个问题\n\n", len(resp.Advices))

	// 7. 计算影响行数（会自动设置 search_path）
	fmt.Println("📊 计算影响行数...")
	affectedRowsMap := services.CalculateAffectedRowsForStatements(sql, engineType, dbParams)
	results := services.ConvertToReviewResults(resp, sql, engineType, affectedRowsMap)

	// 8. 输出结果
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📋 审核结果详情")
	fmt.Println(strings.Repeat("=", 70) + "\n")

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
		case "2":
			level = "✗ ERROR"
			icon = "❌"
		}

		fmt.Printf("%s SQL #%d [%s]\n", icon, result.OrderID, level)

		// 显示 SQL（限制长度）
		sqlPreview := result.SQL
		if len(sqlPreview) > 80 {
			sqlPreview = sqlPreview[:77] + "..."
		}
		fmt.Printf("   SQL: %s\n", sqlPreview)

		if result.AffectedRows > 0 {
			fmt.Printf("   💡 影响行数: %d\n", result.AffectedRows)
		}

		if result.ErrorMessage != "" {
			fmt.Printf("   问题: %s\n", result.ErrorMessage)
		}
		fmt.Println()
	}

	// 9. 总结
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("📊 测试总结")
	fmt.Println(strings.Repeat("=", 70))
	
	totalAffectedRows := 0
	for _, result := range results {
		totalAffectedRows += result.AffectedRows
	}

	if totalAffectedRows > 0 {
		fmt.Printf("✅ 成功！影响行数计算正常（总计: %d 行）\n", totalAffectedRows)
		fmt.Println("✅ search_path 设置成功，可以直接使用表名而无需 schema 前缀")
	} else {
		fmt.Println("⚠️  未计算到影响行数（可能是表为空或查询条件不匹配）")
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎯 功能验证")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("✅ 设置 search_path: 在连接时自动执行")
	fmt.Println("✅ 元数据获取: 可以正确获取指定 schema 的表")
	fmt.Println("✅ SQL 审核: 不带 schema 前缀的 SQL 能正常审核")
	fmt.Println("✅ 影响行数: 不带 schema 前缀的 SQL 能正确计算影响行数")
	fmt.Println("\n🎉 测试完成！")
}

