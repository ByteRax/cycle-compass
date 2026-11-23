# -*- coding: utf-8 -*-
"""真实数据接入：中证指数、腾讯/新浪指数行情、东方财富国债收益率。

在线获取成功后写入本地缓存；网络失败时可显式降级到缓存，并在结果中留下警告。
不将不同基金、不同估值口径静默拼接成同一指标。
"""
from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date, datetime, timedelta
from pathlib import Path
import os
import tempfile

import pandas as pd


class DataSourceError(RuntimeError):
    pass


@dataclass
class MarketData:
    dividend: pd.DataFrame
    growth: pd.DataFrame
    valuation: pd.DataFrame
    bond: pd.DataFrame
    sources: dict[str, str] = field(default_factory=dict)
    warnings: list[str] = field(default_factory=list)


def _atomic_write_csv(df: pd.DataFrame, path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, tmp = tempfile.mkstemp(prefix=path.name, suffix=".tmp", dir=path.parent)
    os.close(fd)
    try:
        df.to_csv(tmp, index=False, encoding="utf-8-sig")
        os.replace(tmp, path)
    finally:
        if os.path.exists(tmp):
            os.unlink(tmp)


def _read_cache(path: Path) -> pd.DataFrame:
    if not path.exists() or path.stat().st_size == 0:
        raise DataSourceError(f"缓存不存在：{path}")
    return pd.read_csv(path)


def _normalise_date(df: pd.DataFrame, column: str = "date") -> pd.DataFrame:
    out = df.copy()
    out[column] = pd.to_datetime(out[column], errors="coerce").dt.normalize()
    out = out.dropna(subset=[column]).sort_values(column)
    return out.drop_duplicates(subset=[column], keep="last").reset_index(drop=True)


def _normalise_dividend(raw: pd.DataFrame) -> pd.DataFrame:
    out = pd.DataFrame({
        "date": raw["日期"],
        "close": pd.to_numeric(raw["收盘"], errors="coerce"),
        "pe_ttm": pd.to_numeric(raw["滚动市盈率"], errors="coerce"),
    })
    out = _normalise_date(out).dropna(subset=["close"])
    return out[(out["close"] > 0)].reset_index(drop=True)


def _normalise_growth(raw: pd.DataFrame) -> pd.DataFrame:
    date_col = "date" if "date" in raw.columns else "日期"
    close_col = "close" if "close" in raw.columns else "收盘"
    out = pd.DataFrame({
        "date": raw[date_col],
        "close": pd.to_numeric(raw[close_col], errors="coerce"),
    })
    out = _normalise_date(out).dropna(subset=["close"])
    return out[(out["close"] > 0)].reset_index(drop=True)


def _normalise_valuation(raw: pd.DataFrame) -> pd.DataFrame:
    out = pd.DataFrame({
        "date": raw["日期"],
        "pe1": pd.to_numeric(raw["市盈率1"], errors="coerce"),
        "pe2": pd.to_numeric(raw["市盈率2"], errors="coerce"),
        "dividend_yield_1": pd.to_numeric(raw["股息率1"], errors="coerce"),
        "dividend_yield_2": pd.to_numeric(raw["股息率2"], errors="coerce"),
    })
    return _normalise_date(out)


def _normalise_bond(raw: pd.DataFrame) -> pd.DataFrame:
    out = pd.DataFrame({
        "date": raw["日期"],
        "bond10y": pd.to_numeric(raw["中国国债收益率10年"], errors="coerce"),
    })
    return _normalise_date(out).dropna(subset=["bond10y"])


def _load_series(
    name: str,
    cache_path: Path,
    fetcher,
    normaliser,
    source_label: str,
    offline: bool,
    warnings: list[str],
    sources: dict[str, str],
) -> pd.DataFrame:
    if not offline:
        try:
            raw = fetcher()
            if raw is None or raw.empty:
                raise DataSourceError("接口返回空数据")
            frame = normaliser(raw)
            if frame.empty:
                raise DataSourceError("标准化后无有效数据")
            _atomic_write_csv(frame, cache_path)
            sources[name] = source_label
            return frame
        except Exception as exc:
            warnings.append(f"{name}在线获取失败，改用缓存：{type(exc).__name__}: {str(exc)[:160]}")
    try:
        frame = _normalise_date(_read_cache(cache_path))
        sources[name] = source_label + "（缓存）"
        if offline:
            warnings.append(f"{name}使用离线缓存")
        return frame
    except Exception as cache_exc:
        raise DataSourceError(f"{name}在线与缓存均不可用：{cache_exc}") from cache_exc


def load_market_data(config: dict, project_dir: Path, offline: bool = False,
                     as_of: date | None = None) -> MarketData:
    try:
        import akshare as ak
    except ImportError as exc:
        raise DataSourceError("缺少 akshare，请先安装 requirements.txt") from exc

    data_dir = (project_dir / config.get("data_dir", "data")).resolve()
    data_dir.mkdir(parents=True, exist_ok=True)
    end = as_of or datetime.now().date()
    start = str(config.get("start_date", "20160101"))
    end_s = end.strftime("%Y%m%d")
    warnings: list[str] = []
    sources: dict[str, str] = {}

    dividend = _load_series(
        "红利低波100行情", data_dir / "dividend_930955.csv",
        lambda: ak.stock_zh_index_hist_csindex(symbol="930955", start_date=start, end_date=end_s),
        _normalise_dividend, "中证指数官网 930955", offline, warnings, sources,
    )

    def growth_fetch():
        try:
            raw = ak.stock_zh_index_daily_tx(symbol="sz399673")
            sources["创业板50主源"] = "腾讯证券"
            return raw
        except Exception as exc:
            warnings.append(f"创业板50腾讯源失败，切换新浪源：{type(exc).__name__}")
            sources["创业板50主源"] = "新浪财经"
            return ak.stock_zh_index_daily(symbol="sz399673")

    growth = _load_series(
        "创业板50行情", data_dir / "growth_399673.csv", growth_fetch,
        _normalise_growth, "腾讯/新浪 399673", offline, warnings, sources,
    )
    valuation = _load_series(
        "红利低波100估值", data_dir / "valuation_930955.csv",
        lambda: ak.stock_zh_index_value_csindex(symbol="930955"),
        _normalise_valuation, "中证指数官网估值（近20个交易日）", offline, warnings, sources,
    )
    bond_start = max(datetime.strptime(start, "%Y%m%d").date(), end - timedelta(days=45))
    bond = _load_series(
        "十年期国债", data_dir / "bond10y.csv",
        lambda: ak.bond_zh_us_rate(start_date=bond_start.strftime("%Y%m%d")),
        _normalise_bond, "东方财富/中国债券信息网", offline, warnings, sources,
    )

    cutoff = pd.Timestamp(end)
    return MarketData(
        dividend=dividend[dividend["date"] <= cutoff].reset_index(drop=True),
        growth=growth[growth["date"] <= cutoff].reset_index(drop=True),
        valuation=valuation[valuation["date"] <= cutoff].reset_index(drop=True),
        bond=bond[bond["date"] <= cutoff].reset_index(drop=True),
        sources=sources,
        warnings=warnings,
    )
