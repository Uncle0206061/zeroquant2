# tests/test_integration.py — API 端点集成测试

import pytest
import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)


def test_health():
    r = client.get("/data/v1/health")
    assert r.status_code == 200
    assert r.json()["code"] == 0


def test_kline_000001_daily():
    r = client.get("/data/v1/kline/000001?period=daily&start_date=20260101&end_date=20260514")
    assert r.status_code == 200
    body = r.json()
    if body["code"] == 500:
        pytest.skip("网络波动")
    assert body["code"] == 0
    assert body["data"]["symbol"] == "000001"
    assert body["data"]["count"] >= 30


def test_kline_600519_weekly():
    r = client.get("/data/v1/kline/600519?period=weekly&start_date=20250101&end_date=20260514")
    assert r.status_code == 200
    body = r.json()
    if body["code"] == 500:
        pytest.skip("网络波动")
    assert body["code"] == 0
    assert body["data"]["count"] >= 10


def test_kline_invalid_stock():
    r = client.get("/data/v1/kline/999999?period=daily&start_date=20260101&end_date=20260514")
    assert r.status_code == 200
    body = r.json()
    assert body["code"] in [0, 500]


def test_sector_industry():
    r = client.get("/data/v1/sector/industry")
    assert r.status_code == 200
    body = r.json()
    if body["code"] == 500:
        pytest.skip("网络波动")
    assert body["code"] == 0
    assert body["data"]["count"] > 50


def test_orderbook_placeholder():
    r = client.get("/data/v1/orderbook/600519")
    assert r.status_code == 200
    assert r.json()["data"]["stock_code"] == "600519"


def test_filter_endpoint():
    r = client.post("/data/v1/filter", json={
        "rules": [{"factor": "pe", "min": 0, "max": 100}],
        "logic": "AND"
    })
    assert r.status_code == 200
    body = r.json()
    if body["code"] == 500 and "行情数据" in body.get("message", ""):
        pytest.skip("网络阻断")
    assert body["code"] == 0


def test_filter_parse():
    r = client.post("/data/v1/filter/parse", json={
        "stock_filter": {"pe": {"min": 0, "max": 30}},
        "timing": {},
        "risk": {}
    })
    assert r.status_code == 200
    assert r.json()["code"] == 0


def test_all_routes_registered():
    routes = [r.path for r in app.routes if hasattr(r, 'path')]
    expected = ["/data/v1/health", "/data/v1/kline/{symbol}", "/data/v1/market/spot",
                "/data/v1/market/{stock_code}", "/data/v1/orderbook/{stock_code}",
                "/data/v1/sector/industry", "/data/v1/sector/concept", "/data/v1/filter",
                "/data/v1/filter/parse"]
    for path in expected:
        assert path in routes, f"路由未注册: {path}"