package config

import (
	"os"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				BitfinexApiKey:     "test_api_key",
				BitfinexSecretKey:  "test_secret_key",
				Currency:           "USD",
				MinLoan:            150.0,
				MaxLoan:            1000.0,
				MinDailyLendRate:   0.02,
				SpreadLend:         30,
				GapBottom:          10,
				GapTop:             5000,
				EnableSmartStrategy: true,
				VolatilityThreshold: 0.002,
				MaxRateMultiplier:   2.0,
				MinRateMultiplier:   0.8,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: Config{
				BitfinexApiKey:    "",
				BitfinexSecretKey: "test_secret_key",
				Currency:          "USD",
				MinLoan:           150.0,
				MinDailyLendRate:  0.02,
				SpreadLend:        30,
				GapBottom:         10,
				GapTop:            5000,
			},
			wantErr: true,
		},
		{
			name: "placeholder API key",
			config: Config{
				BitfinexApiKey:    "your_api_key_here",
				BitfinexSecretKey: "test_secret_key",
				Currency:          "USD",
				MinLoan:           150.0,
				MinDailyLendRate:  0.02,
				SpreadLend:        30,
				GapBottom:         10,
				GapTop:            5000,
			},
			wantErr: true,
		},
		{
			name: "invalid min loan",
			config: Config{
				BitfinexApiKey:    "test_api_key",
				BitfinexSecretKey: "test_secret_key",
				Currency:          "USD",
				MinLoan:           -150.0,
				MinDailyLendRate:  0.02,
				SpreadLend:        30,
				GapBottom:         10,
				GapTop:            5000,
			},
			wantErr: true,
		},
		{
			name: "max loan less than min loan",
			config: Config{
				BitfinexApiKey:    "test_api_key",
				BitfinexSecretKey: "test_secret_key",
				Currency:          "USD",
				MinLoan:           1000.0,
				MaxLoan:           150.0,
				MinDailyLendRate:  0.02,
				SpreadLend:        30,
				GapBottom:         10,
				GapTop:            5000,
			},
			wantErr: true,
		},
		{
			name: "invalid smart strategy config",
			config: Config{
				BitfinexApiKey:      "test_api_key",
				BitfinexSecretKey:   "test_secret_key",
				Currency:            "USD",
				MinLoan:             150.0,
				MinDailyLendRate:    0.02,
				SpreadLend:          30,
				GapBottom:           10,
				GapTop:              5000,
				EnableSmartStrategy: true,
				VolatilityThreshold: 0.02, // 太大
				MaxRateMultiplier:   2.0,
				MinRateMultiplier:   0.8,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// 創建測試配置文件
	testConfigContent := `
BITFINEX_API_KEY: "test_api_key"
BITFINEX_SECRET_KEY: "test_secret_key"
CURRENCY: "USD"
MIN_LOAN: 150.0
MIN_DAILY_LEND_RATE: 0.02
SPREAD_LEND: 30
GAP_BOTTOM: 10
GAP_TOP: 5000
`

	// 創建臨時配置文件
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testConfigContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// 測試加載配置
	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// 驗證配置值
	if config.BitfinexApiKey != "test_api_key" {
		t.Errorf("Expected BitfinexApiKey to be 'test_api_key', got '%s'", config.BitfinexApiKey)
	}
	if config.Currency != "USD" {
		t.Errorf("Expected Currency to be 'USD', got '%s'", config.Currency)
	}
	if config.MinLoan != 150.0 {
		t.Errorf("Expected MinLoan to be 150.0, got %f", config.MinLoan)
	}
}

func TestGetFundingSymbol(t *testing.T) {
	config := &Config{Currency: "USD"}
	expected := "fUSD"
	result := config.GetFundingSymbol()
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGetMinDailyRateDecimal(t *testing.T) {
	config := &Config{MinDailyLendRate: 0.02}
	expected := 0.0002
	result := config.GetMinDailyRateDecimal()
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}