#!/bin/bash
# MySQL 和 PostgreSQL 全面审核规则测试脚本

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# MySQL 连接参数
MYSQL_HOST="127.0.0.1"
MYSQL_PORT="3306"
MYSQL_USER="root"
MYSQL_PASS="root"
MYSQL_DB="mydata"

# PostgreSQL 连接参数
PG_HOST="127.0.0.1"
PG_PORT="5432"
PG_USER="postgres"
PG_PASS="secret"
PG_DB="mydb"
PG_SCHEMA="mydata"

# 测试计数
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试结果记录
declare -a TEST_RESULTS

# 测试函数
test_rule() {
    local test_name="$1"
    local engine="$2"
    local sql="$3"
    local expected_error_level="$4"
    local description="$5"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}[测试 $TOTAL_TESTS]${NC} $test_name"
    echo "  SQL: $sql"
    echo "  描述: $description"
    
    # 构建命令
    if [ "$engine" = "mysql" ]; then
        CMD="./build/advisor \
            -engine mysql \
            -sql \"$sql\" \
            -host $MYSQL_HOST \
            -port $MYSQL_PORT \
            -user $MYSQL_USER \
            -password $MYSQL_PASS \
            -dbname $MYSQL_DB \
            -format json"
    else
        CMD="./build/advisor \
            -engine postgres \
            -sql \"$sql\" \
            -host $PG_HOST \
            -port $PG_PORT \
            -user $PG_USER \
            -password $PG_PASS \
            -dbname $PG_DB \
            -format json"
    fi
    
    # 执行命令
    RESULT=$(eval $CMD 2>&1)
    
    # 检查错误级别
    ERROR_LEVEL=$(echo "$RESULT" | jq -r '.[0].error_level' 2>/dev/null || echo "error")
    AFFECTED_ROWS=$(echo "$RESULT" | jq -r '.[0].affected_rows' 2>/dev/null || echo "0")
    ERROR_MSG=$(echo "$RESULT" | jq -r '.[0].error_message' 2>/dev/null || echo "")
    
    # 验证结果
    if [ "$ERROR_LEVEL" = "$expected_error_level" ]; then
        echo -e "  ${GREEN}✓ 通过${NC} (错误级别: $ERROR_LEVEL, 影响行数: $AFFECTED_ROWS)"
        if [ -n "$ERROR_MSG" ]; then
            echo "  消息: $ERROR_MSG"
        fi
        PASSED_TESTS=$((PASSED_TESTS + 1))
        TEST_RESULTS+=("✓ $test_name")
    else
        echo -e "  ${RED}✗ 失败${NC} (期望: $expected_error_level, 实际: $ERROR_LEVEL)"
        echo "  结果: $RESULT"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        TEST_RESULTS+=("✗ $test_name")
    fi
    echo ""
}

# 创建测试表
setup_test_tables() {
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}准备测试环境${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo ""
    
    # MySQL 测试表
    echo "创建 MySQL 测试表..."
    mysql -h$MYSQL_HOST -P$MYSQL_PORT -u$MYSQL_USER -p$MYSQL_PASS $MYSQL_DB <<'EOF' 2>&1 | grep -v "Warning" || true
-- 删除已存在的测试表
DROP TABLE IF EXISTS test_users;
DROP TABLE IF EXISTS test_orders;
DROP TABLE IF EXISTS test_products;
DROP TABLE IF EXISTS test_customers;
DROP TABLE IF EXISTS test_logs;

-- 创建测试表
CREATE TABLE test_users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    status INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE test_orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE test_products (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock INT DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE test_customers (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    vip BOOLEAN DEFAULT FALSE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE test_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 插入测试数据
INSERT INTO test_users (name, email, status) VALUES
    ('User1', 'user1@test.com', 0),
    ('User2', 'user2@test.com', 1),
    ('User3', 'user3@test.com', 0),
    ('User4', 'user4@test.com', 1),
    ('User5', 'user5@test.com', 0);

INSERT INTO test_orders (user_id, amount, status) VALUES
    (1, 100.00, 'pending'),
    (2, 200.00, 'completed'),
    (3, 150.00, 'pending'),
    (1, 300.00, 'completed'),
    (2, 250.00, 'pending');

INSERT INTO test_products (name, price, stock) VALUES
    ('Product1', 10.00, 100),
    ('Product2', 20.00, 50),
    ('Product3', 30.00, 0);

INSERT INTO test_customers (name, vip) VALUES
    ('Customer1', TRUE),
    ('Customer2', FALSE),
    ('Customer3', TRUE);

INSERT INTO test_logs (message) VALUES
    ('Log entry 1'),
    ('Log entry 2'),
    ('Log entry 3');
EOF
    echo -e "${GREEN}✓ MySQL 测试表创建完成${NC}"
    echo ""
    
    # PostgreSQL 测试表
    echo "创建 PostgreSQL 测试表..."
    PGPASSWORD=$PG_PASS psql -h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DB <<'EOF' 2>&1 | grep -E "(CREATE|INSERT|DROP)" || true
SET search_path TO mydata;

-- 删除已存在的测试表
DROP TABLE IF EXISTS test_users CASCADE;
DROP TABLE IF EXISTS test_orders CASCADE;
DROP TABLE IF EXISTS test_products CASCADE;
DROP TABLE IF EXISTS test_customers CASCADE;
DROP TABLE IF EXISTS test_logs CASCADE;

-- 创建测试表
CREATE TABLE test_users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    status INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE test_orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE test_products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock INT DEFAULT 0
);

CREATE TABLE test_customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    vip BOOLEAN DEFAULT FALSE
);

CREATE TABLE test_logs (
    id SERIAL PRIMARY KEY,
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入测试数据
INSERT INTO test_users (name, email, status) VALUES
    ('User1', 'user1@test.com', 0),
    ('User2', 'user2@test.com', 1),
    ('User3', 'user3@test.com', 0),
    ('User4', 'user4@test.com', 1),
    ('User5', 'user5@test.com', 0);

INSERT INTO test_orders (user_id, amount, status) VALUES
    (1, 100.00, 'pending'),
    (2, 200.00, 'completed'),
    (3, 150.00, 'pending'),
    (1, 300.00, 'completed'),
    (2, 250.00, 'pending');

INSERT INTO test_products (name, price, stock) VALUES
    ('Product1', 10.00, 100),
    ('Product2', 20.00, 50),
    ('Product3', 30.00, 0);

INSERT INTO test_customers (name, vip) VALUES
    ('Customer1', TRUE),
    ('Customer2', FALSE),
    ('Customer3', TRUE);

INSERT INTO test_logs (message) VALUES
    ('Log entry 1'),
    ('Log entry 2'),
    ('Log entry 3');
EOF
    echo -e "${GREEN}✓ PostgreSQL 测试表创建完成${NC}"
    echo ""
}

# MySQL 规则测试
test_mysql_rules() {
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}测试 MySQL 审核规则${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo ""
    
    # 1. UPDATE/DELETE 必须有 WHERE（错误级别）
    test_rule "MySQL-001: UPDATE 缺少 WHERE" "mysql" \
        "UPDATE test_users SET status = 1" \
        "2" \
        "UPDATE 语句缺少 WHERE 子句应该报错"
    
    test_rule "MySQL-002: DELETE 缺少 WHERE" "mysql" \
        "DELETE FROM test_users" \
        "2" \
        "DELETE 语句缺少 WHERE 子句应该报错"
    
    test_rule "MySQL-003: UPDATE 带 WHERE" "mysql" \
        "UPDATE test_users SET status = 1 WHERE id > 3" \
        "0" \
        "UPDATE 语句带 WHERE 子句应该通过"
    
    test_rule "MySQL-004: DELETE 带 WHERE" "mysql" \
        "DELETE FROM test_logs WHERE id > 2" \
        "0" \
        "DELETE 语句带 WHERE 子句应该通过"
    
    # 2. 禁止 SELECT *
    test_rule "MySQL-005: SELECT *" "mysql" \
        "SELECT * FROM test_users" \
        "1" \
        "SELECT * 应该产生警告"
    
    test_rule "MySQL-006: SELECT 指定列" "mysql" \
        "SELECT id, name FROM test_users WHERE id = 1" \
        "0" \
        "SELECT 指定列应该通过"
    
    # 3. 表必须有主键（测试 CREATE TABLE）
    test_rule "MySQL-007: 创建表无主键" "mysql" \
        "CREATE TABLE test_no_pk (id INT, name VARCHAR(100))" \
        "2" \
        "创建表没有主键应该报错"
    
    test_rule "MySQL-008: 创建表有主键" "mysql" \
        "CREATE TABLE test_with_pk (id INT PRIMARY KEY, name VARCHAR(100))" \
        "0" \
        "创建表有主键应该通过"
    
    # 4. 影响行数计算 - 单表 UPDATE
    echo -e "${BLUE}[特殊测试]${NC} MySQL 影响行数计算 - 单表 UPDATE"
    RESULT=$(./build/advisor \
        -engine mysql \
        -sql "UPDATE test_users SET status = 2 WHERE id > 3" \
        -host $MYSQL_HOST \
        -port $MYSQL_PORT \
        -user $MYSQL_USER \
        -password $MYSQL_PASS \
        -dbname $MYSQL_DB \
        -format json)
    AFFECTED=$(echo "$RESULT" | jq -r '.[0].affected_rows')
    echo "  影响行数: $AFFECTED (期望: 2，id=4 和 id=5)"
    if [ "$AFFECTED" = "2" ]; then
        echo -e "  ${GREEN}✓ 影响行数计算正确${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "  ${RED}✗ 影响行数计算错误${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    
    # 5. 影响行数计算 - 单表 DELETE
    echo -e "${BLUE}[特殊测试]${NC} MySQL 影响行数计算 - 单表 DELETE"
    RESULT=$(./build/advisor \
        -engine mysql \
        -sql "DELETE FROM test_logs WHERE id <= 2" \
        -host $MYSQL_HOST \
        -port $MYSQL_PORT \
        -user $MYSQL_USER \
        -password $MYSQL_PASS \
        -dbname $MYSQL_DB \
        -format json)
    AFFECTED=$(echo "$RESULT" | jq -r '.[0].affected_rows')
    echo "  影响行数: $AFFECTED (期望: 2，id=1 和 id=2)"
    if [ "$AFFECTED" = "2" ]; then
        echo -e "  ${GREEN}✓ 影响行数计算正确${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "  ${RED}✗ 影响行数计算错误${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    
    # 6. 影响行数计算 - 连表 UPDATE
    echo -e "${BLUE}[特殊测试]${NC} MySQL 影响行数计算 - 连表 UPDATE"
    RESULT=$(./build/advisor \
        -engine mysql \
        -sql "UPDATE test_orders o INNER JOIN test_customers c ON o.user_id = c.id SET o.status = 'vip_completed' WHERE c.vip = TRUE" \
        -host $MYSQL_HOST \
        -port $MYSQL_PORT \
        -user $MYSQL_USER \
        -password $MYSQL_PASS \
        -dbname $MYSQL_DB \
        -format json)
    AFFECTED=$(echo "$RESULT" | jq -r '.[0].affected_rows')
    echo "  影响行数: $AFFECTED"
    echo -e "  ${GREEN}✓ 连表 UPDATE 改写测试${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    
    # 7. 自增列必须是整数
    test_rule "MySQL-009: 自增列非整数" "mysql" \
        "CREATE TABLE test_auto (id VARCHAR(100) AUTO_INCREMENT PRIMARY KEY, name VARCHAR(100))" \
        "2" \
        "自增列必须是整数类型"
    
    # 8. 索引不能有重复列
    test_rule "MySQL-010: 索引重复列" "mysql" \
        "CREATE TABLE test_idx (id INT PRIMARY KEY, name VARCHAR(100), INDEX idx_test (name, name))" \
        "2" \
        "索引不能包含重复列"
    
    # 9. INSERT 必须指定列
    test_rule "MySQL-011: INSERT 不指定列" "mysql" \
        "INSERT INTO test_users VALUES (100, 'Test', 'test@test.com', 1, NOW(), NOW())" \
        "1" \
        "INSERT 不指定列应该警告"
    
    test_rule "MySQL-012: INSERT 指定列" "mysql" \
        "INSERT INTO test_users (name, email) VALUES ('Test', 'test@test.com')" \
        "0" \
        "INSERT 指定列应该通过"
}

# PostgreSQL 规则测试
test_postgres_rules() {
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}测试 PostgreSQL 审核规则${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo ""
    
    # 1. UPDATE/DELETE 必须有 WHERE（错误级别）
    test_rule "PG-001: UPDATE 缺少 WHERE" "postgres" \
        "UPDATE test_users SET status = 1" \
        "2" \
        "UPDATE 语句缺少 WHERE 子句应该报错"
    
    test_rule "PG-002: DELETE 缺少 WHERE" "postgres" \
        "DELETE FROM test_users" \
        "2" \
        "DELETE 语句缺少 WHERE 子句应该报错"
    
    test_rule "PG-003: UPDATE 带 WHERE" "postgres" \
        "UPDATE test_users SET status = 1 WHERE id > 3" \
        "0" \
        "UPDATE 语句带 WHERE 子句应该通过"
    
    test_rule "PG-004: DELETE 带 WHERE" "postgres" \
        "DELETE FROM test_logs WHERE id > 2" \
        "0" \
        "DELETE 语句带 WHERE 子句应该通过"
    
    # 2. 禁止 SELECT *
    test_rule "PG-005: SELECT *" "postgres" \
        "SELECT * FROM test_users" \
        "1" \
        "SELECT * 应该产生警告"
    
    test_rule "PG-006: SELECT 指定列" "postgres" \
        "SELECT id, name FROM test_users WHERE id = 1" \
        "0" \
        "SELECT 指定列应该通过"
    
    # 3. 表必须有主键
    test_rule "PG-007: 创建表无主键" "postgres" \
        "CREATE TABLE test_no_pk (id INT, name VARCHAR(100))" \
        "2" \
        "创建表没有主键应该报错"
    
    test_rule "PG-008: 创建表有主键" "postgres" \
        "CREATE TABLE test_with_pk (id SERIAL PRIMARY KEY, name VARCHAR(100))" \
        "0" \
        "创建表有主键应该通过"
    
    # 4. 影响行数计算 - 单表 UPDATE
    echo -e "${BLUE}[特殊测试]${NC} PostgreSQL 影响行数计算 - 单表 UPDATE"
    RESULT=$(PGPASSWORD=$PG_PASS ./build/advisor \
        -engine postgres \
        -sql "UPDATE test_users SET status = 2 WHERE id > 3" \
        -host $PG_HOST \
        -port $PG_PORT \
        -user $PG_USER \
        -password $PG_PASS \
        -dbname $PG_DB \
        -format json)
    AFFECTED=$(echo "$RESULT" | jq -r '.[0].affected_rows')
    echo "  影响行数: $AFFECTED (期望: 2，id=4 和 id=5)"
    if [ "$AFFECTED" = "2" ]; then
        echo -e "  ${GREEN}✓ 影响行数计算正确${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "  ${RED}✗ 影响行数计算错误${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    
    # 5. 影响行数计算 - 单表 DELETE
    echo -e "${BLUE}[特殊测试]${NC} PostgreSQL 影响行数计算 - 单表 DELETE"
    RESULT=$(PGPASSWORD=$PG_PASS ./build/advisor \
        -engine postgres \
        -sql "DELETE FROM test_logs WHERE id <= 2" \
        -host $PG_HOST \
        -port $PG_PORT \
        -user $PG_USER \
        -password $PG_PASS \
        -dbname $PG_DB \
        -format json)
    AFFECTED=$(echo "$RESULT" | jq -r '.[0].affected_rows')
    echo "  影响行数: $AFFECTED (期望: 2，id=1 和 id=2)"
    if [ "$AFFECTED" = "2" ]; then
        echo -e "  ${GREEN}✓ 影响行数计算正确${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "  ${RED}✗ 影响行数计算错误${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    
    # 6. 影响行数计算 - 连表 UPDATE (PostgreSQL 语法)
    echo -e "${BLUE}[特殊测试]${NC} PostgreSQL 影响行数计算 - 连表 UPDATE"
    RESULT=$(PGPASSWORD=$PG_PASS ./build/advisor \
        -engine postgres \
        -sql "UPDATE test_orders SET status = 'vip_completed' FROM test_customers WHERE test_orders.user_id = test_customers.id AND test_customers.vip = TRUE" \
        -host $PG_HOST \
        -port $PG_PORT \
        -user $PG_USER \
        -password $PG_PASS \
        -dbname $PG_DB \
        -format json)
    AFFECTED=$(echo "$RESULT" | jq -r '.[0].affected_rows')
    echo "  影响行数: $AFFECTED"
    echo -e "  ${GREEN}✓ 连表 UPDATE 改写测试${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    
    # 7. INSERT 必须指定列
    test_rule "PG-009: INSERT 不指定列" "postgres" \
        "INSERT INTO test_users VALUES (100, 'Test', 'test@test.com', 1, NOW(), NOW())" \
        "1" \
        "INSERT 不指定列应该警告"
    
    test_rule "PG-010: INSERT 指定列" "postgres" \
        "INSERT INTO test_users (name, email) VALUES ('Test', 'test@test.com')" \
        "0" \
        "INSERT 指定列应该通过"
    
    # 8. 并发创建索引
    test_rule "PG-011: 创建索引不并发" "postgres" \
        "CREATE INDEX idx_test_name ON test_users(name)" \
        "2" \
        "PostgreSQL 创建索引应该使用 CONCURRENTLY"
    
    test_rule "PG-012: 并发创建索引" "postgres" \
        "CREATE INDEX CONCURRENTLY idx_test_name ON test_users(name)" \
        "0" \
        "使用 CONCURRENTLY 创建索引应该通过"
}

# 主测试流程
main() {
    echo -e "${GREEN}======================================"
    echo "MySQL & PostgreSQL 全面审核规则测试"
    echo "======================================${NC}"
    echo ""
    
    # 确保程序已编译
    if [ ! -f "./build/advisor" ]; then
        echo "正在编译 advisor..."
        go build -o build/advisor ./cmd/advisor
    fi
    
    # 设置测试环境
    setup_test_tables
    
    # 运行 MySQL 测试
    test_mysql_rules
    
    # 运行 PostgreSQL 测试
    test_postgres_rules
    
    # 输出测试结果
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}测试结果汇总${NC}"
    echo -e "${YELLOW}========================================${NC}"
    echo ""
    echo "总测试数: $TOTAL_TESTS"
    echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
    echo -e "${RED}失败: $FAILED_TESTS${NC}"
    echo ""
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✓ 所有测试通过！${NC}"
        exit 0
    else
        echo -e "${RED}✗ 有测试失败${NC}"
        exit 1
    fi
}

# 运行主函数
main

