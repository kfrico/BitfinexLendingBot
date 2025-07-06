package bitfinex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/common"
	"github.com/bitfinexcom/bitfinex-api-go/pkg/models/fundingoffer"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"

	"github.com/kfrico/BitfinexLendingBot/internal/constants"
	"github.com/kfrico/BitfinexLendingBot/internal/errors"
)

// Client Bitfinex API 客戶端封裝
type Client struct {
	restClient *rest.Client
}

// NewClient 創建新的 Bitfinex 客戶端
func NewClient(apiKey, secretKey string) *Client {
	client := rest.NewClient().Credentials(apiKey, secretKey)
	return &Client{
		restClient: client,
	}
}

// FundingOffer 代表一個資金貸出訂單
type FundingOffer struct {
	ID     int64
	Amount float64
	Rate   float64 // 日利率（小數格式）
	Period int
}

// Wallet 代表錢包信息
type Wallet struct {
	Currency  string
	Type      string
	Balance   float64
	Available float64
}

// FundingBookEntry 代表資金訂單簿條目
type FundingBookEntry struct {
	Rate   float64 // 日利率（小數格式）
	Amount float64
	Period int
	Count  int
}

// FundingCredit 代表活躍的借貸訂單
type FundingCredit struct {
	ID         int64
	Symbol     string
	Amount     float64
	Rate       float64 // 日利率（小數格式）
	Period     int64   // 期間（天）
	MTSCreated int64   // 創建時間戳（毫秒）
	MTSOpened  int64   // 開始時間戳（毫秒）
	Status     string  // 狀態
}

// Candle 代表 K 線數據
type Candle struct {
	MTS    int64   // 時間戳（毫秒）
	Open   float64 // 開盤價
	Close  float64 // 收盤價
	High   float64 // 最高價
	Low    float64 // 最低價
	Volume float64 // 成交量
}

// GetFundingOffers 獲取未完成的資金貸出訂單
func (c *Client) GetFundingOffers(symbol string) ([]*FundingOffer, error) {
	offers, err := c.restClient.Funding.Offers(symbol)
	if err != nil {
		// 處理特殊的空響應錯誤
		if strings.Contains(err.Error(), "data slice too short for funding offer") {
			return []*FundingOffer{}, nil
		}
		return nil, errors.NewAPIError("failed to get funding offers", err)
	}

	// 處理空響應或無數據的情況
	if offers == nil || offers.Snapshot == nil || len(offers.Snapshot) == 0 {
		return []*FundingOffer{}, nil
	}

	result := make([]*FundingOffer, 0, len(offers.Snapshot))
	for _, offer := range offers.Snapshot {
		// 添加安全檢查，防止空數據導致panic
		if offer == nil {
			continue
		}
		result = append(result, &FundingOffer{
			ID:     offer.ID,
			Amount: offer.Amount,
			Rate:   offer.Rate, // API 已返回日利率
			Period: int(offer.Period),
		})
	}

	return result, nil
}

// CancelFundingOffer 取消資金貸出訂單
func (c *Client) CancelFundingOffer(offerID int64) error {
	cancelReq := &fundingoffer.CancelRequest{
		ID: offerID,
	}

	_, err := c.restClient.Funding.CancelOffer(cancelReq)
	if err != nil {
		return errors.NewOrderError("failed to cancel funding offer", err)
	}

	return nil
}

// SubmitFundingOffer 提交新的資金貸出訂單
func (c *Client) SubmitFundingOffer(symbol string, amount float64, dailyRate float64, period int, hidden bool) error {
	offerReq := &fundingoffer.SubmitRequest{
		Type:   constants.OfferTypeLIMIT,
		Symbol: symbol,
		Amount: amount,
		Rate:   dailyRate, // v2 API 使用日利率
		Period: int64(period),
		Hidden: hidden,
	}

	_, err := c.restClient.Funding.SubmitOffer(offerReq)
	if err != nil {
		return errors.NewOrderError("failed to submit funding offer", err)
	}

	return nil
}

// GetWallets 獲取錢包信息
func (c *Client) GetWallets() ([]*Wallet, error) {
	wallets, err := c.restClient.Wallet.Wallet()
	if err != nil {
		return nil, errors.NewAPIError("failed to get wallets", err)
	}

	result := make([]*Wallet, 0, len(wallets.Snapshot))
	for _, w := range wallets.Snapshot {
		result = append(result, &Wallet{
			Currency:  w.Currency,
			Type:      w.Type,
			Balance:   w.Balance,
			Available: w.BalanceAvailable,
		})
	}

	return result, nil
}

// GetFundingBalance 獲取指定幣種的資金錢包餘額
func (c *Client) GetFundingBalance(currency string) (float64, error) {
	wallets, err := c.GetWallets()
	if err != nil {
		return 0, err
	}

	for _, wallet := range wallets {
		if wallet.Currency == currency && wallet.Type == constants.WalletTypeFunding {
			return wallet.Available, nil
		}
	}

	return 0, nil
}

// GetFundingBook 獲取資金訂單簿
func (c *Client) GetFundingBook(symbol string, limit int) ([]*FundingBookEntry, error) {
	if limit > constants.MaxPriceLevels {
		limit = constants.MaxPriceLevels
	}
	if limit <= 0 {
		limit = constants.DefaultPriceLevels
	}

	book, err := c.restClient.Book.All(symbol, common.PrecisionRawBook, limit)
	if err != nil {
		return nil, errors.NewAPIError("failed to get funding book", err)
	}

	if len(book.Snapshot) == 0 {
		return []*FundingBookEntry{}, nil
	}

	result := make([]*FundingBookEntry, 0, len(book.Snapshot))
	for _, entry := range book.Snapshot {
		result = append(result, &FundingBookEntry{
			Rate:   entry.Rate, // API 已返回日利率
			Amount: entry.Amount,
			Period: int(entry.Period),
			Count:  int(entry.Count),
		})
	}

	return result, nil
}

// GetCurrentFundingRate 獲取當前資金利率
func (c *Client) GetCurrentFundingRate(symbol string) (float64, error) {
	book, err := c.GetFundingBook(symbol, 1)
	if err != nil {
		return 0, err
	}

	if len(book) == 0 {
		return 0, errors.NewAPIError("no funding book data available", nil)
	}

	return book[0].Rate, nil
}

// GetFundingCredits 獲取活躍的借貸訂單
func (c *Client) GetFundingCredits(symbol string) ([]*FundingCredit, error) {
	credits, err := c.restClient.Funding.Credits(symbol)
	if err != nil {
		// 處理特殊的空響應錯誤
		if strings.Contains(err.Error(), "data slice too short") {
			return []*FundingCredit{}, nil
		}
		return nil, errors.NewAPIError("failed to get funding credits", err)
	}

	// 處理空響應或無數據的情況
	if credits == nil || credits.Snapshot == nil || len(credits.Snapshot) == 0 {
		return []*FundingCredit{}, nil
	}

	result := make([]*FundingCredit, 0, len(credits.Snapshot))
	for _, credit := range credits.Snapshot {
		// 添加安全檢查，防止空數據導致panic
		if credit == nil {
			continue
		}
		result = append(result, &FundingCredit{
			ID:         credit.ID,
			Symbol:     credit.Symbol,
			Amount:     credit.Amount,
			Rate:       credit.Rate, // API 已返回日利率
			Period:     credit.Period,
			MTSCreated: credit.MTSCreated,
			MTSOpened:  credit.MTSOpened,
			Status:     credit.Status,
		})
	}

	return result, nil
}

// GetFundingCandles 獲取資金 K 線數據
func (c *Client) GetFundingCandles(symbol string, timeFrame string, limit int) ([]*Candle, error) {
	// 構建 candle key，格式: trade:15m:fUSD:a30:p2:p30
	candleKey := fmt.Sprintf("trade:%s:%s:a30:p2:p30", timeFrame, symbol)

	// 構建 API URL
	url := fmt.Sprintf("https://api-pub.bitfinex.com/v2/candles/%s/hist?limit=%d", candleKey, limit)

	// 發送 HTTP 請求
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.NewAPIError("failed to get funding candles", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewAPIError(fmt.Sprintf("API returned status code %d", resp.StatusCode), nil)
	}

	// 解析響應
	var rawData [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, errors.NewAPIError("failed to decode candles response", err)
	}

	// 轉換為 Candle 結構
	candles := make([]*Candle, 0, len(rawData))
	for _, raw := range rawData {
		if len(raw) != 6 {
			continue // 跳過無效數據
		}

		// 安全地轉換每個字段
		mts, ok := raw[0].(float64)
		if !ok {
			continue
		}

		open, ok := raw[1].(float64)
		if !ok {
			continue
		}

		close, ok := raw[2].(float64)
		if !ok {
			continue
		}

		high, ok := raw[3].(float64)
		if !ok {
			continue
		}

		low, ok := raw[4].(float64)
		if !ok {
			continue
		}

		volume, ok := raw[5].(float64)
		if !ok {
			continue
		}

		candle := &Candle{
			MTS:    int64(mts),
			Open:   open,
			Close:  close,
			High:   high,
			Low:    low,
			Volume: volume,
		}
		candles = append(candles, candle)
	}

	return candles, nil
}
