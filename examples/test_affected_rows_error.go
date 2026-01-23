package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/tianyuso/advisorTool/pkg/advisor"
	"github.com/tianyuso/advisorTool/services"
)

func main() {
	fmt.Println("🧪 测试影响行数错误处理")
	fmt.Println(strings.Repeat("=", 70))

	// 测试 SQL - 包含一个会失败的查询
	sql := `
-- 正常的 UPDATE 语句
UPDATE mydata.test_users SET email = 'new@example.com' WHERE id = 1;

-- 会失败的 UPDATE（表不存在）
UPDATE nonexistent_table SET name = 'test' WHERE id = 1;

-- 正常的 DELETE 语句
DELETE FROM mydata.test_orders WHERE id < 100;
`

	// 引擎类型
	engineType := advisor.EnginePostgres

	// 数据库连接参数（用于计算影响行数）
	dbParams := &services.DBConnectionParams{
		Host:     "10.1.1.239",
		Port:     5432,
		User:     "postgres",
		Password: "123456",
		DbName:   "mydata",
		Schema:   "mydata",
		SSLMode:  "disable",
	}

	// 创建审核请求（不使用元数据，所以不会有规则检查错误）
	req := &advisor.ReviewRequest{
		Statement: sql,
		Engine:    engineType,
	}

	// 执行 SQL 审核
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		log.Fatalf("❌ SQL 审核失败: %v", err)
	}

	// 计算影响行数（包含错误信息）
	affectedRowsMap := services.CalculateAffectedRowsForStatements(sql, engineType, dbParams)

	// 打印影响行数详情
	fmt.Println("\n📊 影响行数计算结果:")
	for i, info := range affectedRowsMap {
		if info.Error != "" {
			fmt.Printf("  SQL #%d: Count=%d, Error=%s\n", i+1, info.Count, info.Error)
		} else {
			fmt.Printf("  SQL #%d: Count=%d, Error=nil\n", i+1, info.Count)
		}
	}

	// 转换为结构化结果
	results := services.ConvertToReviewResults(resp, sql, engineType, affectedRowsMap)

	// 输出 JSON 格式结果
	fmt.Println("\n📋 审核结果（JSON 格式）:")
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("❌ JSON 序列化失败: %v", err)
	}
	fmt.Println(string(jsonData))

	// 统计结果
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📈 统计信息:")

	okCount := 0
	warnCount := 0
	errorCount := 0

	for _, result := range results {
		switch result.ErrorLevel {
		case "0":
			okCount++
		case "1":
			warnCount++
		case "2":
			errorCount++
		}

		// 打印每个结果的详细信息
		fmt.Printf("\nSQL #%d [ErrorLevel=%s]:\n", result.OrderID, result.ErrorLevel)
		fmt.Printf("  AffectedRows: %d\n", result.AffectedRows)
		if result.ErrorMessage != "" {
			fmt.Printf("  ErrorMessage: %s\n", result.ErrorMessage)
		}
	}

	fmt.Printf("\n✅ OK: %d\n", okCount)
	fmt.Printf("⚠️  WARNING: %d\n", warnCount)
	fmt.Printf("❌ ERROR: %d\n", errorCount)
	fmt.Println(strings.Repeat("=", 70))
}
