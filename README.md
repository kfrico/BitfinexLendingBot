# BitfinexLendingBot 綠葉放貸機器人

一個自動化的 Bitfinex 放貸機器人，支持智能策略、實時監控和 Telegram 通知。

## ✨ 主要功能

- 🔄 **自動放貸**: 根據市場狀況自動執行放貸策略
- 🧠 **智能策略**: 動態調整利率和期間選擇
- 📱 **Telegram 通知**: 即時通知新借貸成交和利率提醒
- 🛡️ **測試模式**: 安全測試模式，不執行真實交易
- 📊 **實時監控**: 查看活躍借貸訂單和收益統計
- ⚙️ **靈活配置**: 豐富的參數設定和實時調整

## 🚀 快速開始

### 安裝運行

```bash
# 編譯
go build -o bitfinex-lending-bot

# 運行（默認使用 config.yaml）
./bitfinex-lending-bot

# 指定配置文件
./bitfinex-lending-bot -c config.yaml
```

### 基本配置

複製 `config.yaml` 並設定您的 API 密鑰：

```yaml
# 必要設定
BITFINEX_API_KEY: "your_api_key_here"
BITFINEX_SECRET_KEY: "your_secret_key_here"
CURRENCY: "USD"
```

## 📋 完整配置參數

### 🔑 API 設定
```yaml
BITFINEX_API_KEY: "xxxxxxxxxx"        # Bitfinex API 金鑰
BITFINEX_SECRET_KEY: "xxxxxxxxxx"     # Bitfinex API 密鑰
```

### ⚙️ 基本設定
```yaml
CURRENCY: "USD"                       # 交易幣種 (USD, BTC, ETH...)
MINUTES_RUN: 15                       # 主要任務執行間隔 (分鐘)
ORDER_LIMIT: 2                        # 單次最大下單數量
MIN_LOAN: 150                         # 最小貸出金額
MAX_LOAN: 155                         # 最大貸出金額
```

### 📈 利率策略設定
```yaml
MIN_DAILY_LEND_RATE: 0.038           # 最低日利率 (%)
SPREAD_LEND: 30                       # 資金分散筆數
GAP_BOTTOM: 10                        # 訂單深度下限
GAP_TOP: 5000                         # 訂單深度上限
THIRTY_DAY_LEND_RATE_THRESHOLD: 0.04  # 30天期閾值 (%)
ONE_TWENTY_DAY_LEND_RATE_THRESHOLD: 0.045  # 120天期閾值 (%)
RATE_BONUS: 0.002                     # 無掛單時利率加成 (%)
```

### 💎 高額持有策略
```yaml
HIGH_HOLD_RATE: 0.1                  # 高額持有利率 (%)
HIGH_HOLD_AMOUNT: 155                # 高額持有金額 (0=停用)
HIGH_HOLD_ORDERS: 1                  # 高額持有訂單數
```

### 🧠 智能策略設定
```yaml
ENABLE_SMART_STRATEGY: true          # 啟用智能策略
VOLATILITY_THRESHOLD: 0.002          # 市場波動閾值
MAX_RATE_MULTIPLIER: 2.0            # 最大利率倍數
MIN_RATE_MULTIPLIER: 0.8            # 最小利率倍數
```

### 📱 通知設定
```yaml
TELEGRAM_BOT_TOKEN: "xxxxxxxxxx"     # Telegram 機器人 Token
TELEGRAM_AUTH_TOKEN: "your_auth"     # Telegram 驗證密碼
NOTIFY_RATE_THRESHOLD: 0.1           # 利率通知閾值 (%)
LENDING_CHECK_MINUTES: 10            # 借貸檢查間隔 (分鐘)
```

### 🛡️ 其他設定
```yaml
TEST_MODE: true                      # 測試模式 (true=不執行真實下單)
RESERVE_AMOUNT: 100                  # 保留金額
```

## 📱 Telegram 指令

### 📊 資訊查詢
```
/status                             - 系統狀態
/rate                              - 當前利率
/check                             - 利率閾值檢查
/lending                           - 活躍借貸訂單
/strategy                          - 策略狀態
```

### ⚙️ 參數調整
```
/threshold [數值]                   - 利率通知閾值
/reserve [數值]                     - 保留金額
/orderlimit [數值]                  - 下單數量限制
/mindailylendrate [數值]            - 最低日利率
/highholdrate [數值]               - 高額持有利率
/highholdamount [數值]             - 高額持有金額
/highholdorders [數值]             - 高額持有訂單數
```

### 🔧 控制指令
```
/smartstrategy on/off              - 智能策略開關
/restart                           - 重新啟動
/help                              - 指令說明
```

## 🎯 策略說明

### 傳統策略
- **高額持有**: 固定高利率長期放貸
- **分散放貸**: 根據市場深度分散投資
- **期間選擇**: 基於利率閾值選擇期間

### 智能策略
- **動態利率**: 根據市場趨勢調整利率
- **競爭分析**: 分析市場競爭狀況
- **自適應配置**: 動態調整資金分配比例
- **市場預測**: 基於歷史數據預測趨勢

## 🔄 運行模式

### 測試模式 (`TEST_MODE: true`)
- ✅ 不執行真實下單
- ✅ 執行真實取消 (清理用)
- ✅ 完整策略計算和日誌
- ✅ 正常檢查借貸訂單

### 正式模式 (`TEST_MODE: false`)
- 🚀 執行真實交易操作
- 💰 真實資金投入
- ⚠️ 請確保策略參數正確

## 📊 調度器架構

應用程式運行三個獨立調度器：

1. **主要任務** (`MINUTES_RUN`)
   - 執行下單、取消、策略計算
   
2. **借貸檢查** (`LENDING_CHECK_MINUTES`)
   - 檢查新借貸成交
   - 發送 Telegram 通知
   
3. **利率監控** (每小時)
   - 監控利率變化
   - 閾值提醒

## ⚠️ 注意事項

1. **API 權限**: 需要 Bitfinex API 的交易權限
2. **資金安全**: 建議先用小額測試
3. **利率設定**: 過低利率可能無法成交
4. **測試模式**: 首次使用建議啟用測試模式
5. **定期監控**: 建議定期檢查運行狀態

## 📚 相關文檔

- [借貸通知功能說明](LENDING_NOTIFICATION.md)
- [智能策略詳細說明](SMART_STRATEGY.md)

## 🆕 更新日誌

### v2.1.0
- ✨ 新增借貸訂單自動通知
- 🔧 修復統計信息計算錯誤
- ⚙️ 增加獨立的借貸檢查調度器
- 🧠 優化智能策略邏輯

### v2.0.0
- 🧠 新增智能策略系統
- 📱 完整的 Telegram 機器人功能
- 🛡️ 測試模式支持
- 📊 實時監控和統計

---

💡 **提示**: 首次使用建議啟用 `TEST_MODE: true` 進行測試，確認策略運行正常後再切換至正式模式。