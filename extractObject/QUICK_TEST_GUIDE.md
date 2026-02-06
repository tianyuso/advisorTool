# extractObject 快速测试验证脚本

本文档提供快速验证extractObject工具功能的测试命令。

## 测试文件位置

所有综合测试SQL文件已生成在 `/tmp/` 目录下：
- `/tmp/comprehensive_test_mysql.sql` (219行)
- `/tmp/comprehensive_test_sqlserver.sql` (226行)
- `/tmp/comprehensive_test_postgres.sql` (228行)
- `/tmp/comprehensive_test_oracle.sql` (228行)

## 快速测试命令

### 1. MySQL 全面测试
```bash
cd /data/dev_go/advisorTool/extractObject/cmd
./extractobject -db MYSQL -file /tmp/comprehensive_test_mysql.sql
```
**期望结果**: 找到 52 个表 (42个物理表 + 10个CTE)

### 2. SQL Server 全面测试
```bash
cd /data/dev_go/advisorTool/extractObject/cmd
./extractobject -db SQLSERVER -file /tmp/comprehensive_test_sqlserver.sql
```
**期望结果**: 找到 60 个表 (45个物理表 + 15个CTE)

### 3. PostgreSQL 全面测试
```bash
cd /data/dev_go/advisorTool/extractObject/cmd
./extractobject -db POSTGRESQL -file /tmp/comprehensive_test_postgres.sql
```
**期望结果**: 找到 49 个表 (38个物理表 + 11个CTE)

### 4. Oracle 全面测试
```bash
cd /data/dev_go/advisorTool/extractObject/cmd
./extractobject -db ORACLE -file /tmp/comprehensive_test_oracle.sql
```
**期望结果**: 找到 61 个表 (46个物理表 + 15个CTE)

## 边界测试：同表多别名引用

```bash
cd /data/dev_go/advisorTool/extractObject/cmd
cat > /tmp/test_multi_alias.sql << 'EOF'
SELECT u1.name, u2.name, u3.name
FROM users u1
LEFT JOIN users u2 ON u1.referred_by = u2.id
LEFT JOIN users u3 ON u1.manager_id = u3.id;
EOF
./extractobject -db MYSQL -file /tmp/test_multi_alias.sql
```
**期望结果**: 找到 3 个表引用 (users u1, users u2, users u3)

## JSON输出格式测试

```bash
cd /data/dev_go/advisorTool/extractObject/cmd
./extractobject -db MYSQL -format json -file /tmp/comprehensive_test_mysql.sql | jq '.[0:3]'
```

## 特定场景验证

### CTE识别验证
```bash
cat > /tmp/test_cte.sql << 'EOF'
WITH active_users AS (
    SELECT id, name FROM users WHERE status = 'active'
),
premium_users AS (
    SELECT id FROM users WHERE tier = 'premium'
)
SELECT a.name, p.id
FROM active_users a
INNER JOIN premium_users p ON a.id = p.id;
EOF

./extractobject -db POSTGRESQL -file /tmp/test_cte.sql
```
**期望**: 识别出 users(物理表), active_users(CTE), premium_users(CTE)

### 递归CTE验证
```bash
cat > /tmp/test_recursive.sql << 'EOF'
WITH RECURSIVE emp_hierarchy AS (
    SELECT id, name, manager_id, 1 as level
    FROM employees
    WHERE manager_id IS NULL
    UNION ALL
    SELECT e.id, e.name, e.manager_id, eh.level + 1
    FROM employees e
    INNER JOIN emp_hierarchy eh ON e.manager_id = eh.id
)
SELECT * FROM emp_hierarchy;
EOF

./extractobject -db MYSQL -file /tmp/test_recursive.sql
```
**期望**: 识别出 employees(物理表), emp_hierarchy(CTE)

### 跨schema/database引用验证
```bash
cat > /tmp/test_cross_schema.sql << 'EOF'
SELECT u.name, o.total, p.product_name
FROM mydb.users u
INNER JOIN sales_db.orders o ON u.id = o.user_id
LEFT JOIN catalog.products p ON o.product_id = p.id;
EOF

./extractobject -db MYSQL -file /tmp/test_cross_schema.sql
```
**期望**: 正确识别dbname和别名

## 性能测试

```bash
cd /data/dev_go/advisorTool/extractObject/cmd
time ./extractobject -db MYSQL -file /tmp/comprehensive_test_mysql.sql > /dev/null
```
**期望**: 执行时间 < 1秒

## 一键运行所有测试

```bash
#!/bin/bash
cd /data/dev_go/advisorTool/extractObject/cmd

echo "========== MySQL 测试 =========="
./extractobject -db MYSQL -file /tmp/comprehensive_test_mysql.sql | grep "找到"

echo "========== SQL Server 测试 =========="
./extractobject -db SQLSERVER -file /tmp/comprehensive_test_sqlserver.sql | grep "找到"

echo "========== PostgreSQL 测试 =========="
./extractobject -db POSTGRESQL -file /tmp/comprehensive_test_postgres.sql | grep "找到"

echo "========== Oracle 测试 =========="
./extractobject -db ORACLE -file /tmp/comprehensive_test_oracle.sql | grep "找到"

echo ""
echo "✅ 所有测试完成"
```

## 验证检查清单

### 基本功能
- [ ] 单表查询识别
- [ ] 多表JOIN识别
- [ ] INSERT语句表识别
- [ ] UPDATE语句表识别
- [ ] DELETE语句表识别

### 表名格式
- [ ] 简单表名 (table)
- [ ] Schema表名 (schema.table)
- [ ] 完整表名 (database.schema.table) - SQL Server

### 别名
- [ ] AS别名格式 (table AS alias)
- [ ] 无AS别名格式 (table alias)
- [ ] 同表多别名引用

### CTE
- [ ] 简单CTE识别
- [ ] 多个CTE识别
- [ ] 递归CTE识别
- [ ] CTE与物理表区分
- [ ] CTE别名识别

### 输出格式
- [ ] TEXT格式输出
- [ ] JSON格式输出

## 故障排查

### 问题：编译失败
```bash
cd /data/dev_go/advisorTool/extractObject/cmd
go build -o extractobject main.go
```
检查编译错误信息

### 问题：找不到SQL文件
确认测试文件存在：
```bash
ls -lh /tmp/comprehensive_test_*.sql
```

### 问题：识别结果不符合预期
使用JSON格式查看详细信息：
```bash
./extractobject -db MYSQL -format json -file /tmp/test.sql | jq .
```

## 相关文档
- 详细测试报告: `COMPREHENSIVE_TEST_REPORT.md`
- PostgreSQL别名修复: `POSTGRESQL_ALIAS_FIX.md`
- 别名修复报告: `ALIAS_FIX_REPORT.md`

---
**最后更新**: 2026-02-04
**测试版本**: v1.0 (生产就绪)

