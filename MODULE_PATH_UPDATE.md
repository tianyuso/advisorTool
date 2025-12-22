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

## 发布新版本

修改完成后，需要打新的 tag 并推送到 GitHub：

```bash
# 提交所有修改
git add .
git commit -m "fix: update module path to github.com/tianyuso/advisorTool"

# 打新标签（递增版本号）
git tag v1.0.5

# 推送到 GitHub
git push origin main
git push origin v1.0.5
```

## 示例文件

更新后的示例文件已经使用正确的导入路径：

- `examples/postgres_external_usage_example.go` ✅
- `examples/library_usage_basic.go` ✅
- `examples/library_usage_all_rules.go` ✅
- `examples/external_usage_example.go` ✅

## 注意事项

1. ✅ **所有 `.go` 文件的导入路径已更新**
2. ✅ **go.mod 模块路径已更新**
3. ✅ **核心包编译成功**
4. ⚠️ **需要推送到 GitHub 并打新标签**
5. ⚠️ **examples 目录中有多个 main 函数，需要单独编译运行**

## 单独编译示例

由于 examples 目录中有多个 `main` 函数，不能同时编译。需要单独运行：

```bash
# 运行 PostgreSQL 示例
cd examples
go run postgres_external_usage_example.go

# 运行其他示例
go run library_usage_basic.go
go run library_usage_all_rules.go
```

---

**更新日期：** 2024-12-22  
**状态：** ✅ 已完成

