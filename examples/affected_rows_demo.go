package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"advisorTool/db"
	"advisorTool/pkg/advisor"
)

// 演示如何使用影响行数计算功能
func main() {
	// 示例 1: 不连接数据库的情况（affected_rows 为 0）
	fmt.Println("========================================")
	fmt.Println("示例 1: 不连接数据库")
	fmt.Println("========================================")

	statement1 := "UPDATE users SET status = 1 WHERE created_at < '2024-01-01'"
	count1, err := db.CalculateAffectedRows(context.Background(), nil, statement1, advisor.EngineMySQL)
	fmt.Printf("SQL: %s\n", statement1)
	fmt.Printf("影响行数: %d (错误: %v)\n\n", count1, err)

	// 示例 2: 连接数据库的情况（需要实际数据库）
	// 请根据实际情况修改连接参数
	fmt.Println("========================================")
	fmt.Println("示例 2: 连接数据库计算影响行数")
	fmt.Println("========================================")

	config := &db.ConnectionConfig{
		DbType:   "mysql",
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "password", // 请修改为实际密码
		DbName:   "testdb",   // 请修改为实际数据库名
		Timeout:  5,
	}

	// 尝试连接数据库
	conn, err := db.OpenConnection(context.Background(), config)
	if err != nil {
		log.Printf("无法连接数据库: %v\n", err)
		log.Println("跳过需要数据库连接的示例")
	} else {
		defer conn.Close()

		// 测试各种 SQL 语句
		testStatements := []string{
			"UPDATE users SET status = 1 WHERE id > 1000",
			"DELETE FROM logs WHERE created_at < '2023-01-01'",
			"UPDATE orders o INNER JOIN customers c ON o.customer_id = c.id SET o.status = 'completed' WHERE c.vip = 1",
		}

		for i, stmt := range testStatements {
			fmt.Printf("\n测试 %d:\n", i+1)
			fmt.Printf("SQL: %s\n", stmt)

			count, err := db.CalculateAffectedRows(context.Background(), conn, stmt, advisor.EngineMySQL)
			if err != nil {
				fmt.Printf("错误: %v\n", err)
			} else {
				fmt.Printf("预计影响行数: %d\n", count)
			}
		}
	}

	// 示例 3: 测试不同数据库引擎的 SQL 改写
	fmt.Println("\n========================================")
	fmt.Println("示例 3: SQL 改写测试（不执行）")
	fmt.Println("========================================")

	testCases := []struct {
		engine advisor.Engine
		sql    string
	}{
		{advisor.EngineMySQL, "UPDATE users SET name = 'test' WHERE id = 1"},
		{advisor.EnginePostgres, "DELETE FROM logs WHERE created_at < NOW() - INTERVAL '30 days'"},
		{advisor.EngineMSSQL, "UPDATE orders SET status = 'completed' WHERE order_date < '2024-01-01'"},
		{advisor.EngineOracle, "DELETE FROM audit_logs WHERE log_date < SYSDATE - 90"},
	}

	for _, tc := range testCases {
		fmt.Printf("\n引擎: %s\n", tc.engine.String())
		fmt.Printf("原始 SQL: %s\n", tc.sql)

		// 注意：这里只是展示改写逻辑，不实际执行
		// 实际使用中，改写后的 SQL 会在 CalculateAffectedRows 内部执行
		fmt.Println("(改写后的 SQL 将在 CalculateAffectedRows 内部生成)")
	}

	fmt.Println("\n========================================")
	fmt.Println("演示完成")
	fmt.Println("========================================")
}

// demonstrateRewriteLogic 演示 SQL 改写逻辑（仅用于说明）
func demonstrateRewriteLogic() {
	examples := map[string]string{
		"MySQL 单表 UPDATE":      "UPDATE users SET name='x' WHERE id>100 → SELECT COUNT(1) FROM users WHERE id>100",
		"MySQL 连表 UPDATE":      "UPDATE t1 JOIN t2 ON t1.id=t2.id SET t1.x=t2.x → SELECT COUNT(1) FROM t1 JOIN t2 ON t1.id=t2.id",
		"PostgreSQL 连表 UPDATE": "UPDATE t1 SET x=t2.x FROM t2 WHERE t1.id=t2.id → SELECT COUNT(1) FROM t1 INNER JOIN t2 ON t1.id=t2.id",
		"SQL Server 连表 UPDATE": "UPDATE t1 SET t1.x=t2.x FROM t1 JOIN t2 ON t1.id=t2.id → SELECT COUNT(1) FROM t1 JOIN t2 ON t1.id=t2.id",
	}

	fmt.Println("\n改写规则示例：")
	fmt.Println("========================================")
	for name, example := range examples {
		fmt.Printf("%s:\n  %s\n\n", name, example)
	}
}

// 辅助函数：创建测试表（可选）
func createTestTables(conn *sql.DB) error {
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100),
			status INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			last_login TIMESTAMP NULL
		)`,
		`CREATE TABLE IF NOT EXISTS logs (
			id INT PRIMARY KEY AUTO_INCREMENT,
			message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id INT PRIMARY KEY AUTO_INCREMENT,
			customer_id INT,
			customer_name VARCHAR(100),
			status VARCHAR(50),
			total_amount DECIMAL(10,2),
			order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS customers (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100),
			status VARCHAR(50),
			vip BOOLEAN DEFAULT FALSE,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range sqls {
		if _, err := conn.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// 辅助函数：插入测试数据（可选）
func insertTestData(conn *sql.DB) error {
	sqls := []string{
		"INSERT INTO users (name, status) VALUES ('User1', 0), ('User2', 1), ('User3', 0)",
		"INSERT INTO logs (message) VALUES ('Log entry 1'), ('Log entry 2')",
		"INSERT INTO customers (name, status, vip) VALUES ('Customer1', 'active', TRUE), ('Customer2', 'deleted', FALSE)",
		"INSERT INTO orders (customer_id, customer_name, status) VALUES (1, 'Customer1', 'pending'), (2, 'Customer2', 'completed')",
	}

	for _, query := range sqls {
		if _, err := conn.Exec(query); err != nil {
			// 忽略重复插入错误
			continue
		}
	}

	return nil
}
