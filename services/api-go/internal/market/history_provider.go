package market

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultEastmoneyKlineURL = "https://push2his.eastmoney.com/api/qt/stock/kline/get"
	DefaultTencentKlineURL   = "https://web.ifzq.gtimg.cn/appstock/app/fqkline/get"
	DefaultHistoryTimeout    = 5 * time.Second
	DefaultMinReqInterval    = 120 * time.Millisecond
)

// DailyBar 表示标准化后的单日历史行情/指标数据点（可空字段使用指针，绝不用 0 填充缺失）。
type DailyBar struct {
	AssetID          string      `json:"asset_id"`
	Market           string      `json:"market"`
	Symbol           string      `json:"symbol"`
	MeasureType      MeasureType `json:"measure_type"`
	PriceType        string      `json:"price_type"`
	ReturnType       string      `json:"return_type"`
	Unit             string      `json:"unit"`
	Currency         string      `json:"currency"`
	Timezone         string      `json:"timezone"`
	TradingDate      string      `json:"trading_date"`
	AsOf             string      `json:"as_of"`
	AvailableAt      string      `json:"available_at"`
	SourceName       string      `json:"source_name"`
	QualityStatus    string      `json:"quality_status"`
	Open             *float64    `json:"open"`
	High             *float64    `json:"high"`
	Low              *float64    `json:"low"`
	Close            float64     `json:"close"`
	AdjustedClose    *float64    `json:"adjusted_close"`
	TotalReturnIndex *float64    `json:"total_return_index"`
	Volume           *float64    `json:"volume"`
	Turnover         *float64    `json:"turnover"`
	YieldPct         *float64    `json:"yield_pct"`
}

// HistoryResult 封装单资产历史查询结果与数据质量元数据。
type HistoryResult struct {
	Asset         CatalogAsset `json:"asset"`
	Bars          []DailyBar   `json:"bars"`
	QualityStatus string       `json:"quality_status"`
	AsOf          string       `json:"as_of"`
	SourceName    string       `json:"source_name"`
	Message       string       `json:"message,omitempty"`
}

// CompositeIndexPoint 表示链式收益合成指数的单日点位。
type CompositeIndexPoint struct {
	TradingDate     string  `json:"trading_date"`
	Value           float64 `json:"value"`
	DailyReturn     float64 `json:"daily_return"`
	Coverage        float64 `json:"coverage"`
	ValidAssetCount int     `json:"valid_asset_count"`
}

// HistoricalProviderConfig 配置 HistoricalProvider。
type HistoricalProviderConfig struct {
	HTTPClient       *http.Client
	EastmoneyBaseURL string
	TencentBaseURL   string
	SnapshotDir      string
	SeedFallback     *SeedProvider
	MinInterval      time.Duration
	OfflineOnly      bool
	NowFunc          func() time.Time
}

// HistoricalProvider 负责合规拉取与清洗真实日频历史数据，并在离线/不可用时安全降级。
type HistoricalProvider struct {
	client           *http.Client
	eastmoneyBaseURL string
	tencentBaseURL   string
	snapshotDir      string
	seedFallback     *SeedProvider
	minInterval      time.Duration
	offlineOnly      bool
	nowFunc          func() time.Time

	rateMu      sync.Mutex
	lastRequest time.Time
}

// NewHistoricalProvider 创建历史行情提供者（默认 5s 超时与限流保护）。
func NewHistoricalProvider(cfg HistoricalProviderConfig) *HistoricalProvider {
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: DefaultHistoryTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     60 * time.Second,
			},
		}
	}
	emURL := strings.TrimSpace(cfg.EastmoneyBaseURL)
	if emURL == "" {
		emURL = DefaultEastmoneyKlineURL
	}
	txURL := strings.TrimSpace(cfg.TencentBaseURL)
	if txURL == "" {
		txURL = DefaultTencentKlineURL
	}
	minInt := cfg.MinInterval
	if minInt < 0 {
		minInt = 0
	} else if minInt == 0 && !cfg.OfflineOnly {
		minInt = DefaultMinReqInterval
	}
	nowFn := cfg.NowFunc
	if nowFn == nil {
		nowFn = time.Now
	}

	return &HistoricalProvider{
		client:           client,
		eastmoneyBaseURL: emURL,
		tencentBaseURL:   txURL,
		snapshotDir:      cfg.SnapshotDir,
		seedFallback:     cfg.SeedFallback,
		minInterval:      minInt,
		offlineOnly:      cfg.OfflineOnly,
		nowFunc:          nowFn,
	}
}

func (p *HistoricalProvider) waitRateLimit(ctx context.Context) error {
	if p.minInterval <= 0 {
		return nil
	}
	p.rateMu.Lock()
	elapsed := time.Since(p.lastRequest)
	wait := p.minInterval - elapsed
	if wait > 0 {
		timer := time.NewTimer(wait)
		p.rateMu.Unlock()
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		p.rateMu.Lock()
	}
	p.lastRequest = time.Now()
	p.rateMu.Unlock()
	return nil
}

// FetchDailyHistory 按 asset_id 获取日频历史序列（自动校验权限、日期与价格，绝不伪造行情）。
func (p *HistoricalProvider) FetchDailyHistory(ctx context.Context, assetID string, limit int) HistoryResult {
	asset, ok := GetCatalogAsset(assetID)
	if !ok {
		return HistoryResult{
			QualityStatus: "not_covered",
			Message:       fmt.Sprintf("unknown asset_id: %s", assetID),
		}
	}

	if asset.LicenseStatus == LicenseNotCovered {
		return HistoryResult{
			Asset:         asset,
			QualityStatus: "not_covered",
			Message:       firstNonEmpty(asset.LicenseNote, "asset is not covered under current license policy"),
		}
	}

	var bars []DailyBar
	var sourceName string
	var fetchErr error

	if !p.offlineOnly {
		switch {
		case asset.Provider == "eastmoney" && asset.EastmoneySecID != "":
			bars, sourceName, fetchErr = p.fetchFromEastmoney(ctx, asset, limit)
			if (fetchErr != nil || len(bars) == 0) && asset.TencentCode != "" {
				txBars, txSrc, txErr := p.fetchFromTencent(ctx, asset, limit)
				if txErr == nil && len(txBars) > 0 {
					bars, sourceName, fetchErr = txBars, txSrc, nil
				}
			}
		case asset.Provider == "tencent" && asset.TencentCode != "":
			bars, sourceName, fetchErr = p.fetchFromTencent(ctx, asset, limit)
		}
	}

	if len(bars) > 0 {
		if limit > 0 && len(bars) > limit {
			bars = bars[len(bars)-limit:]
		}
		return HistoryResult{
			Asset:         asset,
			Bars:          bars,
			QualityStatus: "ok",
			AsOf:          bars[len(bars)-1].AsOf,
			SourceName:    sourceName,
		}
	}

	// 网络不可用、离线模式或仅本地种子资产：尝试加载本地历史快照（保留原日期，绝不编造）
	snapBars, snapSource := p.loadLocalSnapshot(asset)
	if len(snapBars) > 0 {
		if limit > 0 && len(snapBars) > limit {
			snapBars = snapBars[len(snapBars)-limit:]
		}
		msg := "loaded from local offline snapshot"
		if fetchErr != nil {
			msg = fmt.Sprintf("upstream unavailable (%v); downgraded to local snapshot", fetchErr)
		}
		return HistoryResult{
			Asset:         asset,
			Bars:          snapBars,
			QualityStatus: "offline_seed",
			AsOf:          snapBars[len(snapBars)-1].AsOf,
			SourceName:    snapSource,
			Message:       msg,
		}
	}

	status := "insufficient"
	msg := "no historical bars available"
	if fetchErr != nil {
		msg = fmt.Sprintf("upstream error (%v) and no local snapshot available", fetchErr)
	}
	return HistoryResult{
		Asset:         asset,
		QualityStatus: status,
		Message:       msg,
	}
}

func (p *HistoricalProvider) fetchFromEastmoney(ctx context.Context, asset CatalogAsset, limit int) ([]DailyBar, string, error) {
	if err := p.waitRateLimit(ctx); err != nil {
		return nil, "", err
	}

	reqCtx, cancel := context.WithTimeout(ctx, DefaultHistoryTimeout)
	defer cancel()

	lmt := 1000
	if limit > 0 && limit < lmt {
		lmt = limit
	}

	u, err := url.Parse(p.eastmoneyBaseURL)
	if err != nil {
		return nil, "", err
	}
	q := u.Query()
	q.Set("secid", asset.EastmoneySecID)
	q.Set("fields1", "f1,f2,f3,f4,f5,f6")
	q.Set("fields2", "f51,f52,f53,f54,f55,f56,f57,f58")
	q.Set("klt", "101")
	q.Set("fqt", "1")
	q.Set("lmt", strconv.Itoa(lmt))
	q.Set("end", "20500101")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "CycleCompass/4.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("eastmoney http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, "", err
	}

	var payload struct {
		RC   int `json:"rc"`
		Data *struct {
			Code   string   `json:"code"`
			Name   string   `json:"name"`
			Klines []string `json:"klines"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, "", err
	}
	if payload.Data == nil || len(payload.Data.Klines) == 0 {
		return nil, "", fmt.Errorf("eastmoney empty klines")
	}

	source := "Eastmoney Daily Kline"
	rawBars := make([]DailyBar, 0, len(payload.Data.Klines))
	for _, line := range payload.Data.Klines {
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		dateStr := strings.TrimSpace(parts[0])
		closeVal, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		if err != nil {
			continue
		}

		bar := DailyBar{
			AssetID:       asset.AssetID,
			Market:        asset.Market,
			Symbol:        asset.Symbol,
			MeasureType:   asset.MeasureType,
			PriceType:     asset.PriceType,
			ReturnType:    asset.ReturnType,
			Unit:          asset.Unit,
			Currency:      asset.Currency,
			Timezone:      asset.Timezone,
			TradingDate:   dateStr,
			AsOf:          dateStr,
			SourceName:    source,
			QualityStatus: "ok",
			Open:          parseOptionalFloat(parts, 1),
			Close:         closeVal,
			High:          parseOptionalFloat(parts, 3),
			Low:           parseOptionalFloat(parts, 4),
			Volume:        parseOptionalFloat(parts, 5),
			Turnover:      parseOptionalFloat(parts, 6),
		}
		assignMeasureSpecificFields(&bar, asset)
		rawBars = append(rawBars, bar)
	}

	cleaned := p.sanitizeBars(asset, rawBars)
	return cleaned, source, nil
}

func (p *HistoricalProvider) fetchFromTencent(ctx context.Context, asset CatalogAsset, limit int) ([]DailyBar, string, error) {
	if err := p.waitRateLimit(ctx); err != nil {
		return nil, "", err
	}

	reqCtx, cancel := context.WithTimeout(ctx, DefaultHistoryTimeout)
	defer cancel()

	lmt := 640
	if limit > 0 && limit < lmt {
		lmt = limit
	}

	u, err := url.Parse(p.tencentBaseURL)
	if err != nil {
		return nil, "", err
	}
	q := u.Query()
	q.Set("param", fmt.Sprintf("%s,day,,,%d,qfq", asset.TencentCode, lmt))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "CycleCompass/4.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("tencent http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, "", err
	}

	var payload struct {
		Code int                        `json:"code"`
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, "", err
	}

	rawNode, ok := payload.Data[asset.TencentCode]
	if !ok {
		return nil, "", fmt.Errorf("tencent missing key %s", asset.TencentCode)
	}

	var seriesMap map[string]json.RawMessage
	if err := json.Unmarshal(rawNode, &seriesMap); err != nil {
		return nil, "", err
	}

	var rows [][]json.RawMessage
	for _, key := range []string{"qfqday", "day"} {
		if rawRows, exists := seriesMap[key]; exists {
			if err := json.Unmarshal(rawRows, &rows); err == nil && len(rows) > 0 {
				break
			}
		}
	}
	if len(rows) == 0 {
		return nil, "", fmt.Errorf("tencent empty rows")
	}

	source := "Tencent Daily Kline"
	rawBars := make([]DailyBar, 0, len(rows))
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		dateStr := parseRawString(row[0])
		closeVal, ok := parseRawFloat(row[2])
		if !ok {
			continue
		}

		bar := DailyBar{
			AssetID:       asset.AssetID,
			Market:        asset.Market,
			Symbol:        asset.Symbol,
			MeasureType:   asset.MeasureType,
			PriceType:     asset.PriceType,
			ReturnType:    asset.ReturnType,
			Unit:          asset.Unit,
			Currency:      asset.Currency,
			Timezone:      asset.Timezone,
			TradingDate:   dateStr,
			AsOf:          dateStr,
			SourceName:    source,
			QualityStatus: "ok",
			Open:          parseRawFloatPtr(row, 1),
			Close:         closeVal,
			High:          parseRawFloatPtr(row, 3),
			Low:           parseRawFloatPtr(row, 4),
			Volume:        parseRawFloatPtr(row, 5),
		}
		assignMeasureSpecificFields(&bar, asset)
		rawBars = append(rawBars, bar)
	}

	cleaned := p.sanitizeBars(asset, rawBars)
	return cleaned, source, nil
}

func (p *HistoricalProvider) loadLocalSnapshot(asset CatalogAsset) ([]DailyBar, string) {
	if p.snapshotDir != "" {
		jsonCandidates := []string{
			filepath.Join(p.snapshotDir, asset.AssetID+".json"),
			filepath.Join(p.snapshotDir, asset.Symbol+".json"),
		}
		for _, path := range jsonCandidates {
			if bars, ok := p.readJSONSnapshot(path, asset); ok {
				return bars, "Local JSON Snapshot"
			}
		}

		csvCandidates := []string{}
		if asset.LocalSnapshot != "" {
			csvCandidates = append(csvCandidates, filepath.Join(p.snapshotDir, asset.LocalSnapshot))
		}
		csvCandidates = append(csvCandidates,
			filepath.Join(p.snapshotDir, asset.AssetID+".csv"),
			filepath.Join(p.snapshotDir, asset.Symbol+".csv"),
		)
		for _, path := range csvCandidates {
			if bars, ok := p.readCSVSnapshot(path, asset); ok {
				return bars, "Local CSV Snapshot"
			}
		}
	}

	if p.seedFallback != nil && asset.LocalSnapshot != "" {
		pts := p.seedFallback.GetHistoricalKlines(asset.Symbol, 0)
		if len(pts) > 0 {
			bars := make([]DailyBar, 0, len(pts))
			for _, pt := range pts {
				bar := DailyBar{
					AssetID:       asset.AssetID,
					Market:        asset.Market,
					Symbol:        asset.Symbol,
					MeasureType:   asset.MeasureType,
					PriceType:     asset.PriceType,
					ReturnType:    asset.ReturnType,
					Unit:          asset.Unit,
					Currency:      asset.Currency,
					Timezone:      asset.Timezone,
					TradingDate:   pt.Date,
					AsOf:          pt.Date,
					SourceName:    "Offline Seed CSV",
					QualityStatus: "offline_seed",
					Close:         pt.Close,
				}
				assignMeasureSpecificFields(&bar, asset)
				bars = append(bars, bar)
			}
			return p.sanitizeBars(asset, bars), "Offline Seed CSV"
		}
	}

	return nil, ""
}

func (p *HistoricalProvider) readJSONSnapshot(path string, asset CatalogAsset) ([]DailyBar, bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var bars []DailyBar
	if err := json.Unmarshal(content, &bars); err != nil || len(bars) == 0 {
		var wrapper struct {
			Bars []DailyBar `json:"bars"`
		}
		if err2 := json.Unmarshal(content, &wrapper); err2 != nil || len(wrapper.Bars) == 0 {
			return nil, false
		}
		bars = wrapper.Bars
	}
	for i := range bars {
		if bars[i].AssetID == "" {
			bars[i].AssetID = asset.AssetID
		}
		if bars[i].Market == "" {
			bars[i].Market = asset.Market
		}
		if bars[i].Symbol == "" {
			bars[i].Symbol = asset.Symbol
		}
		if bars[i].MeasureType == "" {
			bars[i].MeasureType = asset.MeasureType
		}
		if bars[i].PriceType == "" {
			bars[i].PriceType = asset.PriceType
		}
		if bars[i].Unit == "" {
			bars[i].Unit = asset.Unit
		}
		if bars[i].Currency == "" {
			bars[i].Currency = asset.Currency
		}
		if bars[i].Timezone == "" {
			bars[i].Timezone = asset.Timezone
		}
		if bars[i].AsOf == "" {
			bars[i].AsOf = bars[i].TradingDate
		}
		bars[i].QualityStatus = "offline_seed"
		if bars[i].SourceName == "" {
			bars[i].SourceName = "Local JSON Snapshot"
		}
		assignMeasureSpecificFields(&bars[i], asset)
	}
	cleaned := p.sanitizeBars(asset, bars)
	return cleaned, len(cleaned) > 0
}

func (p *HistoricalProvider) readCSVSnapshot(path string, asset CatalogAsset) ([]DailyBar, bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))
	r := csv.NewReader(bytes.NewReader(content))
	records, err := r.ReadAll()
	if err != nil || len(records) < 2 {
		return nil, false
	}
	bars := make([]DailyBar, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < 2 {
			continue
		}
		dateStr := strings.TrimSpace(row[0])
		val, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if err != nil {
			continue
		}
		bar := DailyBar{
			AssetID:       asset.AssetID,
			Market:        asset.Market,
			Symbol:        asset.Symbol,
			MeasureType:   asset.MeasureType,
			PriceType:     asset.PriceType,
			ReturnType:    asset.ReturnType,
			Unit:          asset.Unit,
			Currency:      asset.Currency,
			Timezone:      asset.Timezone,
			TradingDate:   dateStr,
			AsOf:          dateStr,
			SourceName:    "Local CSV Snapshot",
			QualityStatus: "offline_seed",
			Close:         val,
		}
		assignMeasureSpecificFields(&bar, asset)
		bars = append(bars, bar)
	}
	cleaned := p.sanitizeBars(asset, bars)
	return cleaned, len(cleaned) > 0
}

// sanitizeBars 严格校验日期、未来数据泄露与价格有效性，按交易日升序去重。
func (p *HistoricalProvider) sanitizeBars(asset CatalogAsset, bars []DailyBar) []DailyBar {
	loc, err := time.LoadLocation(asset.Timezone)
	if err != nil {
		loc = time.UTC
	}
	nowInMarket := p.nowFunc().In(loc)
	todayStr := nowInMarket.Format("2006-01-02")

	byDate := make(map[string]DailyBar, len(bars))
	for _, b := range bars {
		dateStr := strings.TrimSpace(b.TradingDate)
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		// D-05: 过滤未来日期，禁止未来数据泄露
		if dateStr > todayStr {
			continue
		}
		// D-04 / D-06: 过滤 NaN / Inf，常规价格与指数必须严格 > 0
		if math.IsNaN(b.Close) || math.IsInf(b.Close, 0) {
			continue
		}
		if !asset.AllowNegative && asset.MeasureType != MeasureYieldPct && b.Close <= 0 {
			continue
		}

		if b.AsOf == "" {
			b.AsOf = dateStr
		}
		if b.AvailableAt == "" {
			avail := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 15, 30, 0, 0, loc)
			b.AvailableAt = avail.Format(time.RFC3339)
		}
		byDate[dateStr] = b
	}

	if len(byDate) == 0 {
		return nil
	}

	dates := make([]string, 0, len(byDate))
	for d := range byDate {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	out := make([]DailyBar, 0, len(dates))
	for _, d := range dates {
		out = append(out, byDate[d])
	}
	return out
}

// BuildCompositeIndex 按链式收益公式 I_t = I_{t-1} * (1 + sum(w_{i,t-1} * r_{i,t})) 计算综合指数（基期 100）。
// 当任意交易日有效权重覆盖率低于 minCoverage（如 0.90）或样本不足时，返回 nil 与明确原因。
func BuildCompositeIndex(barsByAsset map[string][]DailyBar, weights map[string]float64, minCoverage float64) ([]CompositeIndexPoint, string) {
	if len(weights) == 0 {
		return nil, "empty weights"
	}
	if minCoverage <= 0 || minCoverage > 1.0 {
		minCoverage = 0.90
	}

	totalWeight := 0.0
	for _, w := range weights {
		if w > 0 && !math.IsNaN(w) && !math.IsInf(w, 0) {
			totalWeight += w
		}
	}
	if totalWeight <= 0 {
		return nil, "total weight must be positive"
	}

	// 建立 assetID -> date -> close 索引，并收集全部有效日期
	priceMap := make(map[string]map[string]float64, len(weights))
	dateSet := make(map[string]struct{})
	for assetID, w := range weights {
		if w <= 0 {
			continue
		}
		bars := barsByAsset[assetID]
		m := make(map[string]float64, len(bars))
		for _, b := range bars {
			if b.Close > 0 && !math.IsNaN(b.Close) && !math.IsInf(b.Close, 0) && b.TradingDate != "" {
				m[b.TradingDate] = b.Close
				dateSet[b.TradingDate] = struct{}{}
			}
		}
		priceMap[assetID] = m
	}

	if len(dateSet) < 2 {
		return nil, "insufficient dates for composite index (need at least 2 trading dates)"
	}

	dates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	// 检查基期 t_0 的覆盖率
	baseDate := dates[0]
	baseCoveredWeight := 0.0
	baseCount := 0
	for assetID, w := range weights {
		if w <= 0 {
			continue
		}
		if c, ok := priceMap[assetID][baseDate]; ok && c > 0 {
			baseCoveredWeight += w
			baseCount++
		}
	}
	baseCoverage := baseCoveredWeight / totalWeight
	if baseCoverage+1e-9 < minCoverage {
		return nil, fmt.Sprintf("base date %s coverage %.4f below minimum %.4f", baseDate, baseCoverage, minCoverage)
	}

	points := make([]CompositeIndexPoint, 0, len(dates))
	points = append(points, CompositeIndexPoint{
		TradingDate:     baseDate,
		Value:           100.0,
		DailyReturn:     0.0,
		Coverage:        math.Round(baseCoverage*10000) / 10000,
		ValidAssetCount: baseCount,
	})

	prevIndex := 100.0
	for i := 1; i < len(dates); i++ {
		prevDate := dates[i-1]
		currDate := dates[i]

		coveredWeight := 0.0
		weightedRetSum := 0.0
		validCount := 0

		for assetID, w := range weights {
			if w <= 0 {
				continue
			}
			m := priceMap[assetID]
			prevClose, okPrev := m[prevDate]
			currClose, okCurr := m[currDate]
			if okPrev && okCurr && prevClose > 0 && currClose > 0 {
				ret := (currClose / prevClose) - 1.0
				coveredWeight += w
				weightedRetSum += w * ret
				validCount++
			}
		}

		coverage := coveredWeight / totalWeight
		if coverage+1e-9 < minCoverage {
			return nil, fmt.Sprintf("date %s coverage %.4f below minimum %.4f", currDate, coverage, minCoverage)
		}

		portfolioRet := weightedRetSum / coveredWeight
		currIndex := prevIndex * (1.0 + portfolioRet)
		points = append(points, CompositeIndexPoint{
			TradingDate:     currDate,
			Value:           currIndex,
			DailyReturn:     portfolioRet,
			Coverage:        math.Round(coverage*10000) / 10000,
			ValidAssetCount: validCount,
		})
		prevIndex = currIndex
	}

	return points, ""
}

func assignMeasureSpecificFields(bar *DailyBar, asset CatalogAsset) {
	c := bar.Close
	switch asset.MeasureType {
	case MeasureYieldPct:
		bar.YieldPct = &c
	case MeasurePrice:
		if asset.ReturnType == "total_return" {
			bar.TotalReturnIndex = &c
		} else if asset.PriceType == "qfq_close" {
			bar.AdjustedClose = &c
		}
	}
}

func parseOptionalFloat(parts []string, idx int) *float64 {
	if idx < 0 || idx >= len(parts) {
		return nil
	}
	s := strings.TrimSpace(parts[idx])
	if s == "" || s == "-" || strings.EqualFold(s, "null") {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}

func parseRawString(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	return ""
}

func parseRawFloat(raw json.RawMessage) (float64, bool) {
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return f, true
		}
		return 0, false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		s = strings.TrimSpace(s)
		if s == "" || s == "-" {
			return 0, false
		}
		v, err := strconv.ParseFloat(s, 64)
		if err == nil && !math.IsNaN(v) && !math.IsInf(v, 0) {
			return v, true
		}
	}
	return 0, false
}

func parseRawFloatPtr(row []json.RawMessage, idx int) *float64 {
	if idx < 0 || idx >= len(row) {
		return nil
	}
	v, ok := parseRawFloat(row[idx])
	if !ok {
		return nil
	}
	return &v
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
