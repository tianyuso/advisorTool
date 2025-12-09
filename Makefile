.PHONY: build clean test help

# 默认目标
all: build

# 编译
build:
	@echo "Building advisor tool..."
	@mkdir -p build
	go build -o build/advisor ./cmd/advisor
	@echo "Done! Binary: build/advisor"

# 清理
clean:
	@echo "Cleaning..."
	rm -rf build/
	@echo "Done!"

# 测试（基本功能测试）
test: build
	@echo "Running basic tests..."
	@./build/advisor -version
	@echo "---"
	@echo "Test 1: MySQL SELECT * detection"
	-@./build/advisor -engine mysql -sql "SELECT * FROM users"
	@echo "---"
	@echo "Test 2: PostgreSQL DELETE without WHERE detection"
	-@./build/advisor -engine postgres -sql "DELETE FROM users"
	@echo "---"
	@echo "Test 3: MySQL naming convention check"
	-@./build/advisor -engine mysql -config examples/basic-config.yaml -sql "CREATE TABLE test (id INT PRIMARY KEY)"
	@echo ""
	@echo "All tests completed! (Exit codes indicate found issues, which is expected)"

# 列出所有规则
list-rules: build
	@./build/advisor -list-rules

# 生成示例配置
gen-config: build
	@echo "Generating MySQL config..."
	@./build/advisor -engine mysql -generate-config > examples/generated-mysql-config.yaml
	@echo "Generating PostgreSQL config..."
	@./build/advisor -engine postgres -generate-config > examples/generated-postgres-config.yaml
	@echo "Done!"

# 帮助
help:
	@echo "SQL Advisor Tool Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build       - Build the advisor tool"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make test        - Run basic tests"
	@echo "  make list-rules  - List all available rules"
	@echo "  make gen-config  - Generate sample config files"
	@echo "  make help        - Show this help message"

