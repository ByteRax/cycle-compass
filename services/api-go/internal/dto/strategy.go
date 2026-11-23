package dto

type RotationSignal struct {
	Date        string  `json:"date"`
	Ratio       float64 `json:"ratio"`
	Zone        string  `json:"zone"`
	QuantilePct float64 `json:"quantile_pct"`
	Action      string  `json:"action"`
	ToHighPct   float64 `json:"to_high_pct"`
	ToLowPct    float64 `json:"to_low_pct"`
}

type ScoreResult struct {
	Tech            float64 `json:"tech"`
	Spread          float64 `json:"spread"`
	Value           float64 `json:"value"`
	Style           float64 `json:"style"`
	SpreadPP        float64 `json:"spread_pp"`
	Total           float64 `json:"total"`
	Action          string  `json:"action"`
	Multiplier      float64 `json:"multiplier"`
	Amount          float64 `json:"amount"`
	ReduceRatio     float64 `json:"reduce_ratio"`
	Close           float64 `json:"close,omitempty"`
	PE              float64 `json:"pe,omitempty"`
	PEPercentile    float64 `json:"pe_percentile,omitempty"`
	DividendYield   float64 `json:"dividend_yield,omitempty"`
	Bond10Y         float64 `json:"bond10y,omitempty"`
	RSI70           float64 `json:"rsi70,omitempty"`
	BPct100         float64 `json:"b_pct100,omitempty"`
	StylePercentile float64 `json:"style_percentile,omitempty"`
	DataSourceNote  string  `json:"data_source_note,omitempty"`
}

type CombinedStrategySignal struct {
	Date           string         `json:"date"`
	Rotation       RotationSignal `json:"rotation"`
	Scoring        ScoreResult    `json:"scoring"`
	CombinedAction string         `json:"combined_action"`
	SummaryTips    string         `json:"summary_tips"`
}
