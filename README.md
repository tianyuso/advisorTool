# SQL Advisor Tool

一个基于 [Bytebase](https://github.com/bytebase/bytebase) SQL 审核引擎的独立命令行工具。完整保留 Bytebase 原有的 SQL 解析器和审核规则实现，支持 MySQL、PostgreSQL、Oracle、SQL Server 等多种数据库。

## 特性

- 🔍 **多数据库支持**: MySQL, MariaDB, PostgreSQL, Oracle, SQL Server, TiDB, Snowflake, OceanBase
- 📋 **完整的审核规则**: 70+ 种内置规则，覆盖命名规范、语句规范、表设计、索引优化等
- 🛠️ **原生解析器**: 使用 Bytebase 原有的 ANTLR 解析器，保证解析准确性
  - MySQL: `github.com/bytebase/parser/mysql`
  - PostgreSQL: `github.com/bytebase/parser/postgresql`  
  - Oracle: `github.com/bytebase/parser/plsql`
  - SQL Server: `github.com/bytebase/parser/tsql`
- ⚙️ **高度可配置**: 通过 YAML/JSON 配置文件自定义规则和级别
- 📊 **多种输出格式**: 文本、JSON、YAML

## 安装

### 前置要求

- Go 1.23 或更高版本
- Bytebase 源码（本工具是 Bytebase backend 的子模块）

### 从源码构建

本工具现在可以独立编译，只需要确保 Bytebase 源码在正确位置即可：

```bash
# 进入 advisorTool 目录
cd /path/to/advisorTool

# 编译
go build -o build/advisor ./cmd/advisor

# 或者使用 make（如果有）
make build
```

**注意**：首次编译时会下载较多依赖，请耐心等待。

## 快速开始

### 命令行使用

```bash
# 审核 SQL 语句
./advisor -engine mysql -sql "SELECT * FROM users"

# 审核 SQL 文件
./advisor -engine postgres -file schema.sql

# 从标准输入读取 SQL
cat schema.sql | ./advisor -engine mysql -sql -

# 使用自定义配置文件
./advisor -engine mysql -config review-config.yaml -file schema.sql

# 输出 JSON 格式
./advisor -engine mysql -sql "SELECT * FROM users" -format json

# 列出所有可用规则
./advisor -list-rules

# 生成示例配置文件
./advisor -engine mysql -generate-config > mysql-config.yaml
```

### 命令行参数

| 参数 | 说明 |
|------|------|
| `-engine` | 数据库类型（必需）: mysql, postgres, tidb, oracle, mssql, snowflake, mariadb, oceanbase |
| `-sql` | SQL 语句（使用 `-` 从标准输入读取） |
| `-file` | SQL 文件路径 |
| `-config` | 审核配置文件路径（YAML 或 JSON） |
| `-format` | 输出格式: text, json, yaml（默认: text） |
| `-list-rules` | 列出所有可用规则 |
| `-generate-config` | 生成指定数据库的示例配置文件 |
| `-version` | 显示版本信息 |

### 退出码

- `0`: 审核通过，没有问题
- `1`: 发现警告级别的问题
- `2`: 发现错误级别的问题

### 作为 Go 库使用

```go
package main

import (
    "context"
    "fmt"
    
    "advisorTool/pkg/advisor"
)

func main() {
    // 定义审核规则
    rules := []*advisor.SQLReviewRule{
        advisor.NewRule(advisor.RuleStatementNoSelectAll, advisor.RuleLevelWarning),
        advisor.NewRule(advisor.RuleStatementRequireWhereForUpdateDelete, advisor.RuleLevelError),
        advisor.NewRule(advisor.RuleTableRequirePK, advisor.RuleLevelError),
    }
    
    // 创建审核请求
    req := &advisor.ReviewRequest{
        Engine:    advisor.EngineMySQL,
        Statement: "SELECT * FROM users; DELETE FROM orders;",
        Rules:     rules,
    }
    
    // 执行审核
    resp, err := advisor.SQLReviewCheck(context.Background(), req)
    if err != nil {
        panic(err)
    }
    
    // 处理结果
    for _, advice := range resp.Advices {
        fmt.Printf("[%s] %s: %s\n", advice.Status, advice.Title, advice.Content)
    }
    
    if resp.HasError {
        fmt.Println("审核发现错误!")
    }
}
```

## 配置文件格式

### YAML 格式

```yaml
name: mysql-review-config
rules:
  - type: statement.select.no-select-all
    level: WARNING
    comment: 禁止使用 SELECT *
    
  - type: statement.where.require.update-delete
    level: ERROR
    comment: UPDATE/DELETE 必须包含 WHERE 子句
    
  - type: table.require-pk
    level: ERROR
    comment: 表必须有主键
    
  - type: naming.table
    level: WARNING
    payload: '{"format":"^[a-z][a-z0-9_]*$","maxLength":64}'
    comment: 表名必须使用小写字母和下划线
    
  - type: column.required
    level: ERROR
    payload: '{"list":["id","created_at","updated_at"]}'
    comment: 每个表必须包含指定列
```

### 规则级别

| 级别 | 说明 |
|------|------|
| `ERROR` | 错误级别，必须修复 |
| `WARNING` | 警告级别，建议修复 |
| `DISABLED` | 禁用此规则 |

## 支持的解析器

本工具使用 Bytebase 原有的解析器，基于 ANTLR4：

| 数据库 | 解析器包 |
|--------|----------|
| MySQL | `github.com/bytebase/parser/mysql` |
| MariaDB | `github.com/bytebase/parser/mysql` |
| PostgreSQL | `github.com/bytebase/parser/postgresql` |
| Oracle | `github.com/bytebase/parser/plsql` |
| SQL Server | `github.com/bytebase/parser/tsql` |
| TiDB | `github.com/pingcap/tidb/parser` |
| Snowflake | `github.com/bytebase/parser/snowflake` |
| OceanBase | `github.com/bytebase/parser/mysql` |

## 支持的审核规则

### Engine 规则
- `engine.mysql.use-innodb` - 要求使用 InnoDB 存储引擎

### Naming 命名规则
- `naming.fully-qualified` - 要求使用完全限定的对象名
- `naming.table` - 表命名规范
- `naming.column` - 列命名规范
- `naming.index.pk` - 主键命名规范
- `naming.index.uk` - 唯一键命名规范
- `naming.index.fk` - 外键命名规范
- `naming.index.idx` - 索引命名规范
- `naming.column.auto-increment` - 自增列命名规范
- `naming.table.no-keyword` - 禁止使用关键字作为表名
- `naming.identifier.no-keyword` - 禁止使用关键字作为标识符
- `naming.identifier.case` - 标识符大小写规范

### Statement 语句规则
- `statement.select.no-select-all` - 禁止使用 SELECT *
- `statement.where.require.select` - SELECT 必须包含 WHERE
- `statement.where.require.update-delete` - UPDATE/DELETE 必须包含 WHERE
- `statement.where.no-leading-wildcard-like` - 禁止前导通配符 LIKE
- `statement.disallow-commit` - 禁止 COMMIT 语句
- `statement.disallow-limit` - 禁止 LIMIT 子句
- `statement.disallow-order-by` - 禁止 ORDER BY 子句
- `statement.merge-alter-table` - 合并 ALTER TABLE 语句
- `statement.insert.must-specify-column` - INSERT 必须指定列名
- `statement.insert.disallow-order-by-rand` - 禁止 ORDER BY RAND
- `statement.insert.row-limit` - INSERT 行数限制
- `statement.affected-row-limit` - 影响行数限制
- `statement.dml-dry-run` - DML 空运行验证
- `statement.disallow-add-column-with-default` - 禁止 ADD COLUMN 带默认值（PostgreSQL）
- `statement.add-check-not-valid` - CHECK 约束必须 NOT VALID（PostgreSQL）
- `statement.disallow-add-not-null` - 禁止添加 NOT NULL（PostgreSQL）
- `statement.disallow-cross-db-queries` - 禁止跨库查询（MSSQL）

### Table 表规则
- `table.require-pk` - 表必须有主键
- `table.no-foreign-key` - 禁止外键
- `table.drop-naming-convention` - 删除表命名规范
- `table.comment` - 表注释规范
- `table.disallow-partition` - 禁止分区表
- `table.disallow-trigger` - 禁止触发器
- `table.no-duplicate-index` - 禁止重复索引
- `table.disallow-ddl` - 禁止 DDL 操作
- `table.disallow-dml` - 禁止 DML 操作
- `table.limit-size` - 限制表大小

### Column 列规则
- `column.required` - 必需列
- `column.no-null` - 禁止 NULL 值
- `column.disallow-change-type` - 禁止改变列类型
- `column.set-default-for-not-null` - NOT NULL 列需要默认值
- `column.disallow-change` - 禁止 CHANGE COLUMN
- `column.disallow-changing-order` - 禁止改变列顺序
- `column.auto-increment-must-integer` - 自增列必须为整数
- `column.type-disallow-list` - 列类型黑名单
- `column.disallow-set-charset` - 禁止设置字符集
- `column.auto-increment-must-unsigned` - 自增列必须无符号
- `column.comment` - 列注释规范
- `column.maximum-character-length` - CHAR 最大长度
- `column.maximum-varchar-length` - VARCHAR 最大长度
- `column.require-default` - 列必须有默认值
- `column.disallow-drop-in-index` - 禁止删除索引列

### Index 索引规则
- `index.no-duplicate-column` - 禁止重复列
- `index.key-number-limit` - 索引键数量限制
- `index.pk-type-limit` - 主键类型限制
- `index.type-no-blob` - 禁止 BLOB/TEXT 索引
- `index.total-number-limit` - 索引总数限制
- `index.primary-key-type-allowlist` - 主键类型白名单
- `index.create-concurrently` - 并发创建索引（PostgreSQL）
- `index.not-redundant` - 禁止冗余索引

### Schema 模式规则
- `schema.backward-compatibility` - 向后兼容性检查

### System 系统规则
- `system.charset.allowlist` - 字符集白名单
- `system.collation.allowlist` - 排序规则白名单
- `system.comment.length` - 注释长度限制
- `system.procedure.disallow-create` - 禁止创建存储过程
- `system.function.disallow-create` - 禁止创建函数

## 项目结构

```
advisorTool/
├── build/
│   └── advisor                      # 编译输出
├── cmd/
│   └── advisor/
│       └── main.go                  # 命令行入口
├── pkg/
│   └── advisor/
│       ├── advisor.go               # 核心封装层（引用 Bytebase advisor）
│       └── rules.go                 # 规则常量定义
├── examples/
│   ├── mysql-review-config.yaml     # MySQL 配置示例
│   ├── postgres-review-config.yaml  # PostgreSQL 配置示例
│   ├── basic-config.yaml            # 基础配置示例
│   └── test.sql                     # 测试 SQL
├── go.mod                           # Go 模块定义（含 replace 指令）
├── go.sum                           # 依赖校验
├── Makefile                         # 编译脚本
└── README.md                        # 使用说明
```

## ⚠️ 使用注意事项

### 规则分类

根据是否需要数据库元数据，规则可分为两类：

**1. 无需数据库连接的规则（静态分析）：**
- 命名规范规则（naming.table, naming.column 等）
- 基础语句规则（statement.select.no-select-all, statement.where.require.* 等）
- 表结构规则（table.require-pk, table.no-foreign-key 等）
- 大部分语法检查规则

**2. 需要数据库元数据的规则（需谨慎使用）：**
- `column.no-null` - 需要现有表的元数据
- `column.disallow-drop-in-index` - 需要索引信息
- `schema.backward-compatibility` - 需要完整的 schema 信息
- `table.limit-size` - 需要表大小信息
- DML 空运行规则

当使用需要数据库元数据的规则但未提供元数据时，可能会报错。建议在独立使用时仅启用静态分析规则。

### 推荐的基础配置

参见 `examples/basic-config.yaml`，仅包含不需要数据库元数据的规则。

## 依赖说明

本工具有独立的 `go.mod` 文件，使用 `replace` 指令引用本地 Bytebase 代码：

```go
// go.mod
replace github.com/bytebase/bytebase => ../..

// 以及从主项目复制的其他 replace 指令（ANTLR、TiDB Parser 等）
```

这种设计的优点：
1. **独立编译**：可以直接在 advisorTool 目录下运行 `go build`
2. **依赖一致性**：通过 replace 指令确保依赖版本与主项目一致
3. **完整功能**：使用 Bytebase 原有的 SQL 解析器和全部审核规则
4. **易于维护**：当主项目更新时，只需同步 go.mod 中的 replace 指令

## 与 Bytebase 的关系

本工具是 Bytebase SQL 审核引擎的命令行封装。Bytebase 是一个开源的数据库 DevOps 平台，提供数据库 CI/CD、变更管理、SQL 审核等功能。

如果你需要更完整的数据库管理功能（Web UI、工作流、权限管理等），建议使用完整的 Bytebase 平台。

## 许可证

遵循 Bytebase 项目的许可证。

## 相关链接

- [Bytebase 官网](https://www.bytebase.com)
- [Bytebase GitHub](https://github.com/bytebase/bytebase)
- [SQL Review 文档](https://www.bytebase.com/docs/sql-review/overview)
- [审核规则文档](https://www.bytebase.com/docs/sql-review/review-rules)
