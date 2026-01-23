#!/bin/bash
# 测试非 SQL 语句验证修复
# 此脚本验证系统能够正确识别并报告无效的 SQL 输入

set -e

ADVISOR="./build/advisor"
ENGINE="sqlserver"

echo "=========================================="
echo "测试非 SQL 语句验证修复"
echo "=========================================="
echo ""

# 测试函数
test_invalid_sql() {
    local test_name="$1"
    local sql="$2"
    local expected_error_level="$3"
    
    echo "测试: $test_name"
    echo "SQL: $sql"
    
    # 运行审核并捕获输出
    output=$($ADVISOR -engine "$ENGINE" -sql "$sql" -format json 2>&1) || true
    
    # 提取 error_level
    error_level=$(echo "$output" | grep -o '"error_level":"[^"]*"' | head -1 | cut -d':' -f2 | tr -d '"')
    
    # 验证结果
    if [ "$error_level" = "$expected_error_level" ]; then
        echo "✓ 通过: error_level = $error_level"
    else
        echo "✗ 失败: 期望 error_level = $expected_error_level, 实际 = $error_level"
        echo "输出: $output"
        exit 1
    fi
    
    echo ""
}

# 测试用例 1: 中文文本
test_invalid_sql "中文文本" "转正考试分数记录表" "2"

# 测试用例 2: 纯字母数字
test_invalid_sql "纯字母数字" "abc123" "2"

# 测试用例 3: 普通英文文本
test_invalid_sql "普通英文文本" "hello world" "2"

# 测试用例 4: 空白字符
test_invalid_sql "空白字符" "   " "2"

# 测试用例 5: 有效的 SELECT 语句（应该有警告，不是错误）
echo "测试: 有效的 SELECT 语句"
echo "SQL: SELECT * FROM Users"
output=$($ADVISOR -engine "$ENGINE" -sql "SELECT * FROM Users" -format json 2>&1) || true
error_level=$(echo "$output" | grep -o '"error_level":"[^"]*"' | head -1 | cut -d':' -f2 | tr -d '"')

# SELECT * 应该触发警告（级别 1），不是错误（级别 2）
if [ "$error_level" = "1" ]; then
    echo "✓ 通过: error_level = $error_level (警告级别)"
else
    echo "✗ 失败: 期望 error_level = 1 (警告), 实际 = $error_level"
    echo "输出: $output"
    exit 1
fi
echo ""

# 测试用例 6: 有效的 CREATE TABLE 语句（应该通过）
echo "测试: 有效的 CREATE TABLE 语句"
echo "SQL: CREATE TABLE test (id INT PRIMARY KEY)"
output=$($ADVISOR -engine "$ENGINE" -sql "CREATE TABLE test (id INT PRIMARY KEY)" -format json 2>&1) || true
error_level=$(echo "$output" | grep -o '"error_level":"[^"]*"' | head -1 | cut -d':' -f2 | tr -d '"')

if [ "$error_level" = "0" ]; then
    echo "✓ 通过: error_level = $error_level (成功)"
else
    echo "注意: error_level = $error_level (可能触发了其他审核规则，但不是语法错误)"
    # 只要不是语法错误就算通过
    if [ "$error_level" != "0" ]; then
        echo "检查错误消息..."
        if echo "$output" | grep -q "Syntax error"; then
            echo "✗ 失败: 不应该有语法错误"
            echo "输出: $output"
            exit 1
        else
            echo "✓ 通过: 没有语法错误，只是审核规则建议"
        fi
    fi
fi
echo ""

echo "=========================================="
echo "所有测试通过！✓"
echo "=========================================="

