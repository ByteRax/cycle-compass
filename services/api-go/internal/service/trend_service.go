package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/market"
	"cyclecompass/api/internal/storage"
	"cyclecompass/api/internal/trend"
)

var ErrNotFound = errors.New("not found")

const (
	defaultScopeID    = "all"
	defaultCurrency   = "ORIGINAL"
	defaultMode       = "latest_available"
	defaultRange      = "1Y"
	defaultFreq       = "daily"
	defaultLimit      = 50
	maxQueryLimit     = 200
	defaultChartColor = "#72A7D9"
)

var seriesColors = []string{
	"#E5B352",
	"#FF736A",
	"#2EC4B6",
	"#72A7D9",
	"#818CF8",
	"#34D399",
	"#F472B6",
	"#FBBF24",
}

// TrendService 提供 V4 趋势发现、快照计算与多维分析查询服务。
type TrendService struct {
	store *storage.SQLiteStore
	hist  *market.HistoricalProvider
}

// NewTrendService 构造 TrendService。
func NewTrendService(store *storage.SQLiteStore, hist *market.HistoricalProvider) *TrendService {
	return &TrendService{
		store: store,
		hist:  hist,
	}
}

type snapshotMetaExtra struct {
	ScopeLabel  string               `json:"scope_label"`
	SourceDates []dto.SourceDateMeta `json:"source_dates"`
	Warnings    []string             `json:"warnings"`
}

// BootstrapAndComputeSnapshot 同步目录、拉取/读取真实日线并原子发布不可变趋势快照。
// 当 fetchOnline=false 时不发起在线请求，仅使用本地 SQLite 中已有日线（无数据时诚实标记 insufficient）。
func (s *TrendService) BootstrapAndComputeSnapshot(ctx context.Context, fetchOnline bool) error {
	nowRFC3339 := time.Now().UTC().Format(time.RFC3339)

	// 1. 同步分类
	cats := make([]storage.AssetCategoryEntity, 0, len(market.V4Categories))
	for _, c := range market.V4Categories {
		cats = append(cats, storage.AssetCategoryEntity{
			CategoryID: c.ID,
			Name:       c.Name,
			Level:      1,
			SortOrder:  c.SortOrder,
			HasIndex:   false,
		})
	}
	if err := s.store.UpsertAssetCategories(cats); err != nil {
		return fmt.Errorf("sync categories: %w", err)
	}

	// 2. 同步资产目录与分类归属
	assetEntities := make([]storage.AssetEntity, 0, len(market.V4CatalogAssets))
	memberships := make([]storage.AssetMembershipEntity, 0, len(market.V4CatalogAssets))
	for _, a := range market.V4CatalogAssets {
		assetEntities = append(assetEntities, storage.AssetEntity{
			AssetID:       a.AssetID,
			Market:        a.Market,
			Symbol:        a.Symbol,
			Name:          a.Name,
			ShortName:     a.Name,
			AssetType:     a.AssetType,
			MeasureType:   string(a.MeasureType),
			PriceType:     a.PriceType,
			Unit:          a.Unit,
			Currency:      a.Currency,
			Timezone:      a.Timezone,
			IsActive:      true,
			LicenseStatus: string(a.LicenseStatus),
			SourceName:    a.Provider,
			CreatedAt:     nowRFC3339,
			UpdatedAt:     nowRFC3339,
		})
		if a.CategoryID != "" {
			memberships = append(memberships, storage.AssetMembershipEntity{
				CategoryID: a.CategoryID,
				AssetID:    a.AssetID,
				ValidFrom:  "1970-01-01",
			})
		}
	}
	if err := s.store.UpsertAssets(assetEntities); err != nil {
		return fmt.Errorf("sync assets: %w", err)
	}
	if err := s.store.UpsertAssetMemberships(memberships); err != nil {
		return fmt.Errorf("sync memberships: %w", err)
	}

	// 3. 同步观察池与成员
	universes := make([]storage.UniverseEntity, 0, len(market.V4Universes))
	var uniMembers []storage.UniverseMemberEntity
	for i, u := range market.V4Universes {
		universes = append(universes, storage.UniverseEntity{
			UniverseID:  u.ID,
			Name:        u.Name,
			Description: u.Description,
			IsPublic:    true,
			SortOrder:   i + 1,
		})
		for j, mID := range u.MemberAssetIDs {
			uniMembers = append(uniMembers, storage.UniverseMemberEntity{
				UniverseID: u.ID,
				AssetID:    mID,
				SortOrder:  j + 1,
			})
		}
	}
	if err := s.store.UpsertUniverses(universes); err != nil {
		return fmt.Errorf("sync universes: %w", err)
	}
	if err := s.store.UpsertUniverseMembers(uniMembers); err != nil {
		return fmt.Errorf("sync universe members: %w", err)
	}

	// 4. 加载/拉取真实历史日线（绝不使用 time.Now() 冒充行情日期）
	barsByAsset := make(map[string][]storage.DailyBarEntity, len(market.V4CatalogAssets))
	sourceByMarket := make(map[string]dto.SourceDateMeta)
	hasOnlineOK := false
	hasOfflineSeed := false
	candidateCount := 0
	validCount := 0
	maxAsOf := ""

	for _, asset := range market.V4CatalogAssets {
		if asset.LicenseStatus == market.LicenseNotCovered {
			continue
		}
		candidateCount++

		if fetchOnline && s.hist != nil {
			res := s.hist.FetchDailyHistory(ctx, asset.AssetID, 1000)
			if len(res.Bars) > 0 {
				dbBars := make([]storage.DailyBarEntity, 0, len(res.Bars))
				for _, b := range res.Bars {
					c := b.Close
					dbBars = append(dbBars, storage.DailyBarEntity{
						AssetID:          b.AssetID,
						TradingDate:      b.TradingDate,
						Provider:         firstNonEmpty(b.SourceName, asset.Provider),
						Revision:         1,
						Open:             b.Open,
						High:             b.High,
						Low:              b.Low,
						Close:            &c,
						AdjustedClose:    b.AdjustedClose,
						TotalReturnIndex: b.TotalReturnIndex,
						Volume:           b.Volume,
						Turnover:         b.Turnover,
						YieldPct:         b.YieldPct,
						QualityStatus:    b.QualityStatus,
						AsOf:             b.AsOf,
						AvailableAt:      b.AvailableAt,
						UpdatedAt:        nowRFC3339,
					})
				}
				if err := s.store.UpsertDailyBars(dbBars); err != nil {
					return fmt.Errorf("upsert bars for %s: %w", asset.AssetID, err)
				}
			}
		}

		storedBars, err := s.loadDedupedBars(asset.AssetID)
		if err != nil {
			return fmt.Errorf("load bars for %s: %w", asset.AssetID, err)
		}
		if len(storedBars) > 0 {
			barsByAsset[asset.AssetID] = storedBars
			validCount++
			lastBar := storedBars[len(storedBars)-1]
			if lastBar.AsOf > maxAsOf {
				maxAsOf = lastBar.AsOf
			}
			switch lastBar.QualityStatus {
			case "ok":
				hasOnlineOK = true
			case "offline_seed":
				hasOfflineSeed = true
			}
			prevMeta, exists := sourceByMarket[asset.Market]
			if !exists || lastBar.AsOf > prevMeta.AsOf {
				sourceByMarket[asset.Market] = dto.SourceDateMeta{
					Market: asset.Market,
					AsOf:   lastBar.AsOf,
					Source: lastBar.Provider,
					Status: lastBar.QualityStatus,
				}
			}
		}
	}

	// 5. 计算全量资产趋势指标、状态机、节气与事件
	cfg := trend.DefaultAlgorithmConfig()
	snapStamp := strings.ReplaceAll(maxAsOf, "-", "")
	if snapStamp == "" {
		snapStamp = "insufficient"
	}
	snapshotID := fmt.Sprintf("snap_%s_%d", snapStamp, time.Now().UTC().UnixNano())

	metricMap := make(map[string]storage.TrendDailyMetricEntity, len(market.V4CatalogAssets))
	trendBarsMap := make(map[string][]trend.DailyBar, len(barsByAsset))
	var allEvents []storage.TrendEventEntity

	for _, asset := range market.V4CatalogAssets {
		if asset.LicenseStatus == market.LicenseNotCovered {
			reason := firstNonEmpty(asset.LicenseNote, "not_covered")
			temp := string(trend.TempUnavailable)
			metricMap[asset.AssetID] = storage.TrendDailyMetricEntity{
				SnapshotID:   snapshotID,
				AssetID:      asset.AssetID,
				TradingDate:  "",
				Temperature:  &temp,
				Direction:    string(trend.DirectionUnknown),
				Confirmation: string(trend.ConfirmationUnknown),
				Side:         string(trend.SideUnknown),
				Reason:       &reason,
			}
			continue
		}

		rawBars := barsByAsset[asset.AssetID]
		if len(rawBars) == 0 {
			reason := "insufficient historical bars"
			temp := string(trend.TempUnavailable)
			metricMap[asset.AssetID] = storage.TrendDailyMetricEntity{
				SnapshotID:   snapshotID,
				AssetID:      asset.AssetID,
				TradingDate:  "",
				Temperature:  &temp,
				Direction:    string(trend.DirectionUnknown),
				Confirmation: string(trend.ConfirmationUnknown),
				Side:         string(trend.SideUnknown),
				Reason:       &reason,
			}
			continue
		}

		tBars := toTrendBars(rawBars)
		trendBarsMap[asset.AssetID] = tBars

		tempRes := trend.ComputeTemperature(tBars, cfg)
		states, evs := trend.ComputeStateSeries(tBars, cfg)
		lastState := trend.StateResult{
			TradingDate:  rawBars[len(rawBars)-1].TradingDate,
			Direction:    trend.DirectionUnknown,
			Confirmation: trend.ConfirmationUnknown,
			Side:         trend.SideUnknown,
		}
		if len(states) > 0 {
			lastState = states[len(states)-1]
		}
		phaseRes := trend.ComputePhase(lastState)

		tempStr := string(tempRes.Level)
		var tempScore *float64
		if tempRes.Score != nil {
			v := round2(*tempRes.Score)
			tempScore = &v
		}
		var startedAt, confirmedAt *string
		if lastState.StartedAt != "" {
			v := lastState.StartedAt
			startedAt = &v
		}
		if lastState.ConfirmedAt != "" {
			v := lastState.ConfirmedAt
			confirmedAt = &v
		}
		var durTrading, durCal *int
		if lastState.DurationTradingDays > 0 {
			d := lastState.DurationTradingDays
			durTrading = &d
		}
		if lastState.DurationCalendarDays > 0 {
			d := lastState.DurationCalendarDays
			durCal = &d
		}
		phaseName := phaseRes.PhaseName
		var reasonPtr *string
		if tempRes.Reason != "" {
			r := tempRes.Reason
			reasonPtr = &r
		} else if lastState.Reason != "" {
			r := lastState.Reason
			reasonPtr = &r
		}

		metricMap[asset.AssetID] = storage.TrendDailyMetricEntity{
			SnapshotID:           snapshotID,
			AssetID:              asset.AssetID,
			TradingDate:          rawBars[len(rawBars)-1].TradingDate,
			Temperature:          &tempStr,
			TemperatureScore:     tempScore,
			Direction:            string(lastState.Direction),
			Confirmation:         string(lastState.Confirmation),
			Side:                 string(lastState.Side),
			StartedAt:            startedAt,
			ConfirmedAt:          confirmedAt,
			DurationTradingDays:  durTrading,
			DurationCalendarDays: durCal,
			Phase:                &phaseName,
			Reason:               reasonPtr,
		}

		// 保留最近 20 条事件写入存储
		startEv := 0
		if len(evs) > 20 {
			startEv = len(evs) - 20
		}
		for _, ev := range evs[startEv:] {
			oldV := ev.OldValue
			newV := ev.NewValue
			allEvents = append(allEvents, storage.TrendEventEntity{
				EventID:          ev.EventID,
				AssetID:          ev.AssetID,
				EventType:        ev.EventType,
				OldValue:         &oldV,
				NewValue:         &newV,
				AsOf:             ev.AsOf,
				DetectedAt:       nowRFC3339,
				SnapshotID:       snapshotID,
				AlgorithmVersion: cfg.AlgorithmVersion,
				ScopeID:          defaultScopeID,
				DedupeKey:        ev.DedupeKey + ":" + snapshotID,
			})
		}
	}

	// 6. 计算全市场 ("all") 与各公开观察池的相对强度排名
	var allRankings []storage.StrengthDailyRankingEntity
	computeScopeRankings := func(scopeID, scopeLabel string, assetIDs []string, updateTier bool) {
		inputs := make([]trend.StrengthInput, 0, len(assetIDs))
		for _, id := range assetIDs {
			if tBars, ok := trendBarsMap[id]; ok && len(tBars) > 0 {
				inputs = append(inputs, trend.StrengthInput{AssetID: id, Bars: tBars})
			}
		}
		if len(inputs) < 2 {
			return
		}
		results := trend.ComputeRelativeStrength(inputs, scopeID, scopeLabel, defaultCurrency, cfg.AlgorithmVersion)
		for _, r := range results {
			var rankInt *int
			if r.Rank != nil {
				v := int(math.Round(*r.Rank))
				rankInt = &v
			}
			var scorePtr *float64
			if r.Score != nil {
				v := round2(*r.Score)
				scorePtr = &v
			}
			var rawPtr *float64
			if r.RawFactor != nil {
				v := round4(*r.RawFactor)
				rawPtr = &v
			}
			var reasonPtr *string
			if r.Reason != "" {
				v := r.Reason
				reasonPtr = &v
			}
			allRankings = append(allRankings, storage.StrengthDailyRankingEntity{
				SnapshotID:     snapshotID,
				ScopeID:        scopeID,
				Currency:       defaultCurrency,
				AssetID:        r.AssetID,
				RawFactor:      rawPtr,
				Rank:           rankInt,
				Score:          scorePtr,
				CandidateCount: r.CandidateCount,
				ValidCount:     r.ValidCount,
				Reason:         reasonPtr,
			})

			if updateTier {
				if m, ok := metricMap[r.AssetID]; ok {
					tierPtr, _ := trend.ComputeTier(r.Score, trend.Side(m.Side))
					if tierPtr != nil {
						ts := string(*tierPtr)
						m.Tier = &ts
						metricMap[r.AssetID] = m
					}
				}
			}
		}
	}

	allCoveredIDs := make([]string, 0, len(trendBarsMap))
	for id := range trendBarsMap {
		allCoveredIDs = append(allCoveredIDs, id)
	}
	sort.Strings(allCoveredIDs)
	computeScopeRankings(defaultScopeID, "全市场观察池", allCoveredIDs, true)
	for _, u := range market.V4Universes {
		computeScopeRankings(u.ID, u.Name, u.MemberAssetIDs, false)
	}

	// 7. 组装快照状态与覆盖率
	status := "ok"
	var warnings []string
	if validCount == 0 {
		status = "insufficient"
		warnings = append(warnings, "当前尚无有效日线数据，所有趋势指标诚实标记为 insufficient")
	} else if !hasOnlineOK && hasOfflineSeed {
		status = "offline_seed"
		warnings = append(warnings, "当前使用本地离线快照数据")
	} else if validCount < candidateCount {
		status = "degraded"
		warnings = append(warnings, fmt.Sprintf("部分标的缺少历史日线（有效 %d/%d）", validCount, candidateCount))
	}

	var covPtr *float64
	if candidateCount > 0 {
		cov := round4(float64(validCount) / float64(candidateCount))
		covPtr = &cov
	}

	sourceDates := make([]dto.SourceDateMeta, 0, len(sourceByMarket))
	for _, sd := range sourceByMarket {
		sourceDates = append(sourceDates, sd)
	}
	sort.Slice(sourceDates, func(i, j int) bool {
		return sourceDates[i].Market < sourceDates[j].Market
	})

	metaBytes, _ := json.Marshal(snapshotMetaExtra{
		ScopeLabel:  "全市场观察池",
		SourceDates: sourceDates,
		Warnings:    warnings,
	})

	metricsSlice := make([]storage.TrendDailyMetricEntity, 0, len(metricMap))
	for _, m := range metricMap {
		metricsSlice = append(metricsSlice, m)
	}

	snapshot := storage.TrendSnapshotEntity{
		SnapshotID:       snapshotID,
		ScopeID:          defaultScopeID,
		Currency:         defaultCurrency,
		Mode:             defaultMode,
		AsOf:             maxAsOf,
		ComputedAt:       nowRFC3339,
		AlgorithmVersion: cfg.AlgorithmVersion,
		CandidateCount:   candidateCount,
		ValidCount:       validCount,
		Coverage:         covPtr,
		Status:           status,
		IsCurrent:        true,
		MetaJSON:         string(metaBytes),
	}

	if err := s.store.PublishTrendSnapshot(snapshot, metricsSlice, allRankings, allEvents); err != nil {
		return fmt.Errorf("publish snapshot: %w", err)
	}
	return nil
}

// GetCatalog 返回资产目录、观察池、分类树与全市场趋势概览。
func (s *TrendService) GetCatalog(ctx context.Context, scopeID, currency string) (*dto.CatalogResponse, error) {
	snap, meta, err := s.resolveSnapshot("", scopeID, currency, defaultMode)
	if err != nil {
		return nil, err
	}

	universes, err := s.store.ListUniverses(true)
	if err != nil {
		return nil, err
	}
	uniItems := make([]dto.CatalogUniverseItem, 0, len(universes))
	for _, u := range universes {
		members, _ := s.store.ListUniverseMembers(u.UniverseID)
		uniItems = append(uniItems, dto.CatalogUniverseItem{
			UniverseID:  u.UniverseID,
			Name:        u.Name,
			Description: u.Description,
			MemberCount: len(members),
		})
	}

	cats, err := s.store.ListAssetCategories()
	if err != nil {
		return nil, err
	}
	catItems := make([]dto.CatalogCategoryItem, 0, len(cats))
	for _, c := range cats {
		catItems = append(catItems, dto.CatalogCategoryItem{
			CategoryID:   c.CategoryID,
			ParentID:     c.ParentID,
			Name:         c.Name,
			Level:        c.Level,
			HasIndex:     c.HasIndex,
			IndexAssetID: c.IndexAssetID,
		})
	}

	overview := dto.CatalogOverview{}
	if snap != nil {
		metrics, err := s.store.ListTrendDailyMetrics(snap.SnapshotID, nil)
		if err == nil {
			overview.TotalCount = len(metrics)
			for _, m := range metrics {
				switch m.Side {
				case string(trend.SideRight):
					overview.RightCount++
					overview.ValidCount++
				case string(trend.SideLeft):
					overview.LeftCount++
					overview.ValidCount++
				case string(trend.SideNeutral):
					overview.OtherCount++
					overview.ValidCount++
				}
			}
			if overview.ValidCount > 0 {
				rr := round2(float64(overview.RightCount) * 100 / float64(overview.ValidCount))
				lr := round2(float64(overview.LeftCount) * 100 / float64(overview.ValidCount))
				or := round2(float64(overview.OtherCount) * 100 / float64(overview.ValidCount))
				overview.RightRatioPct = &rr
				overview.LeftRatioPct = &lr
				overview.OtherRatioPct = &or
			}
		}
	}

	return &dto.CatalogResponse{
		Snapshot:   meta,
		Universes:  uniItems,
		Categories: catItems,
		Overview:   overview,
	}, nil
}

// GetUniverses 返回所有公开观察池列表。
func (s *TrendService) GetUniverses(ctx context.Context) (*dto.UniversesResponse, error) {
	_, meta, err := s.resolveSnapshot("", defaultScopeID, defaultCurrency, defaultMode)
	if err != nil {
		return nil, err
	}
	universes, err := s.store.ListUniverses(true)
	if err != nil {
		return nil, err
	}
	items := make([]dto.CatalogUniverseItem, 0, len(universes))
	for _, u := range universes {
		members, _ := s.store.ListUniverseMembers(u.UniverseID)
		items = append(items, dto.CatalogUniverseItem{
			UniverseID:  u.UniverseID,
			Name:        u.Name,
			Description: u.Description,
			MemberCount: len(members),
		})
	}
	return &dto.UniversesResponse{
		Items:    items,
		Snapshot: meta,
	}, nil
}

// SearchAssets 按关键词、资产类型检索标准化资产目录。
func (s *TrendService) SearchAssets(ctx context.Context, q, assetType, cursor string, limit int) (*dto.SearchAssetsResponse, error) {
	snap, _, err := s.resolveSnapshot("", defaultScopeID, defaultCurrency, defaultMode)
	if err != nil {
		return nil, err
	}
	assets, err := s.store.ListAssets(true)
	if err != nil {
		return nil, err
	}

	qLower := strings.ToLower(strings.TrimSpace(q))
	typeLower := strings.ToLower(strings.TrimSpace(assetType))

	var matched []storage.AssetEntity
	for _, a := range assets {
		if typeLower != "" && strings.ToLower(a.AssetType) != typeLower {
			continue
		}
		if qLower != "" {
			hit := strings.Contains(strings.ToLower(a.AssetID), qLower) ||
				strings.Contains(strings.ToLower(a.Symbol), qLower) ||
				strings.Contains(strings.ToLower(a.Name), qLower)
			if !hit {
				if catAsset, ok := market.GetCatalogAsset(a.AssetID); ok {
					for _, alias := range catAsset.Aliases {
						if strings.Contains(strings.ToLower(alias), qLower) {
							hit = true
							break
						}
					}
				}
			}
			if !hit {
				continue
			}
		}
		matched = append(matched, a)
	}

	total := len(matched)
	offset := parseCursor(cursor)
	if limit <= 0 {
		limit = defaultLimit
	}
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	page := matched[offset:end]
	items, err := s.buildAssetSummaries(snap, defaultScopeID, page)
	if err != nil {
		return nil, err
	}

	var nextCursor *string
	if end < total {
		nc := strconv.Itoa(end)
		nextCursor = &nc
	}

	return &dto.SearchAssetsResponse{
		Items:      items,
		NextCursor: nextCursor,
		Total:      total,
	}, nil
}

// GetAssetDetail 返回单个资产（或兼容旧 ID）详情、可用图表、周期涨跌与最近事件。
func (s *TrendService) GetAssetDetail(ctx context.Context, id, scopeID, currency, snapshotID string) (*dto.AssetDetailResponse, error) {
	target := s.resolveTarget(id)
	if target.Asset == nil && target.Universe == nil && target.Category == nil {
		return nil, ErrNotFound
	}

	effScope := firstNonEmpty(scopeID, target.ScopeID, defaultScopeID)
	snap, meta, err := s.resolveSnapshot(snapshotID, effScope, currency, defaultMode)
	if err != nil {
		return nil, err
	}

	var summary dto.AssetSummaryItem
	var breadcrumbs []dto.BreadcrumbItem
	var availableCharts []string
	var benchmarkID, benchmarkName *string
	var month1Pct, year1Pct *float64
	var recentEvents []dto.TrendEventItem

	if target.Asset != nil {
		items, err := s.buildAssetSummaries(snap, effScope, []storage.AssetEntity{*target.Asset})
		if err != nil {
			return nil, err
		}
		summary = items[0]
		availableCharts = []string{"curve", "annual"}
		if target.Universe != nil {
			availableCharts = append(availableCharts, "components", "dashboard", "season", "distribution", "tier")
			breadcrumbs = append(breadcrumbs, dto.BreadcrumbItem{ID: target.Universe.ID, Name: target.Universe.Name})
			if target.Universe.BenchmarkAssetID != "" {
				bID := target.Universe.BenchmarkAssetID
				benchmarkID = &bID
				if ba, ok := market.GetCatalogAsset(bID); ok {
					bn := ba.Name
					benchmarkName = &bn
				}
			}
		}
		breadcrumbs = append(breadcrumbs, dto.BreadcrumbItem{ID: summary.AssetID, Name: summary.Name})

		bars, _ := s.loadDedupedBars(summary.AssetID)
		month1Pct = computeTrailingReturnPct(bars, 21)
		year1Pct = computeTrailingReturnPct(bars, 252)

		if snap != nil {
			evs, _ := s.store.ListTrendEvents(snap.SnapshotID, summary.AssetID, nil, 10, 0)
			recentEvents = mapEventEntities(evs)
		}
	} else if target.Universe != nil {
		summary = s.buildUniverseSummary(snap, *target.Universe)
		availableCharts = []string{"components", "dashboard", "season", "distribution", "tier"}
		breadcrumbs = []dto.BreadcrumbItem{{ID: target.Universe.ID, Name: target.Universe.Name}}
		if target.Universe.BenchmarkAssetID != "" {
			bID := target.Universe.BenchmarkAssetID
			benchmarkID = &bID
			if ba, ok := market.GetCatalogAsset(bID); ok {
				bn := ba.Name
				benchmarkName = &bn
			}
		}
	} else if target.Category != nil {
		summary = dto.AssetSummaryItem{
			AssetID:       target.Category.ID,
			Market:        "CN",
			Symbol:        target.Category.ID,
			Name:          target.Category.Name,
			ShortName:     target.Category.Name,
			AssetType:     "category",
			MeasureType:   "index",
			HasMembers:    true,
			QualityStatus: meta.Status,
			Metrics:       emptyMetrics(),
			FeatureTags:   []string{},
		}
		availableCharts = []string{"components", "dashboard", "season", "distribution", "tier"}
		breadcrumbs = []dto.BreadcrumbItem{{ID: target.Category.ID, Name: target.Category.Name}}
	}

	if recentEvents == nil {
		recentEvents = []dto.TrendEventItem{}
	}

	return &dto.AssetDetailResponse{
		Snapshot:            meta,
		Asset:               summary,
		Breadcrumbs:         breadcrumbs,
		AvailableCharts:     availableCharts,
		AvailableRanges:     []string{"1M", "3M", "6M", "1Y", "3Y", "5Y", "ALL"},
		AvailableCurrencies: []string{"ORIGINAL", "CNY", "USD"},
		BenchmarkID:         benchmarkID,
		BenchmarkName:       benchmarkName,
		PeriodChanges: dto.PeriodChanges{
			Month1Pct: month1Pct,
			Year1Pct:  year1Pct,
		},
		RecentEvents: recentEvents,
	}, nil
}

// GetAssetMembers 返回观察池/分类成员列表，支持排序白名单与游标分页。
func (s *TrendService) GetAssetMembers(ctx context.Context, id, scopeID, sortField, order, cursor string, limit int, currency string) (*dto.MembersResponse, error) {
	parentID, parentName, members, err := s.resolvePoolMembers(id)
	if err != nil {
		return nil, err
	}

	effScope := firstNonEmpty(scopeID, parentID, defaultScopeID)
	snap, meta, err := s.resolveSnapshot("", effScope, currency, defaultMode)
	if err != nil {
		return nil, err
	}

	items, err := s.buildAssetSummaries(snap, effScope, members)
	if err != nil {
		return nil, err
	}

	sortAssetSummaries(items, sortField, order)

	total := len(items)
	offset := parseCursor(cursor)
	if limit <= 0 {
		limit = defaultLimit
	}
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}

	var nextCursor *string
	if end < total {
		nc := strconv.Itoa(end)
		nextCursor = &nc
	}

	return &dto.MembersResponse{
		Snapshot:   meta,
		ParentID:   parentID,
		ParentName: parentName,
		Items:      items[offset:end],
		NextCursor: nextCursor,
		Total:      total,
	}, nil
}

// GetLatestSnapshot 返回当前生效快照元数据。
func (s *TrendService) GetLatestSnapshot(ctx context.Context, scopeID, currency, mode string) (*dto.SnapshotMeta, error) {
	_, meta, err := s.resolveSnapshot("", scopeID, currency, mode)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

// GetTrendCurve 返回单资产趋势曲线、极值（最大回撤/最大涨幅）、状态机区间与事件。
func (s *TrendService) GetTrendCurve(ctx context.Context, id, rangeStr, freq, currency, benchmarkID, snapshotID string) (*dto.TrendCurveResponse, error) {
	target := s.resolveTarget(id)
	if target.Asset == nil {
		return nil, ErrNotFound
	}
	snap, meta, err := s.resolveSnapshot(snapshotID, firstNonEmpty(target.ScopeID, defaultScopeID), currency, defaultMode)
	if err != nil {
		return nil, err
	}

	if rangeStr == "" {
		rangeStr = defaultRange
	}
	if freq == "" {
		freq = defaultFreq
	}

	allBars, err := s.loadDedupedBars(target.Asset.AssetID)
	if err != nil {
		return nil, err
	}
	rangedBars := filterBarsByRange(allBars, rangeStr)
	points := buildSeriesPoints(rangedBars)

	effBenchID := benchmarkID
	if effBenchID == "" && target.Universe != nil && target.Universe.BenchmarkAssetID != "" && target.Universe.BenchmarkAssetID != target.Asset.AssetID {
		effBenchID = target.Universe.BenchmarkAssetID
	}
	var benchIDPtr, benchNamePtr *string
	benchPoints := []dto.SeriesPoint{}
	if effBenchID != "" {
		if ba, ok := market.GetCatalogAsset(effBenchID); ok {
			bID := ba.AssetID
			bName := ba.Name
			benchIDPtr = &bID
			benchNamePtr = &bName
			bBars, _ := s.loadDedupedBars(bID)
			benchPoints = buildSeriesPoints(filterBarsByRange(bBars, rangeStr))
		}
	}

	tBars := toTrendBars(rangedBars)
	ext := trend.ComputeExtremum(tBars)
	extInfo := buildExtremumInfo(ext, rangedBars)

	_, allEvents := trend.ComputeStateSeries(toTrendBars(allBars), trend.DefaultAlgorithmConfig())
	rangedStates, _ := trend.ComputeStateSeries(tBars, trend.DefaultAlgorithmConfig())
	intervals := buildStateIntervals(rangedStates)

	var curveEvents []dto.TrendEventItem
	minDate := ""
	if len(rangedBars) > 0 {
		minDate = rangedBars[0].TradingDate
	}
	for _, ev := range allEvents {
		if minDate != "" && ev.AsOf < minDate {
			continue
		}
		curveEvents = append(curveEvents, dto.TrendEventItem{
			EventID:          ev.EventID,
			AssetID:          ev.AssetID,
			EventType:        ev.EventType,
			OldValue:         ev.OldValue,
			NewValue:         ev.NewValue,
			AsOf:             ev.AsOf,
			DetectedAt:       meta.ComputedAt,
			SnapshotID:       meta.SnapshotID,
			AlgorithmVersion: ev.AlgorithmVersion,
			ScopeID:          firstNonEmpty(target.ScopeID, defaultScopeID),
		})
	}
	if curveEvents == nil {
		curveEvents = []dto.TrendEventItem{}
	}
	if snap != nil && len(curveEvents) == 0 {
		if evs, err := s.store.ListTrendEvents(snap.SnapshotID, target.Asset.AssetID, nil, 20, 0); err == nil {
			curveEvents = mapEventEntities(evs)
		}
	}

	return &dto.TrendCurveResponse{
		Snapshot:        meta,
		AssetID:         target.Asset.AssetID,
		AssetName:       target.Asset.Name,
		BenchmarkID:     benchIDPtr,
		BenchmarkName:   benchNamePtr,
		Range:           rangeStr,
		Freq:            freq,
		Points:          points,
		BenchmarkPoints: benchPoints,
		Extremum:        extInfo,
		StateIntervals:  intervals,
		Events:          curveEvents,
	}, nil
}

// GetTrendAnnual 返回多年度自然周同比收益与统计摘要。
func (s *TrendService) GetTrendAnnual(ctx context.Context, id, yearsStr, currency, snapshotID string) (*dto.AnnualTrendResponse, error) {
	target := s.resolveTarget(id)
	if target.Asset == nil {
		return nil, ErrNotFound
	}
	_, meta, err := s.resolveSnapshot(snapshotID, firstNonEmpty(target.ScopeID, defaultScopeID), currency, defaultMode)
	if err != nil {
		return nil, err
	}

	allBars, err := s.loadDedupedBars(target.Asset.AssetID)
	if err != nil {
		return nil, err
	}

	asOf := meta.AsOf
	if len(allBars) > 0 {
		asOf = allBars[len(allBars)-1].TradingDate
	}
	annualRes := trend.ComputeAnnualComparison(toTrendBars(allBars), asOf)

	yearFilter := parseYearFilter(yearsStr)
	yearsOut := make([]dto.AnnualYearSeries, 0, len(annualRes.Years))
	for _, ys := range annualRes.Years {
		if len(yearFilter) > 0 && !yearFilter[ys.Year] {
			continue
		}
		weeks := make([]dto.AnnualWeekPoint, 0, len(ys.Weeks))
		for _, w := range ys.Weeks {
			pct := round2(w.Return * 100)
			weeks = append(weeks, dto.AnnualWeekPoint{
				Week:      w.Week,
				EndDate:   w.WeekEndDate,
				ReturnPct: &pct,
			})
		}
		yearsOut = append(yearsOut, dto.AnnualYearSeries{
			Year:              ys.Year,
			IsCurrent:         ys.Year == annualRes.CurrentYear,
			IsComplete:        ys.IsComplete,
			SameWeekReturnPct: fracPtrToPctPtr(ys.SameWeekReturn),
			FullYearReturnPct: fracPtrToPctPtr(ys.FullYearReturn),
			WeeklyReturns:     weeks,
		})
	}

	var curRank *int
	if annualRes.SameWeekPercentile != nil && len(annualRes.Years) > 0 {
		rank := int(math.Round((1.0 - *annualRes.SameWeekPercentile) * float64(len(annualRes.Years))))
		if rank < 1 {
			rank = 1
		}
		curRank = &rank
	}

	return &dto.AnnualTrendResponse{
		Snapshot:   meta,
		AssetID:    target.Asset.AssetID,
		AssetName:  target.Asset.Name,
		CutoffWeek: annualRes.CurrentWeek,
		Years:      yearsOut,
		Stats: dto.AnnualStats{
			SampleCount:          annualRes.CompleteYearCount,
			CurrentYearReturnPct: fracPtrToPctPtr(annualRes.CurrentReturn),
			SameWeekMedianPct:    fracPtrToPctPtr(annualRes.SameWeekMedian),
			SameWeekMeanPct:      fracPtrToPctPtr(annualRes.SameWeekMean),
			CurrentRank:          curRank,
			UpRatioPct:           fracPtrToPctPtr(annualRes.SameWeekUpRatio),
		},
	}, nil
}

// GetTrendComponents 返回观察池成分标准化对比曲线。
func (s *TrendService) GetTrendComponents(ctx context.Context, id, scopeID, rangeStr string, topN int, sortField, currency, snapshotID string) (*dto.ComponentTrendResponse, error) {
	parentID, parentName, members, err := s.resolvePoolMembers(id)
	if err != nil {
		return nil, err
	}
	if rangeStr == "" {
		rangeStr = defaultRange
	}
	effScope := firstNonEmpty(scopeID, parentID, defaultScopeID)
	snap, meta, err := s.resolveSnapshot(snapshotID, effScope, currency, defaultMode)
	if err != nil {
		return nil, err
	}

	summaries, err := s.buildAssetSummaries(snap, effScope, members)
	if err != nil {
		return nil, err
	}
	if sortField != "" {
		sortAssetSummaries(summaries, sortField, "desc")
	}
	if topN > 0 && len(summaries) > topN {
		summaries = summaries[:topN]
	}

	// 收集所有成员区间内的共同/并集时间轴
	dateSet := make(map[string]struct{})
	rangedByAsset := make(map[string][]storage.DailyBarEntity, len(summaries))
	for _, sm := range summaries {
		bars, _ := s.loadDedupedBars(sm.AssetID)
		rb := filterBarsByRange(bars, rangeStr)
		rangedByAsset[sm.AssetID] = rb
		for _, b := range rb {
			dateSet[b.TradingDate] = struct{}{}
		}
	}
	timeline := make([]string, 0, len(dateSet))
	for d := range dateSet {
		timeline = append(timeline, d)
	}
	sort.Strings(timeline)

	benchID := ""
	if u, ok := market.GetUniverse(parentID); ok {
		benchID = u.BenchmarkAssetID
	}

	series := make([]dto.ComponentTrendSeries, 0, len(summaries))
	for i, sm := range summaries {
		rb := rangedByAsset[sm.AssetID]
		color := seriesColors[i%len(seriesColors)]
		if len(rb) == 0 {
			series = append(series, dto.ComponentTrendSeries{
				AssetID:     sm.AssetID,
				Name:        sm.Name,
				Color:       color,
				IsBenchmark: sm.AssetID == benchID,
				Temperature: sm.Metrics.Temperature,
				LatestValue: sm.LatestValue,
				Turnover:    sm.Turnover,
				Points:      make([]*float64, len(timeline)),
				Reason:      sm.Metrics.Reason,
			})
			continue
		}

		valByDate := make(map[string]float64, len(rb))
		var baseVal float64
		for _, b := range rb {
			if b.Close != nil && *b.Close > 0 {
				valByDate[b.TradingDate] = *b.Close
				if baseVal == 0 {
					baseVal = *b.Close
				}
			}
		}

		pts := make([]*float64, len(timeline))
		for ti, d := range timeline {
			if v, ok := valByDate[d]; ok && baseVal > 0 {
				norm := round2((v/baseVal - 1.0) * 100)
				pts[ti] = &norm
			}
		}

		var periodPct *float64
		if baseVal > 0 && rb[len(rb)-1].Close != nil {
			p := round2((*rb[len(rb)-1].Close/baseVal - 1.0) * 100)
			periodPct = &p
		}

		series = append(series, dto.ComponentTrendSeries{
			AssetID:         sm.AssetID,
			Name:            sm.Name,
			Color:           color,
			IsBenchmark:     sm.AssetID == benchID,
			Temperature:     sm.Metrics.Temperature,
			LatestValue:     sm.LatestValue,
			PeriodChangePct: periodPct,
			Turnover:        sm.Turnover,
			TotalReturnPct:  periodPct,
			Points:          pts,
			Reason:          sm.Metrics.Reason,
		})
	}

	return &dto.ComponentTrendResponse{
		Snapshot:  meta,
		AssetID:   parentID,
		AssetName: parentName,
		Range:     rangeStr,
		Timeline:  timeline,
		Series:    series,
	}, nil
}

// GetTrendDashboard 返回分组看板与迷你走势 sparkline。
func (s *TrendService) GetTrendDashboard(ctx context.Context, id, changePeriod, sparkRange, sortField, order, scopeID, group, currency, snapshotID string) (*dto.TrendDashboardResponse, error) {
	parentID, parentName, members, err := s.resolvePoolMembers(id)
	if err != nil {
		return nil, err
	}
	if changePeriod == "" {
		changePeriod = "1D"
	}
	if sparkRange == "" {
		sparkRange = "1M"
	}
	effScope := firstNonEmpty(scopeID, parentID, defaultScopeID)
	snap, meta, err := s.resolveSnapshot(snapshotID, effScope, currency, defaultMode)
	if err != nil {
		return nil, err
	}

	items, err := s.buildAssetSummaries(snap, effScope, members)
	if err != nil {
		return nil, err
	}
	if sortField != "" {
		sortAssetSummaries(items, sortField, order)
	}

	groupMap := make(map[string][]dto.DashboardRow)
	var groupNames []string
	for _, it := range items {
		gName := firstNonEmpty(it.GroupName, it.Market, "其他")
		if group != "" && gName != group {
			continue
		}
		bars, _ := s.loadDedupedBars(it.AssetID)
		rb := filterBarsByRange(bars, sparkRange)
		spark, sMin, sMax := buildSparkline(rb, 24)
		chg := computePeriodChangeByCode(bars, changePeriod, it.DayChangePct)

		if _, ok := groupMap[gName]; !ok {
			groupNames = append(groupNames, gName)
		}
		groupMap[gName] = append(groupMap[gName], dto.DashboardRow{
			Asset:     it,
			ChangePct: chg,
			Sparkline: spark,
			SparkMin:  sMin,
			SparkMax:  sMax,
		})
	}

	groups := make([]dto.DashboardGroup, 0, len(groupNames))
	for _, gn := range groupNames {
		groups = append(groups, dto.DashboardGroup{
			GroupName: gn,
			Rows:      groupMap[gn],
		})
	}

	return &dto.TrendDashboardResponse{
		Snapshot:     meta,
		AssetID:      parentID,
		AssetName:    parentName,
		ChangePeriod: changePeriod,
		SparkRange:   sparkRange,
		Groups:       groups,
	}, nil
}

// GetTrendSeason 返回趋势节气生命周期分布。
func (s *TrendService) GetTrendSeason(ctx context.Context, id, direction, snapshotID string) (*dto.TrendSeasonResponse, error) {
	parentID, parentName, members, err := s.resolvePoolMembers(id)
	if err != nil {
		return nil, err
	}
	snap, meta, err := s.resolveSnapshot(snapshotID, parentID, defaultCurrency, defaultMode)
	if err != nil {
		return nil, err
	}
	items, err := s.buildAssetSummaries(snap, parentID, members)
	if err != nil {
		return nil, err
	}

	stages := []dto.SeasonStageBucket{
		{StageKey: "guyu", StageLabel: "谷雨", Direction: "up", RuleDesc: "上行形成中（未确认）", Members: []dto.SeasonMemberItem{}},
		{StageKey: "lixia", StageLabel: "立夏", Direction: "up", RuleDesc: "已确认上行 1–10 交易日", Members: []dto.SeasonMemberItem{}},
		{StageKey: "xiazhi", StageLabel: "夏至", Direction: "up", RuleDesc: "已确认上行 11–40 交易日", Members: []dto.SeasonMemberItem{}},
		{StageKey: "xiaoshu", StageLabel: "小暑", Direction: "up", RuleDesc: "已确认上行 41–90 交易日", Members: []dto.SeasonMemberItem{}},
		{StageKey: "dashu", StageLabel: "大暑", Direction: "up", RuleDesc: "已确认上行 >90 交易日", Members: []dto.SeasonMemberItem{}},
		{StageKey: "oscillation", StageLabel: "震荡", Direction: "neutral", RuleDesc: "趋势中性震荡", Members: []dto.SeasonMemberItem{}},
		{StageKey: "hanlu", StageLabel: "寒露", Direction: "down", RuleDesc: "下行形成中（未确认）", Members: []dto.SeasonMemberItem{}},
		{StageKey: "lidong", StageLabel: "立冬", Direction: "down", RuleDesc: "已确认下行 1–10 交易日", Members: []dto.SeasonMemberItem{}},
		{StageKey: "dongzhi", StageLabel: "冬至", Direction: "down", RuleDesc: "已确认下行 11–40 交易日", Members: []dto.SeasonMemberItem{}},
		{StageKey: "xiaohan", StageLabel: "小寒", Direction: "down", RuleDesc: "已确认下行 41–90 交易日", Members: []dto.SeasonMemberItem{}},
		{StageKey: "dahan", StageLabel: "大寒", Direction: "down", RuleDesc: "已确认下行 >90 交易日", Members: []dto.SeasonMemberItem{}},
	}
	stageIdx := make(map[string]int, len(stages))
	for i, st := range stages {
		stageIdx[st.StageLabel] = i
	}

	var upCount, downCount, otherCount, validCount int
	for _, it := range items {
		dir := it.Metrics.Direction
		if dir == string(trend.DirectionUp) {
			upCount++
			validCount++
		} else if dir == string(trend.DirectionDown) {
			downCount++
			validCount++
		} else if dir == string(trend.DirectionNeutral) {
			otherCount++
			validCount++
		}

		if direction == "up" && dir != "up" {
			continue
		}
		if direction == "down" && dir != "down" {
			continue
		}

		if it.Metrics.Phase != nil {
			if idx, ok := stageIdx[*it.Metrics.Phase]; ok {
				stages[idx].Members = append(stages[idx].Members, dto.SeasonMemberItem{
					AssetID:             it.AssetID,
					Name:                it.Name,
					DurationTradingDays: it.Metrics.DurationTradingDays,
					Temperature:         it.Metrics.Temperature,
					Side:                it.Metrics.Side,
				})
				stages[idx].Count++
			}
		}
	}

	return &dto.TrendSeasonResponse{
		Snapshot:   meta,
		AssetID:    parentID,
		AssetName:  parentName,
		UpCount:    upCount,
		DownCount:  downCount,
		OtherCount: otherCount,
		ValidCount: validCount,
		Stages:     stages,
	}, nil
}

// GetTrendDistribution 返回资金权重与右侧占比分布矩阵。
func (s *TrendService) GetTrendDistribution(ctx context.Context, id, weightBy, colorBy, snapshotID string) (*dto.TrendDistributionResponse, error) {
	parentID, parentName, members, err := s.resolvePoolMembers(id)
	if err != nil {
		return nil, err
	}
	if weightBy == "" {
		weightBy = "equal"
	}
	if colorBy == "" {
		colorBy = "side"
	}
	snap, meta, err := s.resolveSnapshot(snapshotID, parentID, defaultCurrency, defaultMode)
	if err != nil {
		return nil, err
	}
	items, err := s.buildAssetSummaries(snap, parentID, members)
	if err != nil {
		return nil, err
	}

	type rawNode struct {
		item   dto.AssetSummaryItem
		weight float64
		hasW   bool
	}
	rawNodes := make([]rawNode, 0, len(items))
	totalWeight := 0.0
	weightedValid := 0
	for _, it := range items {
		w := 1.0
		hasW := true
		switch weightBy {
		case "turnover":
			if it.Turnover != nil && *it.Turnover > 0 {
				w = *it.Turnover
			} else {
				w = 0
				hasW = false
			}
		case "float_market_cap":
			if it.ScaleValue != nil && *it.ScaleValue > 0 {
				w = *it.ScaleValue
			} else {
				w = 0
				hasW = false
			}
		}
		if hasW {
			totalWeight += w
			weightedValid++
		}
		rawNodes = append(rawNodes, rawNode{item: it, weight: w, hasW: hasW})
	}

	var rightCount, validCount, unknownCount int
	rightWeight := 0.0
	nodes := make([]dto.DistributionNode, 0, len(rawNodes))
	for _, rn := range rawNodes {
		it := rn.item
		if it.Metrics.Side == string(trend.SideUnknown) {
			unknownCount++
		} else {
			validCount++
			if it.Metrics.Side == string(trend.SideRight) {
				rightCount++
				rightWeight += rn.weight
			}
		}
		wPct := 0.0
		if totalWeight > 0 && rn.hasW {
			wPct = round2(rn.weight * 100 / totalWeight)
		}
		var strScore *float64
		if it.Metrics.Strength != nil {
			strScore = it.Metrics.Strength.Score
		}
		nodes = append(nodes, dto.DistributionNode{
			AssetID:             it.AssetID,
			Name:                it.Name,
			GroupName:           firstNonEmpty(it.GroupName, it.Market, "其他"),
			WeightValue:         rn.weight,
			WeightPct:           wPct,
			Side:                it.Metrics.Side,
			Temperature:         it.Metrics.Temperature,
			DayChangePct:        it.DayChangePct,
			DurationTradingDays: it.Metrics.DurationTradingDays,
			StrengthScore:       strScore,
		})
	}

	var rightRatioPtr, weightedRightPtr, weightCovPtr *float64
	if validCount > 0 {
		r := round2(float64(rightCount) * 100 / float64(validCount))
		rightRatioPtr = &r
	}
	if totalWeight > 0 {
		wr := round2(rightWeight * 100 / totalWeight)
		weightedRightPtr = &wr
	}
	if len(items) > 0 {
		wc := round2(float64(weightedValid) * 100 / float64(len(items)))
		weightCovPtr = &wc
	}

	return &dto.TrendDistributionResponse{
		Snapshot:              meta,
		AssetID:               parentID,
		AssetName:             parentName,
		WeightBy:              weightBy,
		ColorBy:               colorBy,
		RightCount:            rightCount,
		ValidCount:            validCount,
		UnknownCount:          unknownCount,
		RightRatioPct:         rightRatioPtr,
		WeightedRightRatioPct: weightedRightPtr,
		WeightCoveragePct:     weightCovPtr,
		Nodes:                 nodes,
	}, nil
}

// GetTrendTier 返回五档相对强度梯队榜单。
func (s *TrendService) GetTrendTier(ctx context.Context, id, scopeID, snapshotID string) (*dto.TrendTierResponse, error) {
	parentID, parentName, members, err := s.resolvePoolMembers(id)
	if err != nil {
		return nil, err
	}
	effScope := firstNonEmpty(scopeID, parentID, defaultScopeID)
	snap, meta, err := s.resolveSnapshot(snapshotID, effScope, defaultCurrency, defaultMode)
	if err != nil {
		return nil, err
	}
	items, err := s.buildAssetSummaries(snap, effScope, members)
	if err != nil {
		return nil, err
	}

	buckets := []dto.TierBucket{
		{Tier: "S", ProLabel: "S 领跑", FunLabel: "夯", RuleText: "强度 ≥ 90 且右侧确认", Items: []dto.TierMemberItem{}},
		{Tier: "A", ProLabel: "A 强势", FunLabel: "顶级", RuleText: "强度 ≥ 70（或 ≥ 90 非右侧）", Items: []dto.TierMemberItem{}},
		{Tier: "B", ProLabel: "B 中坚", FunLabel: "人上人", RuleText: "30 ≤ 强度 < 70", Items: []dto.TierMemberItem{}},
		{Tier: "C", ProLabel: "C 偏弱", FunLabel: "NPC", RuleText: "10 ≤ 强度 < 30（或 < 10 非左侧）", Items: []dto.TierMemberItem{}},
		{Tier: "D", ProLabel: "D 落后", FunLabel: "拉完了", RuleText: "强度 < 10 且左侧确认", Items: []dto.TierMemberItem{}},
	}
	bucketIdx := map[string]int{"S": 0, "A": 1, "B": 2, "C": 3, "D": 4}
	excluded := []dto.TierExcludedItem{}
	validCount := 0

	for _, it := range items {
		var scorePtr *float64
		var deltaPtr *float64
		if it.Metrics.Strength != nil {
			scorePtr = it.Metrics.Strength.Score
			deltaPtr = it.Metrics.Strength.Delta
		}
		tierPtr, reason := trend.ComputeTier(scorePtr, trend.Side(it.Metrics.Side))
		if tierPtr == nil || scorePtr == nil {
			excluded = append(excluded, dto.TierExcludedItem{
				AssetID: it.AssetID,
				Name:    it.Name,
				Reason:  firstNonEmpty(reason, it.Metrics.Reason, "样本不足"),
			})
			continue
		}
		idx := bucketIdx[string(*tierPtr)]
		buckets[idx].Items = append(buckets[idx].Items, dto.TierMemberItem{
			AssetID:             it.AssetID,
			Name:                it.Name,
			Score:               *scorePtr,
			Side:                it.Metrics.Side,
			DurationTradingDays: it.Metrics.DurationTradingDays,
			Delta:               deltaPtr,
			Temperature:         it.Metrics.Temperature,
		})
		buckets[idx].Count++
		validCount++
	}

	for i := range buckets {
		sort.Slice(buckets[i].Items, func(a, b int) bool {
			return buckets[i].Items[a].Score > buckets[i].Items[b].Score
		})
	}

	return &dto.TrendTierResponse{
		Snapshot:       meta,
		AssetID:        parentID,
		AssetName:      parentName,
		CandidateCount: len(items),
		ValidCount:     validCount,
		ExcludedCount:  len(excluded),
		Tiers:          buckets,
		ExcludedItems:  excluded,
	}, nil
}

// GetTrendEvents 返回状态迁移事件列表，支持类型过滤与游标分页。
func (s *TrendService) GetTrendEvents(ctx context.Context, id, typesStr string, limit int, cursor, snapshotID string) (*dto.TrendEventsResponse, error) {
	target := s.resolveTarget(id)
	assetID := ""
	if target.Asset != nil {
		assetID = target.Asset.AssetID
	} else if target.Universe == nil && target.Category == nil && id != "all" {
		return nil, ErrNotFound
	}

	snap, _, err := s.resolveSnapshot(snapshotID, firstNonEmpty(target.ScopeID, defaultScopeID), defaultCurrency, defaultMode)
	if err != nil {
		return nil, err
	}
	if snap == nil {
		return &dto.TrendEventsResponse{Items: []dto.TrendEventItem{}}, nil
	}

	var types []string
	for _, t := range strings.Split(typesStr, ",") {
		if tt := strings.TrimSpace(t); tt != "" {
			types = append(types, tt)
		}
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	offset := parseCursor(cursor)

	entities, err := s.store.ListTrendEvents(snap.SnapshotID, assetID, types, limit+1, offset)
	if err != nil {
		return nil, err
	}

	var nextCursor *string
	if len(entities) > limit {
		entities = entities[:limit]
		nc := strconv.Itoa(offset + limit)
		nextCursor = &nc
	}

	return &dto.TrendEventsResponse{
		Items:      mapEventEntities(entities),
		NextCursor: nextCursor,
	}, nil
}

// GetMethodology 返回量化算法阈值、梯队规则与免责声明。
func (s *TrendService) GetMethodology(ctx context.Context) *dto.MethodologyResponse {
	cfg := trend.DefaultAlgorithmConfig()
	return &dto.MethodologyResponse{
		AlgorithmVersion: cfg.AlgorithmVersion,
		TemperatureThresholds: []dto.TemperatureThresholdItem{
			{Key: "boiling", Label: "沸", Min: 86, Max: 100, Desc: "极强上行动量"},
			{Key: "hot", Label: "热", Min: 72, Max: 86, Desc: "强势上行"},
			{Key: "warm", Label: "温", Min: 58, Max: 72, Desc: "偏强运行"},
			{Key: "flat", Label: "平", Min: 43, Max: 58, Desc: "中性震荡"},
			{Key: "cool", Label: "凉", Min: 29, Max: 43, Desc: "偏弱运行"},
			{Key: "cold", Label: "寒", Min: 15, Max: 29, Desc: "下行趋势"},
			{Key: "freezing", Label: "冻", Min: 0, Max: 15, Desc: "极弱下行动量"},
		},
		TierRules: []dto.TierRuleItem{
			{Tier: "S", Rule: "相对强度 ≥ 90 且右侧确认"},
			{Tier: "A", Rule: "相对强度 ≥ 70（或 ≥ 90 非右侧）"},
			{Tier: "B", Rule: "30 ≤ 相对强度 < 70"},
			{Tier: "C", Rule: "10 ≤ 相对强度 < 30（或 < 10 非左侧）"},
			{Tier: "D", Rule: "相对强度 < 10 且左侧确认"},
		},
		Disclaimers: []string{
			"本系统所有趋势温度、节气与梯队排名均基于公开历史日线计算，仅供量化研究与趋势观察参考，不构成投资建议。",
			"样本不足或未获授权覆盖的标的严格标示为 insufficient / not_covered，绝不以模拟曲线或当前时间充当历史行情。",
		},
	}
}

// GetDataStatus 返回各市场的数据覆盖情况与最新交易日状态。
func (s *TrendService) GetDataStatus(ctx context.Context) (*dto.DataStatusResponse, error) {
	_, meta, err := s.resolveSnapshot("", defaultScopeID, defaultCurrency, defaultMode)
	if err != nil {
		return nil, err
	}

	marketLabels := map[string]string{
		"CN":         "A股与国内利率",
		"US":         "美股指数",
		"HK":         "港股指数",
		"JP":         "亚太指数",
		"CN_FUT":     "国内商品期货",
		"GLOBAL_FUT": "国际商品期货",
		"FX":         "全球外汇",
	}

	type mAgg struct {
		total   int
		covered int
		asOf    string
		source  string
		status  string
	}
	aggs := make(map[string]*mAgg)
	for _, a := range market.V4CatalogAssets {
		ag, ok := aggs[a.Market]
		if !ok {
			ag = &mAgg{status: "insufficient", source: a.Provider}
			aggs[a.Market] = ag
		}
		ag.total++
		if a.LicenseStatus == market.LicenseNotCovered {
			continue
		}
		bars, _ := s.loadDedupedBars(a.AssetID)
		if len(bars) > 0 {
			ag.covered++
			last := bars[len(bars)-1]
			if last.AsOf > ag.asOf {
				ag.asOf = last.AsOf
				ag.source = last.Provider
				ag.status = last.QualityStatus
			}
		}
	}

	markets := make([]dto.DataStatusMarketItem, 0, len(aggs))
	for m, ag := range aggs {
		var asOfPtr *string
		if ag.asOf != "" {
			v := ag.asOf
			asOfPtr = &v
		}
		note := "正常覆盖"
		if ag.covered == 0 {
			note = "暂无本地或在线历史日线"
		} else if ag.covered < ag.total {
			note = fmt.Sprintf("部分标的未覆盖或样本不足（%d/%d）", ag.covered, ag.total)
		}
		markets = append(markets, dto.DataStatusMarketItem{
			Market:       m,
			Label:        firstNonEmpty(marketLabels[m], m),
			AsOf:         asOfPtr,
			SourceName:   ag.source,
			Status:       ag.status,
			CoveredCount: ag.covered,
			TotalCount:   ag.total,
			Note:         note,
		})
	}
	sort.Slice(markets, func(i, j int) bool {
		return markets[i].Market < markets[j].Market
	})

	return &dto.DataStatusResponse{
		Snapshot: meta,
		Markets:  markets,
	}, nil
}

// --- 内部辅助方法 ---

type resolvedTarget struct {
	Asset    *storage.AssetEntity
	Universe *market.UniverseDef
	Category *market.CategoryDef
	ScopeID  string
}

func (s *TrendService) resolveTarget(id string) resolvedTarget {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return resolvedTarget{}
	}
	if leg, ok := market.LegacyDivergenceMappings[cleanID]; ok {
		var res resolvedTarget
		res.ScopeID = leg.TargetUniverseID
		if u, uok := market.GetUniverse(leg.TargetUniverseID); uok {
			res.Universe = &u
		}
		if a, err := s.store.GetAssetByID(leg.TargetAssetID); err == nil {
			res.Asset = a
		}
		return res
	}
	if u, ok := market.GetUniverse(cleanID); ok {
		return resolvedTarget{Universe: &u, ScopeID: u.ID}
	}
	for _, c := range market.V4Categories {
		if c.ID == cleanID {
			cc := c
			return resolvedTarget{Category: &cc, ScopeID: defaultScopeID}
		}
	}
	if a, err := s.store.GetAssetByID(cleanID); err == nil {
		return resolvedTarget{Asset: a, ScopeID: defaultScopeID}
	}
	return resolvedTarget{}
}

func (s *TrendService) resolvePoolMembers(id string) (string, string, []storage.AssetEntity, error) {
	target := s.resolveTarget(id)
	if target.Universe != nil {
		members, err := s.store.ListUniverseMembers(target.Universe.ID)
		return target.Universe.ID, target.Universe.Name, members, err
	}
	if target.Category != nil {
		catAssets := market.GetCatalogAssetsByCategory(target.Category.ID)
		members := make([]storage.AssetEntity, 0, len(catAssets))
		for _, ca := range catAssets {
			if a, err := s.store.GetAssetByID(ca.AssetID); err == nil {
				members = append(members, *a)
			}
		}
		return target.Category.ID, target.Category.Name, members, nil
	}
	if target.Asset != nil {
		return target.Asset.AssetID, target.Asset.Name, []storage.AssetEntity{*target.Asset}, nil
	}
	if id == "" || id == "all" {
		all, err := s.store.ListAssets(true)
		return "all", "全市场观察池", all, err
	}
	return "", "", nil, ErrNotFound
}

func (s *TrendService) resolveSnapshot(snapshotID, scopeID, currency, mode string) (*storage.TrendSnapshotEntity, dto.SnapshotMeta, error) {
	if currency == "" {
		currency = defaultCurrency
	}
	if mode == "" {
		mode = defaultMode
	}
	if scopeID == "" {
		scopeID = defaultScopeID
	}

	var snap *storage.TrendSnapshotEntity
	var err error
	if snapshotID != "" {
		snap, err = s.store.GetSnapshotByID(snapshotID)
	} else {
		snap, err = s.store.GetCurrentSnapshot(scopeID, currency, mode)
		if errors.Is(err, sql.ErrNoRows) {
			snap, err = s.store.GetCurrentSnapshot(defaultScopeID, defaultCurrency, defaultMode)
		}
	}

	if errors.Is(err, sql.ErrNoRows) || snap == nil {
		return nil, dto.SnapshotMeta{
			SnapshotID:       "",
			AsOf:             "",
			ComputedAt:       "",
			AlgorithmVersion: trend.DefaultAlgorithmConfig().AlgorithmVersion,
			Currency:         currency,
			ScopeID:          scopeID,
			ScopeLabel:       resolveScopeLabel(scopeID),
			Mode:             mode,
			Status:           "insufficient",
			SourceDates:      []dto.SourceDateMeta{},
			Warnings:         []string{"尚未生成趋势快照"},
		}, nil
	}
	if err != nil {
		return nil, dto.SnapshotMeta{}, err
	}

	var extra snapshotMetaExtra
	_ = json.Unmarshal([]byte(snap.MetaJSON), &extra)
	if extra.SourceDates == nil {
		extra.SourceDates = []dto.SourceDateMeta{}
	}
	if extra.Warnings == nil {
		extra.Warnings = []string{}
	}

	meta := dto.SnapshotMeta{
		SnapshotID:       snap.SnapshotID,
		AsOf:             snap.AsOf,
		ComputedAt:       snap.ComputedAt,
		AlgorithmVersion: snap.AlgorithmVersion,
		Currency:         snap.Currency,
		ScopeID:          scopeID,
		ScopeLabel:       firstNonEmpty(resolveScopeLabel(scopeID), extra.ScopeLabel),
		CandidateCount:   snap.CandidateCount,
		ValidCount:       snap.ValidCount,
		Coverage:         snap.Coverage,
		Mode:             snap.Mode,
		Status:           snap.Status,
		SourceDates:      extra.SourceDates,
		Warnings:         extra.Warnings,
	}
	return snap, meta, nil
}

func resolveScopeLabel(scopeID string) string {
	if u, ok := market.GetUniverse(scopeID); ok {
		return u.Name
	}
	for _, c := range market.V4Categories {
		if c.ID == scopeID {
			return c.Name
		}
	}
	return "全市场观察池"
}

func (s *TrendService) loadDedupedBars(assetID string) ([]storage.DailyBarEntity, error) {
	raw, err := s.store.GetDailyBars(assetID, "", "", 0)
	if err != nil || len(raw) == 0 {
		return raw, err
	}
	dedup := make([]storage.DailyBarEntity, 0, len(raw))
	lastDate := ""
	for _, b := range raw {
		if b.TradingDate == lastDate {
			continue
		}
		lastDate = b.TradingDate
		dedup = append(dedup, b)
	}
	return dedup, nil
}

func (s *TrendService) buildAssetSummaries(snap *storage.TrendSnapshotEntity, scopeID string, assets []storage.AssetEntity) ([]dto.AssetSummaryItem, error) {
	metricMap := make(map[string]storage.TrendDailyMetricEntity)
	rankMap := make(map[string]storage.StrengthDailyRankingEntity)
	eventMap := make(map[string][]dto.TrendEventItem)

	if snap != nil && len(assets) > 0 {
		ids := make([]string, len(assets))
		for i, a := range assets {
			ids[i] = a.AssetID
		}
		if metrics, err := s.store.ListTrendDailyMetrics(snap.SnapshotID, ids); err == nil {
			for _, m := range metrics {
				metricMap[m.AssetID] = m
			}
		}
		rankings, err := s.store.ListStrengthDailyRankings(snap.SnapshotID, scopeID, defaultCurrency)
		if err != nil || len(rankings) == 0 {
			rankings, _ = s.store.ListStrengthDailyRankings(snap.SnapshotID, defaultScopeID, defaultCurrency)
		}
		for _, r := range rankings {
			rankMap[r.AssetID] = r
		}
		if evs, err := s.store.ListTrendEvents(snap.SnapshotID, "", nil, 200, 0); err == nil {
			for _, ev := range evs {
				eventMap[ev.AssetID] = append(eventMap[ev.AssetID], mapEventEntity(ev))
			}
		}
	}

	items := make([]dto.AssetSummaryItem, 0, len(assets))
	for _, a := range assets {
		bars, _ := s.loadDedupedBars(a.AssetID)
		var asOfPtr *string
		var latestVal, dayChg, turnover *float64
		qStatus := "insufficient"
		if a.LicenseStatus == string(market.LicenseNotCovered) {
			qStatus = "not_covered"
		}
		if len(bars) > 0 {
			last := bars[len(bars)-1]
			asOf := last.AsOf
			asOfPtr = &asOf
			latestVal = last.Close
			turnover = last.Turnover
			qStatus = last.QualityStatus
			if len(bars) >= 2 && bars[len(bars)-2].Close != nil && *bars[len(bars)-2].Close > 0 && last.Close != nil {
				chg := round2((*last.Close / *bars[len(bars)-2].Close - 1.0) * 100)
				dayChg = &chg
			}
		}

		m := emptyMetrics()
		if dm, ok := metricMap[a.AssetID]; ok {
			if dm.Temperature != nil {
				m.Temperature = *dm.Temperature
			}
			m.TemperatureScore = dm.TemperatureScore
			m.Direction = dm.Direction
			m.Confirmation = dm.Confirmation
			m.Side = dm.Side
			m.StartedAt = dm.StartedAt
			m.ConfirmedAt = dm.ConfirmedAt
			m.DurationTradingDays = dm.DurationTradingDays
			m.DurationCalendarDays = dm.DurationCalendarDays
			m.Phase = dm.Phase
			m.Tier = dm.Tier
			if dm.Reason != nil {
				m.Reason = *dm.Reason
			}
		}
		if rk, ok := rankMap[a.AssetID]; ok {
			reason := ""
			if rk.Reason != nil {
				reason = *rk.Reason
			}
			m.Strength = &dto.StrengthInfo{
				ScopeID:        rk.ScopeID,
				ScopeLabel:     resolveScopeLabel(rk.ScopeID),
				Score:          rk.Score,
				Rank:           rk.Rank,
				RawFactor:      rk.RawFactor,
				ValidCount:     rk.ValidCount,
				CandidateCount: rk.CandidateCount,
				Delta:          rk.Delta,
				Reason:         reason,
			}
		}
		if evs, ok := eventMap[a.AssetID]; ok {
			m.Events = evs
		}

		items = append(items, dto.AssetSummaryItem{
			AssetID:       a.AssetID,
			Market:        a.Market,
			Symbol:        a.Symbol,
			Name:          a.Name,
			ShortName:     a.ShortName,
			AssetType:     a.AssetType,
			MeasureType:   a.MeasureType,
			Unit:          a.Unit,
			Currency:      a.Currency,
			AsOf:          asOfPtr,
			LatestValue:   latestVal,
			DayChangePct:  dayChg,
			Turnover:      turnover,
			ScaleLabel:    a.Unit,
			GroupName:     a.Market,
			HasMembers:    false,
			MemberCount:   0,
			LicenseStatus: a.LicenseStatus,
			QualityStatus: qStatus,
			Metrics:       m,
			FeatureTags:   []string{},
		})
	}
	return items, nil
}

func (s *TrendService) buildUniverseSummary(snap *storage.TrendSnapshotEntity, u market.UniverseDef) dto.AssetSummaryItem {
	members, _ := s.store.ListUniverseMembers(u.ID)
	items, _ := s.buildAssetSummaries(snap, u.ID, members)
	var rightCount, validCount int
	var asOfPtr *string
	for _, it := range items {
		if it.AsOf != nil && (asOfPtr == nil || *it.AsOf > *asOfPtr) {
			v := *it.AsOf
			asOfPtr = &v
		}
		if it.Metrics.Side != string(trend.SideUnknown) {
			validCount++
			if it.Metrics.Side == string(trend.SideRight) {
				rightCount++
			}
		}
	}
	var rr *float64
	if validCount > 0 {
		v := round2(float64(rightCount) * 100 / float64(validCount))
		rr = &v
	}
	return dto.AssetSummaryItem{
		AssetID:       u.ID,
		Market:        "MULTI",
		Symbol:        u.ID,
		Name:          u.Name,
		ShortName:     u.Name,
		AssetType:     "universe",
		MeasureType:   "index",
		AsOf:          asOfPtr,
		HasMembers:    true,
		MemberCount:   len(members),
		RightRatioPct: rr,
		RightCount:    rightCount,
		ValidCount:    validCount,
		LicenseStatus: string(u.LicenseStatus),
		QualityStatus: "ok",
		Metrics:       emptyMetrics(),
		FeatureTags:   []string{},
	}
}

func emptyMetrics() dto.TrendMetrics {
	return dto.TrendMetrics{
		Temperature:  string(trend.TempUnavailable),
		Direction:    string(trend.DirectionUnknown),
		Confirmation: string(trend.ConfirmationUnknown),
		Side:         string(trend.SideUnknown),
		Events:       []dto.TrendEventItem{},
	}
}

func toTrendBars(bars []storage.DailyBarEntity) []trend.DailyBar {
	out := make([]trend.DailyBar, 0, len(bars))
	for _, b := range bars {
		val := b.Close
		if b.AdjustedClose != nil && *b.AdjustedClose > 0 {
			val = b.AdjustedClose
		}
		out = append(out, trend.DailyBar{
			AssetID:     b.AssetID,
			TradingDate: b.TradingDate,
			Value:       val,
			Volume:      b.Volume,
			Turnover:    b.Turnover,
		})
	}
	return out
}

func filterBarsByRange(bars []storage.DailyBarEntity, rangeStr string) []storage.DailyBarEntity {
	if len(bars) == 0 || rangeStr == "" || rangeStr == "ALL" {
		return bars
	}
	lastDate, err := time.Parse("2006-01-02", bars[len(bars)-1].TradingDate)
	if err != nil {
		return bars
	}
	var cutoff time.Time
	switch rangeStr {
	case "1M":
		cutoff = lastDate.AddDate(0, -1, 0)
	case "3M":
		cutoff = lastDate.AddDate(0, -3, 0)
	case "6M":
		cutoff = lastDate.AddDate(0, -6, 0)
	case "1Y":
		cutoff = lastDate.AddDate(-1, 0, 0)
	case "3Y":
		cutoff = lastDate.AddDate(-3, 0, 0)
	case "5Y":
		cutoff = lastDate.AddDate(-5, 0, 0)
	default:
		return bars
	}
	cutStr := cutoff.Format("2006-01-02")
	idx := sort.Search(len(bars), func(i int) bool {
		return bars[i].TradingDate >= cutStr
	})
	return bars[idx:]
}

func buildSeriesPoints(bars []storage.DailyBarEntity) []dto.SeriesPoint {
	pts := make([]dto.SeriesPoint, 0, len(bars))
	var baseVal float64
	for _, b := range bars {
		if b.Close != nil && *b.Close > 0 && baseVal == 0 {
			baseVal = *b.Close
		}
		var retPtr *float64
		if baseVal > 0 && b.Close != nil {
			r := round2((*b.Close/baseVal - 1.0) * 100)
			retPtr = &r
		}
		pts = append(pts, dto.SeriesPoint{
			TS:          b.TradingDate,
			TradingDate: b.TradingDate,
			Value:       b.Close,
			ReturnPct:   retPtr,
			Turnover:    b.Turnover,
			Status:      b.QualityStatus,
		})
	}
	return pts
}

func buildExtremumInfo(ext trend.ExtremumResult, bars []storage.DailyBarEntity) dto.ExtremumInfo {
	if !ext.Valid || len(bars) == 0 {
		return dto.ExtremumInfo{}
	}
	ddPct := round2(ext.MaxDrawdown * 100)
	mgPct := round2(ext.MaxGain * 100)
	var hiVal, loVal *float64
	var hiDate, loDate *string
	for _, b := range bars {
		if b.Close == nil || *b.Close <= 0 {
			continue
		}
		v := *b.Close
		d := b.TradingDate
		if hiVal == nil || v > *hiVal {
			vv := v
			dd := d
			hiVal = &vv
			hiDate = &dd
		}
		if loVal == nil || v < *loVal {
			vv := v
			dd := d
			loVal = &vv
			loDate = &dd
		}
	}
	return dto.ExtremumInfo{
		MaxDrawdownPct:     &ddPct,
		DrawdownPeakDate:   strOrNil(ext.DrawdownPeakDate),
		DrawdownTroughDate: strOrNil(ext.DrawdownTroughDate),
		MaxGainPct:         &mgPct,
		GainTroughDate:     strOrNil(ext.GainTroughDate),
		GainPeakDate:       strOrNil(ext.GainPeakDate),
		HighValue:          hiVal,
		HighDate:           hiDate,
		LowValue:           loVal,
		LowDate:            loDate,
	}
}

func buildStateIntervals(states []trend.StateResult) []dto.StateInterval {
	if len(states) == 0 {
		return []dto.StateInterval{}
	}
	intervals := make([]dto.StateInterval, 0)
	cur := dto.StateInterval{
		StartDate: states[0].TradingDate,
		EndDate:   states[0].TradingDate,
		Side:      string(states[0].Side),
	}
	for _, st := range states[1:] {
		if string(st.Side) == cur.Side {
			cur.EndDate = st.TradingDate
		} else {
			intervals = append(intervals, cur)
			cur = dto.StateInterval{
				StartDate: st.TradingDate,
				EndDate:   st.TradingDate,
				Side:      string(st.Side),
			}
		}
	}
	intervals = append(intervals, cur)
	return intervals
}

func buildSparkline(bars []storage.DailyBarEntity, maxPts int) ([]*float64, *float64, *float64) {
	if len(bars) == 0 {
		return []*float64{}, nil, nil
	}
	step := 1
	if maxPts > 0 && len(bars) > maxPts {
		step = len(bars) / maxPts
		if step < 1 {
			step = 1
		}
	}
	var pts []*float64
	var minV, maxV *float64
	for i := 0; i < len(bars); i += step {
		v := bars[i].Close
		pts = append(pts, v)
		if v != nil {
			if minV == nil || *v < *minV {
				vv := *v
				minV = &vv
			}
			if maxV == nil || *v > *maxV {
				vv := *v
				maxV = &vv
			}
		}
	}
	return pts, minV, maxV
}

func computeTrailingReturnPct(bars []storage.DailyBarEntity, tradingDays int) *float64 {
	if len(bars) < 2 {
		return nil
	}
	last := bars[len(bars)-1]
	if last.Close == nil || *last.Close <= 0 {
		return nil
	}
	idx := len(bars) - 1 - tradingDays
	if idx < 0 {
		idx = 0
	}
	base := bars[idx]
	if base.Close == nil || *base.Close <= 0 {
		return nil
	}
	v := round2((*last.Close / *base.Close - 1.0) * 100)
	return &v
}

func computePeriodChangeByCode(bars []storage.DailyBarEntity, period string, dayChg *float64) *float64 {
	switch period {
	case "1D", "":
		return dayChg
	case "1W":
		return computeTrailingReturnPct(bars, 5)
	case "1M":
		return computeTrailingReturnPct(bars, 21)
	case "3M":
		return computeTrailingReturnPct(bars, 63)
	case "6M":
		return computeTrailingReturnPct(bars, 126)
	case "1Y":
		return computeTrailingReturnPct(bars, 252)
	default:
		return dayChg
	}
}

func sortAssetSummaries(items []dto.AssetSummaryItem, sortField, order string) {
	if sortField == "" {
		return
	}
	desc := strings.ToLower(order) != "asc"
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		va, oka := extractSortVal(a, sortField)
		vb, okb := extractSortVal(b, sortField)
		if oka != okb {
			return oka
		}
		if !oka {
			return a.AssetID < b.AssetID
		}
		if desc {
			return va > vb
		}
		return va < vb
	})
}

func extractSortVal(item dto.AssetSummaryItem, field string) (float64, bool) {
	switch field {
	case "temperature_score", "temperature":
		if item.Metrics.TemperatureScore != nil {
			return *item.Metrics.TemperatureScore, true
		}
	case "strength_score", "strength":
		if item.Metrics.Strength != nil && item.Metrics.Strength.Score != nil {
			return *item.Metrics.Strength.Score, true
		}
	case "day_change_pct":
		if item.DayChangePct != nil {
			return *item.DayChangePct, true
		}
	case "duration_trading_days", "days":
		if item.Metrics.DurationTradingDays != nil {
			return float64(*item.Metrics.DurationTradingDays), true
		}
	case "latest_value":
		if item.LatestValue != nil {
			return *item.LatestValue, true
		}
	case "turnover":
		if item.Turnover != nil {
			return *item.Turnover, true
		}
	}
	return 0, false
}

func mapEventEntities(evs []storage.TrendEventEntity) []dto.TrendEventItem {
	out := make([]dto.TrendEventItem, 0, len(evs))
	for _, e := range evs {
		out = append(out, mapEventEntity(e))
	}
	return out
}

func mapEventEntity(e storage.TrendEventEntity) dto.TrendEventItem {
	oldV := ""
	if e.OldValue != nil {
		oldV = *e.OldValue
	}
	newV := ""
	if e.NewValue != nil {
		newV = *e.NewValue
	}
	return dto.TrendEventItem{
		EventID:          e.EventID,
		AssetID:          e.AssetID,
		EventType:        e.EventType,
		OldValue:         oldV,
		NewValue:         newV,
		AsOf:             e.AsOf,
		DetectedAt:       e.DetectedAt,
		SnapshotID:       e.SnapshotID,
		AlgorithmVersion: e.AlgorithmVersion,
		ScopeID:          e.ScopeID,
	}
}

func parseCursor(c string) int {
	if c == "" {
		return 0
	}
	v, err := strconv.Atoi(strings.TrimSpace(c))
	if err != nil || v < 0 {
		return 0
	}
	return v
}

func parseYearFilter(yearsStr string) map[int]bool {
	if strings.TrimSpace(yearsStr) == "" {
		return nil
	}
	res := make(map[int]bool)
	for _, p := range strings.Split(yearsStr, ",") {
		if y, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			res[y] = true
		}
	}
	return res
}

func fracPtrToPctPtr(f *float64) *float64 {
	if f == nil {
		return nil
	}
	v := round2(*f * 100)
	return &v
}

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
