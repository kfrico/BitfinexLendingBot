package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli"

	"github.com/kfrico/BitfinexLendingBot/internal/bitfinex"
	"github.com/kfrico/BitfinexLendingBot/internal/config"
	"github.com/kfrico/BitfinexLendingBot/internal/constants"
	"github.com/kfrico/BitfinexLendingBot/internal/rates"
	"github.com/kfrico/BitfinexLendingBot/internal/strategy"
	"github.com/kfrico/BitfinexLendingBot/internal/telegram"
)

// Application 應用程式主結構
type Application struct {
	config        *config.Config
	bfxClient     *bitfinex.Client
	telegramBot   *telegram.Bot
	lendingBot    *strategy.LendingBot
	rateConverter *rates.Converter
}

// NewApplication 創建新的應用程式實例
func NewApplication(configPath string) (*Application, error) {
	// 載入配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 創建 Bitfinex 客戶端
	bfxClient := bitfinex.NewClient(cfg.BitfinexApiKey, cfg.BitfinexSecretKey)

	// 創建 Telegram 機器人
	telegramBot, err := telegram.NewBot(cfg, bfxClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	// 創建貸出機器人
	lendingBot := strategy.NewLendingBot(cfg, bfxClient)

	// 創建利率轉換器
	rateConverter := rates.NewConverter()

	app := &Application{
		config:        cfg,
		bfxClient:     bfxClient,
		telegramBot:   telegramBot,
		lendingBot:    lendingBot,
		rateConverter: rateConverter,
	}

	// 設置 Telegram bot 重啟回調
	telegramBot.SetRestartCallback(app.handleRestart)

	// 設置借貸機器人的通知回調
	lendingBot.SetNotifyCallback(telegramBot.SendNotification)

	// 設置 Telegram bot 的借貸機器人引用
	telegramBot.SetLendingBot(lendingBot)

	return app, nil
}

// Run 運行應用程式
func (app *Application) Run() error {
	log.Printf("Config loaded successfully: %+v", app.config)

	// 顯示運行模式
	if app.config.TestMode {
		log.Println("🧪 === 測試模式啟動 ===")
		log.Println("🧪 不會執行真實的下單操作")
		log.Println("🧪 但會執行真實的取消操作")
	} else {
		log.Println("🚀 === 正式模式啟動 ===")
		log.Println("🚀 將執行真實的交易操作")
	}

	// 啟動 Telegram 機器人
	go app.telegramBot.Start()

	// 啟動每小時利率檢查
	go app.scheduleHourlyRateCheck()

	// 啟動借貸訂單檢查
	go app.scheduleLendingCheck()

	log.Printf("Scheduler started at: %v", time.Now())
	log.Printf("⚙️ 主要任務間隔: %d 分鐘", app.config.MinutesRun)
	log.Printf("💰 借貸檢查間隔: %d 分鐘", app.config.LendingCheckMinutes)
	log.Printf("📊 利率檢查: 每小時")

	// 啟動主要業務邏輯調度
	app.scheduleMainTask()

	return nil
}

// scheduleMainTask 調度主要任務
func (app *Application) scheduleMainTask() {
	// 先執行第一次
	app.executeMainTask()

	ticker := time.NewTicker(time.Duration(app.config.MinutesRun) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		app.executeMainTask()
	}
}

// executeMainTask 執行主要任務
func (app *Application) executeMainTask() {
	if err := app.lendingBot.Execute(); err != nil {
		log.Printf("執行貸出策略失敗: %v", err)
	}
}

// scheduleHourlyRateCheck 調度每小時利率檢查
func (app *Application) scheduleHourlyRateCheck() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), constants.HourlyCheckMinute, 0, 0, now.Location())
		if now.After(next) || now.Equal(next) {
			next = next.Add(time.Hour)
		}

		delay := next.Sub(now)
		log.Printf("下次執行時間: %s, 等待時間: %s", next.Format("2006-01-02 15:04:05"), delay)

		time.Sleep(delay)
		app.checkRateThreshold()
	}
}

// checkRateThreshold 檢查利率閾值
func (app *Application) checkRateThreshold() {
	log.Println("定時檢查貸出利率...")

	exceeded, percentageRate, err := app.lendingBot.CheckRateThreshold()
	if err != nil {
		log.Printf("取得貸出利率失敗: %v", err)
		return
	}

	log.Printf("當前貸出利率: %.4f%%, 閾值: %.4f%%", percentageRate, app.config.NotifyRateThreshold)

	if exceeded {
		message := fmt.Sprintf("⚠️ 定時檢查提醒: 目前貸出利率 %.4f%% 已超過閾值 %.4f%%",
			percentageRate, app.config.NotifyRateThreshold)

		if err := app.telegramBot.SendNotification(message); err != nil {
			log.Printf("發送 Telegram 通知失敗: %v", err)
		} else {
			log.Printf("成功發送利率提醒")
		}
	} else {
		log.Println("當前利率低於閾值，無需發送通知")
	}
}

// handleRestart 處理重啟請求
func (app *Application) handleRestart() error {
	log.Println("收到重啟請求，開始執行重啟邏輯...")

	// 執行主要任務（這會取消所有訂單並重新下單）
	if err := app.lendingBot.Execute(); err != nil {
		log.Printf("重啟執行失敗: %v", err)
		return fmt.Errorf("重啟執行失敗: %w", err)
	}

	log.Println("重啟完成！")
	return nil
}

// scheduleLendingCheck 調度借貸訂單檢查
func (app *Application) scheduleLendingCheck() {
	// 先執行第一次檢查
	app.executeLendingCheck()

	ticker := time.NewTicker(time.Duration(app.config.LendingCheckMinutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		app.executeLendingCheck()
	}
}

// executeLendingCheck 執行借貸訂單檢查
func (app *Application) executeLendingCheck() {
	if err := app.lendingBot.CheckNewLendingCredits(); err != nil {
		log.Printf("檢查借貸訂單失敗: %v", err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "bitfinex-lending-bot"
	app.Version = "v2.0.0"
	app.Usage = "Automated Bitfinex lending bot with v2 API"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Value:  "config.yaml",
			Usage:  "Configuration file path",
			EnvVar: "CONFIG_PATH",
		},
	}

	app.Action = func(c *cli.Context) error {
		configPath := c.String("config")

		application, err := NewApplication(configPath)
		if err != nil {
			log.Fatalf("Failed to create application: %v", err)
		}

		return application.Run()
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
