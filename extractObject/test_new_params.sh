#!/bin/bash

# 测试新的小写数据库类型参数

echo "=========================================="
echo "测试新的小写数据库类型参数"
echo "=========================================="
echo ""

cd "$(dirname "$0")/cmd" || exit 1

# 确保工具已编译
if [ ! -f "./extractobject" ]; then
    echo "正在编译工具..."
    go build -o extractobject main.go
    echo ""
fi

echo "✓ 工具已就绪"
echo ""

# 测试计数器
passed=0
failed=0

# 测试函数
test_db() {
    local db_type=$1
    local db_name=$2
    local sql=$3
    
    echo "► 测试 $db_name (参数: -db $db_type)"
    output=$(./extractobject -db "$db_type" -sql "$sql" 2>&1)
    if [ $? -eq 0 ]; then
        echo "  ✓ 通过"
        ((passed++))
    else
        echo "  ✗ 失败: $output"
        ((failed++))
    fi
}

echo "【1】测试小写参数（推荐格式）"
echo "----------------------------------------"
test_db "mysql" "MySQL" "SELECT * FROM users"
test_db "postgres" "PostgreSQL" "SELECT * FROM users"
test_db "oracle" "Oracle" "SELECT * FROM users"
test_db "sqlserver" "SQL Server" "SELECT * FROM users"
echo ""

echo "【2】测试大写参数（向后兼容）"
echo "----------------------------------------"
test_db "MYSQL" "MySQL (大写)" "SELECT * FROM users"
test_db "POSTGRESQL" "PostgreSQL (大写)" "SELECT * FROM users"
test_db "ORACLE" "Oracle (大写)" "SELECT * FROM users"
test_db "SQLSERVER" "SQL Server (大写)" "SELECT * FROM users"
echo ""

echo "【3】测试别名参数"
echo "----------------------------------------"
test_db "postgresql" "PostgreSQL (全称)" "SELECT * FROM users"
test_db "mssql" "SQL Server (mssql)" "SELECT * FROM users"
echo ""

echo "【4】测试其他数据库类型"
echo "----------------------------------------"
test_db "tidb" "TiDB" "SELECT * FROM users"
test_db "mariadb" "MariaDB" "SELECT * FROM users"
test_db "oceanbase" "OceanBase" "SELECT * FROM users"
# Snowflake 暂不支持完整的SQL解析，跳过测试
# test_db "snowflake" "Snowflake" "SELECT * FROM users"
echo ""

echo "【5】测试无效参数（应该失败）"
echo "----------------------------------------"
echo "► 测试无效数据库类型"
output=$(./extractobject -db "invalid_db" -sql "SELECT * FROM users" 2>&1)
if [ $? -ne 0 ] && echo "$output" | grep -q "不支持的数据库类型"; then
    echo "  ✓ 正确返回错误信息"
    ((passed++))
else
    echo "  ✗ 错误处理异常"
    ((failed++))
fi
echo ""

echo "=========================================="
echo "测试结果汇总"
echo "=========================================="
echo "通过: $passed"
echo "失败: $failed"
echo ""

if [ $failed -eq 0 ]; then
    echo "✅ 所有测试通过！"
    echo ""
    echo "数据库类型参数已成功更新为小写格式"
    echo "支持的参数: mysql, postgres, oracle, sqlserver"
    echo "同时保持向后兼容: MYSQL, POSTGRESQL, ORACLE, SQLSERVER"
    exit 0
else
    echo "❌ 部分测试失败"
    exit 1
fi

