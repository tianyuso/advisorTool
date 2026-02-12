# 数据库类型参数更新说明

## 更新内容

将数据库类型参数从大写改为小写格式，提升用户体验和命令行工具的易用性。

## 变更对比

### 之前（大写格式）
```bash
./extractobject -dbtype MYSQL -sql "SELECT * FROM users"
./extractobject -dbtype POSTGRESQL -sql "SELECT * FROM users"
./extractobject -dbtype ORACLE -sql "SELECT * FROM users"
./extractobject -dbtype SQLSERVER -sql "SELECT * FROM users"
```

### 现在（小写格式，推荐）
```bash
./extractobject -dbtype mysql -sql "SELECT * FROM users"
./extractobject -dbtype postgres -sql "SELECT * FROM users"
./extractobject -dbtype oracle -sql "SELECT * FROM users"
./extractobject -dbtype sqlserver -sql "SELECT * FROM users"
```

## 支持的数据库类型

| 数据库 | 参数（推荐小写） | 别名支持 |
|--------|----------------|----------|
| MySQL | `mysql` | MYSQL, MySQL |
| PostgreSQL | `postgres` | postgresql, POSTGRES, POSTGRESQL, PostgreSQL |
| Oracle | `oracle` | ORACLE, Oracle |
| SQL Server | `sqlserver` | SQLSERVER, SQLServer, mssql, MSSQL |
| TiDB | `tidb` | TIDB, TiDB |
| MariaDB | `mariadb` | MARIADB, MariaDB |
| OceanBase | `oceanbase` | OCEANBASE, OceanBase |
| Snowflake | `snowflake` | SNOWFLAKE, Snowflake |

## 向后兼容性

✅ **完全向后兼容**：旧的大写参数仍然可以正常使用，不会影响现有脚本和代码。

```bash
# 以下两种方式都可以正常工作
./extractobject -dbtype mysql -sql "SELECT * FROM users"
./extractobject -dbtype MYSQL -sql "SELECT * FROM users"
```

## 代码使用

### Go 代码中使用
```go
import extractor "github.com/tianyuso/advisorTool/extractObject"

// 使用常量（推荐）
tables, err := extractor.ExtractTables(extractor.MySQL, sql)

// 使用字符串（需要解析）
dbType, err := extractor.ParseDBType("mysql")
if err != nil {
    log.Fatal(err)
}
tables, err := extractor.ExtractTables(dbType, sql)
```

### 命令行使用
```bash
# 推荐使用小写
./extractobject -dbtype mysql -file query.sql
./extractobject -dbtype postgres -sql "SELECT * FROM users"
./extractobject -dbtype oracle -file query.sql -json
./extractobject -dbtype sqlserver -file query.sql

# 也支持大写（向后兼容）
./extractobject -dbtype MYSQL -file query.sql
./extractobject -dbtype POSTGRESQL -file query.sql
```

## 更新的文件列表

### 核心代码
- ✅ `types.go` - 数据库类型常量定义，新增 ParseDBType 函数
- ✅ `cmd/main.go` - 命令行工具，更新帮助文本和默认值

### Shell 脚本
- ✅ `cmd/demo_cte_feature.sh` - CTE 功能演示脚本
- ✅ `cmd/demo_cte_all_databases.sh` - 全数据库 CTE 演示
- ✅ `cmd/test_mysql.sh` - MySQL 测试脚本
- ✅ `final_demo.sh` - 最终演示脚本
- ✅ `test.sh` - 基础测试脚本

### 测试文件
- ✅ `extractor_test.go` - 单元测试（使用常量，无需修改）
- ✅ `example.go` - 示例代码（使用常量，无需修改）

## 测试验证

### 单元测试
```bash
cd /data/dev_go/advisorTool/extractObject
go test -v -run "TestMySQL|TestPostgreSQL|TestSQLServer|TestOracle"
```
✅ 所有测试通过

### 功能测试
```bash
cd /data/dev_go/advisorTool/extractObject/cmd

# 测试小写参数
./extractobject -db mysql -sql "SELECT * FROM users"
./extractobject -db postgres -sql "SELECT * FROM users"
./extractobject -db oracle -sql "SELECT * FROM users"
./extractobject -db sqlserver -sql "SELECT * FROM users"

# 测试大写参数（向后兼容）
./extractobject -db MYSQL -sql "SELECT * FROM users"
./extractobject -db POSTGRESQL -sql "SELECT * FROM users"
```
✅ 所有测试通过

## 迁移指南

### 对于命令行用户
- 推荐开始使用小写参数（如 `mysql`, `postgres`）
- 旧的大写参数仍然可用，无需立即修改现有脚本

### 对于开发者
- 在 Go 代码中继续使用常量（`extractor.MySQL`, `extractor.PostgreSQL` 等）
- 如果需要解析字符串为数据库类型，使用 `extractor.ParseDBType()` 函数
- 该函数支持大小写不敏感的解析

## 优势

1. **更符合命令行惯例**：大多数命令行工具使用小写参数
2. **更易于输入**：小写输入更快，无需按 Shift 键
3. **向后兼容**：不会破坏现有的脚本和代码
4. **灵活性强**：支持多种格式，包括缩写（如 `mssql` for SQL Server）

## 更新日期

2026-02-06

