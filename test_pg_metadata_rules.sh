#!/bin/bash
# PostgreSQL 需要 Metadata 的规则测试脚本
# 
# 使用方法：
#   chmod +x test_pg_metadata_rules.sh
#   ./test_pg_metadata_rules.sh

# 数据库连接参数
HOST="127.0.0.1"
PORT="5432"
USER="postgres"
PASSWORD="secret"
DBNAME="mydb"

# advisor 路径
ADVISOR="./build/advisor"

echo "============================================================"
echo "PostgreSQL 需要 Metadata 的规则测试"
echo "============================================================"
echo "连接参数: host=$HOST port=$PORT user=$USER dbname=$DBNAME"
echo ""

# 函数：运行测试并格式化输出
run_test() {
    local desc="$1"
    local sql="$2"
    echo "-----------------------------------------------------------"
    echo "测试: $desc"
    echo "SQL: $sql"
    echo ""
    $ADVISOR -engine postgres \
        -host $HOST -port $PORT -user $USER -password $PASSWORD -dbname $DBNAME \
        -sql "$sql" -format json 2>&1 | python3 -m json.tool 2>/dev/null || cat
    echo ""
}

echo ""
echo "========== 1. column.no-null 规则测试 =========="
echo "该规则检查是否添加了允许 NULL 的列"
echo ""

run_test "添加允许 NULL 的列（应该警告）" \
    "ALTER TABLE mydata.employee ADD COLUMN nickname VARCHAR(50) NULL"

run_test "添加 NOT NULL 的列（应该通过）" \
    "ALTER TABLE mydata.employee ADD COLUMN salary NUMERIC(10,2) NOT NULL DEFAULT 0"

echo ""
echo "========== 2. column.require-default 规则测试 =========="
echo "该规则检查列是否有默认值"
echo ""

run_test "添加 NOT NULL 列但没有默认值（应该警告）" \
    "ALTER TABLE mydata.employee ADD COLUMN hire_date DATE NOT NULL"

run_test "添加列带默认值（应该通过）" \
    "ALTER TABLE mydata.employee ADD COLUMN status VARCHAR(20) DEFAULT 'active'"

echo ""
echo "========== 3. schema.backward-compatibility 规则测试 =========="
echo "该规则检查 Schema 变更是否向后兼容"
echo ""

run_test "删除列（不向后兼容，应该警告）" \
    "ALTER TABLE mydata.employee DROP COLUMN phone"

run_test "删除表（不向后兼容，应该警告）" \
    "DROP TABLE mydata.employee"

run_test "修改列类型变小（不向后兼容，应该警告）" \
    "ALTER TABLE mydata.employee ALTER COLUMN name TYPE VARCHAR(30)"

run_test "重命名列（不向后兼容，应该警告）" \
    "ALTER TABLE mydata.employee RENAME COLUMN phone TO mobile"

run_test "添加列（向后兼容，应该通过）" \
    "ALTER TABLE mydata.employee ADD COLUMN extra_info TEXT"

echo ""
echo "========== 4. 其他基础规则测试（不需要 Metadata）=========="
echo ""

run_test "CREATE TABLE 无主键（应该报错）" \
    "CREATE TABLE mydata.test_table (id INT, name VARCHAR(100))"

run_test "CREATE TABLE 有主键（应该通过）" \
    "CREATE TABLE mydata.test_table (id SERIAL PRIMARY KEY, name VARCHAR(100))"

run_test "UPDATE 不带 WHERE（应该报错）" \
    "UPDATE mydata.employee SET name = 'test'"

run_test "DELETE 不带 WHERE（应该报错）" \
    "DELETE FROM mydata.employee"

run_test "SELECT * 不带 WHERE（应该警告）" \
    "SELECT * FROM mydata.employee"

run_test "INSERT 不指定列（应该警告）" \
    "INSERT INTO mydata.employee VALUES (1, 'test', 'IT', 'dev', 25, '123456')"

echo ""
echo "============================================================"
echo "测试完成"
echo "============================================================"

