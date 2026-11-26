package service

import (
	"context"
	"fmt"
	"math"

	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/market"
)

type CycleService struct {
	provider     *market.OnlineProvider
	seedProvider *market.SeedProvider
}

func NewCycleService(provider *market.OnlineProvider, seedProvider *market.SeedProvider) *CycleService {
	return &CycleService{
		provider:     provider,
		seedProvider: seedProvider,
	}
}

func (s *CycleService) GetCycleOverview(ctx context.Context) dto.CycleCompassState {
	macros := s.seedProvider.GetMacroIndicators()

	bondYield := 0.0
	bondDate := ""
	bondSource := "bond10y.csv"
	for _, m := range macros {
		if m.Code == "CN10Y" && m.Value > 0 {
			bondYield = m.Value
			bondDate = m.AsOf
			break
		}
	}
	if s.provider != nil {
		liveBond := s.provider.GetQuote(ctx, "CN10Y")
		if liveBond.Meta.Status == "ok" && liveBond.Price > 0 && liveBond.AsOf >= bondDate {
			bondYield = liveBond.Price
			bondDate = liveBond.AsOf
			if liveBond.Meta.SourceName != "" {
				bondSource = liveBond.Meta.SourceName
			}
			for i := range macros {
				if macros[i].Code == "CN10Y" {
					macros[i].Value = liveBond.Price
					macros[i].AsOf = liveBond.AsOf
					macros[i].Meta = liveBond.Meta
					break
				}
			}
		}
	}
	divYield := 4.39 // 静态股息率样例基线（2025-01-24）
	spread := 0.0
	if bondYield > 0 {
		spread = math.Round((divYield-bondYield)*10000) / 10000
	}

	histRatios := s.seedProvider.GetCachedHistRatios()
	quantilePct := 0.0
	if len(histRatios) > 0 {
		win := histRatios
		if len(win) > 1500 {
			win = win[len(win)-1500:]
		}
		cur := win[len(win)-1]
		cnt := 0
		for _, r := range win {
			if r <= cur {
				cnt++
			}
		}
		quantilePct = math.Round((float64(cnt)/float64(len(win)))*1000) / 10
	}

	dataDate := bondDate
	if dataDate == "" {
		dataDate = "2025-01-24"
	}

	evidence := []string{
		fmt.Sprintf("CN10Y十年期国债收益率 %.2f%%（观测日 %s，数据源 %s）", bondYield, bondDate, bondSource),
		fmt.Sprintf("股债利差 %.2f%%（基于红利低波参考股息率 %.2f%% 与 CN10Y 计算）", spread, divYield),
		fmt.Sprintf("创50/红利比值处于 1500 日 %.1f%% 历史分位（基于历史收盘CSV）", quantilePct),
		"宏观阶段与占优资产文字研判为离线研究归档结论，不随盘中行情自动改写",
	}

	return dto.CycleCompassState{
		CurrentPhase:      "弱复苏与结构性景气分化（研究归档）",
		PhaseCode:         1,
		DominantAsset:     "红利高股息防御 / 结构性AI主线",
		MacroIndicators:   macros,
		EquityBondSpread:  spread,
		StylePreference:   "红利防守为主，等待低区信号配置成长",
		Summary:           fmt.Sprintf("实算CN10Y %.2f%%(%s)，股债利差约 %.2fpp，创50/红利比值位于 %.1f%% 分位；宏观描述部分源自归档研究参考。", bondYield, dataDate, spread, quantilePct),
		DataDate:          dataDate,
		SourceName:        "国债/指数历史CSV + 宏观研究归档参考",
		IsStaticReference: true,
		EvidenceNotes:     evidence,
	}
}
