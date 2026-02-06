# CTE临时表识别功能更新

## 更新日期
2026-02-04

## 功能说明
在 `extractObject` 工具中添加了CTE（Common Table Expression，公用表表达式）临时表的识别功能。

## 更新内容

### 1. TableInfo结构体更新
在 `extractObject/types.go` 中，为 `TableInfo` 结构体添加了 `IsCTE` 字段：

```go
type TableInfo struct {
    DBName string // 数据库名
    Schema string // 模式名
    TBName string // 表名
    Alias  string // 别名
    IsCTE  bool   // 是否是CTE临时表 (新增)
}
```

### 2. 支持的数据库类型
目前已为以下数据库类型实现了CTE识别功能：
- ✅ **PostgreSQL** - 完全支持，包括递归CTE
- ✅ **MySQL** - 完全支持，包括递归CTE
- ✅ **Oracle** - 完全支持，包括递归CTE（使用Subquery Factoring子句）
- ✅ **SQL Server** - 完全支持，包括递归CTE

### 3. 输出格式更新

#### 文本格式输出
添加了"类型"列，显示表是"物理表"还是"CTE临时表"：

```
找到 5 个表:

数据库名                 模式名                  表名                             别名                   类型        
--------------------------------------------------------------------------------------------
-                    iss_dwd              dwd_hr_taskinfo                -                    物理表       
-                    -                    first_date_per_user            -                    CTE临时表    
-                    -                    date_series                    -                    CTE临时表    
-                    -                    daily_new_users                -                    CTE临时表    
-                    -                    daily_active_users             -                    CTE临时表    
```

#### JSON格式输出
在JSON输出中包含 `IsCTE` 字段：

```json
[
  {
    "DBName": "",
    "Schema": "iss_dwd",
    "TBName": "dwd_hr_taskinfo",
    "Alias": "",
    "IsCTE": false
  },
  {
    "DBName": "",
    "Schema": "",
    "TBName": "first_date_per_user",
    "Alias": "",
    "IsCTE": true
  }
]
```

## 使用示例

### PostgreSQL示例

```sql
WITH first_date_per_user AS (
    SELECT creatorno, min(creationtime::date) as first_date
    FROM iss_dwd.dwd_hr_taskinfo
    GROUP BY creatorno
)
SELECT * FROM first_date_per_user;
```

运行命令：
```bash
./extractobject -db POSTGRESQL -file test.sql
```

### MySQL示例

```sql
WITH RECURSIVE employee_hierarchy AS (
    SELECT id, name, manager_id, 1 as level
    FROM employees
    WHERE manager_id IS NULL
    UNION ALL
    SELECT e.id, e.name, e.manager_id, eh.level + 1
    FROM employees e
    INNER JOIN employee_hierarchy eh ON e.manager_id = eh.id
)
SELECT * FROM employee_hierarchy;
```

运行命令：
```bash
./extractobject -db MYSQL -file test.sql
```

### Oracle示例

```sql
WITH employee_hierarchy AS (
    SELECT employee_id, first_name, last_name, manager_id, 1 as level
    FROM hr.employees
    WHERE manager_id IS NULL
    UNION ALL
    SELECT e.employee_id, e.first_name, e.last_name, e.manager_id, eh.level + 1
    FROM hr.employees e
    INNER JOIN employee_hierarchy eh ON e.manager_id = eh.employee_id
)
SELECT * FROM employee_hierarchy;
```

运行命令：
```bash
./extractobject -db ORACLE -file test.sql
```

### SQL Server示例

```sql
WITH CategoryHierarchy AS (
    SELECT CategoryID, CategoryName, ParentCategoryID, 1 as Level
    FROM dbo.Categories
    WHERE ParentCategoryID IS NULL
    UNION ALL
    SELECT c.CategoryID, c.CategoryName, c.ParentCategoryID, ch.Level + 1
    FROM dbo.Categories c
    INNER JOIN CategoryHierarchy ch ON c.ParentCategoryID = ch.CategoryID
)
SELECT * FROM CategoryHierarchy;
```

运行命令：
```bash
./extractobject -db SQLSERVER -file test.sql
```

## 技术实现

### PostgreSQL实现
- 通过监听 `EnterCommon_table_expr` 事件捕获CTE定义
- 在 `cteNames` map中记录所有CTE名称
- 在处理表引用时检查表名是否在CTE集合中

### MySQL实现
- 通过监听 `EnterCommonTableExpression` 事件捕获CTE定义
- 使用相同的机制标记CTE临时表

### Oracle实现
- 通过监听 `EnterFactoring_element` 事件捕获CTE定义（Oracle使用Subquery Factoring Clause）
- 表名规范化为大写（Oracle默认不区分大小写）
- 在 `cteNames` map中记录所有CTE名称

### SQL Server实现
- 通过监听 `EnterCommon_table_expression` 事件捕获CTE定义
- 使用 `NormalizeTSQLIdentifierText` 规范化CTE名称
- SQL Server只允许顶层CTE（不支持嵌套）

## 后续计划
- 为其他数据库类型（TiDB, Snowflake等）添加CTE识别功能
- 考虑添加CTE作用域信息（嵌套CTE的父子关系）
- 添加递归CTE的特殊标记

## 相关文件
- `extractObject/types.go` - 数据结构定义
- `extractObject/postgresql_extractor.go` - PostgreSQL实现
- `extractObject/mysql_extractor.go` - MySQL实现
- `extractObject/oracle_extractor.go` - Oracle实现
- `extractObject/sqlserver_extractor.go` - SQL Server实现
- `extractObject/cmd/main.go` - 命令行工具和输出格式

