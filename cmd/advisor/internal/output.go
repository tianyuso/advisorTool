package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"advisorTool/pkg/advisor"
)

// OutputResults outputs the review results in the specified format.
func OutputResults(resp *advisor.ReviewResponse, statement string, engineType advisor.Engine, format string, dbParams *DBConnectionParams) error {
	// 先计算所有 SQL 语句的影响行数（适用于所有格式）
	affectedRowsMap := CalculateAffectedRowsForStatements(statement, engineType, dbParams)

	// 转换为统一的结果格式
	results := ConvertToReviewResults(resp, statement, engineType, affectedRowsMap)

	switch format {
	case "json":
		data, err := json.Marshal(results)
		if err != nil {
			return err
		}
		fmt.Println(string(data))

	case "table":
		renderTable(results)

	default:
		return fmt.Errorf("unsupported format: %s (supported: json, table)", format)
	}

	return nil
}

// renderTable renders results as a formatted table using go-pretty
func renderTable(results []ReviewResult) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// 设置表头（包含所有字段）
	t.AppendHeader(table.Row{
		"Order",
		"Stage",
		"Level",
		"Status",
		"SQL",
		"Affected",
		"Sequence",
		"Backup DB",
		"Exec Time",
		"SQL SHA1",
		"Backup Time",
		"Error Message",
	})

	// 设置列配置（包含所有字段）
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: 6, WidthMax: 6, AlignHeader: text.AlignCenter, Align: text.AlignCenter},    // Order
		{Number: 2, WidthMin: 8, WidthMax: 8, AlignHeader: text.AlignCenter, Align: text.AlignCenter},    // Stage
		{Number: 3, WidthMin: 8, WidthMax: 10, AlignHeader: text.AlignCenter, Align: text.AlignCenter},   // Level
		{Number: 4, WidthMin: 14, WidthMax: 18, AlignHeader: text.AlignCenter, Align: text.AlignLeft},    // Status
		{Number: 5, WidthMin: 30, WidthMax: 60, AlignHeader: text.AlignCenter, Align: text.AlignLeft},    // SQL
		{Number: 6, WidthMin: 9, WidthMax: 9, AlignHeader: text.AlignCenter, Align: text.AlignCenter},    // Affected
		{Number: 7, WidthMin: 12, WidthMax: 15, AlignHeader: text.AlignCenter, Align: text.AlignCenter},  // Sequence
		{Number: 8, WidthMin: 10, WidthMax: 15, AlignHeader: text.AlignCenter, Align: text.AlignCenter},  // Backup DB
		{Number: 9, WidthMin: 10, WidthMax: 10, AlignHeader: text.AlignCenter, Align: text.AlignCenter},  // Exec Time
		{Number: 10, WidthMin: 12, WidthMax: 20, AlignHeader: text.AlignCenter, Align: text.AlignLeft},   // SQL SHA1
		{Number: 11, WidthMin: 12, WidthMax: 12, AlignHeader: text.AlignCenter, Align: text.AlignCenter}, // Backup Time
		{Number: 12, WidthMin: 20, WidthMax: 50, AlignHeader: text.AlignCenter, Align: text.AlignLeft},   // Error Message
	})

	// 添加数据行（包含所有字段）
	for _, result := range results {
		// 根据错误级别设置颜色
		levelText := getLevelText(result.ErrorLevel)
		statusColor := getStatusColor(result.ErrorLevel)

		// 格式化 SQL（保留适当长度）
		sql := formatSQL(result.SQL, 60)

		// 格式化错误消息
		errorMsg := formatErrorMessage(result.ErrorMessage, 50)

		// 格式化 SHA1（显示前 20 个字符）
		sqlSha1 := formatString(result.SQLSha1, 20)

		t.AppendRow(table.Row{
			result.OrderID,
			result.Stage,
			levelText,
			applyColor(result.StageStatus, statusColor),
			sql,
			result.AffectedRows,
			result.Sequence,
			formatString(result.BackupDBName, 15),
			result.ExecuteTime,
			sqlSha1,
			result.BackupTime,
			errorMsg,
		})
	}

	// 设置样式
	t.SetStyle(table.StyleColoredBright)
	t.Style().Options.SeparateRows = false
	t.Style().Options.DrawBorder = true

	// 渲染表格
	t.Render()

	// 输出汇总信息
	fmt.Println()
	printSummary(results)
}

// printSummary prints a summary of the results
func printSummary(results []ReviewResult) {
	totalCount := len(results)
	errorCount := 0
	warningCount := 0
	successCount := 0
	totalAffectedRows := 0

	for _, result := range results {
		switch result.ErrorLevel {
		case "2":
			errorCount++
		case "1":
			warningCount++
		case "0":
			successCount++
		}
		totalAffectedRows += result.AffectedRows
	}

	fmt.Printf("Summary:\n")
	fmt.Printf("  Total Statements: %d\n", totalCount)
	fmt.Printf("  ✓ Success: %d\n", successCount)
	if warningCount > 0 {
		fmt.Printf("  ⚠ Warnings: %d\n", warningCount)
	}
	if errorCount > 0 {
		fmt.Printf("  ✗ Errors: %d\n", errorCount)
	}
	if totalAffectedRows > 0 {
		fmt.Printf("  Total Affected Rows: %d\n", totalAffectedRows)
	}
	fmt.Println()
}

// formatSQL formats SQL string for display
func formatSQL(sql string, maxLen int) string {
	// 移除多余的空白字符
	sql = removeExtraSpaces(sql)

	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen-3] + "..."
}

// formatErrorMessage formats error message for display
func formatErrorMessage(msg string, maxLen int) string {
	if msg == "" {
		return "-"
	}

	// 移除换行符，只显示单行
	msg = removeNewlines(msg)

	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen-3] + "..."
}

// removeExtraSpaces removes extra spaces from string
func removeExtraSpaces(s string) string {
	// 简单实现：替换多个空格为一个
	result := ""
	lastSpace := false
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if !lastSpace {
				result += " "
				lastSpace = true
			}
		} else {
			result += string(c)
			lastSpace = false
		}
	}
	return result
}

// removeNewlines removes newlines from string
func removeNewlines(s string) string {
	result := ""
	for _, c := range s {
		if c != '\n' && c != '\r' {
			result += string(c)
		} else {
			result += " "
		}
	}
	return result
}

// formatString formats a string with truncation
func formatString(s string, maxLen int) string {
	if s == "" {
		return "-"
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getLevelText converts error level to readable text
func getLevelText(errorLevel string) string {
	switch errorLevel {
	case "0":
		return "✓ OK"
	case "1":
		return "⚠ WARN"
	case "2":
		return "✗ ERROR"
	default:
		return errorLevel
	}
}

// getStatusColor returns color code based on error level
func getStatusColor(errorLevel string) string {
	switch errorLevel {
	case "0":
		return "\033[32m" // Green
	case "1":
		return "\033[33m" // Yellow
	case "2":
		return "\033[31m" // Red
	default:
		return "\033[0m" // Reset
	}
}

// applyColor applies color to text
func applyColor(text, color string) string {
	if color == "" {
		return text
	}
	return color + text + "\033[0m"
}

// ListAvailableRules lists all available SQL review rules.
func ListAvailableRules() {
	fmt.Println("Available SQL Review Rules:")
	fmt.Println("===========================")
	fmt.Println()

	rules := advisor.AllRules()
	for _, ruleType := range rules {
		desc := advisor.GetRuleDescription(ruleType)
		fmt.Printf("  %s\n    %s\n\n", ruleType, desc)
	}

	fmt.Printf("\nTotal: %d rules\n", len(rules))
}
