#!/bin/bash

# 測試優雅關閉功能

echo "🧪 測試優雅關閉功能..."

# 檢查配置文件是否存在
if [ ! -f "config.yaml" ]; then
    echo "⚠️  config.yaml 不存在，複製範例配置..."
    cp config.yaml.example config.yaml
fi

# 編譯程式
echo "🔨 編譯程式..."
go build -o test-graceful-shutdown .

if [ $? -ne 0 ]; then
    echo "❌ 編譯失敗"
    exit 1
fi

echo "✅ 編譯成功"

# 啟動程式（背景運行）
echo "🚀 啟動程式..."
./test-graceful-shutdown -c config.yaml &
PID=$!

echo "📝 程式 PID: $PID"

# 等待幾秒讓程式完全啟動
echo "⏳ 等待程式啟動..."
sleep 5

# 檢查程式是否還在運行
if ! kill -0 $PID 2>/dev/null; then
    echo "❌ 程式未能正常啟動"
    exit 1
fi

echo "✅ 程式已啟動"

# 發送 SIGINT 信號 (Ctrl+C)
echo "📤 發送 SIGINT 信號進行優雅關閉..."
kill -INT $PID

# 等待程式關閉，最多等待 15 秒
echo "⏳ 等待程式關閉（最多15秒）..."
for i in {1..15}; do
    if ! kill -0 $PID 2>/dev/null; then
        echo "✅ 程式在 ${i} 秒內優雅關閉"
        break
    fi
    sleep 1
    if [ $i -eq 15 ]; then
        echo "❌ 程式未在 15 秒內關閉，強制終止..."
        kill -KILL $PID 2>/dev/null
        echo "❌ 優雅關閉測試失敗"
        exit 1
    fi
done

echo "🎉 優雅關閉測試成功！"

# 清理
rm -f test-graceful-shutdown

echo "✨ 測試完成"