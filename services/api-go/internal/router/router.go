package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"cyclecompass/api/internal/config"
	"cyclecompass/api/internal/handler"
)

type RouterParams struct {
	Config    *config.Config
	Root      *handler.RootHandler
	Market    *handler.MarketHandler
	Sector    *handler.SectorHandler
	Cycle     *handler.CycleHandler
	Strategy  *handler.StrategyHandler
	Portfolio *handler.PortfolioHandler
	Backtest  *handler.BacktestHandler
	Trend     *handler.TrendHandler
}

func getOrHead(r chi.Router, pattern string, h http.HandlerFunc) {
	r.Get(pattern, h)
	r.Head(pattern, h)
}

func Setup(p RouterParams) http.Handler {
	r := chi.NewRouter()

	// 全局高性能中间件
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Gzip 压缩：响应大于 1KB 自动压缩，压缩等级 5
	r.Use(middleware.Compress(5))
	// ETag 协商缓存 (304 Not Modified)
	r.Use(ETagMiddleware)

	// CORS 中间件：按配置白名单输出；通配符 * 时关闭 Credentials，符合浏览器安全规范
	allowAllOrigins := false
	allowedOriginSet := make(map[string]struct{}, len(p.Config.CorsOrigins))
	for _, o := range p.Config.CorsOrigins {
		if o == "*" {
			allowAllOrigins = true
		} else {
			allowedOriginSet[o] = struct{}{}
		}
	}

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				allowedHeaders := "Accept, Authorization, Content-Type, X-CSRF-Token, Version, version, AppKey, appkey, X-Request-Id"
				if reqHeaders := r.Header.Get("Access-Control-Request-Headers"); reqHeaders != "" {
					allowedHeaders = reqHeaders
				}
				if allowAllOrigins {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
					w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
					w.Header().Set("Access-Control-Expose-Headers", "ETag, X-Request-Id")
					w.Header().Set("Access-Control-Max-Age", "86400")
				} else if _, ok := allowedOriginSet[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
					w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Expose-Headers", "ETag, X-Request-Id")
					w.Header().Set("Access-Control-Max-Age", "86400")
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// 根路径
	getOrHead(r, "/", p.Root.HandleRoot)

	// API v1 路由组
	r.Route("/api/v1", func(api chi.Router) {
		getOrHead(api, "/health", p.Root.HandleHealth)

		// Markets
		api.Route("/markets", func(m chi.Router) {
			getOrHead(m, "/quotes", p.Market.GetQuotes)
			getOrHead(m, "/quotes/{symbol}", p.Market.GetSingleQuote)
		})

		// Sectors
		api.Route("/sectors", func(s chi.Router) {
			getOrHead(s, "/metrics", p.Sector.GetIndustryMetrics)
		})

		// Cycles
		api.Route("/cycles", func(c chi.Router) {
			getOrHead(c, "/overview", p.Cycle.GetCycleOverview)
		})

		// Strategies
		api.Route("/strategies", func(st chi.Router) {
			getOrHead(st, "/rotation/signal", p.Strategy.GetRotationSignal)
			getOrHead(st, "/scoring/detail", p.Strategy.GetScoringDetail)
			getOrHead(st, "/combined", p.Strategy.GetCombinedSignal)
		})

		// Portfolios
		api.Route("/portfolios", func(pt chi.Router) {
			getOrHead(pt, "/overview", p.Portfolio.GetPortfolioOverview)
		})

		// Backtests
		api.Route("/backtests", func(b chi.Router) {
			b.Post("/run", p.Backtest.RunBacktest)
		})

		// V4 Assets & Trends
		if p.Trend != nil {
			api.Route("/assets", func(a chi.Router) {
				getOrHead(a, "/catalog", p.Trend.GetCatalog)
				getOrHead(a, "/universes", p.Trend.GetUniverses)
				getOrHead(a, "/search", p.Trend.SearchAssets)
				getOrHead(a, "/{id}", p.Trend.GetAssetDetail)
				getOrHead(a, "/{id}/members", p.Trend.GetAssetMembers)
			})

			api.Route("/trends", func(tr chi.Router) {
				getOrHead(tr, "/snapshots/latest", p.Trend.GetLatestSnapshot)
				getOrHead(tr, "/methodology", p.Trend.GetMethodology)
				getOrHead(tr, "/data-status", p.Trend.GetDataStatus)
				getOrHead(tr, "/{id}/curve", p.Trend.GetTrendCurve)
				getOrHead(tr, "/{id}/annual", p.Trend.GetTrendAnnual)
				getOrHead(tr, "/{id}/components", p.Trend.GetTrendComponents)
				getOrHead(tr, "/{id}/dashboard", p.Trend.GetTrendDashboard)
				getOrHead(tr, "/{id}/season", p.Trend.GetTrendSeason)
				getOrHead(tr, "/{id}/distribution", p.Trend.GetTrendDistribution)
				getOrHead(tr, "/{id}/tier", p.Trend.GetTrendTier)
				getOrHead(tr, "/{id}/events", p.Trend.GetTrendEvents)
			})
		}
	})

	return r
}
