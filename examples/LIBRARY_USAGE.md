# SQL Advisor Tool - Go 库使用示例

本目录包含使用 advisorTool 作为 Go 库的各种示例代码。

## 示例列表

### 1. 基础使用示例 (`library_usage_basic.go`)

演示如何使用指定的几个规则进行 SQL 审核。

**运行方式**:
```bash
go run library_usage_basic.go
```

**功能**:
- 定义 3 条基础审核规则
- 审核简单的 SQL 语句
- 输出审核结果和问题详情

---

### 2. 全规则验证示例 (`library_usage_all_rules.go`)

演示如何使用所有可用规则进行全面的 SQL 审核。

**运行方式**:
```bash
go run library_usage_all_rules.go
```

**功能**:
- 自动加载所有可用规则（90+ 条）
- 对包含多种违规的 SQL 进行全面审核
- 详细的统计和分类输出
- 按规则类型汇总问题

---

### 3. Payload 配置示例 (`library_usage_with_payload.go`)

演示如何为规则配置参数（Payload）。

**运行方式**:
```bash
go run library_usage_with_payload.go
```

**功能**:
- 配置表命名规范（正则表达式）
- 配置 INSERT 行数限制
- 配置列类型黑名单
- 配置字符集白名单
- 配置必需列
- 配置注释规范

---

## 其他已有示例

### `affected_rows_demo.go`
演示如何使用 affected rows 功能。

### 配置文件示例
- `basic-config.yaml` - 基础配置
- `mysql-review-config.yaml` - MySQL 完整配置
- `postgres-review-config.yaml` - PostgreSQL 配置

### 测试 SQL 文件
- `test.sql` - 通用测试 SQL
- `test-affected-rows.sql` - 影响行数测试 SQL

---

## 使用建议

1. **开发阶段**：使用 `library_usage_basic.go` 的方式，只启用核心规则
2. **测试阶段**：使用 `library_usage_all_rules.go` 的方式，启用所有规则进行全面审核
3. **生产环境**：使用 `library_usage_with_payload.go` 的方式，根据团队规范配置参数

## 集成到项目

### 作为 Go Module 依赖

在你的项目中使用：

```bash
# 在你的项目目录
go mod init your-project
# 添加本地依赖（假设 advisorTool 在上级目录）
go mod edit -replace advisorTool=../advisorTool
```

### 示例代码

```go
import "advisorTool/pkg/advisor"

// 创建规则
rules := []*advisor.SQLReviewRule{
    advisor.NewRule(advisor.RuleStatementNoSelectAll, advisor.RuleLevelWarning),
}

// 创建请求
req := &advisor.ReviewRequest{
    Engine:    advisor.EngineMySQL,
    Statement: "SELECT * FROM users",
    Rules:     rules,
}

// 执行审核
resp, err := advisor.SQLReviewCheck(context.Background(), req)
```

## 注意事项

1. **规则适用性**：不同数据库支持的规则有所不同，某些规则只适用于特定数据库
2. **元数据依赖**：部分规则需要数据库元数据（如 `RuleColumnNotNull`），需要通过 `ReviewRequest.DBSchema` 提供
3. **性能考虑**：启用所有规则会增加审核时间，建议按需选择规则
4. **Payload 格式**：配置 Payload 时必须是有效的 JSON 格式

## 常见问题

### Q: 如何知道某个规则需要什么 Payload？
A: 查看 `advisor/sql_review.go` 中的规则定义，或参考 README.md 中的配置文件格式说明。

### Q: 规则执行出错怎么办？
A: 检查规则是否适用于当前数据库引擎，以及是否提供了必要的元数据。

### Q: 如何自定义规则？
A: 实现 `advisor.Advisor` 接口并通过 `advisor.Register()` 注册。

---

更多信息请参考项目根目录的 [README.md](../README.md)









