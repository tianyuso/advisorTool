# advisorTool 外部使用指南

## ✅ 最新修复 (v1.0.6)

已修复外部项目使用时的依赖版本问题。现在可以正常使用 `go get` 安装。

## 快速开始

### 1. 安装

```bash
go get github.com/tianyuso/advisorTool@latest
```

### 2. 最简单的使用示例

创建 `main.go`：

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/tianyuso/advisorTool/pkg/advisor"
    "github.com/tianyuso/advisorTool/services"
)

func main() {
    // 1. 加载规则
    rules, err := services.LoadRules("", advisor.EnginePostgres, false)
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 准备 SQL
    sql := `CREATE TABLE users (id INT PRIMARY KEY);`
    
    // 3. 审核
    req := &advisor.ReviewRequest{
        Engine:          advisor.EnginePostgres,
        Statement:       sql,
        CurrentDatabase: "mydb",
        Rules:           rules,
    }
    
    resp, err := advisor.SQLReviewCheck(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. 查看结果
    fmt.Printf("发现 %d 个问题\n", len(resp.Advices))
}
```

### 3. 运行

```bash
go mod init myproject
go mod tidy
go run main.go
```

## 完整示例（带数据库连接）

如果需要连接数据库进行高级审核和影响行数计算：

```go
package main

import (
    "context"
    "log"
    
    "github.com/tianyuso/advisorTool/pkg/advisor"
    "github.com/tianyuso/advisorTool/services"
)

func main() {
    // 配置数据库连接
    dbParams := &services.DBConnectionParams{
        Host:     "127.0.0.1",
        Port:     5432,
        User:     "postgres",
        Password: "secret",
        DbName:   "mydb",
        SSLMode:  "disable",
        Timeout:  10,
    }
    
    engineType := advisor.EnginePostgres
    
    // 获取元数据（用于高级规则）
    metadata, err := services.FetchDatabaseMetadata(engineType, dbParams)
    if err != nil {
        log.Printf("警告: %v", err)
        metadata = nil
    }
    
    // 加载规则
    hasMetadata := (metadata != nil)
    rules, err := services.LoadRules("", engineType, hasMetadata)
    if err != nil {
        log.Fatal(err)
    }
    
    // 审核 SQL
    sql := `
    UPDATE myschema.users SET status = 'active';
    DELETE FROM myschema.orders WHERE id > 1000;
    `
    
    req := &advisor.ReviewRequest{
        Engine:          engineType,
        Statement:       sql,
        CurrentDatabase: dbParams.DbName,
        Rules:           rules,
        DBSchema:        metadata,
    }
    
    resp, err := advisor.SQLReviewCheck(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    // 计算影响行数
    affectedRows := services.CalculateAffectedRowsForStatements(sql, engineType, dbParams)
    
    // 转换结果
    results := services.ConvertToReviewResults(resp, sql, engineType, affectedRows)
    
    // 输出 JSON 格式
    services.OutputResults(resp, sql, engineType, "json", dbParams)
}
```

## 主要功能

### 1. SQL 审核

支持多种审核规则：
- 命名规范检查
- 语法规范检查  
- 性能优化建议
- 安全性检查
- 向后兼容性检查

### 2. 影响行数计算

自动计算 UPDATE/DELETE 语句的影响行数：
```go
affectedRows := services.CalculateAffectedRowsForStatements(sql, engineType, dbParams)
```

### 3. 元数据获取

从数据库获取 schema 信息：
```go
metadata, err := services.FetchDatabaseMetadata(engineType, dbParams)
```

### 4. 多种输出格式

- JSON 格式（兼容 Inception）
- 表格格式
- 结构化数据

```go
// JSON 输出
services.OutputResults(resp, sql, engineType, "json", dbParams)

// 表格输出
services.OutputResults(resp, sql, engineType, "table", dbParams)

// 结构化数据
results := services.ConvertToReviewResults(resp, sql, engineType, affectedRows)
```

## 支持的数据库

- ✅ PostgreSQL
- ✅ MySQL
- ✅ TiDB
- ✅ MSSQL
- ✅ Oracle
- ✅ MariaDB
- ✅ OceanBase
- ✅ Snowflake

## 常见问题

### Q: go get 时报错 "invalid version"？

A: 请确保使用 v1.0.6 或更高版本：
```bash
go get github.com/tianyuso/advisorTool@v1.0.6
```

### Q: 不连接数据库可以使用吗？

A: 可以！基础审核功能不需要数据库连接。只有以下功能需要连接：
- 元数据相关的高级规则
- 影响行数计算

### Q: 如何只审核不计算影响行数？

A: 不调用 `CalculateAffectedRowsForStatements` 即可，只使用 `SQLReviewCheck`。

### Q: 支持自定义规则吗？

A: 可以通过配置文件或代码自定义规则级别和参数。参考 `services.LoadRules()` 函数。

## 更多示例

项目包含多个示例文件：

```bash
# 克隆仓库查看示例
git clone https://github.com/tianyuso/advisorTool.git
cd advisorTool/examples

# 运行示例
go run simple_test.go                          # 简单测试
go run postgres_external_usage_example.go      # PostgreSQL 完整示例
go run library_usage_basic.go                  # 基础使用
go run library_usage_all_rules.go              # 所有规则示例
```

## 文档

- [完整 README](https://github.com/tianyuso/advisorTool/blob/main/README.md)
- [依赖版本修复说明](https://github.com/tianyuso/advisorTool/blob/main/DEPENDENCY_VERSION_FIX.md)
- [模块路径更新说明](https://github.com/tianyuso/advisorTool/blob/main/MODULE_PATH_UPDATE.md)
- [PostgreSQL Schema 使用](https://github.com/tianyuso/advisorTool/blob/main/POSTGRESQL_SCHEMA_NOTES.md)

## 问题反馈

遇到问题请提交 Issue：https://github.com/tianyuso/advisorTool/issues

---

**最新版本：** v1.0.6  
**更新日期：** 2024-12-22  
**状态：** ✅ 可用

