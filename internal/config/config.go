package config

import (
	"strings"

	"github.com/kfrico/BitfinexLendingBot/internal/constants"
	"github.com/kfrico/BitfinexLendingBot/internal/errors"
	"github.com/spf13/viper"
)

// Config 應用程式配置結構
type Config struct {
	// API 配置
	BitfinexApiKey    string `mapstructure:"BITFINEX_API_KEY"`
	BitfinexSecretKey string `mapstructure:"BITFINEX_SECRET_KEY"`

	// 基本設定
	Currency   string `mapstructure:"CURRENCY"`
	OrderLimit int    `mapstructure:"ORDER_LIMIT"`

	RunOnlyOnNewCredits bool `mapstructure:"RUN_ONLY_ON_NEW_CREDITS"` // 是否僅在有新借貸訂單時執行重新計算訂單的任務(增加排隊訂單成交)
	// 如果 RunOnlyOnNewCredits 設定 true 則 MinutesRun 無用
	MinutesRun int `mapstructure:"MINUTES_RUN"` // 每隔幾分鐘清除訂單重新產生新訂單

	// 貸出限制
	MinLoan float64 `mapstructure:"MIN_LOAN"`
	MaxLoan float64 `mapstructure:"MAX_LOAN"`

	// 利率策略
	MinDailyLendRate              float64 `mapstructure:"MIN_DAILY_LEND_RATE"`
	SpreadLend                    int     `mapstructure:"SPREAD_LEND"`
	GapBottom                     float64 `mapstructure:"GAP_BOTTOM"`
	GapTop                        float64 `mapstructure:"GAP_TOP"`
	ThirtyDayLendRateThreshold    float64 `mapstructure:"THIRTY_DAY_LEND_RATE_THRESHOLD"`
	OneTwentyDayLendRateThreshold float64 `mapstructure:"ONE_TWENTY_DAY_LEND_RATE_THRESHOLD"`
	RateBonus                     float64 `mapstructure:"RATE_BONUS"`

	// 高額持有策略
	HighHoldRate   float64 `mapstructure:"HIGH_HOLD_RATE"`
	HighHoldAmount float64 `mapstructure:"HIGH_HOLD_AMOUNT"`
	HighHoldOrders int     `mapstructure:"HIGH_HOLD_ORDERS"`

	// Telegram 設定
	TelegramBotToken  string `mapstructure:"TELEGRAM_BOT_TOKEN"`
	TelegramAuthToken string `mapstructure:"TELEGRAM_AUTH_TOKEN"`

	// 通知設定
	NotifyRateThreshold float64 `mapstructure:"NOTIFY_RATE_THRESHOLD"`
	ReserveAmount       float64 `mapstructure:"RESERVE_AMOUNT"`

	// 智能策略設定
	EnableSmartStrategy      bool    `mapstructure:"ENABLE_SMART_STRATEGY"`
	VolatilityThreshold      float64 `mapstructure:"VOLATILITY_THRESHOLD"`
	MaxRateMultiplier        float64 `mapstructure:"MAX_RATE_MULTIPLIER"`
	MinRateMultiplier        float64 `mapstructure:"MIN_RATE_MULTIPLIER"`
	RateRangeIncreasePercent float64 `mapstructure:"RATE_RANGE_INCREASE_PERCENT"` // 利率範圍增加百分比

	// K線策略設定
	EnableKlineStrategy bool    `mapstructure:"ENABLE_KLINE_STRATEGY"` // 啟用K線策略
	KlineTimeFrame      string  `mapstructure:"KLINE_TIME_FRAME"`      // K線時間框架，預設15m
	KlinePeriod         int     `mapstructure:"KLINE_PERIOD"`          // K線週期數量，預設24（6小時）
	KlineSpreadPercent  float64 `mapstructure:"KLINE_SPREAD_PERCENT"`  // K線最高點加成百分比，預設0%
	KlineSmoothMethod   string  `mapstructure:"KLINE_SMOOTH_METHOD"`   // K線利率平滑方法：max, sma, ema, hla, p90

	// 測試模式設定
	TestMode bool `mapstructure:"TEST_MODE"`

	// 借貸通知設定
	LastLendingCheckTime int64 // 上次檢查借貸訂單的時間戳
	LendingCheckMinutes  int   `mapstructure:"LENDING_CHECK_MINUTES"` // 借貸訂單檢查間隔（分鐘）
}

// LoadConfig 從文件加載配置
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.NewConfigError("failed to read config file", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, errors.NewConfigError("failed to unmarshal config", err)
	}

	// 設置智能策略參數的預設值
	config.setSmartStrategyDefaults()

	// 設置K線策略參數的預設值
	config.setKlineStrategyDefaults()

	// 設置借貸檢查間隔的預設值
	config.setLendingCheckDefaults()

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate 驗證配置有效性
func (c *Config) Validate() error {
	if c.BitfinexApiKey == "" || c.BitfinexApiKey == "your_api_key_here" {
		return errors.NewValidationError("BITFINEX_API_KEY is required and must be set to your actual API key")
	}
	if c.BitfinexSecretKey == "" || c.BitfinexSecretKey == "your_secret_key_here" {
		return errors.NewValidationError("BITFINEX_SECRET_KEY is required and must be set to your actual secret key")
	}
	if c.Currency == "" {
		return errors.NewValidationError("CURRENCY is required")
	}
	if c.MinLoan <= 0 {
		return errors.NewValidationError("MIN_LOAN must be positive")
	}
	if c.MaxLoan > 0 && c.MaxLoan < c.MinLoan {
		return errors.NewValidationError("MAX_LOAN cannot be less than MIN_LOAN")
	}
	if c.MinDailyLendRate <= 0 {
		return errors.NewValidationError("MIN_DAILY_LEND_RATE must be positive")
	}
	if c.SpreadLend <= 0 {
		return errors.NewValidationError("SPREAD_LEND must be positive")
	}
	if c.GapBottom < 0 || c.GapTop < 0 || c.GapTop <= c.GapBottom {
		return errors.NewValidationError("invalid GAP_BOTTOM or GAP_TOP values")
	}

	// 驗證智能策略參數
	if c.EnableSmartStrategy {
		if c.VolatilityThreshold <= 0 || c.VolatilityThreshold > 0.01 {
			return errors.NewValidationError("VOLATILITY_THRESHOLD must be between 0 and 0.01")
		}
		if c.MaxRateMultiplier <= 1.0 || c.MaxRateMultiplier > 5.0 {
			return errors.NewValidationError("MAX_RATE_MULTIPLIER must be between 1.0 and 5.0")
		}
		if c.MinRateMultiplier < 0.1 || c.MinRateMultiplier >= 1.0 {
			return errors.NewValidationError("MIN_RATE_MULTIPLIER must be between 0.1 and 1.0")
		}
		if c.MinRateMultiplier >= c.MaxRateMultiplier {
			return errors.NewValidationError("MIN_RATE_MULTIPLIER must be less than MAX_RATE_MULTIPLIER")
		}
		if c.RateRangeIncreasePercent <= 0 || c.RateRangeIncreasePercent > 1.0 {
			return errors.NewValidationError("RATE_RANGE_INCREASE_PERCENT must be between 0 and 1.0 (0-100%)")
		}
	}

	// 驗證K線策略參數
	if c.EnableKlineStrategy {
		if c.KlineTimeFrame == "" {
			return errors.NewValidationError("KLINE_TIME_FRAME is required when ENABLE_KLINE_STRATEGY is true")
		}
		if c.KlinePeriod <= 0 {
			return errors.NewValidationError("KLINE_PERIOD must be positive")
		}
		if c.KlineSpreadPercent < 0 || c.KlineSpreadPercent > 100 {
			return errors.NewValidationError("KLINE_SPREAD_PERCENT must be between 0 and 100")
		}
		// 驗證平滑方法
		validMethods := []string{"max", "sma", "ema", "hla", "p90"}
		isValidMethod := false
		for _, method := range validMethods {
			if c.KlineSmoothMethod == method {
				isValidMethod = true
				break
			}
		}
		if !isValidMethod {
			return errors.NewValidationError("KLINE_SMOOTH_METHOD must be one of: max, sma, ema, hla, p90")
		}
	}

	// 驗證借貸檢查間隔
	if c.LendingCheckMinutes <= 0 {
		return errors.NewValidationError("LENDING_CHECK_MINUTES must be positive")
	}

	return nil
}

// GetFundingSymbol 獲取 funding symbol
func (c *Config) GetFundingSymbol() string {
	return constants.FundingSymbolPrefix + strings.ToUpper(c.Currency)
}

// GetMinDailyRateDecimal 獲取最低日利率（小數格式）
func (c *Config) GetMinDailyRateDecimal() float64 {
	return c.MinDailyLendRate / constants.PercentageToDecimal
}

// GetHighHoldRateDecimal 獲取高額持有利率（小數格式）
func (c *Config) GetHighHoldRateDecimal() float64 {
	return c.HighHoldRate / constants.PercentageToDecimal
}

// GetThirtyDayThresholdDecimal 獲取30天閾值（小數格式）
func (c *Config) GetThirtyDayThresholdDecimal() float64 {
	return c.ThirtyDayLendRateThreshold / constants.PercentageToDecimal
}

// GetOneTwentyDayThresholdDecimal 獲取120天閾值（小數格式）
func (c *Config) GetOneTwentyDayThresholdDecimal() float64 {
	return c.OneTwentyDayLendRateThreshold / constants.PercentageToDecimal
}

// setSmartStrategyDefaults 設置智能策略參數的預設值
func (c *Config) setSmartStrategyDefaults() {
	// 如果智能策略啟用但參數為零，設置建議的預設值
	if c.EnableSmartStrategy {
		if c.VolatilityThreshold == 0 {
			c.VolatilityThreshold = constants.DefaultVolatilityThreshold
		}
		if c.MaxRateMultiplier == 0 {
			c.MaxRateMultiplier = constants.DefaultMaxRateMultiplier
		}
		if c.MinRateMultiplier == 0 {
			c.MinRateMultiplier = constants.DefaultMinRateMultiplier
		}
		if c.RateRangeIncreasePercent == 0 {
			c.RateRangeIncreasePercent = constants.RateRangeIncreasePercent
		}
	} else {
		// 如果智能策略未啟用，確保參數有預設值以防止驗證錯誤
		if c.VolatilityThreshold == 0 {
			c.VolatilityThreshold = constants.DefaultVolatilityThreshold
		}
		if c.MaxRateMultiplier == 0 {
			c.MaxRateMultiplier = constants.DefaultMaxRateMultiplier
		}
		if c.MinRateMultiplier == 0 {
			c.MinRateMultiplier = constants.DefaultMinRateMultiplier
		}
		if c.RateRangeIncreasePercent == 0 {
			c.RateRangeIncreasePercent = constants.RateRangeIncreasePercent
		}
	}
}

// setKlineStrategyDefaults 設置K線策略參數的預設值
func (c *Config) setKlineStrategyDefaults() {
	// 如果K線策略啟用但參數為空，設置預設值
	if c.EnableKlineStrategy {
		if c.KlineTimeFrame == "" {
			c.KlineTimeFrame = "15m"
		}
		if c.KlinePeriod == 0 {
			c.KlinePeriod = 24 // 6小時的15分鐘K線
		}
		if c.KlineSpreadPercent == 0 {
			c.KlineSpreadPercent = 0.0 // 0%加成
		}
		if c.KlineSmoothMethod == "" {
			c.KlineSmoothMethod = "ema" // 預設使用指數移動平均
		}
	}
}

// setLendingCheckDefaults 設置借貸檢查間隔的預設值
func (c *Config) setLendingCheckDefaults() {
	// 如果未設置借貸檢查間隔，預設為 10 分鐘
	if c.LendingCheckMinutes == 0 {
		c.LendingCheckMinutes = 10
	}
}
