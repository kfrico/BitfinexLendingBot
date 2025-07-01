package strategy

import (
	"math"
	"time"

	"github.com/kfrico/BitfinexLendingBot/internal/bitfinex"
)

// MarketAnalyzer 市場分析器
type MarketAnalyzer struct {
	rateHistory    []RateSnapshot
	maxHistorySize int
}

// RateSnapshot 利率快照
type RateSnapshot struct {
	Rate      float64
	Timestamp time.Time
	Volume    float64
}

// MarketCondition 市場狀況
type MarketCondition struct {
	Trend          string  // "rising", "falling", "stable"
	Volatility     float64 // 波動率
	LiquidityDepth int     // 流動性深度
	AvgRate        float64 // 平均利率
	RateRatio      float64 // 當前利率/平均利率
}

// NewMarketAnalyzer 創建市場分析器
func NewMarketAnalyzer() *MarketAnalyzer {
	return &MarketAnalyzer{
		rateHistory:    make([]RateSnapshot, 0),
		maxHistorySize: 48, // 保留48個數據點 (12小時，每15分鐘一次)
	}
}

// AddRateSnapshot 添加利率快照
func (ma *MarketAnalyzer) AddRateSnapshot(rate float64, volume float64) {
	snapshot := RateSnapshot{
		Rate:      rate,
		Timestamp: time.Now(),
		Volume:    volume,
	}

	ma.rateHistory = append(ma.rateHistory, snapshot)

	// 保持歷史數據大小限制
	if len(ma.rateHistory) > ma.maxHistorySize {
		ma.rateHistory = ma.rateHistory[1:]
	}
}

// AnalyzeMarket 分析市場狀況
func (ma *MarketAnalyzer) AnalyzeMarket(fundingBook []*bitfinex.FundingBookEntry) *MarketCondition {
	if len(ma.rateHistory) < 3 {
		// 數據不足，返回默認狀況
		return &MarketCondition{
			Trend:          "stable",
			Volatility:     0.0,
			LiquidityDepth: len(fundingBook),
			AvgRate:        0.0,
			RateRatio:      1.0,
		}
	}

	avgRate := ma.calculateAverageRate()
	volatility := ma.calculateVolatility()
	trend := ma.determineTrend()
	currentRate := ma.rateHistory[len(ma.rateHistory)-1].Rate
	rateRatio := 1.0
	if avgRate > 0 {
		rateRatio = currentRate / avgRate
	}

	return &MarketCondition{
		Trend:          trend,
		Volatility:     volatility,
		LiquidityDepth: len(fundingBook),
		AvgRate:        avgRate,
		RateRatio:      rateRatio,
	}
}

// calculateAverageRate 計算平均利率
func (ma *MarketAnalyzer) calculateAverageRate() float64 {
	if len(ma.rateHistory) == 0 {
		return 0.0
	}

	var sum float64
	for _, snapshot := range ma.rateHistory {
		sum += snapshot.Rate
	}

	return sum / float64(len(ma.rateHistory))
}

// calculateVolatility 計算波動率
func (ma *MarketAnalyzer) calculateVolatility() float64 {
	if len(ma.rateHistory) < 2 {
		return 0.0
	}

	avgRate := ma.calculateAverageRate()
	var sumSquaredDiff float64

	for _, snapshot := range ma.rateHistory {
		diff := snapshot.Rate - avgRate
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(ma.rateHistory))
	return math.Sqrt(variance)
}

// determineTrend 判斷趨勢
func (ma *MarketAnalyzer) determineTrend() string {
	if len(ma.rateHistory) < 6 {
		return "stable"
	}

	// 取最近6個點進行趨勢分析
	recentHistory := ma.rateHistory[len(ma.rateHistory)-6:]

	var upCount, downCount int
	for i := 1; i < len(recentHistory); i++ {
		diff := recentHistory[i].Rate - recentHistory[i-1].Rate
		threshold := 0.0001 // 0.01%的變化閾值

		if diff > threshold {
			upCount++
		} else if diff < -threshold {
			downCount++
		}
	}

	if upCount >= 4 {
		return "rising"
	} else if downCount >= 4 {
		return "falling"
	}

	return "stable"
}

// AnalyzeCompetition 分析競爭對手
func (ma *MarketAnalyzer) AnalyzeCompetition(fundingBook []*bitfinex.FundingBookEntry) float64 {
	if len(fundingBook) < 10 {
		return 0.0
	}

	// 分析前10層的平均利率差
	var totalSpread float64
	validSpreads := 0

	for i := 0; i < 9 && i < len(fundingBook)-1; i++ {
		spread := fundingBook[i+1].Rate - fundingBook[i].Rate
		if spread > 0 {
			totalSpread += spread
			validSpreads++
		}
	}

	if validSpreads == 0 {
		return 0.0
	}

	avgSpread := totalSpread / float64(validSpreads)

	// 建議利率：略高於當前最佳利率
	return fundingBook[0].Rate + avgSpread*0.3
}

// GetOptimalDepthRange 獲取最佳深度範圍
func (ma *MarketAnalyzer) GetOptimalDepthRange(fundsAvailable float64, condition *MarketCondition) (bottom, top float64) {
	// 基礎範圍根據資金量調整
	baseBottom := 10.0
	baseTop := 1000.0

	if fundsAvailable > 1000 {
		baseBottom = 5.0
		baseTop = 3000.0
	} else if fundsAvailable > 500 {
		baseBottom = 8.0
		baseTop = 2000.0
	}

	// 根據市場狀況調整
	switch condition.Trend {
	case "rising":
		// 利率上升時縮小範圍，提升競爭力
		baseBottom *= 1.2
		baseTop *= 0.8
	case "falling":
		// 利率下降時擴大範圍，分散風險
		baseBottom *= 0.8
		baseTop *= 1.2
	}

	// 根據波動率調整 (使用配置的波動率閾值)
	if condition.Volatility > 0.001 { // 高波動
		baseTop *= 1.3 // 擴大範圍應對波動
	}

	return baseBottom, baseTop
}
