package extractobject

import (
	"context"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/bytebase/parser/mysql"

	storepb "github.com/tianyuso/advisorTool/generated-go/store"
	"github.com/tianyuso/advisorTool/parser/base"
	mysqlparser "github.com/tianyuso/advisorTool/parser/mysql"
)

// MySQLExtractor MySQL表名提取器
type MySQLExtractor struct{}

// Extract 从MySQL SQL语句中提取表名
func (e *MySQLExtractor) Extract(ctx context.Context, sql string) ([]TableInfo, error) {
	asts, err := parseSQL(storepb.Engine_MYSQL, sql)
	if err != nil {
		return nil, err
	}

	var tables []TableInfo
	for _, ast := range asts {
		antlrAST, ok := ast.(*base.ANTLRAST)
		if !ok {
			continue
		}

		listener := &mysqlTableExtractListener{
			tables:   make([]TableInfo, 0),
			cteNames: make(map[string]bool),
			tableMap: make(map[string]string),
		}

		antlr.ParseTreeWalkerDefault.Walk(listener, antlrAST.Tree)
		tables = append(tables, listener.tables...)
	}

	return tables, nil
}

// mysqlTableExtractListener MySQL表名提取监听器
type mysqlTableExtractListener struct {
	*parser.BaseMySQLParserListener
	tables   []TableInfo
	cteNames map[string]bool   // CTE名称集合
	tableMap map[string]string // 表名到别名的映射，用于避免重复处理
}

// EnterSingleTable 进入单表节点 (优先处理，包含别名信息)
func (l *mysqlTableExtractListener) EnterSingleTable(ctx *parser.SingleTableContext) {
	if ctx.TableRef() == nil {
		return
	}

	dbName, tableName := mysqlparser.NormalizeMySQLTableRef(ctx.TableRef())
	if tableName == "" {
		return
	}

	// 提取别名
	alias := ""
	if ctx.TableAlias() != nil && ctx.TableAlias().Identifier() != nil {
		alias = mysqlparser.NormalizeMySQLIdentifier(ctx.TableAlias().Identifier())
	}

	// 创建唯一key标记已处理
	key := dbName + "." + tableName
	if processedAlias, exists := l.tableMap[key]; exists {
		// 如果已经处理过，检查别名是否不同（可能是同一个表的不同别名引用）
		if alias != processedAlias {
			// 不同别名，添加新记录
			tableInfo := TableInfo{
				DBName: dbName,
				TBName: tableName,
				Alias:  alias,
				IsCTE:  l.cteNames[tableName],
			}
			l.tables = append(l.tables, tableInfo)
		}
		return
	}

	// 标记为已处理
	l.tableMap[key] = alias

	// 添加表信息
	tableInfo := TableInfo{
		DBName: dbName,
		TBName: tableName,
		Alias:  alias,
		IsCTE:  l.cteNames[tableName],
	}
	l.tables = append(l.tables, tableInfo)
}

// EnterTableRef 进入表引用节点（处理没有通过SingleTable的情况）
func (l *mysqlTableExtractListener) EnterTableRef(ctx *parser.TableRefContext) {
	// 只处理那些不在SingleTable上下文中的TableRef
	// 检查父节点是否是SingleTable
	if parent := ctx.GetParent(); parent != nil {
		if _, ok := parent.(*parser.SingleTableContext); ok {
			// 如果父节点是SingleTable，由EnterSingleTable处理
			return
		}
	}

	// 使用parser包的NormalizeMySQLTableRef函数
	dbName, tableName := mysqlparser.NormalizeMySQLTableRef(ctx)
	if tableName != "" {
		key := dbName + "." + tableName
		if _, exists := l.tableMap[key]; !exists {
			l.tableMap[key] = ""
			tableInfo := TableInfo{
				DBName: dbName,
				TBName: tableName,
				IsCTE:  l.cteNames[tableName],
			}
			l.tables = append(l.tables, tableInfo)
		}
	}
}

// EnterCommonTableExpression 进入CTE定义节点
func (l *mysqlTableExtractListener) EnterCommonTableExpression(ctx *parser.CommonTableExpressionContext) {
	if ctx.Identifier() == nil {
		return
	}
	
	// 获取CTE名称
	cteName := mysqlparser.NormalizeMySQLIdentifier(ctx.Identifier())
	l.cteNames[cteName] = true
}

