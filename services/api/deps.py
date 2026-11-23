# -*- coding: utf-8 -*-
"""FastAPI 依赖注入与服务单例。"""
from __future__ import annotations

from packages.market_data.online_provider import OnlineDataProvider
from packages.market_data.seed_provider import SeedDataProvider

# 单例提供者实例
_seed_provider = SeedDataProvider()
_online_provider = OnlineDataProvider(fallback_provider=_seed_provider)


def get_market_data_provider() -> OnlineDataProvider:
    """获取带三级降级支持的市场数据服务。"""
    return _online_provider


def get_seed_provider() -> SeedDataProvider:
    """获取纯离线种子数据提供者。"""
    return _seed_provider
