# 影响行数计算修复总结

## 问题描述

当 SQL 语句以注释开头时（例如：`-- 注释\nUPDATE ...`），`CalculateAffectedRows` 函数无法正确识别 UPDATE/DELETE 语句类型，导致影响行数始终返回 0。

## 根本原因

1. **`getStatementType` 函数**：使用 `strings.HasPrefix()` 检查语句类型，但没有跳过前导注释
2. **SQL 改写函数**：`rewriteToCountSQL` 及其各个引擎的实现函数在解析 SQL 时，也受到前导注释的干扰

## 修复方案

### 1. 修复 `getStatementType` 函数 (db/affected_rows.go)

**修改前：**
```go
func getStatementType(statement string, engine advisor.Engine) string {
	trimmed := strings.TrimSpace(statement)
	upper := strings.ToUpper(trimmed)

	if strings.HasPrefix(upper, "UPDATE") {
		return "UPDATE"
	}
	if strings.HasPrefix(upper, "DELETE") {
		return "DELETE"
	}
	return "UNKNOWN"
}
```

**修改后：**
```go
func getStatementType(statement string, engine advisor.Engine) string {
	// 逐行扫描，跳过注释和空行，找到第一个实际的 SQL 语句
	lines := strings.Split(statement, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 跳过空行和注释行
		if trimmed == "" || strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
			continue
		}
		
		// 找到第一个非注释行，检查语句类型
		upper := strings.ToUpper(trimmed)
		if strings.HasPrefix(upper, "UPDATE") {
			return "UPDATE"
		}
		if strings.HasPrefix(upper, "DELETE") {
			return "DELETE"
		}
		return "UNKNOWN"
	}
	return "UNKNOWN"
}
```

### 2. 添加 `removeLeadingComments` 辅助函数

```go
// removeLeadingComments 移除 SQL 语句前导的注释行
func removeLeadingComments(statement string) string {
	lines := strings.Split(statement, "\n")
	var sqlLines []string
	foundSQL := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !foundSQL {
			if trimmed == "" || strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") {
				continue
			}
			foundSQL = true
		}
		sqlLines = append(sqlLines, line)
	}
	
	return strings.Join(sqlLines, "\n")
}
```

### 3. 为所有引擎的改写函数应用注释移除

- `rewriteMySQLToCount`
- `rewritePostgresToCount`
- `rewriteTSQLToCount`
- `rewriteOracleToCount`

每个函数在解析前都先调用 `removeLeadingComments(statement)`。

## 测试验证

### 测试用例

```sql
-- ===== UPDATE 语句 =====
-- 正常的 UPDATE（有 WHERE 条件）
UPDATE mydata.test_users 
SET status = 'inactive', updated_at = CURRENT_TIMESTAMP 
WHERE id = 100;

-- 危险的 UPDATE（没有 WHERE 条件）
UPDATE mydata.test_users SET status = 'active';

-- ===== DELETE 语句 =====
-- 正常的 DELETE（有 WHERE 条件）
DELETE FROM mydata.test_orders WHERE order_date < '2023-01-01';

-- 危险的 DELETE（没有 WHERE 条件）
DELETE FROM mydata.test_users;
```

### 测试结果

✅ **UPDATE with WHERE**: 影响行数 = 0 (没有匹配记录)
✅ **UPDATE without WHERE**: 影响行数 = 5 (表中有 5 条记录)
✅ **DELETE with WHERE**: 影响行数 = 0 (没有匹配记录)
✅ **DELETE without WHERE**: 影响行数 = 5 (表中有 5 条记录)

### SQL 改写示例

原始 SQL:
```sql
-- 危险的 UPDATE（没有 WHERE 条件）
UPDATE mydata.test_users SET status = 'active'
```

改写后的 COUNT SQL:
```sql
SELECT COUNT(1) FROM mydata.test_users
```

## 影响范围

- ✅ PostgreSQL
- ✅ MySQL/MariaDB/TiDB/OceanBase
- ✅ SQL Server
- ✅ Oracle

所有支持的数据库引擎都已应用此修复。

## 相关文件

- `db/affected_rows.go` - 核心修复
- `services/result.go` - 调用层（无需修改）
- `examples/postgres_external_usage_example.go` - 测试用例

## 总结

此修复解决了当 SQL 语句包含前导注释时，无法正确计算 UPDATE/DELETE 影响行数的问题。修复方案通过跳过注释行来识别真正的 SQL 语句类型，并在改写 SQL 前移除前导注释，确保 SQL 解析器能够正确处理。

