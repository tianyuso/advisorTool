#!/bin/bash
# PostgreSQL 表存在性检查测试脚本

set -e

echo "=== PostgreSQL 表存在性检查测试 ==="
echo ""

# 数据库连接参数
DB_HOST="127.0.0.1"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="secret"
DB_NAME="mydb"
SCHEMA="mydata"

# 测试 SQL
TEST_SQL='CREATE TABLE "mydata"."user" (
  id BIGSERIAL not NULL,
  name TEXT NOT NULL UNIQUE,
  role TEXT NOT NULL DEFAULT '\''dev'\'' CHECK (role IN ('\''admin'\'', '\''dev'\'', '\''viewer'\'')),
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);'

echo "测试 SQL 语句:"
echo "$TEST_SQL"
echo ""

# 测试 1: 有数据库连接 - 应该检测到表已存在
echo "----------------------------------------"
echo "测试 1: 有数据库连接（应该检测到表已存在）"
echo "----------------------------------------"
./build/advisor -engine postgres \
  -sql "$TEST_SQL" \
  -host "$DB_HOST" \
  -port "$DB_PORT" \
  -user "$DB_USER" \
  -password "$DB_PASSWORD" \
  -dbname "$DB_NAME" \
  -schema "$SCHEMA" || echo "✅ 预期行为：检测到表已存在并返回错误"

echo ""
echo ""

# 测试 2: 无数据库连接 - 应该正常运行（不崩溃）
echo "----------------------------------------"
echo "测试 2: 无数据库连接（应该正常运行不崩溃）"
echo "----------------------------------------"
./build/advisor -engine postgres -sql "$TEST_SQL" && echo "✅ 预期行为：跳过元数据检查，正常运行"

echo ""
echo ""

# 测试 3: 创建不存在的表 - 应该通过审核（可能有警告）
NEW_TABLE_SQL='CREATE TABLE "mydata"."test_new_table_'$(date +%s)'" (
  id BIGSERIAL not NULL,
  name TEXT NOT NULL,
  PRIMARY KEY (id)
);'

echo "----------------------------------------"
echo "测试 3: 创建不存在的表（应该通过审核）"
echo "----------------------------------------"
echo "测试 SQL:"
echo "$NEW_TABLE_SQL"
echo ""
./build/advisor -engine postgres \
  -sql "$NEW_TABLE_SQL" \
  -host "$DB_HOST" \
  -port "$DB_PORT" \
  -user "$DB_USER" \
  -password "$DB_PASSWORD" \
  -dbname "$DB_NAME" \
  -schema "$SCHEMA" && echo "✅ 预期行为：表不存在，可以创建"

echo ""
echo "=== 测试完成 ==="

