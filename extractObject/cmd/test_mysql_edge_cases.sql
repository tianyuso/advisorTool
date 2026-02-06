-- ============================================
-- MySQL 边缘情况测试（修正版）
-- ============================================

-- 31. 使用反引号的表名
SELECT * FROM `user_info` WHERE id = 1;

-- 32. 数据库名和表名都带反引号
SELECT * FROM `my_db`.`order_details` WHERE status = 'active';

-- 33. 复杂的多级子查询
SELECT * FROM (
    SELECT * FROM (
        SELECT id, name FROM users WHERE age > 20
    ) AS level2
    WHERE id > 100
) AS level1;

-- 34. 多表DELETE的另一种写法
DELETE t1, t2
FROM table1 t1
INNER JOIN table2 t2 ON t1.id = t2.ref_id
WHERE t1.created_at < '2024-01-01';

-- 35. 使用索引提示
SELECT * FROM orders USE INDEX (idx_created_at) WHERE created_at > '2025-01-01';

-- 36. STRAIGHT_JOIN
SELECT STRAIGHT_JOIN 
    c.name,
    o.order_date
FROM customers c
JOIN orders o ON c.id = o.customer_id;

-- 37. DERIVED表和子查询组合
SELECT 
    main.category,
    main.total_sales,
    sub.avg_price
FROM (
    SELECT category, SUM(amount) as total_sales
    FROM sales_db.sales
    GROUP BY category
) main
JOIN (
    SELECT category, AVG(price) as avg_price
    FROM mydb.products
    GROUP BY category
) sub ON main.category = sub.category
WHERE main.total_sales > 10000;

-- 38. UNION多个表
SELECT id, name, 'customer' as type FROM customers
UNION ALL
SELECT id, name, 'employee' as type FROM employees  
UNION ALL
SELECT id, company_name, 'vendor' as type FROM vendors
UNION ALL
SELECT id, name, 'partner' as type FROM partners;

-- 39. 嵌套EXISTS
SELECT *
FROM projects p
WHERE EXISTS (
    SELECT 1 FROM project_members pm
    WHERE pm.project_id = p.id
      AND EXISTS (
          SELECT 1 FROM users u
          WHERE u.id = pm.user_id AND u.is_active = 1
      )
);

-- 40. NOT IN与子查询
SELECT *
FROM products p
WHERE p.category_id NOT IN (
    SELECT DISTINCT c.id 
    FROM categories c
    JOIN category_filters cf ON c.id = cf.category_id
    WHERE cf.is_hidden = 1
);

-- 41. HAVING子句中的子查询
SELECT 
    customer_id,
    SUM(amount) as total
FROM orders
GROUP BY customer_id
HAVING SUM(amount) > (
    SELECT AVG(total_amount) 
    FROM (
        SELECT customer_id, SUM(amount) as total_amount
        FROM orders
        GROUP BY customer_id
    ) AS avg_calc
);

-- 42. CASE语句中的子查询
SELECT 
    o.id,
    CASE 
        WHEN o.amount > (SELECT AVG(amount) FROM orders) THEN 'high'
        ELSE 'low'
    END as order_level
FROM orders o;

-- 43. 多层嵌套WITH
WITH 
base_data AS (
    SELECT * FROM raw_data WHERE date >= '2025-01-01'
),
filtered_data AS (
    SELECT * FROM base_data WHERE status = 'active'
),
aggregated_data AS (
    SELECT 
        category,
        COUNT(*) as cnt,
        SUM(amount) as total
    FROM filtered_data
    GROUP BY category
)
SELECT * FROM aggregated_data WHERE total > 1000;

-- 44. 带schema的多表JOIN
SELECT 
    db1.table1.col1,
    db2.table2.col2,
    db3.table3.col3
FROM db1.table1
JOIN db2.table2 ON db1.table1.id = db2.table2.ref_id
LEFT JOIN db3.table3 ON db2.table2.id = db3.table3.ref_id;

-- 45. INSERT INTO SELECT 带多个JOIN
INSERT INTO summary_table (id, name, value, date)
SELECT 
    t1.id,
    t2.name,
    t3.value,
    t4.record_date
FROM base_table t1
INNER JOIN name_table t2 ON t1.id = t2.base_id
LEFT JOIN value_table t3 ON t1.id = t3.base_id
RIGHT JOIN date_table t4 ON t1.id = t4.base_id
WHERE t1.status = 'ready';


