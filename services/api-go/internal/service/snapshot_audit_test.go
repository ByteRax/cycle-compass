package service_test

import (
	"context"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cyclecompass/api/internal/market"
	"cyclecompass/api/internal/service"
	"cyclecompass/api/internal/storage"
)

func TestStrategyFactorsAndCycleMetadata_Audit05_06(t *testing.T) {
	dataDir := filepath.Join("..", "..", "..", "..", "references", "strategy_original", "index_rotation", "data")
	seedP := market.NewSeedProvider(dataDir)
	onlineP := market.NewOnlineProvider(seedP)

	stratSvc := service.NewStrategyService(onlineP, seedP)
	score := stratSvc.GetScoringDetail(200.0)
	if len(score.Factors) != 4 {
		t.Fatalf("expected 4 structured factors, got %d", len(score.Factors))
	}
	var techFound, valueFound bool
	for _, f := range score.Factors {
		if f.Key == "tech" {
			techFound = true
			if f.Status != "computed" || f.AsOf == "" {
				t.Fatalf("expected computed tech factor with real CSV date, got %+v", f)
			}
		}
		if f.Key == "value" {
			valueFound = true
			if f.Status != "static_reference" || f.AsOf != "2025-01-24" {
				t.Fatalf("expected static_reference value factor with fixed reference date, got %+v", f)
			}
		}
	}
	if !techFound || !valueFound {
		t.Fatalf("missing required tech/value factors in %+v", score.Factors)
	}

	// Missing CSV inputs must explicitly flag HasStaleInput instead of faking daily official quote
	emptySeed := market.NewSeedProvider(t.TempDir())
	emptyStrat := service.NewStrategyService(market.NewOnlineProvider(emptySeed), emptySeed)
	staleScore := emptyStrat.GetScoringDetail(200.0)
	if !staleScore.HasStaleInput {
		t.Fatalf("expected HasStaleInput=true when CSVs missing")
	}

	cycleSvc := service.NewCycleService(onlineP, seedP)
	cy := cycleSvc.GetCycleOverview(context.Background())
	if !cy.IsStaticReference || cy.DataDate == "" || len(cy.EvidenceNotes) == 0 {
		t.Fatalf("expected structured cycle metadata and evidence notes, got %+v", cy)
	}

	sectorSvc := service.NewSectorService(seedP)
	sectors := sectorSvc.GetIndustryMetrics()
	if len(sectors) == 0 || !sectors[0].IsReference || sectors[0].DataDate == "" {
		t.Fatalf("expected reference metadata on industry metrics, got %+v", sectors[0])
	}
}

func TestSnapshotRevisionFreezeAndCurveFX_Audit07_11(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "audit07_11.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// Seed US.INDEX.NDX (USD) and CN.SSE.000300 (CNY, benchmark) and FX USD/CNY bars
	var bars []storage.DailyBarEntity
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		d := base.AddDate(0, 0, i).Format("2006-01-02")
		ndxClose := 100.0 + float64(i) // rev 1: 100..109
		benchClose := 3000.0 + float64(i*10)
		fxRate := 7.0 + float64(i)*0.1 // USD/CNY rising 7.0..7.9
		bars = append(bars,
			storage.DailyBarEntity{
				AssetID: "US.INDEX.NDX", TradingDate: d, Provider: "test", Revision: 1,
				Close: &ndxClose, QualityStatus: "offline_seed", AsOf: d, AvailableAt: d + "T16:00:00Z", UpdatedAt: d,
			},
			storage.DailyBarEntity{
				AssetID: "CN.SSE.000300", TradingDate: d, Provider: "test", Revision: 1,
				Close: &benchClose, QualityStatus: "offline_seed", AsOf: d, AvailableAt: d + "T16:00:00Z", UpdatedAt: d,
			},
			storage.DailyBarEntity{
				AssetID: "GLOBAL.FX.USDCNH", TradingDate: d, Provider: "test", Revision: 1,
				Close: &fxRate, QualityStatus: "offline_seed", AsOf: d, AvailableAt: d + "T16:00:00Z", UpdatedAt: d,
			},
		)
	}
	if err := store.UpsertDailyBars(bars); err != nil {
		t.Fatalf("UpsertDailyBars rev 1 failed: %v", err)
	}

	hist := market.NewHistoricalProvider(market.HistoricalProviderConfig{OfflineOnly: true})
	svc := service.NewTrendService(store, hist)
	ctx := context.Background()

	if err := svc.BootstrapAndComputeSnapshot(ctx, false); err != nil {
		t.Fatalf("Bootstrap v1 failed: %v", err)
	}
	snap1, err := svc.GetLatestSnapshot(ctx, "all", "ORIGINAL", "latest_available")
	if err != nil {
		t.Fatalf("GetLatestSnapshot v1 failed: %v", err)
	}

	// Verify FX conversion USD -> CNY on curve (AUDIT 11)
	curveCNY, err := svc.GetTrendCurve(ctx, "US.INDEX.NDX", "1M", "daily", "CNY", "CN.SSE.000300", snap1.SnapshotID, "", "")
	if err != nil {
		t.Fatalf("GetTrendCurve CNY failed: %v", err)
	}
	if !curveCNY.FXApplied || curveCNY.EffectiveCurrency != "CNY" || curveCNY.FXPair != "USD/CNY" {
		t.Fatalf("expected USD/CNY FX conversion on curve, got applied=%v eff=%q pair=%q", curveCNY.FXApplied, curveCNY.EffectiveCurrency, curveCNY.FXPair)
	}
	// First point: 100 * 7.0 = 700
	if curveCNY.Points[0].Value == nil || math.Abs(*curveCNY.Points[0].Value-700.0) > 0.01 {
		t.Fatalf("expected first converted point 700.0, got %v", curveCNY.Points[0].Value)
	}

	// Verify unsupported currency HKD (no HKD rate in store) falls back honestly with warning
	curveHKD, err := svc.GetTrendCurve(ctx, "US.INDEX.NDX", "1M", "daily", "HKD", "", snap1.SnapshotID, "", "")
	if err != nil {
		t.Fatalf("GetTrendCurve HKD failed: %v", err)
	}
	if curveHKD.FXApplied || curveHKD.EffectiveCurrency != "USD" || len(curveHKD.Snapshot.Warnings) == 0 {
		t.Fatalf("expected fallback to USD with warning for missing HKD rate, got applied=%v eff=%q warn=%v", curveHKD.FXApplied, curveHKD.EffectiveCurrency, curveHKD.Snapshot.Warnings)
	}

	// Mutate historical revision on date 0 to revision=2 and publish snap2 (AUDIT 07)
	time.Sleep(2 * time.Millisecond) // ensure unique UnixNano snapshotID
	rev2Val := 999.0
	rev2Bench := 9999.0
	d0 := "2026-10-01"
	if err := store.UpsertDailyBars([]storage.DailyBarEntity{
		{
			AssetID: "US.INDEX.NDX", TradingDate: d0, Provider: "test", Revision: 2,
			Close: &rev2Val, QualityStatus: "ok", AsOf: d0, AvailableAt: d0 + "T16:00:00Z", UpdatedAt: d0,
		},
		{
			AssetID: "CN.SSE.000300", TradingDate: d0, Provider: "test", Revision: 2,
			Close: &rev2Bench, QualityStatus: "ok", AsOf: d0, AvailableAt: d0 + "T16:00:00Z", UpdatedAt: d0,
		},
	}); err != nil {
		t.Fatalf("UpsertDailyBars rev 2 failed: %v", err)
	}
	if err := svc.BootstrapAndComputeSnapshot(ctx, false); err != nil {
		t.Fatalf("Bootstrap v2 failed: %v", err)
	}

	// Re-open old snap1 curve (ORIGINAL currency) -> MUST stay frozen at 100.0 (and benchmark 3000.0), NOT 999.0!
	reopenedSnap1, err := svc.GetTrendCurve(ctx, "US.INDEX.NDX", "1M", "daily", "ORIGINAL", "CN.SSE.000300", snap1.SnapshotID, "", "")
	if err != nil {
		t.Fatalf("reopen snap1 failed: %v", err)
	}
	if reopenedSnap1.Points[0].Value == nil || math.Abs(*reopenedSnap1.Points[0].Value-100.0) > 0.01 {
		t.Fatalf("frozen snap1 mutated after rev2! got %.2f, want 100.00", *reopenedSnap1.Points[0].Value)
	}
	if len(reopenedSnap1.BenchmarkPoints) == 0 || reopenedSnap1.BenchmarkPoints[0].Value == nil || math.Abs(*reopenedSnap1.BenchmarkPoints[0].Value-3000.0) > 0.01 {
		t.Fatalf("frozen snap1 benchmark mutated after rev2! got %+v", reopenedSnap1.BenchmarkPoints)
	}

	// Latest snapshot MUST reflect revision 2 (999.0)
	latestCurve, err := svc.GetTrendCurve(ctx, "US.INDEX.NDX", "1M", "daily", "ORIGINAL", "CN.SSE.000300", "", "", "")
	if err != nil {
		t.Fatalf("latest curve failed: %v", err)
	}
	if latestCurve.Points[0].Value == nil || math.Abs(*latestCurve.Points[0].Value-999.0) > 0.01 {
		t.Fatalf("expected latest curve to reflect rev2 (999.0), got %v", latestCurve.Points[0].Value)
	}
	_ = strings.Contains
}
