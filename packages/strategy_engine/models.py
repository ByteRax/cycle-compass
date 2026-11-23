# -*- coding: utf-8 -*-
"""策略核心数据模型，采用 Pydantic v2 强类型与校验。"""
from __future__ import annotations

from datetime import date
from enum import Enum
from typing import Any, Optional
from pydantic import BaseModel, Field, field_validator


class RotationZone(str, Enum):
    LOW = "低区"
    MID = "中间区"
    HIGH = "高区"


class EventType(str, Enum):
    NONE = "none"
    WEEKLY_INVEST = "weekly_invest"           # 周定投买入
    CASH_DEPOSIT = "cash_deposit"             # 中间区定投资金转入现金池
    CASH_POOL_REBUY = "cash_pool_rebuy"       # 离开中间区次周回补买入
    SELL_POOL_BUYBACK = "sell_pool_buyback"   # 评分>=12分卖出池买回
    REDUCE = "reduce"                         # 减仓
    FEE = "fee"                               # 交易手续费


class LedgerEntry(BaseModel):
    """资金记账流水条目，确保资金守恒对账。"""
    date: str
    event_type: EventType
    symbol: Optional[str] = None
    target_name: Optional[str] = None
    shares: float = 0.0
    price: float = 0.0
    amount: float = 0.0                     # 本次操作名义金额
    cash_inflow: float = 0.0                # 现金流入（例如定投注入或卖出回款）
    cash_outflow: float = 0.0               # 现金流出（例如买入扣款或转入现金池）
    cash_pool_delta: float = 0.0            # 现金池变动（正为存入，负为释放）
    cash_balance_after: float = 0.0         # 结算后可用现金
    cash_pool_after: float = 0.0            # 结算后现金池余额
    description: str = ""


class ScoreInput(BaseModel):
    """四维打分输入参数。"""
    pe_percentile: float = Field(..., ge=0.0, le=1.0, description="PE历史分位 [0, 1]")
    dividend_yield: float = Field(..., ge=0.0, description="股息率(%)，例如 4.49 表示 4.49%")
    bond10y: float = Field(..., ge=0.0, description="十年期国债到期收益率(%)，例如 1.58")
    rsi70: float = Field(..., ge=0.0, le=100.0, description="70日Wilder RSI [0, 100]")
    b_pct100: float = Field(..., description="100日布林%B，允许破轨")
    style_percentile: float = Field(..., ge=0.0, le=1.0, description="风格比值历史分位 [0, 1]")


class ScoreResult(BaseModel):
    """打分结果与定投建议。"""
    tech: float
    spread: float
    value: float
    style: float
    spread_pp: float
    total: float
    action: str
    multiplier: float
    amount: float
    reduce_ratio: float


class RotationConfig(BaseModel):
    """比值轮动策略配置。"""
    low_line: float = 0.16
    high_line: float = 0.28
    weekly_amount: float = 1000.0
    fee: float = 0.0005
    div_yield_growth: float = 0.006
    div_yield_dividend: float = 0.05
    cash_rate: float = 0.0
    quantile_window: int = 1500

    @field_validator("high_line")
    @classmethod
    def validate_lines(cls, v: float, info) -> float:
        low = info.data.get("low_line", 0.16)
        if v <= low:
            raise ValueError("high_line 必须大于 low_line")
        return v


class RotationSignal(BaseModel):
    """比值轮动每日信号。"""
    date: str
    ratio: float
    zone: RotationZone
    quantile_pct: float
    action: str
    to_high_pct: float
    to_low_pct: float
