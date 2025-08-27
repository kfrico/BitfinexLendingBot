package tracker

import (
	"sync"
	"time"
)

// BotOrderTracker 追蹤程式創建的訂單
type BotOrderTracker struct {
	mu           sync.RWMutex
	createdOrders map[int64]time.Time // orderID -> 創建時間
	botStartTime  time.Time
}

// NewBotOrderTracker 創建新的訂單追蹤器
func NewBotOrderTracker() *BotOrderTracker {
	return &BotOrderTracker{
		createdOrders: make(map[int64]time.Time),
		botStartTime:  time.Now(),
	}
}

// TrackOrder 記錄程式創建的訂單
func (t *BotOrderTracker) TrackOrder(orderID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.createdOrders[orderID] = time.Now()
}

// IsTrackedOrder 檢查是否為程式創建的訂單
func (t *BotOrderTracker) IsTrackedOrder(orderID int64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, exists := t.createdOrders[orderID]
	return exists
}

// RemoveOrder 移除已追蹤的訂單（訂單完成或取消時）
func (t *BotOrderTracker) RemoveOrder(orderID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.createdOrders, orderID)
}

// GetTrackedOrders 獲取所有追蹤的訂單ID
func (t *BotOrderTracker) GetTrackedOrders() []int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	orders := make([]int64, 0, len(t.createdOrders))
	for orderID := range t.createdOrders {
		orders = append(orders, orderID)
	}
	return orders
}

// CleanOldOrders 清理舊訂單記錄（避免記憶體洩漏）
func (t *BotOrderTracker) CleanOldOrders(maxAge time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	now := time.Now()
	for orderID, createdTime := range t.createdOrders {
		if now.Sub(createdTime) > maxAge {
			delete(t.createdOrders, orderID)
		}
	}
}

// GetOrderCount 獲取追蹤的訂單數量
func (t *BotOrderTracker) GetOrderCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.createdOrders)
}

// GetBotStartTime 獲取機器人啟動時間
func (t *BotOrderTracker) GetBotStartTime() time.Time {
	return t.botStartTime
}