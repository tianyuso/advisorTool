package extractobject

import (
	"context"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/bytebase/parser/snowflake"

	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
)

// SnowflakeExtractor Snowflake表名提取器
type SnowflakeExtractor struct{}

// Extract 从Snowflake SQL语句中提取表名
func (e *SnowflakeExtractor) Extract(ctx context.Context, sql string) ([]TableInfo, error) {
	asts, err := parseSQL(storepb.Engine_SNOWFLAKE, sql)
	if err != nil {
		return nil, err
	}

	var tables []TableInfo
	for _, ast := range asts {
		antlrAST, ok := ast.(*base.ANTLRAST)
		if !ok {
			continue
		}

		listener := &snowflakeTableExtractListener{
			tables:   make([]TableInfo, 0),
			tableMap: make(map[string]bool),
		}

		antlr.ParseTreeWalkerDefault.Walk(listener, antlrAST.Tree)
		tables = append(tables, listener.tables...)
	}

	return tables, nil
}

// snowflakeTableExtractListener Snowflake表名提取监听器
type snowflakeTableExtractListener struct {
	*parser.BaseSnowflakeParserListener
	tables   []TableInfo
	tableMap map[string]bool // 用于去重
}

// EnterObject_name 进入对象名称节点
func (l *snowflakeTableExtractListener) EnterObject_name(ctx *parser.Object_nameContext) {
	tableInfo := extractSnowflakeTableInfo(ctx)
	if tableInfo.TBName != "" {
		// 去重
		key := tableInfo.DBName + "." + tableInfo.Schema + "." + tableInfo.TBName
		if !l.tableMap[key] {
			l.tables = append(l.tables, tableInfo)
			l.tableMap[key] = true
		}
	}
}

// extractSnowflakeTableInfo 从对象名称上下文中提取表信息
func extractSnowflakeTableInfo(ctx *parser.Object_nameContext) TableInfo {
	tableInfo := TableInfo{}

	if ctx == nil {
		return tableInfo
	}

	// Snowflake支持格式: database.schema.table 或 schema.table 或 table
	fullName := ctx.GetText()

	// 移除引号和空白
	fullName = strings.Trim(fullName, "\"' \t\n")

	// 分割名称部分
	parts := strings.Split(fullName, ".")

	// 清理每个部分
	for i := range parts {
		parts[i] = strings.Trim(parts[i], "\"' \t\n")
	}

	switch len(parts) {
	case 1:
		// 只有表名
		tableInfo.TBName = parts[0]
	case 2:
		// schema.table
		tableInfo.Schema = parts[0]
		tableInfo.TBName = parts[1]
	case 3:
		// database.schema.table
		tableInfo.DBName = parts[0]
		tableInfo.Schema = parts[1]
		tableInfo.TBName = parts[2]
	}

	return tableInfo
}
