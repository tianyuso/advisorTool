# Services 包

这个包提供了 SQL Advisor 工具的核心服务功能，可以被外部 Go 程序引用。

## 背景

之前这些功能位于 `cmd/advisor/internal` 包中，由于 Go 语言的 `internal` 包机制，外部程序无法引用，会报错：
```
use of internal package advisorTool/cmd/advisor/internal not allowed
```

现在已迁移到 `services` 包，可以被任何外部程序正常引用。

## 功能模块

### 1. 规则配置 (config.go)

提供 SQL 审核规则的加载和管理功能：

- `LoadRules(configFile, engineType, hasMetadata)` - 从配置文件或默认规则加载审核规则
- `GetDefaultRules(engineType, hasMetadata)` - 获取指定数据库引擎的默认规则
- `GenerateSampleConfig(engineType)` - 生成示例配置文件

### 2. 结果处理 (result.go)

提供审核结果的转换和处理：

- `ConvertToReviewResults()` - 将审核响应转换为 Inception 兼容格式
- `CalculateAffectedRowsForStatements()` - 计算 SQL 语句的影响行数
- `SplitSQL()` - 分割多条 SQL 语句
- `DBConnectionParams` - 数据库连接参数结构

### 3. 输出格式化 (output.go)

提供多种输出格式支持：

- `OutputResults()` - 输出审核结果（支持 JSON 和表格格式）
- `ListAvailableRules()` - 列出所有可用的审核规则

### 4. 元数据获取 (metadata.go)

提供数据库元数据获取功能：

- `FetchDatabaseMetadata()` - 从数据库获取 schema 元数据

## 使用示例

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "advisorTool/pkg/advisor"
    "advisorTool/services"
)

func main() {
    // 1. 加载规则
    rules, err := services.LoadRules("", advisor.EngineMySQL, false)
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. 创建审核请求
    req := &advisor.ReviewRequest{
        Engine:          advisor.EngineMySQL,
        Statement:       "SELECT * FROM users",
        CurrentDatabase: "mydb",
        Rules:           rules,
    }
    
    // 3. 执行审核
    resp, err := advisor.SQLReviewCheck(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    // 4. 输出结果
    err = services.OutputResults(resp, req.Statement, req.Engine, "json", nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 带数据库连接的高级用法

```go
// 设置数据库连接参数
dbParams := &services.DBConnectionParams{
    Host:     "localhost",
    Port:     3306,
    User:     "root",
    Password: "password",
    DbName:   "mydb",
    Charset:  "utf8mb4",
    Timeout:  10,
}

// 获取数据库元数据
metadata, err := services.FetchDatabaseMetadata(advisor.EngineMySQL, dbParams)
if err != nil {
    log.Printf("Warning: %v", err)
} else {
    req.DBSchema = metadata
}

// 执行审核（会包含需要元数据的规则）
resp, err := advisor.SQLReviewCheck(context.Background(), req)
```

### 自定义结果处理

```go
// 转换为结构化结果
affectedRowsMap := services.CalculateAffectedRowsForStatements(
    sql, 
    engineType, 
    dbParams,
)

results := services.ConvertToReviewResults(
    resp, 
    sql, 
    engineType, 
    affectedRowsMap,
)

// 自定义处理结果
for _, result := range results {
    if result.ErrorLevel == "2" {
        fmt.Printf("错误: %s\n", result.ErrorMessage)
    }
}
```

## 完整示例

参考项目中的示例文件：
- `examples/external_usage_example.go` - 外部程序使用示例
- `examples/postgres_library_example.go` - PostgreSQL 完整示例
- `cmd/advisor/main.go` - CLI 工具实现

## 支持的数据库

- MySQL / MariaDB
- PostgreSQL
- TiDB
- OceanBase
- Oracle
- SQL Server (MSSQL)
- Snowflake

## 输出格式

### JSON 格式

```json
[
  {
    "order_id": 1,
    "stage": "CHECKED",
    "error_level": "1",
    "stage_status": "Audit Completed",
    "error_message": "[rule-type] message",
    "sql": "SELECT * FROM users",
    "affected_rows": 0,
    "sequence": "0_0_00000000"
  }
]
```

### 表格格式

使用 `go-pretty` 库输出美观的表格，包含颜色标识和统计信息。

## 注意事项

1. 某些规则需要数据库元数据才能运行（如向后兼容性检查）
2. 计算影响行数功能需要提供有效的数据库连接
3. 不同数据库引擎支持的规则集不同
4. 默认规则集已针对各数据库引擎优化

## 迁移说明

如果你的代码之前使用了 `cmd/advisor/internal` 包：

**旧的导入：**
```go
import "advisorTool/cmd/advisor/internal"
```

**新的导入：**
```go
import "advisorTool/services"
```

所有 API 保持不变，只需更改导入路径即可。

