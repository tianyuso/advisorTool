-- ============================================
-- MySQL 全面测试 SQL 语句集合
-- ============================================

-- 1. 单表查询 - 不带别名
SELECT * FROM users WHERE id > 100;

-- 2. 单表查询 - 带数据库名
SELECT * FROM mydb.orders WHERE status = 'pending';

-- 3. 单表查询 - 使用 AS 别名
SELECT u.id, u.name FROM users AS u WHERE u.age > 18;

-- 4. 单表查询 - 不使用 AS 的别名
SELECT u.id, u.name FROM customers u WHERE u.status = 'active';

-- 5. 两表 JOIN - 混合别名风格
SELECT 
    o.order_id,
    c.customer_name
FROM orders AS o
INNER JOIN customers c ON o.customer_id = c.id;

-- 6. 多表 JOIN - 包含数据库名
SELECT 
    u.user_id,
    o.order_date,
    p.product_name,
    od.quantity
FROM mydb.users u
LEFT JOIN sales_db.orders AS o ON u.id = o.user_id
INNER JOIN products p ON o.product_id = p.id
RIGHT JOIN order_details od ON o.id = od.order_id;

-- 7. 子查询 - 单表
SELECT * FROM (
    SELECT id, name FROM employees WHERE dept_id = 10
) AS emp_subset;

-- 8. 子查询 - 多表
SELECT *
FROM (
    SELECT u.id, u.name, o.total
    FROM users u
    JOIN orders o ON u.id = o.user_id
) t1
WHERE t1.total > 1000;

-- 9. INSERT 语句 - 单表
INSERT INTO users (name, email, age) 
VALUES ('张三', 'zhangsan@example.com', 25);

-- 10. INSERT 语句 - 带数据库名
INSERT INTO mydb.customers (customer_name, phone)
VALUES ('李四', '13800138000');

-- 11. INSERT SELECT 语句
INSERT INTO archive_orders (order_id, customer_id, order_date)
SELECT id, customer_id, created_at
FROM orders
WHERE created_at < '2024-01-01';

-- 12. INSERT SELECT - 多表 JOIN
INSERT INTO sales_summary (customer_id, total_amount, order_count)
SELECT 
    c.id,
    SUM(o.amount) as total,
    COUNT(o.id) as cnt
FROM customers c
JOIN orders o ON c.id = o.customer_id
GROUP BY c.id;

-- 13. UPDATE 语句 - 单表
UPDATE users 
SET status = 'inactive', updated_at = NOW()
WHERE last_login < '2024-01-01';

-- 14. UPDATE 语句 - 带数据库名
UPDATE mydb.products
SET price = price * 1.1
WHERE category = 'electronics';

-- 15. UPDATE 多表 - 使用 JOIN
UPDATE orders o
JOIN customers c ON o.customer_id = c.id
SET o.status = 'vip_order'
WHERE c.level = 'VIP';

-- 16. UPDATE 多表 - 带数据库名和别名
UPDATE sales_db.orders AS o
JOIN mydb.customers c ON o.customer_id = c.id
SET o.discount = 0.15
WHERE c.member_since < '2020-01-01';

-- 17. DELETE 语句 - 单表
DELETE FROM temp_logs WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);

-- 18. DELETE 语句 - 带数据库名
DELETE FROM mydb.old_records WHERE archived = 1;

-- 19. DELETE 多表 - 使用 JOIN
DELETE o
FROM orders o
JOIN customers c ON o.customer_id = c.id
WHERE c.status = 'deleted';

-- 20. DELETE 多表 - 多个数据库
DELETE o, od
FROM sales_db.orders o
JOIN sales_db.order_details od ON o.id = od.order_id
JOIN mydb.customers c ON o.customer_id = c.id
WHERE c.deleted_at IS NOT NULL;

-- 21. WITH CTE - 单个 CTE
WITH high_value_customers AS (
    SELECT customer_id, SUM(amount) as total
    FROM orders
    GROUP BY customer_id
    HAVING total > 10000
)
SELECT c.name, hvc.total
FROM customers c
JOIN high_value_customers hvc ON c.id = hvc.customer_id;

-- 22. WITH CTE - 多个 CTE
WITH 
monthly_sales AS (
    SELECT 
        DATE_FORMAT(order_date, '%Y-%m') as month,
        SUM(amount) as total
    FROM sales_db.orders
    GROUP BY month
),
top_products AS (
    SELECT 
        product_id,
        COUNT(*) as order_count
    FROM order_details
    GROUP BY product_id
    ORDER BY order_count DESC
    LIMIT 10
)
SELECT 
    ms.month,
    ms.total,
    COUNT(DISTINCT tp.product_id) as top_product_count
FROM monthly_sales ms
CROSS JOIN top_products tp
GROUP BY ms.month, ms.total;

-- 23. WITH CTE - 嵌套引用
WITH 
user_orders AS (
    SELECT u.id as user_id, u.name, o.id as order_id, o.amount
    FROM mydb.users AS u
    JOIN sales_db.orders o ON u.id = o.user_id
),
user_totals AS (
    SELECT user_id, name, SUM(amount) as total_spent
    FROM user_orders
    GROUP BY user_id, name
)
SELECT * FROM user_totals WHERE total_spent > 5000;

-- 24. UNION 查询
SELECT id, name, 'customer' as type FROM customers
UNION ALL
SELECT id, name, 'supplier' as type FROM suppliers;

-- 25. UNION 查询 - 带数据库名和别名
SELECT u.id, u.email FROM mydb.active_users u
UNION
SELECT i.id, i.email FROM archive_db.inactive_users AS i;

-- 26. 复杂嵌套查询
SELECT 
    main.customer_id,
    main.order_count,
    detail.avg_amount
FROM (
    SELECT customer_id, COUNT(*) as order_count
    FROM orders
    GROUP BY customer_id
) main
LEFT JOIN (
    SELECT 
        o.customer_id,
        AVG(od.unit_price * od.quantity) as avg_amount
    FROM sales_db.orders AS o
    JOIN order_details od ON o.id = od.order_id
    GROUP BY o.customer_id
) detail ON main.customer_id = detail.customer_id;

-- 27. EXISTS 子查询
SELECT c.id, c.name
FROM customers c
WHERE EXISTS (
    SELECT 1 
    FROM orders o 
    WHERE o.customer_id = c.id 
      AND o.amount > 1000
);

-- 28. IN 子查询 - 多表
SELECT p.product_name, p.price
FROM products p
WHERE p.id IN (
    SELECT DISTINCT od.product_id
    FROM order_details od
    JOIN orders o ON od.order_id = o.id
    WHERE o.order_date >= '2025-01-01'
);

-- 29. REPLACE 语句
REPLACE INTO user_settings (user_id, setting_key, setting_value)
VALUES (1, 'theme', 'dark');

-- 30. REPLACE INTO SELECT
REPLACE INTO mydb.product_cache (product_id, product_name, price)
SELECT id, name, price FROM products WHERE updated_at > NOW() - INTERVAL 1 HOUR;





