package extractobject

import (
	"context"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/bytebase/parser/tsql"

	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
	tsqlparser "github.com/tianyuso/advisorTool/parser/tsql"
	_ "github.com/tianyuso/advisorTool/parser/tsql" // 导入以注册parser
)

// SQLServerExtractor SQL Server表名提取器
type SQLServerExtractor struct{}

// Extract 从SQL Server SQL语句中提取表名
func (e *SQLServerExtractor) Extract(ctx context.Context, sql string) ([]TableInfo, error) {
	asts, err := parseSQL(storepb.Engine_MSSQL, sql)
	if err != nil {
		return nil, err
	}

	var tables []TableInfo
	for _, ast := range asts {
		antlrAST, ok := ast.(*base.ANTLRAST)
		if !ok {
			continue
		}

		listener := &sqlserverTableExtractListener{
			tables:   make([]TableInfo, 0),
			tableMap: make(map[string]bool),
			cteNames: make(map[string]bool),
		}

		antlr.ParseTreeWalkerDefault.Walk(listener, antlrAST.Tree)
		tables = append(tables, listener.tables...)
	}

	return tables, nil
}

// sqlserverTableExtractListener SQL Server表名提取监听器
type sqlserverTableExtractListener struct {
	*parser.BaseTSqlParserListener
	tables   []TableInfo
	tableMap map[string]bool   // 用于去重
	cteNames map[string]bool   // CTE名称集合
}

// EnterTable_source_item 进入表源项节点（只在FROM子句中触发）
func (l *sqlserverTableExtractListener) EnterTable_source_item(ctx *parser.Table_source_itemContext) {
	// 只处理包含Full_table_name的情况
	if ctx.Full_table_name() == nil {
		return
	}

	tableName, err := tsqlparser.NormalizeFullTableName(ctx.Full_table_name())
	if err != nil || tableName == nil {
		return
	}

	tableInfo := TableInfo{
		DBName: tableName.Database,
		Schema: tableName.Schema,
		TBName: tableName.Table,
		IsCTE:  l.cteNames[tableName.Table], // 检查是否是CTE
	}

	if tableInfo.TBName != "" {
		l.tables = append(l.tables, tableInfo)
	}
}

// ExitAs_table_alias 退出表别名节点（处理别名）
func (l *sqlserverTableExtractListener) ExitAs_table_alias(ctx *parser.As_table_aliasContext) {
	if ctx.Table_alias() == nil || ctx.Table_alias().Id_() == nil {
		return
	}

	// 获取别名
	_, alias := tsqlparser.NormalizeTSQLIdentifier(ctx.Table_alias().Id_())

	// 更新最后一个表的别名（假设As_table_alias紧跟在表引用之后）
	if len(l.tables) > 0 {
		lastIdx := len(l.tables) - 1
		if l.tables[lastIdx].Alias == "" {
			l.tables[lastIdx].Alias = alias
		}
	}
}

// EnterCommon_table_expression 进入CTE定义节点
func (l *sqlserverTableExtractListener) EnterCommon_table_expression(ctx *parser.Common_table_expressionContext) {
	if ctx.GetExpression_name() == nil {
		return
	}
	
	// 获取CTE名称 - SQL Server的expression_name直接返回ID
	cteName := ctx.GetParser().GetTokenStream().GetTextFromRuleContext(ctx.GetExpression_name())
	if cteName != "" {
		// 移除方括号和引号，获取规范化的名称
		original, _ := tsqlparser.NormalizeTSQLIdentifierText(cteName)
		l.cteNames[original] = true
	}
}


