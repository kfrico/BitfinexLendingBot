package rates

import (
	"github.com/kfrico/BitfinexLendingBot/internal/constants"
)

// Converter 利率轉換器
type Converter struct{}

// NewConverter 創建新的利率轉換器
func NewConverter() *Converter {
	return &Converter{}
}

// PercentageToDecimal 將百分比轉換為小數
func (c *Converter) PercentageToDecimal(percentage float64) float64 {
	return percentage / constants.PercentageToDecimal
}

// DecimalToPercentage 將小數轉換為百分比
func (c *Converter) DecimalToPercentage(decimal float64) float64 {
	return decimal * constants.PercentageToDecimal
}

// DailyToAnnual 將日利率轉換為年利率
func (c *Converter) DailyToAnnual(dailyRate float64) float64 {
	return dailyRate * constants.DaysPerYear
}

// AnnualToDaily 將年利率轉換為日利率
func (c *Converter) AnnualToDaily(annualRate float64) float64 {
	return annualRate / constants.DaysPerYear
}

// PercentageDailyToDecimalDaily 將百分比日利率轉換為小數日利率
func (c *Converter) PercentageDailyToDecimalDaily(percentageDaily float64) float64 {
	return c.PercentageToDecimal(percentageDaily)
}

// DecimalDailyToPercentageDaily 將小數日利率轉換為百分比日利率
func (c *Converter) DecimalDailyToPercentageDaily(decimalDaily float64) float64 {
	return c.DecimalToPercentage(decimalDaily)
}

// PercentageToAnnualDecimal 將百分比轉換為年化小數
// 例如：0.5% -> 0.005 -> 1.825 (年化)
func (c *Converter) PercentageToAnnualDecimal(percentage float64) float64 {
	dailyDecimal := c.PercentageToDecimal(percentage)
	return c.DailyToAnnual(dailyDecimal)
}

// AnnualDecimalToPercentage 將年化小數轉換為百分比
func (c *Converter) AnnualDecimalToPercentage(annualDecimal float64) float64 {
	dailyDecimal := c.AnnualToDaily(annualDecimal)
	return c.DecimalToPercentage(dailyDecimal)
}

// ValidateDailyRate 驗證日利率是否在合理範圍內
func (c *Converter) ValidateDailyRate(dailyRate float64) bool {
	// Bitfinex 限制每日利率不超過 7%
	const maxDailyRateDecimal = 0.07
	return dailyRate > 0 && dailyRate <= maxDailyRateDecimal
}

// ValidatePercentageRate 驗證百分比利率是否在合理範圍內
func (c *Converter) ValidatePercentageRate(percentageRate float64) bool {
	// 轉換為日利率進行驗證
	dailyRate := c.PercentageToDecimal(percentageRate)
	return c.ValidateDailyRate(dailyRate)
}
