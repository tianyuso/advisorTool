# 外部项目使用指南

## 问题说明

当您在外部项目中使用 `github.com/tianyuso/advisorTool` 时，可能会遇到依赖版本冲突的问题。这是因为本项目依赖于 Bytebase 的一些 fork 版本的库。

## 解决方案

在您的项目的 `go.mod` 文件中添加以下 `replace` 指令：

```go
replace (
    github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos => github.com/bytebase/azure-sdk-for-go/sdk/data/azcosmos v0.0.0-20250109032656-87cf24d45689
    
    github.com/antlr4-go/antlr/v4 => github.com/bytebase/antlr/v4 v4.0.0-20240827034948-8c385f108920
    
    github.com/beltran/gohive => github.com/bytebase/gohive v0.0.0-20240422092929-d76993a958a4
    github.com/beltran/gosasl => github.com/bytebase/gosasl v0.0.0-20240422091407-6b7481e86f08
    
    github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v3.2.6-0.20210809144907-32ab6a8243d7+incompatible
    
    github.com/github/gh-ost => github.com/bytebase/gh-ost2 v1.1.7-0.20251002210738-35e5dddaad7c
    
    github.com/jackc/pgx/v5 => github.com/bytebase/pgx/v5 v5.0.0-20250212161523-96ff8aed8767
    
    github.com/mattn/go-oci8 => github.com/bytebase/go-obo v0.0.0-20231026081615-705a7fffbfd2
    
    github.com/microsoft/go-mssqldb => github.com/bytebase/go-mssqldb v0.0.0-20240801091126-3ff3ca07d898
    
    github.com/pingcap/tidb => github.com/bytebase/tidb v0.0.0-20251104040057-d29df9dd1b3b
    
    github.com/pingcap/tidb/pkg/parser => github.com/bytebase/tidb/pkg/parser v0.0.0-20251104040057-d29df9dd1b3b
    
    github.com/youmark/pkcs8 => github.com/bytebase/pkcs8 v0.0.0-20240612095628-fcd0a7484c94
)
```

## 完整示例

创建一个新项目：

```bash
mkdir myproject
cd myproject
go mod init myproject
```

编辑 `go.mod`，添加 replace 指令：

```go
module myproject

go 1.24

replace (
    github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos => github.com/bytebase/azure-sdk-for-go/sdk/data/azcosmos v0.0.0-20250109032656-87cf24d45689
    github.com/antlr4-go/antlr/v4 => github.com/bytebase/antlr/v4 v4.0.0-20240827034948-8c385f108920
    github.com/beltran/gohive => github.com/bytebase/gohive v0.0.0-20240422092929-d76993a958a4
    github.com/beltran/gosasl => github.com/bytebase/gosasl v0.0.0-20240422091407-6b7481e86f08
    github.com/dgrijalva/jwt-go => github.com/form3tech-oss/jwt-go v3.2.6-0.20210809144907-32ab6a8243d7+incompatible
    github.com/github/gh-ost => github.com/bytebase/gh-ost2 v1.1.7-0.20251002210738-35e5dddaad7c
    github.com/jackc/pgx/v5 => github.com/bytebase/pgx/v5 v5.0.0-20250212161523-96ff8aed8767
    github.com/mattn/go-oci8 => github.com/bytebase/go-obo v0.0.0-20231026081615-705a7fffbfd2
    github.com/microsoft/go-mssqldb => github.com/bytebase/go-mssqldb v0.0.0-20240801091126-3ff3ca07d898
    github.com/pingcap/tidb => github.com/bytebase/tidb v0.0.0-20251104040057-d29df9dd1b3b
    github.com/pingcap/tidb/pkg/parser => github.com/bytebase/tidb/pkg/parser v0.0.0-20251104040057-d29df9dd1b3b
    github.com/youmark/pkcs8 => github.com/bytebase/pkcs8 v0.0.0-20240612095628-fcd0a7484c94
)

require github.com/tianyuso/advisorTool v1.0.6
```

然后运行：

```bash
go mod tidy
go build
```

## 为什么需要这些 replace 指令？

advisorTool 基于 Bytebase 项目，使用了 Bytebase 团队维护的一些第三方库的 fork 版本，这些版本包含了特定的修复和增强功能。由于 Go 模块系统的限制，`replace` 指令不会传递到依赖项目，因此您需要在自己的项目中手动添加这些指令。

## 支持的数据库

- MySQL
- PostgreSQL
- SQL Server (MSSQL)
- Oracle
- TiDB
- OceanBase
- Snowflake
- Redshift

## 参考示例

请参考 `examples/` 目录下的示例代码：

- `postgres_external_usage_example.go` - PostgreSQL 审核示例
- `library_usage_basic.go` - 基本使用示例
- `library_usage_all_rules.go` - 所有规则示例
- `library_usage_with_payload.go` - 带有元数据的审核示例

