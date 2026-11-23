package service

import (
	"context"
	"strings"

	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/market"
)

var DefaultWatchlist = []string{
	"930955", "399673", "000300", "000905", "000688",
	"BRENT", "VIX", "DXY", "US20Y", "GOLD_OZ", "ICNYXAU", "SILVER_OZ", "COPPER_HG", "NATGAS",
	"SPX", "NDX", "N225", "USDCNH", "GOLD", "COPPER", "CN10Y", "US10Y",
	"KOSPI", "KOSDAQ", "TOPIX", "VN_INDEX", "SENSEX",
	"CNYKRW", "CNYJPY", "USDKRW", "USDJPY",
}

type MarketService struct {
	provider *market.OnlineProvider
}

func NewMarketService(provider *market.OnlineProvider) *MarketService {
	return &MarketService{provider: provider}
}

func (s *MarketService) GetQuotes(ctx context.Context, category, symbols string) []dto.MarketQuote {
	var symList []string
	if strings.TrimSpace(symbols) != "" {
		for _, part := range strings.Split(symbols, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				symList = append(symList, trimmed)
			}
		}
	} else {
		symList = DefaultWatchlist
	}

	quotes := s.provider.GetQuotesBatch(ctx, symList)
	if category != "" {
		var filtered []dto.MarketQuote
		for _, q := range quotes {
			if q.Category == category {
				filtered = append(filtered, q)
			}
		}
		return filtered
	}
	return quotes
}

func (s *MarketService) GetSingleQuote(ctx context.Context, symbol string) dto.MarketQuote {
	return s.provider.GetQuote(ctx, symbol)
}
