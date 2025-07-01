package constants

import "time"

// API 相關常量
const (
	FundingSymbolPrefix = "f"
	WalletTypeFunding   = "funding"
	OfferTypeLIMIT      = "LIMIT"
	DefaultPeriodDays   = 2
	Period30Days        = 30
	Period120Days       = 120
)

// 利率轉換常量
const (
	DaysPerYear         = 365
	PercentageToDecimal = 100.0
)

// 默認配置值
const (
	DefaultPriceLevels = 25
	MaxPriceLevels     = 100 // Bitfinex 最大允許值，根據 API 文檔只能是 1、25 或 100
	DefaultOrderLimit  = 3
	DefaultMinutesRun  = 15
)

// 時間相關常量
const (
	DefaultTimeout    = 30 * time.Second
	RetryDelay        = 5 * time.Second
	HourlyCheckMinute = 6
)

// Telegram 相關常量
const (
	TelegramCommandPrefix = "/"
	MaxMessageLength      = 4096
)

// 智能策略預設值
const (
	DefaultVolatilityThreshold = 0.002 // 0.2% 日利率波動閾值
	DefaultMaxRateMultiplier   = 2.0   // 最大2倍基礎利率
	DefaultMinRateMultiplier   = 0.8   // 最小0.8倍基礎利率

	// 建議值範圍
	RecommendedVolatilityMin = 0.001 // 保守用戶建議值
	RecommendedVolatilityMax = 0.003 // 激進用戶建議值
	RecommendedMaxRateMin    = 1.5   // 保守用戶建議值
	RecommendedMaxRateMax    = 3.0   // 激進用戶建議值
	RecommendedMinRateMin    = 0.7   // 激進用戶建議值
	RecommendedMinRateMax    = 0.9   // 保守用戶建議值
)
