package dto

type IndustryMetric struct {
	SectorID        string         `json:"sector_id"`
	Name            string         `json:"name"`
	Icon            string         `json:"icon"`
	Stage           string         `json:"stage"`
	ProsperityScore float64        `json:"prosperity_score"`
	PriceIndex      float64        `json:"price_index"`
	PriceChangePct  float64        `json:"price_change_pct"`
	Unit            string         `json:"unit"`
	Catalyst        string         `json:"catalyst"`
	AsOf            string         `json:"as_of"`
	Meta            DataSourceMeta `json:"meta"`
}
