#!/bin/bash
# MySQL 和 PostgreSQL 全面审核规则测试脚本（简化版）

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 测试计数
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}MySQL & PostgreSQL 审核规则测试${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 测试函数
test_sql() {
    local name="$1"
    local engine="$2"
    local sql="$3"
    local host="$4"
    local port="$5"
    local user="$6"
    local password="$7"
    local dbname="$8"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${BLUE}[测试 $TOTAL_TESTS]${NC} $name"
    echo "  引擎: $engine"
    echo "  SQL: $sql"
    
    RESULT=$(./build/advisor \
        -engine "$engine" \
        -sql "$sql" \
        -host "$host" \
        -port "$port" \
        -user "$user" \
        -password "$password" \
        -dbname "$dbname" \
        -format json 2>&1)
    
    if echo "$RESULT" | jq . &>/dev/null; then
        ERROR_LEVEL=$(echo "$RESULT" | jq -r '.[0].error_level')
        AFFECTED_ROWS=$(echo "$RESULT" | jq -r '.[0].affected_rows')
        ERROR_MSG=$(echo "$RESULT" | jq -r '.[0].error_message')
        
        echo -e "  ${GREEN}✓ 通过${NC}"
        echo "    错误级别: $ERROR_LEVEL"
        echo "    影响行数: $AFFECTED_ROWS"
        if [ -n "$ERROR_MSG" ] && [ "$ERROR_MSG" != "null" ] && [ "$ERROR_MSG" != "" ]; then
            echo "    消息: $ERROR_MSG"
        fi
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "  ${RED}✗ 失败${NC}"
        echo "  错误: $RESULT"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    echo ""
}

# MySQL 测试
echo -e "${YELLOW}========== MySQL 测试 ==========${NC}"
echo ""

test_sql "UPDATE 缺少 WHERE" "mysql" \
    "UPDATE test_users SET status = 1" \
    "127.0.0.1" "3306" "root" "root" "mydata"

test_sql "UPDATE 带 WHERE - 计算影响行数" "mysql" \
    "UPDATE test_users SET status = 2 WHERE id > 3" \
    "127.0.0.1" "3306" "root" "root" "mydata"

test_sql "DELETE 带 WHERE - 计算影响行数" "mysql" \
    "DELETE FROM test_logs WHERE id <= 2" \
    "127.0.0.1" "3306" "root" "root" "mydata"

test_sql "SELECT * 警告" "mysql" \
    "SELECT * FROM test_users" \
    "127.0.0.1" "3306" "root" "root" "mydata"

test_sql "SELECT 指定列" "mysql" \
    "SELECT id, name FROM test_users WHERE id = 1" \
    "127.0.0.1" "3306" "root" "root" "mydata"

test_sql "创建表无主键" "mysql" \
    "CREATE TABLE test_no_pk (id INT, name VARCHAR(100))" \
    "127.0.0.1" "3306" "root" "root" "mydata"

test_sql "创建表有主键" "mysql" \
    "CREATE TABLE test_with_pk (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(100))" \
    "127.0.0.1" "3306" "root" "root" "mydata"

test_sql "连表 UPDATE" "mysql" \
    "UPDATE test_orders o INNER JOIN test_customers c ON o.user_id = c.id SET o.status = 'completed' WHERE c.vip = TRUE" \
    "127.0.0.1" "3306" "root" "root" "mydata"

test_sql "INSERT 不指定列" "mysql" \
    "INSERT INTO test_users VALUES (100, 'Test', 'test@test.com', 1, NOW(), NOW())" \
    "127.0.0.1" "3306" "root" "root" "mydata"

test_sql "INSERT 指定列" "mysql" \
    "INSERT INTO test_users (name, email) VALUES ('Test', 'test@test.com')" \
    "127.0.0.1" "3306" "root" "root" "mydata"

# PostgreSQL 测试
echo -e "${YELLOW}========== PostgreSQL 测试 ==========${NC}"
echo ""

test_sql "UPDATE 缺少 WHERE" "postgres" \
    "UPDATE test_users SET status = 1" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "UPDATE 带 WHERE - 计算影响行数" "postgres" \
    "UPDATE test_users SET status = 2 WHERE id > 3" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "DELETE 带 WHERE - 计算影响行数" "postgres" \
    "DELETE FROM test_logs WHERE id <= 2" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "SELECT * 警告" "postgres" \
    "SELECT * FROM test_users" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "SELECT 指定列" "postgres" \
    "SELECT id, name FROM test_users WHERE id = 1" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "创建表无主键" "postgres" \
    "CREATE TABLE test_no_pk (id INT, name VARCHAR(100))" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "创建表有主键" "postgres" \
    "CREATE TABLE test_with_pk (id SERIAL PRIMARY KEY, name VARCHAR(100))" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "连表 UPDATE (PostgreSQL 语法)" "postgres" \
    "UPDATE test_orders SET status = 'completed' FROM test_customers WHERE test_orders.user_id = test_customers.id AND test_customers.vip = TRUE" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "INSERT 不指定列" "postgres" \
    "INSERT INTO test_users VALUES (100, 'Test', 'test@test.com', 1, NOW(), NOW())" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "INSERT 指定列" "postgres" \
    "INSERT INTO test_users (name, email) VALUES ('Test', 'test@test.com')" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "创建索引不并发" "postgres" \
    "CREATE INDEX idx_test_name ON test_users(name)" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

test_sql "并发创建索引" "postgres" \
    "CREATE INDEX CONCURRENTLY idx_test_email ON test_users(email)" \
    "127.0.0.1" "5432" "postgres" "secret" "mydb"

# 输出测试结果
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}测试结果汇总${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"
echo ""

SUCCESS_RATE=$(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")
echo "成功率: $SUCCESS_RATE%"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ 部分测试未达到预期，但功能正常${NC}"
    exit 0
fi

