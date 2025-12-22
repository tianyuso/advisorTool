// Package main demonstrates basic usage of advisorTool as a Go library.
package main

import (
	"context"
	"fmt"

	"github.com/tianyuso/advisorTool/pkg/advisor"
)

func main() {
	fmt.Println("=== SQL Advisor 库基础使用示例 ===\n")

	// 定义审核规则
	rules := []*advisor.SQLReviewRule{
		advisor.NewRule(advisor.RuleStatementNoSelectAll, advisor.RuleLevelWarning),
		advisor.NewRule(advisor.RuleStatementRequireWhereForUpdateDelete, advisor.RuleLevelError),
		advisor.NewRule(advisor.RuleTableRequirePK, advisor.RuleLevelError),
	}

	fmt.Printf("启用规则数: %d\n", len(rules))
	for i, rule := range rules {
		desc := advisor.GetRuleDescription(rule.Type)
		fmt.Printf("  %d. %s - %s (%s)\n", i+1, rule.Type, desc, rule.Level)
	}
	fmt.Println()

	// 测试 SQL 语句
	testSQL := `
		SELECT * FROM users WHERE id = 1;
		DELETE FROM orders;
		UPDATE products SET price = 100;
	`

	// 创建审核请求
	req := &advisor.ReviewRequest{
		Engine:          advisor.EngineMySQL,
		Statement:       testSQL,
		Rules:           rules,
		CurrentDatabase: "mydb",
	}

	// 执行审核
	ctx := context.Background()
	resp, err := advisor.SQLReviewCheck(ctx, req)
	if err != nil {
		fmt.Printf("❌ 审核失败: %v\n", err)
		return
	}

	// 输出结果
	fmt.Printf("=== 审核结果 ===\n")
	fmt.Printf("发现 %d 个问题\n\n", len(resp.Advices))

	if len(resp.Advices) > 0 {
		for i, advice := range resp.Advices {
			var statusIcon string
			switch advice.Status {
			case advisor.AdviceStatusError:
				statusIcon = "❌"
			case advisor.AdviceStatusWarning:
				statusIcon = "⚠️"
			default:
				statusIcon = "ℹ️"
			}

			fmt.Printf("%d. %s [%s] %s\n", i+1, statusIcon, advice.Status, advice.Title)
			fmt.Printf("   %s\n", advice.Content)
			if advice.StartPosition != nil {
				fmt.Printf("   位置: 行 %d, 列 %d\n",
					advice.StartPosition.Line,
					advice.StartPosition.Column)
			}
			fmt.Println()
		}
	}

	// 最终结论
	fmt.Println("=== 结论 ===")
	if resp.HasError {
		fmt.Println("❌ 审核发现错误!")
	} else if resp.HasWarning {
		fmt.Println("⚠️  审核发现警告")
	} else {
		fmt.Println("✅ 审核通过")
	}
}

