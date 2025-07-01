# Bitfinex Lending Bot Makefile

# 變數定義
BINARY_NAME=BitfinexLendingBot
BINARY_NAME_LEGACY=BitfinexLendingBot-legacy
CONFIG_FILE=config.yaml

# Go 相關
GO_FILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
GOMOD_FILE := go.mod

.PHONY: help build build-legacy run run-legacy clean test lint format check-deps

# 默認目標
help: ## 顯示幫助信息
	@echo "可用的目標:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 構建目標
build: ## 構建主版本（重構版本）
	@echo "構建主版本..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) main.go

build-legacy: ## 構建舊版本（僅供參考）
	@echo "構建舊版本..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME_LEGACY) backup/main_original.go

# 運行目標
run: build ## 運行主版本
	@echo "運行主版本..."
	./$(BINARY_NAME) -c $(CONFIG_FILE)

run-legacy: build-legacy ## 運行舊版本
	@echo "運行舊版本..."
	./$(BINARY_NAME_LEGACY) -c $(CONFIG_FILE)

# 代碼質量目標
format: ## 格式化代碼
	@echo "格式化代碼..."
	gofmt -w $(GO_FILES)

lint: ## 檢查代碼規範
	@echo "檢查代碼規範..."
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安裝，跳過高級檢查"; \
	fi

test: ## 運行測試
	@echo "運行測試..."
	go test -v ./...

# 依賴管理
check-deps: ## 檢查並整理依賴
	@echo "檢查依賴..."
	go mod verify
	go mod tidy
	go mod download

# 清理目標
clean: ## 清理構建產物
	@echo "清理構建產物..."
	rm -f $(BINARY_NAME) $(BINARY_NAME_LEGACY)
	go clean

# 開發目標
dev: format lint build ## 開發環境準備（格式化、檢查、構建）
	@echo "開發環境準備完成"

# 部署目標
release: clean check-deps format lint test build ## 發佈準備
	@echo "發佈版本準備完成"

# 比較目標
compare: build build-legacy ## 構建兩個版本用於比較
	@echo "已構建兩個版本："
	@echo "  主版本: $(BINARY_NAME)"
	@echo "  舊版本: $(BINARY_NAME_LEGACY)"
	@echo ""
	@echo "運行指令："
	@echo "  make run         # 運行主版本"
	@echo "  make run-legacy  # 運行舊版本"