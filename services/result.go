package services

import (
	"context"
	"fmt"
	"strings"

	"advisorTool/db"
	"advisorTool/pkg/advisor"
)

// ReviewResult represents the review result in Inception-compatible format.
type ReviewResult struct {
	OrderID      int    `json:"order_id"`
	Stage        string `json:"stage"`
	ErrorLevel   string `json:"error_level"`
	StageStatus  string `json:"stage_status"`
	ErrorMessage string `json:"error_message"`
	SQL          string `json:"sql"`
	AffectedRows int    `json:"affected_rows"`
	Sequence     string `json:"sequence"`
	BackupDBName string `json:"backup_dbname"`
	ExecuteTime  string `json:"execute_time"`
	SQLSha1      string `json:"sqlsha1"`
	BackupTime   string `json:"backup_time"`
}

// DBConnectionParams holds database connection parameters.
type DBConnectionParams struct {
	Host        string
	Port        int
	User        string
	Password    string
	DbName      string
	Charset     string
	ServiceName string
	Sid         string
	SSLMode     string
	Timeout     int
	Schema      string
}

// CalculateAffectedRowsForStatements calculates affected rows for all SQL statements.
// Returns a map of SQL index to affected rows count.
func CalculateAffectedRowsForStatements(statement string, engineType advisor.Engine, dbParams *DBConnectionParams) map[int]int {
	affectedRowsMap := make(map[int]int)

	if dbParams == nil || dbParams.Host == "" || dbParams.Port == 0 {
		return affectedRowsMap
	}

	sqlStatements := SplitSQL(statement)

	// 打开数据库连接
	config := &db.ConnectionConfig{
		DbType:      GetDbTypeString(engineType),
		Host:        dbParams.Host,
		Port:        dbParams.Port,
		User:        dbParams.User,
		Password:    dbParams.Password,
		DbName:      dbParams.DbName,
		Charset:     dbParams.Charset,
		ServiceName: dbParams.ServiceName,
		Sid:         dbParams.Sid,
		SSLMode:     dbParams.SSLMode,
		Timeout:     dbParams.Timeout,
		Schema:      dbParams.Schema,
	}

	dbConn, err := db.OpenConnection(context.Background(), config)
	if err != nil {
		return affectedRowsMap
	}
	defer dbConn.Close()

	// 计算每个 SQL 语句的影响行数
	for i, sql := range sqlStatements {
		count, err := db.CalculateAffectedRows(context.Background(), dbConn, sql, engineType)
		if err == nil {
			affectedRowsMap[i] = count
		}
	}

	return affectedRowsMap
}

// ConvertToReviewResults converts advisor response to Inception-compatible format.
func ConvertToReviewResults(resp *advisor.ReviewResponse, statement string, engineType advisor.Engine, affectedRowsMap map[int]int) []ReviewResult {
	// Split SQL statements by semicolon
	sqlStatements := SplitSQL(statement)

	// If no issues found, return success for each statement
	if len(resp.Advices) == 0 {
		var results []ReviewResult
		for i, sql := range sqlStatements {
			affectedRows := 0
			if count, ok := affectedRowsMap[i]; ok {
				affectedRows = count
			}

			results = append(results, ReviewResult{
				OrderID:      i + 1,
				Stage:        "CHECKED",
				ErrorLevel:   "0",
				StageStatus:  "Audit Completed",
				ErrorMessage: "",
				SQL:          strings.TrimSpace(sql),
				AffectedRows: affectedRows,
				Sequence:     fmt.Sprintf("0_0_%08d", i),
				BackupDBName: "",
				ExecuteTime:  "0",
				SQLSha1:      "",
				BackupTime:   "0",
			})
		}
		return results
	}

	// Group advices by SQL statement (using line number)
	advicesBySQL := make(map[int][]*advisor.Advice)
	for _, advice := range resp.Advices {
		line := 1
		if advice.StartPosition != nil {
			line = int(advice.StartPosition.Line)
		}
		// Find which SQL statement this line belongs to
		sqlIndex := FindSQLIndexByLine(sqlStatements, statement, line)
		advicesBySQL[sqlIndex] = append(advicesBySQL[sqlIndex], advice)
	}

	var results []ReviewResult
	for i, sql := range sqlStatements {
		advices := advicesBySQL[i]

		errorLevel := "0"
		stageStatus := "Audit Completed"
		var errorMessages []string

		for _, advice := range advices {
			switch advice.Status {
			case advisor.AdviceStatusError:
				errorLevel = "2"
			case advisor.AdviceStatusWarning:
				if errorLevel != "2" {
					errorLevel = "1"
				}
			}
			// Collect error messages: [rule_type] message
			errorMessages = append(errorMessages, fmt.Sprintf("[%s] %s", advice.Title, advice.Content))
		}

		// 计算影响行数
		affectedRows := 0
		if count, ok := affectedRowsMap[i]; ok {
			affectedRows = count
		}

		results = append(results, ReviewResult{
			OrderID:      i + 1,
			Stage:        "CHECKED",
			ErrorLevel:   errorLevel,
			StageStatus:  stageStatus,
			ErrorMessage: strings.Join(errorMessages, "\n"),
			SQL:          strings.TrimSpace(sql),
			AffectedRows: affectedRows,
			Sequence:     fmt.Sprintf("0_0_%08d", i),
			BackupDBName: "",
			ExecuteTime:  "0",
			SQLSha1:      "",
			BackupTime:   "0",
		})
	}

	return results
}

// SplitSQL splits SQL statements by semicolon.
func SplitSQL(statement string) []string {
	// Simple split by semicolon - handles most cases
	parts := strings.Split(statement, ";")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		// If no semicolon found, treat entire statement as one SQL
		result = append(result, strings.TrimSpace(statement))
	}
	return result
}

// FindSQLIndexByLine finds which SQL statement a line belongs to.
func FindSQLIndexByLine(sqlStatements []string, fullStatement string, line int) int {
	if len(sqlStatements) <= 1 {
		return 0
	}

	// Count lines to find which statement the line belongs to
	lines := strings.Split(fullStatement, "\n")
	currentLine := 1
	currentSQLIndex := 0
	currentSQLStart := 0

	for i, l := range lines {
		if i > 0 {
			currentLine++
		}

		// Check if this line contains a semicolon (end of statement)
		if strings.Contains(l, ";") {
			if line >= currentSQLStart && line <= currentLine {
				return currentSQLIndex
			}
			currentSQLIndex++
			currentSQLStart = currentLine + 1
			if currentSQLIndex >= len(sqlStatements) {
				currentSQLIndex = len(sqlStatements) - 1
			}
		}
	}

	// If not found, return the last index or 0
	if line >= currentSQLStart {
		return currentSQLIndex
	}
	return 0
}

// GetDbTypeString converts Engine type to database type string.
func GetDbTypeString(engineType advisor.Engine) string {
	switch engineType {
	case advisor.EngineMySQL:
		return "mysql"
	case advisor.EngineMariaDB:
		return "mariadb"
	case advisor.EngineTiDB:
		return "tidb"
	case advisor.EngineOceanBase:
		return "oceanbase"
	case advisor.EnginePostgres:
		return "postgres"
	case advisor.EngineMSSQL:
		return "mssql"
	case advisor.EngineOracle:
		return "oracle"
	case advisor.EngineSnowflake:
		return "snowflake"
	default:
		return "mysql"
	}
}
