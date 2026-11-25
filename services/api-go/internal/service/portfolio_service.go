package service

import (
	"context"
	"math"
	"strings"

	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/market"
	"cyclecompass/api/internal/storage"
)

type PortfolioService struct {
	store    *storage.SQLiteStore
	provider *market.OnlineProvider
}

func NewPortfolioService(store *storage.SQLiteStore, provider *market.OnlineProvider) *PortfolioService {
	return &PortfolioService{
		store:    store,
		provider: provider,
	}
}

func (s *PortfolioService) GetPortfolioOverview(ctx context.Context, accountID string) (*dto.PortfolioOverview, error) {
	if accountID == "" {
		accountID = "ACC_DEMO_001"
	}

	acc, err := s.store.GetAccount(accountID)
	if err != nil {
		return nil, err
	}

	positions, err := s.store.GetPositions(accountID)
	if err != nil {
		return nil, err
	}

	txs, err := s.store.GetTransactions(accountID, 50)
	if err != nil {
		return nil, err
	}

	var posItems []dto.PositionItem
	var warnings []string
	totPosVal := 0.0
	totCost := 0.0
	latestValuationDate := ""
	hasFallback := false

	for _, p := range positions {
		quote := s.provider.GetQuote(ctx, p.Symbol)
		currPrice := 0.0
		priceStatus := "unavailable"
		valDate := ""
		valSource := ""

		// 严格按交易 ETF 实际报价估值，绝不使用 >=50 除 10000 或信号指数点位估算
		if quote.Price > 0 && quote.Meta.Status != "unavailable" && quote.Meta.Status != "error" {
			currPrice = math.Round(quote.Price*10000) / 10000
			if quote.Meta.Status == "ok" {
				priceStatus = "live"
			} else {
				priceStatus = quote.Meta.Status
			}
			valDate = quote.AsOf
			valSource = quote.Meta.SourceName
			_ = s.store.UpdatePositionLastPrice(accountID, p.Symbol, currPrice, valDate, valSource)
		} else if p.LastPrice > 0 {
			currPrice = math.Round(p.LastPrice*10000) / 10000
			priceStatus = "last_known"
			valDate = p.LastPriceDate
			valSource = p.LastPriceSource
			if valSource == "" {
				valSource = "历史已知估值"
			}
			hasFallback = true
			warnings = append(warnings, p.Name+" 暂未获取最新实时行情，沿用上次有效估值("+valDate+")")
		} else if p.CostPrice > 0 {
			currPrice = math.Round(p.CostPrice*10000) / 10000
			priceStatus = "cost_fallback"
			valSource = "建仓成本暂估（暂无公开行情）"
			hasFallback = true
			warnings = append(warnings, p.Name+" 暂无可用交易报价，按成本价暂估，禁止按未知价模拟成交")
		} else {
			hasFallback = true
			warnings = append(warnings, p.Name+" 暂无可用估值报价")
		}

		if valDate > latestValuationDate {
			latestValuationDate = valDate
		}

		mVal := p.Shares * currPrice
		cVal := p.Shares * p.CostPrice
		pnl := mVal - cVal
		pnlPct := 0.0
		if cVal > 0 {
			pnlPct = (pnl / cVal) * 100.0
		}

		signalSym := strings.TrimSpace(p.SignalSymbol)
		if signalSym == "" {
			switch strings.ToUpper(strings.TrimSpace(p.Symbol)) {
			case "515100":
				signalSym = "930955"
			case "159949":
				signalSym = "399673"
			}
		}

		posItems = append(posItems, dto.PositionItem{
			Symbol:              p.Symbol,
			SignalSymbol:        signalSym,
			Name:                p.Name,
			Shares:              p.Shares,
			CurrentPrice:        currPrice,
			UnitPrice1Share:     currPrice,
			PriceStatus:         priceStatus,
			ValuationDate:       valDate,
			ValuationSource:     valSource,
			MarketValue:         math.Round(mVal*100) / 100,
			CostPrice:           p.CostPrice,
			TotalCost:           math.Round(cVal*100) / 100,
			UnrealizedProfit:    math.Round(pnl*100) / 100,
			UnrealizedProfitPct: math.Round(pnlPct*100) / 100,
		})

		totPosVal += mVal
		totCost += cVal
	}

	availCash := acc.AvailableCash
	cPool := acc.CashPool
	totAsset := totPosVal + availCash + cPool
	totProfit := totAsset - (totCost + availCash + cPool)
	totProfitPct := 0.0
	if totCost > 0 {
		totProfitPct = (totProfit / totCost) * 100.0
	}

	valNote := "基于交易所 ETF 最新行情估值"
	if hasFallback {
		valNote = "部分标的实时行情暂不可用，已按上次有效估值或成本价保留估值（不可用于未报价模拟成交）"
	}

	return &dto.PortfolioOverview{
		AccountID:          acc.AccountID,
		IsDemo:             acc.AccountID == "ACC_DEMO_001",
		TotalAsset:         math.Round(totAsset*100) / 100,
		AvailableCash:      math.Round(availCash*100) / 100,
		CashPool:           math.Round(cPool*100) / 100,
		PositionsValue:     math.Round(totPosVal*100) / 100,
		TotalCost:          math.Round(totCost*100) / 100,
		TotalProfit:        math.Round(totProfit*100) / 100,
		TotalProfitPct:     math.Round(totProfitPct*100) / 100,
		ValuationDate:      latestValuationDate,
		ValuationNote:      valNote,
		Warnings:           warnings,
		Positions:          posItems,
		RecentTransactions: txs,
	}, nil
}
