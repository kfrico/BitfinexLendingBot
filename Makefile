# BitfinexLendingBot Makefile

.PHONY: build test clean run dev format lint help install

# 默認目標
.DEFAULT_GOAL := help

# 變量定義
BINARY_NAME=bitfinex-lending-bot
BUILD_DIR=build
CONFIG_FILE=config.yaml

# 顏色定義
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

## build: 編譯應用程式
build:
	@echo "$(GREEN)🔨 編譯應用程式...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✓ 編譯完成: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build: 編譯linux應用程式
build-linux:
	@echo "$(GREEN)🔨 編譯應用程式...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)✓ 編譯完成: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## test: 運行測試套件
test:
	@echo "$(GREEN)🧪 運行測試套件...$(NC)"
	@./test.sh

## test-quick: 快速測試（不包含覆蓋率）
test-quick:
	@echo "$(GREEN)⚡ 快速測試...$(NC)"
	@go test ./... -short

## test-verbose: 詳細測試輸出
test-verbose:
	@echo "$(GREEN)🔍 詳細測試...$(NC)"
	@go test -v ./...

## run: 運行應用程式（生產模式）
run: build
	@echo "$(GREEN)🚀 運行應用程式...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME) -c $(CONFIG_FILE)

## dev: 開發模式運行（測試模式）
dev:
	@echo "$(YELLOW)🧪 開發模式運行（測試模式）...$(NC)"
	@go run . -c $(CONFIG_FILE)

## format: 格式化代碼
format:
	@echo "$(GREEN)🎨 格式化代碼...$(NC)"
	@gofmt -w .
	@echo "$(GREEN)✓ 代碼格式化完成$(NC)"

## lint: 代碼靜態分析
lint:
	@echo "$(GREEN)🔍 代碼靜態分析...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ 靜態分析完成$(NC)"

## mod-tidy: 整理模組依賴
mod-tidy:
	@echo "$(GREEN)📦 整理模組依賴...$(NC)"
	@go mod tidy
	@go mod vendor
	@echo "$(GREEN)✓ 依賴整理完成$(NC)"

## clean: 清理編譯文件
clean:
	@echo "$(GREEN)🧹 清理編譯文件...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ 清理完成$(NC)"

## install: 安裝到系統路徑
install: build
	@echo "$(GREEN)📥 安裝應用程式...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✓ 安裝完成: /usr/local/bin/$(BINARY_NAME)$(NC)"

## uninstall: 從系統路徑卸載
uninstall:
	@echo "$(GREEN)📤 卸載應用程式...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ 卸載完成$(NC)"

## docker-build: 構建 Docker 鏡像
docker-build:
	@echo "$(GREEN)🐳 構建 Docker 鏡像...$(NC)"
	@docker build -t $(BINARY_NAME):latest .
	@echo "$(GREEN)✓ Docker 鏡像構建完成$(NC)"

## config-example: 複製配置範例
config-example:
	@echo "$(GREEN)📋 複製配置範例...$(NC)"
	@cp config.yaml.example config.yaml
	@echo "$(YELLOW)⚠ 請編輯 config.yaml 填入您的 API 密鑰$(NC)"

## security-check: 安全檢查
security-check:
	@echo "$(GREEN)🔒 安全檢查...$(NC)"
	@echo "檢查是否有敏感信息..."
	@! git log --oneline | grep -i "api\|key\|secret\|token" || echo "$(YELLOW)⚠ 發現可能包含敏感信息的提交$(NC)"
	@! find . -name "*.go" -o -name "*.yaml" -o -name "*.yml" | xargs grep -l "api.*key\|secret.*key" | grep -v "_test.go" | grep -v "config.yaml.example" || echo "$(YELLOW)⚠ 發現可能包含敏感信息的文件$(NC)"
	@echo "$(GREEN)✓ 安全檢查完成$(NC)"

## deps: 檢查和更新依賴
deps:
	@echo "$(GREEN)🔍 檢查依賴...$(NC)"
	@go list -u -m all
	@echo "$(GREEN)📥 下載依賴...$(NC)"
	@go mod download

## release: 構建發布版本
release: clean test build
	@echo "$(GREEN)🚀 構建發布版本...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/release/
	@cp config.yaml.example $(BUILD_DIR)/release/
	@cp README.md $(BUILD_DIR)/release/
	@cp SECURITY.md $(BUILD_DIR)/release/
	@cd $(BUILD_DIR)/release && tar -czf ../$(BINARY_NAME)-release.tar.gz .
	@echo "$(GREEN)✓ 發布包已生成: $(BUILD_DIR)/$(BINARY_NAME)-release.tar.gz$(NC)"

## help: 顯示幫助信息
help:
	@echo "$(GREEN)BitfinexLendingBot - 可用命令:$(NC)"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "$(YELLOW)使用示例:$(NC)"
	@echo "  make config-example  # 創建配置文件"
	@echo "  make dev            # 開發模式運行"
	@echo "  make test           # 運行完整測試"
	@echo "  make build          # 編譯應用程式"
	@echo "  make run            # 運行應用程式"