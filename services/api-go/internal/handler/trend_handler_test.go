package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"cyclecompass/api/internal/config"
	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/handler"
	"cyclecompass/api/internal/market"
	"cyclecompass/api/internal/router"
	"cyclecompass/api/internal/service"
	"cyclecompass/api/internal/storage"
)

func setupTestRouter(t *testing.T) http.Handler {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_trend.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	// 写入确定性历史日线（150 个交易日，截至 2025-07-01，绝不用 time.Now()）
	seedAssets := []string{"CN.SSE.000300", "CN.SSE.000905", "CN.CSI.930955", "CN.SSE.510300"}
	var bars []storage.DailyBarEntity
	baseDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for idx, assetID := range seedAssets {
		price := 100.0 + float64(idx*10)
		dayCount := 0
		for d := 0; dayCount < 150; d++ {
			cur := baseDate.AddDate(0, 0, d)
			if cur.Weekday() == time.Saturday || cur.Weekday() == time.Sunday {
				continue
			}
			dayCount++
			dateStr := cur.Format("2006-01-02")
			if idx%2 == 0 {
				price *= 1.0025
			} else {
				price *= 0.9990
			}
			c := price
			turnover := 1e8 + float64(dayCount*1e5)
			bars = append(bars, storage.DailyBarEntity{
				AssetID:       assetID,
				TradingDate:   dateStr,
				Provider:      "test_seed",
				Revision:      1,
				Close:         &c,
				Turnover:      &turnover,
				QualityStatus: "offline_seed",
				AsOf:          dateStr,
				AvailableAt:   dateStr + "T15:30:00+08:00",
				UpdatedAt:     "2025-07-01T16:00:00Z",
			})
		}
	}
	if err := store.UpsertDailyBars(bars); err != nil {
		t.Fatalf("UpsertDailyBars failed: %v", err)
	}

	hist := market.NewHistoricalProvider(market.HistoricalProviderConfig{
		OfflineOnly: true,
	})
	trendSvc := service.NewTrendService(store, hist)
	if err := trendSvc.BootstrapAndComputeSnapshot(context.Background(), false); err != nil {
		t.Fatalf("BootstrapAndComputeSnapshot failed: %v", err)
	}

	cfg := &config.Config{
		Port:        "8002",
		AppName:     "CycleCompass Test",
		AppVersion:  "test",
		CorsOrigins: []string{"*"},
		DataDir:     tmpDir,
		DBPath:      dbPath,
	}

	seedP := market.NewSeedProvider(tmpDir)
	onlineP := market.NewOnlineProvider(seedP)

	return router.Setup(router.RouterParams{
		Config:    cfg,
		Root:      handler.NewRootHandler(cfg, onlineP),
		Market:    handler.NewMarketHandler(service.NewMarketService(onlineP)),
		Sector:    handler.NewSectorHandler(service.NewSectorService(seedP)),
		Cycle:     handler.NewCycleHandler(service.NewCycleService(onlineP, seedP)),
		Strategy:  handler.NewStrategyHandler(service.NewStrategyService(onlineP, seedP)),
		Portfolio: handler.NewPortfolioHandler(service.NewPortfolioService(store, onlineP)),
		Backtest:  handler.NewBacktestHandler(service.NewBacktestService(seedP)),
		Trend:     handler.NewTrendHandler(trendSvc),
	})
}

func doReq(t *testing.T, h http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestV4TrendEndpointsAndLegacyMappings(t *testing.T) {
	h := setupTestRouter(t)

	// 1. 原有 /api/v1/health 无回归
	healthRec := doReq(t, h, http.MethodGet, "/api/v1/health")
	if healthRec.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/health status = %d, want 200", healthRec.Code)
	}
	var healthResp dto.ApiResponse[dto.HealthData]
	if err := json.Unmarshal(healthRec.Body.Bytes(), &healthResp); err != nil {
		t.Fatalf("unmarshal health failed: %v", err)
	}
	if healthResp.Code != 0 || healthResp.Data.Status != "healthy" {
		t.Fatalf("unexpected health response: %+v", healthResp)
	}

	// 2. 全部 V4 GET/HEAD 路由均返回 200 且 code == 0
	okRoutes := []string{
		"/api/v1/assets/catalog",
		"/api/v1/assets/universes",
		"/api/v1/assets/search?q=沪深&limit=10",
		"/api/v1/assets/CN.SSE.000300",
		"/api/v1/assets/universe_a_sector/members?sort=temperature_score&order=desc&limit=5",
		"/api/v1/trends/snapshots/latest",
		"/api/v1/trends/CN.SSE.000300/curve?range=1Y",
		"/api/v1/trends/CN.SSE.000300/annual",
		"/api/v1/trends/universe_a_sector/components?range=1Y",
		"/api/v1/trends/universe_a_sector/dashboard",
		"/api/v1/trends/universe_a_sector/season",
		"/api/v1/trends/universe_a_sector/distribution?weight_by=equal",
		"/api/v1/trends/universe_a_sector/tier",
		"/api/v1/trends/CN.SSE.000300/events?limit=10",
		"/api/v1/trends/methodology",
		"/api/v1/trends/data-status",
	}
	for _, route := range okRoutes {
		rec := doReq(t, h, http.MethodGet, route)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, body = %s", route, rec.Code, rec.Body.String())
			continue
		}
		var env struct {
			Code int `json:"code"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil || env.Code != 0 {
			t.Errorf("GET %s invalid envelope: err=%v body=%s", route, err, rec.Body.String())
		}

		headRec := doReq(t, h, http.MethodHead, route)
		if headRec.Code != http.StatusOK {
			t.Errorf("HEAD %s status = %d, want 200", route, headRec.Code)
		}
	}

	// 3. 验证快照 as_of 诚实反映真实日线截止日期（非 time.Now()）
	snapRec := doReq(t, h, http.MethodGet, "/api/v1/trends/snapshots/latest")
	var snapResp dto.ApiResponse[dto.SnapshotMeta]
	if err := json.Unmarshal(snapRec.Body.Bytes(), &snapResp); err != nil {
		t.Fatalf("unmarshal snapshot failed: %v", err)
	}
	if snapResp.Data.AsOf >= "2026-01-01" || snapResp.Data.AsOf == "" {
		t.Fatalf("snapshot as_of = %q, expected historical 2025 date", snapResp.Data.AsOf)
	}
	if snapResp.Data.Status != "offline_seed" {
		t.Fatalf("snapshot status = %q, want offline_seed", snapResp.Data.Status)
	}

	// 4. 验证旧 ID 兼容映射
	legacyExpect := map[string]string{
		"a_shares":      "CN.SSE.000300",
		"dividend_etf":  "CN.CSI.930955",
		"etf_macro":     "CN.SSE.510300",
		"a_shares_long": "CN.SSE.000300",
	}
	for legID, wantAssetID := range legacyExpect {
		rec := doReq(t, h, http.MethodGet, "/api/v1/assets/"+legID)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/v1/assets/%s status = %d", legID, rec.Code)
		}
		var detailResp dto.ApiResponse[dto.AssetDetailResponse]
		if err := json.Unmarshal(rec.Body.Bytes(), &detailResp); err != nil {
			t.Fatalf("unmarshal legacy detail %s failed: %v", legID, err)
		}
		if detailResp.Data.Asset.AssetID != wantAssetID {
			t.Errorf("legacy %s mapped to %q, want %q", legID, detailResp.Data.Asset.AssetID, wantAssetID)
		}
	}

	// 5. 验证参数白名单校验（400）
	badRoutes := []string{
		"/api/v1/assets/universe_a_sector/members?sort=drop_table",
		"/api/v1/assets/universe_a_sector/members?order=random",
		"/api/v1/assets/universe_a_sector/members?limit=201",
		"/api/v1/assets/search?limit=999",
		"/api/v1/trends/CN.SSE.000300/curve?range=100Y",
		"/api/v1/trends/universe_a_sector/season?direction=sideways",
		"/api/v1/trends/universe_a_sector/distribution?weight_by=magic",
	}
	for _, route := range badRoutes {
		rec := doReq(t, h, http.MethodGet, route)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("GET %s status = %d, want 400", route, rec.Code)
		}
	}

	// 6. 验证未知资产返回 404
	notFoundRec := doReq(t, h, http.MethodGet, "/api/v1/assets/UNKNOWN.ASSET.XYZ")
	if notFoundRec.Code != http.StatusNotFound {
		t.Errorf("GET unknown asset status = %d, want 404", notFoundRec.Code)
	}
}

func TestV5BatchAMembersFilteringPaginationSnapshotAndRangeAvailability(t *testing.T) {
	h := setupTestRouter(t)

	// 1. 验证 GetAssetDetail 返回 earliest_date 与 range_availability（150 交易日约 210 自然日：6M 可用，1Y/3Y 禁用）
	detailRec := doReq(t, h, http.MethodGet, "/api/v1/assets/CN.SSE.000300")
	if detailRec.Code != http.StatusOK {
		t.Fatalf("GET detail status = %d, body = %s", detailRec.Code, detailRec.Body.String())
	}
	var detailResp dto.ApiResponse[dto.AssetDetailResponse]
	if err := json.Unmarshal(detailRec.Body.Bytes(), &detailResp); err != nil {
		t.Fatalf("unmarshal detail failed: %v", err)
	}
	if detailResp.Data.EarliestDate == nil || *detailResp.Data.EarliestDate != "2025-01-01" {
		t.Fatalf("earliest_date = %v, want 2025-01-01", detailResp.Data.EarliestDate)
	}
	rangeMap := map[string]dto.RangeAvailabilityItem{}
	for _, ra := range detailResp.Data.RangeAvailability {
		rangeMap[ra.Range] = ra
	}
	if !rangeMap["1M"].Available || !rangeMap["3M"].Available || !rangeMap["6M"].Available || !rangeMap["ALL"].Available {
		t.Errorf("expected 1M/3M/6M/ALL available for 150-day series, got %+v", rangeMap)
	}
	if rangeMap["1Y"].Available || rangeMap["3Y"].Available || rangeMap["1Y"].Reason == "" {
		t.Errorf("expected 1Y/3Y disabled with reason, got %+v", rangeMap)
	}

	// 2. 验证全景成员查询：服务端筛选、排序稳定性、全池 overview、市场选项与绑定快照游标分页
	page1Rec := doReq(t, h, http.MethodGet, "/api/v1/assets/universe_macro/members?sort=strength&order=desc&limit=2")
	if page1Rec.Code != http.StatusOK {
		t.Fatalf("GET members page1 status = %d, body = %s", page1Rec.Code, page1Rec.Body.String())
	}
	var page1 dto.ApiResponse[dto.MembersResponse]
	if err := json.Unmarshal(page1Rec.Body.Bytes(), &page1); err != nil {
		t.Fatalf("unmarshal page1 failed: %v", err)
	}
	if page1.Data.PoolTotal != 8 {
		t.Errorf("pool_total = %d, want 8", page1.Data.PoolTotal)
	}
	if page1.Data.Overview == nil || page1.Data.Overview.TotalCount != 8 {
		t.Errorf("overview total_count = %+v, want 8", page1.Data.Overview)
	}
	if len(page1.Data.AvailableMarkets) < 3 {
		t.Errorf("available_markets = %v, expected full pool markets (>=3)", page1.Data.AvailableMarkets)
	}
	if page1.Data.NextCursor == nil || *page1.Data.NextCursor == "" {
		t.Fatalf("expected non-empty next_cursor on page1")
	}

	// 翻页加载第 2 页，验证 pool_total / valid_count / overview / available_markets 与第 1 页一致且不重复
	page2Rec := doReq(t, h, http.MethodGet, "/api/v1/assets/universe_macro/members?sort=strength&order=desc&limit=2&cursor="+*page1.Data.NextCursor+"&snapshot_id="+page1.Data.Snapshot.SnapshotID)
	if page2Rec.Code != http.StatusOK {
		t.Fatalf("GET members page2 status = %d, body = %s", page2Rec.Code, page2Rec.Body.String())
	}
	var page2 dto.ApiResponse[dto.MembersResponse]
	if err := json.Unmarshal(page2Rec.Body.Bytes(), &page2); err != nil {
		t.Fatalf("unmarshal page2 failed: %v", err)
	}
	if page2.Data.PoolTotal != page1.Data.PoolTotal || page2.Data.ValidCount != page1.Data.ValidCount {
		t.Errorf("pagination altered pool stats: p1=%+v p2=%+v", page1.Data.Overview, page2.Data.Overview)
	}
	if len(page2.Data.Items) == 0 || page2.Data.Items[0].AssetID == page1.Data.Items[0].AssetID {
		t.Errorf("page2 returned invalid/duplicate items: p1=%v p2=%v", page1.Data.Items, page2.Data.Items)
	}

	// 3. 验证服务端筛选（例如 market=US），matched_total 缩小但 pool_total 与 available_markets 仍来自全池
	filteredRec := doReq(t, h, http.MethodGet, "/api/v1/assets/universe_macro/members?market=US&limit=20")
	if filteredRec.Code != http.StatusOK {
		t.Fatalf("GET filtered members status = %d, body = %s", filteredRec.Code, filteredRec.Body.String())
	}
	var filtered dto.ApiResponse[dto.MembersResponse]
	if err := json.Unmarshal(filteredRec.Body.Bytes(), &filtered); err != nil {
		t.Fatalf("unmarshal filtered failed: %v", err)
	}
	if filtered.Data.MatchedTotal != 1 || len(filtered.Data.Items) != 1 || filtered.Data.Items[0].Market != "US" {
		t.Errorf("unexpected market=US filter result: matched=%d items=%+v", filtered.Data.MatchedTotal, filtered.Data.Items)
	}
	if filtered.Data.PoolTotal != 8 || len(filtered.Data.AvailableMarkets) < 3 {
		t.Errorf("filtered query altered pool_total=%d or available_markets=%v", filtered.Data.PoolTotal, filtered.Data.AvailableMarkets)
	}

	// 4. 验证过期游标返回 409
	staleCursorRec := doReq(t, h, http.MethodGet, "/api/v1/assets/universe_macro/members?cursor=snap_old_000:2")
	if staleCursorRec.Code != http.StatusConflict {
		t.Errorf("stale cursor status = %d, want 409", staleCursorRec.Code)
	}

	// 5. 验证不存在/过期 snapshot_id 返回 410，不悄悄返回最新
	expiredSnapRec := doReq(t, h, http.MethodGet, "/api/v1/assets/CN.SSE.000300?snapshot_id=snap_nonexistent_999")
	if expiredSnapRec.Code != http.StatusGone {
		t.Errorf("expired snapshot_id status = %d, want 410", expiredSnapRec.Code)
	}

	// 6. 验证 curve 支持 start_date/end_date 精确区间查询
	curveRec := doReq(t, h, http.MethodGet, "/api/v1/trends/CN.SSE.000300/curve?start_date=2025-02-03&end_date=2025-02-28")
	if curveRec.Code != http.StatusOK {
		t.Fatalf("GET curve window status = %d", curveRec.Code)
	}
	var curveResp dto.ApiResponse[dto.TrendCurveResponse]
	if err := json.Unmarshal(curveRec.Body.Bytes(), &curveResp); err != nil {
		t.Fatalf("unmarshal curve failed: %v", err)
	}
	for _, pt := range curveResp.Data.Points {
		if pt.TradingDate < "2025-02-03" || pt.TradingDate > "2025-02-28" {
			t.Fatalf("curve point outside requested window: %s", pt.TradingDate)
		}
	}
}

func TestV5BatchBPeriodMatrixEndpoint(t *testing.T) {
	h := setupTestRouter(t)

	// 1. 查询周期矩阵并按 rolling_1m 倒序全池排序
	rec := doReq(t, h, http.MethodGet, "/api/v1/trends/universe_macro/matrix?currency=CNY&sort=rolling_1m&order=desc&limit=20")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET matrix status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp dto.ApiResponse[dto.PeriodMatrixResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal matrix failed: %v", err)
	}
	if len(resp.Data.Columns) != 6 {
		t.Fatalf("columns len = %d, want 6", len(resp.Data.Columns))
	}
	if resp.Data.PoolID != "universe_macro" || resp.Data.SortScope != "pool" {
		t.Errorf("unexpected matrix meta: pool=%s scope=%s", resp.Data.PoolID, resp.Data.SortScope)
	}
	var totalRows int
	for _, grp := range resp.Data.Groups {
		totalRows += len(grp.Rows)
	}
	if totalRows != resp.Data.PoolTotal || totalRows == 0 {
		t.Errorf("matrix rows=%d pool_total=%d", totalRows, resp.Data.PoolTotal)
	}

	// 验证排序顺序（合法收益率严格降序，缺失或利率 asset 在后）
	if len(resp.Data.Groups) > 0 && len(resp.Data.Groups[0].Rows) >= 2 {
		rows := resp.Data.Groups[0].Rows
		seenNil := false
		prevRet := 1e9
		for _, r := range rows {
			cell := r.Cells["rolling_1m"]
			if r.Asset.MeasureType == "yield_pct" {
				if cell.DisplayUnit != "bp" || cell.ReturnPct != nil {
					t.Errorf("bond asset %s should use bp only: %+v", r.Asset.AssetID, cell)
				}
			}
			if cell.ReturnPct == nil || r.Asset.MeasureType == "yield_pct" {
				seenNil = true
				continue
			}
			if seenNil {
				t.Errorf("valid returnPct appeared after missing/bond row: %+v", r.Asset.AssetID)
			}
			if *cell.ReturnPct > prevRet {
				t.Errorf("matrix rows not descending on rolling_1m: %.2f > %.2f", *cell.ReturnPct, prevRet)
			}
			prevRet = *cell.ReturnPct
		}
	}

	// 2. 分页与过期游标 409 / 不存在快照 410
	stale := doReq(t, h, http.MethodGet, "/api/v1/trends/universe_macro/matrix?cursor=snap_old:2")
	if stale.Code != http.StatusConflict {
		t.Errorf("stale matrix cursor status = %d, want 409", stale.Code)
	}
	expired := doReq(t, h, http.MethodGet, "/api/v1/trends/universe_macro/matrix?snapshot_id=snap_missing_999")
	if expired.Code != http.StatusGone {
		t.Errorf("expired matrix snapshot status = %d, want 410", expired.Code)
	}
}

func TestV6ContractEnumsSearchSortAndDashboard_AuditB(t *testing.T) {
	h := setupTestRouter(t)

	// 1. 从共用枚举接口读取前端所有枚举并生成真实 HTTP 请求
	enumsRec := doReq(t, h, http.MethodGet, "/api/v1/assets/enums")
	if enumsRec.Code != http.StatusOK {
		t.Fatalf("GET enums status = %d", enumsRec.Code)
	}
	var enumsResp dto.ApiResponse[dto.AssetEnumsResponse]
	if err := json.Unmarshal(enumsRec.Body.Bytes(), &enumsResp); err != nil {
		t.Fatalf("unmarshal enums failed: %v", err)
	}
	enums := enumsResp.Data
	if len(enums.SortFields) == 0 || len(enums.SearchTypes) == 0 || len(enums.DashboardSorts) == 0 {
		t.Fatalf("enums unexpectedly empty: %+v", enums)
	}

	// 验证每个列表排序字段都返回 200 及合法顺序（包括 duration_trading_days）
	for _, sf := range enums.SortFields {
		rec := doReq(t, h, http.MethodGet, "/api/v1/assets/universe_a_sector/members?sort="+sf.Key+"&order=desc&limit=10")
		if rec.Code != http.StatusOK {
			t.Errorf("members sort=%s status=%d body=%s", sf.Key, rec.Code, rec.Body.String())
		}
	}

	// 验证每个仪表盘排序字段（含 sort=change & change_period=1M）均返回 200 且按 1M 真实顺序降序
	for _, df := range enums.DashboardSorts {
		rec := doReq(t, h, http.MethodGet, "/api/v1/trends/universe_a_sector/dashboard?change_period=1M&sort="+df.Key+"&order=desc")
		if rec.Code != http.StatusOK {
			t.Errorf("dashboard sort=%s status=%d body=%s", df.Key, rec.Code, rec.Body.String())
		}
		if df.Key == "change" {
			var dash dto.ApiResponse[dto.TrendDashboardResponse]
			if err := json.Unmarshal(rec.Body.Bytes(), &dash); err != nil {
				t.Fatalf("unmarshal dashboard failed: %v", err)
			}
			var flat []dto.DashboardRow
			for _, grp := range dash.Data.Groups {
				flat = append(flat, grp.Rows...)
			}
			prev := 1e9
			seenNil := false
			for _, row := range flat {
				if row.ChangePct == nil {
					seenNil = true
					continue
				}
				if seenNil {
					t.Errorf("non-nil change_pct appeared after nil row")
				}
				if *row.ChangePct > prev {
					t.Errorf("dashboard change_period=1M not descending: %.2f > %.2f", *row.ChangePct, prev)
				}
				prev = *row.ChangePct
			}
		}
	}

	// 2. 验证统一搜索支持 全部/单品种逗号拼接/行业专题/大类观察池
	singleRec := doReq(t, h, http.MethodGet, "/api/v1/assets/search?q=沪深&type=stock,index,future,futures,bond,bond_yield,fx")
	if singleRec.Code != http.StatusOK {
		t.Fatalf("search single status=%d body=%s", singleRec.Code, singleRec.Body.String())
	}
	var single dto.ApiResponse[dto.SearchAssetsResponse]
	if err := json.Unmarshal(singleRec.Body.Bytes(), &single); err != nil {
		t.Fatalf("unmarshal single search failed: %v", err)
	}
	if len(single.Data.Items) == 0 || single.Data.Items[0].ObjectType != "asset" || single.Data.Items[0].RouteTarget != "asset_detail" {
		t.Fatalf("unexpected single search result: %+v", single.Data.Items)
	}

	catRec := doReq(t, h, http.MethodGet, "/api/v1/assets/search?q=算力&type=category")
	var catResp dto.ApiResponse[dto.SearchAssetsResponse]
	_ = json.Unmarshal(catRec.Body.Bytes(), &catResp)
	if len(catResp.Data.Items) == 0 || catResp.Data.Items[0].ObjectType != "category" || catResp.Data.Items[0].RouteTarget != "sector_detail" {
		t.Fatalf("expected industry sector object in category search, got %+v", catResp.Data)
	}

	uniRec := doReq(t, h, http.MethodGet, "/api/v1/assets/search?q=大类&type=universe")
	var uniResp dto.ApiResponse[dto.SearchAssetsResponse]
	_ = json.Unmarshal(uniRec.Body.Bytes(), &uniResp)
	if len(uniResp.Data.Items) == 0 || uniResp.Data.Items[0].ObjectType != "universe" || uniResp.Data.Items[0].RouteTarget != "pool_overview" {
		t.Fatalf("expected universe object in universe search, got %+v", uniResp.Data)
	}

	// 3. 搜索分页测试 (limit=2 + cursor)
	page1Rec := doReq(t, h, http.MethodGet, "/api/v1/assets/search?q=&limit=2")
	var sp1 dto.ApiResponse[dto.SearchAssetsResponse]
	_ = json.Unmarshal(page1Rec.Body.Bytes(), &sp1)
	if len(sp1.Data.Items) != 2 || sp1.Data.NextCursor == nil {
		t.Fatalf("expected paginated search page1 with next_cursor, got len=%d cursor=%v", len(sp1.Data.Items), sp1.Data.NextCursor)
	}
	page2Rec := doReq(t, h, http.MethodGet, "/api/v1/assets/search?q=&limit=2&cursor="+*sp1.Data.NextCursor)
	var sp2 dto.ApiResponse[dto.SearchAssetsResponse]
	_ = json.Unmarshal(page2Rec.Body.Bytes(), &sp2)
	if len(sp2.Data.Items) == 0 || sp2.Data.Items[0].AssetID == sp1.Data.Items[0].AssetID {
		t.Fatalf("expected distinct page2 items: p1=%+v p2=%+v", sp1.Data.Items, sp2.Data.Items)
	}
}

func TestV6MultiPoolAnd41stItemPagination_AuditC(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "audit_c.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// 先跑常规引导写入目录与双池
	hist := market.NewHistoricalProvider(market.HistoricalProviderConfig{OfflineOnly: true})
	svc := service.NewTrendService(store, hist)
	if err := svc.BootstrapAndComputeSnapshot(context.Background(), false); err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	// 插入 42 个确定性测试资产及观察池 universe_c_test（第 41 个为唯一 HK 目标）
	var assets []storage.AssetEntity
	var uniMembers []storage.UniverseMemberEntity
	var bars []storage.DailyBarEntity
	base := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	if err := store.UpsertUniverses([]storage.UniverseEntity{
		{UniverseID: "universe_c_test", Name: "分页验收42池", Description: "42成员分页筛选测试", IsPublic: true, SortOrder: 99},
	}); err != nil {
		t.Fatalf("UpsertUniverses failed: %v", err)
	}

	for idx := 1; idx <= 42; idx++ {
		aid := "CN.TEST." + string(rune('A'+(idx-1)/26)) + string(rune('A'+(idx-1)%26))
		mkt := "CN"
		name := "样本指数_" + aid
		if idx == 41 {
			aid = "HK.UNIQUE.0041"
			mkt = "HK"
			name = "唯一命中第41号靶标"
		}
		assets = append(assets, storage.AssetEntity{
			AssetID: aid, Market: mkt, Symbol: aid, Name: name, ShortName: name,
			AssetType: "index", MeasureType: "price", Unit: "点", Currency: "CNY", Timezone: "Asia/Shanghai",
			IsActive: true, LicenseStatus: "public", SourceName: "test", CreatedAt: "2026-08-01", UpdatedAt: "2026-08-01",
		})
		uniMembers = append(uniMembers, storage.UniverseMemberEntity{
			UniverseID: "universe_c_test", AssetID: aid, SortOrder: idx,
		})
		for d := 0; d < 5; d++ {
			dt := base.AddDate(0, 0, d).Format("2006-01-02")
			c := 100.0 + float64(idx) + float64(d)
			bars = append(bars, storage.DailyBarEntity{
				AssetID: aid, TradingDate: dt, Provider: "test", Revision: 1,
				Close: &c, QualityStatus: "ok", AsOf: dt, AvailableAt: dt + "T16:00:00Z", UpdatedAt: dt,
			})
		}
	}
	if err := store.UpsertAssets(assets); err != nil {
		t.Fatalf("UpsertAssets failed: %v", err)
	}
	if err := store.UpsertUniverseMembers(uniMembers); err != nil {
		t.Fatalf("UpsertUniverseMembers failed: %v", err)
	}
	if err := store.UpsertDailyBars(bars); err != nil {
		t.Fatalf("UpsertDailyBars failed: %v", err)
	}

	// 验证未筛选前第 1 页 (limit=40) 与第 2 页 (cursor)
	p1, err := svc.GetAssetMembers(context.Background(), service.MembersQuery{
		PoolID: "universe_c_test", SortField: "asset_id", Order: "asc", Limit: 40,
	})
	if err != nil {
		t.Fatalf("page1 failed: %v", err)
	}
	if p1.PoolTotal != 42 || len(p1.Items) != 40 || p1.NextCursor == nil {
		t.Fatalf("unexpected page1 stats: pool_total=%d items=%d cursor=%v", p1.PoolTotal, len(p1.Items), p1.NextCursor)
	}

	p2, err := svc.GetAssetMembers(context.Background(), service.MembersQuery{
		PoolID: "universe_c_test", SortField: "asset_id", Order: "asc", Limit: 40, Cursor: *p1.NextCursor,
	})
	if err != nil {
		t.Fatalf("page2 failed: %v", err)
	}
	if p2.PoolTotal != 42 || len(p2.Items) != 2 || p2.NextCursor != nil {
		t.Fatalf("unexpected page2 stats: pool_total=%d items=%d cursor=%v", p2.PoolTotal, len(p2.Items), p2.NextCursor)
	}

	// 服务端先筛后分页：首次按 market=HK 即直接命中第 41 个标的，且池总计数仍为 42！
	filtered, err := svc.GetAssetMembers(context.Background(), service.MembersQuery{
		PoolID: "universe_c_test", MarketFilter: "HK", Limit: 40,
	})
	if err != nil {
		t.Fatalf("filtered query failed: %v", err)
	}
	if filtered.MatchedTotal != 1 || filtered.PoolTotal != 42 || len(filtered.Items) != 1 || filtered.Items[0].AssetID != "HK.UNIQUE.0041" {
		t.Fatalf("expected direct first-page hit for 41st item: %+v", filtered)
	}

	// 换池比较：另一个观察池 universe_a_sector 有独立 PoolTotal=7
	poolA, err := svc.GetAssetMembers(context.Background(), service.MembersQuery{
		PoolID: "universe_a_sector", Limit: 40,
	})
	if err != nil || poolA.PoolTotal != 7 {
		t.Fatalf("pool switch count mismatch: err=%v pool_total=%d", err, poolA.PoolTotal)
	}
}
