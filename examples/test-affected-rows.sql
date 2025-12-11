-- 示例 SQL 文件：用于测试影响行数计算功能
-- 此文件包含各种类型的 UPDATE 和 DELETE 语句

-- ============================================
-- MySQL 示例
-- ============================================

-- 1. 简单的单表 UPDATE（带 WHERE）
UPDATE users 
SET status = 'active', updated_at = NOW() 
WHERE last_login < DATE_SUB(NOW(), INTERVAL 30 DAY);

-- 2. 简单的单表 DELETE（带 WHERE）
DELETE FROM logs 
WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY);

-- 3. 连表 UPDATE
UPDATE orders o
INNER JOIN customers c ON o.customer_id = c.id
SET o.customer_name = c.name
WHERE c.updated_at > '2024-01-01';

-- 4. 连表 DELETE
DELETE o
FROM orders o
INNER JOIN customers c ON o.customer_id = c.id
WHERE c.status = 'deleted';

-- 5. 多表 JOIN 的 UPDATE
UPDATE products p
INNER JOIN categories c ON p.category_id = c.id
LEFT JOIN discounts d ON p.id = d.product_id
SET p.discount_rate = COALESCE(d.rate, 0)
WHERE c.name = 'Electronics';

-- ============================================
-- PostgreSQL 示例（注释掉，避免语法错误）
-- ============================================

-- 单表 UPDATE 带 WHERE
-- UPDATE users 
-- SET status = 'active', updated_at = CURRENT_TIMESTAMP 
-- WHERE last_login < CURRENT_TIMESTAMP - INTERVAL '30 days';

-- 连表 UPDATE（PostgreSQL 语法）
-- UPDATE orders 
-- SET customer_name = customers.name 
-- FROM customers 
-- WHERE orders.customer_id = customers.id 
--   AND customers.updated_at > '2024-01-01';

-- 单表 DELETE 带 WHERE
-- DELETE FROM logs 
-- WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '90 days';

-- 连表 DELETE（PostgreSQL 语法）
-- DELETE FROM orders 
-- USING customers 
-- WHERE orders.customer_id = customers.id 
--   AND customers.status = 'deleted';

-- ============================================
-- SQL Server 示例（注释掉，避免语法错误）
-- ============================================

-- 单表 UPDATE 带 WHERE
-- UPDATE users 
-- SET status = 'active', updated_at = GETDATE() 
-- WHERE last_login < DATEADD(day, -30, GETDATE());

-- 连表 UPDATE（SQL Server 语法）
-- UPDATE o
-- SET o.customer_name = c.name
-- FROM orders o
-- INNER JOIN customers c ON o.customer_id = c.id
-- WHERE c.updated_at > '2024-01-01';

-- 单表 DELETE 带 WHERE
-- DELETE FROM logs 
-- WHERE created_at < DATEADD(day, -90, GETDATE());

-- 连表 DELETE（SQL Server 语法）
-- DELETE o
-- FROM orders o
-- INNER JOIN customers c ON o.customer_id = c.id
-- WHERE c.status = 'deleted';

-- ============================================
-- 测试场景
-- ============================================

-- 不带 WHERE 的 UPDATE（应该触发警告）
UPDATE test_table SET status = 1;

-- 不带 WHERE 的 DELETE（应该触发警告）
DELETE FROM temp_data;

-- 复杂的 WHERE 条件
UPDATE products 
SET price = price * 1.1 
WHERE category IN ('Electronics', 'Computers') 
  AND stock > 0 
  AND price < 1000;

-- 使用子查询的 UPDATE（简单情况）
UPDATE orders 
SET total_amount = (
    SELECT SUM(quantity * price) 
    FROM order_items 
    WHERE order_items.order_id = orders.id
)
WHERE status = 'pending';

