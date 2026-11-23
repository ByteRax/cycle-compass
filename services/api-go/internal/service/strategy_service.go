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
	for _, q := range quotes {
		if q.Symbol == "399673" {
			gPrice = q.Price
		} else if q.Symbol == "930955" {
			dPrice = q.Price
		}
	}
	if gPrice <= 0 {
		gPrice = s.seedProvider.GetQuote("399673").Price
	}
	if dPrice <= 0 {
		dPrice = s.seedProvider.GetQuote("930955").Price
	}

	ratio := gPrice / dPrice
	histRatios := s.seedProvider.GetCachedHistRatios()

	// 计算分位数
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
	quantilePct := float64(count) / float64(len(targetWindow)) * 100.0

	var zone, action string
	if ratio < 0.16 {
		zone = "低区"
		action = "低区：周定投创业板50(159949) 1000元；有到期现金池则次周回补"
	} else if ratio >= 0.28 {
		zone = "高区"
		action = "高区：周定投红利低波100(515100) 1000元；有到期现金池则次周回补"
	} else {
		zone = "中间区"
		action = "中间区：暂停指数买入，周定投资金1000元进现金池"
	}

	toHigh := (0.28/ratio - 1.0) * 100.0
	toLow := (1.0 - 0.16/ratio) * 100.0

	return dto.RotationSignal{
		Date:        time.Now().Format("2006-01-02"),
		Ratio:       ratio,
		Zone:        zone,
		QuantilePct: quantilePct,
		Action:      action,
		ToHighPct:   toHigh,
		ToLowPct:    toLow,
	}
}

func (s *StrategyService) GetScoringDetail(baseAmount float64) dto.ScoreResult {
	if baseAmount <= 0 {
		baseAmount = 200.0
	}

	pePercentile := 0.9077
	dividendYield := 4.39
	bond10y := 1.5803
	rsi70 := 53.57
	bPct100 := 0.9620
	stylePercentile := 0.7573

	// 技术面
	tech := 0.45*(100.0-rsi70) + 55.0*(1.0-bPct100)
	tech = math.Max(0.0, math.Min(100.0, tech))

	// 利差
	spreadPP := dividendYield - bond10y
	spread := math.Max(0.0, math.Min(100.0, spreadPP/4.0*100.0))

	// 估值
	value := math.Max(0.0, math.Min(100.0, (1.0-pePercentile)*100.0))

	// 风格
	style := math.Max(0.0, math.Min(100.0, (1.0-stylePercentile)*100.0))

	// 加权折算
	total := (tech*WeightTech + spread*WeightSpread + value*WeightValue + style*WeightStyle) / 5.0

	// 档位判定
	var action string
	var multiplier, amount, reduceRatio float64

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
		Close:           11878.27,
		PE:              9.61,
		PEPercentile:    pePercentile,
		DividendYield:   dividendYield,
		Bond10Y:         bond10y,
		RSI70:           rsi70,
		BPct100:         bPct100,
		StylePercentile: stylePercentile,
		DataSourceNote:  "中证指数官网官方收盘（日频）",
	}
}

func (s *StrategyService) GetCombinedSignal(ctx context.Context) dto.CombinedStrategySignal {
	rot := s.GetRotationSignal(ctx)
	score := s.GetScoringDetail(200.0)

	var action, tips string
	switch rot.Zone {
	case "中间区":
		action = "轮动中间区：红利打分仅作估值观察，不新增指数买入；资金1000元进入现金池等待"
		tips = "当前比值在 0.16~0.28 中间区，维持现金池储备，耐受风格拉锯。"
	case "低区":
		action = fmt.Sprintf("轮动低区：定投买入创业板50；红利四维评分 %.2f 分作为参考", score.Total)
		tips = "成长指数比值极度超跌，周定投倾向进攻标的。"
	default:
		action = fmt.Sprintf("轮动高区：定投买入红利低波100，结合打分建议买入 %.0f 元", score.Amount)
		tips = "红利指数处于相对优势区间，按打分档位执行买入或减仓防守。"
	}

	return dto.CombinedStrategySignal{
		Date:           rot.Date,
		Rotation:       rot,
		Scoring:        score,
		CombinedAction: action,
		SummaryTips:    tips,
	}
}
