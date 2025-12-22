# 快速开始 - 使用 GitHub 包

## 安装

```bash
go get github.com/tianyuso/advisorTool@latest
```

## 基础使用示例

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
    // 1. 配置数据库连接
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
    
    // 2. 获取数据库元数据
    metadata, err := services.FetchDatabaseMetadata(engineType, dbParams)
    if err != nil {
        log.Printf("警告: 获取元数据失败: %v", err)
        metadata = nil
    }
    
    // 3. 加载审核规则
    hasMetadata := (metadata != nil)
    rules, err := services.LoadRules("", engineType, hasMetadata)
    if err != nil {
        log.Fatalf("加载规则失败: %v", err)
    }
    
    // 4. 准备要审核的 SQL
    sql := `
    CREATE TABLE mydata.test_users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(50) NOT NULL,
        email VARCHAR(100)
    );
    
    UPDATE mydata.test_users SET status = 'active';
    DELETE FROM mydata.test_users;
    `
    
    // 5. 创建审核请求
    req := &advisor.ReviewRequest{
        Engine:          engineType,
        Statement:       sql,
        CurrentDatabase: dbParams.DbName,
        Rules:           rules,
        DBSchema:        metadata,
    }
    
    // 6. 执行 SQL 审核
    resp, err := advisor.SQLReviewCheck(context.Background(), req)
    if err != nil {
        log.Fatalf("SQL 审核失败: %v", err)
    }
    
    // 7. 输出结果
    fmt.Printf("审核完成，发现 %d 个问题\n", len(resp.Advices))
    
    // 8. 计算影响行数
    affectedRowsMap := services.CalculateAffectedRowsForStatements(sql, engineType, dbParams)
    results := services.ConvertToReviewResults(resp, sql, engineType, affectedRowsMap)
    
    // 9. 输出 JSON 格式
    if err := services.OutputResults(resp, sql, engineType, "json", dbParams); err != nil {
        log.Printf("输出结果失败: %v", err)
    }
}
```

## 运行示例

项目包含多个示例文件，可以直接运行：

```bash
# PostgreSQL 完整示例
cd examples
go run postgres_external_usage_example.go

# 基础使用示例
go run library_usage_basic.go

# 完整规则示例
go run library_usage_all_rules.go
```

## 支持的数据库

- PostgreSQL ✅
- MySQL ✅
- TiDB ✅
- MSSQL ✅
- Oracle ✅
- MariaDB ✅
- OceanBase ✅
- Snowflake ✅

## 主要功能

1. **SQL 审核** - 基于预定义规则审核 SQL 语句
2. **影响行数计算** - 计算 UPDATE/DELETE 语句的影响行数
3. **元数据获取** - 从数据库获取 schema 元数据
4. **多种输出格式** - JSON、表格、文本格式输出

## 文档

- [完整文档](README.md)
- [模块路径更新说明](MODULE_PATH_UPDATE.md)
- [PostgreSQL Schema 使用说明](POSTGRESQL_SCHEMA_NOTES.md)
- [库使用指南](examples/LIBRARY_USAGE.md)

## 问题反馈

如果遇到问题，请在 GitHub 上提交 Issue：
https://github.com/tianyuso/advisorTool/issues

