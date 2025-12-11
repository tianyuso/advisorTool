// Package db provides database-related utilities including affected rows calculation.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"advisorTool/parser/base"
	mysqlparser "advisorTool/parser/mysql"
	pgparser "advisorTool/parser/pg"
	plsqlparser "advisorTool/parser/plsql"
	tsqlparser "advisorTool/parser/tsql"
	"advisorTool/pkg/advisor"
)

// CalculateAffectedRows 计算 UPDATE/DELETE 语句的影响行数
// 通过将 UPDATE/DELETE 改写为 SELECT COUNT(1) 查询来估算
func CalculateAffectedRows(ctx context.Context, conn *sql.DB, statement string, engine advisor.Engine) (int, error) {
	// 首先判断是否是 UPDATE 或 DELETE 语句
	stmtType := getStatementType(statement, engine)
	if stmtType != "UPDATE" && stmtType != "DELETE" {
		return 0, nil
	}

	// 根据引擎类型改写 SQL
	countSQL, err := rewriteToCountSQL(statement, engine)
	if err != nil {
		return 0, errors.Wrap(err, "failed to rewrite SQL to count statement")
	}

	// 执行 COUNT 查询
	var count int
	err = conn.QueryRowContext(ctx, countSQL).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute count query")
	}

	return count, nil
}

// getStatementType 获取 SQL 语句类型
func getStatementType(statement string, engine advisor.Engine) string {
	trimmed := strings.TrimSpace(statement)
	upper := strings.ToUpper(trimmed)

	if strings.HasPrefix(upper, "UPDATE") {
		return "UPDATE"
	}
	if strings.HasPrefix(upper, "DELETE") {
		return "DELETE"
	}
	return "UNKNOWN"
}

// rewriteToCountSQL 将 UPDATE/DELETE 语句改写为 SELECT COUNT(1) 语句
func rewriteToCountSQL(statement string, engine advisor.Engine) (string, error) {
	switch engine {
	case advisor.EngineMySQL, advisor.EngineMariaDB, advisor.EngineTiDB, advisor.EngineOceanBase:
		return rewriteMySQLToCount(statement)
	case advisor.EnginePostgres:
		return rewritePostgresToCount(statement)
	case advisor.EngineMSSQL:
		return rewriteTSQLToCount(statement)
	case advisor.EngineOracle:
		return rewriteOracleToCount(statement)
	default:
		return "", fmt.Errorf("unsupported engine: %s", engine)
	}
}

// rewriteMySQLToCount 改写 MySQL UPDATE/DELETE 语句为 COUNT 查询
// MySQL 支持两种 UPDATE 语法：
// 1. 单表: UPDATE table SET ... WHERE ...
// 2. 连表: UPDATE t1 JOIN t2 ON ... SET ... WHERE ...
func rewriteMySQLToCount(statement string) (string, error) {
	// 解析 MySQL SQL
	res, err := mysqlparser.ParseMySQL(statement)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse MySQL statement")
	}

	if len(res) == 0 {
		return "", errors.New("no statements found")
	}

	// 使用第一个语句
	stmt := res[0]

	// 根据语句类型处理
	stmtType := strings.ToUpper(strings.TrimSpace(statement))

	if strings.HasPrefix(stmtType, "UPDATE") {
		return rewriteMySQLUpdateToCount(stmt, statement)
	} else if strings.HasPrefix(stmtType, "DELETE") {
		return rewriteMySQLDeleteToCount(stmt, statement)
	}

	return "", errors.New("not an UPDATE or DELETE statement")
}

// rewriteMySQLUpdateToCount 改写 MySQL UPDATE 为 COUNT
func rewriteMySQLUpdateToCount(stmt *base.ParseResult, original string) (string, error) {
	// 简单的基于文本的改写方法
	// 这是一个简化版本，处理常见情况

	upper := strings.ToUpper(original)

	// 查找关键字位置
	updateIdx := strings.Index(upper, "UPDATE")
	if updateIdx == -1 {
		return "", errors.New("UPDATE keyword not found")
	}

	setIdx := strings.Index(upper, "SET")
	if setIdx == -1 {
		return "", errors.New("SET keyword not found")
	}

	// 提取 UPDATE 和 SET 之间的表名和 JOIN 部分
	tableAndJoin := strings.TrimSpace(original[updateIdx+6 : setIdx])

	// 查找 WHERE 子句
	whereIdx := findWhereClauseIndex(upper, setIdx)
	var whereClause string
	if whereIdx > 0 {
		// 找到 WHERE 子句的结束位置（去除尾部的分号和空格）
		whereEnd := len(original)
		trimmed := strings.TrimRight(original[whereIdx:], "; \t\n\r")
		whereEnd = whereIdx + len(trimmed)
		whereClause = " " + strings.TrimSpace(original[whereIdx:whereEnd])
	}

	// 检查是否有 JOIN（连表更新）
	if containsJoin(tableAndJoin) {
		// 连表更新: UPDATE t1 JOIN t2 ON ... SET ... WHERE ...
		// 改写为: SELECT COUNT(1) FROM t1 JOIN t2 ON ... WHERE ...
		return fmt.Sprintf("SELECT COUNT(1) FROM %s%s", tableAndJoin, whereClause), nil
	} else {
		// 单表更新: UPDATE table SET ... WHERE ...
		// 改写为: SELECT COUNT(1) FROM table WHERE ...
		tableName := strings.TrimSpace(tableAndJoin)
		return fmt.Sprintf("SELECT COUNT(1) FROM %s%s", tableName, whereClause), nil
	}
}

// rewriteMySQLDeleteToCount 改写 MySQL DELETE 为 COUNT
func rewriteMySQLDeleteToCount(stmt *base.ParseResult, original string) (string, error) {
	upper := strings.ToUpper(original)

	// 查找关键字位置
	deleteIdx := strings.Index(upper, "DELETE")
	if deleteIdx == -1 {
		return "", errors.New("DELETE keyword not found")
	}

	fromIdx := strings.Index(upper, "FROM")
	if fromIdx == -1 {
		return "", errors.New("FROM keyword not found")
	}

	// 查找 WHERE 子句
	whereIdx := findWhereClauseIndex(upper, fromIdx)
	var restPart string
	if whereIdx > 0 {
		// 包括 WHERE 之后的所有内容（去除分号）
		trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
		restPart = strings.TrimSpace(trimmed)
	} else {
		// 没有 WHERE 子句，只有表名和可能的 JOIN
		trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
		restPart = strings.TrimSpace(trimmed)
	}

	// MySQL DELETE 语法：
	// 单表: DELETE FROM table WHERE ...
	// 连表: DELETE t1 FROM t1 JOIN t2 ON ... WHERE ...

	// 检查 DELETE 和 FROM 之间是否有表名（连表删除）
	betweenDeleteFrom := strings.TrimSpace(original[deleteIdx+6 : fromIdx])
	if betweenDeleteFrom != "" {
		// 连表删除
		return fmt.Sprintf("SELECT COUNT(1) FROM %s", restPart), nil
	} else {
		// 单表删除
		return fmt.Sprintf("SELECT COUNT(1) FROM %s", restPart), nil
	}
}

// rewritePostgresToCount 改写 PostgreSQL UPDATE/DELETE 语句为 COUNT 查询
func rewritePostgresToCount(statement string) (string, error) {
	// 解析 PostgreSQL SQL
	res, err := pgparser.ParsePostgreSQL(statement)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse PostgreSQL statement")
	}

	if len(res) == 0 {
		return "", errors.New("no statements found")
	}

	stmtType := strings.ToUpper(strings.TrimSpace(statement))

	if strings.HasPrefix(stmtType, "UPDATE") {
		return rewritePostgresUpdateToCount(statement)
	} else if strings.HasPrefix(stmtType, "DELETE") {
		return rewritePostgresDeleteToCount(statement)
	}

	return "", errors.New("not an UPDATE or DELETE statement")
}

// rewritePostgresUpdateToCount 改写 PostgreSQL UPDATE 为 COUNT
// PostgreSQL UPDATE 语法:
// 单表: UPDATE table SET ... WHERE ...
// 连表: UPDATE table1 SET ... FROM table2 WHERE ...
func rewritePostgresUpdateToCount(original string) (string, error) {
	upper := strings.ToUpper(original)

	updateIdx := strings.Index(upper, "UPDATE")
	if updateIdx == -1 {
		return "", errors.New("UPDATE keyword not found")
	}

	setIdx := strings.Index(upper, "SET")
	if setIdx == -1 {
		return "", errors.New("SET keyword not found")
	}

	// 提取表名
	tableName := strings.TrimSpace(original[updateIdx+6 : setIdx])

	// 查找 FROM 子句（连表更新）
	fromIdx := findKeywordIndex(upper, "FROM", setIdx)
	whereIdx := findWhereClauseIndex(upper, setIdx)

	if fromIdx > 0 && (whereIdx == -1 || fromIdx < whereIdx) {
		// 连表更新: UPDATE table1 SET ... FROM table2 WHERE ...
		// 改写为: SELECT COUNT(1) FROM table1 INNER JOIN table2 ON ...

		var fromPart string
		if whereIdx > 0 {
			fromPart = strings.TrimSpace(original[fromIdx+4 : whereIdx])
		} else {
			trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
			fromPart = strings.TrimSpace(trimmed)
		}

		// 提取 WHERE 子句中的 JOIN 条件
		var whereClause string
		if whereIdx > 0 {
			trimmed := strings.TrimRight(original[whereIdx:], "; \t\n\r")
			whereClause = strings.TrimSpace(trimmed)

			// PostgreSQL 的 WHERE 子句包含 JOIN 条件，需要转换为 ON 子句
			// 这是一个简化处理，实际可能需要更复杂的解析
			// WHERE table1.id = table2.id -> ON table1.id = table2.id
			whereClause = strings.Replace(whereClause, "WHERE", "ON", 1)
		}

		return fmt.Sprintf("SELECT COUNT(1) FROM %s INNER JOIN %s %s", tableName, fromPart, whereClause), nil
	} else {
		// 单表更新
		var whereClause string
		if whereIdx > 0 {
			trimmed := strings.TrimRight(original[whereIdx:], "; \t\n\r")
			whereClause = " " + strings.TrimSpace(trimmed)
		}
		return fmt.Sprintf("SELECT COUNT(1) FROM %s%s", tableName, whereClause), nil
	}
}

// rewritePostgresDeleteToCount 改写 PostgreSQL DELETE 为 COUNT
func rewritePostgresDeleteToCount(original string) (string, error) {
	upper := strings.ToUpper(original)

	deleteIdx := strings.Index(upper, "DELETE")
	if deleteIdx == -1 {
		return "", errors.New("DELETE keyword not found")
	}

	fromIdx := strings.Index(upper, "FROM")
	if fromIdx == -1 {
		return "", errors.New("FROM keyword not found")
	}

	// PostgreSQL DELETE 语法:
	// 单表: DELETE FROM table WHERE ...
	// 连表: DELETE FROM table1 USING table2 WHERE ...

	usingIdx := findKeywordIndex(upper, "USING", fromIdx)
	whereIdx := findWhereClauseIndex(upper, fromIdx)

	var tableName string
	if usingIdx > 0 && (whereIdx == -1 || usingIdx < whereIdx) {
		// 连表删除
		tableName = strings.TrimSpace(original[fromIdx+4 : usingIdx])

		var usingPart string
		if whereIdx > 0 {
			usingPart = strings.TrimSpace(original[usingIdx+5 : whereIdx])
		} else {
			trimmed := strings.TrimRight(original[usingIdx+5:], "; \t\n\r")
			usingPart = strings.TrimSpace(trimmed)
		}

		var whereClause string
		if whereIdx > 0 {
			trimmed := strings.TrimRight(original[whereIdx:], "; \t\n\r")
			whereClause = strings.TrimSpace(trimmed)
			// 转换 WHERE 为 ON
			whereClause = strings.Replace(whereClause, "WHERE", "ON", 1)
		}

		return fmt.Sprintf("SELECT COUNT(1) FROM %s INNER JOIN %s %s", tableName, usingPart, whereClause), nil
	} else {
		// 单表删除
		var restPart string
		if whereIdx > 0 {
			trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
			restPart = strings.TrimSpace(trimmed)
		} else {
			trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
			restPart = strings.TrimSpace(trimmed)
		}
		return fmt.Sprintf("SELECT COUNT(1) FROM %s", restPart), nil
	}
}

// rewriteTSQLToCount 改写 SQL Server UPDATE/DELETE 语句为 COUNT 查询
func rewriteTSQLToCount(statement string) (string, error) {
	// 解析 T-SQL
	res, err := tsqlparser.ParseTSQL(statement)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse T-SQL statement")
	}

	if len(res) == 0 {
		return "", errors.New("no statements found")
	}

	stmtType := strings.ToUpper(strings.TrimSpace(statement))

	if strings.HasPrefix(stmtType, "UPDATE") {
		return rewriteTSQLUpdateToCount(statement)
	} else if strings.HasPrefix(stmtType, "DELETE") {
		return rewriteTSQLDeleteToCount(statement)
	}

	return "", errors.New("not an UPDATE or DELETE statement")
}

// rewriteTSQLUpdateToCount 改写 SQL Server UPDATE 为 COUNT
// SQL Server UPDATE 语法:
// 单表: UPDATE table SET ... WHERE ...
// 连表: UPDATE t1 SET ... FROM table1 t1 INNER JOIN table2 t2 ON ... WHERE ...
func rewriteTSQLUpdateToCount(original string) (string, error) {
	upper := strings.ToUpper(original)

	updateIdx := strings.Index(upper, "UPDATE")
	if updateIdx == -1 {
		return "", errors.New("UPDATE keyword not found")
	}

	setIdx := strings.Index(upper, "SET")
	if setIdx == -1 {
		return "", errors.New("SET keyword not found")
	}

	// 提取 UPDATE 后的表别名或表名
	tableAlias := strings.TrimSpace(original[updateIdx+6 : setIdx])

	// 查找 FROM 子句
	fromIdx := findKeywordIndex(upper, "FROM", setIdx)
	whereIdx := findWhereClauseIndex(upper, setIdx)

	if fromIdx > 0 && (whereIdx == -1 || fromIdx < whereIdx) {
		// 连表更新: UPDATE t1 SET ... FROM table1 t1 INNER JOIN table2 t2 ON ... WHERE ...
		// 改写为: SELECT COUNT(1) FROM table1 t1 INNER JOIN table2 t2 ON ... WHERE ...

		var fromPart string
		if whereIdx > 0 {
			fromPart = strings.TrimSpace(original[fromIdx+4 : whereIdx])
		} else {
			trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
			fromPart = strings.TrimSpace(trimmed)
		}

		var whereClause string
		if whereIdx > 0 {
			trimmed := strings.TrimRight(original[whereIdx:], "; \t\n\r")
			whereClause = " " + strings.TrimSpace(trimmed)
		}

		return fmt.Sprintf("SELECT COUNT(1) FROM %s%s", fromPart, whereClause), nil
	} else {
		// 单表更新
		var whereClause string
		if whereIdx > 0 {
			trimmed := strings.TrimRight(original[whereIdx:], "; \t\n\r")
			whereClause = " " + strings.TrimSpace(trimmed)
		}
		return fmt.Sprintf("SELECT COUNT(1) FROM %s%s", tableAlias, whereClause), nil
	}
}

// rewriteTSQLDeleteToCount 改写 SQL Server DELETE 为 COUNT
func rewriteTSQLDeleteToCount(original string) (string, error) {
	upper := strings.ToUpper(original)

	deleteIdx := strings.Index(upper, "DELETE")
	if deleteIdx == -1 {
		return "", errors.New("DELETE keyword not found")
	}

	fromIdx := strings.Index(upper, "FROM")
	if fromIdx == -1 {
		return "", errors.New("FROM keyword not found")
	}

	// SQL Server DELETE 语法:
	// 单表: DELETE FROM table WHERE ...
	// 连表: DELETE t1 FROM table1 t1 INNER JOIN table2 t2 ON ... WHERE ...

	// 检查 DELETE 和 FROM 之间是否有表别名
	betweenDeleteFrom := strings.TrimSpace(original[deleteIdx+6 : fromIdx])

	whereIdx := findWhereClauseIndex(upper, fromIdx)

	if betweenDeleteFrom != "" {
		// 连表删除
		var fromPart string
		if whereIdx > 0 {
			fromPart = strings.TrimSpace(original[fromIdx+4 : whereIdx])
		} else {
			trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
			fromPart = strings.TrimSpace(trimmed)
		}

		var whereClause string
		if whereIdx > 0 {
			trimmed := strings.TrimRight(original[whereIdx:], "; \t\n\r")
			whereClause = " " + strings.TrimSpace(trimmed)
		}

		return fmt.Sprintf("SELECT COUNT(1) FROM %s%s", fromPart, whereClause), nil
	} else {
		// 单表删除
		var restPart string
		if whereIdx > 0 {
			trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
			restPart = strings.TrimSpace(trimmed)
		} else {
			trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
			restPart = strings.TrimSpace(trimmed)
		}
		return fmt.Sprintf("SELECT COUNT(1) FROM %s", restPart), nil
	}
}

// rewriteOracleToCount 改写 Oracle UPDATE/DELETE 语句为 COUNT 查询
func rewriteOracleToCount(statement string) (string, error) {
	// 解析 Oracle PL/SQL
	res, err := plsqlparser.ParsePLSQL(statement)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse PL/SQL statement")
	}

	if len(res) == 0 {
		return "", errors.New("no statements found")
	}

	stmtType := strings.ToUpper(strings.TrimSpace(statement))

	if strings.HasPrefix(stmtType, "UPDATE") {
		return rewriteOracleUpdateToCount(statement)
	} else if strings.HasPrefix(stmtType, "DELETE") {
		return rewriteOracleDeleteToCount(statement)
	}

	return "", errors.New("not an UPDATE or DELETE statement")
}

// rewriteOracleUpdateToCount 改写 Oracle UPDATE 为 COUNT
// Oracle UPDATE 语法与 PostgreSQL 类似
func rewriteOracleUpdateToCount(original string) (string, error) {
	upper := strings.ToUpper(original)

	updateIdx := strings.Index(upper, "UPDATE")
	if updateIdx == -1 {
		return "", errors.New("UPDATE keyword not found")
	}

	setIdx := strings.Index(upper, "SET")
	if setIdx == -1 {
		return "", errors.New("SET keyword not found")
	}

	// 提取表名
	tableName := strings.TrimSpace(original[updateIdx+6 : setIdx])

	// Oracle 支持子查询更新，这里处理简单情况
	whereIdx := findWhereClauseIndex(upper, setIdx)

	var whereClause string
	if whereIdx > 0 {
		trimmed := strings.TrimRight(original[whereIdx:], "; \t\n\r")
		whereClause = " " + strings.TrimSpace(trimmed)
	}

	return fmt.Sprintf("SELECT COUNT(1) FROM %s%s", tableName, whereClause), nil
}

// rewriteOracleDeleteToCount 改写 Oracle DELETE 为 COUNT
func rewriteOracleDeleteToCount(original string) (string, error) {
	upper := strings.ToUpper(original)

	deleteIdx := strings.Index(upper, "DELETE")
	if deleteIdx == -1 {
		return "", errors.New("DELETE keyword not found")
	}

	fromIdx := strings.Index(upper, "FROM")
	if fromIdx == -1 {
		// Oracle 支持 DELETE table WHERE ... 语法（不带 FROM）
		whereIdx := findWhereClauseIndex(upper, deleteIdx)

		var tableName string
		if whereIdx > 0 {
			tableName = strings.TrimSpace(original[deleteIdx+6 : whereIdx])
		} else {
			trimmed := strings.TrimRight(original[deleteIdx+6:], "; \t\n\r")
			tableName = strings.TrimSpace(trimmed)
		}

		var whereClause string
		if whereIdx > 0 {
			trimmed := strings.TrimRight(original[whereIdx:], "; \t\n\r")
			whereClause = " " + strings.TrimSpace(trimmed)
		}

		return fmt.Sprintf("SELECT COUNT(1) FROM %s%s", tableName, whereClause), nil
	}

	// 带 FROM 的语法
	whereIdx := findWhereClauseIndex(upper, fromIdx)

	var restPart string
	if whereIdx > 0 {
		trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
		restPart = strings.TrimSpace(trimmed)
	} else {
		trimmed := strings.TrimRight(original[fromIdx+4:], "; \t\n\r")
		restPart = strings.TrimSpace(trimmed)
	}

	return fmt.Sprintf("SELECT COUNT(1) FROM %s", restPart), nil
}

// 辅助函数

// findWhereClauseIndex 查找 WHERE 关键字的位置
func findWhereClauseIndex(upper string, startFrom int) int {
	return findKeywordIndex(upper, "WHERE", startFrom)
}

// findKeywordIndex 查找关键字的位置（必须是独立的单词）
func findKeywordIndex(text string, keyword string, startFrom int) int {
	searchText := text[startFrom:]
	idx := strings.Index(searchText, keyword)
	if idx == -1 {
		return -1
	}

	// 检查是否是独立单词
	absoluteIdx := startFrom + idx

	// 检查前面是否是空白字符或开头
	if absoluteIdx > 0 {
		prevChar := text[absoluteIdx-1]
		if !isWhitespace(prevChar) {
			// 继续查找下一个
			next := findKeywordIndex(text, keyword, absoluteIdx+1)
			return next
		}
	}

	// 检查后面是否是空白字符或结尾
	endIdx := absoluteIdx + len(keyword)
	if endIdx < len(text) {
		nextChar := text[endIdx]
		if !isWhitespace(nextChar) && nextChar != '(' {
			// 继续查找下一个
			next := findKeywordIndex(text, keyword, absoluteIdx+1)
			return next
		}
	}

	return absoluteIdx
}

// isWhitespace 检查字符是否是空白字符
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// containsJoin 检查字符串是否包含 JOIN 关键字
func containsJoin(text string) bool {
	upper := strings.ToUpper(text)
	joinKeywords := []string{"JOIN", "INNER JOIN", "LEFT JOIN", "RIGHT JOIN", "FULL JOIN", "CROSS JOIN"}
	for _, keyword := range joinKeywords {
		if strings.Contains(upper, keyword) {
			return true
		}
	}
	return false
}
