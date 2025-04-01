### BitfinexLendingBot 綠葉放貸機器人


### 運行服務

```
go run go run main.go
```

### 設定檔範例

```
BITFINDEX_API_KEY: "xxxxxxxxxx"
BITFINDEX_SECRET_KEY: "xxxxxxxxxx"

CURRENCY: "usd"

ORDER_LIMIT: 3 # 每次掛單只掛幾筆，避免一次掛太多都成立，錯過高利率
MINUTES_RUN: 60 # 每隔幾分鐘執行一次

MIN_LOAN: 150 # 每筆最小金額
MAX_LOAN: 150 # 每筆最大金額

MIN_DAILY_LEND_RATE: 0.015 # 最小利率
SPREAD_LEND: 30 # 資金分配幾份
GAP_BOTTOM: 10 # 參數是指ask掛單裡面第幾個 index 下限
GAP_TOP: 5000 # 參數是指ask掛單裡面第幾個 index 上限
THIRTY_DAY_DAILY_THRESHOLD: 0.04 # 利率超過多少就掛120天的單
HIGH_HOLD_DAILY_RATE: 0.0
HIGH_HOLD_AMOUNT: 0

APPLY_BONUS_IF_NO_PENDING_ORDERS: 0.002 # 當下次執行時沒有未成功訂單時就加利率(避免訂單成功全都在低利率上)
```