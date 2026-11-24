# -*- coding: utf-8 -*-
"""实时公开行情在线适配器（支持腾讯/新浪/东方财富公开接口，带内存防抖缓存与批量聚合）。"""
from __future__ import annotations

import asyncio
from datetime import datetime
import time
from typing import Optional
import httpx
import pandas as pd

from packages.market_data.base import BaseDataProvider
from packages.market_data.cross_asset_fetcher import CROSS_ASSET_MAPPING, CrossAssetFetcher
from packages.market_data.models import (
    DataQualityStatus,
    DataSourceMeta,
    GlobalMacroIndicator,
    IndustryMetric,
    MarketQuote,
)
from packages.market_data.seed_provider import KNOWN_ASSETS, SeedDataProvider

TENCENT_API = "https://qt.gtimg.cn/q="


class OnlineDataProvider(BaseDataProvider):
    """通过公开 HTTP API 在线获取行情，带连接池复用、批量聚合与内存 TTL 缓存。"""

    def __init__(
        self,
        timeout: float = 3.0,
        cache_ttl: float = 3.0,
        fallback_provider: Optional[BaseDataProvider] = None,
    ):
        self.timeout = timeout
        self.cache_ttl = cache_ttl
        self.fallback = fallback_provider or SeedDataProvider()
        self.cross_asset = CrossAssetFetcher(cache_ttl=max(cache_ttl, 10.0))
        self._client: Optional[httpx.AsyncClient] = None
        self._cache: dict[str, tuple[float, MarketQuote]] = {}
        self._lock = asyncio.Lock()

    async def _get_client(self) -> httpx.AsyncClient:
        """获取共享的长连接池客户端单例。"""
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(
                timeout=self.timeout,
                limits=httpx.Limits(max_keepalive_connections=20, max_connections=50),
            )
        return self._client

    async def close(self) -> None:
        """释放底层 HTTP 连接池资源。"""
        if self._client is not None and not self._client.is_closed:
            await self._client.aclose()
            self._client = None

    def _symbol_to_tencent_code(self, symbol: str) -> Optional[str]:
        # A股代码映射
        if symbol == "930955":
            return "csi930955"  # 中证红利低波100
        if symbol == "399673":
            return "sz399673"   # 创业板50
        if symbol == "000300":
            return "sh000300"   # 沪深300
        if symbol == "000905":
            return "sh000905"   # 中证500
        if symbol == "000688":
            return "sh000688"   # 科创50
        return None

    def _tencent_code_to_symbol(self, code: str) -> Optional[str]:
        mapping = {
            "csi930955": "930955",
            "sz399673": "399673",
            "sh000300": "000300",
            "sh000905": "000905",
            "sh000688": "000688",
        }
        return mapping.get(code.lower())

    def _parse_tencent_line(self, line: str, latency_ms: float) -> Optional[MarketQuote]:
        """解析腾讯行情单行数据。格式: v_sz399673="51~创业板50~399673~2192.80~...";"""
        line = line.strip()
        if not line or '="' not in line:
            return None
        prefix, data_part = line.split('="', 1)
        raw_code = prefix.split("v_")[-1]
        data_part = data_part.rstrip('";')
        fields = data_part.split("~")
        if len(fields) < 35:
            return None

        symbol = self._tencent_code_to_symbol(raw_code) or fields[2]
        name = fields[1]
        try:
            price = float(fields[3])
            yesterday_close = float(fields[4])
        except (ValueError, TypeError):
            return None

        change = price - yesterday_close
        if len(fields) > 32 and fields[32]:
            try:
                change_pct = float(fields[32])
            except ValueError:
                change_pct = (change / yesterday_close) * 100 if yesterday_close else 0.0
        else:
            change_pct = (change / yesterday_close) * 100 if yesterday_close else 0.0

        as_of = datetime.now().strftime("%Y-%m-%d")
        meta = DataSourceMeta(
            source_name="Tencent Finance Realtime",
            symbol=symbol,
            as_of=as_of,
            status=DataQualityStatus.OK,
            latency_ms=round(latency_ms, 1),
        )
        return MarketQuote(
            symbol=symbol,
            name=name,
            category="equity",
            price=round(price, 4),
            change=round(change, 4),
            change_pct=round(change_pct, 2),
            currency="CNY",
            as_of=as_of,
            meta=meta,
        )

    async def get_quote(self, symbol: str) -> MarketQuote:
        """获取单个标的最新行情（优先命中内存 TTL 缓存）。"""
        norm_sym = symbol.upper()
        now = time.monotonic()

        # 1. 尝试读缓存
        cached = self._cache.get(norm_sym)
        if cached and (now - cached[0] < self.cache_ttl):
            return cached[1]

        # 2. 未命中，委托给批量方法防击穿
        quotes = await self.get_quotes_batch([norm_sym])
        return quotes[0] if quotes else await self.fallback.get_quote(norm_sym)

    async def get_quotes_batch(self, symbols: list[str]) -> list[MarketQuote]:
        """批量获取行情：多标的聚合单个 HTTP 请求 + 内存 TTL 缓存防击穿。"""
        norm_symbols = [s.strip().upper() for s in symbols if s.strip()]
        now = time.monotonic()

        results: dict[str, MarketQuote] = {}
        missing_symbols: list[str] = []

        # 1. 快速检查缓存
        for sym in norm_symbols:
            cached = self._cache.get(sym)
            if cached and (now - cached[0] < self.cache_ttl):
                results[sym] = cached[1]
            else:
                missing_symbols.append(sym)

        if not missing_symbols:
            return [results[s] for s in norm_symbols if s in results]

        # 2. 加互斥锁，防止并发击穿源站
        async with self._lock:
            # 双重检查
            now = time.monotonic()
            still_missing: list[str] = []
            for sym in missing_symbols:
                cached = self._cache.get(sym)
                if cached and (now - cached[0] < self.cache_ttl):
                    results[sym] = cached[1]
                else:
                    still_missing.append(sym)

            if not still_missing:
                return [results[s] for s in norm_symbols if s in results]

            # 拆分：跨资产走 Yahoo Finance，A股代码走腾讯 HTTP，其余走 fallback
            cross_asset_symbols: list[str] = []
            tencent_codes_map: dict[str, str] = {}
            fallback_needed: list[str] = []

            for sym in still_missing:
                if sym in CROSS_ASSET_MAPPING:
                    cross_asset_symbols.append(sym)
                else:
                    t_code = self._symbol_to_tencent_code(sym)
                    if t_code:
                        tencent_codes_map[t_code] = sym
                    else:
                        fallback_needed.append(sym)

            # 3.1 跨资产在线拉取（原油、黄金、美债、VIX、日韩指数等）
            if cross_asset_symbols:
                try:
                    ca_quotes = await self.cross_asset.get_quotes_batch(cross_asset_symbols)
                    for sym, quote in ca_quotes.items():
                        self._cache[sym] = (now, quote)
                        results[sym] = quote
                except Exception as ex:
                    pass

                for sym in cross_asset_symbols:
                    if sym not in results:
                        fallback_needed.append(sym)

            # 3.2 批量打向腾讯源站（A股指数等）
            if tencent_codes_map:
                start_t = time.perf_counter()
                client = await self._get_client()
                codes_query = ",".join(tencent_codes_map.keys())
                try:
                    resp = await client.get(f"{TENCENT_API}{codes_query}")
                    latency = (time.perf_counter() - start_t) * 1000.0
                    if resp.status_code == 200 and resp.text:
                        lines = resp.text.strip().split(";")
                        for line in lines:
                            parsed = self._parse_tencent_line(line, latency)
                            if parsed and parsed.symbol in still_missing:
                                self._cache[parsed.symbol] = (now, parsed)
                                results[parsed.symbol] = parsed
                except Exception:
                    pass

                # 对拉取失败或源站未返回的，降级回退
                for t_code, sym in tencent_codes_map.items():
                    if sym not in results:
                        fallback_needed.append(sym)

            # 4. 统一处理 fallback
            for sym in fallback_needed:
                fb_quote = await self.fallback.get_quote(sym)
                fb_quote.meta.warning = "网络请求超时或源站未收录，已自动平滑降级至最新离线种子快照"
                self._cache[sym] = (now, fb_quote)
                results[sym] = fb_quote

        return [results[s] for s in norm_symbols if s in results]

    async def get_macro_indicators(self) -> list[GlobalMacroIndicator]:
        return await self.fallback.get_macro_indicators()

    async def get_industry_metrics(self) -> list[IndustryMetric]:
        metrics = await self.fallback.get_industry_metrics()
        try:
            loop = asyncio.get_running_loop()
            proxies = await loop.run_in_executor(None, self.cross_asset.fetch_sector_proxies_sync)
            if proxies:
                updated = []
                now_str = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                for m in metrics:
                    if m.sector_id in proxies:
                        _, pct = proxies[m.sector_id]
                        new_m = m.model_copy(update={
                            "price_change_pct": pct,
                            "price_index": round(100.0 + pct, 2),
                            "meta": DataSourceMeta(
                                source_name="Global Proxy Realtime",
                                symbol=m.sector_id,
                                as_of=now_str,
                                status=DataQualityStatus.OK,
                                latency_ms=100.0,
                            )
                        })
                        updated.append(new_m)
                    else:
                        updated.append(m)
                return updated
        except Exception as e:
            logger.debug("Failed online industry metric calibration: %s", e)
        return metrics

    async def get_historical_klines(self, symbol: str, days: int = 1500) -> pd.DataFrame:
        return await self.fallback.get_historical_klines(symbol, days)
