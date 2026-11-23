# -*- coding: utf-8 -*-
"""行情与宏观产业数据模型，支持三级数据质量状态声明。"""
from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Any, Optional
from pydantic import BaseModel, Field


class DataQualityStatus(str, Enum):
    OK = "ok"                     # 主源实时/日级最新数据
    DEGRADED = "degraded"         # 备用源降级（可能有短暂延迟或次优口径）
    OFFLINE_SEED = "offline_seed" # 离线种子快照降级（网络不可用时使用本地存档）
    ERROR = "error"               # 数据不可用或校验异常


class DataSourceMeta(BaseModel):
    """数据源元数据，坚决不造假，公开透明。"""
    source_name: str
    symbol: str
    as_of: str
    status: DataQualityStatus = DataQualityStatus.OK
    latency_ms: Optional[float] = None
    warning: Optional[str] = None
    updated_at: str = Field(default_factory=lambda: datetime.now().isoformat())


class MarketQuote(BaseModel):
    """单个金融资产/指数行情报价。"""
    symbol: str
    name: str
    category: str                   # global, equity, commodity, fx, bond, sector, ai
    price: float
    change: float
    change_pct: float
    open: Optional[float] = None
    high: Optional[float] = None
    low: Optional[float] = None
    volume: Optional[float] = None
    turnover: Optional[float] = None
    currency: str = "CNY"
    as_of: str
    meta: DataSourceMeta


class IndustryMetric(BaseModel):
    """重点产业链景气度与价格监测。"""
    sector_id: str
    name: str                       # e.g., "AI算力与Token", "半导体存储DRAM", "铜/有色金属", "光模块CPO"
    icon: Optional[str] = None      # e.g., "🧠", "💡", "🔬"
    stage: str                      # 周期阶段：复苏/繁荣/滞胀/衰退
    prosperity_score: float         # 景气度指数 0~100
    price_index: float              # 核心价格指数
    price_change_pct: float
    unit: str                       # e.g., "$/M tokens", "$/8Gb", "$/吨"
    catalyst: str                   # 核心催化剂/驱动因素
    as_of: str
    meta: DataSourceMeta


class GlobalMacroIndicator(BaseModel):
    """全球宏观经济与周期观察指标。"""
    code: str                       # e.g., "US10Y", "CN10Y", "DXY", "GOLD", "BRENT"
    name: str
    value: float
    unit: str
    period: str
    trend: str                      # "up", "down", "neutral"
    percentile_10y: Optional[float] = None # 10年分位
    as_of: str
    meta: DataSourceMeta
