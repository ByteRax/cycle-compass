package service

import (
	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/market"
)

type SectorService struct {
	seedProvider *market.SeedProvider
}

func NewSectorService(seedProvider *market.SeedProvider) *SectorService {
	return &SectorService{seedProvider: seedProvider}
}

func (s *SectorService) GetIndustryMetrics() []dto.IndustryMetric {
	return s.seedProvider.GetIndustryMetrics()
}
