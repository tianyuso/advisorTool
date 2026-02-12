# 参数名称更新说明

## 更新内容

将命令行参数 `-db` 改为 `-dbtype`，使参数名称更加明确，减少误解。

## 变更对比

### 之前
```bash
./extractobject -db mysql -sql "SELECT * FROM users"
```

### 现在
```bash
./extractobject -dbtype mysql -sql "SELECT * FROM users"
```

## 更新原因

1. **更明确**：`-dbtype` 比 `-db` 更清楚地表达是"数据库类型"
2. **减少误解**：避免与数据库名称（database name）混淆
3. **更专业**：符合命令行参数的命名惯例

## 使用示例

### 基本用法

```bash
# MySQL
./extractobject -dbtype mysql -sql "SELECT * FROM users"

# PostgreSQL
./extractobject -dbtype postgres -file query.sql

# Oracle (JSON输出)
./extractobject -dbtype oracle -sql "SELECT * FROM hr.employees" -json

# SQL Server
./extractobject -dbtype sqlserver -file query.sql
```

### 所有支持的数据库类型

```bash
./extractobject -dbtype mysql -sql "..."
./extractobject -dbtype postgres -sql "..."
./extractobject -dbtype oracle -sql "..."
./extractobject -dbtype sqlserver -sql "..."
./extractobject -dbtype tidb -sql "..."
./extractobject -dbtype mariadb -sql "..."
./extractobject -dbtype oceanbase -sql "..."
```

### 支持大小写

```bash
# 小写（推荐）
./extractobject -dbtype mysql -sql "SELECT * FROM users"

# 大写（也支持）
./extractobject -dbtype MYSQL -sql "SELECT * FROM users"

# 混合（也支持）
./extractobject -dbtype PostgreSQL -sql "SELECT * FROM users"
```

### 别名支持

```bash
# PostgreSQL 的别名
./extractobject -dbtype postgres -sql "..."
./extractobject -dbtype postgresql -sql "..."

# SQL Server 的别名
./extractobject -dbtype sqlserver -sql "..."
./extractobject -dbtype mssql -sql "..."
```

## 查看帮助

```bash
./extractobject -h
```

输出：
```
  -dbtype string
        数据库类型 (mysql, postgres, oracle, sqlserver, tidb, mariadb, oceanbase, snowflake) (default "mysql")
  -file string
        SQL文件路径
  -json
        以JSON格式输出
  -sql string
        SQL语句
  -version
        显示版本信息
```

## 完整示例

### 从SQL语句提取

```bash
./extractobject -dbtype mysql -sql "
SELECT 
    u.id, 
    u.name, 
    o.order_id
FROM mydb.users AS u
JOIN orders o ON u.id = o.user_id
WHERE u.status = 'active'
"
```

### 从文件提取

```bash
# 创建SQL文件
cat > query.sql << 'EOF'
SELECT p.product_name, c.category_name
FROM public.products p
INNER JOIN public.categories c ON p.category_id = c.id
WHERE c.status = 'active'
EOF

# 提取表名
./extractobject -dbtype postgres -file query.sql
```

### JSON输出

```bash
./extractobject -dbtype mysql -sql "SELECT * FROM mydb.users" -json
```

输出：
```json
[
  {
    "DBName": "mydb",
    "Schema": "",
    "TBName": "users",
    "Alias": "",
    "IsCTE": false
  }
]
```

## 测试验证

所有测试通过：

### 功能测试
✅ mysql - 通过
✅ postgres - 通过  
✅ oracle - 通过
✅ sqlserver - 通过
✅ MYSQL (大写) - 通过
✅ POSTGRESQL (大写) - 通过
✅ postgresql (别名) - 通过
✅ mssql (别名) - 通过
✅ tidb - 通过
✅ mariadb - 通过
✅ oceanbase - 通过
✅ 无效参数错误处理 - 通过

### 运行测试

```bash
# 编译工具
cd /data/dev_go/advisorTool/extractObject/cmd
go build -o extractobject main.go

# 运行完整测试
cd ..
./test_new_params.sh
```

## 更新文件列表

### 代码文件
- ✅ `cmd/main.go` - 参数定义

### Shell 脚本
- ✅ `cmd/demo_cte_feature.sh`
- ✅ `cmd/demo_cte_all_databases.sh`
- ✅ `cmd/test_mysql.sh`
- ✅ `final_demo.sh`
- ✅ `test.sh`
- ✅ `test_new_params.sh`

### 文档文件
- ✅ `README.md`
- ✅ `DATABASE_TYPE_UPDATE.md`
- ✅ `CHANGELOG.md`

## 常见问题

### Q: 为什么要改参数名？
A: `-dbtype` 比 `-db` 更明确，避免与数据库名称（database name）混淆。

### Q: 支持哪些数据库？
A: MySQL, PostgreSQL, Oracle, SQL Server, TiDB, MariaDB, OceanBase。

### Q: 参数大小写敏感吗？
A: 不敏感，支持小写、大写和混合大小写。

### Q: 有别名吗？
A: 有，如 `postgres`/`postgresql`，`sqlserver`/`mssql` 等。

## 更新日期

2026-02-06

---

**现在开始使用新的 `-dbtype` 参数吧！** 🎉

