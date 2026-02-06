package extractobject

import (
	"testing"
)

func TestMySQLExtractor(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected []TableInfo
	}{
		{
			name: "简单SELECT",
			sql:  "SELECT * FROM users",
			expected: []TableInfo{
				{DBName: "", Schema: "", TBName: "users", Alias: ""},
			},
		},
		{
			name: "带数据库名和别名",
			sql:  "SELECT u.id FROM mydb.users AS u",
			expected: []TableInfo{
				{DBName: "mydb", Schema: "", TBName: "users", Alias: "u"},
			},
		},
		{
			name: "JOIN查询",
			sql: `
				SELECT u.id, o.order_id
				FROM mydb.users u
				JOIN orders o ON u.id = o.user_id
			`,
			expected: []TableInfo{
				{DBName: "mydb", Schema: "", TBName: "users", Alias: "u"},
				{DBName: "", Schema: "", TBName: "orders", Alias: "o"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tables, err := ExtractTables(MySQL, tt.sql)
			if err != nil {
				t.Fatalf("ExtractTables() error = %v", err)
			}

			if len(tables) != len(tt.expected) {
				t.Errorf("表数量不匹配: got %d, want %d", len(tables), len(tt.expected))
				return
			}

			// 验证每个表信息
			for i, table := range tables {
				if i >= len(tt.expected) {
					break
				}
				exp := tt.expected[i]
				if table.TBName != exp.TBName {
					t.Errorf("表名[%d]不匹配: got %s, want %s", i, table.TBName, exp.TBName)
				}
			}
		})
	}
}

func TestPostgreSQLExtractor(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected []TableInfo
	}{
		{
			name: "简单SELECT",
			sql:  "SELECT * FROM users",
			expected: []TableInfo{
				{DBName: "", Schema: "", TBName: "users", Alias: ""},
			},
		},
		{
			name: "带schema和别名",
			sql:  "SELECT p.id FROM public.products AS p",
			expected: []TableInfo{
				{DBName: "", Schema: "public", TBName: "products", Alias: "p"},
			},
		},
		{
			name: "JOIN查询",
			sql: `
				SELECT p.product_name, c.category_name
				FROM public.products p
				INNER JOIN public.categories c ON p.category_id = c.id
			`,
			expected: []TableInfo{
				{DBName: "", Schema: "public", TBName: "products", Alias: "p"},
				{DBName: "", Schema: "public", TBName: "categories", Alias: "c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tables, err := ExtractTables(PostgreSQL, tt.sql)
			if err != nil {
				t.Fatalf("ExtractTables() error = %v", err)
			}

			if len(tables) < len(tt.expected) {
				t.Errorf("表数量不足: got %d, want at least %d", len(tables), len(tt.expected))
				return
			}

			// 验证关键表信息存在
			for _, exp := range tt.expected {
				found := false
				for _, table := range tables {
					if table.TBName == exp.TBName && table.Schema == exp.Schema {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("未找到表: schema=%s, table=%s", exp.Schema, exp.TBName)
				}
			}
		})
	}
}

func TestSQLServerExtractor(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected []TableInfo
	}{
		{
			name: "简单SELECT",
			sql:  "SELECT * FROM employees",
			expected: []TableInfo{
				{DBName: "", Schema: "", TBName: "employees", Alias: ""},
			},
		},
		{
			name: "带数据库和schema",
			sql:  "SELECT e.id FROM HRDatabase.dbo.employees AS e",
			expected: []TableInfo{
				{DBName: "HRDatabase", Schema: "dbo", TBName: "employees", Alias: "e"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tables, err := ExtractTables(SQLServer, tt.sql)
			if err != nil {
				t.Fatalf("ExtractTables() error = %v", err)
			}

			if len(tables) < len(tt.expected) {
				t.Errorf("表数量不足: got %d, want at least %d", len(tables), len(tt.expected))
			}
		})
	}
}

func TestOracleExtractor(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected []TableInfo
	}{
		{
			name: "简单SELECT",
			sql:  "SELECT * FROM employees",
			expected: []TableInfo{
				{DBName: "", Schema: "", TBName: "employees", Alias: ""},
			},
		},
		{
			name: "带schema",
			sql:  "SELECT e.emp_id FROM hr.employees e",
			expected: []TableInfo{
				{DBName: "", Schema: "hr", TBName: "employees", Alias: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tables, err := ExtractTables(Oracle, tt.sql)
			if err != nil {
				t.Fatalf("ExtractTables() error = %v", err)
			}

			if len(tables) < len(tt.expected) {
				t.Errorf("表数量不足: got %d, want at least %d", len(tables), len(tt.expected))
			}
		})
	}
}

func TestUnsupportedDBType(t *testing.T) {
	_, err := ExtractTables("UNSUPPORTED", "SELECT * FROM users")
	if err == nil {
		t.Error("期望返回错误，但没有收到错误")
	}
}


