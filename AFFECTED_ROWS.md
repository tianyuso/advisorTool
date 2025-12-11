# 影响行数计算功能说明

## 功能概述

本工具现在支持自动计算 UPDATE 和 DELETE 语句的影响行数。当提供数据库连接参数时，工具会将 UPDATE/DELETE 语句改写为 SELECT COUNT(1) 查询，以估算影响的行数。

## 支持的数据库

- ✅ MySQL
- ✅ MariaDB
- ✅ TiDB
- ✅ OceanBase
- ✅ PostgreSQL
- ✅ SQL Server (MSSQL)
- ✅ Oracle

## 改写规则

### MySQL 语法

#### 单表 UPDATE
```sql
-- 原始语句
UPDATE users SET name = 'test' WHERE id > 100;

-- 改写为
SELECT COUNT(1) FROM users WHERE id > 100;
```

#### 连表 UPDATE
```sql
-- 原始语句
UPDATE t1 INNER JOIN t2 ON t1.id = t2.id 
SET t1.name = t2.name 
WHERE t1.status = 1;

-- 改写为
SELECT COUNT(1) FROM t1 INNER JOIN t2 ON t1.id = t2.id 
WHERE t1.status = 1;
```

#### 单表 DELETE
```sql
-- 原始语句
DELETE FROM users WHERE id > 100;

-- 改写为
SELECT COUNT(1) FROM users WHERE id > 100;
```

#### 连表 DELETE
```sql
-- 原始语句
DELETE t1 FROM t1 INNER JOIN t2 ON t1.id = t2.id WHERE t1.status = 1;

-- 改写为
SELECT COUNT(1) FROM t1 INNER JOIN t2 ON t1.id = t2.id WHERE t1.status = 1;
```

### PostgreSQL 语法

#### 单表 UPDATE
```sql
-- 原始语句
UPDATE users SET name = 'test' WHERE id > 100;

-- 改写为
SELECT COUNT(1) FROM users WHERE id > 100;
```

#### 连表 UPDATE
```sql
-- 原始语句
UPDATE table1 
SET column1 = table2.column1 
FROM table2 
WHERE table1.id = table2.id;

-- 改写为
SELECT COUNT(1) FROM table1 
INNER JOIN table2 ON table1.id = table2.id;
```

#### 单表 DELETE
```sql
-- 原始语句
DELETE FROM users WHERE id > 100;

-- 改写为
SELECT COUNT(1) FROM users WHERE id > 100;
```

#### 连表 DELETE
```sql
-- 原始语句
DELETE FROM table1 USING table2 WHERE table1.id = table2.id;

-- 改写为
SELECT COUNT(1) FROM table1 INNER JOIN table2 ON table1.id = table2.id;
```

### SQL Server 语法

#### 单表 UPDATE
```sql
-- 原始语句
UPDATE users SET name = 'test' WHERE id > 100;

-- 改写为
SELECT COUNT(1) FROM users WHERE id > 100;
```

#### 连表 UPDATE
```sql
-- 原始语句
UPDATE t1 
SET t1.column1 = t2.column1 
FROM table1 t1 
INNER JOIN table2 t2 ON t1.id = t2.id 
WHERE t1.status = 1;

-- 改写为
SELECT COUNT(1) FROM table1 t1 
INNER JOIN table2 t2 ON t1.id = t2.id 
WHERE t1.status = 1;
```

#### 单表 DELETE
```sql
-- 原始语句
DELETE FROM users WHERE id > 100;

-- 改写为
SELECT COUNT(1) FROM users WHERE id > 100;
```

#### 连表 DELETE
```sql
-- 原始语句
DELETE t1 
FROM table1 t1 
INNER JOIN table2 t2 ON t1.id = t2.id 
WHERE t1.status = 1;

-- 改写为
SELECT COUNT(1) FROM table1 t1 
INNER JOIN table2 t2 ON t1.id = t2.id 
WHERE t1.status = 1;
```

### Oracle 语法

#### 单表 UPDATE
```sql
-- 原始语句
UPDATE users SET name = 'test' WHERE id > 100;

-- 改写为
SELECT COUNT(1) FROM users WHERE id > 100;
```

#### 单表 DELETE (带 FROM)
```sql
-- 原始语句
DELETE FROM users WHERE id > 100;

-- 改写为
SELECT COUNT(1) FROM users WHERE id > 100;
```

#### 单表 DELETE (不带 FROM)
```sql
-- 原始语句
DELETE users WHERE id > 100;

-- 改写为
SELECT COUNT(1) FROM users WHERE id > 100;
```

## 使用方法

### 基本命令

要启用影响行数计算功能，需要提供数据库连接参数：

```bash
# MySQL 示例
./advisor \
  -engine mysql \
  -sql "UPDATE users SET status = 1 WHERE created_at < '2024-01-01'" \
  -host localhost \
  -port 3306 \
  -user root \
  -password yourpassword \
  -dbname testdb \
  -format json

# PostgreSQL 示例
./advisor \
  -engine postgres \
  -sql "DELETE FROM logs WHERE created_at < NOW() - INTERVAL '30 days'" \
  -host localhost \
  -port 5432 \
  -user postgres \
  -password yourpassword \
  -dbname testdb \
  -sslmode disable \
  -format json

# SQL Server 示例
./advisor \
  -engine mssql \
  -sql "UPDATE orders SET status = 'completed' WHERE order_date < '2024-01-01'" \
  -host localhost \
  -port 1433 \
  -user sa \
  -password yourpassword \
  -dbname testdb \
  -format json

# Oracle 示例
./advisor \
  -engine oracle \
  -sql "DELETE FROM audit_logs WHERE log_date < SYSDATE - 90" \
  -host localhost \
  -port 1521 \
  -user system \
  -password yourpassword \
  -service-name ORCL \
  -format json
```

### 输出格式

当使用 JSON 输出格式时，`affected_rows` 字段会包含计算的影响行数：

```json
[
  {
    "order_id": 1,
    "stage": "CHECKED",
    "error_level": "0",
    "stage_status": "Audit Completed",
    "error_message": "",
    "sql": "UPDATE users SET status = 1 WHERE created_at < '2024-01-01'",
    "affected_rows": 1523,
    "sequence": "0_0_00000000",
    "backup_dbname": "",
    "execute_time": "0",
    "sqlsha1": "",
    "backup_time": "0"
  }
]
```

### 注意事项

1. **性能考虑**：对于大表，COUNT 查询可能需要较长时间。建议在使用前考虑查询的性能影响。

2. **估算精度**：返回的行数是基于当前数据库状态的估算值，实际执行 UPDATE/DELETE 时可能会因并发操作而略有差异。

3. **连接参数**：如果不提供数据库连接参数，`affected_rows` 将始终为 0。

4. **权限要求**：执行 COUNT 查询需要对相关表具有 SELECT 权限。

5. **WHERE 子句处理**：
   - 支持带 WHERE 子句和不带 WHERE 子句的语句
   - WHERE 子句中的条件会完整保留在 COUNT 查询中

6. **JOIN 支持**：
   - 支持 INNER JOIN、LEFT JOIN、RIGHT JOIN 等各种 JOIN 类型
   - JOIN 条件会正确转换到 COUNT 查询中

## 技术实现

### 改写方法

改写采用基于文本解析的方式，结合语法树验证：

1. 使用各数据库的 ANTLR 解析器验证 SQL 语法
2. 通过关键字定位识别 SQL 结构
3. 提取表名、JOIN 子句和 WHERE 条件
4. 重组为 SELECT COUNT(1) 查询

### 错误处理

- 如果 SQL 语法错误，改写会失败，`affected_rows` 为 0
- 如果数据库连接失败，`affected_rows` 为 0
- 如果 COUNT 查询执行失败，`affected_rows` 为 0
- 所有错误都不会中断审核流程，只是无法提供影响行数

## 测试

运行单元测试验证改写功能：

```bash
cd /data/dev_go/advisorTool
go test -v ./db -run TestRewrite
```

测试覆盖：
- ✅ MySQL 单表/连表 UPDATE/DELETE
- ✅ PostgreSQL 单表/连表 UPDATE/DELETE
- ✅ SQL Server 单表/连表 UPDATE/DELETE
- ✅ Oracle 单表 UPDATE/DELETE
- ✅ 带/不带 WHERE 子句的情况
- ✅ 各种 JOIN 类型

## 示例脚本

### 批量审核并计算影响行数

```bash
#!/bin/bash

# 配置数据库连接
DB_HOST="localhost"
DB_PORT="3306"
DB_USER="root"
DB_PASS="password"
DB_NAME="mydb"

# SQL 文件
SQL_FILE="migration.sql"

# 执行审核
./advisor \
  -engine mysql \
  -file "$SQL_FILE" \
  -host "$DB_HOST" \
  -port "$DB_PORT" \
  -user "$DB_USER" \
  -password "$DB_PASS" \
  -dbname "$DB_NAME" \
  -format json > results.json

# 解析结果
echo "审核结果："
jq '.[] | {sql: .sql, affected_rows: .affected_rows, error_level: .error_level}' results.json
```

## 未来改进

可能的改进方向：

1. 支持更多数据库引擎（如 Snowflake）
2. 支持子查询的改写
3. 提供执行计划分析
4. 支持更复杂的 JOIN 场景
5. 添加缓存机制提高性能

