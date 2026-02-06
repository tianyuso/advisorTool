// Package main demonstrates using advisorTool rules with payload configuration.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tianyuso/advisorTool/pkg/advisor"
)

func main() {
	fmt.Println("=== SQL Advisor 库使用示例 - 规则 Payload 配置 ===\n")

	// 创建带 Payload 的规则
	rules := createRulesWithPayload()

	fmt.Printf("配置了 %d 条规则\n\n", len(rules))
	for i, rule := range rules {
		desc := advisor.GetRuleDescription(rule.Type)
		fmt.Printf("%d. %s\n", i+1, desc)
		fmt.Printf("   类型: %s\n", rule.Type)
		fmt.Printf("   级别: %s\n", rule.Level)
		if rule.Payload != "" {
			fmt.Printf("   配置: %s\n", rule.Payload)
		}
		fmt.Println()
	}

	// 测试 SQL 语句
	testSQL := `
		-- 表名不符合命名规范（应该是小写+下划线）
		CREATE TABLE UserAccounts (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_name VARCHAR(300),  -- VARCHAR 过长
			user_email LONGTEXT,     -- 使用了禁止的类型
			status VARCHAR(50) DEFAULT 'active'
		) CHARSET=latin1;  -- 使用了不在白名单的字符集

		-- INSERT 行数过多
		INSERT INTO products VALUES 
			(1, 'Product 1'), (2, 'Product 2'), (3, 'Product 3'),
			(4, 'Product 4'), (5, 'Product 5'), (6, 'Product 6');
	`

	// 创建审核请求
	req := &advisor.ReviewRequest{
		Engine:          advisor.EngineMySQL,
		Statement:       testSQL,
		Rules:           rules,
		CurrentDatabase: "testdb",
	}

	// 执行审核
	ctx := context.Background()
	resp, err := advisor.SQLReviewCheck(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 审核失败: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	fmt.Printf("=== 审核结果 ===\n")
	fmt.Printf("发现 %d 个问题\n\n", len(resp.Advices))

	if len(resp.Advices) > 0 {
		for i, advice := range resp.Advices {
			statusIcon := getStatusIcon(advice.Status)
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
		os.Exit(2)
	} else if resp.HasWarning {
		fmt.Println("⚠️  审核发现警告")
		os.Exit(1)
	} else {
		fmt.Println("✅ 审核通过")
		os.Exit(0)
	}
}

// createRulesWithPayload 创建带 Payload 配置的规则
func createRulesWithPayload() []*advisor.SQLReviewRule {
	var rules []*advisor.SQLReviewRule

	// 1. 表命名规范：小写+下划线
	namingRule, err := advisor.NewRuleWithPayload(
		advisor.RuleTableNaming,
		advisor.RuleLevelWarning,
		advisor.NamingRulePayload{
			Format:    "^[a-z][a-z0-9_]*$",
			MaxLength: 64,
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建命名规则失败: %v\n", err)
	} else {
		rules = append(rules, namingRule)
	}

	// 2. INSERT 行数限制：最多 5 行
	insertRowLimitRule, err := advisor.NewRuleWithPayload(
		advisor.RuleStatementInsertRowLimit,
		advisor.RuleLevelWarning,
		advisor.NumberTypeRulePayload{
			Number: 5,
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建 INSERT 行数限制规则失败: %v\n", err)
	} else {
		rules = append(rules, insertRowLimitRule)
	}

	// 3. 列类型黑名单：禁止 BLOB 和 TEXT 类型
	typeDisallowRule, err := advisor.NewRuleWithPayload(
		advisor.RuleColumnTypeDisallowList,
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"BLOB", "LONGBLOB", "TEXT", "LONGTEXT"},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建类型黑名单规则失败: %v\n", err)
	} else {
		rules = append(rules, typeDisallowRule)
	}

	// 4. 字符集白名单：只允许 utf8mb4
	charsetRule, err := advisor.NewRuleWithPayload(
		advisor.RuleCharsetAllowlist,
		advisor.RuleLevelWarning,
		advisor.StringArrayTypeRulePayload{
			List: []string{"utf8mb4"},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建字符集白名单规则失败: %v\n", err)
	} else {
		rules = append(rules, charsetRule)
	}

	// 5. VARCHAR 最大长度：不超过 255
	varcharLengthRule, err := advisor.NewRuleWithPayload(
		advisor.RuleColumnMaximumVarcharLength,
		advisor.RuleLevelWarning,
		advisor.NumberTypeRulePayload{
			Number: 255,
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建 VARCHAR 长度规则失败: %v\n", err)
	} else {
		rules = append(rules, varcharLengthRule)
	}

	// 6. 必需列：每个表必须包含 id, created_at, updated_at
	requiredColumnRule, err := advisor.NewRuleWithPayload(
		advisor.RuleRequiredColumn,
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"id", "created_at", "updated_at"},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建必需列规则失败: %v\n", err)
	} else {
		rules = append(rules, requiredColumnRule)
	}

	// 7. 表注释规范：必须有注释，最长 256 字符
	commentRule, err := advisor.NewRuleWithPayload(
		advisor.RuleTableCommentConvention,
		advisor.RuleLevelWarning,
		advisor.CommentConventionRulePayload{
			Required:  true,
			MaxLength: 256,
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建注释规范规则失败: %v\n", err)
	} else {
		rules = append(rules, commentRule)
	}

	return rules
}

// getStatusIcon 返回状态对应的图标
func getStatusIcon(status advisor.AdviceStatus) string {
	switch status {
	case advisor.AdviceStatusError:
		return "❌"
	case advisor.AdviceStatusWarning:
		return "⚠️"
	case advisor.AdviceStatusSuccess:
		return "✅"
	default:
		return "ℹ️"
	}
}











