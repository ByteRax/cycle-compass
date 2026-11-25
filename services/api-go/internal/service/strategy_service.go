package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/market"
)

const (
	WeightTech   = 0.50
	WeightSpread = 0.25
	WeightValue  = 0.15
	WeightStyle  = 0.10
)

type StrategyService struct {
	provider     *market.OnlineProvider
	seedProvider *market.SeedProvider
}

func NewStrategyService(provider *market.OnlineProvider, seedProvider *market.SeedProvider) *StrategyService {
	return &StrategyService{
		provider:     provider,
		seedProvider: seedProvider,
	}
}

func (s *StrategyService) GetRotationSignal(ctx context.Context) dto.RotationSignal {
	quotes := s.provider.GetQuotesBatch(ctx, []string{"399673", "930955"})
	var gPrice, dPrice float64
	var obsDate string
	for _, q := range quotes {
		if q.Symbol == "399673" {
			gPrice = q.Price
			if q.AsOf > obsDate {
				obsDate = q.AsOf
			}
		} else if q.Symbol == "930955" {
			dPrice = q.Price
			if q.AsOf > obsDate {
				obsDate = q.AsOf
			}
		}
	}
	if gPrice <= 0 {
		q := s.seedProvider.GetQuote("399673")
		gPrice = q.Price
		if q.AsOf > obsDate {
			obsDate = q.AsOf
		}
	}
	if dPrice <= 0 {
		q := s.seedProvider.GetQuote("930955")
		dPrice = q.Price
		if q.AsOf > obsDate {
			obsDate = q.AsOf
		}
	}
	if obsDate == "" {
		obsDate = time.Now().Format("2006-01-02")
	}

	ratio := 0.0
	if dPrice > 0 {
		ratio = gPrice / dPrice
	}
	histRatios := s.seedProvider.GetCachedHistRatios()

	// 计算 1500 日滚动分位数
	window := 1500
	targetWindow := histRatios
	if len(histRatios) > window {
		targetWindow = histRatios[len(histRatios)-window:]
	}
	count := 0
	for _, r := range targetWindow {
		if r <= ratio {
			count++
		}
	}
	quantilePct := 0.0
	if len(targetWindow) > 0 {
		quantilePct = float64(count) / float64(len(targetWindow)) * 100.0
	}

	var zone, action string
	if ratio < 0.16 {
		zone = "低区"
		action = "低区：周定投创业板50ETF(159949) 1000元；若离开中间区已满一个自然周则触发回补"
	} else if ratio >= 0.28 {
		zone = "高区"
		action = "高区：周定投红利低波100ETF(515100) 1000元；若离开中间区已满一个自然周则触发回补"
	} else {
		zone = "中间区"
		action = "中间区：暂停指数买入，周定投资金1000元存现金池等待突破回补"
	}

	var toHigh, toLow float64
	if ratio > 0 {
		toHigh = (0.28/ratio - 1.0) * 100.0
		toLow = (1.0 - 0.16/ratio) * 100.0
	}

	return dto.RotationSignal{
		Date:        obsDate,
		Ratio:       ratio,
		Zone:        zone,
		QuantilePct: math.Round(quantilePct*100) / 100,
		Action:      action,
		ToHighPct:   math.Round(toHigh*100) / 100,
		ToLowPct:    math.Round(toLow*100) / 100,
	}
}

// calcRSIFromKlines computes standard Wilder/SMA RSI(period) over closing prices.
func calcRSIFromKlines(points []market.KlinePoint, period int) (float64, bool) {
	if len(points) < period+1 {
		return 0, false
	}
	sub := points[len(points)-(period+1):]
	var gainSum, lossSum float64
	for i := 1; i < len(sub); i++ {
		diff := sub[i].Close - sub[i-1].Close
		if diff > 0 {
			gainSum += diff
		} else {
			lossSum -= diff
		}
	}
	if gainSum+lossSum == 0 {
		return 50.0, true
	}
	if lossSum == 0 {
		return 100.0, true
	}
	rs := (gainSum / float64(period)) / (lossSum / float64(period))
	return 100.0 - 100.0/(1.0+rs), true
}

// calcBollingerPctB computes %B = (Close - LowerBand) / (UpperBand - LowerBand) over lookback days (2 std dev).
func calcBollingerPctB(points []market.KlinePoint, lookback int) (float64, bool) {
	if len(points) < lookback {
		return 0, false
	}
	sub := points[len(points)-lookback:]
	var sum float64
	for _, p := range sub {
		sum += p.Close
	}
	mean := sum / float64(lookback)
	var varSum float64
	for _, p := range sub {
		d := p.Close - mean
		varSum += d * d
	}
	std := math.Sqrt(varSum / float64(lookback))
	if std < 1e-9 {
		return 0.5, true
	}
	upper := mean + 2.0*std
	lower := mean - 2.0*std
	lastClose := sub[len(sub)-1].Close
	return (lastClose - lower) / (upper - lower), true
}

func (s *StrategyService) GetScoringDetail(baseAmount float64) dto.ScoreResult {
	if baseAmount <= 0 {
		baseAmount = 200.0
	}

	hasStaleOrMissing := false

	// 1. 技术面真实计算：基于 930955 历史 CSV（RSI70 + 100日布林带位置）
	dKlines := s.seedProvider.GetHistoricalKlines("930955", 1600)
	latestClose := 0.0
	techAsOf := ""
	techStatus := "computed"
	techSource := "本地历史CSV实算(930955)"
	rsi70 := 50.0
	bPct100 := 0.5

	if len(dKlines) > 0 {
		latestClose = dKlines[len(dKlines)-1].Close
		techAsOf = dKlines[len(dKlines)-1].Date
		r, okR := calcRSIFromKlines(dKlines, 70)
		bp, okB := calcBollingerPctB(dKlines, 100)
		if okR {
			rsi70 = math.Round(r*100) / 100
		} else {
			techStatus = "stale"
			hasStaleOrMissing = true
		}
		if okB {
			bPct100 = math.Round(bp*10000) / 10000
		} else {
			techStatus = "stale"
			hasStaleOrMissing = true
		}
	} else {
		techStatus = "unavailable"
		techSource = "缺少930955历史CSV"
		hasStaleOrMissing = true
	}

	tech := 0.45*(100.0-rsi70) + 55.0*(1.0-bPct100)
	tech = math.Max(0.0, math.Min(100.0, tech))

	// 2. 十年期国债收益率真实读取（bond10y.csv）
	bondKlines := s.seedProvider.GetHistoricalKlines("CN10Y", 30)
	bond10y := 0.0
	bondAsOf := ""
	bondStatus := "computed"
	bondSource := "bond10y.csv 最新收盘"
	if len(bondKlines) > 0 {
		bond10y = math.Round(bondKlines[len(bondKlines)-1].Close*10000) / 10000
		bondAsOf = bondKlines[len(bondKlines)-1].Date
	} else {
		bondStatus = "unavailable"
		bondSource = "缺少国债收益率历史CSV"
		hasStaleOrMissing = true
	}

	// 3. 基本面估值因子（PE / PE分位 / 股息率）：当前开源本地未接中证实时全库估值接口，显式标明静态参考样本基准日
	pe := 9.61
	pePercentile := 0.9077
	dividendYield := 4.39
	fundamentalAsOf := "2025-01-24"
	fundamentalSource := "估值样本基线(未接入实时PE/股息源)"
	fundamentalStatus := "static_reference"

	spreadPP := 0.0
	spread := 0.0
	spreadStatus := "static_reference"
	if bondStatus == "unavailable" {
		spreadStatus = "unavailable"
	} else {
		spreadPP = dividendYield - bond10y
		spread = math.Max(0.0, math.Min(100.0, spreadPP/4.0*100.0))
	}

	value := math.Max(0.0, math.Min(100.0, (1.0-pePercentile)*100.0))

	// 4. 风格因子真实计算：基于 399673/930955 历史比值1500日窗口分位
	stylePercentile := 0.50
	styleStatus := "computed"
	styleSource := "创50/红利比值1500日CSV实算"
	styleAsOf := techAsOf
	histRatios := s.seedProvider.GetCachedHistRatios()
	if len(histRatios) >= 2 && techStatus != "unavailable" {
		curRatio := histRatios[len(histRatios)-1]
		win := histRatios
		if len(win) > 1500 {
			win = win[len(win)-1500:]
		}
		cnt := 0
		for _, r := range win {
			if r <= curRatio {
				cnt++
			}
		}
		stylePercentile = math.Round((float64(cnt)/float64(len(win)))*10000) / 10000
	} else {
		styleStatus = "unavailable"
		styleSource = "缺少双指数交集历史CSV"
		hasStaleOrMissing = true
	}
	style := math.Max(0.0, math.Min(100.0, (1.0-stylePercentile)*100.0))

	// 加权折算
	total := (tech*WeightTech + spread*WeightSpread + value*WeightValue + style*WeightStyle) / 5.0

	// 档位判定
	var action string
	var multiplier, amount, reduceRatio float64

	if techStatus == "unavailable" {
		action = "关键日线输入缺失 · 暂停更新定投档位"
	} else {
		switch {
		case total >= 18.0:
			action = "加码·全力"
			multiplier = 5.0
			amount = baseAmount * 5.0
		case total >= 16.0:
			action = "加码·很便宜"
			multiplier = 4.0
			amount = baseAmount * 4.0
		case total >= 14.0:
			action = "加码·便宜"
			multiplier = 3.0
			amount = baseAmount * 3.0
		case total >= 12.0:
			action = "多买；有卖出池则买回"
			multiplier = 2.0
			amount = baseAmount * 2.0
		case total >= 10.0:
			action = "正常定投"
			multiplier = 1.0
			amount = baseAmount * 1.0
		case total >= 8.0:
			action = "少买·偏贵"
			multiplier = 0.5
			amount = baseAmount * 0.5
		case total > 6.0:
			action = "空仓·暂停定投（老仓不动）"
		case total > 4.0:
			action = "空仓+减仓25%（仅首次跨线执行）"
			reduceRatio = 0.25
		case total > 2.0:
			action = "空仓+减仓33%（仅首次跨线执行）"
			reduceRatio = 0.33
		default:
			action = "空仓+减仓50%（仅首次跨线执行）"
			reduceRatio = 0.50
		}
	}

	overallAsOf := techAsOf
	if bondAsOf > overallAsOf {
		overallAsOf = bondAsOf
	}
	dataNote := fmt.Sprintf("技术面与风格分位按历史CSV实算(截至 %s)；国债利率取自 bond10y.csv(%s)；PE与股息率为静态参考样例(%s，尚未接实时估值接口)",
		techAsOf, bondAsOf, fundamentalAsOf)
	if hasStaleOrMissing {
		dataNote = "部分关键计算输入缺失，得分或档位含降级项，已如实标明因子状态"
	}

	factors := []dto.ScoreFactor{
		{
			Key:       "tech",
			Name:      "技术面 (RSI70 + 布林%B)",
			Value:     rsi70,
			Unit:      "RSI",
			WeightPct: WeightTech * 100,
			Score:     math.Round(tech*100) / 100,
			AsOf:      techAsOf,
			Source:    techSource,
			Status:    techStatus,
			Note:      fmt.Sprintf("布林100日位置 %.2f", bPct100),
		},
		{
			Key:       "spread",
			Name:      "股债利差 (红利股息率 - 国债10Y)",
			Value:     math.Round(spreadPP*100) / 100,
			Unit:      "pp",
			WeightPct: WeightSpread * 100,
			Score:     math.Round(spread*100) / 100,
			AsOf:      bondAsOf,
			Source:    fmt.Sprintf("CN10Y(%s) vs 静态股息(%s)", bondSource, fundamentalAsOf),
			Status:    spreadStatus,
			Note:      fmt.Sprintf("CN10Y=%.2f%%", bond10y),
		},
		{
			Key:       "value",
			Name:      "估值 (PE 历史分位)",
			Value:     pe,
			Unit:      "x",
			WeightPct: WeightValue * 100,
			Score:     math.Round(value*100) / 100,
			AsOf:      fundamentalAsOf,
			Source:    fundamentalSource,
			Status:    fundamentalStatus,
			Note:      fmt.Sprintf("PE分位 %.1f%%（静态样例基线）", pePercentile*100),
		},
		{
			Key:       "style",
			Name:      "风格比值 (创50/红利1500日分位)",
			Value:     math.Round(stylePercentile*10000) / 100,
			Unit:      "%",
			WeightPct: WeightStyle * 100,
			Score:     math.Round(style*100) / 100,
			AsOf:      styleAsOf,
			Source:    styleSource,
			Status:    styleStatus,
			Note:      "比值处于低分位有利于偏防御风格",
		},
	}

	return dto.ScoreResult{
		Tech:            math.Round(tech*100) / 100,
		Spread:          math.Round(spread*100) / 100,
		Value:           math.Round(value*100) / 100,
		Style:           math.Round(style*100) / 100,
		SpreadPP:        math.Round(spreadPP*10000) / 10000,
		Total:           math.Round(total*100) / 100,
		Action:          action,
		Multiplier:      multiplier,
		Amount:          amount,
		ReduceRatio:     reduceRatio,
		Close:           math.Round(latestClose*100) / 100,
		AsOf:            overallAsOf,
		HasStaleInput:   hasStaleOrMissing,
		PE:              pe,
		PEPercentile:    pePercentile,
		DividendYield:   dividendYield,
		Bond10Y:         bond10y,
		RSI70:           rsi70,
		BPct100:         bPct100,
		StylePercentile: stylePercentile,
		DataSourceNote:  dataNote,
		RuleSummary:     "红利四维（50%技术 + 25%利差 + 15%估值 + 10%风格）按20分制评估红利低波估值吸引力；实际买卖仍优先服从双指数轮动区域与周度规则。",
		Factors:         factors,
	}
}

func (s *StrategyService) GetCombinedSignal(ctx context.Context) dto.CombinedStrategySignal {
	rot := s.GetRotationSignal(ctx)
	score := s.GetScoringDetail(200.0)

	var action, tips string
	switch rot.Zone {
	case "中间区":
		action = "轮动中间区：红利四维仅作估值观察，本周定投资金存入现金池等待次周突破"
		tips = "当前比值在 0.16~0.28 中间区，维持现金池储备，不盲目抄底。"
	case "低区":
		action = fmt.Sprintf("轮动低区：周定投买入创业板50ETF(159949)；红利四维 %.2f 分作参考", score.Total)
		tips = "成长比值偏低，定投倾向创业板50，按次周触发原则回补现金池。"
	default:
		action = fmt.Sprintf("轮动高区：周定投买入红利低波100ETF(515100)，参考评分投额 %.0f 元", score.Amount)
		tips = "红利相对占优，按打分与周定投纪律配置高股息。"
	}

	return dto.CombinedStrategySignal{
		Date:           rot.Date,
		Rotation:       rot,
		Scoring:        score,
		CombinedAction: action,
		SummaryTips:    tips,
	}
}
