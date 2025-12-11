#!/bin/bash
# 影响行数计算功能演示脚本

set -e

echo "=========================================="
echo "影响行数计算功能演示"
echo "=========================================="
echo ""

# 确保 advisor 已编译
if [ ! -f "./build/advisor" ]; then
    echo "正在编译 advisor..."
    go build -o build/advisor ./cmd/advisor
    echo "编译完成！"
    echo ""
fi

# 注意：以下示例需要实际的数据库连接
# 请根据实际情况修改连接参数

echo "演示 1: MySQL UPDATE 语句（不带数据库连接）"
echo "-------------------------------------------"
./build/advisor \
  -engine mysql \
  -sql "UPDATE users SET status = 1 WHERE created_at < '2024-01-01'" \
  -format json | jq '.[] | {sql: .sql, affected_rows: .affected_rows}'
echo ""

echo "演示 2: MySQL DELETE 语句（不带数据库连接）"
echo "-------------------------------------------"
./build/advisor \
  -engine mysql \
  -sql "DELETE FROM logs WHERE created_at < '2023-01-01'" \
  -format json | jq '.[] | {sql: .sql, affected_rows: .affected_rows}'
echo ""

echo "演示 3: PostgreSQL UPDATE 连表语句（不带数据库连接）"
echo "-------------------------------------------"
./build/advisor \
  -engine postgres \
  -sql "UPDATE table1 SET column1 = table2.column1 FROM table2 WHERE table1.id = table2.id" \
  -format json | jq '.[] | {sql: .sql, affected_rows: .affected_rows}'
echo ""

echo "演示 4: SQL Server DELETE 连表语句（不带数据库连接）"
echo "-------------------------------------------"
./build/advisor \
  -engine mssql \
  -sql "DELETE t1 FROM table1 t1 INNER JOIN table2 t2 ON t1.id = t2.id WHERE t1.status = 0" \
  -format json | jq '.[] | {sql: .sql, affected_rows: .affected_rows}'
echo ""

echo "=========================================="
echo "注意事项："
echo "1. 以上示例未提供数据库连接，affected_rows 为 0"
echo "2. 要获取实际影响行数，需要添加数据库连接参数："
echo "   -host <主机> -port <端口> -user <用户名> -password <密码> -dbname <数据库名>"
echo "=========================================="
echo ""

# 如果有测试数据库，可以取消注释以下代码进行实际测试
# echo "演示 5: MySQL 实际数据库测试（需要配置）"
# echo "-------------------------------------------"
# DB_HOST="localhost"
# DB_PORT="3306"
# DB_USER="root"
# DB_PASS="password"
# DB_NAME="testdb"
#
# ./build/advisor \
#   -engine mysql \
#   -sql "UPDATE users SET status = 1 WHERE id > 1000" \
#   -host "$DB_HOST" \
#   -port "$DB_PORT" \
#   -user "$DB_USER" \
#   -password "$DB_PASS" \
#   -dbname "$DB_NAME" \
#   -format json | jq '.[] | {sql: .sql, affected_rows: .affected_rows, error_message: .error_message}'

echo "单元测试运行："
echo "-------------------------------------------"
go test -v ./db -run TestRewrite 2>&1 | grep -E "(PASS|FAIL|RUN)"
echo ""
echo "所有测试完成！"

