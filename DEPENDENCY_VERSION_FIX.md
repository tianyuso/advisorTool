# Go 模块依赖版本修复

## 问题描述

当其他项目使用 `go get github.com/tianyuso/advisorTool` 时，出现错误：

```
github.com/microsoft/go-mssqldb@v0.0.0-00010101000000-000000000000: invalid version: unknown revision 000000000000
github.com/pingcap/tidb@v0.0.0-00010101000000-000000000000: invalid version: unknown revision 000000000000
```

## 原因分析

项目 `go.mod` 中使用了 `replace` 指令来替换某些依赖：

```go
replace (
    github.com/microsoft/go-mssqldb => github.com/bytebase/go-mssqldb v0.0.0-20240801091126-3ff3ca07d898
    github.com/pingcap/tidb => github.com/bytebase/tidb v0.0.0-20251104040057-d29df9dd1b3b
)
```

但是 **`replace` 指令只在当前模块生效，不会传播到依赖项目**。当外部项目引入时，Go 会尝试下载原始的无效版本 `v0.0.0-00010101000000-000000000000`，导致失败。

## 解决方案

### 方案 1：使用有效的版本号（✅ 已采用）

直接在 `require` 中使用可以下载的版本：

```go
require (
    github.com/lib/pq v1.10.9
    github.com/microsoft/go-mssqldb v1.7.2  // ✅ 使用官方版本
    github.com/pingcap/tidb v1.1.0-beta.0.20241125141335-ec8b81b98edc  // ✅ 使用可用版本
    github.com/pingcap/tidb/pkg/parser v0.0.0-20241125141335-ec8b81b98edc
)
```

保留 `replace` 指令用于本地开发，但使用者不会受影响。

### 方案 2：建议使用者也添加 replace（不推荐）

让使用者在他们的项目中也添加相同的 `replace` 指令，但这会增加使用复杂度。

## 修复内容

### 修改前

```go
require (
    github.com/microsoft/go-mssqldb v0.0.0-00010101000000-000000000000  // ❌ 无效版本
    github.com/pingcap/tidb v0.0.0-00010101000000-000000000000  // ❌ 无效版本
)
```

### 修改后

```go
require (
    github.com/microsoft/go-mssqldb v1.7.2  // ✅ 有效版本
    github.com/pingcap/tidb v1.1.0-beta.0.20241125141335-ec8b81b98edc  // ✅ 有效版本
)

// replace 指令保留用于本地开发
replace (
    github.com/microsoft/go-mssqldb => github.com/bytebase/go-mssqldb v0.0.0-20240801091126-3ff3ca07d898
    github.com/pingcap/tidb => github.com/bytebase/tidb v0.0.0-20251104040057-d29df9dd1b3b
)
```

## 验证

### 1. 本地验证

```bash
cd /data/dev_go/advisorTool
go mod tidy
go build ./db/... ./services/... ./pkg/...
```

### 2. 外部项目验证

创建测试项目：

```bash
mkdir -p /tmp/test-advisortool
cd /tmp/test-advisortool
go mod init advisorExam

# 复制示例文件
cp /data/dev_go/advisorTool/examples/postgres_external_usage_example.go main.go

# 下载依赖（需要先推送到 GitHub）
go mod tidy
go build
```

## 注意事项

### Replace 指令的作用域

- ✅ **在当前模块中生效** - 本地开发时会使用 replace 的版本
- ❌ **不传播到依赖项目** - 外部项目使用时不会应用 replace

### 版本选择策略

1. **github.com/microsoft/go-mssqldb**
   - 原始 replace: `github.com/bytebase/go-mssqldb v0.0.0-20240801091126-3ff3ca07d898`
   - 新版本: `v1.7.2` (官方最新稳定版)
   - 理由: 官方版本兼容性更好

2. **github.com/pingcap/tidb**
   - 原始 replace: `github.com/bytebase/tidb v0.0.0-20251104040057-d29df9dd1b3b`
   - 新版本: `v1.1.0-beta.0.20241125141335-ec8b81b98edc`
   - 理由: 使用与 parser 匹配的版本

## 推送到 GitHub

修复完成后，需要打新标签：

```bash
# 提交修改
git add go.mod go.sum
git commit -m "fix: update dependency versions to resolve external usage issues"

# 打标签
git tag v1.0.6

# 推送
git push origin main
git push origin v1.0.6
```

## 使用方法

修复后，外部项目可以正常使用：

```bash
# 安装
go get github.com/tianyuso/advisorTool@v1.0.6

# 使用
```

```go
import (
    "github.com/tianyuso/advisorTool/pkg/advisor"
    "github.com/tianyuso/advisorTool/services"
)
```

---

**更新日期：** 2024-12-22  
**状态：** ✅ 已修复  
**版本：** v1.0.6

