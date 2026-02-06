package main

import (
	"fmt"
	"log"
	"strings"

	extractor "github.com/tianyuso/advisorTool/extractObject"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║          SQL表名提取工具 - 综合功能演示                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 测试用例
	testCases := []struct {
		name   string
		dbType extractor.DBType
		sql    string
	}{
		{
			name:   "MySQL - 简单查询",
			dbType: extractor.MySQL,
			sql:    "SELECT * FROM users WHERE status = 'active'",
		},
		{
			name:   "MySQL - 多表JOIN",
			dbType: extractor.MySQL,
			sql: `
				SELECT u.id, o.order_id, p.product_name
				FROM ecommerce.users u
				JOIN ecommerce.orders o ON u.id = o.user_id
				LEFT JOIN products p ON o.product_id = p.id
			`,
		},
		{
			name:   "PostgreSQL - 带schema",
			dbType: extractor.PostgreSQL,
			sql: `
				SELECT p.id, c.name
				FROM public.products p
				INNER JOIN public.categories c ON p.category_id = c.id
			`,
		},
		{
			name:   "PostgreSQL - 子查询",
			dbType: extractor.PostgreSQL,
			sql: `
				SELECT * FROM public.users u
				WHERE u.id IN (
					SELECT user_id FROM public.orders WHERE total > 1000
				)
			`,
		},
		{
			name:   "SQL Server - 三段式表名",
			dbType: extractor.SQLServer,
			sql: `
				SELECT e.id, d.name
				FROM HRDatabase.dbo.employees e
				LEFT JOIN HRDatabase.dbo.departments d ON e.dept_id = d.id
			`,
		},
	}

	// 执行测试
	for i, tc := range testCases {
		fmt.Printf("【测试 %d/%d】%s\n", i+1, len(testCases), tc.name)
		fmt.Println(strings.Repeat("-", 70))
		
		tables, err := extractor.ExtractTables(tc.dbType, tc.sql)
		if err != nil {
			log.Printf("  ❌ 错误: %v\n", err)
			fmt.Println()
			continue
		}

		if len(tables) == 0 {
			fmt.Println("  ⚠️  未找到表")
			fmt.Println()
			continue
		}

		fmt.Printf("  ✅ 找到 %d 个表:\n", len(tables))
		for j, table := range tables {
			fullName := buildFullTableName(table)
			fmt.Printf("     %d. %s", j+1, fullName)
			if table.Alias != "" {
				fmt.Printf(" (别名: %s)", table.Alias)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      测试完成！                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}

func buildFullTableName(table extractor.TableInfo) string {
	parts := []string{}
	if table.DBName != "" {
		parts = append(parts, table.DBName)
	}
	if table.Schema != "" {
		parts = append(parts, table.Schema)
	}
	if table.TBName != "" {
		parts = append(parts, table.TBName)
	}
	
	if len(parts) == 0 {
		return "<未知>"
	}
	
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "."
		}
		result += part
	}
	return result
}

