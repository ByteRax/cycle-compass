package trend

import (
	"math"
	"strings"
	"testing"
)

func fptr(v float64) *float64 {
	return &v
}

func assertNear(t *testing.T, got *float64, want float64, msg string) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s: got nil, want %.4f", msg, want)
	}
	if math.Abs(*got-want) > 0.02 {
		t.Fatalf("%s: got %.4f, want %.4f", msg, *got, want)
	}
}

func TestComputeAssetPeriodMatrix_CalendarRollingZeroAndMissing(t *testing.T) {
	asset := MatrixAssetInput{
		AssetID:     "CN.SSE.000300",
		Market:      "CN",
		MeasureType: "price",
		Currency:    "CNY",
		Timezone:    "Asia/Shanghai",
		SourceName:  "eastmoney",
	}

	bars := []MatrixBarPoint{
		{TradingDate: "2024-12-31", Close: fptr(100)},
		{TradingDate: "2025-12-31", Close: fptr(120)},
		{TradingDate: "2026-09-30", Close: fptr(130)},
		{TradingDate: "2026-10-19", Close: fptr(125)},
		{TradingDate: "2026-10-30", Close: fptr(140)},
		{TradingDate: "2026-11-13", Close: fptr(150)},
		{TradingDate: "2026-11-20", Close: fptr(150)}, // zero WTD change!
	}

	cells := ComputeAssetPeriodMatrix(asset, bars, "CNY", "2026-11-20", nil)

	// WTD: 150 vs 150 -> exact 0.00%, NOT nil
	assertNear(t, cells["wtd"].ReturnPct, 0.0, "wtd zero change")
	if cells["wtd"].StartDate == nil || *cells["wtd"].StartDate != "2026-11-13" {
		t.Fatalf("unexpected wtd start date: %v", cells["wtd"].StartDate)
	}

	// MTD: 150 vs 140 = +7.14%
	assertNear(t, cells["mtd"].ReturnPct, 7.14, "mtd")
	// QTD: 150 vs 130 = +15.38%
	assertNear(t, cells["qtd"].ReturnPct, 15.38, "qtd")
	// YTD: 150 vs 120 = +25.00%
	assertNear(t, cells["ytd"].ReturnPct, 25.00, "ytd")
	// Prev Year: 120 vs 100 = +20.00%
	assertNear(t, cells["prev_year"].ReturnPct, 20.00, "prev_year")
	// Rolling 1M: cutoff 2026-11-20 -> 2026-10-20 holiday fallback -> 2026-10-19 (125) -> 150/125 - 1 = +20.00%
	assertNear(t, cells["rolling_1m"].ReturnPct, 20.00, "rolling_1m holiday fallback")
	if cells["rolling_1m"].StartDate == nil || *cells["rolling_1m"].StartDate != "2026-10-19" {
		t.Fatalf("expected holiday prior bar 2026-10-19, got %v", cells["rolling_1m"].StartDate)
	}
}

func TestComputeAssetPeriodMatrix_FXAndMissingFXAndBonds(t *testing.T) {
	usAsset := MatrixAssetInput{
		AssetID:     "US.NASDAQ.NDX",
		Market:      "US",
		MeasureType: "price",
		Currency:    "USD",
		Timezone:    "America/New_York",
		SourceName:  "eastmoney",
	}
	usBars := []MatrixBarPoint{
		{TradingDate: "2025-12-31", Close: fptr(100)},
		{TradingDate: "2026-10-19", Close: fptr(105)},
		{TradingDate: "2026-11-19", Close: fptr(110)}, // US market T-1 vs 2026-11-20 cutoff
	}

	// USD asset price +10% YTD, USD/CNY depreciated from 7.00 to 7.35 (+5%) -> CNY return = 1.10 * 1.05 - 1 = +15.50%
	fx := NewMatrixFXTable(map[string][]MatrixFXPoint{
		"USD/CNY": {
			{TradingDate: "2025-12-31", Rate: 7.00},
			{TradingDate: "2026-10-19", Rate: 7.20},
			{TradingDate: "2026-11-19", Rate: 7.35},
		},
	})

	cnyCells := ComputeAssetPeriodMatrix(usAsset, usBars, "CNY", "2026-11-20", fx)
	assertNear(t, cnyCells["ytd"].ReturnPct, 15.50, "FX-converted USD->CNY ytd")
	if !cnyCells["ytd"].FXApplied || cnyCells["ytd"].FXPair != "USD/CNY" {
		t.Fatalf("expected FXApplied true with USD/CNY, got %+v", cnyCells["ytd"])
	}

	// Missing HKD FX rate must return nil + explicit reason, never fake pegged fallback
	hkdCells := ComputeAssetPeriodMatrix(usAsset, usBars, "HKD", "2026-11-20", fx)
	if hkdCells["ytd"].ReturnPct != nil || !strings.Contains(hkdCells["ytd"].Reason, "missing_fx_rate") {
		t.Fatalf("expected nil + missing_fx_rate reason, got %+v", hkdCells["ytd"])
	}

	// Bond yield asset must output bp_change only and leave ReturnPct nil
	bond := MatrixAssetInput{
		AssetID:     "CN.BOND.CN10Y",
		Market:      "CN",
		MeasureType: "yield_pct",
		Currency:    "CNY",
		Timezone:    "Asia/Shanghai",
		SourceName:  "seed_csv",
	}
	bondBars := []MatrixBarPoint{
		{TradingDate: "2025-12-31", YieldPct: fptr(2.15), Close: fptr(2.15)},
		{TradingDate: "2026-11-20", YieldPct: fptr(2.05), Close: fptr(2.05)},
	}
	bondCells := ComputeAssetPeriodMatrix(bond, bondBars, "USD", "2026-11-20", fx)
	if bondCells["ytd"].ReturnPct != nil || bondCells["ytd"].DisplayUnit != "bp" || bondCells["ytd"].FXApplied {
		t.Fatalf("unexpected bond ytd cell: %+v", bondCells["ytd"])
	}
	assertNear(t, bondCells["ytd"].BpChange, -10.0, "bond bp_change")
	// Missing prev_year base bar must be nil + reason
	if bondCells["prev_year"].ReturnPct != nil || bondCells["prev_year"].BpChange != nil || bondCells["prev_year"].Reason == "" {
		t.Fatalf("expected nil bond prev_year cell with reason, got %+v", bondCells["prev_year"])
	}
}

func TestComputeAssetPeriodMatrix_StaleSuspended(t *testing.T) {
	asset := MatrixAssetInput{
		AssetID:     "CN.CSI.SUSPENDED",
		Market:      "CN",
		MeasureType: "price",
		Currency:    "CNY",
		Timezone:    "Asia/Shanghai",
		SourceName:  "eastmoney",
	}
	bars := []MatrixBarPoint{
		{TradingDate: "2025-12-31", Close: fptr(100)},
		{TradingDate: "2026-08-01", Close: fptr(110)},
	}
	cells := ComputeAssetPeriodMatrix(asset, bars, "CNY", "2026-11-20", nil)
	if cells["rolling_1m"].ReturnPct != nil || !strings.Contains(cells["rolling_1m"].Reason, "停牌或未更新") {
		t.Fatalf("expected stale suspension reason for rolling_1m, got %+v", cells["rolling_1m"])
	}
}
