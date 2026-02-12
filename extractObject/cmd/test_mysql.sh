#!/bin/bash
# MySQL extractObject 工具快速测试脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=========================================="
echo "MySQL extractObject 工具测试"
echo "=========================================="
echo ""

# 检查工具是否存在
if [ ! -f "./extractobject" ]; then
    echo "错误: extractobject 工具不存在"
    echo "请先编译: go build -o extractobject main.go"
    exit 1
fi

echo "✓ 工具已就绪"
echo ""

# 测试1: 基础全面测试
echo "► 测试1: 基础全面测试 (30个场景)"
echo "----------------------------------------"
./extractobject -dbtype mysql -file test_mysql_comprehensive.sql > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ 基础测试通过"
    RESULT1=$(./extractobject -dbtype mysql -file test_mysql_comprehensive.sql 2>/dev/null | head -1)
    echo "  $RESULT1"
else
    echo "✗ 基础测试失败"
    exit 1
fi
echo ""

# 测试2: 边缘情况测试
echo "► 测试2: 边缘情况测试 (15个场景)"
echo "----------------------------------------"
./extractobject -dbtype mysql -file test_mysql_edge_cases.sql > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ 边缘测试通过"
    RESULT2=$(./extractobject -dbtype mysql -file test_mysql_edge_cases.sql 2>/dev/null | head -1)
    echo "  $RESULT2"
else
    echo "✗ 边缘测试失败"
    exit 1
fi
echo ""

# 测试3: JSON输出测试
echo "► 测试3: JSON输出格式"
echo "----------------------------------------"
JSON_OUTPUT=$(./extractobject -dbtype mysql -sql "SELECT * FROM users" -json 2>/dev/null)
if echo "$JSON_OUTPUT" | grep -q "TBName"; then
    echo "✓ JSON输出正常"
else
    echo "✗ JSON输出异常"
    exit 1
fi
echo ""

# 测试4: 单条SQL测试
echo "► 测试4: 单条SQL命令行测试"
echo "----------------------------------------"
SINGLE_TEST=$(./extractobject -dbtype mysql -sql "SELECT * FROM mydb.users u JOIN orders o ON u.id = o.user_id" 2>/dev/null | grep "找到")
if [ -n "$SINGLE_TEST" ]; then
    echo "✓ 单条SQL测试通过"
    echo "  $SINGLE_TEST"
else
    echo "✗ 单条SQL测试失败"
    exit 1
fi
echo ""

# 汇总
echo "=========================================="
echo "测试完成汇总"
echo "=========================================="
echo "✓ 基础全面测试 (30场景) - 通过"
echo "✓ 边缘情况测试 (15场景) - 通过"
echo "✓ JSON输出格式 - 正常"
echo "✓ 命令行SQL - 正常"
echo ""
echo "总体状态: ✅ 所有测试通过"
echo ""
echo "详细报告: 查看 MYSQL_TEST_REPORT.md"
echo ""


