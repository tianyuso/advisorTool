# PostgreSQL Schema 使用说明

## 重要说明

由于底层审核库（Bytebase advisor）在处理不带 schema 前缀的表名时存在 bug，**我们建议在 SQL 中始终使用完整的 schema.table 格式**。

## 当前实现

### SetSearchPath 标志

我们在 `ConnectionConfig` 中添加了 `SetSearchPath` 标志：

```go
type ConnectionConfig struct {
    // ... 其他参数 ...
    Schema        string // For PostgreSQL
    SetSearchPath bool   // 是否设置 search_path（仅用于影响行数计算）
}
```

### 使用场景

| 场景 | SetSearchPath | 说明 |
|------|--------------|------|
| 获取元数据（审核） | `false` | 不设置 search_path，避免审核时的表名解析问题 |
| 计算影响行数 | `true` | 设置 search_path，支持不带 schema 前缀的表名 |

## 推荐用法

### ✅ 推荐：使用完整表名

```sql
-- CREATE
CREATE TABLE mydata.test_users (id SERIAL PRIMARY KEY, ...);

-- UPDATE  
UPDATE mydata.test_users SET status = 'active' WHERE id = 1;

-- DELETE
DELETE FROM mydata.test_orders WHERE id > 100;

-- ALTER
ALTER TABLE mydata.test_users ADD COLUMN phone VARCHAR(20);
```

**优点：**
- ✅ 审核正常
- ✅ 影响行数计算正常
- ✅ SQL 明确易懂
- ✅ 不依赖 search_path

### ⚠️ 不推荐：混合使用

```sql
-- ❌ 不推荐：混合使用会导致审核混乱
CREATE TABLE mydata.test_users (...);   -- 带 schema
UPDATE test_users SET ...;               -- 不带 schema
DELETE FROM mydata.test_orders ...;      -- 带 schema
```

## 已知问题

### 底层库 Bug

在某些情况下，审核错误消息会引用错误的 SQL 文本片段：

```
错误SQL: UPDATE mydata.test_users SET status = 'active'
错误消息: "CREATE TABLE mydata.test_users (" requires WHERE clause
                     ^^^^^^^^^^^^^^^^^^^^^ 错误的引用
```

这是 Bytebase advisor 库的问题，不在我们的控制范围内。

## 配置示例

```go
// 数据库连接参数
dbParams := &services.DBConnectionParams{
    Host:     "127.0.0.1",
    Port:     5432,
    User:     "postgres",
    Password: "secret",
    DbName:   "mydb",
    SSLMode:  "disable",
    Timeout:  10,
    // 可选：指定 schema（仅用于影响行数计算）
    Schema:   "mydata",
}

// SQL 中使用完整表名
sql := `
CREATE TABLE mydata.test_users (id SERIAL PRIMARY KEY, ...);
UPDATE mydata.test_users SET status = 'active';
DELETE FROM mydata.test_users WHERE id > 100;
`
```

## 影响行数计算验证

即使使用完整表名，影响行数计算仍然正常工作：

```
order_id 9 (UPDATE 无 WHERE): affected_rows = 3 ✅
order_id 11 (DELETE 无 WHERE): affected_rows = 3 ✅
```

## 总结

1. ✅ **影响行数计算功能正常** - 带注释的 UPDATE/DELETE 能正确计算
2. ✅ **支持 search_path** - 但仅用于影响行数计算
3. ⚠️ **建议使用完整表名** - 避免审核库的 bug
4. 📝 **最佳实践** - 在 SQL 中始终使用 `schema.table` 格式

---

**更新日期：** 2024-12-17  
**状态：** 已实现并测试

