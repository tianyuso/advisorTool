package extractobject

import (
	"context"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/bytebase/parser/postgresql"

	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
	pgparser "github.com/tianyuso/advisorTool/parser/pg"
)

// PostgreSQLExtractor PostgreSQL表名提取器
type PostgreSQLExtractor struct{}

// Extract 从PostgreSQL SQL语句中提取表名
func (e *PostgreSQLExtractor) Extract(ctx context.Context, sql string) ([]TableInfo, error) {
	asts, err := parseSQL(storepb.Engine_POSTGRES, sql)
	if err != nil {
		return nil, err
	}

	var tables []TableInfo
	for _, ast := range asts {
		antlrAST, ok := ast.(*base.ANTLRAST)
		if !ok {
			continue
		}

		listener := &postgresqlTableExtractListener{
			tables:   make([]TableInfo, 0),
			tableMap: make(map[string]bool),
			cteNames: make(map[string]bool),
		}

		antlr.ParseTreeWalkerDefault.Walk(listener, antlrAST.Tree)
		tables = append(tables, listener.tables...)
	}

	return tables, nil
}

// postgresqlTableExtractListener PostgreSQL表名提取监听器
type postgresqlTableExtractListener struct {
	*parser.BasePostgreSQLParserListener
	tables   []TableInfo
	tableMap map[string]bool // 用于去重
	cteNames map[string]bool // CTE名称集合
}

// EnterTable_ref 进入表引用节点
func (l *postgresqlTableExtractListener) EnterTable_ref(ctx *parser.Table_refContext) {
	if ctx.Relation_expr() == nil || ctx.Relation_expr().Qualified_name() == nil {
		return
	}

	// 使用parser包的NormalizePostgreSQLQualifiedName函数
	parts := pgparser.NormalizePostgreSQLQualifiedName(ctx.Relation_expr().Qualified_name())
	if len(parts) == 0 {
		return
	}

	tableInfo := TableInfo{}
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

	if tableInfo.TBName != "" {
		// 检查是否是CTE
		if l.cteNames[tableInfo.TBName] {
			tableInfo.IsCTE = true
		}

		// 提取别名
		if ctx.Opt_alias_clause() != nil && ctx.Opt_alias_clause().Table_alias_clause() != nil {
			aliasClause := ctx.Opt_alias_clause().Table_alias_clause()
			if aliasClause.Table_alias() != nil && aliasClause.Table_alias().Identifier() != nil {
				// 规范化别名（PostgreSQL不区分大小写，转小写）
				aliasText := aliasClause.Table_alias().Identifier().GetText()
				tableInfo.Alias = strings.ToLower(aliasText)
			}
		}

		// 不再使用简单的去重，允许同一张表的多次引用（可能有不同的别名）
		l.tables = append(l.tables, tableInfo)
	}
}

// EnterCommon_table_expr 进入CTE定义节点
func (l *postgresqlTableExtractListener) EnterCommon_table_expr(ctx *parser.Common_table_exprContext) {
	if ctx.Name() == nil {
		return
	}

	// 获取CTE名称
	cteName := pgparser.NormalizePostgreSQLName(ctx.Name())
	l.cteNames[cteName] = true
}
