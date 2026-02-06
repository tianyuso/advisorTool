# extractObject 命令行工具

SQL表名提取命令行工具，可以从SQL语句中提取表名、数据库名、模式名和别名信息。

## 安装

```bash
cd extractObject/cmd
go build -o extractobject
```

## 使用方法

### 基本用法

```bash
# 从命令行提取
./extractobject -db MYSQL -sql "SELECT * FROM users"

# 从文件提取
./extractobject -db POSTGRESQL -file query.sql

# JSON格式输出
./extractobject -db MYSQL -sql "SELECT * FROM users" -json
```

### 参数说明

- `-db`: 数据库类型（必需）
  - `MYSQL` - MySQL数据库
  - `POSTGRESQL` - PostgreSQL数据库
  - `ORACLE` - Oracle数据库
  - `SQLSERVER` - SQL Server数据库
  - `TIDB` - TiDB数据库
  - `MARIADB` - MariaDB数据库
  - `OCEANBASE` - OceanBase数据库
  - `SNOWFLAKE` - Snowflake数据库

- `-sql`: SQL语句（与-file二选一）

- `-file`: SQL文件路径（与-sql二选一）

- `-json`: 以JSON格式输出结果

- `-version`: 显示版本信息

## 使用示例

### MySQL示例

```bash
./extractobject -db MYSQL -sql "SELECT u.id, o.order_id FROM mydb.users u JOIN orders o ON u.id = o.user_id"
```

输出:
```
找到 2 个表:

数据库名              模式名                表名                            别名
--------------------------------------------------------------------------------
mydb                -                    users                          u
-                   -                    orders                         o
```

### PostgreSQL示例

```bash
./extractobject -db POSTGRESQL -file query.sql
```

query.sql内容:
```sql
SELECT p.product_name, c.category_name
FROM public.products p
INNER JOIN public.categories c ON p.category_id = c.id
WHERE c.status = 'active'
```

输出:
```
找到 2 个表:

数据库名              模式名                表名                            别名
--------------------------------------------------------------------------------
-                   public               products                       p
-                   public               categories                     c
```

### JSON格式输出

```bash
./extractobject -db MYSQL -sql "SELECT * FROM mydb.users u" -json
```

输出:
```json
[
  {
    "DBName": "mydb",
    "Schema": "",
    "TBName": "users",
    "Alias": "u"
  }
]
```

## 支持的SQL语句

- SELECT查询
- INSERT语句
- UPDATE语句
- DELETE语句
- JOIN操作（INNER JOIN, LEFT JOIN, RIGHT JOIN等）
- 子查询
- CTE（公用表表达式）

## 错误处理

如果SQL语句有语法错误或无法解析，工具会输出错误信息：

```bash
./extractobject -db MYSQL -sql "SELECT * FORM users"
# 输出: 提取表名失败: 解析SQL失败: ...
```


