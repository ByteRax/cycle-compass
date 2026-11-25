package trend

// TempLevel represents the seven-tier absolute trend temperature level plus unavailable.
type TempLevel string

const (
	TempBoiling     TempLevel = "boiling"
	TempHot         TempLevel = "hot"
	TempWarm        TempLevel = "warm"
	TempFlat        TempLevel = "flat"
	TempCool        TempLevel = "cool"
	TempCold        TempLevel = "cold"
	TempFreezing    TempLevel = "freezing"
	TempUnavailable TempLevel = "unavailable"
)

// Direction represents the trend direction in the state machine.
type Direction string

const (
	DirectionUp      Direction = "up"
	DirectionDown    Direction = "down"
	DirectionNeutral Direction = "neutral"
	DirectionUnknown Direction = "unknown"
)

// Confirmation represents the confirmation status of a trend direction.
type Confirmation string

const (
	ConfirmationForming   Confirmation = "forming"
	ConfirmationConfirmed Confirmation = "confirmed"
	ConfirmationNone      Confirmation = "none"
	ConfirmationUnknown   Confirmation = "unknown"
)

// Side represents the product-defined trend side (right = confirmed up, left = confirmed down).
type Side string

const (
	SideRight   Side = "right"
	SideLeft    Side = "left"
	SideNeutral Side = "neutral"
	SideUnknown Side = "unknown"
)

// Tier represents the five-tier relative strength echelon.
type Tier string

const (
	TierS Tier = "S"
	TierA Tier = "A"
	TierB Tier = "B"
	TierC Tier = "C"
	TierD Tier = "D"
)

// DailyBar represents a single daily observation of a positive wealth/price index.
type DailyBar struct {
	AssetID     string   `json:"asset_id,omitempty"`
	TradingDate string   `json:"trading_date"` // YYYY-MM-DD
	Value       *float64 `json:"value"`        // positive wealth/price index W; nil if missing/closed
	Volume      *float64 `json:"volume,omitempty"`
	Turnover    *float64 `json:"turnover,omitempty"`
}

// AlgorithmConfig holds versioned parameters for the trend-v1.0.0 quantitative model.
type AlgorithmConfig struct {
	AlgorithmVersion  string  `json:"algorithm_version"`
	TempWindow        int     `json:"temp_window"`
	MinCoverage       float64 `json:"min_coverage"`
	SigmaWindow       int     `json:"sigma_window"`
	Epsilon           float64 `json:"epsilon"`
	EMAFast           int     `json:"ema_fast"`
	EMASlow           int     `json:"ema_slow"`
	SlopeLag          int     `json:"slope_lag"`
	ConfirmDays       int     `json:"confirm_days"`
	ExitDays          int     `json:"exit_days"`
	UpEntryQ          float64 `json:"up_entry_q"`
	UpExitQ           float64 `json:"up_exit_q"`
	DownEntryQ        float64 `json:"down_entry_q"`
	DownExitQ         float64 `json:"down_exit_q"`
	StrengthMinPoints int     `json:"strength_min_points"`
}

// DefaultAlgorithmConfig returns the default trend-v1.0.0 configuration.
func DefaultAlgorithmConfig() AlgorithmConfig {
	return AlgorithmConfig{
		AlgorithmVersion:  "trend-v1.0.0",
		TempWindow:        121,
		MinCoverage:       0.90,
		SigmaWindow:       60,
		Epsilon:           1e-6,
		EMAFast:           20,
		EMASlow:           60,
		SlopeLag:          5,
		ConfirmDays:       3,
		ExitDays:          2,
		UpEntryQ:          0.35,
		UpExitQ:           0.10,
		DownEntryQ:        -0.35,
		DownExitQ:         -0.10,
		StrengthMinPoints: 61,
	}
}

// TempResult holds the computed absolute trend temperature and intermediate factors.
type TempResult struct {
	TradingDate  string    `json:"trading_date"`
	Score        *float64  `json:"score"`
	DisplayScore *int      `json:"display_score"`
	Level        TempLevel `json:"level"`
	Q            *float64  `json:"q,omitempty"`
	Z20          *float64  `json:"z20,omitempty"`
	Z60          *float64  `json:"z60,omitempty"`
	Z120         *float64  `json:"z120,omitempty"`
	Sigma60      *float64  `json:"sigma60,omitempty"`
	EMA20        *float64  `json:"ema20,omitempty"`
	EMA60        *float64  `json:"ema60,omitempty"`
	Slope20      *float64  `json:"slope20,omitempty"`
	Slope60      *float64  `json:"slope60,omitempty"`
	ValidCount   int       `json:"valid_count"`
	Coverage     float64   `json:"coverage"`
	Reason       string    `json:"reason,omitempty"`
}

// StateResult holds a daily snapshot of the direction/confirmation/duration state machine.
type StateResult struct {
	TradingDate          string       `json:"trading_date"`
	Direction            Direction    `json:"direction"`
	Confirmation         Confirmation `json:"confirmation"`
	Side                 Side         `json:"side"`
	StartedAt            string       `json:"started_at,omitempty"`
	ConfirmedAt          string       `json:"confirmed_at,omitempty"`
	DurationTradingDays  int          `json:"duration_trading_days"`
	DurationCalendarDays int          `json:"duration_calendar_days"`
	FormingDays          int          `json:"forming_days"`
	ExitStreak           int          `json:"exit_streak"`
	StateVersion         string       `json:"state_version"`
	Stale                bool         `json:"stale"`
	Reason               string       `json:"reason,omitempty"`
}

// TrendEvent represents an idempotent state transition event.
type TrendEvent struct {
	EventID          string `json:"event_id"`
	AssetID          string `json:"asset_id"`
	EventType        string `json:"event_type"`
	OldValue         string `json:"old_value"`
	NewValue         string `json:"new_value"`
	AsOf             string `json:"as_of"`
	DetectedAt       string `json:"detected_at"`
	SnapshotID       string `json:"snapshot_id,omitempty"`
	AlgorithmVersion string `json:"algorithm_version"`
	ScopeID          string `json:"scope_id,omitempty"`
	DedupeKey        string `json:"dedupe_key"`
}

// StrengthInput provides the series for one candidate asset in a relative-strength pool.
type StrengthInput struct {
	AssetID string     `json:"asset_id"`
	Bars    []DailyBar `json:"bars"`
}

// StrengthResult holds the cross-sectional relative strength ranking for one asset.
type StrengthResult struct {
	AssetID          string   `json:"asset_id"`
	RawFactor        *float64 `json:"raw_factor"`
	Rank             *float64 `json:"rank"`
	Score            *float64 `json:"score"`
	DisplayScore     *float64 `json:"display_score"`
	CandidateCount   int      `json:"candidate_count"`
	ValidCount       int      `json:"valid_count"`
	ScopeID          string   `json:"scope_id"`
	ScopeLabel       string   `json:"scope_label"`
	Currency         string   `json:"currency"`
	AlgorithmVersion string   `json:"algorithm_version"`
	SourceDate       string   `json:"source_date,omitempty"`
	Reason           string   `json:"reason,omitempty"`
}

// PhaseResult holds the trend solar-term lifecycle mapping.
type PhaseResult struct {
	Direction           Direction    `json:"direction"`
	Confirmation        Confirmation `json:"confirmation"`
	DurationTradingDays int          `json:"duration_trading_days"`
	PhaseCode           string       `json:"phase_code"`
	PhaseName           string       `json:"phase_name"`
	Description         string       `json:"description"`
}

// TierResult holds the five-tier echelon classification and labels.
type TierResult struct {
	Tier   *Tier  `json:"tier"`
	Label  string `json:"label,omitempty"`
	Alias  string `json:"alias,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// ExtremumResult holds the O(n) maximum drawdown (peak before trough) and maximum gain (trough before peak).
type ExtremumResult struct {
	MaxDrawdown         float64 `json:"max_drawdown"`
	DrawdownPeakDate    string  `json:"drawdown_peak_date"`
	DrawdownPeakValue   float64 `json:"drawdown_peak_value"`
	DrawdownTroughDate  string  `json:"drawdown_trough_date"`
	DrawdownTroughValue float64 `json:"drawdown_trough_value"`
	MaxGain             float64 `json:"max_gain"`
	GainTroughDate      string  `json:"gain_trough_date"`
	GainTroughValue     float64 `json:"gain_trough_value"`
	GainPeakDate        string  `json:"gain_peak_date"`
	GainPeakValue       float64 `json:"gain_peak_value"`
	Valid               bool    `json:"valid"`
	Reason              string  `json:"reason,omitempty"`
}

// AnnualWeekResult holds the return observation for one natural-year week bucket (1..53).
type AnnualWeekResult struct {
	Year        int     `json:"year"`
	Week        int     `json:"week"`
	WeekEndDate string  `json:"week_end_date"`
	ActualDate  string  `json:"actual_date"`
	Close       float64 `json:"close"`
	Return      float64 `json:"return"`
}

// AnnualYearSummary holds the annual comparison summary for a single calendar year.
type AnnualYearSummary struct {
	Year           int                `json:"year"`
	BaseDate       string             `json:"base_date,omitempty"`
	BaseClose      float64            `json:"base_close,omitempty"`
	SameWeekReturn *float64           `json:"same_week_return"`
	SameWeekDate   string             `json:"same_week_date,omitempty"`
	FullYearReturn *float64           `json:"full_year_return"`
	FullYearDate   string             `json:"full_year_date,omitempty"`
	IsComplete     bool               `json:"is_complete"`
	Weeks          []AnnualWeekResult `json:"weeks,omitempty"`
}

// AnnualComparisonResult holds multi-year same-week and full-year comparison statistics.
type AnnualComparisonResult struct {
	AsOf               string              `json:"as_of"`
	CurrentYear        int                 `json:"current_year"`
	CurrentWeek        int                 `json:"current_week"`
	CurrentReturn      *float64            `json:"current_return"`
	Years              []AnnualYearSummary `json:"years"`
	CompleteYearCount  int                 `json:"complete_year_count"`
	SameWeekPercentile *float64            `json:"same_week_percentile"`
	SameWeekMedian     *float64            `json:"same_week_median"`
	SameWeekMean       *float64            `json:"same_week_mean"`
	SameWeekUpRatio    *float64            `json:"same_week_up_ratio"`
	FullYearMedian     *float64            `json:"full_year_median"`
	FullYearMean       *float64            `json:"full_year_mean"`
	FullYearUpRatio    *float64            `json:"full_year_up_ratio"`
	Reason             string              `json:"reason,omitempty"`
}
