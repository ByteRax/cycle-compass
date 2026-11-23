package dto

type PositionItem struct {
	Symbol              string  `json:"symbol"`
	Name                string  `json:"name"`
	Shares              float64 `json:"shares"`
	CurrentPrice        float64 `json:"current_price"`
	MarketValue         float64 `json:"market_value"`
	CostPrice           float64 `json:"cost_price"`
	TotalCost           float64 `json:"total_cost"`
	UnrealizedProfit    float64 `json:"unrealized_profit"`
	UnrealizedProfitPct float64 `json:"unrealized_profit_pct"`
}

type LedgerEntry struct {
	Date             string   `json:"date"`
	EventType        string   `json:"event_type"`
	Symbol           *string  `json:"symbol"`
	TargetName       string   `json:"target_name"`
	Shares           float64  `json:"shares"`
	Price            float64  `json:"price"`
	Amount           float64  `json:"amount"`
	CashInflow       float64  `json:"cash_inflow"`
	CashOutflow      float64  `json:"cash_outflow"`
	CashPoolDelta    float64  `json:"cash_pool_delta"`
	CashBalanceAfter float64  `json:"cash_balance_after"`
	CashPoolAfter    float64  `json:"cash_pool_after"`
	Description      string   `json:"description"`
}

type PortfolioOverview struct {
	AccountID          string         `json:"account_id"`
	TotalAsset         float64        `json:"total_asset"`
	AvailableCash      float64        `json:"available_cash"`
	CashPool           float64        `json:"cash_pool"`
	PositionsValue     float64        `json:"positions_value"`
	TotalCost          float64        `json:"total_cost"`
	TotalProfit        float64        `json:"total_profit"`
	TotalProfitPct     float64        `json:"total_profit_pct"`
	Positions          []PositionItem `json:"positions"`
	RecentTransactions []LedgerEntry  `json:"recent_transactions"`
}
