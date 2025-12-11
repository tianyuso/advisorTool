#!/bin/bash

echo "========================================"
echo "测试 SQL Advisor Tool Demo"
echo "========================================"
echo ""

echo "1. 检查文件结构..."
ls -la

echo ""
echo "2. 检查 common 目录..."
ls -la common/

echo ""
echo "3. 验证 Go 模块..."
go mod tidy

echo ""
echo "4. 编译检查..."
echo "   - basic_usage.go"
go build -o /tmp/basic_usage basic_usage.go
if [ $? -eq 0 ]; then
    echo "     ✅ 编译成功"
else
    echo "     ❌ 编译失败"
    exit 1
fi

echo "   - advanced_usage.go"
go build -o /tmp/advanced_usage advanced_usage.go
if [ $? -eq 0 ]; then
    echo "     ✅ 编译成功"
else
    echo "     ❌ 编译失败"
    exit 1
fi

echo "   - batch_review.go"
go build -o /tmp/batch_review batch_review.go
if [ $? -eq 0 ]; then
    echo "     ✅ 编译成功"
else
    echo "     ❌ 编译失败"
    exit 1
fi

echo ""
echo "========================================"
echo "✅ 所有测试通过！"
echo "========================================"
echo ""
echo "可以运行以下命令测试 demo:"
echo "  go run basic_usage.go"
echo "  go run advanced_usage.go"
echo "  go run batch_review.go"

