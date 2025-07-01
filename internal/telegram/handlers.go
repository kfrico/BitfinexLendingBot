package telegram

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// handleRate 處理利率查詢指令
func (b *Bot) handleRate(chatID int64) {
	rate, err := b.bitfinexClient.GetCurrentFundingRate(b.config.GetFundingSymbol())
	if err != nil {
		b.sendMessage(chatID, "取得貸出利率失敗")
		return
	}

	thresholdInfo := ""
	if b.config.NotifyRateThreshold > 0 {
		thresholdInfo = fmt.Sprintf("\n目前設定的閾值為: %.4f%%", b.config.NotifyRateThreshold)
	}

	message := fmt.Sprintf("目前貸出利率: %.4f%%%s",
		b.rateConverter.DecimalDailyToPercentageDaily(rate), thresholdInfo)
	b.sendMessage(chatID, message)
}

// handleCheck 處理利率檢查指令
func (b *Bot) handleCheck(chatID int64) {
	rate, err := b.bitfinexClient.GetCurrentFundingRate(b.config.GetFundingSymbol())
	if err != nil {
		b.sendMessage(chatID, "取得貸出利率失敗")
		return
	}

	percentageRate := b.rateConverter.DecimalDailyToPercentageDaily(rate)
	replyMsg := fmt.Sprintf("當前貸出利率: %.4f%%\n閾值: %.4f%%",
		percentageRate, b.config.NotifyRateThreshold)

	if percentageRate > b.config.NotifyRateThreshold {
		replyMsg += "\n⚠️ 注意: 當前利率已超過閾值!"
	} else {
		replyMsg += "\n✓ 當前利率低於閾值"
	}

	b.sendMessage(chatID, replyMsg)
}

// handleStatus 處理狀態查詢指令
func (b *Bot) handleStatus(chatID int64) {
	statusMsg := fmt.Sprintf("目前系統狀態正常\n幣種: %s\n最小貸出金額: %.2f\n最大貸出金額: %.2f",
		b.config.Currency, b.config.MinLoan, b.config.MaxLoan)

	// 添加保留金額信息
	if b.config.ReserveAmount > 0 {
		statusMsg += fmt.Sprintf("\n保留金額: %.2f", b.config.ReserveAmount)
	} else {
		statusMsg += "\n未設置保留金額"
	}

	// 添加機器人運行參數
	statusMsg += fmt.Sprintf("\n\n機器人運行參數:")
	statusMsg += fmt.Sprintf("\n單次執行最大下單數量限制: %d", b.config.OrderLimit)
	statusMsg += fmt.Sprintf("\n最低每日貸出利率: %.4f%%", b.config.MinDailyLendRate)

	// 添加運行模式信息
	if b.config.TestMode {
		statusMsg += fmt.Sprintf("\n\n🧪 運行模式: 測試模式 (模擬交易)")
	} else {
		statusMsg += fmt.Sprintf("\n\n🚀 運行模式: 正式模式 (真實交易)")
	}

	// 添加高額持有策略信息
	statusMsg += fmt.Sprintf("\n\n高額持有策略:")
	if b.config.HighHoldAmount > 0 {
		statusMsg += fmt.Sprintf("\n金額: %.2f", b.config.HighHoldAmount)
		statusMsg += fmt.Sprintf("\n日利率: %.4f%%", b.config.HighHoldRate)
		statusMsg += fmt.Sprintf("\n訂單數量: %d", b.config.HighHoldOrders)
	} else {
		statusMsg += "\n未啟用高額持有策略"
	}

	b.sendMessage(chatID, statusMsg)
}

// handleSetThreshold 處理設置閾值指令
func (b *Bot) handleSetThreshold(chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		b.sendMessage(chatID, "格式錯誤，請使用 /threshold [數值] 格式")
		return
	}

	threshold, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || threshold <= 0 {
		b.sendMessage(chatID, "請輸入有效的正數值")
		return
	}

	b.config.NotifyRateThreshold = threshold
	b.sendMessage(chatID, fmt.Sprintf("閾值已設定為: %.4f%%", threshold))
}

// handleSetReserve 處理設置保留金額指令
func (b *Bot) handleSetReserve(chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		b.sendMessage(chatID, "格式錯誤，請使用 /reserve [數值] 格式")
		return
	}

	reserve, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || reserve < 0 {
		b.sendMessage(chatID, "請輸入有效的非負數值")
		return
	}

	b.config.ReserveAmount = reserve
	b.sendMessage(chatID, fmt.Sprintf("保留金額已設定為: %.2f", reserve))
}

// handleSetOrderLimit 處理設置訂單限制指令
func (b *Bot) handleSetOrderLimit(chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		b.sendMessage(chatID, "格式錯誤，請使用 /orderlimit [數值] 格式")
		return
	}

	limit, err := strconv.Atoi(parts[1])
	if err != nil || limit < 0 {
		b.sendMessage(chatID, "請輸入有效的非負整數")
		return
	}

	b.config.OrderLimit = limit
	b.sendMessage(chatID, fmt.Sprintf("單次執行最大下單數量限制已設定為: %d", limit))
}

// handleSetMinDailyRate 處理設置最低日利率指令
func (b *Bot) handleSetMinDailyRate(chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		b.sendMessage(chatID, "格式錯誤，請使用 /mindailylendrate [數值] 格式")
		return
	}

	rate, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || rate <= 0 {
		b.sendMessage(chatID, "請輸入有效的正數值")
		return
	}

	if !b.rateConverter.ValidatePercentageRate(rate) {
		b.sendMessage(chatID, "利率超出有效範圍 (0-7%)")
		return
	}

	b.config.MinDailyLendRate = rate
	b.sendMessage(chatID, fmt.Sprintf("最低每日貸出利率已設定為: %.4f%%", rate))
}

// handleSetHighHoldRate 處理設置高額持有利率指令
func (b *Bot) handleSetHighHoldRate(chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		b.sendMessage(chatID, "格式錯誤，請使用 /highholdrate [數值] 格式")
		return
	}

	rate, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || rate <= 0 {
		b.sendMessage(chatID, "請輸入有效的正數值")
		return
	}

	if !b.rateConverter.ValidatePercentageRate(rate) {
		b.sendMessage(chatID, "利率超出有效範圍 (0-7%)")
		return
	}

	b.config.HighHoldRate = rate
	b.sendMessage(chatID, fmt.Sprintf("高額持有策略的日利率已設定為: %.4f%%", rate))
}

// handleSetHighHoldAmount 處理設置高額持有金額指令
func (b *Bot) handleSetHighHoldAmount(chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		b.sendMessage(chatID, "格式錯誤，請使用 /highholdamount [數值] 格式")
		return
	}

	amount, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || amount <= 0 {
		b.sendMessage(chatID, "請輸入有效的正數值")
		return
	}

	b.config.HighHoldAmount = amount
	b.sendMessage(chatID, fmt.Sprintf("高額持有策略的金額已設定為: %.2f", amount))
}

// handleSetHighHoldOrders 處理設置高額持有訂單數量指令
func (b *Bot) handleSetHighHoldOrders(chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) != 2 {
		b.sendMessage(chatID, "格式錯誤，請使用 /highholdorders [數值] 格式")
		return
	}

	orders, err := strconv.Atoi(parts[1])
	if err != nil || orders < 1 {
		b.sendMessage(chatID, "請輸入有效的正整數")
		return
	}

	b.config.HighHoldOrders = orders
	b.sendMessage(chatID, fmt.Sprintf("高額持有訂單數量已設定為: %d", orders))
}

// handleRestart 處理重啟指令
func (b *Bot) handleRestart(chatID int64) {
	b.sendMessage(chatID, "🔄 開始手動重啟...")

	if b.restartCallback == nil {
		b.sendMessage(chatID, "❌ 重啟功能未初始化，請聯繫管理員")
		return
	}

	// 執行重啟邏輯
	err := b.restartCallback()
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ 重啟失敗: %v", err))
		return
	}

	b.sendMessage(chatID, "✅ 重啟完成！所有訂單已清除並重新下單")
}

// handleStrategyStatus 處理策略狀態查詢指令
func (b *Bot) handleStrategyStatus(chatID int64) {
	var strategyType string
	if b.config.EnableSmartStrategy {
		strategyType = "智能策略 (啟用)"
	} else {
		strategyType = "傳統策略 (啟用)"
	}

	statusMsg := fmt.Sprintf("📊 當前策略狀態\n策略類型: %s", strategyType)

	if b.config.EnableSmartStrategy {
		statusMsg += fmt.Sprintf("\n\n🧠 智能策略設定:")
		statusMsg += fmt.Sprintf("\n波動率閾值: %.4f", b.config.VolatilityThreshold)
		statusMsg += fmt.Sprintf("\n最大利率倍數: %.1fx", b.config.MaxRateMultiplier)
		statusMsg += fmt.Sprintf("\n最小利率倍數: %.1fx", b.config.MinRateMultiplier)

		// 添加建議值提示
		statusMsg += fmt.Sprintf("\n\n📋 參數建議值:")
		statusMsg += fmt.Sprintf("\n🛡️ 保守: 波動率 0.001, 最大倍數 1.5x, 最小倍數 0.9x")
		statusMsg += fmt.Sprintf("\n⚖️ 平衡: 波動率 0.002, 最大倍數 2.0x, 最小倍數 0.8x")
		statusMsg += fmt.Sprintf("\n⚡ 激進: 波動率 0.003, 最大倍數 3.0x, 最小倍數 0.7x")

		statusMsg += fmt.Sprintf("\n\n智能功能:")
		statusMsg += fmt.Sprintf("\n✅ 動態利率調整")
		statusMsg += fmt.Sprintf("\n✅ 市場趨勢分析")
		statusMsg += fmt.Sprintf("\n✅ 智能期間選擇")
		statusMsg += fmt.Sprintf("\n✅ 競爭對手分析")
		statusMsg += fmt.Sprintf("\n✅ 自適應資金配置")
	} else {
		statusMsg += fmt.Sprintf("\n\n⚙️ 傳統策略設定:")
		statusMsg += fmt.Sprintf("\n固定高額持有利率: %.4f%%", b.config.HighHoldRate)
		statusMsg += fmt.Sprintf("\n固定分散貸出參數")
		statusMsg += fmt.Sprintf("\n固定期間選擇邏輯")
	}

	statusMsg += fmt.Sprintf("\n\n💡 提示: 使用 /smartstrategy on/off 切換策略")

	b.sendMessage(chatID, statusMsg)
}

// handleToggleSmartStrategy 處理智能策略切換指令
func (b *Bot) handleToggleSmartStrategy(chatID int64, enable bool) {
	b.config.EnableSmartStrategy = enable

	var message string
	if enable {
		message = "✅ 智能策略已啟用\n\n智能功能:\n🧠 動態利率調整\n📈 市場趨勢分析\n⏰ 智能期間選擇\n🏆 競爭對手分析\n💰 自適應資金配置\n\n下次執行時將使用智能策略"
	} else {
		message = "❌ 智能策略已停用\n\n已切換回傳統策略:\n⚙️ 固定參數配置\n📊 傳統分散邏輯\n\n下次執行時將使用傳統策略"
	}

	b.sendMessage(chatID, message)
}

// handleLendingCredits 處理借貸訂單查看指令
func (b *Bot) handleLendingCredits(chatID int64) {
	if b.lendingBot == nil {
		b.sendMessage(chatID, "❌ 借貸機器人未初始化")
		return
	}

	credits, err := b.lendingBot.GetActiveLendingCredits()
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ 獲取借貸訂單失敗: %v", err))
		return
	}

	if len(credits) == 0 {
		b.sendMessage(chatID, "📭 目前沒有活躍的借貸訂單")
		return
	}

	message := "💰 當前活躍的借貸訂單\n\n"

	// 先計算所有訂單的統計信息
	totalAmount := 0.0
	totalDailyEarnings := 0.0
	totalPeriodEarnings := 0.0

	for _, credit := range credits {
		dailyEarnings := credit.Amount * credit.Rate
		periodEarnings := dailyEarnings * float64(credit.Period)

		totalAmount += credit.Amount
		totalDailyEarnings += dailyEarnings
		totalPeriodEarnings += periodEarnings
	}

	// 限制顯示數量，避免消息過長
	displayCount := len(credits)
	if displayCount > 10 {
		displayCount = 10
	}

	for i := 0; i < displayCount; i++ {
		credit := credits[i]

		// 計算收益
		dailyEarnings := credit.Amount * credit.Rate
		periodEarnings := dailyEarnings * float64(credit.Period)

		// 格式化開始時間
		openTime := time.Unix(credit.MTSOpened/1000, 0)

		message += fmt.Sprintf("📊 訂單 #%d (ID: %d)\n", i+1, credit.ID)
		message += fmt.Sprintf("💵 金額: %.2f %s\n", credit.Amount, b.config.Currency)
		message += fmt.Sprintf("📈 日利率: %.4f%%\n", b.rateConverter.DecimalToPercentage(credit.Rate))
		message += fmt.Sprintf("💰 日收益: %.4f %s\n", dailyEarnings, b.config.Currency)
		message += fmt.Sprintf("⏰ 期間: %d 天\n", credit.Period)
		message += fmt.Sprintf("💎 期間總收益: %.4f %s\n", periodEarnings, b.config.Currency)
		message += fmt.Sprintf("🕐 開始時間: %s\n", openTime.Format("2006-01-02 15:04:05"))
		message += fmt.Sprintf("📊 狀態: %s\n", credit.Status)
		message += "\n"
	}

	if len(credits) > 10 {
		message += fmt.Sprintf("... 還有 %d 個訂單未顯示\n\n", len(credits)-10)
	}

	// 添加統計信息
	message += fmt.Sprintf("📊 統計信息:\n")
	message += fmt.Sprintf("📦 總訂單數: %d\n", len(credits))
	message += fmt.Sprintf("💵 總借出金額: %.2f %s\n", totalAmount, b.config.Currency)
	message += fmt.Sprintf("💰 每日總收益: %.4f %s\n", totalDailyEarnings, b.config.Currency)

	if len(credits) <= 10 {
		message += fmt.Sprintf("💎 總期間收益: %.4f %s\n", totalPeriodEarnings, b.config.Currency)
	}

	// 計算年化收益率
	if totalAmount > 0 {
		annualRate := (totalDailyEarnings / totalAmount) * 365 * 100
		message += fmt.Sprintf("📈 年化收益率: %.2f%%", annualRate)
	}

	b.sendMessage(chatID, message)
}
