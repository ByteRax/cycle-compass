# -*- coding: utf-8 -*-
"""量化回测在线计算接口。"""
from __future__ import annotations

import asyncio
from datetime import date
from typing import Optional
from fastapi import APIRouter, Depends
from pydantic import BaseModel, Field, model_validator

from packages.backtesting import BacktestOutput, RotationBacktestEngine
from packages.market_data.base import BaseDataProvider
from packages.strategy_engine import RotationConfig
from services.api.deps import get_seed_provider
from services.api.schemas.response import ApiResponse

router = APIRouter(prefix="/backtests", tags=["Backtests"])

# 内存 LRU 缓存，最多保存 64 组高频回测结果，避免并发重复计算打满 CPU
_BACKTEST_CACHE: dict[tuple, BacktestOutput] = {}
_IN_FLIGHT_TASKS: dict[tuple, asyncio.Future[BacktestOutput]] = {}
_MAX_CACHE_SIZE = 64
_CACHE_LOCK = asyncio.Lock()


class BacktestRequest(BaseModel):
    low_line: float = Field(0.16, ge=0.05, le=0.50, description="轮动低轨")
    high_line: float = Field(0.28, ge=0.10, le=0.80, description="轮动高轨")
    weekly_amount: float = Field(1000.0, ge=100.0, description="每周定投金额")
    days: int = Field(1500, ge=100, le=3000, description="历史回测天数")

    @model_validator(mode="after")
    def validate_thresholds(self) -> "BacktestRequest":
        if self.low_line >= self.high_line:
            raise ValueError("轮动低轨 (low_line) 必须严格小于轮动高轨 (high_line)")
        return self


@router.post("/run", response_model=ApiResponse[BacktestOutput])
async def run_backtest(
    req: BacktestRequest,
    seed_provider: BaseDataProvider = Depends(get_seed_provider),
):
    """运行全历史/指定周期比值轮动回测（带并发线程隔离、Singleflight防击穿与内存 LRU 缓存）。"""
    cache_key = (
        round(req.low_line, 4),
        round(req.high_line, 4),
        round(req.weekly_amount, 2),
        req.days,
    )
    if cache_key in _BACKTEST_CACHE:
        return ApiResponse.success(_BACKTEST_CACHE[cache_key])

    # Singleflight 并发防击穿：仅由创建 Future 的主协程执行计算，其余并发请求等待同一 Future
    is_owner = False
    async with _CACHE_LOCK:
        if cache_key in _BACKTEST_CACHE:
            return ApiResponse.success(_BACKTEST_CACHE[cache_key])
        if cache_key in _IN_FLIGHT_TASKS:
            fut = _IN_FLIGHT_TASKS[cache_key]
        else:
            loop = asyncio.get_running_loop()
            fut = loop.create_future()
            _IN_FLIGHT_TASKS[cache_key] = fut
            is_owner = True

    if not is_owner:
        return await fut

    try:
        g_df = await seed_provider.get_historical_klines("399673", days=req.days + 100)
        d_df = await seed_provider.get_historical_klines("930955", days=req.days + 100)

        if g_df.empty or d_df.empty:
            err_msg = "历史行情数据不足，无法执行回测"
            fail_resp = ApiResponse.fail(err_msg)
            if not fut.done():
                fut.set_result(fail_resp)
            return fail_resp

        merged = g_df.merge(d_df, on="date", suffixes=("_g", "_d")).dropna()
        merged = merged.tail(req.days)

        dates = [date.fromisoformat(d) for d in merged["date"]]
        g_closes = merged["close_g"].astype(float).tolist()
        d_closes = merged["close_d"].astype(float).tolist()

        cfg = RotationConfig(
            low_line=req.low_line,
            high_line=req.high_line,
            weekly_amount=req.weekly_amount,
        )

        engine = RotationBacktestEngine(cfg)
        result = await asyncio.to_thread(engine.run, dates, g_closes, d_closes)

        async with _CACHE_LOCK:
            if len(_BACKTEST_CACHE) >= _MAX_CACHE_SIZE:
                _BACKTEST_CACHE.pop(next(iter(_BACKTEST_CACHE)))
            _BACKTEST_CACHE[cache_key] = result
        success_resp = ApiResponse.success(result)
        if not fut.done():
            fut.set_result(success_resp)
        return success_resp
    except BaseException as e:
        if not fut.done():
            fut.set_exception(e)
        raise
    finally:
        if not fut.done():
            fut.cancel()
        async with _CACHE_LOCK:
            if _IN_FLIGHT_TASKS.get(cache_key) is fut:
                _IN_FLIGHT_TASKS.pop(cache_key, None)
