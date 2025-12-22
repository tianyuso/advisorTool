// 简化的测试程序，用于验证库的外部使用
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tianyuso/advisorTool/pkg/advisor"
	"github.com/tianyuso/advisorTool/services"
)

func main() {
	fmt.Println("=== advisorTool 外部使用测试 ===\n")

	// 1. 测试基础类型
	engineType := advisor.EnginePostgres
	fmt.Printf("✅ 引擎类型: %v\n", engineType)

	// 2. 测试加载规则（不连接数据库）
	rules, err := services.LoadRules("", engineType, false)
	if err != nil {
		log.Fatalf("❌ 加载规则失败: %v", err)
	}
	fmt.Printf("✅ 成功加载 %d 条规则\n", len(rules))

	// 3. 测试简单的 SQL 审核（不需要数据库连接）
	sql := `
	CREATE TABLE test_users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) NOT NULL
	);
	
	SELECT * FROM test_users;
	`

	req := &advisor.ReviewRequest{
		Engine:          engineType,
		Statement:       sql,
		CurrentDatabase: "testdb",
		Rules:           rules,
		DBSchema:        nil, // 不提供元数据
	}

	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		log.Fatalf("❌ SQL 审核失败: %v", err)
	}

	fmt.Printf("✅ 审核完成，发现 %d 个问题\n", len(resp.Advices))

	// 4. 输出审核结果
	for i, advice := range resp.Advices {
		fmt.Printf("   [%d] %s: %s\n", i+1, advice.Status, advice.Title)
	}

	fmt.Println("\n✅ 测试完成！库可以正常使用。")
}

