-- 测试 PostgreSQL 影响行数计算（带 schema）
UPDATE mydata.test_users SET status = 2 WHERE id > 3;
DELETE FROM mydata.test_logs WHERE id <= 2;
UPDATE mydata.test_orders SET status = 'completed' FROM mydata.test_customers WHERE mydata.test_orders.user_id = mydata.test_customers.id AND mydata.test_customers.vip = TRUE;
