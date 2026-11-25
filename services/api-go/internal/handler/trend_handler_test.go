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
