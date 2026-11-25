package dto

// SourceDateMeta 描述单个市场的数据截止日期与来源状态。
type SourceDateMeta struct {
	Market string `json:"market"`
	AsOf   string `json:"as_of"`
	Source string `json:"source"`
	Status string `json:"status,omitempty"`
}

// SnapshotMeta 描述趋势计算快照的元数据与覆盖率。
type SnapshotMeta struct {
	SnapshotID          string           `json:"snapshot_id"`
	AsOf                string           `json:"as_of"`
	ComputedAt          string           `json:"computed_at"`
	AlgorithmVersion    string           `json:"algorithm_version"`
	Currency            string           `json:"currency"`
	RequestedCurrency   string           `json:"requested_currency,omitempty"`
	EffectiveCurrency   string           `json:"effective_currency,omitempty"`
	ScopeID             string           `json:"scope_id"`
	ScopeLabel          string           `json:"scope_label"`
	CandidateCount      int              `json:"candidate_count"`
	ValidCount          int              `json:"valid_count"`
	Coverage            *float64         `json:"coverage"`
	Mode                string           `json:"mode"`
	Status              string           `json:"status"`
	IsCurrent           bool             `json:"is_current"`
	IsExpired           bool             `json:"is_expired,omitempty"`
	RequestedSnapshotID string           `json:"requested_snapshot_id,omitempty"`
	SourceDates         []SourceDateMeta `json:"source_dates"`
	Warnings            []string         `json:"warnings"`
}

// StrengthInfo 描述相对强度池内排名与得分。
type StrengthInfo struct {
	ScopeID        string   `json:"scope_id"`
	ScopeLabel     string   `json:"scope_label"`
	Score          *float64 `json:"score"`
	Rank           *int     `json:"rank"`
	RawFactor      *float64 `json:"raw_factor"`
	ValidCount     int      `json:"valid_count"`
	CandidateCount int      `json:"candidate_count"`
	Delta          *float64 `json:"delta"`
	Reason         string   `json:"reason,omitempty"`
}

// TrendEventItem 描述幂等状态迁移事件。
type TrendEventItem struct {
	EventID          string `json:"event_id"`
	AssetID          string `json:"asset_id"`
	EventType        string `json:"event_type"`
	OldValue         string `json:"old_value"`
	NewValue         string `json:"new_value"`
	AsOf             string `json:"as_of"`
	DetectedAt       string `json:"detected_at"`
	SnapshotID       string `json:"snapshot_id"`
	AlgorithmVersion string `json:"algorithm_version"`
	ScopeID          string `json:"scope_id"`
}

// TrendMetrics 描述单个资产的趋势温度、状态机、节气与梯队指标。
type TrendMetrics struct {
	Temperature          string           `json:"temperature"`
	TemperatureScore     *float64         `json:"temperature_score"`
	Direction            string           `json:"direction"`
	Confirmation         string           `json:"confirmation"`
	Side                 string           `json:"side"`
	StartedAt            *string          `json:"started_at"`
	ConfirmedAt          *string          `json:"confirmed_at"`
	DurationTradingDays  *int             `json:"duration_trading_days"`
	DurationCalendarDays *int             `json:"duration_calendar_days"`
	Phase                *string          `json:"phase"`
	Strength             *StrengthInfo    `json:"strength"`
	Tier                 *string          `json:"tier"`
	Events               []TrendEventItem `json:"events"`
	Reason               string           `json:"reason,omitempty"`
}

// AssetSummaryItem 描述标准化资产或观察池的摘要条目。
type AssetSummaryItem struct {
	AssetID       string       `json:"asset_id"`
	Market        string       `json:"market"`
	Symbol        string       `json:"symbol"`
	Name          string       `json:"name"`
	ShortName     string       `json:"short_name"`
	AssetType     string       `json:"asset_type"`
	ObjectType    string       `json:"object_type,omitempty"`
	RouteTarget   string       `json:"route_target,omitempty"`
	MeasureType   string       `json:"measure_type"`
	Unit          string       `json:"unit"`
	Currency      string       `json:"currency"`
	AsOf          *string      `json:"as_of"`
	LatestValue   *float64     `json:"latest_value"`
	DayChangePct  *float64     `json:"day_change_pct"`
	Turnover      *float64     `json:"turnover"`
	ScaleValue    *float64     `json:"scale_value"`
	ScaleLabel    string       `json:"scale_label"`
	GroupName     string       `json:"group_name,omitempty"`
	HasMembers    bool         `json:"has_members"`
	MemberCount   int          `json:"member_count"`
	RightRatioPct *float64     `json:"right_ratio_pct"`
	RightCount    int          `json:"right_count"`
	ValidCount    int          `json:"valid_count"`
	LicenseStatus string       `json:"license_status"`
	QualityStatus string       `json:"quality_status"`
	Metrics       TrendMetrics `json:"metrics"`
	FeatureTags   []string     `json:"feature_tags"`
}

type CatalogUniverseItem struct {
	UniverseID  string `json:"universe_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MemberCount int    `json:"member_count"`
}

type CatalogCategoryItem struct {
	CategoryID   string  `json:"category_id"`
	ParentID     *string `json:"parent_id"`
	Name         string  `json:"name"`
	Level        int     `json:"level"`
	HasIndex     bool    `json:"has_index"`
	IndexAssetID *string `json:"index_asset_id"`
}

type CatalogOverview struct {
	RightCount     int      `json:"right_count"`
	LeftCount      int      `json:"left_count"`
	FormingCount   int      `json:"forming_count"`
	OtherCount     int      `json:"other_count"`
	UncoveredCount int      `json:"uncovered_count"`
	ValidCount     int      `json:"valid_count"`
	TotalCount     int      `json:"total_count"`
	RightRatioPct  *float64 `json:"right_ratio_pct"`
	LeftRatioPct   *float64 `json:"left_ratio_pct"`
	OtherRatioPct  *float64 `json:"other_ratio_pct"`
}

type CatalogResponse struct {
	Snapshot   SnapshotMeta          `json:"snapshot"`
	Universes  []CatalogUniverseItem `json:"universes"`
	Categories []CatalogCategoryItem `json:"categories"`
	Overview   CatalogOverview       `json:"overview"`
}

type UniversesResponse struct {
	Items    []CatalogUniverseItem `json:"items"`
	Snapshot SnapshotMeta          `json:"snapshot"`
}

type SearchAssetsResponse struct {
	Items          []AssetSummaryItem `json:"items"`
	NextCursor     *string            `json:"next_cursor"`
	Total          int                `json:"total"`
	UncoveredCount int                `json:"uncovered_count"`
	UncoveredNames []string           `json:"uncovered_names,omitempty"`
}

type EnumOption struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type AssetEnumsResponse struct {
	SearchTypes    []EnumOption `json:"search_types"`
	AssetTypes     []EnumOption `json:"asset_types"`
	SortFields     []EnumOption `json:"sort_fields"`
	DashboardSorts []EnumOption `json:"dashboard_sorts"`
	MatrixSorts    []EnumOption `json:"matrix_sorts"`
	Currencies     []EnumOption `json:"currencies"`
	Statuses       []EnumOption `json:"statuses"`
}

type MembersResponse struct {
	Snapshot         SnapshotMeta       `json:"snapshot"`
	ParentID         string             `json:"parent_id"`
	ParentName       string             `json:"parent_name"`
	Items            []AssetSummaryItem `json:"items"`
	NextCursor       *string            `json:"next_cursor"`
	Total            int                `json:"total"`
	PoolTotal        int                `json:"pool_total"`
	ValidCount       int                `json:"valid_count"`
	MatchedTotal     int                `json:"matched_total"`
	Overview         *CatalogOverview   `json:"overview,omitempty"`
	AvailableMarkets []string           `json:"available_markets"`
	SortScope        string             `json:"sort_scope,omitempty"`
}

type BreadcrumbItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PeriodChanges struct {
	Month1Pct *float64 `json:"month_1_pct"`
	Year1Pct  *float64 `json:"year_1_pct"`
}

type RangeAvailabilityItem struct {
	Range        string  `json:"range"`
	Available    bool    `json:"available"`
	EarliestDate *string `json:"earliest_date,omitempty"`
	Reason       string  `json:"reason,omitempty"`
}

type AssetDetailResponse struct {
	Snapshot            SnapshotMeta            `json:"snapshot"`
	Asset               AssetSummaryItem        `json:"asset"`
	Breadcrumbs         []BreadcrumbItem        `json:"breadcrumbs"`
	AvailableCharts     []string                `json:"available_charts"`
	AvailableRanges     []string                `json:"available_ranges"`
	RangeAvailability   []RangeAvailabilityItem `json:"range_availability,omitempty"`
	EarliestDate        *string                 `json:"earliest_date,omitempty"`
	AvailableCurrencies []string                `json:"available_currencies"`
	CurrencyNotes       []string                `json:"currency_notes,omitempty"`
	BenchmarkID         *string                 `json:"benchmark_id"`
	BenchmarkName       *string                 `json:"benchmark_name"`
	PeriodChanges       PeriodChanges           `json:"period_changes"`
	RecentEvents        []TrendEventItem        `json:"recent_events"`
}

type SeriesPoint struct {
	TS          string   `json:"ts"`
	TradingDate string   `json:"trading_date"`
	Value       *float64 `json:"value"`
	ReturnPct   *float64 `json:"return_pct"`
	Turnover    *float64 `json:"turnover"`
	Status      string   `json:"status"`
}

type ExtremumInfo struct {
	MaxDrawdownPct     *float64 `json:"max_drawdown_pct"`
	DrawdownPeakDate   *string  `json:"drawdown_peak_date"`
	DrawdownTroughDate *string  `json:"drawdown_trough_date"`
	MaxGainPct         *float64 `json:"max_gain_pct"`
	GainTroughDate     *string  `json:"gain_trough_date"`
	GainPeakDate       *string  `json:"gain_peak_date"`
	HighValue          *float64 `json:"high_value"`
	HighDate           *string  `json:"high_date"`
	LowValue           *float64 `json:"low_value"`
	LowDate            *string  `json:"low_date"`
}

type StateInterval struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Side      string `json:"side"`
}

type TrendCurveResponse struct {
	Snapshot          SnapshotMeta     `json:"snapshot"`
	AssetID           string           `json:"asset_id"`
	AssetName         string           `json:"asset_name"`
	RequestedCurrency string           `json:"requested_currency,omitempty"`
	EffectiveCurrency string           `json:"effective_currency,omitempty"`
	FXApplied         bool             `json:"fx_applied,omitempty"`
	FXPair            string           `json:"fx_pair,omitempty"`
	BenchmarkID       *string          `json:"benchmark_id"`
	BenchmarkName     *string          `json:"benchmark_name"`
	Range             string           `json:"range"`
	Freq              string           `json:"freq"`
	Points            []SeriesPoint    `json:"points"`
	BenchmarkPoints   []SeriesPoint    `json:"benchmark_points"`
	Extremum          ExtremumInfo     `json:"extremum"`
	StateIntervals    []StateInterval  `json:"state_intervals"`
	Events            []TrendEventItem `json:"events"`
}

type AnnualWeekPoint struct {
	Week      int      `json:"week"`
	EndDate   string   `json:"end_date"`
	ReturnPct *float64 `json:"return_pct"`
}

type AnnualYearSeries struct {
	Year              int               `json:"year"`
	IsCurrent         bool              `json:"is_current"`
	IsComplete        bool              `json:"is_complete"`
	SameWeekReturnPct *float64          `json:"same_week_return_pct"`
	FullYearReturnPct *float64          `json:"full_year_return_pct"`
	WeeklyReturns     []AnnualWeekPoint `json:"weekly_returns"`
}

type AnnualStats struct {
	SampleCount          int      `json:"sample_count"`
	CurrentYearReturnPct *float64 `json:"current_year_return_pct"`
	SameWeekMedianPct    *float64 `json:"same_week_median_pct"`
	SameWeekMeanPct      *float64 `json:"same_week_mean_pct"`
	CurrentRank          *int     `json:"current_rank"`
	UpRatioPct           *float64 `json:"up_ratio_pct"`
}

type AnnualTrendResponse struct {
	Snapshot   SnapshotMeta       `json:"snapshot"`
	AssetID    string             `json:"asset_id"`
	AssetName  string             `json:"asset_name"`
	CutoffWeek int                `json:"cutoff_week"`
	Years      []AnnualYearSeries `json:"years"`
	Stats      AnnualStats        `json:"stats"`
}

type ComponentTrendSeries struct {
	AssetID         string     `json:"asset_id"`
	Name            string     `json:"name"`
	Color           string     `json:"color"`
	IsBenchmark     bool       `json:"is_benchmark"`
	Temperature     string     `json:"temperature"`
	LatestValue     *float64   `json:"latest_value"`
	PeriodChangePct *float64   `json:"period_change_pct"`
	Turnover        *float64   `json:"turnover"`
	TotalReturnPct  *float64   `json:"total_return_pct"`
	Points          []*float64 `json:"points"`
	Reason          string     `json:"reason,omitempty"`
}

type ComponentTrendResponse struct {
	Snapshot  SnapshotMeta           `json:"snapshot"`
	AssetID   string                 `json:"asset_id"`
	AssetName string                 `json:"asset_name"`
	Range     string                 `json:"range"`
	Timeline  []string               `json:"timeline"`
	Series    []ComponentTrendSeries `json:"series"`
}

type DashboardRow struct {
	Asset     AssetSummaryItem `json:"asset"`
	ChangePct *float64         `json:"change_pct"`
	Sparkline []*float64       `json:"sparkline"`
	SparkMin  *float64         `json:"spark_min"`
	SparkMax  *float64         `json:"spark_max"`
}

type DashboardGroup struct {
	GroupName string         `json:"group_name"`
	Rows      []DashboardRow `json:"rows"`
}

type TrendDashboardResponse struct {
	Snapshot     SnapshotMeta     `json:"snapshot"`
	AssetID      string           `json:"asset_id"`
	AssetName    string           `json:"asset_name"`
	ChangePeriod string           `json:"change_period"`
	SparkRange   string           `json:"spark_range"`
	Groups       []DashboardGroup `json:"groups"`
}

type SeasonMemberItem struct {
	AssetID             string `json:"asset_id"`
	Name                string `json:"name"`
	DurationTradingDays *int   `json:"duration_trading_days"`
	Temperature         string `json:"temperature"`
	Side                string `json:"side"`
}

type SeasonStageBucket struct {
	StageKey   string             `json:"stage_key"`
	StageLabel string             `json:"stage_label"`
	Direction  string             `json:"direction"`
	RuleDesc   string             `json:"rule_desc"`
	Count      int                `json:"count"`
	Members    []SeasonMemberItem `json:"members"`
}

type TrendSeasonResponse struct {
	Snapshot   SnapshotMeta        `json:"snapshot"`
	AssetID    string              `json:"asset_id"`
	AssetName  string              `json:"asset_name"`
	UpCount    int                 `json:"up_count"`
	DownCount  int                 `json:"down_count"`
	OtherCount int                 `json:"other_count"`
	ValidCount int                 `json:"valid_count"`
	Stages     []SeasonStageBucket `json:"stages"`
}

type DistributionNode struct {
	AssetID             string   `json:"asset_id"`
	Name                string   `json:"name"`
	GroupName           string   `json:"group_name"`
	WeightValue         float64  `json:"weight_value"`
	WeightPct           float64  `json:"weight_pct"`
	Side                string   `json:"side"`
	Temperature         string   `json:"temperature"`
	DayChangePct        *float64 `json:"day_change_pct"`
	DurationTradingDays *int     `json:"duration_trading_days"`
	StrengthScore       *float64 `json:"strength_score"`
}

type TrendDistributionResponse struct {
	Snapshot              SnapshotMeta       `json:"snapshot"`
	AssetID               string             `json:"asset_id"`
	AssetName             string             `json:"asset_name"`
	WeightBy              string             `json:"weight_by"`
	ColorBy               string             `json:"color_by"`
	RightCount            int                `json:"right_count"`
	ValidCount            int                `json:"valid_count"`
	UnknownCount          int                `json:"unknown_count"`
	RightRatioPct         *float64           `json:"right_ratio_pct"`
	WeightedRightRatioPct *float64           `json:"weighted_right_ratio_pct"`
	WeightCoveragePct     *float64           `json:"weight_coverage_pct"`
	Nodes                 []DistributionNode `json:"nodes"`
}

type TierMemberItem struct {
	AssetID             string   `json:"asset_id"`
	Name                string   `json:"name"`
	Score               float64  `json:"score"`
	Side                string   `json:"side"`
	DurationTradingDays *int     `json:"duration_trading_days"`
	Delta               *float64 `json:"delta"`
	Temperature         string   `json:"temperature"`
}

type TierBucket struct {
	Tier     string           `json:"tier"`
	ProLabel string           `json:"pro_label"`
	FunLabel string           `json:"fun_label"`
	RuleText string           `json:"rule_text"`
	Count    int              `json:"count"`
	Items    []TierMemberItem `json:"items"`
}

type TierExcludedItem struct {
	AssetID string `json:"asset_id"`
	Name    string `json:"name"`
	Reason  string `json:"reason"`
}

type TrendTierResponse struct {
	Snapshot       SnapshotMeta       `json:"snapshot"`
	AssetID        string             `json:"asset_id"`
	AssetName      string             `json:"asset_name"`
	CandidateCount int                `json:"candidate_count"`
	ValidCount     int                `json:"valid_count"`
	ExcludedCount  int                `json:"excluded_count"`
	Tiers          []TierBucket       `json:"tiers"`
	ExcludedItems  []TierExcludedItem `json:"excluded_items"`
}

type TrendEventsResponse struct {
	Items      []TrendEventItem `json:"items"`
	NextCursor *string          `json:"next_cursor"`
}

type DataStatusMarketItem struct {
	Market       string  `json:"market"`
	Label        string  `json:"label"`
	AsOf         *string `json:"as_of"`
	SourceName   string  `json:"source_name"`
	Status       string  `json:"status"`
	CoveredCount int     `json:"covered_count"`
	TotalCount   int     `json:"total_count"`
	Note         string  `json:"note"`
}

type DataStatusResponse struct {
	Snapshot SnapshotMeta           `json:"snapshot"`
	Markets  []DataStatusMarketItem `json:"markets"`
}

type TemperatureThresholdItem struct {
	Key   string  `json:"key"`
	Label string  `json:"label"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Desc  string  `json:"desc"`
}

type TierRuleItem struct {
	Tier string `json:"tier"`
	Rule string `json:"rule"`
}

type MethodologyResponse struct {
	AlgorithmVersion      string                     `json:"algorithm_version"`
	TemperatureThresholds []TemperatureThresholdItem `json:"temperature_thresholds"`
	TierRules             []TierRuleItem             `json:"tier_rules"`
	Disclaimers           []string                   `json:"disclaimers"`
}

type MatrixColumnDef struct {
	Key        string `json:"key"`
	Label      string `json:"label"`
	PeriodType string `json:"period_type"`
	Sortable   bool   `json:"sortable"`
}

type MatrixCell struct {
	PeriodKey     string   `json:"period_key"`
	StartDate     *string  `json:"start_date"`
	EndDate       *string  `json:"end_date"`
	StartValue    *float64 `json:"start_value"`
	EndValue      *float64 `json:"end_value"`
	ReturnPct     *float64 `json:"return_pct"`
	BpChange      *float64 `json:"bp_change,omitempty"`
	DisplayUnit   string   `json:"display_unit"`
	ReturnType    string   `json:"return_type"`
	Currency      string   `json:"currency"`
	FXApplied     bool     `json:"fx_applied,omitempty"`
	FXPair        string   `json:"fx_pair,omitempty"`
	StartFX       *float64 `json:"start_fx,omitempty"`
	EndFX         *float64 `json:"end_fx,omitempty"`
	QualityStatus string   `json:"quality_status"`
	SourceName    string   `json:"source_name"`
	Reason        string   `json:"reason,omitempty"`
}

type MatrixRow struct {
	Asset AssetSummaryItem      `json:"asset"`
	Cells map[string]MatrixCell `json:"cells"`
}

type MatrixGroup struct {
	GroupName string      `json:"group_name"`
	Rows      []MatrixRow `json:"rows"`
}

type PeriodMatrixResponse struct {
	Snapshot         SnapshotMeta      `json:"snapshot"`
	PoolID           string            `json:"pool_id"`
	PoolName         string            `json:"pool_name"`
	Currency         string            `json:"currency"`
	CutoffDate       string            `json:"cutoff_date"`
	Columns          []MatrixColumnDef `json:"columns"`
	Groups           []MatrixGroup     `json:"groups"`
	NextCursor       *string           `json:"next_cursor"`
	Total            int               `json:"total"`
	PoolTotal        int               `json:"pool_total"`
	ValidCount       int               `json:"valid_count"`
	MatchedTotal     int               `json:"matched_total"`
	Overview         *CatalogOverview  `json:"overview,omitempty"`
	AvailableMarkets []string          `json:"available_markets"`
	SortScope        string            `json:"sort_scope"`
	SortField        string            `json:"sort_field"`
	SortOrder        string            `json:"sort_order"`
}
