package main

import (
	"context"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cyclecompass/api/internal/config"
	"cyclecompass/api/internal/handler"
	"cyclecompass/api/internal/market"
	"cyclecompass/api/internal/router"
	"cyclecompass/api/internal/service"
	"cyclecompass/api/internal/storage"
)

func main() {
	cfg := config.Load()
	log.Printf("[CycleCompass-Go] 启动中... 监听端口: %s, 数据目录: %s", cfg.Port, cfg.DataDir)

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. 初始化 SQLite 存储
	store, err := storage.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("[CycleCompass-Go] SQLite 初始化失败: %v", err)
	}
	defer store.Close()
	log.Printf("[CycleCompass-Go] SQLite 存储已就绪: %s", cfg.DBPath)

	// 2. 初始化行情提供者
	seedProvider := market.NewSeedProvider(cfg.DataDir)
	onlineProvider := market.NewOnlineProvider(seedProvider)
	histProvider := market.NewHistoricalProvider(market.HistoricalProviderConfig{
		SnapshotDir:  cfg.DataDir,
		SeedFallback: seedProvider,
	})

	// 4. 初始化服务层
	marketSvc := service.NewMarketService(onlineProvider)
	sectorSvc := service.NewSectorService(seedProvider)
	cycleSvc := service.NewCycleService(onlineProvider, seedProvider)
	strategySvc := service.NewStrategyService(onlineProvider, seedProvider)
	portfolioSvc := service.NewPortfolioService(store, onlineProvider)
	backtestSvc := service.NewBacktestService(seedProvider)
	trendSvc := service.NewTrendService(store, histProvider)

	// 3. 异步后台预热（不阻塞服务启动）
	go func() {
		warmupCtx, cancel := context.WithTimeout(rootCtx, 15*time.Second)
		defer cancel()
		log.Println("[CycleCompass-Go] 正在执行后台预热...")
		_ = seedProvider.GetCachedHistRatios()
		_ = onlineProvider.GetQuotesBatch(warmupCtx, service.DefaultWatchlist[:5])
		if err := trendSvc.BootstrapAndComputeSnapshot(warmupCtx, false); err != nil {
			log.Printf("[CycleCompass-Go] V4 本地快照初始化警告: %v", err)
		}
		if err := trendSvc.BootstrapAndComputeSnapshot(warmupCtx, true); err != nil {
			log.Printf("[CycleCompass-Go] V4 在线快照更新警告: %v", err)
		}
		log.Println("[CycleCompass-Go] 后台预热完成，启动行情主动刷新器(2.5s)...")
		onlineProvider.StartAutoRefresh(rootCtx, service.DefaultWatchlist, 2500*time.Millisecond)
	}()

	// 5. 初始化 Handler 层
	rootH := handler.NewRootHandler(cfg, onlineProvider)
	marketH := handler.NewMarketHandler(marketSvc)
	sectorH := handler.NewSectorHandler(sectorSvc)
	cycleH := handler.NewCycleHandler(cycleSvc)
	strategyH := handler.NewStrategyHandler(strategySvc)
	portfolioH := handler.NewPortfolioHandler(portfolioSvc)
	backtestH := handler.NewBacktestHandler(backtestSvc)
	trendH := handler.NewTrendHandler(trendSvc)

	// 6. 组装路由
	r := router.Setup(router.RouterParams{
		Config:    cfg,
		Root:      rootH,
		Market:    marketH,
		Sector:    sectorH,
		Cycle:     cycleH,
		Strategy:  strategyH,
		Portfolio: portfolioH,
		Backtest:  backtestH,
		Trend:     trendH,
	})

	// 启动仅限内网可见的 pprof 性能探针 (127.0.0.1:6060)
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		debugSrv := &http.Server{
			Addr:              "127.0.0.1:6060",
			Handler:           mux,
			ReadHeaderTimeout: 3 * time.Second,
		}
		if err := debugSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[CycleCompass-Go] pprof 监听退出: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// 7. 优雅停机
	go func() {
		log.Printf("[CycleCompass-Go] 服务已成功监听在 http://0.0.0.0:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[CycleCompass-Go] 监听异常退出: %v", err)
		}
	}()

	<-rootCtx.Done()
	stop() // 释放信号监听，允许二次 Ctrl+C 强制退出
	log.Println("[CycleCompass-Go] 收到终止信号，正在优雅关闭服务...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[CycleCompass-Go] 优雅关闭发生错误: %v", err)
	}

	log.Println("[CycleCompass-Go] 服务已安全退出")
}
