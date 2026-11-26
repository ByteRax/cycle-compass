package market

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"cyclecompass/api/internal/dto"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	TencentAPI       = "https://qt.gtimg.cn/q="
	YahooAPI         = "https://query1.finance.yahoo.com/v8/finance/chart/"
	EastmoneyBondAPI = "https://datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPTA_WEB_TREASURYYIELD&columns=SOLAR_DATE,EMM00166466&pageSize=2&sortColumns=SOLAR_DATE&sortTypes=-1"
	CacheTTL         = 3 * time.Second
)

type cacheEntry struct {
	quote     dto.MarketQuote
	timestamp time.Time
}

type OnlineProvider struct {
	client           *http.Client
	fallback         *SeedProvider
	yahooBaseURL     string
	eastmoneyBondURL string
	cache            map[string]cacheEntry
	cacheMu          sync.RWMutex
	fetchMu          sync.Mutex
}

func NewOnlineProvider(fallback *SeedProvider) *OnlineProvider {
	return &OnlineProvider{
		client: &http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        50,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		fallback:         fallback,
		yahooBaseURL:     YahooAPI,
		eastmoneyBondURL: EastmoneyBondAPI,
		cache:            make(map[string]cacheEntry),
	}
}

func symbolToTencentCode(symbol string) string {
	switch strings.ToUpper(strings.TrimSpace(symbol)) {
	case "930955":
		return "csi930955"
	case "399673":
		return "sz399673"
	case "000300":
		return "sh000300"
	case "000905":
		return "sh000905"
	case "000688":
		return "sh000688"
	case "515100":
		return "sh515100"
	case "159949":
		return "sz159949"
	default:
		return ""
	}
}

func tencentCodeToSymbol(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "csi930955":
		return "930955"
	case "sz399673":
		return "399673"
	case "sh000300":
		return "000300"
	case "sh000905":
		return "000905"
	case "sh000688":
		return "000688"
	case "sh515100":
		return "515100"
	case "sz159949":
		return "159949"
	default:
		return ""
	}
}

func (p *OnlineProvider) parseTencentLine(line string, latencyMs float64) *dto.MarketQuote {
	line = strings.TrimSpace(line)
	if line == "" || !strings.Contains(line, `="`) {
		return nil
	}

	parts := strings.SplitN(line, `="`, 2)
	rawCode := strings.TrimPrefix(parts[0], "v_")
	dataPart := strings.TrimSuffix(parts[1], `";`)
	fields := strings.Split(dataPart, "~")
	if len(fields) < 35 {
		return nil
	}

	symbol := tencentCodeToSymbol(rawCode)
	if symbol == "" {
		symbol = fields[2]
	}
	name := fields[1]
	if info, ok := KnownAssets[symbol]; ok && info.Name != "" {
		name = info.Name
	}

	price, err1 := strconv.ParseFloat(fields[3], 64)
	yesterdayClose, err2 := strconv.ParseFloat(fields[4], 64)
	if err1 != nil || err2 != nil {
		return nil
	}

	change := price - yesterdayClose
	var changePct float64
	if len(fields) > 32 && fields[32] != "" {
		if cp, err := strconv.ParseFloat(fields[32], 64); err == nil {
			changePct = cp
		} else if yesterdayClose > 0 {
			changePct = (change / yesterdayClose) * 100.0
		}
	} else if yesterdayClose > 0 {
		changePct = (change / yesterdayClose) * 100.0
	}

	asOf := time.Now().Format("2006-01-02")
	if len(fields) > 30 && len(fields[30]) >= 8 {
		rawTS := strings.TrimSpace(fields[30])
		if tParsed, err := time.Parse("20060102", rawTS[:8]); err == nil {
			asOf = tParsed.Format("2006-01-02")
		}
	}
	roundedLat := math.Round(latencyMs*10) / 10
	return &dto.MarketQuote{
		Symbol:    symbol,
		Name:      name,
		Category:  "equity",
		Price:     math.Round(price*10000) / 10000,
		Change:    math.Round(change*10000) / 10000,
		ChangePct: math.Round(changePct*100) / 100,
		Currency:  "CNY",
		AsOf:      asOf,
		Meta: dto.DataSourceMeta{
			SourceName: "Tencent Finance Realtime",
			Symbol:     symbol,
			AsOf:       asOf,
			Status:     "ok",
			LatencyMs:  &roundedLat,
			UpdatedAt:  time.Now().Format("2006-01-02T15:04:05.000000"),
		},
	}
}

func (p *OnlineProvider) GetQuote(ctx context.Context, symbol string) dto.MarketQuote {
	quotes := p.GetQuotesBatch(ctx, []string{symbol})
	if len(quotes) > 0 {
		return quotes[0]
	}
	return p.fallback.GetQuote(symbol)
}

func (p *OnlineProvider) GetQuotesBatch(ctx context.Context, symbols []string) []dto.MarketQuote {
	now := time.Now()
	var normSymbols []string
	for _, s := range symbols {
		trimmed := strings.ToUpper(strings.TrimSpace(s))
		if trimmed != "" {
			normSymbols = append(normSymbols, trimmed)
		}
	}

	results := make(map[string]dto.MarketQuote, len(normSymbols))
	var missing []string

	p.cacheMu.RLock()
	for _, sym := range normSymbols {
		if entry, ok := p.cache[sym]; ok && now.Sub(entry.timestamp) < CacheTTL {
			results[sym] = entry.quote
		} else {
			missing = append(missing, sym)
		}
	}
	p.cacheMu.RUnlock()

	if len(missing) == 0 {
		out := make([]dto.MarketQuote, 0, len(normSymbols))
		for _, sym := range normSymbols {
			if q, ok := results[sym]; ok {
				out = append(out, q)
			}
		}
		return out
	}

	p.fetchMu.Lock()
	defer p.fetchMu.Unlock()

	// Double check cache
	now = time.Now()
	var stillMissing []string
	p.cacheMu.RLock()
	for _, sym := range missing {
		if entry, ok := p.cache[sym]; ok && now.Sub(entry.timestamp) < CacheTTL {
			results[sym] = entry.quote
		} else {
			stillMissing = append(stillMissing, sym)
		}
	}
	p.cacheMu.RUnlock()

	if len(stillMissing) == 0 {
		out := make([]dto.MarketQuote, 0, len(normSymbols))
		for _, sym := range normSymbols {
			if q, ok := results[sym]; ok {
				out = append(out, q)
			}
		}
		return out
	}

	tencentCodesMap := make(map[string]string)
	yahooCodesMap := make(map[string]string)
	var needCN10Y bool
	var fallbackNeeded []string

	for _, sym := range stillMissing {
		if tCode := symbolToTencentCode(sym); tCode != "" {
			tencentCodesMap[tCode] = sym
		} else if sym == "CN10Y" {
			needCN10Y = true
		} else if yCode := symbolToYahooCode(sym); yCode != "" {
			yahooCodesMap[sym] = yCode
		} else {
			fallbackNeeded = append(fallbackNeeded, sym)
		}
	}

	if len(tencentCodesMap) > 0 {
		var codesList []string
		for code := range tencentCodesMap {
			codesList = append(codesList, code)
		}
		url := TencentAPI + strings.Join(codesList, ",")

		startT := time.Now()
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err == nil {
			resp, err := p.client.Do(req)
			latency := float64(time.Since(startT).Microseconds()) / 1000.0
			if err == nil && resp.StatusCode == http.StatusOK {
				utf8Reader := transform.NewReader(resp.Body, simplifiedchinese.GB18030.NewDecoder())
				body, _ := io.ReadAll(utf8Reader)
				resp.Body.Close()

				lines := strings.Split(string(body), ";")
				p.cacheMu.Lock()
				for _, line := range lines {
					parsed := p.parseTencentLine(line, latency)
					if parsed != nil {
						p.cache[parsed.Symbol] = cacheEntry{
							quote:     *parsed,
							timestamp: now,
						}
						results[parsed.Symbol] = *parsed
					}
				}
				p.cacheMu.Unlock()
			} else if resp != nil {
				resp.Body.Close()
			}
		}

		for _, sym := range tencentCodesMap {
			if _, ok := results[sym]; !ok {
				fallbackNeeded = append(fallbackNeeded, sym)
			}
		}
	}

	if len(yahooCodesMap) > 0 {
		var wg sync.WaitGroup
		var resMu sync.Mutex
		sem := make(chan struct{}, 6)
		for sym, yCode := range yahooCodesMap {
			wg.Add(1)
			go func(s, yc string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				if parsed := p.fetchYahooQuote(ctx, s, yc); parsed != nil {
					resMu.Lock()
					p.cacheMu.Lock()
					p.cache[s] = cacheEntry{
						quote:     *parsed,
						timestamp: now,
					}
					p.cacheMu.Unlock()
					results[s] = *parsed
					resMu.Unlock()
				}
			}(sym, yCode)
		}
		wg.Wait()

		for sym := range yahooCodesMap {
			if _, ok := results[sym]; !ok {
				fallbackNeeded = append(fallbackNeeded, sym)
			}
		}
	}

	if needCN10Y {
		if parsed := p.fetchEastmoneyCN10Y(ctx); parsed != nil {
			p.cacheMu.Lock()
			p.cache["CN10Y"] = cacheEntry{
				quote:     *parsed,
				timestamp: now,
			}
			p.cacheMu.Unlock()
			results["CN10Y"] = *parsed
		} else {
			fallbackNeeded = append(fallbackNeeded, "CN10Y")
		}
	}

	p.cacheMu.Lock()
	for _, sym := range fallbackNeeded {
		fbQuote := p.fallback.GetQuote(sym)
		fbQuote.Meta.Warning = strPtr("网络请求超时或源站未收录，已自动平滑降级至最新离线种子快照")
		p.cache[sym] = cacheEntry{
			quote:     fbQuote,
			timestamp: now,
		}
		results[sym] = fbQuote
	}
	p.cacheMu.Unlock()

	out := make([]dto.MarketQuote, 0, len(normSymbols))
	for _, sym := range normSymbols {
		if q, ok := results[sym]; ok {
			out = append(out, q)
		}
	}
	return out
}

// isTradingTime 快速判断当前是否处于 A 股交易或盘前时段（工作日 09:10-11:35, 12:55-15:05 北京时间）
func isTradingTime(t time.Time) bool {
	// 转换为东八区
	cst := t.In(time.FixedZone("CST", 8*3600))
	weekday := cst.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	hour, min, _ := cst.Clock()
	timeNum := hour*100 + min

	// 09:10 ~ 11:35 早盘与盘前
	if timeNum >= 910 && timeNum <= 1135 {
		return true
	}
	// 12:55 ~ 15:05 午盘
	if timeNum >= 1255 && timeNum <= 1505 {
		return true
	}
	return false
}

func (p *OnlineProvider) StartAutoRefresh(ctx context.Context, symbols []string, interval time.Duration) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				now := time.Now()
				currentInterval := interval
				if !isTradingTime(now) {
					// 休市或周末自适应降频至 60 秒，大幅减少无效外部网络 I/O 与源站压力
					currentInterval = 60 * time.Second
				}

				refreshCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				p.GetQuotesBatch(refreshCtx, symbols)
				cancel()

				select {
				case <-ctx.Done():
					return
				case <-time.After(currentInterval):
				}
			}
		}
	}()
}

func (p *OnlineProvider) GetFallback() *SeedProvider {
	return p.fallback
}

var yahooSymbolMap = map[string]string{
	"VIX":       "^VIX",
	"BRENT":     "BZ=F",
	"DXY":       "DX-Y.NYB",
	"US20Y":     "TLT",
	"GOLD_OZ":   "GC=F",
	"GOLD":      "GC=F",
	"SILVER_OZ": "SI=F",
	"COPPER_HG": "HG=F",
	"COPPER":    "HG=F",
	"NATGAS":    "NG=F",
	"SPX":       "^GSPC",
	"NDX":       "^NDX",
	"N225":      "^N225",
	"USDCNH":    "CNH=X",
	"US10Y":     "^TNX",
	"KOSPI":     "^KS11",
	"KOSDAQ":    "^KQ11",
	"SENSEX":    "^BSESN",
	"CNYKRW":    "CNYKRW=X",
	"CNYJPY":    "CNYJPY=X",
	"USDKRW":    "KRW=X",
	"USDJPY":    "JPY=X",
}

func symbolToYahooCode(symbol string) string {
	return yahooSymbolMap[strings.ToUpper(strings.TrimSpace(symbol))]
}

type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency                   string   `json:"currency"`
				RegularMarketPrice         *float64 `json:"regularMarketPrice"`
				RegularMarketChangePercent *float64 `json:"regularMarketChangePercent"`
				ChartPreviousClose         *float64 `json:"chartPreviousClose"`
				PreviousClose              *float64 `json:"previousClose"`
				RegularMarketTime          int64    `json:"regularMarketTime"`
				ExchangeTimezoneName       string   `json:"exchangeTimezoneName"`
			} `json:"meta"`
		} `json:"result"`
	} `json:"chart"`
}

func (p *OnlineProvider) fetchYahooQuote(ctx context.Context, symbol, yahooSym string) *dto.MarketQuote {
	base := p.yahooBaseURL
	if base == "" {
		base = YahooAPI
	}
	reqURL := base + url.PathEscape(yahooSym) + "?range=2d&interval=1d"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	startT := time.Now()
	resp, err := p.client.Do(req)
	latency := float64(time.Since(startT).Microseconds()) / 1000.0
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var parsed yahooChartResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&parsed); err != nil {
		return nil
	}
	if len(parsed.Chart.Result) == 0 {
		return nil
	}
	meta := parsed.Chart.Result[0].Meta
	if meta.RegularMarketPrice == nil {
		return nil
	}

	price := *meta.RegularMarketPrice
	var changePct float64
	if meta.RegularMarketChangePercent != nil {
		changePct = *meta.RegularMarketChangePercent
	} else if meta.PreviousClose != nil && *meta.PreviousClose > 0 {
		changePct = ((price - *meta.PreviousClose) / *meta.PreviousClose) * 100.0
	} else if meta.ChartPreviousClose != nil && *meta.ChartPreviousClose > 0 {
		changePct = ((price - *meta.ChartPreviousClose) / *meta.ChartPreviousClose) * 100.0
	}

	var change float64
	if meta.PreviousClose != nil && *meta.PreviousClose > 0 {
		change = price - *meta.PreviousClose
	} else if changePct != 0 {
		prev := price / (1.0 + changePct/100.0)
		change = price - prev
	} else if meta.ChartPreviousClose != nil && *meta.ChartPreviousClose > 0 {
		change = price - *meta.ChartPreviousClose
	}

	asOf := time.Now().Format("2006-01-02")
	if meta.RegularMarketTime > 0 {
		t := time.Unix(meta.RegularMarketTime, 0).UTC()
		if meta.ExchangeTimezoneName != "" {
			if loc, err := time.LoadLocation(meta.ExchangeTimezoneName); err == nil {
				t = t.In(loc)
			}
		}
		asOf = t.Format("2006-01-02")
	}

	info, ok := KnownAssets[symbol]
	name := symbol
	category := "global"
	currency := meta.Currency
	if ok {
		if info.Name != "" {
			name = info.Name
		}
		if info.Category != "" {
			category = info.Category
		}
		currency = info.Currency
	}

	roundedLat := math.Round(latency*10) / 10
	return &dto.MarketQuote{
		Symbol:    symbol,
		Name:      name,
		Category:  category,
		Price:     math.Round(price*10000) / 10000,
		Change:    math.Round(change*10000) / 10000,
		ChangePct: math.Round(changePct*100) / 100,
		Currency:  currency,
		AsOf:      asOf,
		Meta: dto.DataSourceMeta{
			SourceName: "Yahoo Finance",
			Symbol:     symbol,
			AsOf:       asOf,
			Status:     "ok",
			LatencyMs:  &roundedLat,
			UpdatedAt:  time.Now().Format("2006-01-02T15:04:05.000000"),
		},
	}
}

type eastmoneyTreasuryResp struct {
	Success bool `json:"success"`
	Result  struct {
		Data []struct {
			SolarDate   string   `json:"SOLAR_DATE"`
			Yield10Y    *float64 `json:"EMM00166466"`
		} `json:"data"`
	} `json:"result"`
}

func (p *OnlineProvider) fetchEastmoneyCN10Y(ctx context.Context) *dto.MarketQuote {
	reqURL := p.eastmoneyBondURL
	if reqURL == "" {
		reqURL = EastmoneyBondAPI
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil
	}

	startT := time.Now()
	resp, err := p.client.Do(req)
	latency := float64(time.Since(startT).Microseconds()) / 1000.0
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var parsed eastmoneyTreasuryResp
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&parsed); err != nil {
		return nil
	}
	if !parsed.Success || len(parsed.Result.Data) == 0 || parsed.Result.Data[0].Yield10Y == nil {
		return nil
	}

	latest := parsed.Result.Data[0]
	price := *latest.Yield10Y
	asOf := strings.TrimSpace(latest.SolarDate)
	if len(asOf) >= 10 {
		asOf = asOf[:10]
	} else {
		asOf = time.Now().Format("2006-01-02")
	}

	var change, changePct float64
	if len(parsed.Result.Data) > 1 && parsed.Result.Data[1].Yield10Y != nil && *parsed.Result.Data[1].Yield10Y > 0 {
		prev := *parsed.Result.Data[1].Yield10Y
		change = price - prev
		changePct = (change / prev) * 100.0
	}

	info := KnownAssets["CN10Y"]
	roundedLat := math.Round(latency*10) / 10
	return &dto.MarketQuote{
		Symbol:    "CN10Y",
		Name:      info.Name,
		Category:  info.Category,
		Price:     math.Round(price*10000) / 10000,
		Change:    math.Round(change*10000) / 10000,
		ChangePct: math.Round(changePct*100) / 100,
		Currency:  info.Currency,
		AsOf:      asOf,
		Meta: dto.DataSourceMeta{
			SourceName: "Eastmoney Treasury Yield",
			Symbol:     "CN10Y",
			AsOf:       asOf,
			Status:     "ok",
			LatencyMs:  &roundedLat,
			UpdatedAt:  time.Now().Format("2006-01-02T15:04:05.000000"),
		},
	}
}
