# extractObject - SQL表名提取工具

这是一个独立的SQL表名提取工具，可以从SQL语句中提取表名、数据库名、模式名和别名信息。

## 功能特性

- 支持多种数据库类型：MySQL、PostgreSQL、Oracle、SQL Server、TiDB、MariaDB、OceanBase、Snowflake
- 提取完整的表信息：数据库名、模式名、表名、别名、是否为CTE临时表
- 简单易用的API接口和命令行工具
- 基于成熟的SQL解析器（来自Bytebase项目）
- 命令行参数支持小写格式，更符合Unix惯例

## 快速开始

### 命令行工具

```bash
# 编译工具
cd cmd
go build -o extractobject main.go

# 使用小写参数（推荐）
./extractobject -dbtype mysql -sql "SELECT * FROM users"
./extractobject -dbtype postgres -file query.sql
./extractobject -dbtype oracle -sql "SELECT * FROM hr.employees" -json
./extractobject -dbtype sqlserver -file query.sql

# 也支持大写参数（向后兼容）
./extractobject -dbtype MYSQL -sql "SELECT * FROM users"
```

### Go API 使用

```go
package main

import (
    "fmt"
    "log"
    
    extractor "github.com/tianyuso/advisorTool/extractObject"
)

func main() {
    // 定义SQL语句
    sql := `
        SELECT u.id, u.name, o.order_id
        FROM mydb.users AS u
        JOIN orders o ON u.id = o.user_id
        WHERE u.status = 'active'
    `
    
    // 提取表名
    tables, err := extractor.ExtractTables(extractor.MySQL, sql)
    if err != nil {
        log.Fatal(err)
    }
    
    // 打印结果
    for _, table := range tables {
        fmt.Printf("数据库: %s, 模式: %s, 表名: %s, 别名: %s\n",
            table.DBName, table.Schema, table.TBName, table.Alias)
    }
}
```

### 输出示例

```
数据库: mydb, 模式: , 表名: users, 别名: u
数据库: , 模式: , 表名: orders, 别名: o
```

## 命令行工具

### 安装

```bash
cd cmd
go build -o extractobject main.go
```

### 使用方法

```bash
# 基本用法
./extractobject -dbtype <数据库类型> -sql "<SQL语句>"
./extractobject -dbtype <数据库类型> -file <SQL文件路径>

# JSON输出
./extractobject -dbtype <数据库类型> -sql "<SQL语句>" -json

# 查看版本
./extractobject -version
```

### 支持的数据库类型

| 数据库 | 参数（推荐） | 别名支持 |
|--------|------------|----------|
| MySQL | `mysql` | MYSQL, MySQL |
| PostgreSQL | `postgres` | postgresql, POSTGRES, POSTGRESQL |
| Oracle | `oracle` | ORACLE, Oracle |
| SQL Server | `sqlserver` | SQLSERVER, SQLServer, mssql, MSSQL |
| TiDB | `tidb` | TIDB, TiDB |
| MariaDB | `mariadb` | MARIADB, MariaDB |
| OceanBase | `oceanbase` | OCEANBASE, OceanBase |

### 命令行示例

```bash
# MySQL
./extractobject -dbtype mysql -sql "SELECT u.id FROM mydb.users u"

# PostgreSQL
./extractobject -dbtype postgres -file query.sql

# Oracle - JSON输出
./extractobject -dbtype oracle -sql "SELECT * FROM hr.employees" -json

# SQL Server
./extractobject -dbtype sqlserver -file complex_query.sql
```

## API说明

### ExtractTables

```go
func ExtractTables(dbType DBType, sql string) ([]TableInfo, error)
```

从SQL语句中提取表名信息。

**参数：**
- `dbType`: 数据库类型，可选值：
  - `MySQL`
  - `PostgreSQL`
  - `Oracle`
  - `SQLServer`
  - `TiDB`
  - `MariaDB`
  - `OceanBase`
  - `Snowflake`
- `sql`: 要解析的SQL语句

**返回：**
- `[]TableInfo`: 表信息列表
- `error`: 错误信息

### ExtractTablesWithContext

```go
func ExtractTablesWithContext(ctx context.Context, dbType DBType, sql string) ([]TableInfo, error)
```

带上下文的表名提取函数，支持超时控制和取消操作。

### TableInfo 结构

```go
type TableInfo struct {
    DBName string // 数据库名
    Schema string // 模式名
    TBName string // 表名
    Alias  string // 别名
    IsCTE  bool   // 是否是CTE临时表
}
```

## 支持的数据库类型

### Go API 常量

在 Go 代码中使用以下常量：

```go
extractor.MySQL       // MySQL / MariaDB / OceanBase
extractor.PostgreSQL  // PostgreSQL
extractor.Oracle      // Oracle
extractor.SQLServer   // SQL Server
extractor.TiDB        // TiDB
extractor.MariaDB     // MariaDB
extractor.OceanBase   // OceanBase
```

### 字符串解析

如果需要从字符串解析数据库类型，使用 `ParseDBType` 函数：

```go
dbType, err := extractor.ParseDBType("mysql")  // 支持大小写不敏感
if err != nil {
    log.Fatal(err)
}
tables, err := extractor.ExtractTables(dbType, sql)
```

## 支持的SQL语句类型

- SELECT查询
- INSERT语句
- UPDATE语句
- DELETE语句
- JOIN操作
- 子查询
- CTE（公用表表达式）

## 使用示例

### MySQL示例

```go
sql := `
    SELECT u.id, u.name, o.order_id
    FROM mydb.users AS u
    JOIN orders o ON u.id = o.user_id
`
tables, _ := extractor.ExtractTables(extractor.MySQL, sql)
// 输出: mydb.users (u), orders (o)
```

### PostgreSQL示例

```go
sql := `
    SELECT p.product_name, c.category_name
    FROM public.products p
    INNER JOIN public.categories c ON p.category_id = c.id
`
tables, _ := extractor.ExtractTables(extractor.PostgreSQL, sql)
// 输出: public.products (p), public.categories (c)
```

### SQL Server示例

```go
sql := `
    SELECT e.employee_id, d.department_name
    FROM HRDatabase.dbo.employees AS e
    LEFT JOIN HRDatabase.dbo.departments d ON e.dept_id = d.id
`
tables, _ := extractor.ExtractTables(extractor.SQLServer, sql)
// 输出: HRDatabase.dbo.employees (e), HRDatabase.dbo.departments (d)
```

### Oracle示例

```go
sql := `
    SELECT e.emp_id, d.dept_name
    FROM hr.employees e, hr.departments d
    WHERE e.dept_id = d.dept_id
`
tables, _ := extractor.ExtractTables(extractor.Oracle, sql)
// 输出: hr.employees (e), hr.departments (d)
```

## 注意事项

1. 不同数据库的完全限定表名格式不同：
   - MySQL: `database.table`
   - PostgreSQL: `schema.table` 或 `database.schema.table`
   - SQL Server: `database.schema.table` 或 `server.database.schema.table`
   - Oracle: `schema.table`

2. 如果表名中没有指定数据库或模式，相应字段将为空字符串

3. 别名是可选的，如果SQL中没有指定别名，Alias字段将为空字符串

## 依赖

本工具依赖以下解析库：
- MySQL/MariaDB/OceanBase: `github.com/bytebase/parser/mysql`
- PostgreSQL: `github.com/bytebase/parser/postgresql`
- Oracle: `github.com/bytebase/parser/plsql`
- SQL Server: `github.com/bytebase/parser/tsql`
- TiDB: `github.com/pingcap/tidb/parser` (通过MySQL解析器)
- Snowflake: `github.com/bytebase/parser/snowflake`

## 许可证

本项目基于Bytebase SQL解析引擎，继承其开源许可证。


