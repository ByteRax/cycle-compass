# -*- coding: utf-8 -*-
"""模拟投资账户、虚拟持仓与账本资金流水接口。"""
from __future__ import annotations

from typing import Optional
from fastapi import APIRouter, Depends
from pydantic import BaseModel, Field

from packages.market_data.base import BaseDataProvider
from packages.strategy_engine import EventType, LedgerEntry
from services.api.deps import get_market_data_provider
from services.api.schemas.response import ApiResponse

router = APIRouter(prefix="/portfolios", tags=["Portfolios"])


class PositionItem(BaseModel):
    symbol: str
    name: str
    shares: float
    current_price: float
    market_value: float
    cost_price: float
    total_cost: float
    unrealized_profit: float
    unrealized_profit_pct: float


class PortfolioOverview(BaseModel):
    account_id: str
    total_asset: float
    available_cash: float
    cash_pool: float
    positions_value: float
    total_cost: float
    total_profit: float
    total_profit_pct: float
    positions: list[PositionItem]
    recent_transactions: list[LedgerEntry]


# 内存中模拟账户单例数据
_MOCK_ACCOUNT = {
    "account_id": "ACC_DEMO_001",
    "available_cash": 5200.0,
    "cash_pool": 1000.0,
    "positions": [
        {"symbol": "930955", "name": "红利低波100(515100)", "shares": 12500.0, "cost_price": 1.12},
        {"symbol": "399673", "name": "创业板50(159949)", "shares": 8000.0, "cost_price": 0.85},
    ],
    "transactions": [
        LedgerEntry(
            date="2026-11-16",
            event_type=EventType.CASH_DEPOSIT,
            target_name="计划现金池",
            amount=1000.0,
            cash_inflow=1000.0,
            cash_outflow=1000.0,
            cash_pool_delta=1000.0,
            cash_pool_after=1000.0,
            description="中间区：本周定投资金 1000 元存入计划现金池",
        ),
        LedgerEntry(
            date="2026-11-09",
            event_type=EventType.WEEKLY_INVEST,
            symbol="515100",
            target_name="红利低波100(515100)",
            shares=880.0,
            price=1.136,
            amount=1000.0,
            cash_inflow=1000.0,
            cash_outflow=1000.0,
            cash_pool_after=0.0,
            description="高区：周定投 1000 元买入红利低波100",
        ),
    ],
}


@router.get("/overview", response_model=ApiResponse[PortfolioOverview])
async def get_portfolio_overview(provider: BaseDataProvider = Depends(get_market_data_provider)):
    """获取模拟投资账户总览与持仓。"""
    pos_items: list[PositionItem] = []
    tot_pos_val = 0.0
    tot_cost = 0.0

    for p in _MOCK_ACCOUNT["positions"]:
        quote = await provider.get_quote(p["symbol"])
        curr_price = quote.price if quote.price < 50 else (quote.price / 10000.0) # ETF净值折算
        m_val = p["shares"] * curr_price
        c_val = p["shares"] * p["cost_price"]
        pnl = m_val - c_val
        pnl_pct = (pnl / c_val * 100.0) if c_val > 0 else 0.0

        pos_items.append(
            PositionItem(
                symbol=p["symbol"],
                name=p["name"],
                shares=p["shares"],
                current_price=round(curr_price, 4),
                market_value=round(m_val, 2),
                cost_price=p["cost_price"],
                total_cost=round(c_val, 2),
                unrealized_profit=round(pnl, 2),
                unrealized_profit_pct=round(pnl_pct, 2),
            )
        )
        tot_pos_val += m_val
        tot_cost += c_val

    avail_cash = float(_MOCK_ACCOUNT["available_cash"])
    c_pool = float(_MOCK_ACCOUNT["cash_pool"])
    tot_asset = tot_pos_val + avail_cash + c_pool
    tot_profit = tot_asset - (tot_cost + avail_cash + c_pool)
    tot_profit_pct = (tot_profit / tot_cost * 100.0) if tot_cost > 0 else 0.0

    overview = PortfolioOverview(
        account_id=_MOCK_ACCOUNT["account_id"],
        total_asset=round(tot_asset, 2),
        available_cash=round(avail_cash, 2),
        cash_pool=round(c_pool, 2),
        positions_value=round(tot_pos_val, 2),
        total_cost=round(tot_cost, 2),
        total_profit=round(tot_profit, 2),
        total_profit_pct=round(tot_profit_pct, 2),
        positions=pos_items,
        recent_transactions=_MOCK_ACCOUNT["transactions"],
    )

    return ApiResponse.success(overview)
