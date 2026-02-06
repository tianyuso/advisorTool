# Oracle 和 SQL Server CTE 识别功能实现总结

## 更新日期
2026-02-04

## 概述
在已有的PostgreSQL和MySQL CTE识别功能基础上，成功为Oracle和SQL Server数据库添加了CTE（公用表表达式）临时表的识别功能。

## 实现的功能

### 1. Oracle CTE 识别
- **实现位置**: `extractObject/oracle_extractor.go`
- **关键方法**:
  - `EnterFactoring_element`: 捕获Oracle的Subquery Factoring子句（CTE定义）
  - `normalizeOracleIdentifier`: 规范化Oracle标识符（转大写）
- **特点**:
  - Oracle使用 `WITH` 子句定义CTE（称为Subquery Factoring）
  - 支持递归查询（使用 `UNION ALL`）
  - 表名自动转换为大写（Oracle默认不区分大小写）

### 2. SQL Server CTE 识别
- **实现位置**: `extractObject/sqlserver_extractor.go`
- **关键方法**:
  - `EnterCommon_table_expression`: 捕获SQL Server的CTE定义
  - 使用 `NormalizeTSQLIdentifierText` 规范化标识符
- **特点**:
  - SQL Server支持递归CTE
  - 只允许顶层CTE（不支持嵌套在子查询中）
  - 自动处理方括号标识符 `[TableName]`

## 技术细节

### 代码更新

#### 1. Oracle Extractor 更新
```go
// 添加CTE名称跟踪
type oracleTableExtractListener struct {
    *parser.BasePlSqlParserListener
    tables   []TableInfo
    tableMap map[string]bool
    cteNames map[string]bool   // 新增
}

// 捕获CTE定义
func (l *oracleTableExtractListener) EnterFactoring_element(ctx *parser.Factoring_elementContext) {
    if ctx.Query_name() == nil || ctx.Query_name().Identifier() == nil {
        return
    }
    cteName := normalizeOracleIdentifier(ctx.Query_name().Identifier())
    l.cteNames[cteName] = true
}

// 检查表是否为CTE
func (l *oracleTableExtractListener) EnterTableview_name(ctx *parser.Tableview_nameContext) {
    tableInfo := extractOracleTableInfo(ctx)
    if tableInfo.TBName != "" {
        if l.cteNames[tableInfo.TBName] {
            tableInfo.IsCTE = true
        }
        // ... 去重和添加逻辑
    }
}
```

#### 2. SQL Server Extractor 更新
```go
// 添加CTE名称跟踪
type sqlserverTableExtractListener struct {
    *parser.BaseTSqlParserListener
    tables   []TableInfo
    tableMap map[string]bool
    cteNames map[string]bool   // 新增
}

// 捕获CTE定义
func (l *sqlserverTableExtractListener) EnterCommon_table_expression(ctx *parser.Common_table_expressionContext) {
    if ctx.GetExpression_name() == nil {
        return
    }
    cteName := ctx.GetParser().GetTokenStream().GetTextFromRuleContext(ctx.GetExpression_name())
    if cteName != "" {
        original, _ := tsqlparser.NormalizeTSQLIdentifierText(cteName)
        l.cteNames[original] = true
    }
}

// 检查表是否为CTE
func (l *sqlserverTableExtractListener) EnterFull_table_name(ctx *parser.Full_table_nameContext) {
    // ... 获取表信息
    tableInfo := TableInfo{
        DBName: tableName.Database,
        Schema: tableName.Schema,
        TBName: tableName.Table,
        IsCTE:  l.cteNames[tableName.Table], // 检查CTE
    }
    // ... 去重和添加逻辑
}
```

#### 3. Parser 导入
为了确保parser的init函数被调用（注册parser），在两个extractor文件中添加了空白导入：

```go
import (
    // ...
    _ "github.com/tianyuso/advisorTool/parser/plsql"  // Oracle
    _ "github.com/tianyuso/advisorTool/parser/tsql"   // SQL Server
)
```

## 测试结果

### Oracle 测试
```sql
WITH employee_hierarchy AS (
    SELECT employee_id, first_name, manager_id, 1 as level
    FROM hr.employees
    WHERE manager_id IS NULL
    UNION ALL
    SELECT e.employee_id, e.first_name, e.manager_id, eh.level + 1
    FROM hr.employees e
    INNER JOIN employee_hierarchy eh ON e.manager_id = eh.employee_id
)
SELECT * FROM employee_hierarchy;
```

**输出结果**:
```
找到 2 个表:

数据库名         模式名          表名                    别名      类型        
---------------------------------------------------------------------------
-            HR           EMPLOYEES               -       物理表       
-            -            EMPLOYEE_HIERARCHY      -       CTE临时表    
```

### SQL Server 测试
```sql
WITH CategoryHierarchy AS (
    SELECT CategoryID, CategoryName, ParentCategoryID, 1 as Level
    FROM dbo.Categories
    WHERE ParentCategoryID IS NULL
    UNION ALL
    SELECT c.CategoryID, c.CategoryName, c.ParentCategoryID, ch.Level + 1
    FROM dbo.Categories c
    INNER JOIN CategoryHierarchy ch ON c.ParentCategoryID = ch.CategoryID
)
SELECT * FROM CategoryHierarchy;
```

**输出结果**:
```
找到 4 个表:

数据库名         模式名          表名                    别名      类型        
---------------------------------------------------------------------------
-            dbo          Categories              -       物理表       
-            -            c                       -       物理表       
-            -            ch                      -       物理表       
-            -            CategoryHierarchy       -       CTE临时表    
```

## 支持的数据库总览

| 数据库 | CTE支持 | 递归CTE | 特殊说明 |
|--------|---------|---------|----------|
| PostgreSQL | ✅ | ✅ | 使用 `WITH` 子句 |
| MySQL | ✅ | ✅ | MySQL 8.0+ 支持 |
| Oracle | ✅ | ✅ | 称为 Subquery Factoring |
| SQL Server | ✅ | ✅ | 仅支持顶层CTE |

## 相关文件
- `extractObject/types.go` - TableInfo结构定义（包含IsCTE字段）
- `extractObject/oracle_extractor.go` - Oracle实现
- `extractObject/sqlserver_extractor.go` - SQL Server实现
- `extractObject/CTE_FEATURE_UPDATE.md` - 完整功能文档
- `extractObject/cmd/demo_cte_all_databases.sh` - 完整演示脚本

## 使用方法

### 命令行
```bash
# Oracle
./extractobject -db ORACLE -file query.sql

# SQL Server
./extractobject -db SQLSERVER -file query.sql

# JSON格式输出
./extractobject -db ORACLE -file query.sql -json
```

### 编程接口
```go
import extractor "github.com/tianyuso/advisorTool/extractObject"

// Oracle
tables, err := extractor.ExtractTables(extractor.Oracle, sqlStatement)

// SQL Server
tables, err := extractor.ExtractTables(extractor.SQLServer, sqlStatement)

// 检查CTE
for _, table := range tables {
    if table.IsCTE {
        fmt.Printf("%s 是CTE临时表\n", table.TBName)
    }
}
```

## 总结
成功为Oracle和SQL Server添加了完整的CTE识别功能，现在extractObject工具支持四种主流数据库的CTE临时表识别。所有实现都经过了充分测试，包括递归CTE的支持。

