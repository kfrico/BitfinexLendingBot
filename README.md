### BitfinexLendingBot 綠葉放貸機器人


### 運行服務

```
go run go run main.go
```

### 設定檔範例

```yaml
# Bitfinex API 設定
BITFINEX_API_KEY: "xxxxxxxxxx"    # Bitfinex API 金鑰
BITFINEX_SECRET_KEY: "xxxxxxxxxx" # Bitfinex API 密鑰

# 基本設定
CURRENCY: "usd"                   # 交易幣種 (例如：USD, BTC, ETH)
MINUTES_RUN: 60                   # 機器人執行間隔時間 (分鐘)

# 下單限制設定
ORDER_LIMIT: 3                    # 單次執行最大下單數量限制，避免一次掛太多都成立，錯過高利率
MIN_LOAN: 150                     # 最小貸出金額
MAX_LOAN: 150                     # 最大貸出金額限制

# 利率策略設定
MIN_DAILY_LEND_RATE: 0.015                  # 最低每日貸出利率
SPREAD_LEND: 30                             # 資金分散貸出的筆數
GAP_BOTTOM: 10                              # 利率階梯的底部區間（指ask掛單裡面第幾個index）
GAP_TOP: 5000                               # 利率階梯的頂部區間（指ask掛單裡面第幾個index）
THIRTY_DAY_LEND_RATE_THRESHOLD: 0.04        # 觸發30天期貸出的日利率閾值
ONE_TWENTY_DAY_LEND_RATE_THRESHOLD: 0.05    # 觸發120天期貸出的日利率閾值
RATE_BONUS: 0.002                           # 無掛單時的利率加成，避免訂單成功全都在低利率上

# 高額持有策略
HIGH_HOLD_RATE: 0.02                        # 高額持有策略的日利率
HIGH_HOLD_AMOUNT: 0                         # 高額持有策略的金額（設為0表示不使用）
HIGH_HOLD_ORDERS: 1                         # 高額持有策略的訂單數量

# 其他進階設定
RESERVE_AMOUNT: 0                           # 保留金額，不參與借貸
NOTIFY_RATE_THRESHOLD: 0.04                 # 利率通知閾值

# Telegram Bot 設定（可選）
TELEGRAM_BOT_TOKEN: ""                      # Telegram 機器人 Token
TELEGRAM_AUTH_TOKEN: ""                     # Telegram 驗證 Token
```

### Telegram 命令列表

透過Telegram機器人，您可以使用以下命令實時調整設定：

```
/rate                        - 顯示當前貸出利率和閾值
/check                       - 檢查貸出利率是否超過閾值
/status                      - 顯示系統狀態
/threshold [數值]            - 設置利率通知閾值
/reserve [數值]              - 設置不參與借貸的保留金額
/orderlimit [數值]           - 設置單次執行最大下單數量限制
/mindailylendrate [數值]     - 設置最低每日貸出利率
/highholdrate [數值]         - 設置高額持有策略的日利率
/highholdamount [數值]       - 設置高額持有策略的金額
/highholdorders [數值]       - 設置高額持有策略的訂單數量
/restart                     - 手動重新啟動，清除所有訂單，重新運行
/help                        - 顯示所有可用命令
```