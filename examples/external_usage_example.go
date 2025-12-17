// Package main 演示如何在外部程序中使用 advisorTool/services 包
// 这个示例展示了 services 包现在可以被任何外部程序引用，不会再出现
// "use of internal package not allowed" 错误
package main

import (
	"context"
	"fmt"
	"log"

	"advisorTool/pkg/advisor"
	"advisorTool/services"
)

func main() {
	fmt.Println("=== 外部程序使用 advisorTool/services 包示例 ===\n")

	// 1. 使用 services 包加载默认规则
	engineType := advisor.EngineMySQL
	hasMetadata := false

	rules, err := services.LoadRules("", engineType, hasMetadata)
	if err != nil {
		log.Fatalf("加载规则失败: %v", err)
	}

	fmt.Printf("✅ 成功加载 %d 条规则\n\n", len(rules))

	// 2. 准备要审核的 SQL
	sql := `
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

SELECT * FROM users WHERE id = 1;

UPDATE users SET email = 'new@email.com';
`

	// 3. 创建审核请求
	req := &advisor.ReviewRequest{
		Engine:          engineType,
		Statement:       sql,
		CurrentDatabase: "testdb",
		Rules:           rules,
	}

	// 4. 执行 SQL 审核
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		log.Fatalf("SQL 审核失败: %v", err)
	}

	fmt.Printf("审核完成，发现 %d 个问题\n\n", len(resp.Advices))

	// 5. 使用 services 包的 ConvertToReviewResults 转换结果
	// 注意：这里不需要数据库连接，所以 affectedRowsMap 为空
	affectedRowsMap := make(map[int]int)
	results := services.ConvertToReviewResults(resp, sql, engineType, affectedRowsMap)

	// 6. 输出结果（可以选择 JSON 或表格格式）
	fmt.Println("=== 审核结果 ===")
	for _, result := range results {
		level := "✓ OK"
		if result.ErrorLevel == "1" {
			level = "⚠ WARNING"
		} else if result.ErrorLevel == "2" {
			level = "✗ ERROR"
		}

		fmt.Printf("%d. [%s] %s\n", result.OrderID, level, result.SQL)
		if result.ErrorMessage != "" {
			fmt.Printf("   问题: %s\n", result.ErrorMessage)
		}
		fmt.Println()
	}

	// 7. 也可以使用 services.OutputResults 直接输出格式化结果
	fmt.Println("\n=== 使用 services.OutputResults 输出（JSON 格式） ===")
	if err := services.OutputResults(resp, sql, engineType, "json", nil); err != nil {
		log.Printf("输出结果失败: %v", err)
	}

	// // 8. 演示生成示例配置
	// fmt.Println("\n=== 生成 MySQL 示例配置 ===")
	// sampleConfig := services.GenerateSampleConfig(advisor.EngineMySQL)
	// fmt.Println(sampleConfig)
}
