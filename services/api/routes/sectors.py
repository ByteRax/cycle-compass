# -*- coding: utf-8 -*-
"""产业周期景气度与特定价格监测接口（AI、半导体存储、光模块、有色金属等）。"""
from __future__ import annotations

from fastapi import APIRouter, Depends, Response
from packages.market_data.base import BaseDataProvider
from packages.market_data.models import IndustryMetric
from services.api.deps import get_market_data_provider
from services.api.schemas.response import ApiResponse

router = APIRouter(prefix="/sectors", tags=["Sectors"])


@router.get("/metrics", response_model=ApiResponse[list[IndustryMetric]])
async def get_industry_metrics(
    response: Response,
    provider: BaseDataProvider = Depends(get_market_data_provider),
):
    """获取重点产业链景气度评分、价格指数与周期阶段（支持客户端 15s 缓存）。"""
    response.headers["Cache-Control"] = "public, max-age=15"
    metrics = await provider.get_industry_metrics()
    return ApiResponse.success(metrics)
