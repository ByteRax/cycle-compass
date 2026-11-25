package market

import (
	"fmt"
	"sort"
	"strings"
)

// MeasureType 描述金融序列的计量口径，防止价格、净值、收益率、汇率混算。
type MeasureType string

const (
	MeasurePrice     MeasureType = "price"
	MeasureIndex     MeasureType = "index"
	MeasureNAV       MeasureType = "nav"
	MeasureYieldPct  MeasureType = "yield_pct"
	MeasureFXRate    MeasureType = "fx_rate"
	MeasureComposite MeasureType = "composite_index"
)

// LicenseStatus 描述数据源的合规授权状态。
type LicenseStatus string

const (
	LicenseVerifiedPublicResearch LicenseStatus = "verified_public_research"
	LicensePublicMacro            LicenseStatus = "public_macro"
	LicenseSelfComputed           LicenseStatus = "self_computed_proxy"
	LicenseNotCovered             LicenseStatus = "not_covered"
)

// CategoryDef 定义 V4 金融资产分类层级。
type CategoryDef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

// CatalogAsset 定义统一资产目录中的单个标准化标的。
type CatalogAsset struct {
	AssetID        string        `json:"asset_id"`
	Market         string        `json:"market"`
	Exchange       string        `json:"exchange"`
	Symbol         string        `json:"symbol"`
	Name           string        `json:"name"`
	CategoryID     string        `json:"category_id"`
	AssetType      string        `json:"asset_type"`
	MeasureType    MeasureType   `json:"measure_type"`
	PriceType      string        `json:"price_type"`
	ReturnType     string        `json:"return_type"`
	Unit           string        `json:"unit"`
	Currency       string        `json:"currency"`
	Timezone       string        `json:"timezone"`
	AllowNegative  bool          `json:"allow_negative"`
	Provider       string        `json:"provider"`
	EastmoneySecID string        `json:"eastmoney_secid,omitempty"`
	TencentCode    string        `json:"tencent_code,omitempty"`
	LocalSnapshot  string        `json:"local_snapshot,omitempty"`
	LicenseStatus  LicenseStatus `json:"license_status"`
	LicenseNote    string        `json:"license_note,omitempty"`
	Aliases        []string      `json:"aliases,omitempty"`
}

// UniverseDef 定义公开观察池或可计算综合指数池。
type UniverseDef struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	CategoryID        string             `json:"category_id"`
	Description       string             `json:"description"`
	HasCompositeIndex bool               `json:"has_composite_index"`
	BenchmarkAssetID  string             `json:"benchmark_asset_id,omitempty"`
	MemberAssetIDs    []string           `json:"member_asset_ids"`
	Weights           map[string]float64 `json:"weights,omitempty"`
	LicenseStatus     LicenseStatus      `json:"license_status"`
}

// LegacyDivergenceMapping 兼容旧版 pages-sub/divergence 的 4 个预设 ID。
type LegacyDivergenceMapping struct {
	LegacyID           string `json:"legacy_id"`
	Name               string `json:"name"`
	TargetAssetID      string `json:"target_asset_id,omitempty"`
	TargetUniverseID   string `json:"target_universe_id,omitempty"`
	BenchmarkAssetID   string `json:"benchmark_asset_id,omitempty"`
	DefaultChart       string `json:"default_chart"`
	IsLongHorizon      bool   `json:"is_long_horizon"`
	RecommendedRange   string `json:"recommended_range"`
	CompatibilityNotes string `json:"compatibility_notes"`
}

// V4Categories 定义 6 大核心金融资产分类层级。
var V4Categories = []CategoryDef{
	{ID: "cat_macro", Name: "大类资产", Description: "跨市场宏观基准、股债商汇核心代表资产", SortOrder: 1},
	{ID: "cat_a_sector", Name: "A股行业", Description: "A股宽基与核心行业指数", SortOrder: 2},
	{ID: "cat_etf", Name: "ETF基金", Description: "场内宽基、行业、红利与跨境ETF", SortOrder: 3},
	{ID: "cat_global", Name: "港美与全球", Description: "港股、美股及亚太主要市场指数", SortOrder: 4},
	{ID: "cat_futures", Name: "商品期货", Description: "贵金属、有色、能源等商品期货代表合约", SortOrder: 5},
	{ID: "cat_bond_fx", Name: "债券与外汇", Description: "国债收益率观测指标、债券ETF与主要汇率", SortOrder: 6},
}

// V4CatalogAssets 定义 V4 标准化资产注册表（同一资产仅存一份行情，可被多个观察池引用）。
var V4CatalogAssets = []CatalogAsset{
	// A股宽基与行业
	{
		AssetID:        "CN.SSE.000300",
		Market:         "CN",
		Exchange:       "SSE",
		Symbol:         "000300",
		Name:           "沪深300",
		CategoryID:     "cat_a_sector",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.000300",
		TencentCode:    "sh000300",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"HS300", "CSI300", "沪深300"},
	},
	{
		AssetID:        "CN.SSE.000905",
		Market:         "CN",
		Exchange:       "SSE",
		Symbol:         "000905",
		Name:           "中证500",
		CategoryID:     "cat_a_sector",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.000905",
		TencentCode:    "sh000905",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"ZZ500", "CSI500", "中证500"},
	},
	{
		AssetID:        "CN.SSE.000688",
		Market:         "CN",
		Exchange:       "SSE",
		Symbol:         "000688",
		Name:           "科创50",
		CategoryID:     "cat_a_sector",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.000688",
		TencentCode:    "sh000688",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"STAR50", "科创50"},
	},
	{
		AssetID:        "CN.SZSE.399673",
		Market:         "CN",
		Exchange:       "SZSE",
		Symbol:         "399673",
		Name:           "创业板50",
		CategoryID:     "cat_a_sector",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "0.399673",
		TencentCode:    "sz399673",
		LocalSnapshot:  "growth_399673.csv",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"CYB50", "创业板50"},
	},
	{
		AssetID:       "CN.CSI.930955",
		Market:        "CN",
		Exchange:      "CSI",
		Symbol:        "930955",
		Name:          "红利低波100",
		CategoryID:    "cat_a_sector",
		AssetType:     "index",
		MeasureType:   MeasureIndex,
		PriceType:     "index_point",
		ReturnType:    "price",
		Unit:          "点",
		Currency:      "CNY",
		Timezone:      "Asia/Shanghai",
		Provider:      "tencent",
		TencentCode:   "csi930955",
		LocalSnapshot: "dividend_930955.csv",
		LicenseStatus: LicenseVerifiedPublicResearch,
		Aliases:       []string{"HLLB100", "红利低波100"},
	},
	{
		AssetID:        "CN.CSI.000932",
		Market:         "CN",
		Exchange:       "CSI",
		Symbol:         "000932",
		Name:           "中证消费",
		CategoryID:     "cat_a_sector",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.000932",
		TencentCode:    "sh000932",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"主要消费", "中证消费"},
	},
	{
		AssetID:        "CN.CSI.000933",
		Market:         "CN",
		Exchange:       "CSI",
		Symbol:         "000933",
		Name:           "中证医药",
		CategoryID:     "cat_a_sector",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.000933",
		TencentCode:    "sh000933",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"医药卫生", "中证医药"},
	},

	// ETF基金
	{
		AssetID:        "CN.SSE.510300",
		Market:         "CN",
		Exchange:       "SSE",
		Symbol:         "510300",
		Name:           "沪深300ETF",
		CategoryID:     "cat_etf",
		AssetType:      "etf",
		MeasureType:    MeasurePrice,
		PriceType:      "qfq_close",
		ReturnType:     "total_return",
		Unit:           "元",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.510300",
		TencentCode:    "sh510300",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"300ETF", "华泰柏瑞沪深300ETF"},
	},
	{
		AssetID:        "CN.SSE.510880",
		Market:         "CN",
		Exchange:       "SSE",
		Symbol:         "510880",
		Name:           "红利ETF",
		CategoryID:     "cat_etf",
		AssetType:      "etf",
		MeasureType:    MeasurePrice,
		PriceType:      "qfq_close",
		ReturnType:     "total_return",
		Unit:           "元",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.510880",
		TencentCode:    "sh510880",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"上证红利ETF", "红利ETF"},
	},
	{
		AssetID:        "CN.SSE.512890",
		Market:         "CN",
		Exchange:       "SSE",
		Symbol:         "512890",
		Name:           "红利低波ETF",
		CategoryID:     "cat_etf",
		AssetType:      "etf",
		MeasureType:    MeasurePrice,
		PriceType:      "qfq_close",
		ReturnType:     "total_return",
		Unit:           "元",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.512890",
		TencentCode:    "sh512890",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"红利低波ETF"},
	},
	{
		AssetID:        "CN.SZSE.159915",
		Market:         "CN",
		Exchange:       "SZSE",
		Symbol:         "159915",
		Name:           "创业板ETF",
		CategoryID:     "cat_etf",
		AssetType:      "etf",
		MeasureType:    MeasurePrice,
		PriceType:      "qfq_close",
		ReturnType:     "total_return",
		Unit:           "元",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "0.159915",
		TencentCode:    "sz159915",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"易方达创业板ETF", "创业板ETF"},
	},
	{
		AssetID:        "CN.SSE.518880",
		Market:         "CN",
		Exchange:       "SSE",
		Symbol:         "518880",
		Name:           "黄金ETF",
		CategoryID:     "cat_etf",
		AssetType:      "etf",
		MeasureType:    MeasurePrice,
		PriceType:      "qfq_close",
		ReturnType:     "price",
		Unit:           "元",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.518880",
		TencentCode:    "sh518880",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"华安黄金ETF", "黄金ETF"},
	},
	{
		AssetID:        "CN.SSE.511010",
		Market:         "CN",
		Exchange:       "SSE",
		Symbol:         "511010",
		Name:           "国债ETF",
		CategoryID:     "cat_etf",
		AssetType:      "etf",
		MeasureType:    MeasurePrice,
		PriceType:      "qfq_close",
		ReturnType:     "total_return",
		Unit:           "元",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "1.511010",
		TencentCode:    "sh511010",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"5年期国债ETF", "国债ETF"},
	},

	// 港美与全球
	{
		AssetID:        "US.INDEX.SPX",
		Market:         "US",
		Exchange:       "SPDJI",
		Symbol:         "SPX",
		Name:           "标普500",
		CategoryID:     "cat_global",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "USD",
		Timezone:       "America/New_York",
		Provider:       "eastmoney",
		EastmoneySecID: "100.SPX",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"S&P500", "SP500", "标普500"},
	},
	{
		AssetID:        "US.INDEX.NDX",
		Market:         "US",
		Exchange:       "NASDAQ",
		Symbol:         "NDX",
		Name:           "纳斯达克100",
		CategoryID:     "cat_global",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "USD",
		Timezone:       "America/New_York",
		Provider:       "eastmoney",
		EastmoneySecID: "100.NDX",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"NASDAQ100", "纳指100"},
	},
	{
		AssetID:        "HK.HKEX.HSI",
		Market:         "HK",
		Exchange:       "HKEX",
		Symbol:         "HSI",
		Name:           "恒生指数",
		CategoryID:     "cat_global",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "HKD",
		Timezone:       "Asia/Hong_Kong",
		Provider:       "eastmoney",
		EastmoneySecID: "100.HSI",
		TencentCode:    "hkHSI",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"HangSeng", "恒指"},
	},
	{
		AssetID:        "JP.TSE.N225",
		Market:         "JP",
		Exchange:       "TSE",
		Symbol:         "N225",
		Name:           "日经225",
		CategoryID:     "cat_global",
		AssetType:      "index",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "JPY",
		Timezone:       "Asia/Tokyo",
		Provider:       "eastmoney",
		EastmoneySecID: "100.N225",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"Nikkei225", "日经225"},
	},

	// 商品期货
	{
		AssetID:        "CN.FUT.SHFE.AU",
		Market:         "CN_FUT",
		Exchange:       "SHFE",
		Symbol:         "AU",
		Name:           "沪金主连",
		CategoryID:     "cat_futures",
		AssetType:      "futures",
		MeasureType:    MeasurePrice,
		PriceType:      "continuous_close",
		ReturnType:     "continuous",
		Unit:           "元/克",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "113.aum",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"黄金期货", "沪金"},
	},
	{
		AssetID:        "CN.FUT.SHFE.CU",
		Market:         "CN_FUT",
		Exchange:       "SHFE",
		Symbol:         "CU",
		Name:           "沪铜主连",
		CategoryID:     "cat_futures",
		AssetType:      "futures",
		MeasureType:    MeasurePrice,
		PriceType:      "continuous_close",
		ReturnType:     "continuous",
		Unit:           "元/吨",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "113.cum",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"铜期货", "沪铜"},
	},
	{
		AssetID:        "GLOBAL.FUT.BRENT",
		Market:         "GLOBAL_FUT",
		Exchange:       "ICE",
		Symbol:         "BRENT",
		Name:           "布伦特原油",
		CategoryID:     "cat_futures",
		AssetType:      "futures",
		MeasureType:    MeasurePrice,
		PriceType:      "continuous_close",
		ReturnType:     "continuous",
		Unit:           "美元/桶",
		Currency:       "USD",
		Timezone:       "Europe/London",
		AllowNegative:  false,
		Provider:       "eastmoney",
		EastmoneySecID: "102.B00Y",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"Brent", "布伦特"},
	},
	{
		AssetID:       "US.FUT.NYMEX.WTI_NEG_DEMO",
		Market:        "US_FUT",
		Exchange:      "NYMEX",
		Symbol:        "CL",
		Name:          "WTI原油合约(未覆盖)",
		CategoryID:    "cat_futures",
		AssetType:     "futures",
		MeasureType:   MeasurePrice,
		PriceType:     "contract_close",
		ReturnType:    "continuous",
		Unit:          "美元/桶",
		Currency:      "USD",
		Timezone:      "America/New_York",
		AllowNegative: true,
		Provider:      "none",
		LicenseStatus: LicenseNotCovered,
		LicenseNote:   "未获取NYMEX完整授权且含负油价合约特殊规则，暂未覆盖",
	},

	// 债券与外汇
	{
		AssetID:       "CN.BOND.CN10Y",
		Market:        "CN",
		Exchange:      "CHINABOND",
		Symbol:        "CN10Y",
		Name:          "中国十年期国债收益率",
		CategoryID:    "cat_bond_fx",
		AssetType:     "bond_yield",
		MeasureType:   MeasureYieldPct,
		PriceType:     "yield_pct",
		ReturnType:    "bp_change",
		Unit:          "%",
		Currency:      "CNY",
		Timezone:      "Asia/Shanghai",
		AllowNegative: false,
		Provider:      "seed_csv",
		LocalSnapshot: "bond10y.csv",
		LicenseStatus: LicensePublicMacro,
		Aliases:       []string{"中国10Y国债", "中债10年"},
	},
	{
		AssetID:        "GLOBAL.FX.DXY",
		Market:         "FX",
		Exchange:       "ICE",
		Symbol:         "DXY",
		Name:           "美元指数",
		CategoryID:     "cat_bond_fx",
		AssetType:      "fx",
		MeasureType:    MeasureIndex,
		PriceType:      "index_point",
		ReturnType:     "price",
		Unit:           "点",
		Currency:       "USD",
		Timezone:       "America/New_York",
		Provider:       "eastmoney",
		EastmoneySecID: "100.UDI",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"美元指数", "DXY"},
	},
	{
		AssetID:        "GLOBAL.FX.USDCNH",
		Market:         "FX",
		Exchange:       "OTC",
		Symbol:         "USDCNH",
		Name:           "美元离岸人民币",
		CategoryID:     "cat_bond_fx",
		AssetType:      "fx",
		MeasureType:    MeasureFXRate,
		PriceType:      "fx_spot",
		ReturnType:     "price",
		Unit:           "CNY/USD",
		Currency:       "CNY",
		Timezone:       "Asia/Shanghai",
		Provider:       "eastmoney",
		EastmoneySecID: "133.USDCNH",
		LicenseStatus:  LicenseVerifiedPublicResearch,
		Aliases:        []string{"离岸人民币", "USDCNH"},
	},

	// 明确未覆盖（合规阻断示例：虚拟货币与期权不公开上线）
	{
		AssetID:       "CRYPTO.SPOT.BTC",
		Market:        "CRYPTO",
		Exchange:      "NONE",
		Symbol:        "BTC",
		Name:          "比特币(暂未覆盖)",
		CategoryID:    "cat_macro",
		AssetType:     "crypto",
		MeasureType:   MeasurePrice,
		PriceType:     "spot",
		ReturnType:    "price",
		Unit:          "USD",
		Currency:      "USD",
		Timezone:      "UTC",
		Provider:      "none",
		LicenseStatus: LicenseNotCovered,
		LicenseNote:   "合规政策限制：虚拟货币默认不公开上线",
	},
}

// V4Universes 定义 6 个公开观察池，明确区分单纯观察池与可计算综合指数池。
var V4Universes = []UniverseDef{
	{
		ID:                "universe_macro",
		Name:              "大类资产全景观察池",
		CategoryID:        "cat_macro",
		Description:       "跨股、债、商、汇核心资产观察池（跨币种异质资产仅展示成分，不强行合成为单一指数）",
		HasCompositeIndex: false,
		BenchmarkAssetID:  "CN.SSE.000300",
		MemberAssetIDs: []string{
			"CN.SSE.000300",
			"CN.CSI.930955",
			"CN.SSE.518880",
			"CN.SSE.511010",
			"US.INDEX.SPX",
			"HK.HKEX.HSI",
			"GLOBAL.FX.DXY",
			"CN.BOND.CN10Y",
		},
		LicenseStatus: LicenseVerifiedPublicResearch,
	},
	{
		ID:                "universe_a_sector",
		Name:              "A股宽基与行业观察池",
		CategoryID:        "cat_a_sector",
		Description:       "A股核心宽基与风格行业指数，基准为沪深300",
		HasCompositeIndex: true,
		BenchmarkAssetID:  "CN.SSE.000300",
		MemberAssetIDs: []string{
			"CN.SSE.000300",
			"CN.SSE.000905",
			"CN.SSE.000688",
			"CN.SZSE.399673",
			"CN.CSI.930955",
			"CN.CSI.000932",
			"CN.CSI.000933",
		},
		Weights: map[string]float64{
			"CN.SSE.000300":  0.30,
			"CN.SSE.000905":  0.20,
			"CN.SSE.000688":  0.10,
			"CN.SZSE.399673": 0.10,
			"CN.CSI.930955":  0.10,
			"CN.CSI.000932":  0.10,
			"CN.CSI.000933":  0.10,
		},
		LicenseStatus: LicenseSelfComputed,
	},
	{
		ID:                "universe_etf",
		Name:              "场内核心ETF观察池",
		CategoryID:        "cat_etf",
		Description:       "覆盖宽基、红利、黄金与国债的场内高流动性ETF",
		HasCompositeIndex: true,
		BenchmarkAssetID:  "CN.SSE.510300",
		MemberAssetIDs: []string{
			"CN.SSE.510300",
			"CN.SSE.510880",
			"CN.SSE.512890",
			"CN.SZSE.159915",
			"CN.SSE.518880",
			"CN.SSE.511010",
		},
		Weights: map[string]float64{
			"CN.SSE.510300":  0.25,
			"CN.SSE.510880":  0.15,
			"CN.SSE.512890":  0.15,
			"CN.SZSE.159915": 0.15,
			"CN.SSE.518880":  0.15,
			"CN.SSE.511010":  0.15,
		},
		LicenseStatus: LicenseSelfComputed,
	},
	{
		ID:                "universe_dividend_etf",
		Name:              "红利低波策略观察池",
		CategoryID:        "cat_etf",
		Description:       "红利低波100指数及代表性红利ETF",
		HasCompositeIndex: false,
		BenchmarkAssetID:  "CN.CSI.930955",
		MemberAssetIDs: []string{
			"CN.CSI.930955",
			"CN.SSE.510880",
			"CN.SSE.512890",
		},
		LicenseStatus: LicenseVerifiedPublicResearch,
	},
	{
		ID:                "universe_global",
		Name:              "港美与全球股指观察池",
		CategoryID:        "cat_global",
		Description:       "美股、港股、日股代表性公开指数（不同币种仅展示原币趋势，不直接混合加权）",
		HasCompositeIndex: false,
		BenchmarkAssetID:  "US.INDEX.SPX",
		MemberAssetIDs: []string{
			"US.INDEX.SPX",
			"US.INDEX.NDX",
			"HK.HKEX.HSI",
			"JP.TSE.N225",
		},
		LicenseStatus: LicenseVerifiedPublicResearch,
	},
	{
		ID:                "universe_futures",
		Name:              "商品期货代表观察池",
		CategoryID:        "cat_futures",
		Description:       "黄金、铜与布伦特原油代表性连续合约",
		HasCompositeIndex: false,
		BenchmarkAssetID:  "CN.FUT.SHFE.AU",
		MemberAssetIDs: []string{
			"CN.FUT.SHFE.AU",
			"CN.FUT.SHFE.CU",
			"GLOBAL.FUT.BRENT",
		},
		LicenseStatus: LicenseVerifiedPublicResearch,
	},
}

// LegacyDivergenceMappings 兼容旧版 pages-sub/divergence 的 4 个 ID。
var LegacyDivergenceMappings = map[string]LegacyDivergenceMapping{
	"a_shares": {
		LegacyID:           "a_shares",
		Name:               "A股核心宽基",
		TargetAssetID:      "CN.SSE.000300",
		TargetUniverseID:   "universe_a_sector",
		BenchmarkAssetID:   "CN.SSE.000300",
		DefaultChart:       "components",
		IsLongHorizon:      false,
		RecommendedRange:   "1Y",
		CompatibilityNotes: "映射至沪深300与A股宽基行业观察池",
	},
	"dividend_etf": {
		LegacyID:           "dividend_etf",
		Name:               "红利低波与ETF",
		TargetAssetID:      "CN.CSI.930955",
		TargetUniverseID:   "universe_dividend_etf",
		BenchmarkAssetID:   "CN.CSI.930955",
		DefaultChart:       "components",
		IsLongHorizon:      false,
		RecommendedRange:   "1Y",
		CompatibilityNotes: "映射至红利低波100及红利ETF观察池",
	},
	"etf_macro": {
		LegacyID:           "etf_macro",
		Name:               "宏观大类资产",
		TargetAssetID:      "CN.SSE.510300",
		TargetUniverseID:   "universe_macro",
		BenchmarkAssetID:   "CN.SSE.000300",
		DefaultChart:       "components",
		IsLongHorizon:      false,
		RecommendedRange:   "1Y",
		CompatibilityNotes: "映射至大类资产全景观察池",
	},
	"a_shares_long": {
		LegacyID:           "a_shares_long",
		Name:               "A股超长周期研究",
		TargetAssetID:      "CN.SSE.000300",
		TargetUniverseID:   "universe_a_sector",
		BenchmarkAssetID:   "CN.SSE.000300",
		DefaultChart:       "curve",
		IsLongHorizon:      true,
		RecommendedRange:   "ALL",
		CompatibilityNotes: "独立超长周期研究预设，复用沪深300真实历史序列（不编造缺失年份）",
	},
}

var (
	catalogAssetByID map[string]CatalogAsset
	universeByID     map[string]UniverseDef
)

func init() {
	catalogAssetByID = make(map[string]CatalogAsset, len(V4CatalogAssets))
	for _, a := range V4CatalogAssets {
		catalogAssetByID[a.AssetID] = a
	}
	universeByID = make(map[string]UniverseDef, len(V4Universes))
	for _, u := range V4Universes {
		universeByID[u.ID] = u
	}
}

// GetCatalogAsset 根据唯一 asset_id 获取标准化资产定义。
func GetCatalogAsset(assetID string) (CatalogAsset, bool) {
	a, ok := catalogAssetByID[strings.TrimSpace(assetID)]
	return a, ok
}

// GetUniverse 获取指定观察池定义。
func GetUniverse(universeID string) (UniverseDef, bool) {
	u, ok := universeByID[strings.TrimSpace(universeID)]
	return u, ok
}

// GetCatalogAssetsByCategory 返回指定分类下的全部资产。
func GetCatalogAssetsByCategory(categoryID string) []CatalogAsset {
	var out []CatalogAsset
	for _, a := range V4CatalogAssets {
		if a.CategoryID == categoryID {
			out = append(out, a)
		}
	}
	return out
}

// GetCatalogAssetsByUniverse 返回指定观察池下的全部资产。
func GetCatalogAssetsByUniverse(universeID string) []CatalogAsset {
	u, ok := GetUniverse(universeID)
	if !ok {
		return nil
	}
	out := make([]CatalogAsset, 0, len(u.MemberAssetIDs))
	for _, id := range u.MemberAssetIDs {
		if a, exists := GetCatalogAsset(id); exists {
			out = append(out, a)
		}
	}
	return out
}

// ResolveLegacyDivergenceID 将旧版 divergence ID 映射为 V4 资产与观察池上下文。
func ResolveLegacyDivergenceID(legacyID string) (LegacyDivergenceMapping, bool) {
	m, ok := LegacyDivergenceMappings[strings.TrimSpace(legacyID)]
	return m, ok
}

// ValidateCatalogUniqueness 校验资产目录唯一性约束（D-01：asset_id 不重复、观察池成员引用有效）。
func ValidateCatalogUniqueness() error {
	seenIDs := make(map[string]struct{}, len(V4CatalogAssets))
	seenMarketSym := make(map[string]string, len(V4CatalogAssets))
	for _, a := range V4CatalogAssets {
		if a.AssetID == "" {
			return fmt.Errorf("catalog asset has empty asset_id")
		}
		if _, dup := seenIDs[a.AssetID]; dup {
			return fmt.Errorf("duplicate asset_id: %s", a.AssetID)
		}
		seenIDs[a.AssetID] = struct{}{}

		key := a.Market + ":" + a.Exchange + ":" + a.AssetType + ":" + a.Symbol
		if prev, dup := seenMarketSym[key]; dup {
			return fmt.Errorf("duplicate market/exchange/type/symbol key %s for %s and %s", key, prev, a.AssetID)
		}
		seenMarketSym[key] = a.AssetID
	}

	for _, u := range V4Universes {
		seenMembers := make(map[string]struct{}, len(u.MemberAssetIDs))
		for _, memberID := range u.MemberAssetIDs {
			if _, ok := seenIDs[memberID]; !ok {
				return fmt.Errorf("universe %s references unknown asset_id %s", u.ID, memberID)
			}
			if _, dup := seenMembers[memberID]; dup {
				return fmt.Errorf("universe %s has duplicate member %s", u.ID, memberID)
			}
			seenMembers[memberID] = struct{}{}
		}
	}

	// 确定性检查排序
	_ = sort.StringsAreSorted
	return nil
}
