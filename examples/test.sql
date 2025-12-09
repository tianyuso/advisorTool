-- 示例 SQL 文件，用于测试 SQL 审核工具
-- Example SQL file for testing the SQL advisor tool

-- 1. 这个语句会触发 "SELECT *" 警告
SELECT * FROM users;

-- 2. 这个语句会触发 "缺少 WHERE 子句" 警告
SELECT id, name FROM users;

-- 3. 这个语句是正确的
SELECT id, name FROM users WHERE id = 1;

-- 4. 这个 UPDATE 语句会触发 "缺少 WHERE 子句" 错误
UPDATE users SET status = 'active';

-- 5. 这个 DELETE 语句会触发 "缺少 WHERE 子句" 错误
DELETE FROM orders;

-- 6. 这个建表语句会触发 "表没有主键" 错误
CREATE TABLE products (
    name VARCHAR(100),
    price DECIMAL(10,2)
);

-- 7. 这个建表语句是正确的
CREATE TABLE categories (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 8. PostgreSQL: 这个索引创建会触发 "需要使用 CONCURRENTLY" 警告
-- CREATE INDEX idx_users_name ON users(name);

-- 9. PostgreSQL: 正确的索引创建方式
-- CREATE INDEX CONCURRENTLY idx_users_name ON users(name);


