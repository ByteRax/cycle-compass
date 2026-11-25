package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"cyclecompass/api/internal/market"
	"cyclecompass/api/internal/service"
	"cyclecompass/api/internal/storage"
)

func TestPortfolioValuationSeparation_Audit04(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "portfolio_audit04.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// SeedProvider without online access -> ETF 515100/159949 quotes are unavailable
	seedP := market.NewSeedProvider(t.TempDir())
	onlineP := market.NewOnlineProvider(seedP)
	svc := service.NewPortfolioService(store, onlineP)

	// Cancelled context prevents network fetch, forcing offline fallback
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ov, err := svc.GetPortfolioOverview(ctx, "ACC_DEMO_001")
	if err != nil {
		t.Fatalf("GetPortfolioOverview failed: %v", err)
	}

	if !ov.IsDemo {
		t.Fatalf("expected IsDemo=true for ACC_DEMO_001")
	}
	if len(ov.Positions) != 2 {
		t.Fatalf("expected 2 ETF positions, got %d", len(ov.Positions))
	}

	p0 := ov.Positions[0]
	if p0.Symbol != "515100" || p0.SignalSymbol != "930955" {
		t.Fatalf("expected trading symbol 515100 & signal_symbol 930955, got symbol=%q signal=%q", p0.Symbol, p0.SignalSymbol)
	}
	if p0.PriceStatus != "last_known" {
		t.Fatalf("expected last_known price status when offline, got %q", p0.PriceStatus)
	}
	// Verify 1 share ETF valuation matches stored last_known price (1.136), NOT index / 10000
	if p0.UnitPrice1Share != 1.136 || p0.ValuationDate != "2026-11-09" {
		t.Fatalf("unexpected 1-share valuation: unit=%.4f date=%q", p0.UnitPrice1Share, p0.ValuationDate)
	}

	p1 := ov.Positions[1]
	if p1.Symbol != "159949" || p1.SignalSymbol != "399673" {
		t.Fatalf("expected trading symbol 159949 & signal_symbol 399673, got symbol=%q signal=%q", p1.Symbol, p1.SignalSymbol)
	}
	if p1.UnitPrice1Share != 0.85 {
		t.Fatalf("expected 1-share price 0.85, got %.4f", p1.UnitPrice1Share)
	}
}
