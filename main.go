package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bitfinexcom/bitfinex-api-go/v1"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

// envStruct 儲存應用程式的環境變數設定
type envStruct struct {
	BitfindexApiKey         string  `mapstructure:"BITFINDEX_API_KEY" json:"BITFINDEX_API_KEY"`
	BitfindexSecretKey      string  `mapstructure:"BITFINDEX_SECRET_KEY" json:"BITFINDEX_SECRET_KEY"`
	Currency                string  `mapstructure:"CURRENCY" json:"CURRENCY"`
	OrderLimit              int     `mapstructure:"ORDER_LIMIT" json:"ORDER_LIMIT"`
	MinutesRun              int     `mapstructure:"MINUTES_RUN" json:"MINUTES_RUN"`
	MinLoan                 float64 `mapstructure:"MIN_LOAN" json:"MIN_LOAN"`
	MaxLoan                 float64 `mapstructure:"MAX_LOAN" json:"MAX_LOAN"` // 新增最大貸出限制
	MinDailyLendRate        float64 `mapstructure:"MIN_DAILY_LEND_RATE" json:"MIN_DAILY_LEND_RATE"`
	SpreadLend              int     `mapstructure:"SPREAD_LEND" json:"SPREAD_LEND"`
	GapBottom               float64 `mapstructure:"GAP_BOTTOM" json:"GAP_BOTTOM"`
	GapTop                  float64 `mapstructure:"GAP_TOP" json:"GAP_TOP"`
	ThirtyDayDailyThreshold float64 `mapstructure:"THIRTY_DAY_DAILY_THRESHOLD" json:"THIRTY_DAY_DAILY_THRESHOLD"`
	HighHoldDailyRate       float64 `mapstructure:"HIGH_HOLD_DAILY_RATE" json:"HIGH_HOLD_DAILY_RATE"`
	HighHoldAmount          float64 `mapstructure:"HIGH_HOLD_AMOUNT" json:"HIGH_HOLD_AMOUNT"`
}

var (
	env    envStruct
	client *bitfinex.Client
)

// MarginBotConf 設定機器人運作參數
type MarginBotConf struct {
	MinDailyLendRate        float64
	SpreadLend              int
	GapBottom               float64
	GapTop                  float64
	ThirtyDayDailyThreshold float64
	HighHoldDailyRate       float64
	HighHoldAmount          float64
	MinLoan                 float64
	MaxLoan                 float64
}

// MarginBotLoanOffer 貸出訂單資訊
type MarginBotLoanOffer struct {
	Amount float64
	Rate   float64
	Period int
}

// MarginBotLoanOffers 多筆貸出訂單陣列
type MarginBotLoanOffers []MarginBotLoanOffer

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
	client = bitfinex.NewClient().Auth(env.BitfindexApiKey, env.BitfindexSecretKey)
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

// botRun 執行機器人主程式邏輯
func botRun() {
	fmt.Println("取消所有未完成訂單...")
	cancelAllOffers()

	fmt.Println("取得可用額度...")
	fundsAvailable, err := getAvailableFunds(env.Currency)
	if err != nil {
		fmt.Println("取得餘額錯誤:", err)
		return
	}
	fmt.Printf("Currency: %s  Available: %f \n", env.Currency, fundsAvailable)

	// 取得目前 Lendbook
	lendbook, err := client.Lendbook.Get(env.Currency, 0, 10000)
	if err != nil {
		fmt.Println("取得 Lendbook 錯誤:", err)
		return
	}

	// 依據機器人邏輯配置，算出要下單的貸出列表
	loanOffers := marginBotGetLoanOffers(
		fundsAvailable,
		lendbook,
		MarginBotConf{
			MinDailyLendRate:        env.MinDailyLendRate,
			SpreadLend:              env.SpreadLend,
			GapBottom:               env.GapBottom,
			GapTop:                  env.GapTop,
			ThirtyDayDailyThreshold: env.ThirtyDayDailyThreshold,
			HighHoldDailyRate:       env.HighHoldDailyRate,
			HighHoldAmount:          env.HighHoldAmount,
			MinLoan:                 env.MinLoan,
			MaxLoan:                 env.MaxLoan,
		},
	)

	// 依照算出的貸出訂單逐筆下單，且控制在 OrderLimit 以內
	placeLoanOffers(loanOffers, env.OrderLimit)
}

// cancelAllOffers 取消所有未完成訂單
func cancelAllOffers() {
	offers, err := client.Offers.Offers()
	if err != nil {
		fmt.Println("取得未完成訂單失敗:", err)
		return
	}

	for _, offer := range offers {
		if offer.Currency == strings.ToUpper(env.Currency) {
			o, err := client.Offers.Cancel(offer.Id)
			if err != nil {
				fmt.Println(err)
			}

			if err != nil {
				fmt.Println("取消訂單失敗:", err)
			} else {
				fmt.Printf("成功取消訂單: %+v\n", o)
			}
		}
	}
}

// getAvailableFunds 取得指定幣別的可用餘額
func getAvailableFunds(currency string) (float64, error) {
	balances, err := client.Balances.All()
	if err != nil {
		return 0, err
	}

	for _, balance := range balances {
		if balance.Currency == currency && balance.Type == "deposit" {
			return strconv.ParseFloat(balance.Available, 64)
		}
	}
	return 0, nil
}

// placeLoanOffers 依照產生的貸出訂單陣列逐筆下單
func placeLoanOffers(loanOffers MarginBotLoanOffers, orderLimit int) {
	orderCount := 0
	for _, o := range loanOffers {
		if orderLimit != 0 && orderCount >= orderLimit {
			break
		}
		fmt.Printf("下單 => Rate: %.6f, Amount: %.4f, Period: %d \n", o.Rate/365, o.Amount, o.Period)

		_, err := client.Offers.New(env.Currency, o.Amount, o.Rate, int64(o.Period), "lend")
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
	lendbook bitfinex.Lendbook,
	conf MarginBotConf,
) (loanOffers MarginBotLoanOffers) {

	// 如果可用資金小於最小貸出額，則不進行操作
	if fundsAvailable < conf.MinLoan {
		return
	}

	// 初始化可分配資金
	splitFundsAvailable := fundsAvailable

	// 高持有策略: 若 HighHoldAmount 大於最小貸出額，則預留這部分資金於 30 天長期訂單
	// 並同時檢查是否超過最大額度 (MaxLoan) 有設定時則做裁切
	if conf.HighHoldAmount > conf.MinLoan {
		highHold := math.Min(splitFundsAvailable, conf.HighHoldAmount)

		// 若設定了 MaxLoan，而且要扣除的資金超過 MaxLoan，則裁切為 MaxLoan
		if conf.MaxLoan > 0 && highHold > conf.MaxLoan {
			highHold = conf.MaxLoan
		}

		tmp := MarginBotLoanOffer{
			Amount: highHold,
			Rate:   conf.HighHoldDailyRate * 365, // 年化利率
			Period: 30,                           // 固定貸出 30 天
		}
		splitFundsAvailable -= highHold
		loanOffers = append(loanOffers, tmp)
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
		for float64(depthIndex) < nextLend && depthIndex < len(lendbook.Asks)-1 {
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
		rate, _ := strconv.ParseFloat(lendbook.Asks[depthIndex].Rate, 64)
		if rate < conf.MinDailyLendRate*365 {
			tmp.Rate = conf.MinDailyLendRate * 365
		} else {
			tmp.Rate = rate
		}

		// 若市場年化利率高於閾值，則將訂單期間設定為 30 天
		if conf.ThirtyDayDailyThreshold > 0 && rate >= conf.ThirtyDayDailyThreshold*365 {
			tmp.Period = 30
		} else {
			tmp.Period = 2
		}

		loanOffers = append(loanOffers, tmp)
		nextLend += gapClimb
		numSplits--
	}

	return
}
