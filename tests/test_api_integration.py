# -*- coding: utf-8 -*-
"""FastAPI 接口全套集成测试。"""
import pytest
from httpx import ASGITransport, AsyncClient

from services.api.main import app


@pytest.fixture
async def client():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as c:
        yield c


@pytest.mark.asyncio
async def test_root_and_health(client: AsyncClient):
    resp = await client.get("/")
    assert resp.status_code == 200
    assert "CycleCompass" in resp.json()["name"]

    health_resp = await client.get("/api/v1/health")
    assert health_resp.status_code == 200
    h_data = health_resp.json()
    assert h_data["code"] == 0
    assert h_data["data"]["status"] in ("healthy", "degraded")


@pytest.mark.asyncio
async def test_markets_quotes(client: AsyncClient):
    resp = await client.get("/api/v1/markets/quotes")
    assert resp.status_code == 200
    data = resp.json()
    assert data["code"] == 0
    assert len(data["data"]) >= 5

    # 测试过滤分类
    equity_resp = await client.get("/api/v1/markets/quotes?category=equity")
    assert equity_resp.status_code == 200
    for item in equity_resp.json()["data"]:
        assert item["category"] == "equity"


@pytest.mark.asyncio
async def test_sectors_metrics(client: AsyncClient):
    resp = await client.get("/api/v1/sectors/metrics")
    assert resp.status_code == 200
    data = resp.json()
    assert data["code"] == 0
    assert any("AI" in s["name"] for s in data["data"])
    assert any("半导体" in s["name"] for s in data["data"])


@pytest.mark.asyncio
async def test_cycles_overview(client: AsyncClient):
    resp = await client.get("/api/v1/cycles/overview")
    assert resp.status_code == 200
    data = resp.json()
    assert data["code"] == 0
    assert "current_phase" in data["data"]
    assert "equity_bond_spread" in data["data"]


@pytest.mark.asyncio
async def test_strategies_signals(client: AsyncClient):
    # 轮动信号
    rot_resp = await client.get("/api/v1/strategies/rotation/signal")
    assert rot_resp.status_code == 200
    assert rot_resp.json()["data"]["zone"] in ("低区", "中间区", "高区")

    # 打分详情
    score_resp = await client.get("/api/v1/strategies/scoring/detail")
    assert score_resp.status_code == 200
    score_data = score_resp.json()["data"]
    assert 0 <= score_data["total"] <= 20

    # 联动决策
    comb_resp = await client.get("/api/v1/strategies/combined")
    assert comb_resp.status_code == 200
    assert "combined_action" in comb_resp.json()["data"]


@pytest.mark.asyncio
async def test_backtests_run(client: AsyncClient):
    payload = {
        "low_line": 0.16,
        "high_line": 0.28,
        "weekly_amount": 1000.0,
        "days": 300,
    }
    resp = await client.post("/api/v1/backtests/run", json=payload)
    assert resp.status_code == 200
    res_data = resp.json()
    assert res_data["code"] == 0
    summary = res_data["data"]["summary"]
    assert summary["trading_days"] == 300
    assert summary["total_invest"] > 0
    assert "xirr" in summary
    assert "max_drawdown" in summary

    # 校验非法轨线（low_line >= high_line）返回 422
    bad_resp = await client.post(
        "/api/v1/backtests/run",
        json={"low_line": 0.30, "high_line": 0.20, "weekly_amount": 1000.0, "days": 300},
    )
    assert bad_resp.status_code == 422


@pytest.mark.asyncio
async def test_portfolios_overview(client: AsyncClient):
    resp = await client.get("/api/v1/portfolios/overview")
    assert resp.status_code == 200
    data = resp.json()
    assert data["code"] == 0
    assert data["data"]["account_id"] == "ACC_DEMO_001"
    assert len(data["data"]["positions"]) >= 2
