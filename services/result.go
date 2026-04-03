package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/tianyuso/advisorTool/db"
	parserbase "github.com/tianyuso/advisorTool/parser/base"
	"github.com/tianyuso/advisorTool/pkg/advisor"
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

// AffectedRowsInfo holds the count and error information for affected rows calculation.
type AffectedRowsInfo struct {
	Count int
	Error string
}

type splitStatement struct {
	Text      string
	StartLine int
	EndLine   int
}

// CalculateAffectedRowsForStatements calculates affected rows for all SQL statements.
// Returns a map of SQL index to AffectedRowsInfo (count and error).
func CalculateAffectedRowsForStatements(statement string, engineType advisor.Engine, dbParams *DBConnectionParams) map[int]*AffectedRowsInfo {
	affectedRowsMap := make(map[int]*AffectedRowsInfo)

	if dbParams == nil || dbParams.Host == "" || dbParams.Port == 0 {
		return affectedRowsMap
	}

	sqlStatements := splitSQLWithMetadata(statement, engineType)

	// 打开数据库连接
	config := &db.ConnectionConfig{
		DbType:        GetDbTypeString(engineType),
		Host:          dbParams.Host,
		Port:          dbParams.Port,
		User:          dbParams.User,
		Password:      dbParams.Password,
		DbName:        dbParams.DbName,
		Charset:       dbParams.Charset,
		ServiceName:   dbParams.ServiceName,
		Sid:           dbParams.Sid,
		SSLMode:       dbParams.SSLMode,
		Timeout:       dbParams.Timeout,
		Schema:        dbParams.Schema,
		SetSearchPath: true, // 计算影响行数时设置 search_path
	}

	dbConn, err := db.OpenConnection(context.Background(), config)
	if err != nil {
		return affectedRowsMap
	}
	defer dbConn.Close()

	// 计算每个 SQL 语句的影响行数
	for i, sql := range sqlStatements {
		count, err := db.CalculateAffectedRows(context.Background(), dbConn, sql.Text, engineType)
		info := &AffectedRowsInfo{
			Count: count,
		}
		if err != nil {
			info.Error = err.Error()
		}
		affectedRowsMap[i] = info
	}

	return affectedRowsMap
}

// ConvertToReviewResults converts advisor response to Inception-compatible format.
func ConvertToReviewResults(resp *advisor.ReviewResponse, statement string, engineType advisor.Engine, affectedRowsMap map[int]*AffectedRowsInfo) []ReviewResult {
	sqlStatements := splitSQLWithMetadata(statement, engineType)

	// If no issues found, return success for each statement
	if len(resp.Advices) == 0 {
		var results []ReviewResult
		for i, sql := range sqlStatements {
			affectedRows := 0
			errorMessage := ""
			errorLevel := "0"

			if info, ok := affectedRowsMap[i]; ok {
				affectedRows = info.Count
				if info.Error != "" {
					errorMessage = fmt.Sprintf("[AffectedRows] %s", info.Error)
					errorLevel = "2"
				}
			}

			results = append(results, ReviewResult{
				OrderID:      i + 1,
				Stage:        "CHECKED",
				ErrorLevel:   errorLevel,
				StageStatus:  "Audit Completed",
				ErrorMessage: errorMessage,
				SQL:          sql.Text,
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
		sqlIndex := findSQLIndexByLine(sqlStatements, line)
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

		// 计算影响行数和错误信息
		affectedRows := 0
		if info, ok := affectedRowsMap[i]; ok {
			affectedRows = info.Count
			if info.Error != "" {
				errorMessages = append(errorMessages, fmt.Sprintf("[AffectedRows] %s", info.Error))
				errorLevel = "2"
			}
		}

		results = append(results, ReviewResult{
			OrderID:      i + 1,
			Stage:        "CHECKED",
			ErrorLevel:   errorLevel,
			StageStatus:  stageStatus,
			ErrorMessage: strings.Join(errorMessages, "\n"),
			SQL:          sql.Text,
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

// splitSQLWithMetadata splits SQL by engine parser and keeps line range metadata.
func splitSQLWithMetadata(statement string, engineType advisor.Engine) []splitStatement {
	sqlList, err := parserbase.SplitMultiSQL(engineType, statement)
	if err != nil {
		return splitSQLBySemicolon(statement)
	}

	sqlList = parserbase.FilterEmptySQL(sqlList)
	var result []splitStatement
	for _, sql := range sqlList {
		text := strings.TrimSpace(sql.Text)
		if text == "" {
			continue
		}

		startLine := sql.BaseLine + 1
		endLine := startLine + strings.Count(sql.Text, "\n")
		if sql.Start != nil && sql.Start.Line > 0 {
			startLine = int(sql.Start.Line)
		}
		if sql.End != nil && sql.End.Line > 0 {
			endLine = int(sql.End.Line)
		}
		if endLine < startLine {
			endLine = startLine
		}

		result = append(result, splitStatement{
			Text:      text,
			StartLine: startLine,
			EndLine:   endLine,
		})
	}

	if len(result) == 0 {
		return splitSQLBySemicolon(statement)
	}
	return result
}

func splitSQLBySemicolon(statement string) []splitStatement {
	parts := SplitSQL(statement)
	result := make([]splitStatement, 0, len(parts))
	for _, part := range parts {
		result = append(result, splitStatement{
			Text:      part,
			StartLine: 1,
			EndLine:   1,
		})
	}
	return result
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

// findSQLIndexByLine finds which SQL statement a line belongs to.
func findSQLIndexByLine(sqlStatements []splitStatement, line int) int {
	if len(sqlStatements) <= 1 {
		return 0
	}

	for i, sql := range sqlStatements {
		if line >= sql.StartLine && line <= sql.EndLine {
			return i
		}
	}

	// If not found, return nearest range.
	if line < sqlStatements[0].StartLine {
		return 0
	}
	return len(sqlStatements) - 1
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
