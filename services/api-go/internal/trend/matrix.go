package trend

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// MatrixColumnSpec defines a period return matrix column.
type MatrixColumnSpec struct {
	Key        string
	Label      string
	PeriodType string // "calendar" | "rolling"
	Sortable   bool
}

// DefaultMatrixColumns returns the fixed 6-column definition for Financial Assets V5.
func DefaultMatrixColumns() []MatrixColumnSpec {
	return []MatrixColumnSpec{
		{Key: "wtd", Label: "本周", PeriodType: "calendar", Sortable: true},
		{Key: "mtd", Label: "本月", PeriodType: "calendar", Sortable: true},
		{Key: "qtd", Label: "本季", PeriodType: "calendar", Sortable: true},
		{Key: "ytd", Label: "今年来", PeriodType: "calendar", Sortable: true},
		{Key: "prev_year", Label: "去年全年", PeriodType: "calendar", Sortable: true},
		{Key: "rolling_1m", Label: "近1月", PeriodType: "rolling", Sortable: true},
	}
}

// MatrixBarPoint represents an ordered daily observation for period matrix calculation.
type MatrixBarPoint struct {
	TradingDate      string
	Close            *float64
	AdjustedClose    *float64
	TotalReturnIndex *float64
	YieldPct         *float64
	QualityStatus    string
}

// MatrixAssetInput carries asset metadata needed by period matrix rules.
type MatrixAssetInput struct {
	AssetID       string
	Market        string
	MeasureType   string
	Currency      string
	Timezone      string
	LicenseStatus string
	SourceName    string
}

// MatrixFXPoint is a dated FX quote for pair Base/Quote (1 Base = Rate Quote).
type MatrixFXPoint struct {
	TradingDate string
	Rate        float64
}

// MatrixFXTable stores ascending dated FX series keyed by canonical pair "BASE/QUOTE".
type MatrixFXTable struct {
	series map[string][]MatrixFXPoint
}

// NewMatrixFXTable builds a canonical FX table from raw pair maps.
func NewMatrixFXTable(raw map[string][]MatrixFXPoint) *MatrixFXTable {
	tbl := &MatrixFXTable{
		series: make(map[string][]MatrixFXPoint),
	}
	for k, pts := range raw {
		canon := canonicalPairName(k)
		if canon == "" || len(pts) == 0 {
			continue
		}
		cp := make([]MatrixFXPoint, len(pts))
		copy(cp, pts)
		sort.SliceStable(cp, func(i, j int) bool {
			return cp[i].TradingDate < cp[j].TradingDate
		})
		// Merge/overwrite existing canonical key if multiple aliases provided
		tbl.series[canon] = append(tbl.series[canon], cp...)
		if len(tbl.series[canon]) > len(cp) {
			sort.SliceStable(tbl.series[canon], func(i, j int) bool {
				return tbl.series[canon][i].TradingDate < tbl.series[canon][j].TradingDate
			})
		}
	}
	return tbl
}

// MatrixCellResult is the computed cell value and audit metadata for a single period.
type MatrixCellResult struct {
	PeriodKey     string
	StartDate     *string
	EndDate       *string
	StartValue    *float64
	EndValue      *float64
	ReturnPct     *float64
	BpChange      *float64
	DisplayUnit   string // "%" | "bp"
	ReturnType    string
	Currency      string
	FXApplied     bool
	FXPair        string
	StartFX       *float64
	EndFX         *float64
	QualityStatus string
	SourceName    string
	Reason        string
}

// NormalizeMatrixCurrency maps aliases like RAW/CNH into canonical currency codes.
func NormalizeMatrixCurrency(raw string) string {
	u := strings.ToUpper(strings.TrimSpace(raw))
	switch u {
	case "", "RAW", "ORIGINAL":
		return "ORIGINAL"
	case "CNH", "RMB":
		return "CNY"
	default:
		return u
	}
}

func canonicalPairName(raw string) string {
	u := strings.ToUpper(strings.TrimSpace(raw))
	u = strings.ReplaceAll(u, "-", "/")
	u = strings.ReplaceAll(u, "_", "/")
	if strings.Contains(u, "/") {
		parts := strings.Split(u, "/")
		if len(parts) == 2 {
			a := normalizeCcyUnit(parts[0])
			b := normalizeCcyUnit(parts[1])
			if a != "" && b != "" {
				return a + "/" + b
			}
		}
		return ""
	}
	if len(u) == 6 {
		a := normalizeCcyUnit(u[:3])
		b := normalizeCcyUnit(u[3:])
		return a + "/" + b
	}
	return ""
}

func normalizeCcyUnit(c string) string {
	u := strings.ToUpper(strings.TrimSpace(c))
	if u == "CNH" || u == "RMB" {
		return "CNY"
	}
	return u
}

func round2Val(v float64) float64 {
	return math.Round(v*100) / 100
}

func round4Val(v float64) float64 {
	return math.Round(v*10000) / 10000
}

func effectiveMatrixBarValue(b MatrixBarPoint, measureType string) (val float64, returnType string, ok bool) {
	if measureType == "yield_pct" {
		if b.YieldPct != nil {
			return *b.YieldPct, "yield_bp", true
		}
		if b.Close != nil {
			return *b.Close, "yield_bp", true
		}
		return 0, "yield_bp", false
	}
	if b.TotalReturnIndex != nil && *b.TotalReturnIndex > 0 {
		return *b.TotalReturnIndex, "total_return", true
	}
	if b.AdjustedClose != nil && *b.AdjustedClose > 0 {
		return *b.AdjustedClose, "adjusted_price", true
	}
	if b.Close != nil && *b.Close > 0 {
		return *b.Close, "price", true
	}
	return 0, "price", false
}

// LookupRate returns 1 fromCcy in toCcy at or prior to targetDate.
func (t *MatrixFXTable) LookupRate(fromCcy, toCcy, targetDate string) (float64, bool) {
	f := normalizeCcyUnit(fromCcy)
	to := normalizeCcyUnit(toCcy)
	if f == "" || to == "" || f == to {
		return 1.0, true
	}
	if t == nil {
		return 0, false
	}
	if r, ok := t.directRate(f, to, targetDate); ok {
		return r, true
	}
	// Try triangulation through USD if neither endpoint is USD
	if f != "USD" && to != "USD" {
		r1, ok1 := t.directRate(f, "USD", targetDate)
		r2, ok2 := t.directRate("USD", to, targetDate)
		if ok1 && ok2 {
			return r1 * r2, true
		}
	}
	return 0, false
}

func (t *MatrixFXTable) directRate(fromCcy, toCcy, targetDate string) (float64, bool) {
	if pts, ok := t.series[fromCcy+"/"+toCcy]; ok {
		if r, found := priorOrEqualRate(pts, targetDate); found && r > 0 {
			return r, true
		}
	}
	if pts, ok := t.series[toCcy+"/"+fromCcy]; ok {
		if r, found := priorOrEqualRate(pts, targetDate); found && r > 0 {
			return 1.0 / r, true
		}
	}
	return 0, false
}

func priorOrEqualRate(pts []MatrixFXPoint, targetDate string) (float64, bool) {
	if len(pts) == 0 {
		return 0, false
	}
	idx := sort.Search(len(pts), func(i int) bool {
		return pts[i].TradingDate > targetDate
	})
	for i := idx - 1; i >= 0; i-- {
		if pts[i].Rate > 0 {
			return pts[i].Rate, true
		}
	}
	return 0, false
}

type validBar struct {
	date       string
	value      float64
	returnType string
	quality    string
}

func filterValidMatrixBars(bars []MatrixBarPoint, measureType, cutoffDate string) []validBar {
	out := make([]validBar, 0, len(bars))
	lastDate := ""
	for _, b := range bars {
		if cutoffDate != "" && b.TradingDate > cutoffDate {
			continue
		}
		if b.TradingDate == lastDate {
			continue
		}
		v, rt, ok := effectiveMatrixBarValue(b, measureType)
		if !ok {
			continue
		}
		lastDate = b.TradingDate
		out = append(out, validBar{
			date:       b.TradingDate,
			value:      v,
			returnType: rt,
			quality:    b.QualityStatus,
		})
	}
	return out
}

func lastBarBefore(vbars []validBar, boundaryExclusive string) (validBar, bool) {
	idx := sort.Search(len(vbars), func(i int) bool {
		return vbars[i].date >= boundaryExclusive
	})
	if idx-1 >= 0 {
		return vbars[idx-1], true
	}
	return validBar{}, false
}

func lastBarOnOrBefore(vbars []validBar, boundaryInclusive string) (validBar, bool) {
	idx := sort.Search(len(vbars), func(i int) bool {
		return vbars[i].date > boundaryInclusive
	})
	if idx-1 >= 0 {
		return vbars[idx-1], true
	}
	return validBar{}, false
}

func loadAssetLocation(tz string) *time.Location {
	if strings.TrimSpace(tz) != "" {
		if loc, err := time.LoadLocation(strings.TrimSpace(tz)); err == nil {
			return loc
		}
	}
	return time.UTC
}

func parseDateInLocation(dateStr string, loc *time.Location) (time.Time, bool) {
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(dateStr), loc)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// ComputeAssetPeriodMatrix computes all 6 period cells for an asset as of cutoffDate.
func ComputeAssetPeriodMatrix(
	asset MatrixAssetInput,
	bars []MatrixBarPoint,
	targetCurrency string,
	cutoffDate string,
	fxTable *MatrixFXTable,
) map[string]MatrixCellResult {
	cols := DefaultMatrixColumns()
	result := make(map[string]MatrixCellResult, len(cols))
	reqCcy := NormalizeMatrixCurrency(targetCurrency)
	assetCcy := normalizeCcyUnit(asset.Currency)
	isYield := asset.MeasureType == "yield_pct"

	displayCcy := reqCcy
	if isYield || reqCcy == "ORIGINAL" || assetCcy == "" {
		displayCcy = assetCcy
		if displayCcy == "" {
			displayCcy = "ORIGINAL"
		}
	}
	defaultUnit := "%"
	defaultRetType := "price"
	if isYield {
		defaultUnit = "bp"
		defaultRetType = "yield_bp"
	}

	makeMissing := func(periodKey, reason, quality string) MatrixCellResult {
		return MatrixCellResult{
			PeriodKey:     periodKey,
			DisplayUnit:   defaultUnit,
			ReturnType:    defaultRetType,
			Currency:      displayCcy,
			QualityStatus: quality,
			SourceName:    asset.SourceName,
			Reason:        reason,
		}
	}

	if asset.LicenseStatus == "not_covered" {
		for _, c := range cols {
			result[c.Key] = makeMissing(c.Key, "合规与授权暂未覆盖该资产", "not_covered")
		}
		return result
	}

	vbars := filterValidMatrixBars(bars, asset.MeasureType, cutoffDate)
	if len(vbars) == 0 {
		for _, c := range cols {
			result[c.Key] = makeMissing(c.Key, "缺少历史行情数据", "insufficient")
		}
		return result
	}

	loc := loadAssetLocation(asset.Timezone)
	latestBar := vbars[len(vbars)-1]
	effectiveCutoff := strings.TrimSpace(cutoffDate)
	if effectiveCutoff == "" {
		effectiveCutoff = latestBar.date
	}
	refTime, ok := parseDateInLocation(effectiveCutoff, loc)
	if !ok {
		refTime, _ = parseDateInLocation(latestBar.date, loc)
	}

	cutoffLatestTime, _ := parseDateInLocation(latestBar.date, loc)
	refYear, refMonth, _ := refTime.Date()

	// Precompute calendar boundaries (in asset timezone)
	wd := int(refTime.Weekday())
	if wd == 0 {
		wd = 7 // Sunday = 7 in ISO week
	}
	weekStart := refTime.AddDate(0, 0, -(wd - 1)).Format("2006-01-02")
	monthStart := fmt.Sprintf("%04d-%02d-01", refYear, int(refMonth))
	qMonth := ((int(refMonth)-1)/3)*3 + 1
	quarterStart := fmt.Sprintf("%04d-%02d-01", refYear, qMonth)
	yearStart := fmt.Sprintf("%04d-01-01", refYear)
	prevYearStart := fmt.Sprintf("%04d-01-01", refYear-1)
	prevYearEnd := fmt.Sprintf("%04d-12-31", refYear-1)

	buildCell := func(periodKey string, startBar, endBar validBar) MatrixCellResult {
		sd := startBar.date
		ed := endBar.date
		sv := round4Val(startBar.value)
		ev := round4Val(endBar.value)
		retType := endBar.returnType
		quality := endBar.quality
		if quality == "" {
			quality = "ok"
		}

		if isYield {
			bp := round2Val((endBar.value - startBar.value) * 100.0)
			return MatrixCellResult{
				PeriodKey:     periodKey,
				StartDate:     &sd,
				EndDate:       &ed,
				StartValue:    &sv,
				EndValue:      &ev,
				ReturnPct:     nil,
				BpChange:      &bp,
				DisplayUnit:   "bp",
				ReturnType:    "yield_bp",
				Currency:      displayCcy,
				FXApplied:     false,
				QualityStatus: quality,
				SourceName:    asset.SourceName,
				Reason:        "利率类资产展示基点变化 (bp)，不参与收益率排行",
			}
		}

		if startBar.value <= 0 || endBar.value <= 0 {
			return makeMissing(periodKey, "期初或期末净值非正，无法计算收益率", "insufficient")
		}

		needsFX := reqCcy != "ORIGINAL" && assetCcy != "" && reqCcy != assetCcy && asset.MeasureType != "fx_rate"
		startValConv := startBar.value
		endValConv := endBar.value
		var startFXPtr, endFXPtr *float64
		fxApplied := false
		fxPair := ""

		if needsFX {
			rStart, okStart := fxTable.LookupRate(assetCcy, reqCcy, startBar.date)
			rEnd, okEnd := fxTable.LookupRate(assetCcy, reqCcy, endBar.date)
			if !okStart || !okEnd {
				return MatrixCellResult{
					PeriodKey:     periodKey,
					StartDate:     &sd,
					EndDate:       &ed,
					StartValue:    &sv,
					EndValue:      &ev,
					ReturnPct:     nil,
					DisplayUnit:   "%",
					ReturnType:    retType,
					Currency:      reqCcy,
					FXApplied:     false,
					QualityStatus: "degraded",
					SourceName:    asset.SourceName,
					Reason:        fmt.Sprintf("缺少 %s/%s 在 %s 至 %s 的汇率数据 (missing_fx_rate)", assetCcy, reqCcy, startBar.date, endBar.date),
				}
			}
			sfx := round4Val(rStart)
			efx := round4Val(rEnd)
			startFXPtr = &sfx
			endFXPtr = &efx
			startValConv = startBar.value * rStart
			endValConv = endBar.value * rEnd
			fxApplied = true
			fxPair = assetCcy + "/" + reqCcy
		}

		retPct := round2Val(((endValConv / startValConv) - 1.0) * 100.0)
		return MatrixCellResult{
			PeriodKey:     periodKey,
			StartDate:     &sd,
			EndDate:       &ed,
			StartValue:    &sv,
			EndValue:      &ev,
			ReturnPct:     &retPct,
			DisplayUnit:   "%",
			ReturnType:    retType,
			Currency:      displayCcy,
			FXApplied:     fxApplied,
			FXPair:        fxPair,
			StartFX:       startFXPtr,
			EndFX:         endFXPtr,
			QualityStatus: quality,
			SourceName:    asset.SourceName,
		}
	}

	computeCalendarPeriod := func(periodKey, periodStart, noCurrentMsg, noBaseMsg string) MatrixCellResult {
		if latestBar.date < periodStart {
			return makeMissing(periodKey, noCurrentMsg, "insufficient")
		}
		sb, ok := lastBarBefore(vbars, periodStart)
		if !ok {
			return makeMissing(periodKey, noBaseMsg, "insufficient")
		}
		return buildCell(periodKey, sb, latestBar)
	}

	result["wtd"] = computeCalendarPeriod("wtd", weekStart, "本周尚无交易日数据", "缺少上周期末基准日数据")
	result["mtd"] = computeCalendarPeriod("mtd", monthStart, "本月尚无交易日数据", "缺少上月末基准日数据")
	result["qtd"] = computeCalendarPeriod("qtd", quarterStart, "本季尚无交易日数据", "缺少上季末基准日数据")
	result["ytd"] = computeCalendarPeriod("ytd", yearStart, "本年尚无交易日数据", "缺少上年末基准日数据")

	// prev_year: end <= (Y-1)-12-31 and >= (Y-1)-01-01; start < (Y-1)-01-01
	prevEnd, okPrevEnd := lastBarOnOrBefore(vbars, prevYearEnd)
	if !okPrevEnd || prevEnd.date < prevYearStart {
		result["prev_year"] = makeMissing("prev_year", "缺少去年年末有效行情数据", "insufficient")
	} else {
		prevStart, okPrevStart := lastBarBefore(vbars, prevYearStart)
		if !okPrevStart {
			result["prev_year"] = makeMissing("prev_year", "缺少前年年末期初基准日数据", "insufficient")
		} else {
			result["prev_year"] = buildCell("prev_year", prevStart, prevEnd)
		}
	}

	// rolling_1m: lookup prior valid close <= (endBar.date - 1 month); reject if latestBar stale > 35d before cutoff
	if refTime.Sub(cutoffLatestTime) > 35*24*time.Hour {
		result["rolling_1m"] = makeMissing("rolling_1m", fmt.Sprintf("资产截至 %s 已长期停牌或未更新", latestBar.date), "insufficient")
	} else {
		targetRolling := cutoffLatestTime.AddDate(0, -1, 0).Format("2006-01-02")
		sb, okRoll := lastBarOnOrBefore(vbars, targetRolling)
		if !okRoll || sb.date >= latestBar.date {
			result["rolling_1m"] = makeMissing("rolling_1m", "历史起始晚于近1月基准日，样本不足", "insufficient")
		} else {
			result["rolling_1m"] = buildCell("rolling_1m", sb, latestBar)
		}
	}

	return result
}
