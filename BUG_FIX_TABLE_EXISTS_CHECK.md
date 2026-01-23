# Bug 修复：PostgreSQL 表存在性检查

## 问题描述

用户报告在使用 PostgreSQL 审核时，当创建已存在的表时，审核规则没有正确触发错误提示。

## 根本原因分析

经过调查，发现两个关键点：

### 1. 表存在性检查功能正常工作（有数据库连接时）

当提供数据库连接参数时，审核工具能够正确检测表是否已存在：

**代码位置**: `schema/pg/walk_through.go` 第 130-146 行

```go
// Check if table already exists
if schema.GetTable(tableName) != nil {
    // Check IF NOT EXISTS clause
    ifNotExists := ctx.IF_P() != nil && ctx.NOT() != nil && ctx.EXISTS() != nil
    if ifNotExists {
        return
    }
    l.advice = &storepb.Advice{
        Status:  storepb.Advice_ERROR,
        Code:    code.TableExists.Int32(),
        Title:   fmt.Sprintf(`The table %q already exists in the schema %q`, tableName, schema.GetProto().Name),
        Content: fmt.Sprintf(`The table %q already exists in the schema %q`, tableName, schema.GetProto().Name),
        StartPosition: &storepb.Position{
            Line: int32(l.currentLine),
        },
    }
    return
}
```

**测试结果**：
```bash
./build/advisor -engine postgres -file test.sql \
  -host 127.0.0.1 -port 5432 -user postgres -password secret \
  -dbname mydb -schema mydata
```

输出：
```
✗ ERROR: The table "user" already exists in the schema "mydata"
```

### 2. 发现的 Bug：无数据库连接时崩溃

当**没有**提供数据库连接参数时，审核工具会因为空指针异常而崩溃：

**错误堆栈**：
```
runtime error: invalid memory address or nil pointer dereference
advisor/pg/advisor_table_require_pk.go:190
```

**问题代码**：
```go
func (r *tableRequirePKRule) validateFinalState() {
    for tableKey, mention := range r.tableMentions {
        schemaName, tableName := parseTableKey(tableKey)
        
        // 🐛 Bug: 没有检查 finalMetadata 是否为 nil
        schema := r.finalMetadata.GetSchemaMetadata(schemaName)
        // ...
    }
}
```

## 修复方案

在 `advisor/pg/advisor_table_require_pk.go` 的 `validateFinalState()` 方法中添加 nil 检查：

```go
func (r *tableRequirePKRule) validateFinalState() {
    // ✅ 修复：添加 nil 检查
    if r.finalMetadata == nil {
        return
    }
    
    for tableKey, mention := range r.tableMentions {
        // ... 原有逻辑
    }
}
```

## 修复后的行为

### 场景 1：有数据库连接
```bash
./build/advisor -engine postgres -file test.sql \
  -host 127.0.0.1 -port 5432 -user postgres -password secret \
  -dbname mydb -schema mydata
```

**结果**：✅ 正确检测表是否存在
- 表已存在 → 报告 ERROR
- 表不存在 → 通过审核

### 场景 2：无数据库连接
```bash
./build/advisor -engine postgres -file test.sql
```

**修复前**：❌ 崩溃并报错
```
runtime error: invalid memory address or nil pointer dereference
```

**修复后**：✅ 正常运行
- 跳过需要元数据的检查
- 执行其他不需要数据库连接的规则检查

## 测试验证

### 测试 1：检测已存在的表（有数据库连接）

**SQL**:
```sql
CREATE TABLE "mydata"."user" (
  id BIGSERIAL not NULL,
  name TEXT NOT NULL,
  PRIMARY KEY (id)
);
```

**命令**:
```bash
./build/advisor -engine postgres -sql "CREATE TABLE \"mydata\".\"user\" (...)" \
  -host 127.0.0.1 -port 5432 -user postgres -password secret \
  -dbname mydb -schema mydata
```

**结果**:
```
✗ ERROR: The table "user" already exists in the schema "mydata"
```

### 测试 2：无数据库连接不崩溃

**命令**:
```bash
./build/advisor -engine postgres -sql "CREATE TABLE \"mydata\".\"user\" (...)"
```

**结果**:
```
✓ OK: Audit Completed (跳过需要元数据的检查)
```

## 结论

1. **表存在性检查功能本身是正常的**，在有数据库连接时能正确工作
2. **修复了一个 bug**：无数据库连接时的空指针崩溃问题
3. **用户需要确保**：要检查表是否已存在，必须提供数据库连接参数（`-host`, `-port`, `-user`, `-password`, `-dbname`, `-schema`）

## 使用建议

要启用完整的表存在性检查，请使用以下命令格式：

```bash
./build/advisor -engine postgres \
  -file your_sql_file.sql \
  -host 127.0.0.1 \
  -port 5432 \
  -user postgres \
  -password secret \
  -dbname mydb \
  -schema mydata
```

这样审核工具才能连接到数据库，获取元数据，并正确检查表是否已存在。

