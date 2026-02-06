package extractobject

import (
	"fmt"
	"log"
)

// Example 展示如何使用extractObject包提取SQL中的表名
func Example() {
	// MySQL示例
	mysqlSQL := `
		SELECT u.id, u.name, o.order_id
		FROM mydb.users AS u
		JOIN orders o ON u.id = o.user_id
		WHERE u.status = 'active'
	`
	
	mysqlTables, err := ExtractTables(MySQL, mysqlSQL)
	if err != nil {
		log.Printf("MySQL解析错误: %v", err)
	} else {
		fmt.Println("MySQL表名:")
		for _, table := range mysqlTables {
			fmt.Printf("  数据库: %s, 模式: %s, 表名: %s, 别名: %s\n",
				table.DBName, table.Schema, table.TBName, table.Alias)
		}
	}

	// PostgreSQL示例
	pgSQL := `
		SELECT p.product_name, c.category_name
		FROM public.products p
		INNER JOIN public.categories c ON p.category_id = c.id
		WHERE c.status = 'active'
	`
	
	pgTables, err := ExtractTables(PostgreSQL, pgSQL)
	if err != nil {
		log.Printf("PostgreSQL解析错误: %v", err)
	} else {
		fmt.Println("\nPostgreSQL表名:")
		for _, table := range pgTables {
			fmt.Printf("  数据库: %s, 模式: %s, 表名: %s, 别名: %s\n",
				table.DBName, table.Schema, table.TBName, table.Alias)
		}
	}

	// SQL Server示例
	sqlserverSQL := `
		SELECT e.employee_id, d.department_name
		FROM HRDatabase.dbo.employees AS e
		LEFT JOIN HRDatabase.dbo.departments d ON e.dept_id = d.id
	`
	
	sqlserverTables, err := ExtractTables(SQLServer, sqlserverSQL)
	if err != nil {
		log.Printf("SQL Server解析错误: %v", err)
	} else {
		fmt.Println("\nSQL Server表名:")
		for _, table := range sqlserverTables {
			fmt.Printf("  数据库: %s, 模式: %s, 表名: %s, 别名: %s\n",
				table.DBName, table.Schema, table.TBName, table.Alias)
		}
	}

	// Oracle示例
	oracleSQL := `
		SELECT e.emp_id, d.dept_name
		FROM hr.employees e, hr.departments d
		WHERE e.dept_id = d.dept_id
	`
	
	oracleTables, err := ExtractTables(Oracle, oracleSQL)
	if err != nil {
		log.Printf("Oracle解析错误: %v", err)
	} else {
		fmt.Println("\nOracle表名:")
		for _, table := range oracleTables {
			fmt.Printf("  数据库: %s, 模式: %s, 表名: %s, 别名: %s\n",
				table.DBName, table.Schema, table.TBName, table.Alias)
		}
	}
}


