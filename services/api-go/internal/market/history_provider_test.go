package market

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// D-01: 验证资产目录唯一性、6大分类、6大观察池以及旧 divergence 4个ID映射。
func TestCatalogUniquenessAndLegacyMappings_D01(t *testing.T) {
	if err := ValidateCatalogUniqueness(); err != nil {
		t.Fatalf("ValidateCatalogUniqueness failed: %v", err)
	}

	if len(V4Categories) != 6 {
		t.Fatalf("expected 6 categories, got %d", len(V4Categories))
	}

	requiredUniverses := []string{
		"universe_macro",
		"universe_a_sector",
		"universe_etf",
		"universe_dividend_etf",
		"universe_global",
		"universe_futures",
	}
	for _, uid := range requiredUniverses {
		u, ok := GetUniverse(uid)
		if !ok {
			t.Fatalf("missing required universe %s", uid)
		}
		if len(u.MemberAssetIDs) == 0 {
			t.Fatalf("universe %s has no members", uid)
		}
	}

	// 同资产可属于多个观察池，但在目录中仅存一份
	macroMembers := GetCatalogAssetsByUniverse("universe_macro")
	aSectorMembers := GetCatalogAssetsByUniverse("universe_a_sector")
	if len(macroMembers) == 0 || len(aSectorMembers) == 0 {
		t.Fatalf("expected non-empty members in macro and a_sector universes")
	}

	// 验证旧 divergence 4个ID映射
	legacyIDs := []string{"a_shares", "dividend_etf", "etf_macro", "a_shares_long"}
	for _, id := range legacyIDs {
		m, ok := ResolveLegacyDivergenceID(id)
		if !ok {
			t.Fatalf("missing legacy divergence mapping for %s", id)
		}
		if m.TargetAssetID == "" || m.TargetUniverseID == "" {
			t.Fatalf("incomplete mapping for %s: %+v", id, m)
		}
	}
	longMapping, _ := ResolveLegacyDivergenceID("a_shares_long")
	if !longMapping.IsLongHorizon {
		t.Fatalf("expected a_shares_long to have IsLongHorizon=true")
	}
}

// D-04 & D-05: 使用 httptest mock 东方财富日线接口，校验价格/日期过滤、未来数据拦截及字段口径。
func TestHistoricalProvider_EastmoneyMockAndValidation_D04_D05(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"rc": 0,
			"data": map[string]any{
				"code": "000300",
				"name": "沪深300",
				"klines": []string{
					"2026-11-18,3850.00,3860.50,3870.00,3840.00,100000,380000000,0.5",
					"2026-11-19,3860.50,0,3870.00,3840.00,100000,380000000,0.0",       // 无效：零值收盘价
					"2026-11-19,3860.50,-10.5,3870.00,3840.00,100000,380000000,0.0",   // 无效：负值收盘价
					"2026-11-19,3860.50,3880.10,3890.00,-,120000,-,0.51",              // 有效：low/turnover 缺失应为 nil
					"2026-11-25,3900.00,3910.00,3920.00,3895.00,110000,400000000,0.6", // 未来日期：应被拦截
					"bad-date,3900.00,3910.00,3920.00,3895.00,110000,400000000,0.6",   // 无效日期
				},
			},
		})
	}))
	defer srv.Close()

	fixedNow := time.Date(2026, 11, 20, 16, 0, 0, 0, time.FixedZone("CST", 8*3600))
	provider := NewHistoricalProvider(HistoricalProviderConfig{
		EastmoneyBaseURL: srv.URL,
		MinInterval:      10 * time.Millisecond,
		NowFunc:          func() time.Time { return fixedNow },
	})

	res := provider.FetchDailyHistory(context.Background(), "CN.SSE.000300", 0)
	if res.QualityStatus != "ok" {
		t.Fatalf("expected status ok, got %s (%s)", res.QualityStatus, res.Message)
	}
	if hits.Load() != 1 {
		t.Fatalf("expected 1 upstream hit, got %d", hits.Load())
	}
	if len(res.Bars) != 2 {
		t.Fatalf("expected 2 valid bars after filtering zero/negative/future/bad-date, got %d", len(res.Bars))
	}
	if res.AsOf != "2026-11-19" {
		t.Fatalf("expected AsOf=2026-11-19, got %s", res.AsOf)
	}
	// D-06: 缺失字段必须为 nil，不能为 0
	second := res.Bars[1]
	if second.Low != nil || second.Turnover != nil {
		t.Fatalf("expected nil for missing Low/Turnover, got Low=%v Turnover=%v", second.Low, second.Turnover)
	}
	if second.AvailableAt == "" {
		t.Fatalf("expected non-empty AvailableAt")
	}
}

// D-04 & D-05: 测试腾讯日线接口解析。
func TestHistoricalProvider_TencentMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"code": 0,
			"data": {
				"csi930955": {
					"day": [
						["2026-11-18", "11800.10", "11850.20", "11890.00", "11780.00", "500000"],
						["2026-11-19", "11850.20", "11878.27", "11900.00", "11820.00", "520000"]
					]
				}
			}
		}`))
	}))
	defer srv.Close()

	fixedNow := time.Date(2026, 11, 20, 16, 0, 0, 0, time.FixedZone("CST", 8*3600))
	provider := NewHistoricalProvider(HistoricalProviderConfig{
		TencentBaseURL: srv.URL,
		MinInterval:    5 * time.Millisecond,
		NowFunc:        func() time.Time { return fixedNow },
	})

	res := provider.FetchDailyHistory(context.Background(), "CN.CSI.930955", 0)
	if res.QualityStatus != "ok" {
		t.Fatalf("expected ok, got %s (%s)", res.QualityStatus, res.Message)
	}
	if len(res.Bars) != 2 || res.AsOf != "2026-11-19" {
		t.Fatalf("unexpected bars or AsOf: len=%d asOf=%s", len(res.Bars), res.AsOf)
	}
}

// D-06: 验证未覆盖资产、离线快照降级（保留原日期不冒充当前日期）、无快照返回 insufficient。
func TestHistoricalProvider_NotCoveredAndOfflineFallback_D06(t *testing.T) {
	var upstreamCalled atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled.Store(true)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	// 写入本地 CSV 快照（包含固定历史日期，绝不与系统当前日期混淆）
	csvContent := "date,close\n2025-12-30,1.6100\n2025-12-31,1.5803\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "bond10y.csv"), []byte(csvContent), 0o600); err != nil {
		t.Fatal(err)
	}

	provider := NewHistoricalProvider(HistoricalProviderConfig{
		EastmoneyBaseURL: srv.URL,
		TencentBaseURL:   srv.URL,
		SnapshotDir:      tmpDir,
		MinInterval:      1 * time.Millisecond,
		NowFunc:          func() time.Time { return time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC) },
	})

	// 1. 未覆盖资产不得发起上游请求，直接返回 not_covered
	btcRes := provider.FetchDailyHistory(context.Background(), "CRYPTO.SPOT.BTC", 100)
	if btcRes.QualityStatus != "not_covered" {
		t.Fatalf("expected not_covered for BTC, got %s", btcRes.QualityStatus)
	}
	if upstreamCalled.Load() {
		t.Fatalf("not_covered asset should not call upstream")
	}

	// 2. 离线/仅快照资产保留原归档日期，status = offline_seed，且 YieldPct 正确填充
	bondRes := provider.FetchDailyHistory(context.Background(), "CN.BOND.CN10Y", 10)
	if bondRes.QualityStatus != "offline_seed" {
		t.Fatalf("expected offline_seed for CN10Y, got %s (%s)", bondRes.QualityStatus, bondRes.Message)
	}
	if bondRes.AsOf != "2025-12-31" {
		t.Fatalf("expected preserved snapshot AsOf=2025-12-31, got %s", bondRes.AsOf)
	}
	if bondRes.Bars[1].YieldPct == nil || math.Abs(*bondRes.Bars[1].YieldPct-1.5803) > 1e-6 {
		t.Fatalf("expected YieldPct=1.5803, got %v", bondRes.Bars[1].YieldPct)
	}

	// 3. 上游失败且无本地快照 -> insufficient（绝不编造随机走势）
	noSnapRes := provider.FetchDailyHistory(context.Background(), "CN.SSE.000905", 10)
	if noSnapRes.QualityStatus != "insufficient" || len(noSnapRes.Bars) != 0 || noSnapRes.AsOf != "" {
		t.Fatalf("expected insufficient with empty bars and empty AsOf, got status=%s len=%d asOf=%s",
			noSnapRes.QualityStatus, len(noSnapRes.Bars), noSnapRes.AsOf)
	}
}

// D-08: 手算链式综合指数测试（基期 100）及覆盖率不足 minCoverage 返回 nil + reason。
func TestBuildCompositeIndex_HandCalculatedAndCoverage_D08(t *testing.T) {
	bars := map[string][]DailyBar{
		"A": {
			{TradingDate: "2026-11-17", Close: 100.0},
			{TradingDate: "2026-11-18", Close: 110.0}, // +10%
			{TradingDate: "2026-11-19", Close: 115.5}, // +5%
		},
		"B": {
			{TradingDate: "2026-11-17", Close: 200.0},
			{TradingDate: "2026-11-18", Close: 190.0},  // -5%
			{TradingDate: "2026-11-19", Close: 194.75}, // +2.5%
		},
	}
	weights := map[string]float64{
		"A": 0.60,
		"B": 0.40,
	}

	pts, reason := BuildCompositeIndex(bars, weights, 0.90)
	if reason != "" || len(pts) != 3 {
		t.Fatalf("unexpected composite index error: reason=%s len=%d", reason, len(pts))
	}

	// t0: 100.0
	// t1: 100 * (1 + 0.6*0.10 + 0.4*(-0.05)) = 100 * 1.04 = 104.0
	// t2: 104 * (1 + 0.6*0.05 + 0.4*0.025) = 104 * 1.04 = 108.16
	expected := []float64{100.0, 104.0, 108.16}
	for i, exp := range expected {
		if math.Abs(pts[i].Value-exp) > 1e-9 {
			t.Fatalf("day %d: expected %.6f, got %.6f", i, exp, pts[i].Value)
		}
		if pts[i].Coverage != 1.0 {
			t.Fatalf("day %d: expected coverage 1.0, got %.4f", i, pts[i].Coverage)
		}
	}

	// 覆盖率不足 0.90 时（A 占 0.60，B 缺最后一天 -> 覆盖率 0.60 < 0.90）必须返回 nil + reason
	barsLowCov := map[string][]DailyBar{
		"A": bars["A"],
		"B": bars["B"][:2],
	}
	lowPts, lowReason := BuildCompositeIndex(barsLowCov, weights, 0.90)
	if lowPts != nil || lowReason == "" {
		t.Fatalf("expected nil + reason when coverage < 0.90, got pts=%v reason=%q", lowPts, lowReason)
	}
}
