package main

import (
	"fmt"
	"log"

	extractor "github.com/tianyuso/advisorTool/extractObject"
)

func main() {
	// 示例1: MySQL - 简单查询
	fmt.Println("========== MySQL 示例 ==========")
	mysqlExample1()
	fmt.Println()

	// 示例2: MySQL - 复杂JOIN查询
	mysqlExample2()
	fmt.Println()

	// 示例3: PostgreSQL - 带schema的查询
	fmt.Println("========== PostgreSQL 示例 ==========")
	postgresqlExample()
	fmt.Println()

	// 示例4: SQL Server - 三部分表名
	fmt.Println("========== SQL Server 示例 ==========")
	sqlserverExample()
	fmt.Println()

	// 示例5: Oracle - 传统语法
	fmt.Println("========== Oracle 示例 ==========")
	oracleExample()
	fmt.Println()
}

// MySQL示例1: 简单查询
func mysqlExample1() {
	sql := `SELECT * FROM users WHERE status = 'active'`
	
	tables, err := extractor.ExtractTables(extractor.MySQL, sql)
	if err != nil {
		log.Printf("错误: %v", err)
		return
	}

	fmt.Println("SQL:", sql)
	printTables(tables)
}

// MySQL示例2: 复杂JOIN查询
func mysqlExample2() {
	sql := `
		SELECT 
			u.id, 
			u.name, 
			o.order_id, 
			o.total_amount,
			p.product_name
		FROM 
			ecommerce.users AS u
		INNER JOIN 
			ecommerce.orders o ON u.id = o.user_id
		LEFT JOIN 
			products p ON o.product_id = p.id
		WHERE 
			u.status = 'active' 
			AND o.order_date >= '2024-01-01'
	`
	
	tables, err := extractor.ExtractTables(extractor.MySQL, sql)
	if err != nil {
		log.Printf("错误: %v", err)
		return
	}

	fmt.Println("SQL: [复杂JOIN查询]")
	printTables(tables)
}

// PostgreSQL示例: 带schema的查询
func postgresqlExample() {
	sql := `
		SELECT 
			p.product_id,
			p.product_name,
			c.category_name,
			s.supplier_name
		FROM 
			public.products p
		INNER JOIN 
			public.categories c ON p.category_id = c.id
		LEFT JOIN 
			inventory.suppliers s ON p.supplier_id = s.id
		WHERE 
			c.status = 'active'
			AND p.price > 100
	`
	
	tables, err := extractor.ExtractTables(extractor.PostgreSQL, sql)
	if err != nil {
		log.Printf("错误: %v", err)
		return
	}

	fmt.Println("SQL: [带schema的复杂查询]")
	printTables(tables)
}

// SQL Server示例: 三部分表名
func sqlserverExample() {
	sql := `
		SELECT 
			e.employee_id,
			e.first_name,
			e.last_name,
			d.department_name,
			l.location_name
		FROM 
			HRDatabase.dbo.employees AS e
		LEFT JOIN 
			HRDatabase.dbo.departments d ON e.department_id = d.id
		LEFT JOIN 
			HRDatabase.dbo.locations l ON d.location_id = l.id
		WHERE 
			e.status = 'active'
	`
	
	tables, err := extractor.ExtractTables(extractor.SQLServer, sql)
	if err != nil {
		log.Printf("错误: %v", err)
		return
	}

	fmt.Println("SQL: [SQL Server三部分表名]")
	printTables(tables)
}

// Oracle示例: 传统语法
func oracleExample() {
	sql := `
		SELECT 
			e.emp_id,
			e.emp_name,
			d.dept_name,
			j.job_title
		FROM 
			hr.employees e,
			hr.departments d,
			hr.jobs j
		WHERE 
			e.dept_id = d.dept_id
			AND e.job_id = j.job_id
			AND e.status = 'ACTIVE'
	`
	
	tables, err := extractor.ExtractTables(extractor.Oracle, sql)
	if err != nil {
		log.Printf("错误: %v", err)
		return
	}

	fmt.Println("SQL: [Oracle传统JOIN语法]")
	printTables(tables)
}

// 打印表信息
func printTables(tables []extractor.TableInfo) {
	if len(tables) == 0 {
		fmt.Println("未找到任何表")
		return
	}

	fmt.Printf("找到 %d 个表:\n", len(tables))
	fmt.Printf("%-20s %-20s %-30s %-20s\n", "数据库名", "模式名", "表名", "别名")
	fmt.Println(string(make([]byte, 90)))
	
	for _, table := range tables {
		dbName := table.DBName
		if dbName == "" {
			dbName = "-"
		}
		schema := table.Schema
		if schema == "" {
			schema = "-"
		}
		alias := table.Alias
		if alias == "" {
			alias = "-"
		}
		
		fmt.Printf("%-20s %-20s %-30s %-20s\n", dbName, schema, table.TBName, alias)
	}
}


