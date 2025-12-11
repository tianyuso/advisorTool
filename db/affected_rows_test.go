package db

import (
	"testing"

	"advisorTool/pkg/advisor"
)

// TestRewriteMySQLToCount 测试 MySQL UPDATE/DELETE 改写为 COUNT
func TestRewriteMySQLToCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "MySQL单表UPDATE带WHERE",
			input:    "UPDATE users SET name = 'test' WHERE id > 100",
			expected: "SELECT COUNT(1) FROM users WHERE id > 100",
			wantErr:  false,
		},
		{
			name:     "MySQL单表UPDATE不带WHERE",
			input:    "UPDATE users SET name = 'test'",
			expected: "SELECT COUNT(1) FROM users",
			wantErr:  false,
		},
		{
			name:     "MySQL连表UPDATE",
			input:    "UPDATE t1 INNER JOIN t2 ON t1.id = t2.id SET t1.name = t2.name WHERE t1.status = 1",
			expected: "SELECT COUNT(1) FROM t1 INNER JOIN t2 ON t1.id = t2.id WHERE t1.status = 1",
			wantErr:  false,
		},
		{
			name:     "MySQL连表UPDATE不带WHERE",
			input:    "UPDATE t1 LEFT JOIN t2 ON t1.id = t2.id SET t1.name = t2.name",
			expected: "SELECT COUNT(1) FROM t1 LEFT JOIN t2 ON t1.id = t2.id",
			wantErr:  false,
		},
		{
			name:     "MySQL单表DELETE带WHERE",
			input:    "DELETE FROM users WHERE id > 100",
			expected: "SELECT COUNT(1) FROM users WHERE id > 100",
			wantErr:  false,
		},
		{
			name:     "MySQL单表DELETE不带WHERE",
			input:    "DELETE FROM users",
			expected: "SELECT COUNT(1) FROM users",
			wantErr:  false,
		},
		{
			name:     "MySQL连表DELETE",
			input:    "DELETE t1 FROM t1 INNER JOIN t2 ON t1.id = t2.id WHERE t1.status = 1",
			expected: "SELECT COUNT(1) FROM t1 INNER JOIN t2 ON t1.id = t2.id WHERE t1.status = 1",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rewriteMySQLToCount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("rewriteMySQLToCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.expected {
				t.Errorf("rewriteMySQLToCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRewritePostgresToCount 测试 PostgreSQL UPDATE/DELETE 改写为 COUNT
func TestRewritePostgresToCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "PostgreSQL单表UPDATE带WHERE",
			input:    "UPDATE users SET name = 'test' WHERE id > 100",
			expected: "SELECT COUNT(1) FROM users WHERE id > 100",
			wantErr:  false,
		},
		{
			name:     "PostgreSQL单表UPDATE不带WHERE",
			input:    "UPDATE users SET name = 'test'",
			expected: "SELECT COUNT(1) FROM users",
			wantErr:  false,
		},
		{
			name:     "PostgreSQL连表UPDATE",
			input:    "UPDATE table1 SET column1 = table2.column1 FROM table2 WHERE table1.id = table2.id",
			expected: "SELECT COUNT(1) FROM table1 INNER JOIN table2 ON table1.id = table2.id",
			wantErr:  false,
		},
		{
			name:     "PostgreSQL单表DELETE带WHERE",
			input:    "DELETE FROM users WHERE id > 100",
			expected: "SELECT COUNT(1) FROM users WHERE id > 100",
			wantErr:  false,
		},
		{
			name:     "PostgreSQL单表DELETE不带WHERE",
			input:    "DELETE FROM users",
			expected: "SELECT COUNT(1) FROM users",
			wantErr:  false,
		},
		{
			name:     "PostgreSQL连表DELETE",
			input:    "DELETE FROM table1 USING table2 WHERE table1.id = table2.id",
			expected: "SELECT COUNT(1) FROM table1 INNER JOIN table2 ON table1.id = table2.id",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rewritePostgresToCount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("rewritePostgresToCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.expected {
				t.Errorf("rewritePostgresToCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRewriteTSQLToCount 测试 SQL Server UPDATE/DELETE 改写为 COUNT
func TestRewriteTSQLToCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "SQL Server单表UPDATE带WHERE",
			input:    "UPDATE users SET name = 'test' WHERE id > 100",
			expected: "SELECT COUNT(1) FROM users WHERE id > 100",
			wantErr:  false,
		},
		{
			name:     "SQL Server单表UPDATE不带WHERE",
			input:    "UPDATE users SET name = 'test'",
			expected: "SELECT COUNT(1) FROM users",
			wantErr:  false,
		},
		{
			name:     "SQL Server连表UPDATE",
			input:    "UPDATE t1 SET t1.column1 = t2.column1 FROM table1 t1 INNER JOIN table2 t2 ON t1.id = t2.id WHERE t1.status = 1",
			expected: "SELECT COUNT(1) FROM table1 t1 INNER JOIN table2 t2 ON t1.id = t2.id WHERE t1.status = 1",
			wantErr:  false,
		},
		{
			name:     "SQL Server单表DELETE带WHERE",
			input:    "DELETE FROM users WHERE id > 100",
			expected: "SELECT COUNT(1) FROM users WHERE id > 100",
			wantErr:  false,
		},
		{
			name:     "SQL Server单表DELETE不带WHERE",
			input:    "DELETE FROM users",
			expected: "SELECT COUNT(1) FROM users",
			wantErr:  false,
		},
		{
			name:     "SQL Server连表DELETE",
			input:    "DELETE t1 FROM table1 t1 INNER JOIN table2 t2 ON t1.id = t2.id WHERE t1.status = 1",
			expected: "SELECT COUNT(1) FROM table1 t1 INNER JOIN table2 t2 ON t1.id = t2.id WHERE t1.status = 1",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rewriteTSQLToCount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("rewriteTSQLToCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.expected {
				t.Errorf("rewriteTSQLToCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRewriteOracleToCount 测试 Oracle UPDATE/DELETE 改写为 COUNT
func TestRewriteOracleToCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Oracle单表UPDATE带WHERE",
			input:    "UPDATE users SET name = 'test' WHERE id > 100",
			expected: "SELECT COUNT(1) FROM users WHERE id > 100",
			wantErr:  false,
		},
		{
			name:     "Oracle单表UPDATE不带WHERE",
			input:    "UPDATE users SET name = 'test'",
			expected: "SELECT COUNT(1) FROM users",
			wantErr:  false,
		},
		{
			name:     "Oracle单表DELETE带WHERE",
			input:    "DELETE FROM users WHERE id > 100",
			expected: "SELECT COUNT(1) FROM users WHERE id > 100",
			wantErr:  false,
		},
		{
			name:     "Oracle单表DELETE不带WHERE",
			input:    "DELETE FROM users",
			expected: "SELECT COUNT(1) FROM users",
			wantErr:  false,
		},
		{
			name:     "Oracle DELETE不带FROM",
			input:    "DELETE users WHERE id > 100",
			expected: "SELECT COUNT(1) FROM users WHERE id > 100",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rewriteOracleToCount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("rewriteOracleToCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.expected {
				t.Errorf("rewriteOracleToCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRewriteToCountSQL 测试总入口函数
func TestRewriteToCountSQL(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		engine   advisor.Engine
		expected string
		wantErr  bool
	}{
		{
			name:     "MySQL UPDATE",
			stmt:     "UPDATE users SET name = 'test' WHERE id = 1",
			engine:   advisor.EngineMySQL,
			expected: "SELECT COUNT(1) FROM users WHERE id = 1",
			wantErr:  false,
		},
		{
			name:     "PostgreSQL DELETE",
			stmt:     "DELETE FROM users WHERE id = 1",
			engine:   advisor.EnginePostgres,
			expected: "SELECT COUNT(1) FROM users WHERE id = 1",
			wantErr:  false,
		},
		{
			name:     "SQL Server UPDATE",
			stmt:     "UPDATE users SET name = 'test' WHERE id = 1",
			engine:   advisor.EngineMSSQL,
			expected: "SELECT COUNT(1) FROM users WHERE id = 1",
			wantErr:  false,
		},
		{
			name:     "Oracle DELETE",
			stmt:     "DELETE FROM users WHERE id = 1",
			engine:   advisor.EngineOracle,
			expected: "SELECT COUNT(1) FROM users WHERE id = 1",
			wantErr:  false,
		},
		{
			name:    "不支持的引擎",
			stmt:    "UPDATE users SET name = 'test'",
			engine:  advisor.EngineSnowflake,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rewriteToCountSQL(tt.stmt, tt.engine)
			if (err != nil) != tt.wantErr {
				t.Errorf("rewriteToCountSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got != tt.expected {
				t.Errorf("rewriteToCountSQL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGetStatementType 测试语句类型识别
func TestGetStatementType(t *testing.T) {
	tests := []struct {
		name   string
		stmt   string
		engine advisor.Engine
		want   string
	}{
		{
			name:   "UPDATE语句",
			stmt:   "UPDATE users SET name = 'test'",
			engine: advisor.EngineMySQL,
			want:   "UPDATE",
		},
		{
			name:   "DELETE语句",
			stmt:   "DELETE FROM users WHERE id = 1",
			engine: advisor.EngineMySQL,
			want:   "DELETE",
		},
		{
			name:   "SELECT语句",
			stmt:   "SELECT * FROM users",
			engine: advisor.EngineMySQL,
			want:   "UNKNOWN",
		},
		{
			name:   "INSERT语句",
			stmt:   "INSERT INTO users (name) VALUES ('test')",
			engine: advisor.EngineMySQL,
			want:   "UNKNOWN",
		},
		{
			name:   "空语句",
			stmt:   "",
			engine: advisor.EngineMySQL,
			want:   "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStatementType(tt.stmt, tt.engine); got != tt.want {
				t.Errorf("getStatementType() = %v, want %v", got, tt.want)
			}
		})
	}
}
