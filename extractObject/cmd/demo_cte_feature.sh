#!/bin/bash

# CTE临时表识别功能演示脚本

echo "============================================"
echo "extractObject CTE临时表识别功能演示"
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
    SELECT 
        user_id,
        COUNT(*) as order_count,
        SUM(amount) as total_amount
    FROM orders
    WHERE created_at >= '2025-01-01'
    GROUP BY user_id
),
active_users AS (
    SELECT user_id, name
    FROM users
    WHERE status = 'active'
)
SELECT 
    au.name,
    us.order_count,
    us.total_amount
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
SELECT * FROM category_tree ORDER BY level, name;
EOF

echo "测试 1: PostgreSQL CTE 识别"
echo "-------------------------------------------"
./extractobject -db postgres -file /tmp/test_pg_cte.sql
echo ""

echo "测试 2: PostgreSQL CTE JSON输出"
echo "-------------------------------------------"
./extractobject -db postgres -file /tmp/test_pg_cte.sql -json
echo ""

echo "测试 3: MySQL 递归CTE 识别"
echo "-------------------------------------------"
./extractobject -db mysql -file /tmp/test_mysql_cte.sql
echo ""

echo "测试 4: MySQL 递归CTE JSON输出"
echo "-------------------------------------------"
./extractobject -db mysql -file /tmp/test_mysql_cte.sql -json
echo ""

# 清理临时文件
rm -f /tmp/test_pg_cte.sql /tmp/test_mysql_cte.sql

echo "============================================"
echo "演示完成！"
echo "============================================"

