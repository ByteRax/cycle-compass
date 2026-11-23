package market

import (
	"bytes"
	"encoding/csv"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"cyclecompass/api/internal/dto"
)

func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

type KlinePoint struct {
	Date       string
	ParsedDate time.Time
	Close      float64
}

type KnownAssetInfo struct {
	Name             string
	Category         string
	Currency         string
	DefaultPrice     float64
	DefaultChangePct float64
}

var KnownAssets = map[string]KnownAssetInfo{
	// 策略核心指数
	"930955": {Name: "红利低波100", Category: "equity", Currency: "CNY", DefaultPrice: 11878.27},
	"399673": {Name: "创业板50", Category: "equity", Currency: "CNY", DefaultPrice: 2192.80},
	"000300": {Name: "沪深300", Category: "equity", Currency: "CNY", DefaultPrice: 3892.45},
	"000905": {Name: "中证500", Category: "equity", Currency: "CNY", DefaultPrice: 5740.12},
	"000688": {Name: "科创50", Category: "equity", Currency: "CNY", DefaultPrice: 982.30},
	// 全球核心宏观指数与经济数据
	"BRENT":     {Name: "布伦特原油", Category: "global", Currency: "USD", DefaultPrice: 92.63, DefaultChangePct: 0.70},
	"VIX":       {Name: "恐慌指数", Category: "global", Currency: "", DefaultPrice: 35.90, DefaultChangePct: 26.28},
	"DXY":       {Name: "美元指数", Category: "fx", Currency: "USD", DefaultPrice: 103.84, DefaultChangePct: 0.25},
	"US20Y":     {Name: "美债长债", Category: "bond", Currency: "USD", DefaultPrice: 81.45, DefaultChangePct: 1.37},
	"GOLD_OZ":   {Name: "黄金盎司", Category: "commodity", Currency: "USD", DefaultPrice: 3534.64, DefaultChangePct: -2.49},
	"SILVER_OZ": {Name: "白银盎司", Category: "commodity", Currency: "USD", DefaultPrice: 41.18, DefaultChangePct: -3.90},
	"COPPER_HG": {Name: "铜", Category: "commodity", Currency: "USD", DefaultPrice: 554.86, DefaultChangePct: -3.04},
	"NATGAS":    {Name: "天然气", Category: "commodity", Currency: "USD", DefaultPrice: 3.79, DefaultChangePct: 2.99},
	"ICNYXAU":   {Name: "人民币账户黄金", Category: "commodity", Currency: "元/克", DefaultPrice: 628.50, DefaultChangePct: 0.34},
	"SPX":       {Name: "标普500", Category: "global", Currency: "USD", DefaultPrice: 5864.67},
	"NDX":       {Name: "纳斯达克100", Category: "global", Currency: "USD", DefaultPrice: 20450.80},
	"N225":      {Name: "日经225", Category: "global", Currency: "JPY", DefaultPrice: 38283.85},
	"USDCNH":    {Name: "美元离岸人民币", Category: "fx", Currency: "CNY", DefaultPrice: 7.2450},
	// 大宗商品与债券基准
	"GOLD":   {Name: "现货黄金", Category: "commodity", Currency: "USD", DefaultPrice: 2712.50},
	"COPPER": {Name: "COMEX铜", Category: "commodity", Currency: "USD", DefaultPrice: 4.12},
	"CN10Y":  {Name: "中国十年期国债", Category: "bond", Currency: "%", DefaultPrice: 1.5803},
	"US10Y":  {Name: "美国十年期国债", Category: "bond", Currency: "%", DefaultPrice: 4.4120},
	// 日韩宏观与核心产业标的
	"KOSPI":    {Name: "KOSPI", Category: "apac", Currency: "KRW", DefaultPrice: 2510.50, DefaultChangePct: -5.21},
	"KOSDAQ":   {Name: "KOSDAQ", Category: "apac", Currency: "KRW", DefaultPrice: 720.10, DefaultChangePct: -5.71},
	"TOPIX":    {Name: "TOPIX", Category: "apac", Currency: "JPY", DefaultPrice: 2680.30, DefaultChangePct: -3.88},
	"VN_INDEX": {Name: "越南胡志明", Category: "apac", Currency: "VND", DefaultPrice: 1245.20, DefaultChangePct: -0.05},
	"SENSEX":   {Name: "孟买SENSEX", Category: "apac", Currency: "INR", DefaultPrice: 79540.0, DefaultChangePct: 2.30},
	// 亚太汇率
	"CNYKRW": {Name: "人民币/韩元", Category: "fx", Currency: "KRW", DefaultPrice: 201.47, DefaultChangePct: -0.12},
	"CNYJPY": {Name: "人民币/日元", Category: "fx", Currency: "JPY", DefaultPrice: 23.73, DefaultChangePct: -1.02},
	"USDKRW": {Name: "美元/韩元", Category: "fx", Currency: "KRW", DefaultPrice: 1345.48, DefaultChangePct: -0.13},
	"USDJPY": {Name: "美元/日元", Category: "fx", Currency: "JPY", DefaultPrice: 158.53, DefaultChangePct: -0.81},
}

type SeedProvider struct {
	dataDir    string
	mu         sync.RWMutex
	klineCache map[string][]KlinePoint
	histRatios []float64
	ratiosOnce sync.Once
}

func NewSeedProvider(dataDir string) *SeedProvider {
	sp := &SeedProvider{
		dataDir:    dataDir,
		klineCache: make(map[string][]KlinePoint),
	}
	sp.loadCSV("dividend_930955.csv", "930955")
	sp.loadCSV("growth_399673.csv", "399673")
	sp.loadCSV("bond10y.csv", "CN10Y")
	return sp
}

func (s *SeedProvider) loadCSV(filename, cacheKey string) []KlinePoint {
	s.mu.Lock()
	defer s.mu.Unlock()

	if points, ok := s.klineCache[cacheKey]; ok && len(points) > 0 {
		return points
	}

	path := filepath.Join(s.dataDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))

	r := csv.NewReader(bytes.NewReader(content))
	records, err := r.ReadAll()
	if err != nil || len(records) < 2 {
		return nil
	}

	var points []KlinePoint
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < 2 {
			continue
		}
		dateStr := strings.TrimSpace(row[0])
		val, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if err != nil {
			continue
		}
		parsedDate, _ := time.Parse("2006-01-02", dateStr)
		points = append(points, KlinePoint{
			Date:       dateStr,
			ParsedDate: parsedDate,
			Close:      val,
		})
	}

	s.klineCache[cacheKey] = points
	return points
}

func (s *SeedProvider) GetHistoricalKlines(symbol string, days int) []KlinePoint {
	s.mu.RLock()
	points, ok := s.klineCache[symbol]
	s.mu.RUnlock()

	if !ok || len(points) == 0 {
		if symbol == "930955" {
			points = s.loadCSV("dividend_930955.csv", "930955")
		} else if symbol == "399673" {
			points = s.loadCSV("growth_399673.csv", "399673")
		} else if symbol == "CN10Y" || symbol == "bond10y" {
			points = s.loadCSV("bond10y.csv", "CN10Y")
		}
	}

	if len(points) == 0 {
		return nil
	}

	if days <= 0 || days >= len(points) {
		return points
	}
	return points[len(points)-days:]
}

func (s *SeedProvider) GetCachedHistRatios() []float64 {
	s.ratiosOnce.Do(func() {
		gPoints := s.GetHistoricalKlines("399673", 1600)
		dPoints := s.GetHistoricalKlines("930955", 1600)

		dMap := make(map[string]float64, len(dPoints))
		for _, p := range dPoints {
			dMap[p.Date] = p.Close
		}

		var ratios []float64
		for _, g := range gPoints {
			if dClose, ok := dMap[g.Date]; ok && dClose > 0 {
				ratios = append(ratios, g.Close/dClose)
			}
		}

		if len(ratios) == 0 {
			ratios = []float64{0.1846}
		}
		s.histRatios = ratios
	})
	return s.histRatios
}

func (s *SeedProvider) GetQuote(symbol string) dto.MarketQuote {
	norm := strings.ToUpper(strings.TrimSpace(symbol))
	info, exists := KnownAssets[norm]
	if !exists {
		info = KnownAssetInfo{
			Name:         symbol,
			Category:     "equity",
			Currency:     "CNY",
			DefaultPrice: 100.0,
		}
	}

	price := info.DefaultPrice
	changePct := info.DefaultChangePct
	change := math.Round(price*changePct/100.0*10000) / 10000
	asOf := "2026-11-20"

	if symbol == "930955" {
		points := s.GetHistoricalKlines("930955", 2)
		if len(points) > 0 {
			last := points[len(points)-1]
			price = last.Close
			asOf = last.Date
			if len(points) > 1 {
				prev := points[len(points)-2]
				change = price - prev.Close
				if prev.Close > 0 {
					changePct = (change / prev.Close) * 100.0
				}
			}
		}
	} else if symbol == "399673" {
		points := s.GetHistoricalKlines("399673", 2)
		if len(points) > 0 {
			last := points[len(points)-1]
			price = last.Close
			asOf = last.Date
			if len(points) > 1 {
				prev := points[len(points)-2]
				change = price - prev.Close
				if prev.Close > 0 {
					changePct = (change / prev.Close) * 100.0
				}
			}
		}
	} else if symbol == "CN10Y" || symbol == "bond10y" {
		points := s.GetHistoricalKlines("CN10Y", 2)
		if len(points) > 0 {
			last := points[len(points)-1]
			price = last.Close
			asOf = last.Date
			if len(points) > 1 {
				prev := points[len(points)-2]
				change = price - prev.Close
				if prev.Close > 0 {
					changePct = (change / prev.Close) * 100.0
				}
			}
		}
	}

	return dto.MarketQuote{
		Symbol:    symbol,
		Name:      info.Name,
		Category:  info.Category,
		Price:     math.Round(price*10000) / 10000,
		Change:    math.Round(change*10000) / 10000,
		ChangePct: math.Round(changePct*100) / 100,
		Currency:  info.Currency,
		AsOf:      asOf,
		Meta: dto.DataSourceMeta{
			SourceName: "Offline Seed CSV / Snapshot",
			Symbol:     symbol,
			AsOf:       asOf,
			Status:     "offline_seed",
			LatencyMs:  floatPtr(0.0),
			Warning:    strPtr("离线种子快照数据，已按最新归档日期提供"),
			UpdatedAt:  time.Now().Format("2006-01-02T15:04:05.000000"),
		},
	}
}

func (s *SeedProvider) GetQuotesBatch(symbols []string) []dto.MarketQuote {
	res := make([]dto.MarketQuote, 0, len(symbols))
	for _, sym := range symbols {
		res = append(res, s.GetQuote(sym))
	}
	return res
}

func (s *SeedProvider) GetMacroIndicators() []dto.GlobalMacroIndicator {
	nowStr := time.Now().Format("2006-01-02T15:04:05.000000")
	return []dto.GlobalMacroIndicator{
		{
			Code:          "CN10Y",
			Name:          "中国十年期国债收益率",
			Value:         1.5803,
			Unit:          "%",
			Period:        "日度",
			Trend:         "down",
			Percentile10y: floatPtr(0.05),
			AsOf:          "2026-11-20",
			Meta: dto.DataSourceMeta{
				SourceName: "中国债券信息网/种子数据",
				Symbol:     "CN10Y",
				AsOf:       "2026-11-20",
				Status:     "offline_seed",
				UpdatedAt:  nowStr,
			},
		},
		{
			Code:          "US10Y",
			Name:          "美国十年期国债收益率",
			Value:         4.4120,
			Unit:          "%",
			Period:        "日度",
			Trend:         "up",
			Percentile10y: floatPtr(0.82),
			AsOf:          "2026-11-20",
			Meta: dto.DataSourceMeta{
				SourceName: "FRED/种子数据",
				Symbol:     "US10Y",
				AsOf:       "2026-11-20",
				Status:     "offline_seed",
				UpdatedAt:  nowStr,
			},
		},
		{
			Code:          "DXY",
			Name:          "美元指数 (DXY)",
			Value:         106.82,
			Unit:          "点",
			Period:        "实时",
			Trend:         "up",
			Percentile10y: floatPtr(0.88),
			AsOf:          "2026-11-20",
			Meta: dto.DataSourceMeta{
				SourceName: "ICE/种子数据",
				Symbol:     "DXY",
				AsOf:       "2026-11-20",
				Status:     "offline_seed",
				UpdatedAt:  nowStr,
			},
		},
		{
			Code:          "GOLD",
			Name:          "伦敦现货黄金",
			Value:         2712.50,
			Unit:          "美元/盎司",
			Period:        "实时",
			Trend:         "up",
			Percentile10y: floatPtr(0.99),
			AsOf:          "2026-11-20",
			Meta: dto.DataSourceMeta{
				SourceName: "LBMA/种子数据",
				Symbol:     "GOLD",
				AsOf:       "2026-11-20",
				Status:     "offline_seed",
				UpdatedAt:  nowStr,
			},
		},
		{
			Code:          "BRENT",
			Name:          "ICE布伦特原油",
			Value:         75.17,
			Unit:          "美元/桶",
			Period:        "实时",
			Trend:         "neutral",
			Percentile10y: floatPtr(0.48),
			AsOf:          "2026-11-20",
			Meta: dto.DataSourceMeta{
				SourceName: "ICE/种子数据",
				Symbol:     "BRENT",
				AsOf:       "2026-11-20",
				Status:     "offline_seed",
				UpdatedAt:  nowStr,
			},
		},
	}
}

type rawSector struct {
	ID        string
	Name      string
	Icon      string
	ChangePct float64
	Stage     string
	Score     float64
	Catalyst  string
}

var rawSectorsList = []rawSector{
	{"ai_compute", "AI算力", "🧠", 2.58, "繁荣期", 95.0, "前沿大模型持续迭代，算力集群高利用率"},
	{"cpo", "CPO", "💡", 2.17, "繁荣期", 92.5, "1.6T光模块与硅光渗透加速交付"},
	{"semiconductor", "半导体", "🔬", 2.85, "复苏上行", 88.0, "先进制程与汽车芯片需求拉动"},
	{"storage", "存储", "💾", 3.73, "景气高位", 90.0, "HBM与大容量DRAM/NAND价格坚挺"},
	{"datacenter", "数据中心", "🗄️", 1.21, "稳步扩张", 84.0, "智算中心IDC规模化交付上架"},
	{"cloud", "云计算", "☁️", 0.71, "成熟增长", 80.0, "企业AI工作负载云化转型"},
	{"commercial_aerospace", "商业航天", "🚀", 1.30, "早期爆发", 78.0, "商业火箭发射密集，商业化在即"},
	{"satellite", "卫星", "🛰️", 1.37, "早期爆发", 77.0, "低轨互联网星座批量组网"},
	{"robotics", "机器人", "🤖", 1.56, "技术突破", 82.0, "具身智能本体与轻量关节迭代"},
	{"autonomous_driving", "自动驾驶", "🚗", 2.39, "规模试点", 85.0, "端到端智驾与Robotaxi商业化"},
	{"nuclear_power", "核电", "⚛️", 0.58, "稳健扩张", 76.0, "清洁基荷能源与三代四代核电开工"},
	{"power_grid", "电网", "⚡", 0.35, "建设高峰", 79.0, "特高压外送通道与配电网智能化改造"},
	{"military", "军工", "🛡️", 0.56, "周期底部", 72.0, "新型号列装与订单复苏预期"},
	{"new_energy", "新能源", "🔋", 0.86, "出清修复", 68.0, "风光储一体化海外需求稳健"},
	{"solar", "光伏", "☀️", 2.47, "供给见底", 71.0, "产业链价格企稳与落后产能加速出清"},
	{"lithium_battery", "锂电池", "🔌", 1.81, "逐步筑底", 73.0, "储能出货爆发与碳酸锂库存去化"},
	{"petroleum", "石油", "🛢️", 1.43, "高位震荡", 75.0, "OPEC+受控减产与地缘溢价支持"},
	{"natural_gas", "天然气", "🔥", -1.08, "震荡下行", 65.0, "库存相对充裕与季节性淡季"},
	{"copper_metal", "铜 / 有色", "🟠", 0.61, "高位景气", 83.0, "电网改造与AI能耗高用铜驱动"},
	{"gold", "黄金", "🥇", 0.43, "避险繁荣", 89.0, "央行持续购金与全球去美元化叙事"},
	{"banking", "银行金融", "🏦", 0.11, "防守稳定", 69.0, "高股息资产防御与资产质量稳健"},
	{"biomedicine", "生物医药", "💊", -0.09, "估值消化", 67.0, "创新药出海提速但短期震荡"},
	{"consumption", "消费", "🛒", 0.26, "弱复苏", 64.0, "以旧换新政策驱动与性价比消费"},
	{"rare_earth", "稀土", "🧲", 1.37, "底部企稳", 74.0, "出口配额调控与人形机器人磁材需求"},
}

func (s *SeedProvider) GetIndustryMetrics() []dto.IndustryMetric {
	metrics := make([]dto.IndustryMetric, 0, len(rawSectorsList))
	nowStr := time.Now().Format("2006-01-02T15:04:05.000000")
	for _, sec := range rawSectorsList {
		priceIndex := math.Round(100.0*(1.0+sec.ChangePct/100.0)*100) / 100
		metrics = append(metrics, dto.IndustryMetric{
			SectorID:        sec.ID,
			Name:            sec.Name,
			Icon:            sec.Icon,
			Stage:           sec.Stage,
			ProsperityScore: sec.Score,
			PriceIndex:      priceIndex,
			PriceChangePct:  sec.ChangePct,
			Unit:            "点",
			Catalyst:        sec.Catalyst,
			AsOf:            "2026-11-20",
			Meta: dto.DataSourceMeta{
				SourceName: "Global Industry Tracking / 离线快照",
				Symbol:     sec.ID,
				AsOf:       "2026-11-20",
				Status:     "offline_seed",
				UpdatedAt:  nowStr,
			},
		})
	}
	return metrics
}
