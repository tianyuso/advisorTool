# SQL Advisor Tool

一个基于 [Bytebase](https://github.com/bytebase/bytebase) SQL 审核引擎的独立命令行工具。完整保留 Bytebase 原有的 SQL 解析器和审核规则实现，支持 MySQL、PostgreSQL、Oracle、SQL Server 等多种数据库。

## 特性

- 🔍 **多数据库支持**: MySQL, MariaDB, PostgreSQL, Oracle, SQL Server, TiDB, Snowflake, OceanBase
- 📋 **完整的审核规则**: 90+ 种内置规则，覆盖命名规范、语句规范、表设计、索引优化等
- 🛠️ **原生解析器**: 使用 Bytebase 原有的 ANTLR4 解析器，保证解析准确性
  - MySQL/MariaDB/OceanBase: `github.com/bytebase/parser/mysql`
  - PostgreSQL: `github.com/bytebase/parser/postgresql`  
  - Oracle: `github.com/bytebase/parser/plsql`
  - SQL Server: `github.com/bytebase/parser/tsql`
  - TiDB: `github.com/pingcap/tidb/parser`
  - Snowflake: `github.com/bytebase/parser/snowflake`
- ⚙️ **高度可配置**: 通过 YAML/JSON 配置文件自定义规则和级别
- 📊 **多种输出格式**: 文本（可读性强）、JSON（兼容 Inception 格式）、YAML
- 🔌 **数据库连接**: 支持连接真实数据库获取元数据，提供更精确的审核
- 📚 **两种使用方式**: 命令行工具和 Go 库，灵活集成

## 核心架构

### 1. 解析器层（Parser Layer）

使用 ANTLR4 语法树解析器，精确解析 SQL 语句：

```
SQL 输入 → ANTLR Parser → 语法树 (AST) → TreeWalker → 规则检查器
```

- **优势**: 完全理解 SQL 语法结构，不是简单的正则匹配
- **实现**: 基于 Bytebase 原生解析器，各数据库使用对应的官方语法规范

### 2. 审核引擎（Advisor Engine）

采用插件化的规则注册机制：

```go
// 每个规则实现 Advisor 接口
type Advisor interface {
    Check(ctx context.Context, checkCtx Context) ([]*Advice, error)
}

// 通过 init() 函数自动注册
func init() {
    advisor.Register(storepb.Engine_POSTGRES, 
                    advisor.SchemaRuleStatementRequireWhereForUpdateDelete, 
                    &StatementWhereRequiredUpdateDeleteAdvisor{})
}
```

### 3. 规则检查原理

以 "UPDATE/DELETE 必须有 WHERE 条件" 为例：

**检查流程**:
1. **语法树遍历**: 使用 `TreeWalker` 遍历 ANTLR 生成的语法树
2. **节点识别**: 监听 `UpdateStmt` 和 `DeleteStmt` 节点
3. **条件判断**: 检查节点是否包含 `Where_or_current_clause()` 子节点
4. **生成建议**: 如果缺失 WHERE 子句，生成 `Advice` 错误/警告

**代码示例** (`advisor/pg/advisor_statement_where_required_update_delete.go`):

```go
func (r *statementWhereRequiredRule) handleUpdatestmt(ctx *parser.UpdatestmtContext) {
    // 1. 只检查顶层语句（忽略子查询）
    if !isTopLevel(ctx.GetParent()) {
        return
    }

    // 2. 检查 WHERE 子句是否存在
    if ctx.Where_or_current_clause() == nil || ctx.Where_or_current_clause().WHERE() == nil {
        // 3. 提取原始 SQL 文本
        stmtText := extractStatementText(r.statementsText, ctx.GetStart().GetLine(), ctx.GetStop().GetLine())
        
        // 4. 生成审核建议
        r.AddAdvice(&storepb.Advice{
            Status:  r.level,               // ERROR/WARNING
            Code:    code.StatementNoWhere.Int32(),
            Title:   r.title,
            Content: fmt.Sprintf("\"%s\" requires WHERE clause", stmtText),
            StartPosition: &storepb.Position{
                Line:   int32(ctx.GetStart().GetLine()),
                Column: 0,
            },
        })
    }
}
```

**关键技术**:
- ✅ 基于语法树，不是正则匹配
- ✅ 精确定位错误行号和列号
- ✅ 支持复杂 SQL 结构（子查询、CTE、多表 JOIN）
- ✅ 可扩展：新增规则只需实现 `Advisor` 接口

### 4. 规则分类

**静态分析规则**（无需数据库连接）:
- 命名规范检查（表名、列名、索引名）
- 语句结构检查（SELECT *、WHERE 子句、LIMIT）
- DDL 规范检查（主键要求、外键禁止、分区表）
- 索引规范检查（重复索引、BLOB 索引）

**动态分析规则**（需要数据库元数据）:
- 列 NULL 检查（需要知道现有列定义）
- 向后兼容性检查（需要对比变更前后的 schema）
- 索引冗余检查（需要现有索引信息）
- DML 空运行验证（需要实际执行查询计划）

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
# 审核 SQL 语句（使用默认规则）
./advisor -engine mysql -sql "SELECT * FROM users"

# 审核 SQL 文件
./advisor -engine postgres -file schema.sql

# 从标准输入读取 SQL
cat schema.sql | ./advisor -engine mysql -sql -

# 使用自定义配置文件
./advisor -engine mysql -config review-config.yaml -file schema.sql

# 输出 JSON 格式（兼容 Inception 格式）
./advisor -engine mysql -sql "SELECT * FROM users" -format json

# 列出所有可用规则
./advisor -list-rules

# 生成示例配置文件
./advisor -engine mysql -generate-config > mysql-config.yaml

# 连接数据库进行审核（支持需要元数据的规则）
./advisor -engine mysql \
  -host 127.0.0.1 \
  -port 3306 \
  -user root \
  -password xxx \
  -dbname mydb \
  -file schema.sql
```

### 命令行参数

**基础参数**:

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

**数据库连接参数**（可选，用于获取元数据）:

| 参数 | 说明 |
|------|------|
| `-host` | 数据库主机地址 |
| `-port` | 数据库端口 |
| `-user` | 数据库用户名 |
| `-password` | 数据库密码 |
| `-dbname` | 数据库名称 |
| `-charset` | 字符集（MySQL，默认: utf8mb4） |
| `-service-name` | Oracle 服务名 |
| `-sid` | Oracle SID |
| `-sslmode` | PostgreSQL SSL 模式（默认: disable） |
| `-timeout` | 连接超时时间（秒，默认: 5） |

### 退出码

- `0`: 审核通过，没有问题
- `1`: 发现警告级别的问题
- `2`: 发现错误级别的问题

### 作为 Go 库使用

以下示例展示如何在 Go 项目中直接使用 SQL Advisor 库进行 SQL 审核。

> 💡 **提示**: 现在可以使用 `services` 包来简化常见操作，如加载默认规则、格式化输出等。

#### Services 包功能概览

`services` 包（位于 `advisorTool/services`）提供了便捷的辅助函数，可以被外部 Go 程序直接引用：

| 功能 | 函数 | 说明 |
|------|------|------|
| 规则加载 | `LoadRules(configFile, engineType, hasMetadata)` | 从配置文件或获取默认规则 |
| 规则获取 | `GetDefaultRules(engineType, hasMetadata)` | 获取指定数据库的默认规则 |
| 配置生成 | `GenerateSampleConfig(engineType)` | 生成示例配置文件 |
| 结果转换 | `ConvertToReviewResults(resp, sql, engine, affectedRows)` | 转换为 Inception 兼容格式 |
| 结果输出 | `OutputResults(resp, sql, engine, format, dbParams)` | 格式化输出（JSON/表格） |
| 元数据获取 | `FetchDatabaseMetadata(engineType, dbParams)` | 从数据库获取元数据 |
| 影响行数 | `CalculateAffectedRowsForStatements(sql, engine, dbParams)` | 计算 SQL 影响行数 |
| 规则列表 | `ListAvailableRules()` | 列出所有可用规则 |

**使用 services 包的优势**:
- ✅ 无需处理 internal 包限制（之前的 `cmd/advisor/internal` 不能被外部引用）
- ✅ 开箱即用的默认规则集，针对各数据库优化
- ✅ 统一的结果格式转换和输出
- ✅ 简化数据库连接和元数据获取
- ✅ 完整的文档和示例代码

详细文档请参考：[services/README.md](services/README.md)
# 模块路径更新说明

## 问题描述

当将项目发布到 GitHub 后，使用 `go get` 下载时出现错误：

```bash
go get github.com/tianyuso/advisorTool@v1.0.4
# 错误: module declares its path as: advisorTool
#       but was required as: github.com/tianyuso/advisorTool
```

## 原因分析

这是因为 `go.mod` 文件中声明的模块路径与 GitHub 仓库路径不匹配。

**错误配置：**
```go
module advisorTool  // ❌ 本地路径
```

**正确配置：**
```go
module github.com/tianyuso/advisorTool  // ✅ GitHub 完整路径
```

## 已修复内容

### 1. 更新 go.mod

```diff
- module advisorTool
+ module github.com/tianyuso/advisorTool
```

### 2. 更新所有 Go 文件中的导入路径

批量替换所有文件中的导入路径：

```bash
# 旧的导入方式
import "advisorTool/pkg/advisor"
import "advisorTool/services"

# 新的导入方式
import "github.com/tianyuso/advisorTool/pkg/advisor"
import "github.com/tianyuso/advisorTool/services"
```

## 使用方法

### 在本地项目中使用

现在可以正常使用 `go get` 下载：

```bash
go get github.com/tianyuso/advisorTool@latest
# 或指定版本
go get github.com/tianyuso/advisorTool@v1.0.4
```

### 在代码中导入

```go
package main

import (
    "github.com/tianyuso/advisorTool/pkg/advisor"
    "github.com/tianyuso/advisorTool/services"
)

func main() {
    // 使用 advisor 包
    engineType := advisor.EnginePostgres
    
    // 使用 services 包
    dbParams := &services.DBConnectionParams{
        Host:     "127.0.0.1",
        Port:     5432,
        User:     "postgres",
        Password: "secret",
        DbName:   "mydb",
    }
    
    // ... 其他代码
}
```

## 验证

### 1. 本地编译验证

```bash
cd /data/dev_go/advisorTool
go mod tidy
go build ./advisor/... ./db/... ./services/... ./pkg/...
```

### 2. 外部项目使用验证

在其他项目中测试：

```bash
mkdir /tmp/test-advisortool
cd /tmp/test-advisortool
go mod init test-project

# 下载库
go get github.com/tianyuso/advisorTool@latest

# 创建测试文件
cat > main.go << 'EOF'
package main

import (
    "fmt"
    "github.com/tianyuso/advisorTool/pkg/advisor"
)

func main() {
    fmt.Println("Engine:", advisor.EnginePostgres)
}
EOF

# 编译运行
go run main.go
```
#### 基础用法（使用 services 包）

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
	// 1. 使用 services 包加载默认规则（推荐）
	engineType := advisor.EngineMySQL
	hasMetadata := false
	
	rules, err := services.LoadRules("", engineType, hasMetadata)
	if err != nil {
		log.Fatalf("加载规则失败: %v", err)
	}
	
	fmt.Printf("✅ 成功加载 %d 条规则\n\n", len(rules))

	// 2. 准备要审核的 SQL
	sql := `
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

SELECT * FROM users WHERE id = 1;

UPDATE users SET email = 'new@email.com';
`

	// 3. 创建审核请求
	req := &advisor.ReviewRequest{
		Engine:          engineType,
		Statement:       sql,
		CurrentDatabase: "testdb",
		Rules:           rules,
	}

	// 4. 执行 SQL 审核
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		log.Fatalf("SQL 审核失败: %v", err)
	}

	fmt.Printf("审核完成，发现 %d 个问题\n\n", len(resp.Advices))

	// 5. 使用 services 包的 ConvertToReviewResults 转换结果
	affectedRowsMap := make(map[int]int)
	results := services.ConvertToReviewResults(resp, sql, engineType, affectedRowsMap)

	// 6. 输出结果
	fmt.Println("=== 审核结果 ===")
	for _, result := range results {
		level := "✓ OK"
		if result.ErrorLevel == "1" {
			level = "⚠ WARNING"
		} else if result.ErrorLevel == "2" {
			level = "✗ ERROR"
		}

		fmt.Printf("%d. [%s] %s\n", result.OrderID, level, result.SQL)
		if result.ErrorMessage != "" {
			fmt.Printf("   问题: %s\n", result.ErrorMessage)
		}
		fmt.Println()
	}

	// 7. 也可以使用 services.OutputResults 直接输出格式化结果
	fmt.Println("\n=== 使用 services.OutputResults 输出（JSON 格式） ===")
	if err := services.OutputResults(resp, sql, engineType, "json", nil); err != nil {
		log.Printf("输出结果失败: %v", err)
	}
}
```

#### 手动配置规则（传统方式）

```go
package main

import (
	"context"
	"fmt"
	"log"

	"advisorTool/pkg/advisor"
)

func main() {
	// 定义要审核的 SQL 语句
	sql := `
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50),
    email VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

SELECT * FROM users;

UPDATE users SET status = 'active';
`

	// 手动配置审核规则
	rules := []*advisor.SQLReviewRule{
		// 错误级别：UPDATE/DELETE 必须有 WHERE 条件
		advisor.NewRule(
			string(advisor.SchemaRuleStatementRequireWhereForUpdateDelete),
			advisor.RuleLevelError,
		),
		// 警告级别：禁止 SELECT *
		advisor.NewRule(
			string(advisor.SchemaRuleStatementNoSelectAll),
			advisor.RuleLevelWarning,
		),
		// 错误级别：表必须有主键
		advisor.NewRule(
			string(advisor.SchemaRuleTableRequirePK),
			advisor.RuleLevelError,
		),
		// 警告级别：禁止外键
		advisor.NewRule(
			string(advisor.SchemaRuleTableNoFK),
			advisor.RuleLevelWarning,
		),
	}

	// 构建审核请求
	req := &advisor.ReviewRequest{
		Engine:          advisor.EnginePostgres,
		Statement:       sql,
		CurrentDatabase: "mydb",
		Rules:           rules,
	}

	// 执行 SQL 审核
	ctx := context.Background()
	resp, err := advisor.SQLReviewCheck(ctx, req)
	if err != nil {
		log.Fatalf("SQL 审核失败: %v", err)
	}

	// 输出审核结果
	fmt.Printf("审核完成，共发现 %d 个问题\n\n", len(resp.Advices))
	
	for i, advice := range resp.Advices {
		statusStr := ""
		switch advice.Status {
		case advisor.AdviceStatusError:
			statusStr = "❌ [ERROR]"
		case advisor.AdviceStatusWarning:
			statusStr = "⚠️  [WARNING]"
		case advisor.AdviceStatusSuccess:
			statusStr = "✅ [OK]"
		}
		
		fmt.Printf("%d. %s %s\n", i+1, statusStr, advice.Title)
		fmt.Printf("   内容: %s\n", advice.Content)
		if advice.StartPosition != nil {
			fmt.Printf("   位置: Line %d\n", advice.StartPosition.Line)
		}
		fmt.Println()
	}

	// 根据审核结果决定是否允许执行
	if resp.HasError {
		fmt.Println("❌ SQL 审核不通过，存在错误级别的问题，拒绝执行！")
	} else if resp.HasWarning {
		fmt.Println("⚠️  SQL 审核通过，但存在警告，建议修改后再执行")
	} else {
		fmt.Println("✅ SQL 审核通过，可以安全执行")
	}
}
```

#### 高级用法（连接数据库获取元数据）

连接数据库可以启用更多需要元数据的审核规则，如列 NULL 检查、向后兼容性检查等。

**使用 services 包（推荐）**:

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
	// 1. 配置数据库连接参数
	dbParams := &services.DBConnectionParams{
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DbName:   "mydb",
		SSLMode:  "disable",
		Timeout:  10,
	}

	// 2. 使用 services 包获取数据库元数据
	engineType := advisor.EnginePostgres
	metadata, err := services.FetchDatabaseMetadata(engineType, dbParams)
	if err != nil {
		log.Printf("警告: 获取数据库元数据失败: %v", err)
		log.Println("将跳过需要元数据的规则")
		metadata = nil
	} else {
		fmt.Printf("✅ 成功获取数据库元数据，Schema 数量: %d\n\n", len(metadata.Schemas))
	}

	// 3. 加载规则（包括需要元数据的规则）
	hasMetadata := (metadata != nil)
	rules, err := services.LoadRules("", engineType, hasMetadata)
	if err != nil {
		log.Fatalf("加载规则失败: %v", err)
	}

	fmt.Printf("✅ 加载了 %d 条审核规则\n\n", len(rules))

	// 4. 要审核的 SQL（修改现有表）
	sql := `
-- 添加新列
ALTER TABLE mydata.users ADD COLUMN age INT NOT NULL;

-- 修改列类型
ALTER TABLE mydata.users ALTER COLUMN username TYPE VARCHAR(20);

-- 删除列
ALTER TABLE mydata.users DROP COLUMN email;
`

	// 5. 构建带元数据的审核请求
	req := &advisor.ReviewRequest{
		Engine:          engineType,
		Statement:       sql,
		CurrentDatabase: "mydb",
		Rules:           rules,
		DBSchema:        metadata, // 提供元数据
	}

	// 6. 执行审核
	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		log.Fatalf("SQL 审核失败: %v", err)
	}

	// 7. 使用 services 包输出格式化结果
	fmt.Println("=== 审核结果（表格格式） ===")
	if err := services.OutputResults(resp, sql, engineType, "table", dbParams); err != nil {
		log.Printf("输出结果失败: %v", err)
	}

	// 8. 决策建议
	fmt.Println("\n=== 决策建议 ===")
	if resp.HasError {
		fmt.Println("❌ 存在错误级别问题，强烈建议修复后再执行")
	} else if resp.HasWarning {
		fmt.Println("⚠️  存在警告级别问题，建议评估风险")
	} else {
		fmt.Println("✅ 审核通过，可以安全执行")
	}
}
```

**传统方式（直接使用 db 包）**:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"advisorTool/db"
	"advisorTool/pkg/advisor"
)

func main() {
	// 1. 数据库连接配置
	dbConfig := &db.ConnectionConfig{
		DbType:   "postgres",
		Host:     "127.0.0.1",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DbName:   "mydb",
		SSLMode:  "disable",
		Timeout:  10,
	}

	// 2. 连接数据库并获取元数据
	ctx := context.Background()
	conn, err := db.OpenConnection(ctx, dbConfig)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer conn.Close()

	metadata, err := db.GetDatabaseMetadata(ctx, conn, dbConfig)
	if err != nil {
		log.Fatalf("获取数据库元数据失败: %v", err)
	}

	// 3. 要审核的 SQL
	sql := `
ALTER TABLE mydata.users ADD COLUMN age INT NOT NULL;
ALTER TABLE mydata.users ALTER COLUMN username TYPE VARCHAR(20);
ALTER TABLE mydata.users DROP COLUMN email;
`

	// 4. 配置审核规则
	rules := []*advisor.SQLReviewRule{
		advisor.NewRule(
			string(advisor.SchemaRuleStatementRequireWhereForUpdateDelete),
			advisor.RuleLevelError,
		),
		advisor.NewRule(
			string(advisor.SchemaRuleColumnNotNull),
			advisor.RuleLevelWarning,
		),
		advisor.NewRule(
			string(advisor.SchemaRuleSchemaBackwardCompatibility),
			advisor.RuleLevelError,
		),
	}

	// 5. 构建审核请求
	req := &advisor.ReviewRequest{
		Engine:          advisor.EnginePostgres,
		Statement:       sql,
		CurrentDatabase: "mydb",
		Rules:           rules,
		DBSchema:        metadata,
	}

	// 6. 执行审核
	resp, err := advisor.SQLReviewCheck(ctx, req)
	if err != nil {
		log.Fatalf("SQL 审核失败: %v", err)
	}

	// 7. 输出结果
	fmt.Printf("发现 %d 个问题\n", len(resp.Advices))
	for _, advice := range resp.Advices {
		fmt.Printf("- [%s] %s\n", advice.Title, advice.Content)
	}
}
```

#### 使用自定义规则配置

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"advisorTool/pkg/advisor"
)

func main() {
	// 1. 使用 Payload 配置规则参数
	
	// 表命名规范：必须是小写字母和下划线，最大长度 63
	tableNamingRule, err := advisor.NewRuleWithPayload(
		string(advisor.SchemaRuleTableNaming),
		advisor.RuleLevelWarning,
		advisor.NamingRulePayload{
			Format:    "^[a-z][a-z0-9_]*$",
			MaxLength: 63,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// 列命名规范
	columnNamingRule, err := advisor.NewRuleWithPayload(
		string(advisor.SchemaRuleColumnNaming),
		advisor.RuleLevelWarning,
		advisor.NamingRulePayload{
			Format:    "^[a-z][a-z0-9_]*$",
			MaxLength: 63,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// 索引命名规范：idx_表名_列名
	idxNamingRule, err := advisor.NewRuleWithPayload(
		string(advisor.SchemaRuleIDXNaming),
		advisor.RuleLevelWarning,
		advisor.NamingRulePayload{
			Format:    "^idx_{{table}}_{{column_list}}$",
			MaxLength: 63,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// 必需列：每个表必须包含这些列
	requiredColumnsRule, err := advisor.NewRuleWithPayload(
		string(advisor.SchemaRuleRequiredColumn),
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"id", "created_at", "updated_at"},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// INSERT 行数限制
	insertRowLimitRule, err := advisor.NewRuleWithPayload(
		string(advisor.SchemaRuleStatementInsertRowLimit),
		advisor.RuleLevelWarning,
		advisor.NumberTypeRulePayload{
			Number: 1000,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// 影响行数限制（UPDATE/DELETE）
	affectedRowLimitRule, err := advisor.NewRuleWithPayload(
		string(advisor.SchemaRuleStatementAffectedRowLimit),
		advisor.RuleLevelWarning,
		advisor.NumberTypeRulePayload{
			Number: 10000,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// 禁止的列类型
	typeDisallowRule, err := advisor.NewRuleWithPayload(
		string(advisor.SchemaRuleColumnTypeDisallowList),
		advisor.RuleLevelError,
		advisor.StringArrayTypeRulePayload{
			List: []string{"money", "xml"},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// VARCHAR 最大长度
	varcharLengthRule, err := advisor.NewRuleWithPayload(
		string(advisor.SchemaRuleColumnMaximumVarcharLength),
		advisor.RuleLevelWarning,
		advisor.NumberTypeRulePayload{
			Number: 2000,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// 2. 组合所有规则
	rules := []*advisor.SQLReviewRule{
		tableNamingRule,
		columnNamingRule,
		idxNamingRule,
		requiredColumnsRule,
		insertRowLimitRule,
		affectedRowLimitRule,
		typeDisallowRule,
		varcharLengthRule,
		
		// 其他基础规则
		advisor.NewRule(
			string(advisor.SchemaRuleStatementNoSelectAll),
			advisor.RuleLevelWarning,
		),
		advisor.NewRule(
			string(advisor.SchemaRuleStatementRequireWhereForUpdateDelete),
			advisor.RuleLevelError,
		),
		advisor.NewRule(
			string(advisor.SchemaRuleTableRequirePK),
			advisor.RuleLevelError,
		),
	}

	// 3. 测试 SQL
	sql := `
CREATE TABLE UserProfile (  -- 表名不符合命名规范（应该是 user_profile）
    user_id SERIAL PRIMARY KEY,
    UserName VARCHAR(3000),  -- 列名不符合规范，VARCHAR 长度超限
    balance MONEY,           -- 使用了禁止的 money 类型
    notes TEXT
    -- 缺少 created_at 和 updated_at 列
);

CREATE INDEX user_idx ON UserProfile(user_id);  -- 索引名不符合规范

SELECT * FROM UserProfile;  -- 禁止 SELECT *
`

	// 4. 执行审核
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

	// 5. 输出结果（JSON 格式）
	type Result struct {
		TotalIssues int               `json:"total_issues"`
		HasError    bool              `json:"has_error"`
		HasWarning  bool              `json:"has_warning"`
		Issues      []IssueDetail     `json:"issues"`
	}

	type IssueDetail struct {
		Severity string `json:"severity"`
		Rule     string `json:"rule"`
		Title    string `json:"title"`
		Message  string `json:"message"`
		Line     int32  `json:"line"`
		Column   int32  `json:"column"`
	}

	result := Result{
		TotalIssues: len(resp.Advices),
		HasError:    resp.HasError,
		HasWarning:  resp.HasWarning,
		Issues:      make([]IssueDetail, 0),
	}

	for _, advice := range resp.Advices {
		severity := "info"
		if advice.Status == advisor.AdviceStatusError {
			severity = "error"
		} else if advice.Status == advisor.AdviceStatusWarning {
			severity = "warning"
		}

		issue := IssueDetail{
			Severity: severity,
			Rule:     fmt.Sprintf("code-%d", advice.Code),
			Title:    advice.Title,
			Message:  advice.Content,
		}
		
		if advice.StartPosition != nil {
			issue.Line = advice.StartPosition.Line
			issue.Column = advice.StartPosition.Column
		}

		result.Issues = append(result.Issues, issue)
	}

	// 输出 JSON
	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
}
```

#### 批量审核多个 SQL 文件

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"advisorTool/pkg/advisor"
	"advisorTool/services"
)

func main() {
	// 1. 使用 services 包加载规则
	engineType := advisor.EnginePostgres
	rules, err := services.LoadRules("", engineType, false)
	if err != nil {
		log.Fatalf("加载规则失败: %v", err)
	}

	// 2. 扫描 SQL 文件目录
	sqlDir := "./migrations"
	files, err := filepath.Glob(filepath.Join(sqlDir, "*.sql"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("发现 %d 个 SQL 文件，开始审核...\n\n", len(files))

	totalIssues := 0
	failedFiles := 0

	// 3. 遍历审核每个文件
	for _, file := range files {
		fmt.Printf("📄 审核文件: %s\n", filepath.Base(file))

		// 读取 SQL 文件
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("   ❌ 读取失败: %v\n\n", err)
			continue
		}

		// 执行审核
		req := &advisor.ReviewRequest{
			Engine:          engineType,
			Statement:       string(content),
			CurrentDatabase: "mydb",
			Rules:           rules,
		}

		resp, err := advisor.SQLReviewCheck(context.Background(), req)
		if err != nil {
			fmt.Printf("   ❌ 审核失败: %v\n\n", err)
			continue
		}

		// 统计问题
		if len(resp.Advices) == 0 {
			fmt.Printf("   ✅ 通过\n\n")
		} else {
			totalIssues += len(resp.Advices)
			if resp.HasError {
				failedFiles++
			}

			fmt.Printf("   发现 %d 个问题:\n", len(resp.Advices))
			for _, advice := range resp.Advices {
				icon := "⚠️ "
				if advice.Status == advisor.AdviceStatusError {
					icon = "❌"
				}
				fmt.Printf("     %s Line %d: %s\n", 
					icon, advice.StartPosition.GetLine(), advice.Title)
			}
			fmt.Println()
		}
	}

	// 4. 输出总结
	fmt.Println("==================== 审核总结 ====================")
	fmt.Printf("总文件数: %d\n", len(files))
	fmt.Printf("发现问题: %d\n", totalIssues)
	fmt.Printf("不通过的文件: %d\n", failedFiles)
	
	if failedFiles > 0 {
		fmt.Println("\n❌ 存在不符合规范的 SQL 文件，请修复后重新提交")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ 所有 SQL 文件审核通过！")
	}
}
```

#### 集成到 CI/CD 流程

```go
// ci_check.go - 用于 CI/CD 流程的 SQL 审核脚本
package main

import (
	"context"
	"fmt"
	"os"

	"advisorTool/pkg/advisor"
	"advisorTool/services"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("用法: ci_check <engine> <sql-file> [config-file]")
		fmt.Println("示例: ci_check mysql migration.sql strict-review.yaml")
		os.Exit(1)
	}

	engineStr := os.Args[1]
	sqlFile := os.Args[2]
	configFile := ""
	if len(os.Args) > 3 {
		configFile = os.Args[3]
	}

	// 读取 SQL 文件
	content, err := os.ReadFile(sqlFile)
	if err != nil {
		fmt.Printf("❌ 读取文件失败: %v\n", err)
		os.Exit(1)
	}

	// 解析数据库引擎类型
	engineType := advisor.EngineFromString(engineStr)
	if engineType == 0 {
		fmt.Printf("❌ 不支持的数据库引擎: %s\n", engineStr)
		os.Exit(1)
	}

	// 使用 services 包加载规则
	rules, err := services.LoadRules(configFile, engineType, false)
	if err != nil {
		fmt.Printf("❌ 加载规则失败: %v\n", err)
		os.Exit(1)
	}

	// 执行审核
	req := &advisor.ReviewRequest{
		Engine:          engineType,
		Statement:       string(content),
		CurrentDatabase: os.Getenv("DB_NAME"),
		Rules:           rules,
	}

	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		fmt.Printf("❌ 审核失败: %v\n", err)
		os.Exit(1)
	}

	// 输出 GitHub Actions 格式的错误信息
	for _, advice := range resp.Advices {
		level := "warning"
		if advice.Status == advisor.AdviceStatusError {
			level = "error"
		}
		
		// GitHub Actions annotation format
		fmt.Printf("::%s file=%s,line=%d,col=%d::%s - %s\n",
			level,
			sqlFile,
			advice.StartPosition.GetLine(),
			advice.StartPosition.GetColumn(),
			advice.Title,
			advice.Content,
		)
	}

	if resp.HasError {
		fmt.Printf("\n❌ SQL 审核失败，发现 %d 个错误\n", len(resp.Advices))
		os.Exit(2)
	}

	if resp.HasWarning {
		fmt.Printf("\n⚠️  SQL 审核发现 %d 个警告\n", len(resp.Advices))
		os.Exit(1)
	}

	fmt.Printf("✅ SQL 审核通过\n")
}
```

#### 从配置文件加载规则

使用 `services` 包可以轻松从 YAML 或 JSON 配置文件加载规则：

```go
package main

import (
	"context"
	"log"

	"advisorTool/pkg/advisor"
	"advisorTool/services"
)

func main() {
	// 从配置文件加载规则
	rules, err := services.LoadRules("mysql-review-config.yaml", advisor.EngineMySQL, false)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 如果不提供配置文件路径，将使用默认规则
	// rules, err := services.LoadRules("", advisor.EngineMySQL, false)

	sql := "SELECT * FROM users WHERE id = 1"
	
	req := &advisor.ReviewRequest{
		Engine:          advisor.EngineMySQL,
		Statement:       sql,
		CurrentDatabase: "mydb",
		Rules:           rules,
	}

	resp, err := advisor.SQLReviewCheck(context.Background(), req)
	if err != nil {
		log.Fatalf("审核失败: %v", err)
	}

	// 使用 services 包输出结果
	services.OutputResults(resp, sql, advisor.EngineMySQL, "table", nil)
}
```

#### 生成示例配置文件

```go
package main

import (
	"fmt"

	"advisorTool/pkg/advisor"
	"advisorTool/services"
)

func main() {
	// 为不同数据库生成示例配置
	mysqlConfig := services.GenerateSampleConfig(advisor.EngineMySQL)
	fmt.Println("=== MySQL 示例配置 ===")
	fmt.Println(mysqlConfig)
	
	pgConfig := services.GenerateSampleConfig(advisor.EnginePostgres)
	fmt.Println("\n=== PostgreSQL 示例配置 ===")
	fmt.Println(pgConfig)
}
```

更多规则类型和配置，请参考 [配置文件格式](#配置文件格式) 和 [可用规则列表](#可用规则列表) 章节。

**高级用法 - 使用 Payload 配置**:

```go
// 命名规范规则（带参数）
namingRule, _ := advisor.NewRuleWithPayload(
    advisor.RuleTableNaming,
    advisor.RuleLevelWarning,
    advisor.NamingRulePayload{
        Format:    "^[a-z][a-z0-9_]*$",  // 正则表达式
        MaxLength: 64,
    },
)

// 数值限制规则
limitRule, _ := advisor.NewRuleWithPayload(
    advisor.RuleStatementInsertRowLimit,
    advisor.RuleLevelWarning,
    advisor.NumberTypeRulePayload{
        Number: 1000,  // 单次 INSERT 最多 1000 行
    },
)

// 类型禁用规则
typeRule, _ := advisor.NewRuleWithPayload(
    advisor.RuleColumnTypeDisallowList,
    advisor.RuleLevelError,
    advisor.StringArrayTypeRulePayload{
        List: []string{"BLOB", "LONGBLOB", "TEXT"},
    },
)

rules := []*advisor.SQLReviewRule{namingRule, limitRule, typeRule}
```

## 配置文件格式

### YAML 格式示例

```yaml
name: mysql-review-config
rules:
  # 基础规则
  - type: statement.select.no-select-all
    level: WARNING
    comment: 禁止使用 SELECT *
    
  - type: statement.where.require.update-delete
    level: ERROR
    comment: UPDATE/DELETE 必须包含 WHERE 子句
    
  - type: table.require-pk
    level: ERROR
    comment: 表必须有主键
    
  # 带参数的规则
  - type: naming.table
    level: WARNING
    payload: '{"format":"^[a-z][a-z0-9_]*$","maxLength":64}'
    comment: 表名必须使用小写字母和下划线
    
  - type: column.required
    level: ERROR
    payload: '{"list":["id","created_at","updated_at"]}'
    comment: 每个表必须包含指定列
    
  - type: statement.insert.row-limit
    level: WARNING
    payload: '{"number":1000}'
    comment: 限制单次 INSERT 行数
    
  - type: system.charset.allowlist
    level: WARNING
    payload: '{"list":["utf8mb4","utf8"]}'
    comment: 只允许使用指定字符集
```

### 规则级别

| 级别 | 说明 | 退出码 |
|------|------|--------|
| `ERROR` | 错误级别，必须修复 | 2 |
| `WARNING` | 警告级别，建议修复 | 1 |
| `DISABLED` | 禁用此规则 | - |

### Payload 配置类型

不同规则支持不同的 Payload 类型：

**1. 命名规则 (NamingRulePayload)**:
```json
{
  "format": "^[a-z][a-z0-9_]*$",  // 正则表达式
  "maxLength": 64                  // 最大长度
}
```

**2. 数值规则 (NumberTypeRulePayload)**:
```json
{
  "number": 1000  // 数值限制
}
```

**3. 字符串数组规则 (StringArrayTypeRulePayload)**:
```json
{
  "list": ["utf8mb4", "utf8"]  // 允许或禁止的列表
}
```

**4. 注释规范规则 (CommentConventionRulePayload)**:
```json
{
  "required": true,    // 是否必需
  "maxLength": 256     // 最大长度
}
```

## 支持的解析器

本工具使用 Bytebase 原有的解析器，基于 ANTLR4：

| 数据库 | 解析器包 | 语法规范 |
|--------|----------|----------|
| MySQL | `github.com/bytebase/parser/mysql` | MySQL 8.0 语法 |
| MariaDB | `github.com/bytebase/parser/mysql` | 兼容 MySQL 语法 |
| PostgreSQL | `github.com/bytebase/parser/postgresql` | PostgreSQL 14+ 语法 |
| Oracle | `github.com/bytebase/parser/plsql` | Oracle PL/SQL |
| SQL Server | `github.com/bytebase/parser/tsql` | T-SQL |
| TiDB | `github.com/pingcap/tidb/parser` | TiDB 原生解析器 |
| Snowflake | `github.com/bytebase/parser/snowflake` | Snowflake SQL |
| OceanBase | `github.com/bytebase/parser/mysql` | 兼容 MySQL 模式 |

## 支持的审核规则

### Engine 引擎规则
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

**基础检查**:
- `statement.select.no-select-all` - 禁止使用 SELECT *
- `statement.where.require.select` - SELECT 必须包含 WHERE
- `statement.where.require.update-delete` - UPDATE/DELETE 必须包含 WHERE ⭐
- `statement.where.no-leading-wildcard-like` - 禁止前导通配符 LIKE
- `statement.where.no-equal-null` - 禁止使用 WHERE col = NULL
- `statement.where.disallow-functions` - 禁止在 WHERE 中使用函数

**DML 规则**:
- `statement.insert.must-specify-column` - INSERT 必须指定列名
- `statement.insert.disallow-order-by-rand` - 禁止 ORDER BY RAND
- `statement.insert.row-limit` - INSERT 行数限制
- `statement.affected-row-limit` - 影响行数限制
- `statement.dml-dry-run` - DML 空运行验证

**DDL 规则**:
- `statement.merge-alter-table` - 合并 ALTER TABLE 语句
- `statement.disallow-add-column-with-default` - 禁止 ADD COLUMN 带默认值（PostgreSQL）
- `statement.add-check-not-valid` - CHECK 约束必须 NOT VALID（PostgreSQL）
- `statement.disallow-add-not-null` - 禁止添加 NOT NULL（PostgreSQL）
- `statement.add-fk-not-valid` - 外键必须 NOT VALID（PostgreSQL）
- `statement.create-specify-schema` - 创建时指定 schema

**性能和限制**:
- `statement.disallow-commit` - 禁止 COMMIT 语句
- `statement.disallow-limit` - 禁止 LIMIT 子句
- `statement.disallow-order-by` - 禁止 ORDER BY 子句
- `statement.disallow-cross-db-queries` - 禁止跨库查询（MSSQL）
- `statement.select.full-table-scan` - 禁止全表扫描
- `statement.disallow-using-filesort` - 禁止文件排序
- `statement.disallow-using-temporary` - 禁止临时表
- `statement.query-minimum-plan-level` - 最低查询计划级别
- `statement.maximum-limit-value` - 最大 LIMIT 值
- `statement.maximum-join-table-count` - 最大 JOIN 表数
- `statement.maximum-statements-in-transaction` - 事务中最大语句数
- `statement.max-execution-time` - 最大执行时间

**其他**:
- `statement.non-transactional` - 非事务语句检查
- `statement.prior-backup-check` - 变更前备份检查
- `statement.disallow-offline-ddl` - 禁止离线 DDL（OceanBase）

### Table 表规则
- `table.require-pk` - 表必须有主键 ⭐
- `table.no-foreign-key` - 禁止外键
- `table.drop-naming-convention` - 删除表命名规范
- `table.comment` - 表注释规范
- `table.disallow-partition` - 禁止分区表
- `table.disallow-trigger` - 禁止触发器
- `table.no-duplicate-index` - 禁止重复索引
- `table.disallow-ddl` - 禁止特定表的 DDL 操作
- `table.disallow-dml` - 禁止特定表的 DML 操作
- `table.limit-size` - 限制表大小
- `table.text-fields-total-length` - 文本字段总长度限制
- `table.disallow-set-charset` - 禁止设置表字符集
- `table.require-charset` - 要求指定字符集
- `table.require-collation` - 要求指定排序规则

### Column 列规则

**基础规则**:
- `column.required` - 必需列
- `column.no-null` - 禁止 NULL 值
- `column.require-default` - 列必须有默认值
- `column.set-default-for-not-null` - NOT NULL 列需要默认值
- `column.add-not-null-column-require-default` - 添加 NOT NULL 列需要默认值

**变更控制**:
- `column.disallow-change-type` - 禁止改变列类型
- `column.disallow-change` - 禁止 CHANGE COLUMN
- `column.disallow-changing-order` - 禁止改变列顺序
- `column.disallow-drop` - 禁止 DROP COLUMN
- `column.disallow-drop-in-index` - 禁止删除索引列

**类型和长度**:
- `column.type-disallow-list` - 列类型黑名单
- `column.maximum-character-length` - CHAR 最大长度
- `column.maximum-varchar-length` - VARCHAR 最大长度

**自增列**:
- `column.auto-increment-must-integer` - 自增列必须为整数
- `column.auto-increment-must-unsigned` - 自增列必须无符号
- `column.auto-increment-initial-value` - 自增列初始值

**其他**:
- `column.comment` - 列注释规范
- `column.disallow-set-charset` - 禁止设置列字符集
- `column.default-disallow-volatile` - 禁止易变的默认值
- `column.current-time-count-limit` - 当前时间列数量限制
- `column.require-charset` - 要求指定字符集
- `column.require-collation` - 要求指定排序规则

### Index 索引规则
- `index.no-duplicate-column` - 禁止重复列
- `index.key-number-limit` - 索引键数量限制
- `index.pk-type-limit` - 主键类型限制
- `index.type-no-blob` - 禁止 BLOB/TEXT 索引
- `index.total-number-limit` - 索引总数限制
- `index.primary-key-type-allowlist` - 主键类型白名单
- `index.create-concurrently` - 并发创建索引（PostgreSQL）
- `index.type-allowlist` - 索引类型白名单
- `index.not-redundant` - 禁止冗余索引

### Schema 模式规则
- `schema.backward-compatibility` - 向后兼容性检查 ⭐

### Database 数据库规则
- `database.drop-empty-database` - 只能删除空数据库

### System 系统规则
- `system.charset.allowlist` - 字符集白名单
- `system.collation.allowlist` - 排序规则白名单
- `system.comment.length` - 注释长度限制
- `system.procedure.disallow-create` - 禁止创建存储过程
- `system.function.disallow-create` - 禁止创建函数
- `system.event.disallow-create` - 禁止创建事件
- `system.view.disallow-create` - 禁止创建视图
- `system.function.disallow-list` - 函数黑名单

**标注说明**: ⭐ 表示核心规则，建议在所有环境中启用

## 输出格式

### 1. Text 格式（默认）

```
Found 2 issue(s):

1. ❌ [ERROR] statement.where.require.update-delete
   "DELETE FROM orders" requires WHERE clause
   Location: line 2, column 0

2. ⚠️ [WARNING] statement.select.no-select-all
   "SELECT * FROM users" uses SELECT all
   Location: line 1, column 0
```

### 2. JSON 格式（兼容 Inception）

```json
[
  {
    "order_id": 1,
    "stage": "CHECKED",
    "error_level": "2",
    "stage_status": "Audit Completed",
    "error_message": "[statement.where.require.update-delete] \"DELETE FROM orders\" requires WHERE clause",
    "sql": "DELETE FROM orders",
    "affected_rows": 0,
    "sequence": "0_0_00000000",
    "backup_dbname": "",
    "execute_time": "0",
    "sqlsha1": "",
    "backup_time": "0"
  }
]
```

**错误级别说明**:
- `0`: 无问题
- `1`: 警告
- `2`: 错误

### 3. YAML 格式

```yaml
advices:
  - status: ERROR
    code: 201
    title: statement.where.require.update-delete
    content: '"DELETE FROM orders" requires WHERE clause'
    startPosition:
      line: 2
      column: 0
hasError: true
hasWarning: false
```

## 项目结构

```
advisorTool/
├── advisor/                          # Bytebase 原有审核规则实现
│   ├── advisor.go                    # 核心接口定义
│   ├── builtin_rules.go              # 内置规则
│   ├── code/                         # 错误码定义
│   ├── mysql/                        # MySQL 规则实现（50+ 规则）
│   ├── pg/                           # PostgreSQL 规则实现（40+ 规则）
│   ├── oracle/                       # Oracle 规则实现
│   ├── mssql/                        # SQL Server 规则实现
│   ├── tidb/                         # TiDB 规则实现
│   ├── snowflake/                    # Snowflake 规则实现
│   └── oceanbase/                    # OceanBase 规则实现
├── cmd/
│   └── advisor/
│       └── main.go                   # 命令行入口（190 行）
│           ├── 参数解析
│           ├── SQL 输入处理
│           ├── 规则配置加载
│           ├── 数据库元数据获取
│           ├── 审核执行
│           └── 结果输出（支持多种格式）
├── services/                        # 公共服务包（可被外部引用）
│   ├── config.go                     # 规则配置加载（340 行）
│   │   ├── LoadRules()               # 加载规则
│   │   ├── GetDefaultRules()         # 获取默认规则
│   │   └── GenerateSampleConfig()    # 生成示例配置
│   ├── output.go                     # 结果输出（275 行）
│   │   ├── OutputResults()           # 格式化输出
│   │   └── ListAvailableRules()      # 列出可用规则
│   ├── result.go                     # 结果处理（253 行）
│   │   ├── ConvertToReviewResults()  # 转换结果格式
│   │   ├── CalculateAffectedRows()   # 计算影响行数
│   │   └── DBConnectionParams        # 连接参数结构
│   ├── metadata.go                   # 元数据获取（49 行）
│   │   └── FetchDatabaseMetadata()   # 获取数据库元数据
│   └── README.md                     # Services 包文档
├── pkg/
│   └── advisor/
│       ├── advisor.go                # 封装层 API（247 行）
│       │   ├── SQLReviewCheck()      # 主入口函数
│       │   ├── EngineFromString()    # 引擎类型转换
│       │   └── NewRule*()            # 规则构建函数
│       └── rules.go                  # 规则常量定义（380 行）
│           ├── 90+ 规则类型常量
│           ├── AllRules()            # 返回所有规则
│           └── GetRuleDescription()  # 规则描述
├── db/                               # 数据库连接和元数据获取
│   ├── connection.go                 # 连接管理
│   └── metadata.go                   # 元数据提取
├── examples/
│   ├── mysql-review-config.yaml      # MySQL 完整配置（245 行）
│   ├── postgres-review-config.yaml   # PostgreSQL 配置
│   ├── basic-config.yaml             # 基础配置（无需元数据）
│   ├── test.sql                      # 测试 SQL
│   ├── postgres_library_example.go   # PostgreSQL 库使用示例
│   └── external_usage_example.go     # 外部程序使用示例
├── build/
│   └── advisor                       # 编译输出
├── go.mod                            # Go 模块（含 replace 指令）
├── go.sum                            # 依赖校验
├── Makefile                          # 编译脚本
└── README.md                         # 本文档
```

## 规则分类与使用建议

### 静态分析规则（推荐）

**优点**: 无需数据库连接，快速审核，适合 CI/CD 集成

**通用规则**（所有数据库）:
```yaml
rules:
  - type: statement.where.require.update-delete
    level: ERROR
  - type: table.require-pk
    level: ERROR
  - type: statement.select.no-select-all
    level: WARNING
  - type: table.no-foreign-key
    level: WARNING
```

**MySQL 特有规则**:
```yaml
rules:
  - type: engine.mysql.use-innodb
    level: ERROR
  - type: column.auto-increment-must-integer
    level: ERROR
  - type: column.auto-increment-must-unsigned
    level: WARNING
  - type: index.no-duplicate-column
    level: ERROR
```

**PostgreSQL 特有规则**:
```yaml
rules:
  - type: statement.disallow-add-column-with-default
    level: WARNING
  - type: statement.add-check-not-valid
    level: WARNING
  - type: index.create-concurrently
    level: ERROR
  - type: statement.create-specify-schema
    level: WARNING
```

### 动态分析规则（需谨慎）

**需要**: 提供 `-host`、`-port`、`-user`、`-password`、`-dbname` 参数

**元数据依赖规则**:
```yaml
rules:
  - type: column.no-null              # 需要现有表结构
    level: WARNING
  - type: column.disallow-drop-in-index  # 需要索引信息
    level: ERROR
  - type: schema.backward-compatibility  # 需要变更前后对比
    level: ERROR
  - type: index.not-redundant          # 需要现有索引
    level: WARNING
```

## 典型使用场景

### 场景 1: CI/CD 集成

```bash
#!/bin/bash
# pre-deploy-check.sh

# 检查 SQL 变更脚本
./advisor -engine mysql \
  -config production-review.yaml \
  -file migration.sql \
  -format json > review-result.json

# 根据退出码决定是否继续部署
if [ $? -eq 2 ]; then
  echo "❌ SQL 审核失败，发现错误级别问题"
  exit 1
elif [ $? -eq 1 ]; then
  echo "⚠️ SQL 审核发现警告，需人工确认"
  exit 1
else
  echo "✅ SQL 审核通过"
  exit 0
fi
```

### 场景 2: 开发环境快速检查

```bash
# 快速检查本地 SQL 文件
./advisor -engine postgres -file my-changes.sql

# 使用宽松的规则集
./advisor -engine mysql -config basic-config.yaml -file test.sql
```

### 场景 3: 生产环境审核（带元数据）

```bash
# 连接生产数据库进行全面审核
./advisor -engine mysql \
  -host prod-db.example.com \
  -port 3306 \
  -user readonly_user \
  -password ${DB_PASSWORD} \
  -dbname production \
  -config strict-review.yaml \
  -file hotfix.sql \
  -format json
```

### 场景 4: IDE 集成

在 VSCode、IntelliJ 等 IDE 中配置为外部工具：

```json
{
  "name": "SQL Review",
  "command": "/path/to/advisor",
  "args": [
    "-engine", "mysql",
    "-sql", "${selectedText}"
  ]
}
```

## 常见问题（FAQ）

### Q1: 如何选择合适的规则？

**答**: 根据环境和需求分级启用：

- **开发环境**: 使用 `basic-config.yaml`，只启用核心规则
- **测试环境**: 启用大部分 WARNING 规则，ERROR 规则保持严格
- **生产环境**: 严格模式，所有 ERROR 规则必须通过

### Q2: 某些规则报错但我认为合理，如何处理？

**答**: 三种方式：

1. 在配置文件中将该规则设为 `DISABLED`
2. 修改规则级别为 `WARNING`
3. 添加 `comment` 字段说明例外情况

### Q3: 如何添加自定义规则？

**答**: 实现 `Advisor` 接口并注册：

```go
package myrule

import (
    "context"
    "advisorTool/advisor"
    storepb "advisorTool/generated-go/store"
)

type MyCustomAdvisor struct{}

func (a *MyCustomAdvisor) Check(ctx context.Context, checkCtx advisor.Context) ([]*storepb.Advice, error) {
    // 实现检查逻辑
    return advices, nil
}

func init() {
    advisor.Register(storepb.Engine_MYSQL, "my.custom.rule", &MyCustomAdvisor{})
}
```

### Q4: 性能如何？能处理大型 SQL 文件吗？

**答**: 
- 静态分析：单个文件（1000 行 SQL）约 100-500ms
- 动态分析：取决于数据库响应时间
- 建议：超过 10000 行的 SQL 文件建议分批审核

### Q5: 与 Inception 的区别？

**答**:

| 特性 | SQL Advisor Tool | Inception |
|------|------------------|-----------|
| 解析器 | ANTLR4（精确） | 自定义解析器 |
| 规则数量 | 90+ | 30+ |
| 数据库支持 | 8 种 | 主要 MySQL |
| 可扩展性 | 高（插件化） | 中 |
| 输出格式 | JSON 兼容 Inception | JSON |

## 依赖说明

本工具有独立的 `go.mod` 文件，使用 `replace` 指令引用本地 Bytebase 代码：

```go
// go.mod
module advisorTool

go 1.23

replace github.com/bytebase/bytebase => ../..

require (
    github.com/bytebase/bytebase v0.0.0
    github.com/antlr4-go/antlr/v4 v4.13.0
    github.com/pingcap/tidb/parser v0.0.0
    // ... 其他依赖
)
```

**设计优势**:
1. ✅ **独立编译**: 可在 advisorTool 目录直接 `go build`
2. ✅ **依赖一致**: 通过 replace 确保与主项目版本一致
3. ✅ **完整功能**: 使用 Bytebase 原有解析器和规则
4. ✅ **易于维护**: 主项目更新时同步 go.mod

## 与 Bytebase 的关系

本工具是 **Bytebase SQL 审核引擎的命令行封装**。

**Bytebase** 是一个开源的数据库 DevOps 平台，提供：
- 🌐 Web UI 界面
- 👥 团队协作和权限管理
- 📋 变更工作流和审批
- 📊 SQL 审核引擎（本工具使用的核心）
- 🔄 数据库版本控制

**如果你需要**:
- ✅ 命令行工具 → 使用本工具
- ✅ CI/CD 集成 → 使用本工具
- ✅ 快速审核 SQL → 使用本工具
- ✅ 完整的数据库管理平台 → 使用 Bytebase

## 贡献指南

欢迎贡献代码、报告 Bug 或建议新功能！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

遵循 Bytebase 项目的许可证。

## 相关链接

- [Bytebase 官网](https://www.bytebase.com)
- [Bytebase GitHub](https://github.com/bytebase/bytebase)
- [SQL Review 文档](https://www.bytebase.com/docs/sql-review/overview)
- [审核规则文档](https://www.bytebase.com/docs/sql-review/review-rules)
- [ANTLR4 官网](https://www.antlr.org/)

## 更新日志

### v1.0.0 (2024)
- ✅ 初始版本发布
- ✅ 支持 8 种数据库引擎
- ✅ 实现 90+ 审核规则
- ✅ 支持多种输出格式
- ✅ 支持数据库元数据获取
- ✅ 兼容 Inception JSON 格式

---

**Star ⭐ 本项目** 如果你觉得有用！

有问题或建议？欢迎提 [Issue](https://github.com/your-repo/advisorTool/issues)！
