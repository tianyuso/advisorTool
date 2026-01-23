# 修复：非 SQL 语句审核不报错问题

## 问题描述

当用户提交的内容不是有效的 SQL 语句时（例如：`转正考试分数记录表`），审核系统不报错，而是显示审核通过（`error_level: "0"`）。

### 问题样例

**输入**：
```bash
./advisor -engine sqlserver -sql "转正考试分数记录表" -host 127.0.0.1 -port 1433 -user sa -password "xxx" -dbname Psadata -schema dbo -format json
```

**修复前的输出**（错误）：
```json
[{
  "order_id":1,
  "stage":"CHECKED",
  "error_level":"0",  // ← 错误：应该报错
  "stage_status":"Audit Completed",
  "error_message":"",  // ← 错误：应该有语法错误消息
  "sql":"转正考试分数记录表",
  "affected_rows":0
}]
```

## 根本原因

1. **解析器层面**：ANTLR TSQL 解析器对某些非 SQL 文本（如纯标识符）不会触发语法错误，可能返回空的或部分有效的解析树
2. **审核逻辑层面**：
   - 如果解析器没有报错且返回了 AST（即使是无效的），审核逻辑会认为解析成功
   - 如果没有任何审核建议（Advice），系统会默认认为审核通过

## 修复方案

采用**方案3**：多层防御策略

### 修改 1：增强 SQL Review 判断逻辑

**文件**：`/data/dev_go/advisorTool/advisor/sql_review.go`

**修改内容**：在 `SQLReviewCheck` 函数中增加三个优先级判断：

1. **优先级 1**：如果解析器报告了语法错误，优先返回解析错误
2. **优先级 2**：如果没有生成有效的 AST，返回语法错误（捕获解析器未报告的无效输入）
3. **优先级 3**：如果没有审核规则，返回空（表示所有检查通过）

```go
// Priority 1: Return parse errors if any (syntax errors from parser)
if len(parseResult) > 0 {
    return parseResult, nil
}

// Priority 2: If no valid AST was generated, return syntax error
// This catches cases where the input is not valid SQL but parser didn't report error
if asts == nil || len(asts) == 0 {
    return []*storepb.Advice{
        {
            Status:  storepb.Advice_ERROR,
            Code:    201, // StatementSyntaxErrorCode
            Title:   SyntaxErrorTitle,
            Content: "Invalid SQL statement: no valid SQL syntax found",
            StartPosition: &storepb.Position{
                Line:   1,
                Column: 1,
            },
        },
    }, nil
}

// Priority 3: If no rules to check, return empty (all checks passed)
if len(ruleList) == 0 {
    return nil, nil
}
```

### 修改 2：增强 TSQL 解析器验证

**文件**：`/data/dev_go/advisorTool/parser/tsql/tsql.go`

**修改内容**：在 `parseSingleTSQL` 函数中，解析完成后验证解析树是否包含有效的 SQL 语句：

```go
// Validate that the parse tree contains valid SQL statements
if tree == nil || !containsValidTSQLStatement(tree) {
    return nil, &base.SyntaxError{
        Position: startPosition,
        Message: "Invalid SQL statement: no valid SQL syntax found",
        RawMessage: "no valid SQL syntax found",
    }
}
```

**新增函数**：`containsValidTSQLStatement`

该函数检查解析树是否包含至少一个有效的 SQL 语句类型：
- DML 语句（SELECT, INSERT, UPDATE, DELETE）
- DDL 语句（CREATE, ALTER, DROP）
- 其他语句（USE, SET, DECLARE）
- 备份/恢复语句
- 事务控制语句（BEGIN TRAN, COMMIT, ROLLBACK）
- DBCC 语句

## 修复效果

### 测试用例 1：非 SQL 文本（中文）

**输入**：
```bash
./advisor -engine sqlserver -sql "转正考试分数记录表" -format json
```

**输出**：
```json
[{
  "order_id":1,
  "stage":"CHECKED",
  "error_level":"2",  // ✓ 正确：错误级别
  "stage_status":"Audit Completed",
  "error_message":"[Syntax error] Invalid SQL statement: no valid SQL syntax found",  // ✓ 正确：语法错误提示
  "sql":"转正考试分数记录表",
  "affected_rows":0
}]
```

**退出码**：2（表示有错误）

### 测试用例 2：纯字母数字

**输入**：
```bash
./advisor -engine sqlserver -sql "abc123" -format json
```

**输出**：
```json
[{
  "order_id":1,
  "stage":"CHECKED",
  "error_level":"2",
  "stage_status":"Audit Completed",
  "error_message":"[Syntax error] Invalid SQL statement: no valid SQL syntax found",
  "sql":"abc123",
  "affected_rows":0
}]
```

### 测试用例 3：普通文本

**输入**：
```bash
./advisor -engine sqlserver -sql "hello world" -format json
```

**输出**：
```json
[{
  "order_id":1,
  "stage":"CHECKED",
  "error_level":"2",
  "stage_status":"Audit Completed",
  "error_message":"[Syntax error] Invalid SQL statement: no valid SQL syntax found",
  "sql":"hello world",
  "affected_rows":0
}]
```

### 测试用例 4：有效的 SQL 语句（验证正常功能）

**输入**：
```bash
./advisor -engine sqlserver -sql "SELECT * FROM Users" -format json
```

**输出**：
```json
[{
  "order_id":1,
  "stage":"CHECKED",
  "error_level":"1",  // ✓ 警告级别（因为使用了 SELECT *）
  "stage_status":"Audit Completed",
  "error_message":"[statement.select.no-select-all] Avoid using SELECT *.\n[statement.where.require.select] WHERE clause is required for SELETE statement.",
  "sql":"SELECT * FROM Users",
  "affected_rows":0
}]
```

**退出码**：1（表示有警告）

### 测试用例 5：有效的 DDL 语句

**输入**：
```bash
./advisor -engine sqlserver -sql "CREATE TABLE test (id INT PRIMARY KEY, name NVARCHAR(100))" -format table
```

**输出**：
```
+------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
|  ORDER    STAGE     LEVEL        STATUS                                  SQL                              AFFECTED    SEQUENCE     BACKUP DB   EXEC TIME    SQL SHA1     BACKUP TIME      ERROR MESSAGE    |
|    1     CHECKED    ✓ OK    Audit Completed  CREATE TABLE test (id INT PRIMARY KEY, name NVARCHAR(100))      0      0_0_00000000       -           0      -                   0       -                    |
+------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+

Summary:
  Total Statements: 1
  ✓ Success: 1
```

**退出码**：0（表示成功）

## 影响范围

1. **影响的引擎**：主要是 SQL Server (MSSQL/TSQL)
2. **向后兼容性**：完全兼容，不影响有效 SQL 语句的审核
3. **性能影响**：轻微（增加了解析树验证），对性能影响可忽略不计

## 其他数据库引擎

虽然此次修复主要针对 SQL Server，但在 `advisor/sql_review.go` 中的修改对所有数据库引擎都生效，提供了第一层防护。

如果其他数据库引擎也存在类似问题，可以参考 TSQL 的修复方式，在对应的解析器中添加 `containsValidStatement` 函数。

## 修改文件列表

1. `/data/dev_go/advisorTool/advisor/sql_review.go` - 增强审核判断逻辑
2. `/data/dev_go/advisorTool/parser/tsql/tsql.go` - 增强 TSQL 解析器验证

## 测试建议

建议在以下场景进行测试：
1. 非 SQL 文本输入（中文、英文、特殊字符等）
2. 空白输入
3. 各种有效的 SQL 语句（SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, DROP 等）
4. 包含语法错误的 SQL 语句
5. 多语句批处理

## 总结

通过多层防御策略，成功修复了非 SQL 语句不报错的问题：
- **第一层**：SQL Review 层面的 AST 有效性检查
- **第二层**：TSQL 解析器层面的语句类型验证

这种方式确保了即使某一层未能捕获无效输入，下一层仍能提供保护。

