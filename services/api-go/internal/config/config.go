package config

import (
	"os"
	"strings"
)

type Config struct {
	Port        string
	AppName     string
	AppVersion  string
	CorsOrigins []string
	DataDir     string
	DBPath      string
}

func Load() *Config {
	port := getEnv("PORT", "8002")
	appName := getEnv("APP_NAME", "周期罗盘 CycleCompass API")
	appVersion := getEnv("APP_VERSION", "0.1.0")
	dataDir := getEnv("DATA_DIR", "references/strategy_original/index_rotation/data")
	dbPath := getEnv("DB_PATH", "data/cyclecompass.db")

	corsStr := getEnv("CORS_ORIGINS", "*")
	var origins []string
	if corsStr == "*" {
		origins = []string{"*"}
	} else {
		for _, o := range strings.Split(corsStr, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	return &Config{
		Port:        port,
		AppName:     appName,
		AppVersion:  appVersion,
		CorsOrigins: origins,
		DataDir:     dataDir,
		DBPath:      dbPath,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
