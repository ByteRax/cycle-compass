package service

import (
	"context"
	"math"

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

	txs, err := s.store.GetTransactions(accountID, 20)
	if err != nil {
		return nil, err
	}

	var posItems []dto.PositionItem
	totPosVal := 0.0
	totCost := 0.0

	for _, p := range positions {
		quote := s.provider.GetQuote(ctx, p.Symbol)
		currPrice := quote.Price
		if currPrice >= 50.0 {
			currPrice = currPrice / 10000.0
		}

		mVal := p.Shares * currPrice
		cVal := p.Shares * p.CostPrice
		pnl := mVal - cVal
		pnlPct := 0.0
		if cVal > 0 {
			pnlPct = (pnl / cVal) * 100.0
		}

		posItems = append(posItems, dto.PositionItem{
			Symbol:             p.Symbol,
			Name:               p.Name,
			Shares:             p.Shares,
			CurrentPrice:       math.Round(currPrice*10000) / 10000,
			MarketValue:        math.Round(mVal*100) / 100,
			CostPrice:          p.CostPrice,
			TotalCost:          math.Round(cVal*100) / 100,
			UnrealizedProfit:   math.Round(pnl*100) / 100,
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

	return &dto.PortfolioOverview{
		AccountID:          acc.AccountID,
		TotalAsset:         math.Round(totAsset*100) / 100,
		AvailableCash:      math.Round(availCash*100) / 100,
		CashPool:           math.Round(cPool*100) / 100,
		PositionsValue:     math.Round(totPosVal*100) / 100,
		TotalCost:          math.Round(totCost*100) / 100,
		TotalProfit:        math.Round(totProfit*100) / 100,
		TotalProfitPct:     math.Round(totProfitPct*100) / 100,
		Positions:          posItems,
		RecentTransactions: txs,
	}, nil
}
