# Demo 使用完整指南

## 📖 概述

本 demo 示例经过全面改进，现在提供：

1. ✅ **完整规则集支持** - 不再使用少数几条规则，而是使用完整的默认规则集
2. ✅ **数据库元数据支持** - 可以连接真实数据库获取元数据，启用更多高级规则
3. ✅ **公共辅助函数** - 统一的规则加载和数据库连接管理
4. ✅ **规则自动调整** - 根据是否有元数据自动调整规则集

## 🎯 核心改进

### 改进 1: 完整规则集

**之前**:
```go
// 只有 3 条规则
rules := []*advisor.SQLReviewRule{
    advisor.NewRule(advisor.RuleStatementNoSelectAll, advisor.RuleLevelWarning),
    advisor.NewRule(advisor.RuleStatementRequireWhereForUpdateDelete, advisor.RuleLevelError),
    advisor.NewRule(advisor.RuleTableRequirePK, advisor.RuleLevelError),
}
```

**现在**:
```go
// 获取完整的默认规则集
rules := common.GetDefaultRules(advisor.EngineMySQL, false)
// MySQL: 22 条规则（静态分析模式）
// MySQL with metadata: 26 条规则（包含元数据规则）
```

### 改进 2: 数据库元数据支持

**之前**: 没有数据库连接功能

**现在**:
```go
// 配置数据库连接
dbConfig := &common.DBConfig{
    Host:     "127.0.0.1",
    Port:     3306,
    User:     "root",
    Password: "password",
    DBName:   "test_db",
    Charset:  "utf8mb4",
    Timeout:  5,
}

// 获取元数据
metadata, err := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)

// 根据是否有元数据选择规则集
hasMetadata := (metadata != nil && err == nil)
rules := common.GetDefaultRules(advisor.EngineMySQL, hasMetadata)

// 审核时使用元数据
req := &advisor.ReviewRequest{
    Engine:    advisor.EngineMySQL,
    Statement: sql,
    Rules:     rules,
    DBSchema:  metadata,  // 元数据支持
}
```

### 改进 3: 规则分类

规则根据是否需要数据库元数据自动分类：

**静态分析规则（无需元数据）**:
- ✅ 通用规则：UPDATE/DELETE 必须有 WHERE、表必须有主键、禁止 SELECT *
- ✅ MySQL 特有：自增列必须是整数、禁止 BLOB 索引、禁止存储过程
- ✅ PostgreSQL 特有：并发创建索引、禁止易变默认值

**动态分析规则（需要元数据）**:
- ✅ 列 NULL 检查（需要现有表结构）
- ✅ NOT NULL 列默认值检查
- ✅ 向后兼容性检查
- ✅ 索引冗余检查

## 📊 规则数量统计

### MySQL / MariaDB / TiDB / OceanBase

**静态分析模式** (hasMetadata=false): 22 条规则
- 通用错误规则: 2 条
- 通用警告规则: 3 条
- MySQL 错误规则: 2 条
- MySQL 警告规则: 15 条

**动态分析模式** (hasMetadata=true): 26 条规则
- 静态规则: 22 条
- 元数据规则: 4 条（column.no-null, column.set-default-for-not-null, column.require-default, schema.backward-compatibility）

### PostgreSQL

**静态分析模式**: 18 条规则
- 通用规则: 5 条
- PostgreSQL 特有: 13 条

**动态分析模式**: 21 条规则
- 静态规则: 18 条
- 元数据规则: 3 条

### SQL Server (MSSQL)

**静态分析模式**: 6 条规则
**动态分析模式**: 8 条规则

### Oracle

**静态分析模式**: 7 条规则
**动态分析模式**: 9 条规则

### Snowflake

**静态分析模式**: 5 条规则
**动态分析模式**: 6 条规则

## 🛠️ 公共辅助函数说明

### common.GetDefaultRules()

获取指定数据库引擎的完整默认规则集。

```go
func GetDefaultRules(engineType advisor.Engine, hasMetadata bool) []*advisor.SQLReviewRule
```

**参数**:
- `engineType`: 数据库引擎类型
- `hasMetadata`: 是否有数据库元数据（影响规则集）

**返回**: 完整的规则列表

**示例**:
```go
// MySQL 静态规则（22 条）
rules := common.GetDefaultRules(advisor.EngineMySQL, false)

// MySQL 动态规则（26 条）
rules := common.GetDefaultRules(advisor.EngineMySQL, true)

// PostgreSQL 规则
rules := common.GetDefaultRules(advisor.EnginePostgres, false)
```

### common.FetchDatabaseMetadata()

从真实数据库获取元数据。

```go
func FetchDatabaseMetadata(engineType advisor.Engine, dbConfig *DBConfig) (*advisor.DatabaseSchemaMetadata, error)
```

**参数**:
- `engineType`: 数据库引擎类型
- `dbConfig`: 数据库连接配置（如果为 nil，返回 nil）

**返回**: 
- 数据库元数据（成功）
- nil（配置为空或连接失败）

**示例**:
```go
// MySQL 配置
dbConfig := &common.DBConfig{
    Host:     "127.0.0.1",
    Port:     3306,
    User:     "root",
    Password: "password",
    DBName:   "mydb",
    Charset:  "utf8mb4",
    Timeout:  5,
}

metadata, err := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
if err != nil {
    fmt.Printf("获取元数据失败: %v\n", err)
}

// PostgreSQL 配置
dbConfig := &common.DBConfig{
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "password",
    DBName:   "testdb",
    SSLMode:  "disable",
    Timeout:  5,
}

metadata, err := common.FetchDatabaseMetadata(advisor.EnginePostgres, dbConfig)
```

## 📝 使用示例

### 示例 1: 静态分析（推荐）

无需数据库连接，使用完整的静态规则集。

```go
package main

import (
    "context"
    "fmt"
    "demo/common"
    "advisorTool/pkg/advisor"
)

func main() {
    // 获取完整规则集（静态）
    rules := common.GetDefaultRules(advisor.EngineMySQL, false)
    fmt.Printf("规则数量: %d 条\n", len(rules))
    
    sql := `
SELECT * FROM users;
DELETE FROM orders WHERE id = 1;
CREATE TABLE test (name VARCHAR(50));
`
    
    req := &advisor.ReviewRequest{
        Engine:    advisor.EngineMySQL,
        Statement: sql,
        Rules:     rules,
    }
    
    resp, _ := advisor.SQLReviewCheck(context.Background(), req)
    
    for _, advice := range resp.Advices {
        fmt.Printf("[%s] %s\n", advice.Title, advice.Content)
    }
}
```

### 示例 2: 动态分析（高级）

连接数据库获取元数据，使用完整规则集。

```go
package main

import (
    "context"
    "fmt"
    "demo/common"
    "advisorTool/pkg/advisor"
)

func main() {
    // 配置数据库连接
    dbConfig := &common.DBConfig{
        Host:     "127.0.0.1",
        Port:     3306,
        User:     "root",
        Password: "password",
        DBName:   "production_db",
        Charset:  "utf8mb4",
        Timeout:  5,
    }
    
    // 获取元数据
    metadata, err := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
    if err != nil {
        fmt.Printf("警告: 无法获取元数据: %v\n", err)
        fmt.Println("将使用静态分析模式")
    }
    
    // 根据是否有元数据选择规则
    hasMetadata := (metadata != nil && err == nil)
    rules := common.GetDefaultRules(advisor.EngineMySQL, hasMetadata)
    fmt.Printf("规则数量: %d 条 (hasMetadata=%v)\n", len(rules), hasMetadata)
    
    sql := `
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
ALTER TABLE orders ADD COLUMN status VARCHAR(20) NOT NULL;
`
    
    req := &advisor.ReviewRequest{
        Engine:          advisor.EngineMySQL,
        Statement:       sql,
        Rules:           rules,
        DBSchema:        metadata,
        CurrentDatabase: "production_db",
    }
    
    resp, _ := advisor.SQLReviewCheck(context.Background(), req)
    
    if len(resp.Advices) == 0 {
        fmt.Println("✅ 通过审核")
    } else {
        fmt.Printf("发现 %d 个问题:\n", len(resp.Advices))
        for _, advice := range resp.Advices {
            fmt.Printf("  [%s] %s\n", advice.Title, advice.Content)
        }
    }
}
```

### 示例 3: 自定义规则集

在完整规则集基础上添加自定义配置。

```go
package main

import (
    "context"
    "demo/common"
    "advisorTool/pkg/advisor"
)

func main() {
    // 获取基础规则集
    baseRules := common.GetDefaultRules(advisor.EngineMySQL, false)
    
    // 添加自定义规则
    namingRule, _ := advisor.NewRuleWithPayload(
        advisor.RuleTableNaming,
        advisor.RuleLevelError,
        advisor.NamingRulePayload{
            Format:    "^[a-z][a-z0-9_]*$",
            MaxLength: 64,
        },
    )
    
    requiredColumns, _ := advisor.NewRuleWithPayload(
        advisor.RuleRequiredColumn,
        advisor.RuleLevelError,
        advisor.StringArrayTypeRulePayload{
            List: []string{"id", "created_at", "updated_at"},
        },
    )
    
    // 合并规则
    allRules := append(baseRules, namingRule, requiredColumns)
    
    // 使用合并后的规则进行审核
    req := &advisor.ReviewRequest{
        Engine:    advisor.EngineMySQL,
        Statement: sql,
        Rules:     allRules,
    }
    
    // ... 执行审核
}
```

## 🔧 配置说明

### DBConfig 结构

```go
type DBConfig struct {
    Host        string  // 数据库主机地址（必需）
    Port        int     // 数据库端口（必需）
    User        string  // 数据库用户名（必需）
    Password    string  // 数据库密码（必需）
    DBName      string  // 数据库名称（必需）
    Charset     string  // 字符集（MySQL，可选，默认 utf8mb4）
    ServiceName string  // Oracle 服务名（Oracle 专用）
    Sid         string  // Oracle SID（Oracle 专用）
    SSLMode     string  // PostgreSQL SSL 模式（PostgreSQL 专用，默认 disable）
    Timeout     int     // 连接超时时间（秒，可选，默认 5）
}
```

### 不同数据库的配置示例

**MySQL**:
```go
dbConfig := &common.DBConfig{
    Host:     "127.0.0.1",
    Port:     3306,
    User:     "root",
    Password: "password",
    DBName:   "test_db",
    Charset:  "utf8mb4",
    Timeout:  5,
}
```

**PostgreSQL**:
```go
dbConfig := &common.DBConfig{
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "password",
    DBName:   "testdb",
    SSLMode:  "disable",  // 或 "require"
    Timeout:  5,
}
```

**Oracle**:
```go
dbConfig := &common.DBConfig{
    Host:        "oracle.example.com",
    Port:        1521,
    User:        "system",
    Password:    "password",
    DBName:      "ORCL",
    ServiceName: "ORCL",  // 或使用 Sid
    Timeout:     10,
}
```

**SQL Server**:
```go
dbConfig := &common.DBConfig{
    Host:     "sqlserver.example.com",
    Port:     1433,
    User:     "sa",
    Password: "password",
    DBName:   "master",
    Timeout:  5,
}
```

## 💡 最佳实践

### 1. 优先使用静态分析

```go
// 推荐：大多数场景使用静态分析
rules := common.GetDefaultRules(advisor.EngineMySQL, false)
```

**优点**:
- ✅ 无需数据库连接，速度快
- ✅ 适合 CI/CD 集成
- ✅ 22 条规则已覆盖大部分场景

### 2. 生产环境使用动态分析

```go
// 生产环境：连接只读账号获取元数据
dbConfig := &common.DBConfig{
    Host:     os.Getenv("DB_HOST"),
    Port:     3306,
    User:     "readonly_user",  // 只读账号
    Password: os.Getenv("DB_PASSWORD"),
    DBName:   "production",
}

metadata, _ := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
rules := common.GetDefaultRules(advisor.EngineMySQL, metadata != nil)
```

**优点**:
- ✅ 更完整的规则集（26 条）
- ✅ 更精确的审核结果
- ✅ 支持向后兼容性检查

### 3. 错误处理

```go
metadata, err := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
if err != nil {
    fmt.Printf("⚠️ 警告: 无法获取元数据: %v\n", err)
    fmt.Println("将降级为静态分析模式")
    // 使用静态规则集
    rules = common.GetDefaultRules(advisor.EngineMySQL, false)
} else {
    // 使用动态规则集
    rules = common.GetDefaultRules(advisor.EngineMySQL, true)
}
```

### 4. 规则组合

```go
// 基础规则
baseRules := common.GetDefaultRules(advisor.EngineMySQL, false)

// 添加严格的自定义规则
strictRules := []*advisor.SQLReviewRule{
    // ... 自定义规则
}

// 合并
allRules := append(baseRules, strictRules...)
```

## 🎯 总结

通过这些改进，demo 示例现在提供：

1. ✅ **完整功能** - 使用 20+ 条完整规则，而不是 3 条示例规则
2. ✅ **生产就绪** - 支持数据库元数据，可用于生产环境
3. ✅ **易于使用** - 公共辅助函数简化了配置
4. ✅ **灵活扩展** - 可以在基础规则上添加自定义配置

---

**GitHub**: https://github.com/tianyuso/advisorTool
**文档**: ../README.md

