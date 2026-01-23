# 影响行数错误处理改进

## 修改内容

### 1. 新增 `AffectedRowsInfo` 结构体 (services/result.go)

```go
// AffectedRowsInfo holds the count and error information for affected rows calculation.
type AffectedRowsInfo struct {
	Count int
	Error string
}
```

### 2. 修改 `CalculateAffectedRowsForStatements` 函数

**之前的签名:**
```go
func CalculateAffectedRowsForStatements(statement string, engineType advisor.Engine, dbParams *DBConnectionParams) map[int]int
```

**修改后的签名:**
```go
func CalculateAffectedRowsForStatements(statement string, engineType advisor.Engine, dbParams *DBConnectionParams) map[int]*AffectedRowsInfo
```

**主要变化:**
- 返回类型从 `map[int]int` 改为 `map[int]*AffectedRowsInfo`
- 不再只保存成功的结果（`err == nil`），现在所有 SQL 语句都会保存结果
- 当 `err != nil` 时，保存 `count` 和 `err.Error()` 到 `AffectedRowsInfo` 结构体

**新逻辑:**
```go
for i, sql := range sqlStatements {
    count, err := db.CalculateAffectedRows(context.Background(), dbConn, sql, engineType)
    info := &AffectedRowsInfo{
        Count: count,
    }
    if err != nil {
        info.Error = err.Error()
    }
    affectedRowsMap[i] = info
}
```

### 3. 修改 `ConvertToReviewResults` 函数

**之前的签名:**
```go
func ConvertToReviewResults(resp *advisor.ReviewResponse, statement string, engineType advisor.Engine, affectedRowsMap map[int]int) []ReviewResult
```

**修改后的签名:**
```go
func ConvertToReviewResults(resp *advisor.ReviewResponse, statement string, engineType advisor.Engine, affectedRowsMap map[int]*AffectedRowsInfo) []ReviewResult
```

**主要变化:**

#### a) 没有审核建议时的处理（len(resp.Advices) == 0）:
- 从 `affectedRowsMap[i]` 获取 `AffectedRowsInfo`
- 如果 `info.Error != ""`, 则：
  - `ErrorMessage` 设置为 `"[AffectedRows] {error message}"`
  - `ErrorLevel` 设置为 `"2"` (错误级别)
- 否则保持 `ErrorLevel = "0"`, `ErrorMessage = ""`

```go
affectedRows := 0
errorMessage := ""
errorLevel := "0"

if info, ok := affectedRowsMap[i]; ok {
    affectedRows = info.Count
    if info.Error != "" {
        errorMessage = fmt.Sprintf("[AffectedRows] %s", info.Error)
        errorLevel = "2"
    }
}
```

#### b) 有审核建议时的处理:
- 从 `affectedRowsMap[i]` 获取 `AffectedRowsInfo`
- 如果 `info.Error != ""`, 则：
  - 将错误信息添加到 `errorMessages` 数组: `"[AffectedRows] {error message}"`
  - 将 `ErrorLevel` 强制设置为 `"2"` (确保显示为错误)

```go
affectedRows := 0
if info, ok := affectedRowsMap[i]; ok {
    affectedRows = info.Count
    if info.Error != "" {
        errorMessages = append(errorMessages, fmt.Sprintf("[AffectedRows] %s", info.Error))
        errorLevel = "2"
    }
}
```

## 优点

1. **完整保留错误信息**: 不再丢失影响行数计算失败的错误信息
2. **统一的错误处理**: 错误信息以 `[AffectedRows]` 前缀显示，与其他审核规则的格式一致
3. **明确的错误级别**: 当影响行数计算失败时，`ErrorLevel` 自动设置为 `"2"` (错误)
4. **向后兼容**: 当 `err == nil` 时，`info.Error` 为空字符串，不影响现有逻辑

## 使用示例

```go
// 计算影响行数（包含错误信息）
affectedRowsMap := services.CalculateAffectedRowsForStatements(sql, engineType, dbParams)

// 打印详细信息
for i, info := range affectedRowsMap {
    if info.Error != "" {
        fmt.Printf("SQL #%d: Count=%d, Error=%s\n", i+1, info.Count, info.Error)
    } else {
        fmt.Printf("SQL #%d: Count=%d, Error=nil\n", i+1, info.Count)
    }
}

// 转换为结构化结果
results := services.ConvertToReviewResults(resp, sql, engineType, affectedRowsMap)
```

## 测试文件

创建了 `examples/test_affected_rows_error.go` 用于测试新的错误处理逻辑。

## 兼容性说明

此修改**向后不兼容**，因为：
1. `CalculateAffectedRowsForStatements` 的返回类型已改变
2. `ConvertToReviewResults` 的参数类型已改变

但由于这些是内部 API，且项目内所有调用点已同步更新，不会影响现有功能。

