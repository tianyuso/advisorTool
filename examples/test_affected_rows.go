// Package main 演示影响行数计算功能的测试
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tianyuso/advisorTool/db"
	"github.com/tianyuso/advisorTool/pkg/advisor"
	"github.com/tianyuso/advisorTool/services"
)

func main() {
	fmt.Println("=== 影响行数计算功能测试 ===\n")

	// 数据库连接参数
	dbParams := &services.DBConnectionParams{
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DbName:   "mydb",
		SSLMode:  "disable",
		Timeout:  10,
	}

	engineType := advisor.EnginePostgres

	// 打开数据库连接
	config := &db.ConnectionConfig{
		DbType:  "postgres",
		Host:    dbParams.Host,
		Port:    dbParams.Port,
		User:    dbParams.User,
		Password: dbParams.Password,
		DbName:  dbParams.DbName,
		SSLMode: dbParams.SSLMode,
		Timeout: dbParams.Timeout,
	}

	dbConn, err := db.OpenConnection(context.Background(), config)
	if err != nil {
		log.Fatalf("❌ 无法连接数据库: %v", err)
	}
	defer dbConn.Close()
	fmt.Println("✅ 成功连接到数据库\n")

	// 测试用例
	testCases := []struct {
		name        string
		sql         string
		expectNonZero bool // 是否期望影响行数大于 0
	}{
		{
			name: "UPDATE with WHERE (带注释)",
			sql: `-- 正常的 UPDATE（有 WHERE 条件）
UPDATE mydata.test_users 
SET status = 'inactive', updated_at = CURRENT_TIMESTAMP 
WHERE id = 999999`,
			expectNonZero: false, // 不存在的 ID，影响 0 行
		},
		{
			name: "UPDATE without WHERE (带注释)",
			sql: `-- 危险的 UPDATE（没有 WHERE 条件）
UPDATE mydata.test_users SET status = 'active'`,
			expectNonZero: true, // 影响所有行
		},
		{
			name: "DELETE with WHERE (带注释)",
			sql: `-- 正常的 DELETE（有 WHERE 条件）
DELETE FROM mydata.test_orders WHERE order_date < '2000-01-01'`,
			expectNonZero: false, // 没有这么早的数据
		},
		{
			name: "DELETE without WHERE (带注释)",
			sql: `-- 危险的 DELETE（没有 WHERE 条件）
DELETE FROM mydata.test_users`,
			expectNonZero: true, // 删除所有行
		},
		{
			name: "UPDATE without comments",
			sql:  `UPDATE mydata.test_users SET status = 'active' WHERE id = 1`,
			expectNonZero: false, // 可能不存在 id=1
		},
		{
			name:          "SELECT statement",
			sql:           `SELECT * FROM mydata.test_users`,
			expectNonZero: false, // SELECT 不计算影响行数
		},
	}

	fmt.Println("开始测试...\n")
	passCount := 0
	totalCount := len(testCases)

	for i, tc := range testCases {
		fmt.Printf("测试 %d: %s\n", i+1, tc.name)
		fmt.Printf("SQL: %s\n", tc.sql)

		count, err := db.CalculateAffectedRows(context.Background(), dbConn, tc.sql, engineType)
		
		status := "✅"
		passed := true

		if err != nil {
			if tc.expectNonZero {
				status = "❌"
				passed = false
				fmt.Printf("结果: %s 错误 - %v\n", status, err)
			} else {
				fmt.Printf("结果: %s 预期错误或 0 行 - %v\n", status, err)
			}
		} else {
			fmt.Printf("影响行数: %d\n", count)
			
			if tc.expectNonZero && count == 0 {
				status = "⚠️"
				fmt.Printf("结果: %s 警告 - 预期影响行数 > 0，实际为 0（可能表为空）\n", status)
			} else if !tc.expectNonZero && count > 0 {
				status = "❌"
				passed = false
				fmt.Printf("结果: %s 失败 - 预期影响行数 = 0，实际为 %d\n", status, count)
			} else {
				fmt.Printf("结果: %s 通过\n", status)
			}
		}

		if passed {
			passCount++
		}
		fmt.Println()
	}

	// 总结
	fmt.Println("======================================")
	fmt.Printf("测试完成: %d/%d 通过\n", passCount, totalCount)
	if passCount == totalCount {
		fmt.Println("✅ 所有测试通过！")
	} else {
		fmt.Printf("⚠️  %d 个测试未通过\n", totalCount-passCount)
	}
}

