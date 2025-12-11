# SQL Advisor Tool - Go 库使用示例

本目录包含了将 SQL Advisor Tool 作为 Go 库使用的完整示例代码。

## ✨ 主要改进

- ✅ **完整规则集**: 所有示例都使用完整的默认规则集（20+ 规则）
- ✅ **数据库连接支持**: 支持连接真实数据库获取元数据
- ✅ **公共辅助函数**: `common/helpers.go` 提供统一的规则加载和元数据获取
- ✅ **规则分类**: 根据是否需要元数据自动调整规则集

## 📁 示例文件

### 1. `basic_usage.go` - 基础用法（完整规则集）
展示使用完整默认规则集进行审核。

**包含内容**:
- ✅ 静态分析模式（无需数据库连接）
- ✅ 动态分析模式（支持数据库元数据）
- ✅ 批量 SQL 语句审核（使用完整规则）
- ✅ 不同数据库引擎的完整规则集

**运行方式**:
```bash
cd demo
go run basic_usage.go
```

**核心改进**:
```go
// 使用完整的默认规则集
rules := common.GetDefaultRules(advisor.EngineMySQL, false)
// MySQL: 22 条规则（静态分析）
// MySQL with metadata: 26 条规则（包含元数据规则）

// 支持数据库连接
dbConfig := &common.DBConfig{
    Host:     "127.0.0.1",
    Port:     3306,
    User:     "root",
    Password: "password",
    DBName:   "test_db",
}
metadata, _ := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)

req := &advisor.ReviewRequest{
    Engine:    advisor.EngineMySQL,
    Statement: sql,
    Rules:     rules,
    DBSchema:  metadata,  // 可选：元数据支持
}
```

### 2. `advanced_usage.go` - 高级用法（完整规则集 + Payload）
展示如何在完整规则集基础上添加自定义 Payload 配置。

**包含内容**:
- ✅ 完整基础规则集 + 自定义命名规范
- ✅ 综合配置（完整规则 + 类型限制 + 数值限制）
- ✅ 支持数据库元数据的高级审核
- ✅ 生产环境完整配置（30+ 规则）

**运行方式**:
```bash
cd demo
go run advanced_usage.go
```

**核心改进**:
```go
// 获取完整基础规则集
baseRules := common.GetDefaultRules(advisor.EngineMySQL, false)

// 在基础规则上添加自定义配置
namingRule, _ := advisor.NewRuleWithPayload(
    advisor.RuleTableNaming,
    advisor.RuleLevelWarning,
    advisor.NamingRulePayload{
        Format:    "^[a-z][a-z0-9_]*$",
        MaxLength: 64,
    },
)

typeRule, _ := advisor.NewRuleWithPayload(
    advisor.RuleColumnTypeDisallowList,
    advisor.RuleLevelError,
    advisor.StringArrayTypeRulePayload{
        List: []string{"BLOB", "TEXT"},
    },
)

// 合并规则
allRules := append(baseRules, namingRule, typeRule)
// 总规则数: 25+ 条

// 支持元数据
metadata, _ := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
req.DBSchema = metadata
```

### 3. `batch_review.go` - 批量审核（完整规则集 + 元数据）
展示如何批量审核多个 SQL 文件，使用完整规则集。

**包含内容**:
- ✅ 从文件读取 SQL（使用完整规则集）
- ✅ 批量审核多个文件（支持元数据）
- ✅ 生成详细审核报告（问题分类、统计、修复建议）
- ✅ 汇总报告和最终评估

**运行方式**:
```bash
cd demo
go run batch_review.go
```

**核心改进**:
```go
// 支持数据库元数据
metadata, _ := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
hasMetadata := (metadata != nil)

// 获取完整规则集（根据元数据自动调整）
rules := common.GetDefaultRules(advisor.EngineMySQL, hasMetadata)

// 批量审核多个文件
files, _ := filepath.Glob(filepath.Join(tmpDir, "*.sql"))
for _, file := range files {
    content, _ := ioutil.ReadFile(file)
    
    req := &advisor.ReviewRequest{
        Engine:    advisor.EngineMySQL,
        Statement: string(content),
        Rules:     rules,
        DBSchema:  metadata,  // 支持元数据
    }
    
    resp, _ := advisor.SQLReviewCheck(context.Background(), req)
    // 处理结果...
}

// 生成详细报告
generateDetailedReport(resp, sqlContent)
```

**新增功能**:
- 📊 问题分类统计（语句规范、表结构、列规范等）
- 🎯 修复优先级排序
- 💡 详细的修复建议
- 📈 批量审核汇总报告

### 4. `common/helpers.go` - 公共辅助函数
提供统一的规则加载和数据库连接功能。

**核心功能**:
```go
// 获取完整默认规则集
func GetDefaultRules(engineType advisor.Engine, hasMetadata bool) []*advisor.SQLReviewRule

// 获取数据库元数据
func FetchDatabaseMetadata(engineType advisor.Engine, dbConfig *DBConfig) (*advisor.DatabaseSchemaMetadata, error)

// 数据库配置结构
type DBConfig struct {
    Host        string
    Port        int
    User        string
    Password    string
    DBName      string
    Charset     string  // MySQL
    ServiceName string  // Oracle
    Sid         string  // Oracle
    SSLMode     string  // PostgreSQL
    Timeout     int
}
```

**规则数量统计**:
- MySQL（静态）: 22 条规则
- MySQL（含元数据）: 26 条规则
- PostgreSQL（静态）: 18 条规则
- PostgreSQL（含元数据）: 21 条规则
- MSSQL（静态）: 6 条规则
- Oracle（静态）: 7 条规则

## 🚀 快速开始

### 1. 环境准备

```bash
# 克隆项目
git clone https://github.com/tianyuso/advisorTool.git
cd advisorTool/demo

# 安装依赖（首次运行）
go mod tidy
```

### 2. 运行示例

**基础示例（无需数据库）**:
```bash
go run basic_usage.go
```

**高级示例（无需数据库）**:
```bash
go run advanced_usage.go
```

**批量审核示例**:
```bash
go run batch_review.go
```

### 3. 使用数据库元数据（可选）

如需测试元数据相关功能，请在代码中取消注释并配置数据库连接：

```go
// 在示例文件中找到并修改此配置
dbConfig := &common.DBConfig{
    Host:     "127.0.0.1",
    Port:     3306,
    User:     "root",
    Password: "your_password",
    DBName:   "test_db",
    Charset:  "utf8mb4",
    Timeout:  5,
}
```

**元数据规则优势**:
- ✅ 列 NULL 检查（需要现有表结构）
- ✅ 向后兼容性检查（需要变更前后对比）
- ✅ 索引冗余检查（需要现有索引信息）
- ✅ 更精确的 DDL 审核

### 4. 在您的项目中使用

```bash
# 添加依赖
go get github.com/tianyuso/advisorTool
```

在代码中使用:
```go
import (
    "advisorTool/pkg/advisor"
    "github.com/tianyuso/advisorTool/db"  // 如需数据库连接
)
```

## 📚 完整示例场景

### 场景 1: CI/CD 集成（完整规则集）

```go
package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "os"
    
    "github.com/tianyuso/advisorTool/pkg/advisor"
    "github.com/tianyuso/advisorTool/demo/common"
)

func main() {
    // 读取变更脚本
    sql, _ := ioutil.ReadFile("migration.sql")
    
    // 使用完整的严格规则集
    rules := common.GetDefaultRules(advisor.EngineMySQL, false)
    
    // 可选：连接生产数据库获取元数据
    dbConfig := &common.DBConfig{
        Host:     os.Getenv("DB_HOST"),
        Port:     3306,
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        DBName:   os.Getenv("DB_NAME"),
    }
    metadata, _ := common.FetchDatabaseMetadata(advisor.EngineMySQL, dbConfig)
    
    if metadata != nil {
        // 如果有元数据，使用更完整的规则集
        rules = common.GetDefaultRules(advisor.EngineMySQL, true)
    }
    
    req := &advisor.ReviewRequest{
        Engine:    advisor.EngineMySQL,
        Statement: string(sql),
        Rules:     rules,
        DBSchema:  metadata,
    }
    
    resp, _ := advisor.SQLReviewCheck(context.Background(), req)
    
    // 有错误则中止部署
    if resp.HasError {
        fmt.Printf("❌ SQL 审核失败，发现 %d 个问题\n", len(resp.Advices))
        for _, advice := range resp.Advices {
            fmt.Printf("  [%s] %s\n", advice.Title, advice.Content)
        }
        os.Exit(1)
    }
    
    fmt.Printf("✅ SQL 审核通过 (%d 条规则)\n", len(rules))
}
```

### 场景 2: Web 服务集成

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"
    
    "github.com/tianyuso/advisorTool/pkg/advisor"
)

type ReviewRequest struct {
    SQL    string `json:"sql"`
    Engine string `json:"engine"`
}

func reviewHandler(w http.ResponseWriter, r *http.Request) {
    var req ReviewRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 执行审核
    rules := []*advisor.SQLReviewRule{
        advisor.NewRule(advisor.RuleStatementNoSelectAll, advisor.RuleLevelWarning),
    }
    
    reviewReq := &advisor.ReviewRequest{
        Engine:    advisor.EngineFromString(req.Engine),
        Statement: req.SQL,
        Rules:     rules,
    }
    
    resp, _ := advisor.SQLReviewCheck(context.Background(), reviewReq)
    
    // 返回结果
    json.NewEncoder(w).Encode(resp)
}

func main() {
    http.HandleFunc("/api/review", reviewHandler)
    http.ListenAndServe(":8080", nil)
}
```

### 场景 3: 定制规则配置

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/tianyuso/advisorTool/pkg/advisor"
)

func createProductionRules() []*advisor.SQLReviewRule {
    // 命名规范
    tableNaming, _ := advisor.NewRuleWithPayload(
        advisor.RuleTableNaming,
        advisor.RuleLevelWarning,
        advisor.NamingRulePayload{
            Format:    "^[a-z][a-z0-9_]*$",
            MaxLength: 64,
        },
    )
    
    // 必需列
    requiredColumns, _ := advisor.NewRuleWithPayload(
        advisor.RuleRequiredColumn,
        advisor.RuleLevelError,
        advisor.StringArrayTypeRulePayload{
            List: []string{"id", "created_at", "updated_at"},
        },
    )
    
    return []*advisor.SQLReviewRule{
        tableNaming,
        requiredColumns,
        advisor.NewRule(advisor.RuleTableRequirePK, advisor.RuleLevelError),
        advisor.NewRule(advisor.RuleStatementRequireWhereForUpdateDelete, advisor.RuleLevelError),
    }
}

func main() {
    rules := createProductionRules()
    
    req := &advisor.ReviewRequest{
        Engine:    advisor.EngineMySQL,
        Statement: "CREATE TABLE users (name VARCHAR(50));",
        Rules:     rules,
    }
    
    resp, _ := advisor.SQLReviewCheck(context.Background(), req)
    
    for _, advice := range resp.Advices {
        fmt.Printf("[%s] %s\n", advice.Title, advice.Content)
    }
}
```

## 🎯 核心 API 说明

### 1. 创建审核请求

```go
type ReviewRequest struct {
    Engine          Engine                  // 数据库引擎
    Statement       string                  // SQL 语句
    Rules           []*SQLReviewRule        // 审核规则
    CurrentDatabase string                  // 当前数据库（可选）
    DBSchema        *DatabaseSchemaMetadata // 数据库元数据（可选）
}
```

### 2. 执行审核

```go
func SQLReviewCheck(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error)
```

### 3. 审核响应

```go
type ReviewResponse struct {
    Advices    []*Advice  // 审核建议列表
    HasError   bool       // 是否有错误级别问题
    HasWarning bool       // 是否有警告级别问题
}

type Advice struct {
    Status        Status    // ERROR/WARNING/SUCCESS
    Code          int32     // 错误码
    Title         string    // 规则标题
    Content       string    // 问题描述
    StartPosition *Position // 位置信息（行号、列号）
}
```

### 4. 规则创建

```go
// 基础规则
func NewRule(ruleType string, level RuleLevel) *SQLReviewRule

// 带 Payload 的规则
func NewRuleWithPayload(ruleType string, level RuleLevel, payload interface{}) (*SQLReviewRule, error)
```

### 5. Payload 类型

```go
// 命名规则
type NamingRulePayload struct {
    Format    string  // 正则表达式
    MaxLength int     // 最大长度
}

// 数值规则
type NumberTypeRulePayload struct {
    Number int  // 数值限制
}

// 字符串数组规则
type StringArrayTypeRulePayload struct {
    List []string  // 列表
}

// 注释规范规则
type CommentConventionRulePayload struct {
    Required  bool  // 是否必需
    MaxLength int   // 最大长度
}
```

## 🔧 常用规则列表

### 必备规则（推荐在所有环境启用）

| 规则常量 | 说明 | 级别 |
|---------|------|------|
| `RuleStatementRequireWhereForUpdateDelete` | UPDATE/DELETE 必须有 WHERE | ERROR |
| `RuleTableRequirePK` | 表必须有主键 | ERROR |
| `RuleStatementNoSelectAll` | 禁止 SELECT * | WARNING |
| `RuleTableNoFK` | 禁止外键 | WARNING |

### 命名规范

| 规则常量 | 说明 |
|---------|------|
| `RuleTableNaming` | 表命名规范 |
| `RuleColumnNaming` | 列命名规范 |
| `RuleIDXNaming` | 索引命名规范 |

### 列规则

| 规则常量 | 说明 |
|---------|------|
| `RuleRequiredColumn` | 必需列 |
| `RuleColumnNotNull` | 禁止 NULL 值 |
| `RuleColumnTypeDisallowList` | 列类型黑名单 |

### 性能规则

| 规则常量 | 说明 |
|---------|------|
| `RuleStatementNoLeadingWildcardLike` | 禁止前导 % |
| `RuleIndexNoDuplicateColumn` | 禁止重复索引列 |
| `RuleTableNoDuplicateIndex` | 禁止重复索引 |

完整规则列表请参考: [pkg/advisor/rules.go](../pkg/advisor/rules.go)

## 💡 最佳实践

### 1. 规则分级使用

```go
// 开发环境 - 宽松
devRules := []*advisor.SQLReviewRule{
    advisor.NewRule(advisor.RuleTableRequirePK, advisor.RuleLevelWarning),
    advisor.NewRule(advisor.RuleStatementRequireWhereForUpdateDelete, advisor.RuleLevelWarning),
}

// 生产环境 - 严格
prodRules := []*advisor.SQLReviewRule{
    advisor.NewRule(advisor.RuleTableRequirePK, advisor.RuleLevelError),
    advisor.NewRule(advisor.RuleStatementRequireWhereForUpdateDelete, advisor.RuleLevelError),
    advisor.NewRule(advisor.RuleSchemaBackwardCompatibility, advisor.RuleLevelError),
}
```

### 2. 错误处理

```go
resp, err := advisor.SQLReviewCheck(ctx, req)
if err != nil {
    log.Printf("审核失败: %v", err)
    return err
}

// 根据不同级别采取不同行动
if resp.HasError {
    // 阻止部署
    return errors.New("存在错误级别问题")
} else if resp.HasWarning {
    // 记录警告，但允许继续
    log.Println("存在警告，需要人工确认")
}
```

### 3. 结果缓存

```go
// 对相同 SQL 进行缓存
type ReviewCache struct {
    cache map[string]*advisor.ReviewResponse
    mu    sync.RWMutex
}

func (c *ReviewCache) Review(sql string, rules []*advisor.SQLReviewRule) (*advisor.ReviewResponse, error) {
    key := fmt.Sprintf("%x", md5.Sum([]byte(sql)))
    
    c.mu.RLock()
    if cached, ok := c.cache[key]; ok {
        c.mu.RUnlock()
        return cached, nil
    }
    c.mu.RUnlock()
    
    // 执行审核
    req := &advisor.ReviewRequest{...}
    resp, err := advisor.SQLReviewCheck(context.Background(), req)
    
    if err == nil {
        c.mu.Lock()
        c.cache[key] = resp
        c.mu.Unlock()
    }
    
    return resp, err
}
```

## 📖 更多资源

- [项目主页](https://github.com/tianyuso/advisorTool)
- [完整文档](../README.md)
- [配置示例](../examples/)
- [Bytebase 官方文档](https://www.bytebase.com/docs)

## ❓ 常见问题

**Q: 如何添加自定义规则？**

A: 实现 `Advisor` 接口并注册到系统中。参考 [advisor/pg/advisor_statement_where_required_update_delete.go](../advisor/pg/advisor_statement_where_required_update_delete.go)

**Q: 性能如何？**

A: 静态分析通常在 100-500ms 内完成（1000 行 SQL）。建议对大文件进行分批处理。

**Q: 是否支持并发？**

A: 是的，`SQLReviewCheck` 函数是线程安全的，可以并发调用。

## 📄 许可证

遵循 Bytebase 项目的 GPL-3.0 许可证。

---

**GitHub**: https://github.com/tianyuso/advisorTool

有问题或建议？欢迎提 [Issue](https://github.com/tianyuso/advisorTool/issues)！

