package dto

type GlobalMacroIndicator struct {
	Code          string         `json:"code"`
	Name          string         `json:"name"`
	Value         float64        `json:"value"`
	Unit          string         `json:"unit"`
	Period        string         `json:"period"`
	Trend         string         `json:"trend"`
	Percentile10y *float64       `json:"percentile_10y"`
	AsOf          string         `json:"as_of"`
	Meta          DataSourceMeta `json:"meta"`
}

type CycleCompassState struct {
	CurrentPhase     string                 `json:"current_phase"`
	PhaseCode        int                    `json:"phase_code"`
	DominantAsset    string                 `json:"dominant_asset"`
	MacroIndicators  []GlobalMacroIndicator `json:"macro_indicators"`
	EquityBondSpread float64                `json:"equity_bond_spread"`
	StylePreference  string                 `json:"style_preference"`
	Summary          string                 `json:"summary"`
}
