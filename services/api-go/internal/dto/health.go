package dto

// HealthSourceMeta 行情源监控状态元数据
type HealthSourceMeta struct {
	Source string `json:"source"`
	AsOf   string `json:"as_of"`
	Status string `json:"status"`
}

// HealthData 健康检查详情响应
type HealthData struct {
	Status           string                      `json:"status"`
	Timestamp        string                      `json:"timestamp"`
	Services         map[string]string           `json:"services"`
	MonitoredSources map[string]HealthSourceMeta `json:"monitored_sources"`
}

// RootResponse 根路径响应
type RootResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Docs    string `json:"docs"`
	Health  string `json:"health"`
}
