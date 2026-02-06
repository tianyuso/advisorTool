#!/bin/bash

# 测试 extractObject 工具

cd /data/dev_go/advisorTool/extractObject

echo "========== 测试 MySQL =========="
go run examples/demo.go 2>&1 | grep -A 20 "MySQL表名"

echo ""
echo "========== 编译命令行工具 =========="
cd cmd
go build -o extractobject main.go

if [ -f extractobject ]; then
    echo "命令行工具编译成功！"
    
    echo ""
    echo "========== 测试命令行工具 =========="
    ./extractobject -db mysql -sql "SELECT u.id, o.order_id FROM mydb.users u JOIN orders o ON u.id = o.user_id"
    
    echo ""
    echo "========== JSON输出测试 =========="
    ./extractobject -db mysql -sql "SELECT * FROM mydb.users AS u" -json
else
    echo "编译失败"
    exit 1
fi
