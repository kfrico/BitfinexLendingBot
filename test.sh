#!/bin/bash

# BitfinexLendingBot 測試腳本

set -e

echo "🧪 開始運行測試套件..."

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 函數：打印帶顏色的消息
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# 檢查 Go 版本
echo "🔍 檢查 Go 環境..."
if ! command -v go &> /dev/null; then
    print_error "Go 未安裝或不在 PATH 中"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
print_status "Go 版本: $GO_VERSION"

# 運行 go mod tidy
echo "📦 整理模組依賴..."
go mod tidy
print_status "依賴整理完成"

# 運行 go vet
echo "🔍 運行靜態分析 (go vet)..."
if go vet ./...; then
    print_status "靜態分析通過"
else
    print_error "靜態分析發現問題"
    exit 1
fi

# 運行測試
echo "🧪 運行單元測試..."
if go test -v ./... -race -coverprofile=coverage.out; then
    print_status "所有測試通過"
else
    print_error "測試失敗"
    exit 1
fi

# 生成測試覆蓋率報告
echo "📊 生成測試覆蓋率報告..."
if go tool cover -html=coverage.out -o coverage.html; then
    print_status "覆蓋率報告已生成: coverage.html"
else
    print_warning "無法生成覆蓋率報告"
fi

# 顯示簡要覆蓋率統計
if go tool cover -func=coverage.out | tail -1; then
    print_status "測試覆蓋率統計完成"
fi

# 運行格式檢查
echo "🎨 檢查代碼格式..."
UNFORMATTED=$(gofmt -l .)
if [ -z "$UNFORMATTED" ]; then
    print_status "代碼格式正確"
else
    print_warning "以下文件需要格式化:"
    echo "$UNFORMATTED"
    echo "運行 'gofmt -w .' 來修復格式問題"
fi

# 檢查是否有未提交的 go.mod 或 go.sum 變更
echo "📋 檢查模組文件變更..."
if git diff --exit-code go.mod go.sum; then
    print_status "模組文件無變更"
else
    print_warning "go.mod 或 go.sum 有變更，請檢查並提交"
fi

# 編譯檢查
echo "🔨 檢查編譯..."
if go build -o /tmp/bitfinex-lending-bot-test .; then
    print_status "編譯成功"
    rm -f /tmp/bitfinex-lending-bot-test
else
    print_error "編譯失敗"
    exit 1
fi

echo ""
echo "🎉 所有檢查完成！"
echo ""
echo "📋 測試總結:"
echo "   ✓ 靜態分析通過"
echo "   ✓ 單元測試通過"
echo "   ✓ 編譯成功"
echo "   📊 覆蓋率報告: coverage.html"
echo ""
echo "💡 提示："
echo "   - 運行 'go test -v ./...' 來重新運行測試"
echo "   - 運行 'go test -bench=.' 來運行性能測試（如果有的話）"
echo "   - 查看 coverage.html 了解詳細的測試覆蓋率"