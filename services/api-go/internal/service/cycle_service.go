package service

import (
	"context"
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

	bondYield := 1.5803
	for _, m := range macros {
		if m.Code == "CN10Y" {
			bondYield = m.Value
			break
		}
	}
	divYield := 4.39
	spread := math.Round((divYield-bondYield)*10000) / 10000

	return dto.CycleCompassState{
		CurrentPhase:     "弱复苏与海外结构性繁荣分化",
		PhaseCode:        1,
		DominantAsset:    "红利高股息与海外AI核心资产",
		MacroIndicators:  macros,
		EquityBondSpread: spread,
		StylePreference:  "红利防守为主，低区伺机成长",
		Summary:          "国内十年期国债维持低位，红利股息率利差极具性价比；创业板50/红利低波100比值处于24%历史分位中间偏低区。",
	}
}
