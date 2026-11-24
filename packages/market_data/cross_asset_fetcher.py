# -*- coding: utf-8 -*-
"""全球宏观与多资产在线实时抓取器（移植自 daily_stock_analysis 的 yfinance / cross-asset 模块）。"""
from __future__ import annotations

import asyncio
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
import logging
import math
import time
from typing import Dict, List, Optional, Tuple

import yfinance as yf

from packages.market_data.models import (
    DataQualityStatus,
    DataSourceMeta,
    MarketQuote,
)

logger = logging.getLogger(__name__)

# 内部 symbol 到 Yahoo Finance 符号及元信息的映射表
CROSS_ASSET_MAPPING: Dict[str, Tuple[str, str, str, str, float]] = {
    # 格式: symbol -> (yf_ticker, name, category, currency, scale)
    "BRENT": ("BZ=F", "布伦特原油", "global", "USD", 1.0),
    "VIX": ("^VIX", "恐慌指数", "global", "", 1.0),
    "DXY": ("DX-Y.NYB", "美元指数", "fx", "USD", 1.0),
    "US20Y": ("TLT", "美债长债", "bond", "USD", 1.0),
    "GOLD_OZ": ("GC=F", "黄金盘司", "commodity", "USD", 1.0),
    "SILVER_OZ": ("SI=F", "白银盘司", "commodity", "USD", 1.0),
    "COPPER_HG": ("HG=F", "铜", "commodity", "USD", 100.0),  # 转换美分/磅至标准盘口指数
    "NATGAS": ("NG=F", "天然气", "commodity", "USD", 1.0),
    "ICNYXAU": ("GC=F", "人民币账户黄金", "commodity", "元/克", 0.1735),  # 国际金价换算人民币克重参考
    # 亚太指数与核心标的
    "KOSPI": ("^KS11", "KOSPI", "apac", "KRW", 1.0),
    "KOSDAQ": ("^KQ11", "KOSDAQ", "apac", "KRW", 1.0),
    "N225": ("^N225", "日经225", "apac", "JPY", 1.0),
    "TOPIX": ("1306.T", "TOPIX", "apac", "JPY", 1.0),
    "TWII": ("^TWII", "台湾加权", "apac", "TWD", 1.0),
    "TSMC": ("2330.TW", "台积电", "apac", "TWD", 1.0),
    # 港股与离岸中国
    "HSI": ("^HSI", "恒生指数", "hk", "HKD", 1.0),
    "HSTECH": ("HSTECH.HK", "恒生科技", "hk", "HKD", 1.0),
    "HSCE": ("^HSCE", "国企指数", "hk", "HKD", 1.0),
    "FXI": ("FXI", "富时中国50", "china", "USD", 1.0),
    "KWEB": ("KWEB", "中概互联", "china", "USD", 1.0),
    # 美股主指数与基准利率
    "SPX": ("^GSPC", "标普500", "us", "USD", 1.0),
    "NDX": ("^NDX", "纳斯达克100", "us", "USD", 1.0),
    "DJI": ("^DJI", "道琼斯", "us", "USD", 1.0),
    "US10Y": ("^TNX", "美债10年收益率", "bond", "%", 1.0),
    # 大宗商品与宏观流动性
    "GOLD": ("GC=F", "现货黄金", "commodity", "USD", 1.0),
    "COPPER": ("HG=F", "COMEX铜", "commodity", "USD", 1.0),
    "BTC": ("BTC-USD", "比特币", "crypto", "USD", 1.0),
    "ETH": ("ETH-USD", "以太坊", "crypto", "USD", 1.0),
    "SOYBEAN": ("ZS=F", "美大豆", "commodity", "USD", 1.0),
}

# 产业链与板块龙头 ETF / 代理标的映射（美股盘前与实盘跟踪）
SECTOR_YF_PROXY: Dict[str, str] = {
    "ai_compute": "NVDA",       # AI算力
    "cpo": "SMH",              # CPO
    "semiconductor": "SOXX",   # 半导体
    "storage": "MU",           # 存储
    "datacenter": "EQIX",      # 数据中心
    "cloud": "CLOU",           # 云计算
    "commercial_aerospace": "RKLB", # 商业航天
    "satellite": "ASTS",       # 卫星
    "robotics": "BOTZ",        # 机器人
    "autonomous_driving": "TSLA", # 自动驾驶
    "nuclear_power": "CCJ",    # 核电
    "power_grid": "GRID",      # 电网
    "military": "ITA",         # 军工
    "new_energy": "ICLN",      # 新能源
    "solar": "TAN",            # 光伏
    "lithium_battery": "LIT",  # 锂电池
    "petroleum": "XLE",        # 石油
    "natural_gas": "UNG",      # 天然气
    "copper_metal": "COPX",    # 铜 / 有色
    "gold": "GLD",             # 黄金
    "banking": "XLF",          # 银行金融
    "biomedicine": "XBI",      # 生物医药
    "consumption": "XLP",      # 消费
    "rare_earth": "REMX",      # 稀土
}


class CrossAssetFetcher:
    """跨资产多市场实时行情抓取适配器。"""

    def __init__(self, cache_ttl: float = 60.0):
        self.cache_ttl = cache_ttl
        self._cache: Dict[str, Tuple[float, MarketQuote]] = {}
        self._sector_cache: Dict[str, Tuple[float, float, float]] = {}  # sector_id -> (timestamp, price, pct)
        self._lock = asyncio.Lock()

    def _fetch_single_ticker(self, yf_sym: str) -> Tuple[str, Optional[float], Optional[float]]:
        """并行单元拉取单标的行情。"""
        try:
            t = yf.Ticker(yf_sym)
            info = t.fast_info
            last_price = getattr(info, "last_price", None)
            prev_close = getattr(info, "previous_close", None)
            if last_price is not None and not math.isnan(last_price):
                return yf_sym, float(last_price), float(prev_close) if prev_close else None
        except Exception as ex:
            logger.debug("Failed yfinance fetch for %s: %s", yf_sym, ex)
        return yf_sym, None, None

    def fetch_yf_batch_sync(self, symbols: List[str]) -> Dict[str, MarketQuote]:
        """并发执行 Yahoo Finance 批量查询，规避阻塞与顺序等待。"""
        results: Dict[str, MarketQuote] = {}
        if not symbols:
            return results

        # 构造 yf_ticker -> 内部 symbol 映射
        yf_to_internals: Dict[str, List[str]] = {}
        for sym in symbols:
            norm = sym.upper()
            if norm in CROSS_ASSET_MAPPING:
                yf_sym = CROSS_ASSET_MAPPING[norm][0]
                yf_to_internals.setdefault(yf_sym, []).append(norm)

        if not yf_to_internals:
            return results

        as_of = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        unique_yf_syms = list(yf_to_internals.keys())
        workers = min(10, max(1, len(unique_yf_syms)))

        try:
            with ThreadPoolExecutor(max_workers=workers) as executor:
                raw_results = list(executor.map(self._fetch_single_ticker, unique_yf_syms))

            for yf_sym, last_price, prev_close in raw_results:
                if last_price is None:
                    continue

                internal_list = yf_to_internals.get(yf_sym, [])
                for internal_sym in internal_list:
                    _, name, category, currency, scale = CROSS_ASSET_MAPPING[internal_sym]
                    scaled_price = round(last_price * scale, 2)

                    if prev_close and not math.isnan(prev_close) and prev_close > 0:
                        scaled_prev = prev_close * scale
                        change = round(scaled_price - scaled_prev, 2)
                        change_pct = round(((last_price - prev_close) / prev_close) * 100.0, 2)
                    else:
                        change = 0.0
                        change_pct = 0.0

                    meta = DataSourceMeta(
                        source_name="Yahoo Finance Realtime",
                        symbol=yf_sym,
                        as_of=as_of,
                        status=DataQualityStatus.OK,
                        latency_ms=120.0,
                    )

                    results[internal_sym] = MarketQuote(
                        symbol=internal_sym,
                        name=name,
                        category=category,
                        price=scaled_price,
                        change=change,
                        change_pct=change_pct,
                        currency=currency,
                        as_of=as_of,
                        meta=meta,
                    )
        except Exception as e:
            logger.warning("Error fetching batch from yfinance: %s", e)

        return results

    def fetch_sector_proxies_sync(self) -> Dict[str, Tuple[float, float]]:
        """并行拉取 24 大产业链代理 ETF/标的最新价格与涨跌幅。"""
        results: Dict[str, Tuple[float, float]] = {}
        unique_proxies = {sec: proxy for sec, proxy in SECTOR_YF_PROXY.items()}
        proxy_to_secs: Dict[str, List[str]] = {}
        for sec, proxy in unique_proxies.items():
            proxy_to_secs.setdefault(proxy, []).append(sec)

        workers = min(10, len(proxy_to_secs))
        try:
            with ThreadPoolExecutor(max_workers=workers) as executor:
                raw_results = list(executor.map(self._fetch_single_ticker, list(proxy_to_secs.keys())))

            for proxy, last_price, prev_close in raw_results:
                if last_price is None or prev_close is None or prev_close <= 0:
                    continue
                pct = round(((last_price - prev_close) / prev_close) * 100.0, 2)
                for sec in proxy_to_secs.get(proxy, []):
                    results[sec] = (last_price, pct)
        except Exception as ex:
            logger.warning("Failed to fetch sector proxies: %s", ex)

        return results

    async def get_quotes_batch(self, symbols: List[str]) -> Dict[str, MarketQuote]:
        """异步带缓存的跨资产批量行情拉取。"""
        now = time.monotonic()
        results: Dict[str, MarketQuote] = {}
        missing: List[str] = []

        for s in symbols:
            norm = s.upper()
            cached = self._cache.get(norm)
            if cached and (now - cached[0] < self.cache_ttl):
                results[norm] = cached[1]
            elif norm in CROSS_ASSET_MAPPING:
                missing.append(norm)

        if not missing:
            return results

        async with self._lock:
            # 双重检查
            now = time.monotonic()
            still_missing = [
                s for s in missing
                if s not in self._cache or (now - self._cache[s][0] >= self.cache_ttl)
            ]

            if still_missing:
                loop = asyncio.get_running_loop()
                fetched = await loop.run_in_executor(None, self.fetch_yf_batch_sync, still_missing)
                for sym, quote in fetched.items():
                    self._cache[sym] = (now, quote)
                    results[sym] = quote

        return results
