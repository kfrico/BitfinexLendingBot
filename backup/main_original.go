package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/book"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/fundingoffer"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

// envStruct 儲存應用程式的環境變數設定
type envStruct struct {
	BitfinexApiKey                string  `mapstructure:"BITFINEX_API_KEY" json:"BITFINEX_API_KEY"`                                     // Bitfinex API 金鑰
	BitfinexSecretKey             string  `mapstructure:"BITFINEX_SECRET_KEY" json:"BITFINEX_SECRET_KEY"`                               // Bitfinex API 密鑰
	Currency                      string  `mapstructure:"CURRENCY" json:"CURRENCY"`                                                     // 交易幣種 (例如：USD, BTC, ETH)
	OrderLimit                    int     `mapstructure:"ORDER_LIMIT" json:"ORDER_LIMIT"`                                               // 單次執行最大下單數量限制
	MinutesRun                    int     `mapstructure:"MINUTES_RUN" json:"MINUTES_RUN"`                                               // 機器人執行間隔時間 (分鐘)
	MinLoan                       float64 `mapstructure:"MIN_LOAN" json:"MIN_LOAN"`                                                     // 最小貸出金額
	MaxLoan                       float64 `mapstructure:"MAX_LOAN" json:"MAX_LOAN"`                                                     // 最大貸出金額限制
	MinDailyLendRate              float64 `mapstructure:"MIN_DAILY_LEND_RATE" json:"MIN_DAILY_LEND_RATE"`                               // 最低每日貸出利率
	SpreadLend                    int     `mapstructure:"SPREAD_LEND" json:"SPREAD_LEND"`                                               // 資金分散貸出的筆數
	GapBottom                     float64 `mapstructure:"GAP_BOTTOM" json:"GAP_BOTTOM"`                                                 // 利率階梯的底部區間
	GapTop                        float64 `mapstructure:"GAP_TOP" json:"GAP_TOP"`                                                       // 利率階梯的頂部區間
	ThirtyDayLendRateThreshold    float64 `mapstructure:"THIRTY_DAY_LEND_RATE_THRESHOLD" json:"THIRTY_DAY_LEND_RATE_THRESHOLD"`         // 觸發30天期貸出的日利率閾值
	OneTwentyDayLendRateThreshold float64 `mapstructure:"ONE_TWENTY_DAY_LEND_RATE_THRESHOLD" json:"ONE_TWENTY_DAY_LEND_RATE_THRESHOLD"` // 觸發120天期貸出的日利率閾值
	HighHoldRate                  float64 `mapstructure:"HIGH_HOLD_RATE" json:"HIGH_HOLD_RATE"`                                         // 高額持有策略的日利率
	HighHoldAmount                float64 `mapstructure:"HIGH_HOLD_AMOUNT" json:"HIGH_HOLD_AMOUNT"`                                     // 高額持有策略的金額
	HighHoldOrders                int     `mapstructure:"HIGH_HOLD_ORDERS" json:"HIGH_HOLD_ORDERS"`                                     // 高額持有策略的訂單數量
	RateBonus                     float64 `mapstructure:"RATE_BONUS" json:"RATE_BONUS"`                                                 // 無掛單時的利率加成
	TelegramBotToken              string  `mapstructure:"TELEGRAM_BOT_TOKEN" json:"TELEGRAM_BOT_TOKEN"`                                 // Telegram 機器人 Token
	TelegramAuthToken             string  `mapstructure:"TELEGRAM_AUTH_TOKEN" json:"TELEGRAM_AUTH_TOKEN"`                               // Telegram 驗證 Token
	NotifyRateThreshold           float64 `mapstructure:"NOTIFY_RATE_THRESHOLD" json:"NOTIFY_RATE_THRESHOLD"`                           // 利率通知閾值
	ReserveAmount                 float64 `mapstructure:"RESERVE_AMOUNT" json:"RESERVE_AMOUNT"`                                         // 保留金額，不參與借貸
}

var (
	env    envStruct
	client *rest.Client
	bot    *tgbotapi.BotAPI
)

// MarginBotConf 設定機器人運作參數
type MarginBotConf struct {
	MinDailyLendRate              float64
	SpreadLend                    int
	GapBottom                     float64
	GapTop                        float64
	ThirtyDayLendRateThreshold    float64
	OneTwentyDayLendRateThreshold float64
	HighHoldRate                  float64
	HighHoldAmount                float64
	HighHoldOrders                int
	MinLoan                       float64
	MaxLoan                       float64
}

// MarginBotLoanOffer 貸出訂單資訊
type MarginBotLoanOffer struct {
	Amount float64
	Rate   float64
	Period int
}

// MarginBotLoanOffers 多筆貸出訂單陣列
type MarginBotLoanOffers []MarginBotLoanOffer

// 全局驗證映射表改為單一聊天ID
var authenticatedChatID int64
var chatIDMutex sync.Mutex

func main() {
	app := cli.NewApp()
	app.Name = "bitfindex-bot"
	app.Version = "v0.0.1"

	// 設定 CLI 參數
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Value:  "config.yaml",
			Usage:  "app config",
			EnvVar: "CONFIG_PATH",
		},
	}

	app.Action = runApp

	// 執行 CLI
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

// runApp 為主要的執行流程
func runApp(c *cli.Context) {
	loadConfig(c.String("config"))
	initBitfinexClient()
	initTelegramBot()

	go handleTelegramMessages()

	// 啟動每小時06分的貸出利率檢查
	go scheduleHourlyTask(6, checkLendRate)

	log.Println("ENV:", env)
	log.Println("Config 設定成功")

	fmt.Println("Scheduler started at:", time.Now())
	scheduleTask(env.MinutesRun, botRun)

	select {} // 阻塞主程式，使其持續執行
}

// loadConfig 讀取並解析設定檔案及環境變數
func loadConfig(configPath string) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&env); err != nil {
		panic(err)
	}
}

// initBitfinexClient 初始化 Bitfinex 客戶端
func initBitfinexClient() {
	client = rest.NewClient().Credentials(env.BitfinexApiKey, env.BitfinexSecretKey)
}

// initTelegramBot 初始化 Telegram bot 客戶端
func initTelegramBot() {
	var err error
	bot, err = tgbotapi.NewBotAPI(env.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
}

// scheduleTask 定時執行任務，每 n 分鐘執行一次
func scheduleTask(minutes int, task func()) {
	// 先執行第一次
	task()

	ticker := time.NewTicker(time.Duration(minutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		task()
	}
}

// scheduleHourlyTask 在每小時的指定分鐘執行任務
func scheduleHourlyTask(minute int, task func()) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), minute, 0, 0, now.Location())
		if now.After(next) || now.Equal(next) {
			next = next.Add(time.Hour)
		}

		delay := next.Sub(now)
		log.Printf("下次執行時間: %s, 等待時間: %s", next.Format("2006-01-02 15:04:05"), delay)

		time.Sleep(delay)
		task()
	}
}

// botRun 執行機器人主程式邏輯
func botRun() {
	fmt.Println("取消所有未完成訂單...")
	hasPendingOrders := cancelAllOffers()

	// 暫停幾秒避免取消訂單的金額還沒歸還
	time.Sleep(5 * time.Second)

	fmt.Println("取得可用額度...")
	fundsAvailable, err := getAvailableFunds(env.Currency)
	if err != nil {
		fmt.Println("取得餘額錯誤:", err)
		return
	}
	fmt.Printf("Currency: %s  Available: %f \n", env.Currency, fundsAvailable)

	// 扣除保留金額
	if env.ReserveAmount > 0 {
		fundsAvailable = math.Max(0, fundsAvailable-env.ReserveAmount)
		fmt.Printf("扣除保留金額後可用: %f \n", fundsAvailable)
	}

	// 若扣除保留金額後可用資金小於最小貸出額，則不進行操作
	if fundsAvailable < env.MinLoan {
		fmt.Println("可用資金小於最小貸出額，不進行操作")
		return
	}

	// 取得目前 Funding Book (Lendbook)
	fundingSymbol := "f" + strings.ToUpper(env.Currency)                         // 轉換為 funding symbol (例如：fUSD)
	lendbook, err := client.Book.All(fundingSymbol, common.PrecisionRawBook, 25) // 使用 R0 精度和默認價格水平
	if err != nil {
		fmt.Println("取得 Funding Book 錯誤:", err)
		return
	}

	// 依據機器人邏輯配置，算出要下單的貸出列表
	loanOffers := marginBotGetLoanOffers(
		fundsAvailable,
		lendbook,
		MarginBotConf{
			MinDailyLendRate:              env.MinDailyLendRate,
			SpreadLend:                    env.SpreadLend,
			GapBottom:                     env.GapBottom,
			GapTop:                        env.GapTop,
			ThirtyDayLendRateThreshold:    env.ThirtyDayLendRateThreshold,
			OneTwentyDayLendRateThreshold: env.OneTwentyDayLendRateThreshold,
			HighHoldRate:                  env.HighHoldRate,
			HighHoldAmount:                env.HighHoldAmount,
			HighHoldOrders:                env.HighHoldOrders,
			MinLoan:                       env.MinLoan,
			MaxLoan:                       env.MaxLoan,
		},
	)

	// 依照算出的貸出訂單逐筆下單，且控制在 OrderLimit 以內
	placeLoanOffers(loanOffers, env.OrderLimit, hasPendingOrders)
}

// cancelAllOffers 取消所有未完成訂單
func cancelAllOffers() (hasPendingOrders bool) {
	hasPendingOrders = false

	fundingSymbol := "f" + strings.ToUpper(env.Currency) // 轉換為 funding symbol
	offers, err := client.Funding.Offers(fundingSymbol)
	if err != nil {
		fmt.Println("取得未完成訂單失敗:", err)
		return hasPendingOrders
	}

	// 檢查是否有訂單數據
	if offers != nil && len(offers.Snapshot) > 0 {
		for _, offer := range offers.Snapshot {
			hasPendingOrders = true

			// 取消訂單
			cancelReq := &fundingoffer.CancelRequest{
				ID: offer.ID,
			}
			_, err := client.Funding.CancelOffer(cancelReq)
			if err != nil {
				fmt.Println("取消訂單失敗:", err)
			} else {
				fmt.Printf("成功取消訂單 ID: %d\n", offer.ID)
			}
		}
	} else {
		fmt.Println("目前沒有未完成的訂單")
	}

	return hasPendingOrders
}

// getAvailableFunds 取得指定幣別的可用餘額
func getAvailableFunds(currency string) (float64, error) {
	wallets, err := client.Wallet.Wallet()
	if err != nil {
		return 0, err
	}

	// 在 v2 API 中，我們需要尋找 funding 錢包類型
	for _, wallet := range wallets.Snapshot {
		if wallet.Currency == strings.ToUpper(currency) && wallet.Type == "funding" {
			return wallet.BalanceAvailable, nil
		}
	}
	return 0, nil
}

// placeLoanOffers 依照產生的貸出訂單陣列逐筆下單
func placeLoanOffers(loanOffers MarginBotLoanOffers, orderLimit int, hasPendingOrders bool) {
	orderCount := 0
	for _, o := range loanOffers {
		if orderLimit != 0 && orderCount >= orderLimit {
			break
		}

		if !hasPendingOrders {
			o.Rate = (o.Rate/365 + env.RateBonus) * 365
		}

		fmt.Printf("下單 => Rate: %.6f, Amount: %.4f, Period: %d \n", o.Rate/365, o.Amount, o.Period)

		// 創建 funding offer 請求
		fundingSymbol := "f" + strings.ToUpper(env.Currency)
		offerReq := &fundingoffer.SubmitRequest{
			Type:   "LIMIT",
			Symbol: fundingSymbol,
			Amount: o.Amount,
			Rate:   o.Rate / 365, // v2 API 使用日利率
			Period: int64(o.Period),
			Hidden: false,
		}

		_, err := client.Funding.SubmitOffer(offerReq)
		if err != nil {
			fmt.Println("下訂單失敗:", err)
		} else {
			orderCount++
		}
	}
}

// marginBotGetLoanOffers 計算並生成貸出訂單清單
func marginBotGetLoanOffers(
	fundsAvailable float64,
	lendbook *book.Snapshot,
	conf MarginBotConf,
) (loanOffers MarginBotLoanOffers) {

	// 如果可用資金小於最小貸出額，則不進行操作
	if fundsAvailable < conf.MinLoan {
		return
	}

	// 初始化可分配資金
	splitFundsAvailable := fundsAvailable

	// 高持有策略: 若 HighHoldAmount 大於最小貸出額，則執行高額持有策略
	if conf.HighHoldAmount > conf.MinLoan {
		// 檢查高額持有訂單數量設定
		ordersCount := conf.HighHoldOrders
		if ordersCount <= 0 {
			ordersCount = 1 // 如果未設置訂單數量或無效值，則默認為 1 筆
		}

		// 訂單金額
		highHold := conf.HighHoldAmount

		// 若設定了 MaxLoan，且 highHold 大於 MaxLoan，則裁切為 MaxLoan
		if conf.MaxLoan > 0 && highHold > conf.MaxLoan {
			highHold = conf.MaxLoan
		}

		// 創建多筆相同金額的高額持有訂單
		// 計算實際可以創建的訂單數量（基於可用資金）
		possibleOrders := int(splitFundsAvailable / highHold)
		actualOrders := math.Min(float64(ordersCount), float64(possibleOrders))

		// 下訂單
		for i := 0; i < int(actualOrders); i++ {
			// 確保每筆金額不超過剩餘資金
			if splitFundsAvailable < highHold {
				break
			}

			// 創建訂單
			tmp := MarginBotLoanOffer{
				Amount: highHold,
				Rate:   conf.HighHoldRate / 100 * 365, // 配置文件中是百分比，轉換為年化利率
				Period: 120,                           // 固定貸出 120 天
			}
			loanOffers = append(loanOffers, tmp)
			splitFundsAvailable -= highHold
		}
	}

	// 分割資金成多筆貸出
	numSplits := conf.SpreadLend
	if numSplits <= 0 || splitFundsAvailable < conf.MinLoan {
		return
	}

	// 計算每筆貸出金額 (初始)
	amtEach := splitFundsAvailable / float64(numSplits)
	amtEach = float64(int64(amtEach*100)) / 100.0 // 保留小數點後兩位

	// 若每筆金額小於最小貸出額，嘗試調降分割數
	for amtEach <= conf.MinLoan && numSplits > 1 {
		numSplits--
		amtEach = splitFundsAvailable / float64(numSplits)
		amtEach = float64(int64(amtEach*100)) / 100.0
	}
	if numSplits <= 0 {
		return
	}

	// 計算利率遞增量
	gapClimb := (conf.GapTop - conf.GapBottom) / float64(numSplits)
	nextLend := conf.GapBottom

	// 以市場深度遍歷，計算對應利率
	depthIndex := 0

	for numSplits > 0 {
		// 累計市場量至指定利率區間
		for float64(depthIndex) < nextLend && depthIndex < len(lendbook.Snapshot)-1 {
			depthIndex++
		}

		tmp := MarginBotLoanOffer{}

		// 依照計算出的 amtEach 與 MaxLoan 進行裁切
		allocAmount := amtEach
		// 若有設定 MaxLoan，且 allocAmount 大於 MaxLoan，則調整為 MaxLoan
		if conf.MaxLoan > 0 && allocAmount > conf.MaxLoan {
			allocAmount = conf.MaxLoan
		}
		tmp.Amount = allocAmount

		// 若計算後的金額仍小於 MinLoan, 則不需要下單
		if tmp.Amount < conf.MinLoan {
			break
		}

		// 依據市場利率 vs 最低日利率
		// 在 v2 API 中，book.Book.Rate 已經是日利率
		dailyRate := lendbook.Snapshot[depthIndex].Rate
		minDailyRate := conf.MinDailyLendRate / 100 // 配置文件中是百分比
		if dailyRate < minDailyRate {
			tmp.Rate = minDailyRate * 365 // 儲存為年化利率以保持兼容性
		} else {
			tmp.Rate = dailyRate * 365 // 轉換為年化利率以保持兼容性
		}

		if conf.OneTwentyDayLendRateThreshold > 0 && tmp.Rate >= (conf.OneTwentyDayLendRateThreshold/100)*365 {
			tmp.Period = 120 // 若市場年化利率高於閾值，則將訂單期間設定為 120 天
		} else if conf.ThirtyDayLendRateThreshold > 0 && tmp.Rate >= (conf.ThirtyDayLendRateThreshold/100)*365 {
			tmp.Period = 30 // 若市場年化利率高於閾值，則將訂單期間設定為 30 天
		} else {
			tmp.Period = 2 // 若市場年化利率低於閾值，則將訂單期間設定為 2 天
		}

		loanOffers = append(loanOffers, tmp)
		nextLend += gapClimb
		numSplits--
	}

	return
}

func handleTelegramMessages() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		// 處理身份驗證
		isAuthenticated := getAuthenticatedChatID() == chatID

		// 處理驗證過程
		if text == "/auth" {
			msg := tgbotapi.NewMessage(chatID, "請輸入驗證 token：")
			bot.Send(msg)
			continue
		} else if text == env.TelegramAuthToken {
			setAuthenticatedChatID(chatID)
			msg := tgbotapi.NewMessage(chatID, "驗證成功，現在可以傳送指令了")
			bot.Send(msg)
			continue
		} else if !isAuthenticated {
			msg := tgbotapi.NewMessage(chatID, "請先進行驗證，輸入 /auth 開始驗證流程")
			bot.Send(msg)
			continue
		}

		// 處理已驗證用戶的指令
		switch {
		case text == "/help" || text == "/start":
			// 顯示幫助訊息
			helpText := `可用指令:
/rate - 顯示當前貸出利率和閾值
/check - 檢查貸出利率是否超過閾值
/threshold [數值] - 設置利率通知閾值
/reserve [數值] - 設置不參與借貸的保留金額
/orderlimit [數值] - 設置單次執行最大下單數量限制
/mindailylendrate [數值] - 設置最低每日貸出利率
/highholdrate [數值] - 設置高額持有策略的日利率
/highholdamount [數值] - 設置高額持有策略的金額
/highholdorders [數值] - 設置高額持有策略的訂單數量
/status - 顯示系統狀態
/help - 顯示此幫助訊息
/restart - 手動重新啟動，清除所有訂單，重新運行`
			msg := tgbotapi.NewMessage(chatID, helpText)
			bot.Send(msg)

		// 手動重新啟動，清除所有訂單，重新運行
		case text == "/restart":
			botRun()
			msg := tgbotapi.NewMessage(chatID, "機器人已重新啟動，清除所有訂單，重新運行")
			bot.Send(msg)
		case text == "/rate":
			// 顯示當前貸出利率
			rate, err := getLendRate()
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "取得貸出利率失敗")
				bot.Send(msg)
			} else {
				thresholdInfo := ""
				if env.NotifyRateThreshold > 0 {
					thresholdInfo = fmt.Sprintf("\n目前設定的閾值為: %.4f%%", env.NotifyRateThreshold)
				}
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("目前貸出利率: %.4f%%%s", rate*100, thresholdInfo))
				bot.Send(msg)
			}
		case text == "/check":
			// 執行檢查並獲取結果
			rate, err := getLendRate()
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "取得貸出利率失敗")
				bot.Send(msg)
				continue
			}

			log.Printf("手動檢查: 當前貸出利率: %.4f", rate)

			replyMsg := fmt.Sprintf("當前貸出利率: %.4f%%\n閾值: %.4f%%", rate*100, env.NotifyRateThreshold)

			if rate*100 > env.NotifyRateThreshold {
				replyMsg += "\n⚠️ 注意: 當前利率已超過閾值!"
			} else {
				replyMsg += "\n✓ 當前利率低於閾值"
			}

			msg := tgbotapi.NewMessage(chatID, replyMsg)
			bot.Send(msg)

		case strings.HasPrefix(text, "/threshold "):
			// 設置閾值
			parts := strings.Split(text, " ")
			if len(parts) != 2 {
				msg := tgbotapi.NewMessage(chatID, "格式錯誤，請使用 /threshold [數值] 格式")
				bot.Send(msg)
				continue
			}

			threshold, err := strconv.ParseFloat(parts[1], 64)
			if err != nil || threshold <= 0 {
				msg := tgbotapi.NewMessage(chatID, "請輸入有效的正數值")
				bot.Send(msg)
				continue
			}

			env.NotifyRateThreshold = threshold

			// 理想情況下應該將新閾值保存到配置文件中
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("閾值已設定為: %.4f%%", threshold))
			bot.Send(msg)

		case strings.HasPrefix(text, "/reserve "):
			// 設置保留金額
			parts := strings.Split(text, " ")
			if len(parts) != 2 {
				msg := tgbotapi.NewMessage(chatID, "格式錯誤，請使用 /reserve [數值] 格式")
				bot.Send(msg)
				continue
			}

			reserve, err := strconv.ParseFloat(parts[1], 64)
			if err != nil || reserve < 0 {
				msg := tgbotapi.NewMessage(chatID, "請輸入有效的非負數值")
				bot.Send(msg)
				continue
			}

			env.ReserveAmount = reserve

			// 理想情況下應該將新保留金額保存到配置文件中
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("保留金額已設定為: %.2f", reserve))
			bot.Send(msg)

		case strings.HasPrefix(text, "/orderlimit "):
			// 設置單次執行最大下單數量限制
			parts := strings.Split(text, " ")
			if len(parts) != 2 {
				msg := tgbotapi.NewMessage(chatID, "格式錯誤，請使用 /orderlimit [數值] 格式")
				bot.Send(msg)
				continue
			}

			limit, err := strconv.Atoi(parts[1])
			if err != nil || limit < 0 {
				msg := tgbotapi.NewMessage(chatID, "請輸入有效的非負整數")
				bot.Send(msg)
				continue
			}

			env.OrderLimit = limit

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("單次執行最大下單數量限制已設定為: %d", limit))
			bot.Send(msg)

		case strings.HasPrefix(text, "/mindailylendrate "):
			// 設置最低每日貸出利率
			parts := strings.Split(text, " ")
			if len(parts) != 2 {
				msg := tgbotapi.NewMessage(chatID, "格式錯誤，請使用 /mindailylendrate [數值] 格式")
				bot.Send(msg)
				continue
			}

			rate, err := strconv.ParseFloat(parts[1], 64)
			if err != nil || rate <= 0 {
				msg := tgbotapi.NewMessage(chatID, "請輸入有效的正數值")
				bot.Send(msg)
				continue
			}

			env.MinDailyLendRate = rate

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("最低每日貸出利率已設定為: %.4f%%", rate))
			bot.Send(msg)

		case strings.HasPrefix(text, "/highholdrate "):
			// 設置高額持有策略的日利率
			parts := strings.Split(text, " ")
			if len(parts) != 2 {
				msg := tgbotapi.NewMessage(chatID, "格式錯誤，請使用 /highholdrate [數值] 格式")
				bot.Send(msg)
				continue
			}

			rate, err := strconv.ParseFloat(parts[1], 64)
			if err != nil || rate <= 0 {
				msg := tgbotapi.NewMessage(chatID, "請輸入有效的正數值")
				bot.Send(msg)
				continue
			}

			env.HighHoldRate = rate

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("高額持有策略的日利率已設定為: %.4f%%", rate))
			bot.Send(msg)

		case strings.HasPrefix(text, "/highholdamount "):
			// 設置高額持有策略的金額
			parts := strings.Split(text, " ")
			if len(parts) != 2 {
				msg := tgbotapi.NewMessage(chatID, "格式錯誤，請使用 /highholdamount [數值] 格式")
				bot.Send(msg)
				continue
			}

			amount, err := strconv.ParseFloat(parts[1], 64)
			if err != nil || amount <= 0 {
				msg := tgbotapi.NewMessage(chatID, "請輸入有效的正數值")
				bot.Send(msg)
				continue
			}

			env.HighHoldAmount = amount

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("高額持有策略的金額已設定為: %.2f", amount))
			bot.Send(msg)

		case strings.HasPrefix(text, "/highholdorders "):
			// 設置高額持有訂單數量
			parts := strings.Split(text, " ")
			if len(parts) != 2 {
				msg := tgbotapi.NewMessage(chatID, "格式錯誤，請使用 /highholdorders [數值] 格式")
				bot.Send(msg)
				continue
			}

			orders, err := strconv.Atoi(parts[1])
			if err != nil || orders < 1 {
				msg := tgbotapi.NewMessage(chatID, "請輸入有效的正整數")
				bot.Send(msg)
				continue
			}

			env.HighHoldOrders = orders

			// 理想情況下應該將新設定保存到配置文件中
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("高額持有訂單數量已設定為: %d", orders))
			bot.Send(msg)

		case text == "/status":
			statusMsg := fmt.Sprintf("目前系統狀態正常\n幣種: %s\n最小貸出金額: %.2f\n最大貸出金額: %.2f", env.Currency, env.MinLoan, env.MaxLoan)

			// 添加保留金額信息
			if env.ReserveAmount > 0 {
				statusMsg += fmt.Sprintf("\n保留金額: %.2f", env.ReserveAmount)
			} else {
				statusMsg += "\n未設置保留金額"
			}

			// 添加機器人運行參數
			statusMsg += fmt.Sprintf("\n\n機器人運行參數:")
			statusMsg += fmt.Sprintf("\n單次執行最大下單數量限制: %d", env.OrderLimit)
			statusMsg += fmt.Sprintf("\n最低每日貸出利率: %.4f%%", env.MinDailyLendRate)

			// 添加高額持有策略信息
			statusMsg += fmt.Sprintf("\n\n高額持有策略:")
			if env.HighHoldAmount > 0 {
				statusMsg += fmt.Sprintf("\n金額: %.2f", env.HighHoldAmount)
				statusMsg += fmt.Sprintf("\n日利率: %.4f%%", env.HighHoldRate)
				statusMsg += fmt.Sprintf("\n訂單數量: %d", env.HighHoldOrders)
			} else {
				statusMsg += "\n未啟用高額持有策略"
			}

			msg := tgbotapi.NewMessage(chatID, statusMsg)
			bot.Send(msg)

		default:
			msg := tgbotapi.NewMessage(chatID, "無效的指令，輸入 /help 查看所有可用指令")
			bot.Send(msg)
		}
	}
}

func getLendRate() (float64, error) {
	// 在 v2 API 中，我們使用 funding book 來獲取當前利率
	fundingSymbol := "f" + strings.ToUpper(env.Currency)
	book, err := client.Book.All(fundingSymbol, common.PrecisionRawBook, 25) // 使用 R0 精度

	if err != nil {
		return 0, err
	}

	if len(book.Snapshot) == 0 {
		return 0, fmt.Errorf("no funding book data available")
	}

	// 取得第一個 ask (貸出) 利率
	return book.Snapshot[0].Rate, nil
}

// checkLendRate 檢查貸出利率是否超過閾值，並在超過時發送通知
func checkLendRate() {
	log.Println("定時檢查貸出利率...")

	// 獲取當前貸出利率
	rate, err := getLendRate()
	if err != nil {
		log.Printf("取得貸出利率失敗: %v", err)
		return
	}

	log.Printf("當前貸出利率: %.4f%%, 閾值: %.4f%%", rate*100, env.NotifyRateThreshold)

	// 檢查是否需要發送通知
	if rate*100 > env.NotifyRateThreshold {
		chatID := getAuthenticatedChatID()
		if chatID == 0 {
			log.Println("尚未設定聊天ID，無法發送通知")
			return
		}

		notifyMsg := fmt.Sprintf("⚠️ 定時檢查提醒: 目前貸出利率 %.4f%% 已超過閾值 %.4f%%", rate*100, env.NotifyRateThreshold)
		msg := tgbotapi.NewMessage(chatID, notifyMsg)

		if _, err := bot.Send(msg); err != nil {
			log.Printf("發送 Telegram 通知失敗: %v", err)
		} else {
			log.Printf("成功發送利率提醒至聊天ID: %d", chatID)
		}
	} else {
		log.Println("當前利率低於閾值，無需發送通知")
	}
}

// setAuthenticatedChatID 設置已驗證的單一聊天ID
func setAuthenticatedChatID(chatID int64) {
	chatIDMutex.Lock()
	authenticatedChatID = chatID
	chatIDMutex.Unlock()
}

// getAuthenticatedChatID 獲取已驗證的單一聊天ID
func getAuthenticatedChatID() int64 {
	chatIDMutex.Lock()
	defer chatIDMutex.Unlock()

	return authenticatedChatID
}
