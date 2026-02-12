package main

import (
	"fmt"
	
	extractor "github.com/tianyuso/advisorTool/extractObject"
)

func main() {
	// 示例 SQL 语句
	sql := `
		SELECT 
			u.id, 
			u.username,
			o.order_id,
			o.total_amount
		FROM 
			mydb.users AS u
		INNER JOIN 
			orders o ON u.id = o.user_id
		WHERE 
			u.status = 'active'
	`

	fmt.Println("=== SQL表名提取示例 ===\n")
	fmt.Println("SQL语句:")
	fmt.Println(sql)
	fmt.Println("\n提取结果:")
	fmt.Println("----------------------------------------")

	// 提取表名
	tables, err := extractor.ExtractTables(extractor.MySQL, sql)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// 输出结果
	if len(tables) == 0 {
		fmt.Println("未找到任何表")
		return
	}

	fmt.Printf("找到 %d 个表:\n\n", len(tables))
	for i, table := range tables {
		fmt.Printf("%d. 表名: %s\n", i+1, table.TBName)
		if table.DBName != "" {
			fmt.Printf("   数据库: %s\n", table.DBName)
		}
		if table.Schema != "" {
			fmt.Printf("   模式: %s\n", table.Schema)
		}
		if table.Alias != "" {
			fmt.Printf("   别名: %s\n", table.Alias)
		}
		fmt.Println()
	}
}





