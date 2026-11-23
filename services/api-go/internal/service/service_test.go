package service_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/market"
	"cyclecompass/api/internal/service"
	"cyclecompass/api/internal/storage"
)

func TestAllServices(t *testing.T) {
	dataDir := filepath.Join("..", "..", "..", "..", "references", "strategy_original", "index_rotation", "data")
	seedProvider := market.NewSeedProvider(dataDir)
	onlineProvider := market.NewOnlineProvider(seedProvider)

	ctx := context.Background()

	// 1. 测试行情服务
	marketSvc := service.NewMarketService(onlineProvider)
	quote := marketSvc.GetSingleQuote(ctx, "930955")
	if quote.Symbol != "930955" || quote.Price <= 0 {
		t.Fatalf("quote 930955 failed: %+v", quote)
	}

	// 2. 测试宏观周期
	cycleSvc := service.NewCycleService(onlineProvider, seedProvider)
	state := cycleSvc.GetCycleOverview(ctx)
	if state.PhaseCode != 1 || len(state.MacroIndicators) == 0 {
		t.Fatalf("cycle overview failed: %+v", state)
	}

	// 3. 测试策略计算
	stratSvc := service.NewStrategyService(onlineProvider, seedProvider)
	rot := stratSvc.GetRotationSignal(ctx)
	if rot.Zone == "" || rot.Ratio <= 0 {
		t.Fatalf("rotation signal failed: %+v", rot)
	}

	score := stratSvc.GetScoringDetail(200.0)
	if score.Total <= 0 || score.Total > 20.0 {
		t.Fatalf("scoring detail failed: %+v", score)
	}

	// 4. 测试回测引擎
	btSvc := service.NewBacktestService(seedProvider)
	bt, err := btSvc.RunBacktest(dto.BacktestRequest{
		LowLine:      0.16,
		HighLine:     0.28,
		WeeklyAmount: 1000.0,
		Days:         300,
	})
	if err != nil {
		t.Fatalf("backtest run failed: %v", err)
	}
	if bt.Summary.TradingDays != 300 || bt.Summary.TotalInvest != 62000.0 {
		t.Fatalf("backtest summary mismatch: %+v", bt.Summary)
	}
	if bt.Summary.CalmarRatio <= 0 {
		t.Fatalf("expected positive calmar ratio, got: %f", bt.Summary.CalmarRatio)
	}
	if _, err := btSvc.RunBacktest(dto.BacktestRequest{
		LowLine:      0.30,
		HighLine:     0.20,
		WeeklyAmount: 1000.0,
		Days:         300,
	}); err == nil {
		t.Fatalf("expected validation error when low_line >= high_line")
	}

	// 4.1 测试回测 Singleflight 并发防击穿与缓存一致性
	const concurrentCalls = 10
	errCh := make(chan error, concurrentCalls)
	for i := 0; i < concurrentCalls; i++ {
		go func() {
			cbt, err := btSvc.RunBacktest(dto.BacktestRequest{
				LowLine:      0.16,
				HighLine:     0.28,
				WeeklyAmount: 1000.0,
				Days:         300,
			})
			if err != nil {
				errCh <- err
				return
			}
			if cbt.Summary.TradingDays != 300 || cbt.Summary.TotalInvest != 62000.0 {
				errCh <- fmt.Errorf("concurrent result mismatch")
				return
			}
			errCh <- nil
		}()
	}
	for i := 0; i < concurrentCalls; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("concurrent singleflight backtest failed: %v", err)
		}
	}

	// 5. 测试持久化存储
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("init sqlite store failed: %v", err)
	}
	defer store.Close()

	portSvc := service.NewPortfolioService(store, onlineProvider)
	port, err := portSvc.GetPortfolioOverview(ctx, "ACC_DEMO_001")
	if err != nil {
		t.Fatalf("portfolio overview failed: %v", err)
	}
	if port.TotalAsset <= 0 || len(port.Positions) != 2 {
		t.Fatalf("portfolio data mismatch: %+v", port)
	}
}
