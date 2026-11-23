# -*- coding: utf-8 -*-
"""宏观经济与资产配置周期观察接口。"""
from __future__ import annotations

from fastapi import APIRouter, Depends, Response
from pydantic import BaseModel
from packages.market_data.base import BaseDataProvider
from packages.market_data.models import GlobalMacroIndicator
from services.api.deps import get_market_data_provider
from services.api.schemas.response import ApiResponse

router = APIRouter(prefix="/cycles", tags=["Cycles"])


class CycleCompassState(BaseModel):
    current_phase: str                # e.g., "过热期向滞胀期过渡", "衰退后期"
    phase_code: int                   # 1: 复苏, 2: 繁荣, 3: 滞胀, 4: 衰退
    dominant_asset: str               # 占优资产类别：股票/商品/现金/债券
    macro_indicators: list[GlobalMacroIndicator]
    equity_bond_spread: float         # 股债收益利差 (百分点)
    style_preference: str             # 当前风格倾向：价值/红利 vs 成长
    summary: str


@router.get("/overview", response_model=ApiResponse[CycleCompassState])
async def get_cycle_overview(
    response: Response,
    provider: BaseDataProvider = Depends(get_market_data_provider),
):
    """获取全球宏观周期罗盘综合概览（带 10s 客户端缓存）。"""
    response.headers["Cache-Control"] = "public, max-age=10"
    macros = await provider.get_macro_indicators()

    # 简要利差计算
    bond = next((m for m in macros if m.code == "CN10Y"), None)
    bond_yield = bond.value if bond else 1.5803
    div_yield = 4.39
    spread = div_yield - bond_yield

    state = CycleCompassState(
        current_phase="弱复苏与海外结构性繁荣分化",
        phase_code=1,
        dominant_asset="红利高股息与海外AI核心资产",
        macro_indicators=macros,
        equity_bond_spread=round(spread, 4),
        style_preference="红利防守为主，低区伺机成长",
        summary="国内十年期国债维持低位，红利股息率利差极具性价比；创业板50/红利低波100比值处于24%历史分位中间偏低区。",
    )

    return ApiResponse.success(state)
