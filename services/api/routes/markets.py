# -*- coding: utf-8 -*-
"""全球宏观与多资产市场行情接口。"""
from __future__ import annotations

from typing import Optional
from fastapi import APIRouter, Depends, Query
from packages.market_data.base import BaseDataProvider
from packages.market_data.models import MarketQuote
from services.api.deps import get_market_data_provider
from services.api.schemas.response import ApiResponse

router = APIRouter(prefix="/markets", tags=["Markets"])

DEFAULT_WATCHLIST = [
    "930955", "399673", "000300", "000905", "000688",
    "BRENT", "VIX", "DXY", "US20Y", "GOLD_OZ", "ICNYXAU", "SILVER_OZ", "COPPER_HG", "NATGAS",
    "SPX", "NDX", "N225", "USDCNH", "GOLD", "COPPER", "CN10Y", "US10Y",
    "KOSPI", "KOSDAQ", "TOPIX", "VN_INDEX", "SENSEX",
    "CNYKRW", "CNYJPY", "USDKRW", "USDJPY",
]


@router.get("/quotes", response_model=ApiResponse[list[MarketQuote]])
async def get_quotes(
    category: Optional[str] = Query(None, description="资产类别过滤: equity, global, commodity, fx, bond"),
    symbols: Optional[str] = Query(None, description="逗号分隔的代码列表，缺省为全量核心观察池"),
    provider: BaseDataProvider = Depends(get_market_data_provider),
):
    """获取多资产最新报价列表。"""
    sym_list = [s.strip() for s in symbols.split(",") if s.strip()] if symbols else DEFAULT_WATCHLIST
    quotes = await provider.get_quotes_batch(sym_list)

    if category:
        quotes = [q for q in quotes if q.category == category]

    return ApiResponse.success(quotes)


@router.get("/quotes/{symbol}", response_model=ApiResponse[MarketQuote])
async def get_single_quote(
    symbol: str,
    provider: BaseDataProvider = Depends(get_market_data_provider),
):
    """获取单个标的最新行情。"""
    quote = await provider.get_quote(symbol)
    return ApiResponse.success(quote)
