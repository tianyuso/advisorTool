#!/bin/bash

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║       extractObject - SQL表名提取工具 最终演示              ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

echo "【1】运行单元测试"
echo "----------------------------------------"
go test -v -run "TestMySQL|TestPostgreSQL|TestSQLServer" 2>&1 | grep -E "PASS|FAIL|RUN"
echo ""

echo "【2】命令行工具演示 - MySQL"
echo "----------------------------------------"
cd cmd
./extractobject -db mysql -sql "SELECT u.id, u.name FROM mydb.users u WHERE u.status = 'active'"
echo ""

echo "【3】命令行工具演示 - PostgreSQL"
echo "----------------------------------------"
./extractobject -db postgres -sql "SELECT * FROM public.products p JOIN public.categories c ON p.cat_id = c.id"
echo ""

echo "【4】JSON输出演示"
echo "----------------------------------------"
./extractobject -db mysql -sql "SELECT * FROM mydb.users" -json
echo ""

cd ..
echo "【5】库使用演示"
echo "----------------------------------------"
go run examples/simple_example.go
echo ""

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    演示完成！                                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
