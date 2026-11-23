# -*- coding: utf-8 -*-
"""FastAPI 主应用入口、路由装配与高性能中间件（ORJSON/Gzip/Lifespan预热）。"""
from __future__ import annotations

import asyncio
from contextlib import asynccontextmanager
import logging
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.gzip import GZipMiddleware
from fastapi.responses import JSONResponse

from services.api.config import settings
from services.api.deps import get_market_data_provider, get_seed_provider
from services.api.routes import backtests, cycles, health, markets, portfolios, sectors, strategies
from services.api.routes.markets import DEFAULT_WATCHLIST
from services.api.routes.strategies import _get_cached_hist_ratios
from services.api.schemas.response import ApiResponse

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理与异步预热。

    在服务启动时在后台完成核心行情与策略序列预热，消除首个用户的冷启动抖动；
    停机时优雅释放底层 HTTP 连接池。
    """
    async def _warmup():
        try:
            seed = get_seed_provider()
            market = get_market_data_provider()
            # 1. 预热历史比值序列
            await _get_cached_hist_ratios(seed)
            # 2. 预热核心自选池行情缓存
            await market.get_quotes_batch(DEFAULT_WATCHLIST[:5])
        except Exception:
            logger.exception("后台预热任务执行失败")

    warmup_task = asyncio.create_task(_warmup())
    try:
        yield
    finally:
        if not warmup_task.done():
            warmup_task.cancel()
            try:
                await warmup_task
            except asyncio.CancelledError:
                pass
        await get_market_data_provider().close()


app = FastAPI(
    title=settings.app_name,
    version=settings.app_version,
    description="周期罗盘 CycleCompass 后端系统：全球行情监测、宏观产业周期、指数量化策略与模拟投资平台",
    docs_url="/docs",
    redoc_url="/redoc",
    lifespan=lifespan,
)

# 1. 跨域配置：当允许所有来源时关闭 credentials，符合浏览器 CORS 安全规范
allow_credentials = "*" not in settings.cors_origins
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_origins,
    allow_credentials=allow_credentials,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 2. Gzip 压缩中间件：对大于 1KB 的响应自动开启 Gzip 压缩，降低 70%+ 网络传输量
app.add_middleware(GZipMiddleware, minimum_size=1000)


@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    """全局统一异常处理（生产环境隐藏底层异常详情）。"""
    logger.exception("未处理的服务端异常: %s %s", request.method, request.url.path)
    message = f"服务器内部异常: {str(exc)}" if settings.debug else "服务器内部异常，请稍后重试"
    return JSONResponse(
        status_code=500,
        content=ApiResponse.fail(message=message, code=500).model_dump(),
    )


# 挂载业务模块路由
app.include_router(health.router, prefix="/api/v1")
app.include_router(markets.router, prefix="/api/v1")
app.include_router(sectors.router, prefix="/api/v1")
app.include_router(cycles.router, prefix="/api/v1")
app.include_router(strategies.router, prefix="/api/v1")
app.include_router(backtests.router, prefix="/api/v1")
app.include_router(portfolios.router, prefix="/api/v1")


@app.get("/")
async def root():
    return {
        "name": settings.app_name,
        "version": settings.app_version,
        "docs": "/docs",
        "health": "/api/v1/health",
    }
