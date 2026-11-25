package market

import (
	"context"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"cyclecompass/api/internal/dto"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	TencentAPI = "https://qt.gtimg.cn/q="
	CacheTTL   = 3 * time.Second
)

type cacheEntry struct {
	quote     dto.MarketQuote
	timestamp time.Time
}

type OnlineProvider struct {
	client   *http.Client
	fallback *SeedProvider
	cache    map[string]cacheEntry
	cacheMu  sync.RWMutex
	fetchMu  sync.Mutex
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
		fallback: fallback,
		cache:    make(map[string]cacheEntry),
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
	var fallbackNeeded []string

	for _, sym := range stillMissing {
		tCode := symbolToTencentCode(sym)
		if tCode != "" {
			tencentCodesMap[tCode] = sym
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
