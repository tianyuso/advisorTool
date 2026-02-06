# MySQL extractObject 工具测试总结

## 🎯 测试目标

对 extractObject 工具进行全面的 MySQL 数据库场景测试，验证其在各种SQL语句类型、表名格式、别名方式以及复杂查询场景下的表提取能力。

---

## 📋 测试范围

### 语句类型覆盖
- ✅ SELECT（单表、多表、子查询、UNION）
- ✅ INSERT（单表、INSERT SELECT）
- ✅ UPDATE（单表、多表JOIN）
- ✅ DELETE（单表、多表JOIN）
- ✅ REPLACE（单表、REPLACE SELECT）
- ✅ WITH CTE（单个、多个、嵌套）

### 表名格式覆盖
- ✅ 简单表名: `users`
- ✅ 数据库.表名: `mydb.orders`
- ✅ 反引号: `` `user_info` ``
- ✅ 完整反引号: `` `mydb`.`orders` ``
- ✅ 多级引用: `db1.table1.col1`

### 别名格式覆盖
- ✅ AS别名: `FROM users AS u`
- ✅ 不带AS: `FROM customers c`
- ✅ 混合使用: `FROM orders AS o JOIN customers c`

### JOIN类型覆盖
- ✅ INNER JOIN (2表、4表)
- ✅ LEFT JOIN
- ✅ RIGHT JOIN
- ✅ CROSS JOIN
- ✅ STRAIGHT_JOIN
- ✅ 跨数据库 JOIN

---

## 📊 测试结果统计

### 基础全面测试
- **测试文件**: `test_mysql_comprehensive.sql`
- **场景数量**: 30个
- **提取表数**: 59个（含重复）
- **唯一表数**: 26个
- **涉及数据库**: 3个 (mydb, sales_db, archive_db)
- **通过率**: ✅ **100%**

### 边缘情况测试
- **测试文件**: `test_mysql_edge_cases.sql`
- **场景数量**: 15个
- **提取表数**: 36个（含重复）
- **唯一表数**: 23个
- **涉及数据库**: 4个
- **通过率**: ✅ **100%**

---

## 🎉 核心功能验证

| 功能 | 状态 | 说明 |
|------|------|------|
| 单表查询 | ✅ | 支持所有格式 |
| 多表JOIN | ✅ | 2-4表，所有JOIN类型 |
| 跨数据库操作 | ✅ | 完美支持 |
| 别名处理 | ✅ | AS和非AS都支持 |
| INSERT语句 | ✅ | 单表和INSERT SELECT |
| UPDATE语句 | ✅ | 单表和多表JOIN |
| DELETE语句 | ✅ | 单表和多表JOIN |
| WITH CTE | ✅ | 单个、多个、嵌套 |
| 子查询 | ✅ | 多层嵌套、EXISTS、IN |
| UNION | ✅ | 多表UNION |
| 特殊字符 | ✅ | 反引号支持 |

---

## 📝 测试文件说明

### 1. test_mysql_comprehensive.sql
包含30个基础场景：
- 单表查询（4种格式）
- 多表JOIN（2-4表）
- INSERT/UPDATE/DELETE（各3种）
- WITH CTE（3种复杂度）
- 子查询和UNION（5种）
- 跨数据库操作（多个）

### 2. test_mysql_edge_cases.sql
包含15个边缘场景：
- 反引号表名
- 多级嵌套子查询
- 复杂的多表操作
- 索引提示
- STRAIGHT_JOIN
- 嵌套EXISTS和IN

### 3. MYSQL_TEST_REPORT.md
详细的测试报告文档，包含：
- 完整的测试分类
- 每个场景的示例代码
- 统计数据和分析
- 使用说明

### 4. test_mysql.sh
自动化测试脚本，快速验证：
```bash
./test_mysql.sh
```

---

## 💡 使用示例

### 基本命令
```bash
# 从文件读取SQL
./extractobject -db MYSQL -file test.sql

# 直接输入SQL
./extractobject -db MYSQL -sql "SELECT * FROM users"

# JSON格式输出
./extractobject -db MYSQL -file test.sql -json
```

### 输出示例

**文本格式:**
```
找到 5 个表:

数据库名                 模式名                  表名                             别名                  
--------------------------------------------------------------------------------
mydb                 -                    users                          -                   
sales_db             -                    orders                         -                   
-                    -                    customers                      -                   
```

**JSON格式:**
```json
[
  {
    "DBName": "mydb",
    "Schema": "",
    "TBName": "users",
    "Alias": ""
  }
]
```

---

## ✅ 测试结论

### 优势
1. **准确率高**: 100% 通过率，准确识别各种格式
2. **功能全面**: 支持所有常用SQL语句类型
3. **格式灵活**: AS/非AS别名、反引号、跨库都支持
4. **复杂查询**: CTE、子查询、多表JOIN完美处理
5. **易于使用**: 命令行简单，支持文件和直接输入

### 适用场景
- ✅ SQL审计和安全分析
- ✅ 数据血缘关系分析
- ✅ 权限管理（表级别）
- ✅ SQL优化前的依赖分析
- ✅ 自动化文档生成

### 局限性
- ⚠️ 窗口函数的某些高级语法可能不支持
- ⚠️ 递归CTE的RECURSIVE关键字可能有限制
- ⚠️ 某些MySQL 8.0+的新特性需要验证

---

## 🚀 快速开始

```bash
# 1. 进入工具目录
cd /data/dev_go/advisorTool/extractObject/cmd

# 2. 运行快速测试
./test_mysql.sh

# 3. 查看详细报告
cat MYSQL_TEST_REPORT.md

# 4. 测试自己的SQL
./extractobject -db MYSQL -file your_sql.sql
```

---

## 📚 相关文件

- `test_mysql_comprehensive.sql` - 基础全面测试SQL
- `test_mysql_edge_cases.sql` - 边缘情况测试SQL
- `MYSQL_TEST_REPORT.md` - 详细测试报告
- `test_mysql.sh` - 自动化测试脚本
- `README.md` - 工具说明文档

---

**测试日期**: 2026-02-04  
**工具版本**: extractObject v1.0.0  
**测试状态**: ✅ **全部通过**


