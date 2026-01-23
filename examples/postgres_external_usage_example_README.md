# PostgreSQL 外部使用示例

本示例展示如何在外部 Go 程序中使用 `advisorTool/services` 包对 PostgreSQL 进行 SQL 审核。

## 文件说明

- **文件**: `postgres_external_usage_example.go`
- **功能**: 连接真实的 PostgreSQL 数据库，获取元数据，执行全面的 SQL 审核

## 数据库配置

本示例使用以下数据库连接参数：

```go
Host:     "127.0.0.1"
Port:     5432
User:     "postgres"
Password: "secret"
DbName:   "mydb"
Schema:   "mydata"
SSLMode:  "disable"
Timeout:  10
```

## 测试 SQL 类型

示例包含了以下类型的 SQL 语句：

### 1. DDL - 数据定义语句

#### 建表语句
```sql
CREATE TABLE mydata.test_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    age INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE mydata.test_orders (
    order_id SERIAL PRIMARY KEY,
    user_id INT,
    order_no VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2),
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 创建索引
```sql
CREATE INDEX idx_test_users_username ON mydata.test_users(username);
CREATE INDEX idx_test_users_email ON mydata.test_users(email);
CREATE INDEX idx_test_orders_user_id ON mydata.test_orders(user_id);
```

**注意**: PostgreSQL 推荐使用 `CONCURRENTLY` 关键字避免锁表：
```sql
CREATE INDEX CONCURRENTLY idx_name ON table(column);
```

#### 修改表结构
```sql
-- 添加列
ALTER TABLE mydata.test_users ADD COLUMN phone VARCHAR(20);

-- 修改列类型
ALTER TABLE mydata.test_users ALTER COLUMN username TYPE VARCHAR(100);

-- 删除列
ALTER TABLE mydata.test_users DROP COLUMN age;
```

### 2. DML - 数据操作语句

#### SELECT 查询
```sql
-- ❌ 不推荐：使用 SELECT *
SELECT * FROM mydata.test_users WHERE id = 1;

-- ✅ 推荐：明确指定列
SELECT id, username, email FROM mydata.test_users WHERE status = 'active';
```

#### UPDATE 更新
```sql
-- ✅ 正常：有 WHERE 条件
UPDATE mydata.test_users 
SET status = 'inactive', updated_at = CURRENT_TIMESTAMP 
WHERE id = 100;

-- ❌ 危险：没有 WHERE 条件（会触发错误）
UPDATE mydata.test_users SET status = 'active';
```

#### DELETE 删除
```sql
-- ✅ 正常：有 WHERE 条件
DELETE FROM mydata.test_orders WHERE order_date < '2023-01-01';

-- ❌ 危险：没有 WHERE 条件（会触发错误）
DELETE FROM mydata.test_users;
```

## 运行示例

### 1. 准备数据库

确保 PostgreSQL 正在运行，并且：

```bash
# 创建数据库（如果不存在）
psql -U postgres -c "CREATE DATABASE mydb;"

# 创建 schema
psql -U postgres -d mydb -c "CREATE SCHEMA IF NOT EXISTS mydata;"
```

### 2. 编译并运行

```bash
# 编译
go build -o postgres_example postgres_external_usage_example.go

# 运行
./postgres_example
```

或者直接运行：

```bash
go run postgres_external_usage_example.go
```

## 输出说明

示例会输出以下几种格式的审核结果：

### 1. 详细结果格式
```
📋 审核结果详情
======================================================================

❌ SQL #1 [✗ ERROR]
   SQL: UPDATE mydata.test_users SET status = 'active'
   问题: [statement.where.require.update-delete] "UPDATE..." requires WHERE clause

⚠️  SQL #2 [⚠ WARNING]
   SQL: SELECT * FROM mydata.test_users WHERE id = 1
   问题: [statement.select.no-select-all] "SELECT * FROM..." uses SELECT all
```

### 2. 统计信息
```
📊 统计信息
======================================================================
总 SQL 语句数: 14
✅ 通过: 11
⚠️  警告: 1
❌ 错误: 2
```

### 3. JSON 格式（兼容 Inception）
```json
[
  {
    "order_id": 1,
    "stage": "CHECKED",
    "error_level": "2",
    "stage_status": "Audit Completed",
    "error_message": "[statement.where.require] ...",
    "sql": "UPDATE mydata.test_users SET status = 'active'",
    "affected_rows": 0,
    "sequence": "0_0_00000000"
  }
]
```

### 4. 表格格式
使用 `go-pretty` 库输出美观的表格，包含颜色标识。

## 使用的 Services 包功能

本示例展示了 `services` 包的核心功能：

### 1. 数据库元数据获取
```go
metadata, err := services.FetchDatabaseMetadata(engineType, dbParams)
```

### 2. 规则加载
```go
// 自动加载适合 PostgreSQL 的默认规则
// hasMetadata=true 会包含需要元数据的高级规则
rules, err := services.LoadRules("", engineType, hasMetadata)
```

### 3. 影响行数计算
```go
affectedRowsMap := services.CalculateAffectedRowsForStatements(sql, engineType, dbParams)
```

### 4. 结果转换
```go
// 转换为 Inception 兼容的结构化格式
results := services.ConvertToReviewResults(resp, sql, engineType, affectedRowsMap)
```

### 5. 格式化输出
```go
// JSON 格式
services.OutputResults(resp, sql, engineType, "json", dbParams)

// 表格格式
services.OutputResults(resp, sql, engineType, "table", dbParams)
```

## PostgreSQL 特定审核规则

本示例会检查以下 PostgreSQL 特定规则：

1. ✅ **索引并发创建** - 推荐使用 `CONCURRENTLY` 关键字
2. ✅ **添加列默认值** - 避免带默认值直接添加列（可能锁表）
3. ✅ **约束验证** - 推荐使用 `NOT VALID` 然后再验证
4. ✅ **完全限定名** - 推荐使用 `schema.table` 格式
5. ✅ **WHERE 子句要求** - UPDATE/DELETE 必须有 WHERE
6. ✅ **SELECT * 禁止** - 应明确指定列名
7. ✅ **主键要求** - 表必须有主键
8. ✅ **外键建议** - 根据配置可能禁止外键
9. ✅ **向后兼容性** - 检查 schema 变更的兼容性（需要元数据）
10. ✅ **列 NULL 检查** - 检查列定义（需要元数据）

## 常见问题

### Q1: 连接数据库失败怎么办？

如果看到以下错误：
```
⚠️  警告: 获取数据库元数据失败: connection refused
将使用基础规则进行审核（跳过需要元数据的规则）
```

**解决方法**：
1. 确保 PostgreSQL 正在运行
2. 检查连接参数（host, port, user, password）
3. 确保数据库和 schema 存在
4. 检查防火墙设置

即使无法连接数据库，示例仍会继续运行，只是会跳过需要元数据的高级规则。

### Q2: 表已存在的错误

如果看到：
```
❌ SQL #1 [✗ ERROR]
   问题: The table "test_users" already exists in the schema "mydata"
```

这是正常的向后兼容性检查。如果表已存在，尝试再次创建会产生错误。

**解决方法**：
1. 删除现有表：`DROP TABLE mydata.test_users CASCADE;`
2. 或者修改 SQL 使用 `CREATE TABLE IF NOT EXISTS`

### Q3: 如何自定义规则？

可以提供自己的配置文件：

```go
// 使用自定义配置
rules, err := services.LoadRules("my-postgres-config.yaml", engineType, hasMetadata)
```

或者使用 `services.GenerateSampleConfig()` 生成示例配置：

```go
config := services.GenerateSampleConfig(advisor.EnginePostgres)
fmt.Println(config)
```

### Q4: 如何禁用某些规则？

在自定义配置文件中设置规则级别为 `DISABLED`：

```yaml
rules:
  - type: statement.select.no-select-all
    level: DISABLED  # 禁用此规则
```

## PostgreSQL 最佳实践提示

示例结尾会输出 PostgreSQL 特定的最佳实践建议：

1. **创建索引使用 CONCURRENTLY**
   ```sql
   CREATE INDEX CONCURRENTLY idx_name ON table(column);
   ```

2. **添加带默认值的列分两步**
   ```sql
   -- 第一步：添加列（不带默认值）
   ALTER TABLE ADD COLUMN without DEFAULT;
   
   -- 第二步：更新值
   UPDATE TABLE SET column = value;
   ```

3. **添加约束使用 NOT VALID**
   ```sql
   ALTER TABLE ADD CONSTRAINT ... CHECK (...) NOT VALID;
   ALTER TABLE VALIDATE CONSTRAINT ...;
   ```

4. **使用完全限定名**
   ```sql
   SELECT * FROM mydata.test_users;  -- 推荐
   SELECT * FROM test_users;         -- 不推荐
   ```

## 扩展阅读

- [services 包文档](../services/README.md)
- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)
- [Bytebase SQL Review](https://www.bytebase.com/docs/sql-review/overview)

## 许可证

遵循 Bytebase 项目的许可证。









