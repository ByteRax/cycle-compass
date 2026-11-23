# -*- coding: utf-8 -*-
"""系统健康检查与数据源运行状态监控。"""
from __future__ import annotations

from datetime import datetime
from fastapi import APIRouter, Depends
from packages.market_data.base import BaseDataProvider
from packages.market_data.models import DataQualityStatus
from services.api.deps import get_market_data_provider
from services.api.schemas.response import ApiResponse

router = APIRouter(prefix="/health", tags=["System Health"])


@router.get("", response_model=ApiResponse[dict])
async def check_health(provider: BaseDataProvider = Depends(get_market_data_provider)):
    """检查后端服务与数据源运行健康度。"""
    # 探活核心标的
    quote_dividend = await provider.get_quote("930955")
    quote_growth = await provider.get_quote("399673")

    status_code = "healthy"
    if quote_dividend.meta.status == DataQualityStatus.ERROR:
        status_code = "degraded"

    return ApiResponse.success({
        "status": status_code,
        "timestamp": datetime.now().isoformat(),
        "services": {
            "api": "ok",
            "strategy_engine": "ok",
            "backtesting_engine": "ok",
            "market_data": quote_dividend.meta.status.value,
        },
        "monitored_sources": {
            "dividend_930955": {
                "source": quote_dividend.meta.source_name,
                "as_of": quote_dividend.as_of,
                "status": quote_dividend.meta.status.value,
            },
            "growth_399673": {
                "source": quote_growth.meta.source_name,
                "as_of": quote_growth.as_of,
                "status": quote_growth.meta.status.value,
            },
        },
    })
