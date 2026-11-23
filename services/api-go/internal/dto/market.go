package dto

type DataSourceMeta struct {
	SourceName string   `json:"source_name"`
	Symbol     string   `json:"symbol"`
	AsOf       string   `json:"as_of"`
	Status     string   `json:"status"`
	LatencyMs  *float64 `json:"latency_ms"`
	Warning    *string  `json:"warning"`
	UpdatedAt  string   `json:"updated_at"`
}

type MarketQuote struct {
	Symbol    string         `json:"symbol"`
	Name      string         `json:"name"`
	Category  string         `json:"category"`
	Price     float64        `json:"price"`
	Change    float64        `json:"change"`
	ChangePct float64        `json:"change_pct"`
	Open      *float64       `json:"open"`
	High      *float64       `json:"high"`
	Low       *float64       `json:"low"`
	Volume    *float64       `json:"volume"`
	Turnover  *float64       `json:"turnover"`
	Currency  string         `json:"currency"`
	AsOf      string         `json:"as_of"`
	Meta      DataSourceMeta `json:"meta"`
}
