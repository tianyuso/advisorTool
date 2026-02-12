-- ============================================
-- MySQL 边缘情况和高级特性测试
-- ============================================

-- 31. 使用反引号的表名
SELECT * FROM `user-info` WHERE id = 1;

-- 32. 数据库名和表名都带反引号
SELECT * FROM `my-db`.`order-details` WHERE status = 'active';

-- 33. 复杂的多级子查询
SELECT * FROM (
    SELECT * FROM (
        SELECT id, name FROM users WHERE age > 20
    ) AS level2
    WHERE id > 100
) AS level1;

-- 34. LATERAL JOIN (MySQL 8.0+)
SELECT
   d.department_name,
   t.first_name,
   t.last_name,
   t.salary
FROM
   departments d
LEFT JOIN LATERAL (
   SELECT
       e.first_name,
       e.last_name,
       e.salary
   FROM
       employees e
   WHERE
       e.department_id = d.department_id
   ORDER BY
       e.salary DESC
   LIMIT 5
) t ON TRUE
ORDER BY
   d.department_name,
   t.salary DESC;
-- 35. 窗口函数
SELECT
  sales.*,
  AVG( revenue ) OVER ( PARTITION BY category ) AS avg_revenue 
FROM
	sales;

-- 先统计每年每个工程师的工单量，再对每年的工程师按工单量降序排名
SELECT
    YEAR(create_time) AS work_year,
    engineer,
    engineer_display,
    COUNT(*) AS total_workflow,
    -- 生成连续行号
    ROW_NUMBER() OVER (PARTITION BY YEAR(create_time) ORDER BY COUNT(*) DESC) AS row_num,
    -- 不连续排名（同分同名次，后续跳过）
    RANK() OVER (PARTITION BY YEAR(create_time) ORDER BY COUNT(*) DESC) AS rank_num,
    -- 连续排名（同分同名次，后续连续）
    DENSE_RANK() OVER (PARTITION BY YEAR(create_time) ORDER BY COUNT(*) DESC) AS dense_rank_num
FROM sql_workflow
GROUP BY work_year, engineer, engineer_display
ORDER BY work_year ASC, total_workflow DESC;

-- 36. MERGE语句风格的INSERT ON DUPLICATE
INSERT INTO user_stats (user_id, login_count, last_login)
SELECT u.id, COUNT(l.id), MAX(l.login_time)
FROM users u
LEFT JOIN login_logs l ON u.id = l.user_id
GROUP BY u.id
ON DUPLICATE KEY UPDATE 
    login_count = VALUES(login_count),
    last_login = VALUES(last_login);

-- 37. 多个WITH递归CTE
WITH RECURSIVE 
employee_hierarchy AS (
    SELECT id, name, manager_id, 1 as level
    FROM employees
    WHERE manager_id IS NULL
    
    UNION ALL
    
    SELECT e.id, e.name, e.manager_id, eh.level + 1
    FROM employees e
    JOIN employee_hierarchy eh ON e.manager_id = eh.id
),
dept_summary AS (
    SELECT dept_id, COUNT(*) as emp_count
    FROM employees
    GROUP BY dept_id
)
SELECT 
    eh.name,
    eh.level,
    ds.emp_count
FROM employee_hierarchy eh
LEFT JOIN dept_summary ds ON eh.id = ds.dept_id;

-- 38. CASE语句中的子查询
SELECT 
    o.id,
    CASE 
        WHEN o.amount > (SELECT AVG(amount) FROM orders) THEN 'high'
        WHEN o.amount > (SELECT MIN(amount) FROM orders WHERE status = 'completed') THEN 'medium'
        ELSE 'low'
    END as order_level
FROM orders o;

-- 39. JSON表函数
SELECT 
    u.id,
    u.name,
    jt.hobby
FROM users u
CROSS JOIN JSON_TABLE(
    u.hobbies,
    '$[*]' COLUMNS (hobby VARCHAR(50) PATH '$')
) AS jt;

-- 40. 多表DELETE的另一种写法
DELETE t1, t2
FROM table1 t1
INNER JOIN table2 t2 ON t1.id = t2.ref_id
WHERE t1.created_at < '2024-01-01';

-- 41. 使用表值构造器
INSERT INTO test_data (id, value)
VALUES 
    ROW(1, 'a'),
    ROW(2, 'b'),
    ROW(3, 'c');

-- 42. 带PARTITION的表引用
SELECT * FROM sales PARTITION (p0, p1) WHERE year = 2024;

-- 43. 使用索引提示
SELECT * FROM orders USE INDEX (idx_created_at) WHERE created_at > '2025-01-01';

-- 44. STRAIGHT_JOIN
SELECT STRAIGHT_JOIN 
    c.name,
    o.order_date
FROM customers c
JOIN orders o ON c.id = o.customer_id;

-- 45. DERIVED表和子查询组合
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

-- 46. UNION多个表
SELECT id, name, 'customer' as type FROM customers
UNION ALL
SELECT id, name, 'employee' as type FROM employees  
UNION ALL
SELECT id, company_name, 'vendor' as type FROM vendors
UNION ALL
SELECT id, name, 'partner' as type FROM partners;

-- 47. 嵌套EXISTS
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

-- 48. NOT IN与子查询
SELECT *
FROM products p
WHERE p.category_id NOT IN (
    SELECT DISTINCT c.id 
    FROM categories c
    JOIN category_filters cf ON c.id = cf.category_id
    WHERE cf.is_hidden = 1
);

-- 49. HAVING子句中的子查询
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

-- 50. LOAD DATA INFILE (表名提取)
LOAD DATA LOCAL INFILE '/tmp/data.csv'
INTO TABLE mydb.import_temp
FIELDS TERMINATED BY ','
LINES TERMINATED BY '\n'
IGNORE 1 LINES
(col1, col2, col3);





