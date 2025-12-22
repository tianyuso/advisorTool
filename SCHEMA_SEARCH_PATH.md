# PostgreSQL Schema Search Path 功能说明

## 功能概述

为 PostgreSQL 数据库添加自动设置 `search_path` 功能，使得在审核 SQL 时可以直接使用表名，而不需要使用 `schema.table` 的完整形式。

## 问题背景

在 PostgreSQL 中，如果不显式指定 schema，查询和操作会使用当前的 `search_path`。默认情况下，`search_path` 只包含 `public` schema。当需要操作其他 schema（如 `mydata`）中的表时，有两种方式：

1. **完整表名**：`mydata.test_users` ❌ 繁琐
2. **设置 search_path**：`SET search_path TO mydata, public` ✅ 简洁

## 实现方案

### 修改的文件

1. **`db/connection.go`** - 在连接建立后自动设置 search_path
2. **`services/result.go`** - 传递 Schema 参数
3. **`services/metadata.go`** - 传递 Schema 参数

### 核心实现

#### 1. 连接层自动设置 (db/connection.go)

```go
// OpenConnection opens a database connection based on the configuration.
func OpenConnection(ctx context.Context, config *ConnectionConfig) (*sql.DB, error) {
    // ... 连接数据库 ...
    
    // For PostgreSQL, set search_path if Schema is specified
    if config.DbType == "postgres" && config.Schema != "" {
        searchPathSQL := fmt.Sprintf("SET search_path TO %s, public", config.Schema)
        if _, err := db.ExecContext(ctx, searchPathSQL); err != nil {
            db.Close()
            return nil, fmt.Errorf("failed to set search_path: %w", err)
        }
    }
    
    return db, nil
}
```

#### 2. 配置传递

在 `services/result.go` 和 `services/metadata.go` 中，都需要传递 `Schema` 参数：

```go
config := &db.ConnectionConfig{
    DbType:      GetDbTypeString(engineType),
    Host:        dbParams.Host,
    Port:        dbParams.Port,
    // ... 其他参数 ...
    Schema:      dbParams.Schema,  // ✅ 新增
}
```

## 使用方法

### 基本用法

```go
// 1. 设置数据库连接参数，指定 Schema
dbParams := &services.DBConnectionParams{
    Host:     "127.0.0.1",
    Port:     5432,
    User:     "postgres",
    Password: "secret",
    DbName:   "mydb",
    SSLMode:  "disable",
    Timeout:  10,
    Schema:   "mydata",  // ✅ 指定 schema
}

// 2. 获取元数据（自动设置 search_path）
metadata, err := services.FetchDatabaseMetadata(advisor.EnginePostgres, dbParams)

// 3. 审核 SQL（可以直接使用表名）
sql := `
    UPDATE test_users SET status = 'active' WHERE id = 1;
    DELETE FROM test_orders WHERE order_date < '2023-01-01';
`

// 4. 计算影响行数（自动设置 search_path）
affectedRowsMap := services.CalculateAffectedRowsForStatements(sql, engineType, dbParams)
```

### 对比示例

#### 修改前 ❌

```sql
-- 必须使用完整表名
UPDATE mydata.test_users SET status = 'active';
DELETE FROM mydata.test_orders WHERE id > 100;
```

#### 修改后 ✅

```go
// 设置 Schema 参数
dbParams.Schema = "mydata"

// SQL 中可以直接使用表名
sql := `
    UPDATE test_users SET status = 'active';
    DELETE FROM test_orders WHERE id > 100;
`
```

## 功能验证

### 测试结果

运行 `examples/test_schema_search_path.go` 验证功能：

```bash
go run examples/test_schema_search_path.go
```

**测试输出：**

```
✅ 设置 search_path 为: mydata, public
✅ 影响行数计算正常（总计: 10 行）
✅ search_path 设置成功，可以直接使用表名而无需 schema 前缀
```

### 测试覆盖

- ✅ **元数据获取**：可以正确获取指定 schema 的表
- ✅ **SQL 审核**：不带 schema 前缀的 SQL 能正常审核
- ✅ **影响行数计算**：不带 schema 前缀的 SQL 能正确计算影响行数
- ✅ **连接池**：每个连接都正确设置 search_path

## 技术细节

### search_path 语法

```sql
SET search_path TO mydata, public;
```

- **第一个 schema** (`mydata`)：优先搜索
- **第二个 schema** (`public`)：回退搜索
- 当查询 `test_users` 时，PostgreSQL 会：
  1. 先在 `mydata` schema 中查找
  2. 如果找不到，再在 `public` schema 中查找

### 生效范围

- **连接级别**：设置在连接建立后立即执行
- **连接池**：每个新连接都会自动设置
- **会话级别**：仅影响当前连接会话

### 兼容性

| 功能 | 是否兼容 | 说明 |
|------|---------|------|
| 不指定 Schema | ✅ | 不设置 search_path，保持默认行为 |
| 指定 Schema | ✅ | 自动设置 search_path |
| 混合使用 | ✅ | 可以同时使用 `table` 和 `schema.table` |
| 其他数据库 | ✅ | 不影响 MySQL、SQL Server 等 |

## 注意事项

### 1. Schema 必须存在

```go
// ❌ 错误：schema 不存在
dbParams.Schema = "nonexistent"

// ✅ 正确：确保 schema 存在
dbParams.Schema = "mydata"
```

### 2. 表名冲突

如果 `mydata` 和 `public` 中都有 `test_users` 表：
- 优先使用 `mydata.test_users`
- 如果要使用 `public.test_users`，需要显式指定：`public.test_users`

### 3. 权限要求

用户需要有目标 schema 的访问权限：

```sql
GRANT USAGE ON SCHEMA mydata TO postgres;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA mydata TO postgres;
```

## 实际应用场景

### 场景 1: 多租户应用

```go
// 为不同租户使用不同的 schema
tenantID := "tenant_123"
dbParams.Schema = tenantID

// 所有 SQL 都自动路由到对应租户的 schema
sql := "SELECT * FROM users"  // 实际查询 tenant_123.users
```

### 场景 2: 开发/测试环境隔离

```go
// 开发环境
dbParams.Schema = "dev"

// 测试环境
dbParams.Schema = "test"

// 生产环境
dbParams.Schema = "prod"
```

### 场景 3: 简化 SQL 迁移

```go
// 从 MySQL 迁移到 PostgreSQL
// MySQL: 数据库名 = mydata
// PostgreSQL: schema 名 = mydata

dbParams.Schema = "mydata"

// SQL 可以保持不变
sql := "UPDATE users SET status = 'active'"
```

## 总结

✅ **自动化**：连接时自动设置，无需手动执行 SQL  
✅ **透明化**：对现有代码无侵入，只需设置 Schema 参数  
✅ **灵活性**：支持指定或不指定 schema  
✅ **兼容性**：不影响其他数据库类型  
✅ **完整性**：覆盖元数据获取、SQL 审核、影响行数计算

---

**更新日期：** 2024-12-17  
**相关文件：**
- `db/connection.go`
- `services/result.go`
- `services/metadata.go`
- `examples/test_schema_search_path.go`

