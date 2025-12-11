// Package main demonstrates advanced usage with payload configuration and database metadata
package main

import (
	"context"
	"fmt"
	"os"

	"advisorTool/pkg/advisor"
	"demo/common"
)

func main() {
	fmt.Println("=== SQL Advisor Tool - 高级用法示例 ===")
	fmt.Println("使用完整规则集 + Payload 配置 + 数据库元数据\n")

	// 示例 1: 使用 Payload 配置命名规范
	example1()

	fmt.Println("\n" + "="*60 + "\n")

	// 示例 2: 综合配置（完整规则集 + 自定义规则）
	example2()

	fmt.Println("\n" + "="*60 + "\n")

	// 示例 3: 使用数据库元数据进行审核
	example3()

	fmt.Println("\n" + "="*60 + "\n")

	// 示例 4: 生产环境完整配置
	example4()
}

// example1 演示命名规范配置
func example1() {
	fmt.Println("示例 1: 命名规范配置")
	fmt.Println("在完整规则集基础上添加自定义命名规范\n")

	// 获取基础规则集
	baseRules := common.GetDefaultRules(advisor.EngineMySQL, false)

	// 添加自定义命名规范规则
	tableNamingRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleTableNaming,
		advisor.RuleLevelWarning,
		advisor.NamingRulePayload{
			Format:    "^[a-z][a-z0-9_]*$", // 小写字母开头，使用下划线
			MaxLength: 64,
		},
	)

	columnNamingRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleColumnNaming,
		advisor.RuleLevelWarning,
		advisor.NamingRulePayload{
			Format:    "^[a-z][a-z0-9_]*$",
			MaxLength: 64,
		},
	)

	// 合并规则
	allRules := append(baseRules, tableNamingRule, columnNamingRule)
	fmt.Printf("总规则数: %d 条（基础规则 + 自定义规则）\n\n", len(allRules))

	// 测试不同的表名
	testCases := []struct {
		desc string
		sql  string
	}{
		{
			desc: "符合规范的表名",
			sql:  "CREATE TABLE user_orders (id INT PRIMARY KEY, user_id INT);",
		},
		{
			desc: "驼峰命名（不符合规范）",
			sql:  "CREATE TABLE UserOrders (id INT PRIMARY KEY, user_id INT);",
		},
		{
			desc: "数字开头（不符合规范）",
			sql:  "CREATE TABLE 123_orders (id INT PRIMARY KEY, amount DECIMAL);",
		},
	}

	for _, tc := range testCases {
		fmt.Printf("%s:\n", tc.desc)
		fmt.Printf("SQL: %s\n", tc.sql)

		req := &advisor.ReviewRequest{
			Engine:    advisor.EngineMySQL,
			Statement: tc.sql,
			Rules:     allRules,
		}

		resp, err := advisor.SQLReviewCheck(context.Background(), req)
		if err != nil {
			fmt.Printf("  ❌ 审核失败: %v\n\n", err)
			continue
		}

		if len(resp.Advices) == 0 {
			fmt.Println("  ✅ 通过审核")
		} else {
			for _, advice := range resp.Advices {
				icon := "⚠️"
				if advice.Status == advisor.AdviceStatusError {
					icon = "❌"
				}
				fmt.Printf("  %s %s\n", icon, advice.Content)
			}
		}
		fmt.Println()
	}
}

// example2 演示综合配置
func example2() {
	fmt.Println("示例 2: 综合配置（完整规则集 + 自定义限制）")
	fmt.Println("基础规则 + 类型限制 + 数值限制\n")

	// 获取基础规则集
	baseRules := common.GetDefaultRules(advisor.EngineMySQL, false)

	// 添加列类型黑名单
	typeRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleColumnTypeDisallowList,
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"BLOB", "LONGBLOB", "TEXT", "MEDIUMTEXT", "LONGTEXT"},
		},
	)

	// 添加必需列
	requiredColumnRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleRequiredColumn,
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"id", "created_at", "updated_at"},
		},
	)

	// 添加字符集白名单
	charsetRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleCharsetAllowlist,
		advisor.RuleLevelWarning,
		advisor.StringArrayTypeRulePayload{
			List: []string{"utf8mb4", "utf8"},
		},
	)

	// 合并规则
	allRules := append(baseRules, typeRule, requiredColumnRule, charsetRule)
	fmt.Printf("总规则数: %d 条\n\n", len(allRules))

	// 测试 SQL
	testCases := []struct {
		desc string
		sql  string
	}{
		{
			desc: "完全符合规范",
			sql: `CREATE TABLE user_profiles (
				id BIGINT PRIMARY KEY AUTO_INCREMENT,
				username VARCHAR(50) NOT NULL,
				email VARCHAR(100),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		},
		{
			desc: "使用了禁止的 TEXT 类型",
			sql: `CREATE TABLE posts (
				id BIGINT PRIMARY KEY,
				content TEXT,
				created_at TIMESTAMP,
				updated_at TIMESTAMP
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		},
		{
			desc: "缺少必需列",
			sql: `CREATE TABLE products (
				product_id INT PRIMARY KEY,
				name VARCHAR(100)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		},
	}

	for i, tc := range testCases {
		fmt.Printf("[%d] %s:\n", i+1, tc.desc)

		req := &advisor.ReviewRequest{
			Engine:    advisor.EngineMySQL,
			Statement: tc.sql,
			Rules:     allRules,
		}

		resp, err := advisor.SQLReviewCheck(context.Background(), req)
		if err != nil {
			fmt.Printf("  ❌ 审核失败: %v\n\n", err)
			continue
		}

		if len(resp.Advices) == 0 {
			fmt.Println("  ✅ 通过所有审核规则")
		} else {
			fmt.Printf("  发现 %d 个问题:\n", len(resp.Advices))
			for _, advice := range resp.Advices {
				icon := "⚠️"
				statusText := "WARNING"
				if advice.Status == advisor.AdviceStatusError {
					icon = "❌"
					statusText = "ERROR"
				}
				fmt.Printf("    %s [%s] %s\n", icon, statusText, advice.Content)
			}
		}
		fmt.Println()
	}
}

// example3 演示使用数据库元数据进行审核
func example3() {
	fmt.Println("示例 3: 使用数据库元数据进行审核")
	fmt.Println("提示: 需要配置数据库连接才能启用元数据相关规则\n")

	// 数据库连接配置（默认为空）
	// 如需测试，请取消注释并填写实际的数据库连接信息
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
		fmt.Println("⚠️ 未配置数据库连接，使用静态分析模式")
		fmt.Println("以下规则将被跳过:")
		fmt.Println("  • column.no-null (需要现有表结构)")
		fmt.Println("  • column.set-default-for-not-null (需要表元数据)")
		fmt.Println("  • column.require-default (需要表元数据)")
		fmt.Println("  • schema.backward-compatibility (需要变更前后对比)")
		fmt.Println()
	} else {
		fmt.Println("✅ 已连接数据库，启用完整规则集（包含元数据规则）\n")
	}

	// 获取完整规则集
	rules := common.GetDefaultRules(advisor.EngineMySQL, hasMetadata)
	fmt.Printf("已加载 %d 条审核规则 (hasMetadata=%v)\n\n", len(rules), hasMetadata)

	// 测试 SQL
	sql := `
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
ALTER TABLE orders ADD COLUMN status VARCHAR(20) NOT NULL;
`

	fmt.Println("待审核 SQL:")
	fmt.Println(sql)

	req := &advisor.ReviewRequest{
		Engine:          advisor.EngineMySQL,
		Statement:       sql,
		Rules:           rules,
		DBSchema:        metadata,
		CurrentDatabase: "test_db",
	}

	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ 审核失败: %v\n", err)
		return
	}

	// 输出结果
	fmt.Println("\n审核结果:")
	fmt.Println("=" * 60)

	if len(resp.Advices) == 0 {
		fmt.Println("✅ 通过审核")
	} else {
		fmt.Printf("发现 %d 个问题:\n\n", len(resp.Advices))
		for i, advice := range resp.Advices {
			icon := "⚠️"
			statusText := "WARNING"
			if advice.Status == advisor.AdviceStatusError {
				icon = "❌"
				statusText = "ERROR"
			}
			fmt.Printf("%d. %s [%s] %s\n", i+1, icon, statusText, advice.Title)
			fmt.Printf("   %s\n", advice.Content)
		}
	}

	if !hasMetadata {
		fmt.Println("\n💡 提示: 配置数据库连接后可以启用更多高级规则")
	}
}

// example4 演示生产环境完整配置
func example4() {
	fmt.Println("示例 4: 生产环境完整配置")
	fmt.Println("严格模式：完整规则集 + 自定义限制 + 元数据检查\n")

	// 数据库配置（实际使用时填写真实值）
	var dbConfig *common.DBConfig = nil

	// 获取元数据
	metadata, _ := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
	hasMetadata := (metadata != nil)

	// 获取基础规则集
	baseRules := common.GetDefaultRules(advisor.EngineMySQL, hasMetadata)

	// 添加生产环境的严格规则

	// 1. 表命名规范
	tableNamingRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleTableNaming,
		advisor.RuleLevelError,
		advisor.NamingRulePayload{
			Format:    "^[a-z][a-z0-9_]*$",
			MaxLength: 64,
		},
	)

	// 2. 列命名规范
	columnNamingRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleColumnNaming,
		advisor.RuleLevelError,
		advisor.NamingRulePayload{
			Format:    "^[a-z][a-z0-9_]*$",
			MaxLength: 64,
		},
	)

	// 3. 必需列
	requiredColumnRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleRequiredColumn,
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"id", "created_at", "updated_at"},
		},
	)

	// 4. 类型黑名单
	typeRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleColumnTypeDisallowList,
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"BLOB", "TEXT"},
		},
	)

	// 5. 字符集要求
	charsetRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleCharsetAllowlist,
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"utf8mb4"},
		},
	)

	// 6. 表注释规范
	tableCommentRule, _ := advisor.NewRuleWithPayload(
		advisor.RuleTableCommentConvention,
		advisor.RuleLevelWarning,
		advisor.CommentConventionRulePayload{
			Required:  true,
			MaxLength: 256,
		},
	)

	// 合并所有规则
	allRules := append(baseRules,
		tableNamingRule,
		columnNamingRule,
		requiredColumnRule,
		typeRule,
		charsetRule,
		tableCommentRule,
	)

	fmt.Printf("生产环境规则总数: %d 条\n", len(allRules))
	fmt.Printf("规则级别分布:\n")

	errorCount := 0
	warningCount := 0
	for _, rule := range allRules {
		if rule.Level == advisor.RuleLevelError {
			errorCount++
		} else if rule.Level == advisor.RuleLevelWarning {
			warningCount++
		}
	}
	fmt.Printf("  • ERROR 级别: %d 条\n", errorCount)
	fmt.Printf("  • WARNING 级别: %d 条\n\n", warningCount)

	// 测试 SQL
	sql := `
CREATE TABLE user_orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    order_no VARCHAR(50) NOT NULL COMMENT '订单号',
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '订单金额',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '订单状态',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户订单表';
`

	fmt.Println("待审核 SQL:")
	fmt.Println(sql)

	req := &advisor.ReviewRequest{
		Engine:    advisor.EngineMySQL,
		Statement: sql,
		Rules:     allRules,
		DBSchema:  metadata,
	}

	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		fmt.Printf("\n❌ 审核失败: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	fmt.Println("\n生产环境审核结果:")
	fmt.Println("=" * 60)

	if len(resp.Advices) == 0 {
		fmt.Println("✅ 通过所有生产环境审核规则！")
		fmt.Println("该 SQL 符合生产环境部署标准。")
	} else {
		fmt.Printf("发现 %d 个问题:\n\n", len(resp.Advices))

		// 分组显示
		errorAdvices := []*advisor.Advice{}
		warningAdvices := []*advisor.Advice{}

		for _, advice := range resp.Advices {
			if advice.Status == advisor.AdviceStatusError {
				errorAdvices = append(errorAdvices, advice)
			} else {
				warningAdvices = append(warningAdvices, advice)
			}
		}

		if len(errorAdvices) > 0 {
			fmt.Printf("❌ 错误 (%d 个) - 必须修复:\n", len(errorAdvices))
			for i, advice := range errorAdvices {
				fmt.Printf("%d. [%s] %s\n", i+1, advice.Title, advice.Content)
			}
			fmt.Println()
		}

		if len(warningAdvices) > 0 {
			fmt.Printf("⚠️ 警告 (%d 个) - 建议修复:\n", len(warningAdvices))
			for i, advice := range warningAdvices {
				fmt.Printf("%d. [%s] %s\n", i+1, advice.Title, advice.Content)
			}
		}
	}

	// 退出码
	if resp.HasError {
		fmt.Println("\n❌ 状态: 存在错误，禁止部署到生产环境")
		os.Exit(2)
	} else if resp.HasWarning {
		fmt.Println("\n⚠️ 状态: 存在警告，需人工确认后部署")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ 状态: 审核通过，可以部署到生产环境")
	}
}
