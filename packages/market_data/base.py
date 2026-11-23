# -*- coding: utf-8 -*-
"""行情数据提供者抽象基类与接口约定。"""
from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Any, Optional
import pandas as pd
from packages.market_data.models import (
    DataQualityStatus,
    DataSourceMeta,
    GlobalMacroIndicator,
    IndustryMetric,
    MarketQuote,
)


class BaseDataProvider(ABC):
    """所有数据提供者的抽象基类。"""

    @abstractmethod
    async def get_quote(self, symbol: str) -> MarketQuote:
        """获取单个标的实时/最新行情报价。"""
        pass

    @abstractmethod
    async def get_quotes_batch(self, symbols: list[str]) -> list[MarketQuote]:
        """批量获取标的行情报价。"""
        pass

    @abstractmethod
    async def get_macro_indicators(self) -> list[GlobalMacroIndicator]:
        """获取全球宏观指标集合（中美国债、美元指数、大宗商品等）。"""
        pass

    @abstractmethod
    async def get_industry_metrics(self) -> list[IndustryMetric]:
        """获取重点产业链景气度与价格（半导体、存储、AI、光模块、有色等）。"""
        pass

    @abstractmethod
    async def get_historical_klines(self, symbol: str, days: int = 1500) -> pd.DataFrame:
        """获取历史日线序列 DataFrame: ['date', 'close', 'open', 'high', 'low', 'volume']。"""
        pass
