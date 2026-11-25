package dto

type BacktestRequest struct {
	LowLine      float64 `json:"low_line"`
	HighLine     float64 `json:"high_line"`
	WeeklyAmount float64 `json:"weekly_amount"`
	Days         int     `json:"days"`
}

type DailyRecord struct {
	Date           string  `json:"date"`
	Ratio          float64 `json:"ratio"`
	Zone           string  `json:"zone"`
	GrowthPrice    float64 `json:"growth_price"`
	DividendPrice  float64 `json:"dividend_price"`
	GrowthShares   float64 `json:"growth_shares"`
	DividendShares float64 `json:"dividend_shares"`
	MarketValue    float64 `json:"market_value"`
	CashPool       float64 `json:"cash_pool"`
	TotalAsset     float64 `json:"total_asset"`
	TotalInvest    float64 `json:"total_invest"`
	NetProfit      float64 `json:"net_profit"`
	ReturnPct      float64 `json:"return_pct"`
}

type BacktestSummary struct {
	StartDate          string         `json:"start_date"`
	EndDate            string         `json:"end_date"`
	TradingDays        int            `json:"trading_days"`
	TotalInvest        float64        `json:"total_invest"`
	FinalAsset         float64        `json:"final_asset"`
	NetProfit          float64        `json:"net_profit"`
	TotalReturnPct     float64        `json:"total_return_pct"`
	AnnualizedReturn   float64        `json:"annualized_return"`
	XIRR               float64        `json:"xirr"`
	MaxDrawdown        float64        `json:"max_drawdown"`
	SharpeRatio        float64        `json:"sharpe_ratio"`
	CalmarRatio        float64        `json:"calmar_ratio"`
	ZoneDistribution   map[string]int `json:"zone_distribution"`
	TotalRebuyCount    int            `json:"total_rebuy_count"`
	LedgerEntriesCount int            `json:"ledger_entries_count"`
	RuleVersion        string         `json:"rule_version,omitempty"`
	DataSource         string         `json:"data_source,omitempty"`
	ExecutionTiming    string         `json:"execution_timing,omitempty"`
	ClipReason         string         `json:"clip_reason,omitempty"`
}

type BacktestOutput struct {
	Summary       BacktestSummary `json:"summary"`
	DailyRecords  []DailyRecord   `json:"daily_records"`
	LedgerEntries []LedgerEntry   `json:"ledger_entries"`
}
