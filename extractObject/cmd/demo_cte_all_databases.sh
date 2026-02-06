#!/bin/bash

# CTE临时表识别功能完整演示（包括Oracle和SQL Server）

echo "============================================"
echo "extractObject CTE临时表识别功能完整演示"
echo "============================================"
echo ""

# 进入工具目录
cd "$(dirname "$0")" || exit 1

# 确保工具已编译
if [ ! -f "./extractobject" ]; then
    echo "正在编译 extractObject 工具..."
    go build -o extractobject main.go
    echo ""
fi

# 创建PostgreSQL测试SQL
cat > /tmp/test_pg_cte.sql << 'EOF'
WITH 
user_stats AS (
    SELECT user_id, COUNT(*) as order_count
    FROM orders
    GROUP BY user_id
),
active_users AS (
    SELECT user_id, name FROM users WHERE status = 'active'
)
SELECT au.name, us.order_count
FROM active_users au
JOIN user_stats us ON au.user_id = us.user_id;
EOF

# 创建MySQL测试SQL
cat > /tmp/test_mysql_cte.sql << 'EOF'
WITH RECURSIVE category_tree AS (
    SELECT id, name, parent_id, 1 as level
    FROM categories
    WHERE parent_id IS NULL
    UNION ALL
    SELECT c.id, c.name, c.parent_id, ct.level + 1
    FROM categories c
    INNER JOIN category_tree ct ON c.parent_id = ct.id
)
SELECT * FROM category_tree;
EOF

# 创建Oracle测试SQL
cat > /tmp/test_oracle_cte.sql << 'EOF'
WITH employee_hierarchy AS (
    SELECT employee_id, first_name, manager_id, 1 as level
    FROM hr.employees
    WHERE manager_id IS NULL
    UNION ALL
    SELECT e.employee_id, e.first_name, e.manager_id, eh.level + 1
    FROM hr.employees e
    INNER JOIN employee_hierarchy eh ON e.manager_id = eh.employee_id
)
SELECT * FROM employee_hierarchy;
EOF

# 创建SQL Server测试SQL
cat > /tmp/test_sqlserver_cte.sql << 'EOF'
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
EOF

echo "=========================================="
echo "测试 1: PostgreSQL CTE 识别"
echo "=========================================="
./extractobject -db postgres -file /tmp/test_pg_cte.sql
echo ""

echo "=========================================="
echo "测试 2: MySQL 递归CTE 识别"
echo "=========================================="
./extractobject -db mysql -file /tmp/test_mysql_cte.sql
echo ""

echo "=========================================="
echo "测试 3: Oracle CTE 识别"
echo "=========================================="
./extractobject -db oracle -file /tmp/test_oracle_cte.sql
echo ""

echo "=========================================="
echo "测试 4: SQL Server CTE 识别"
echo "=========================================="
./extractobject -db sqlserver -file /tmp/test_sqlserver_cte.sql
echo ""

echo "=========================================="
echo "JSON 格式示例 (PostgreSQL)"
echo "=========================================="
./extractobject -db postgres -file /tmp/test_pg_cte.sql -json
echo ""

# 清理临时文件
rm -f /tmp/test_pg_cte.sql /tmp/test_mysql_cte.sql /tmp/test_oracle_cte.sql /tmp/test_sqlserver_cte.sql

echo "============================================"
echo "演示完成！支持的数据库："
echo "  ✓ PostgreSQL"
echo "  ✓ MySQL"
echo "  ✓ Oracle"
echo "  ✓ SQL Server"
echo "============================================"

