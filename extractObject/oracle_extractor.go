package extractobject

import (
	"context"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/bytebase/parser/plsql"

	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
	plsqlparser "github.com/tianyuso/advisorTool/parser/plsql"
	_ "github.com/tianyuso/advisorTool/parser/plsql" // 导入以注册parser
)

// OracleExtractor Oracle表名提取器
type OracleExtractor struct{}

// Extract 从Oracle SQL语句中提取表名
func (e *OracleExtractor) Extract(ctx context.Context, sql string) ([]TableInfo, error) {
	asts, err := parseSQL(storepb.Engine_ORACLE, sql)
	if err != nil {
		return nil, err
	}

	var tables []TableInfo
	for _, ast := range asts {
		antlrAST, ok := ast.(*base.ANTLRAST)
		if !ok {
			continue
		}

		listener := &oracleTableExtractListener{
			tables:   make([]TableInfo, 0),
			tableMap: make(map[string]bool),
			cteNames: make(map[string]bool),
		}

		antlr.ParseTreeWalkerDefault.Walk(listener, antlrAST.Tree)
		tables = append(tables, listener.tables...)
	}

	return tables, nil
}

// oracleTableExtractListener Oracle表名提取监听器
type oracleTableExtractListener struct {
	*parser.BasePlSqlParserListener
	tables       []TableInfo
	tableMap     map[string]bool   // 用于去重
	cteNames     map[string]bool   // CTE名称集合
	pendingAlias string            // 待处理的别名
}

// EnterTableview_name 进入表/视图名称节点
func (l *oracleTableExtractListener) EnterTableview_name(ctx *parser.Tableview_nameContext) {
	tableInfo := extractOracleTableInfo(ctx)
	if tableInfo.TBName == "" {
		return
	}

	// 检查是否是CTE
	if l.cteNames[tableInfo.TBName] {
		tableInfo.IsCTE = true
	}

	// 不再使用简单的去重，允许同一张表的多次引用（可能有不同的别名）
	l.tables = append(l.tables, tableInfo)
}

// ExitTable_alias 退出表别名节点（处理别名）
func (l *oracleTableExtractListener) ExitTable_alias(ctx *parser.Table_aliasContext) {
	alias := plsqlparser.NormalizeTableAlias(ctx)
	if alias == "" {
		return
	}

	// 规范化为大写
	alias = strings.ToUpper(alias)

	// 更新最后一个表的别名（假设Table_alias紧跟在表引用之后）
	if len(l.tables) > 0 {
		lastIdx := len(l.tables) - 1
		if l.tables[lastIdx].Alias == "" {
			l.tables[lastIdx].Alias = alias
		}
	}
}

// EnterFactoring_element 进入CTE定义节点（Oracle称为Factoring Element）
func (l *oracleTableExtractListener) EnterFactoring_element(ctx *parser.Factoring_elementContext) {
	if ctx.Query_name() == nil || ctx.Query_name().Identifier() == nil {
		return
	}
	
	// 获取CTE名称
	cteName := normalizeOracleIdentifier(ctx.Query_name().Identifier())
	l.cteNames[cteName] = true
}


// normalizeOracleIdentifier 规范化Oracle标识符
func normalizeOracleIdentifier(ctx parser.IIdentifierContext) string {
	if ctx == nil {
		return ""
	}
	text := ctx.GetText()
	// 移除引号
	text = strings.Trim(text, "\"")
	return strings.ToUpper(text) // Oracle默认不区分大小写，转大写
}

// extractOracleTableInfo 从表视图名称上下文中提取表信息
func extractOracleTableInfo(ctx *parser.Tableview_nameContext) TableInfo {
	tableInfo := TableInfo{}

	if ctx == nil {
		return tableInfo
	}

	// 获取完整名称文本
	fullName := ctx.GetText()
	
	// 移除引号和空白
	fullName = strings.Trim(fullName, "\"' \t\n")
	
	// 分割 schema.table 或 database.schema.table
	parts := strings.Split(fullName, ".")
	
	// 清理每个部分并转为大写（Oracle默认不区分大小写）
	for i := range parts {
		parts[i] = strings.Trim(parts[i], "\"' \t\n")
		parts[i] = strings.ToUpper(parts[i])
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
		// database.schema.table (Oracle中不常见，但支持)
		tableInfo.DBName = parts[0]
		tableInfo.Schema = parts[1]
		tableInfo.TBName = parts[2]
	}

	return tableInfo
}


