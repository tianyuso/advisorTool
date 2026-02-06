# extractObject SQL表名提取工具 - 快速开始指南

## 项目概述

`extractObject` 是一个简化版的SQL表名提取工具，基于 [Bytebase](https://github.com/bytebase/bytebase) SQL解析引擎改造。它可以从SQL语句中提取表名、数据库名、模式名和别名信息，支持多种数据库类型。

## 支持的数据库

- ✅ MySQL
- ✅ PostgreSQL
- ✅ SQL Server
- ✅ TiDB
- ✅ MariaDB
- ✅ OceanBase
- ✅ Snowflake
- ⚠️  Oracle (需要parser注册支持)

## 快速开始

### 1. 作为库使用

```go
package main

import (
    "fmt"
    "log"
    
    extractor "github.com/tianyuso/advisorTool/extractObject"
)

func main() {
    sql := `
        SELECT u.id, u.name, o.order_id
        FROM mydb.users AS u
        JOIN orders o ON u.id = o.user_id
    `
    
    tables, err := extractor.ExtractTables(extractor.MySQL, sql)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, table := range tables {
        fmt.Printf("数据库: %s, 模式: %s, 表名: %s, 别名: %s\n",
            table.DBName, table.Schema, table.TBName, table.Alias)
    }
}
```

**输出:**
```
数据库: mydb, 模式: , 表名: users, 别名: u
数据库: , 模式: , 表名: orders, 别名: o
```

### 2. 作为命令行工具使用

#### 编译

```bash
cd extractObject/cmd
go build -o extractobject main.go
```

#### 使用示例

**从命令行提取:**
```bash
./extractobject -db MYSQL -sql "SELECT * FROM users"
```

**从文件提取:**
```bash
./extractobject -db POSTGRESQL -file query.sql
```

**JSON格式输出:**
```bash
./extractobject -db MYSQL -sql "SELECT * FROM mydb.users u" -json
```

输出:
```json
[
  {
    "DBName": "mydb",
    "Schema": "",
    "TBName": "users",
    "Alias": "u"
  }
]
```

## API 文档

### 核心函数

#### ExtractTables

```go
func ExtractTables(dbType DBType, sql string) ([]TableInfo, error)
```

从SQL语句中提取表名信息。

**参数:**
- `dbType`: 数据库类型 (`MySQL`, `PostgreSQL`, `Oracle`, `SQLServer`, `TiDB`, `MariaDB`, `OceanBase`, `Snowflake`)
- `sql`: SQL语句

**返回:**
- `[]TableInfo`: 表信息列表
- `error`: 错误信息

#### ExtractTablesWithContext

```go
func ExtractTablesWithContext(ctx context.Context, dbType DBType, sql string) ([]TableInfo, error)
```

带上下文的表名提取，支持超时和取消。

### 数据结构

#### TableInfo

```go
type TableInfo struct {
    DBName string // 数据库名
    Schema string // 模式名
    TBName string // 表名
    Alias  string // 别名
}
```

## 使用示例

### MySQL 示例

```go
sql := `
    SELECT u.id, u.name, o.order_id
    FROM mydb.users AS u
    JOIN orders o ON u.id = o.user_id
`
tables, _ := extractor.ExtractTables(extractor.MySQL, sql)
// 输出: mydb.users (u), orders (o)
```

### PostgreSQL 示例

```go
sql := `
    SELECT p.product_name, c.category_name
    FROM public.products p
    INNER JOIN public.categories c ON p.category_id = c.id
`
tables, _ := extractor.ExtractTables(extractor.PostgreSQL, sql)
// 输出: public.products (p), public.categories (c)
```

### SQL Server 示例

```go
sql := `
    SELECT e.employee_id, d.department_name
    FROM HRDatabase.dbo.employees AS e
    LEFT JOIN HRDatabase.dbo.departments d ON e.dept_id = d.id
`
tables, _ := extractor.ExtractTables(extractor.SQLServer, sql)
// 输出: HRDatabase.dbo.employees (e), HRDatabase.dbo.departments (d)
```

## 运行测试

```bash
cd extractObject
go test -v
```

## 项目结构

```
extractObject/
├── types.go              # 类型定义
├── extractor.go          # 核心提取逻辑
├── mysql_extractor.go    # MySQL提取器
├── postgresql_extractor.go # PostgreSQL提取器
├── sqlserver_extractor.go  # SQL Server提取器
├── oracle_extractor.go   # Oracle提取器
├── tidb_extractor.go     # TiDB提取器
├── snowflake_extractor.go # Snowflake提取器
├── extractor_test.go     # 单元测试
├── example.go            # 使用示例
├── README.md             # 详细文档
├── QUICKSTART.md         # 本文档
├── cmd/                  # 命令行工具
│   ├── main.go
│   └── README.md
└── examples/             # 示例代码
    └── demo.go
```

## 注意事项

1. **完全限定表名格式** 因数据库而异:
   - MySQL: `database.table`
   - PostgreSQL: `schema.table` 或 `database.schema.table`
   - SQL Server: `database.schema.table`
   - Oracle: `schema.table`

2. **空字段**: 如果SQL中未指定数据库或模式，相应字段将为空字符串

3. **别名**: 目前MySQL支持提取别名，其他数据库正在完善中

## 性能说明

- 解析速度取决于SQL复杂度
- 简单SELECT: < 10ms
- 复杂JOIN查询: < 50ms
- 多语句: 按语句数量线性增长

## 常见问题

### Q: 为什么Oracle测试失败？
A: Oracle解析器需要在parser中注册。如果需要使用Oracle，请检查`parser/plsql`包的注册逻辑。

### Q: 支持子查询吗？
A: 是的，支持子查询、CTE和各种复杂SQL语句。

### Q: 可以提取视图名吗？
A: 可以，视图在SQL中作为表引用，会被正常提取。

### Q: 如何处理语法错误的SQL?
A: 工具会返回错误，建议先验证SQL语法。

## 许可证

本项目基于Bytebase SQL解析引擎，继承其开源许可证。

## 贡献

欢迎提交Issues和Pull Requests!

## 联系方式

- GitHub: [advisorTool](https://github.com/tianyuso/advisorTool)
- 问题反馈: 请通过GitHub Issues提交

---

**版本**: v1.0.0  
**最后更新**: 2026-02-04
