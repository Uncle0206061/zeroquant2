# tests/test_level2_simulated.py — 模拟 Level-2 端到端测试

import pytest
import sys, os, asyncio, time
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)


class TestSimulatedLevel2:
    """模拟 Level-2 数据源端到端"""

    def test_simulator_import(self):
        from app.services.simulated_level2 import simulated_level2
        assert simulated_level2 is not None
        assert simulated_level2.interval == 3.0

    def test_mock_orderbook_generation(self):
        """mock 盘口数据生成"""
        from app.services.simulated_level2 import SimulatedLevel2
        sim = SimulatedLevel2()
        data = sim._generate_mock_orderbook("000001")
        assert data["symbol"] == "000001"
        assert data["type"] == "orderbook"
        assert "timestamp" in data
        assert "price" in data
        assert len(data["bids"]) == 5
        assert len(data["asks"]) == 5
        for bid in data["bids"]:
            assert len(bid) == 2 and bid[0] > 0 and bid[1] > 0
        for ask in data["asks"]:
            assert len(ask) == 2 and ask[0] > 0 and ask[1] > 0

    def test_orderbook_bid_ask_sorted(self):
        """买盘降序（越高越好），卖盘升序（越低越好）"""
        from app.services.simulated_level2 import SimulatedLevel2
        sim = SimulatedLevel2()
        data = sim._generate_mock_orderbook("600519")
        bids = [b[0] for b in data["bids"]]
        asks = [a[0] for a in data["asks"]]
        assert bids == sorted(bids, reverse=True), f"买盘未降序: {bids}"
        assert asks == sorted(asks), f"卖盘未升序: {asks}"

    def test_orderbook_best_bid_ask(self):
        """买一 < 卖一（盘口价差合理）"""
        from app.services.simulated_level2 import SimulatedLevel2
        sim = SimulatedLevel2()
        data = sim._generate_mock_orderbook("300750")
        best_bid = data["bids"][0][0]
        best_ask = data["asks"][0][0]
        assert best_bid < best_ask, f"买一({best_bid}) >= 卖一({best_ask})"


class TestOrderbookAPI:
    """盘口 API"""

    def test_orderbook_cache_miss(self):
        """缓存未命中 → 返回空盘口"""
        r = client.get("/data/v1/orderbook/999999")
        assert r.status_code == 200
        body = r.json()
        assert body["code"] == 0
        assert body["data"]["bids"] == []
        assert body["data"]["source"] == "cache-miss"

    def test_orderbook_with_cached_data(self):
        """Redis 有缓存 → 返回真实盘口"""
        from app.cache import CacheManager as cache
        mock_data = {
            "symbol": "000001",
            "type": "orderbook",
            "price": 10.5,
            "bids": [[10.49, 1000], [10.48, 2000], [10.47, 1500], [10.46, 3000], [10.45, 500]],
            "asks": [[10.51, 800], [10.52, 1200], [10.53, 900], [10.54, 2000], [10.55, 600]],
            "timestamp": "2026-05-14T12:00:00"
        }
        cache.set_orderbook("000001", mock_data)

        r = client.get("/data/v1/orderbook/000001")
        assert r.status_code == 200
        body = r.json()
        if cache.is_available():
            assert body["data"]["bids"] != []
            assert body["data"]["source"] == "level2-cache"
            assert len(body["data"]["bids"]) == 5

    def test_orderbook_subscribe(self):
        """订阅接口"""
        r = client.post("/data/v1/orderbook/subscribe", json=["000001", "600519"])
        assert r.status_code == 200
        assert r.json()["code"] == 0


class TestStartupLifecycle:
    """服务启动时自动拉起模拟器"""

    def test_simulator_running_after_startup(self):
        from app.services.simulated_level2 import simulated_level2
        # lifespan 在 TestClient 创建时已触发
        # 模拟器应已启动（或有 scheduled task）
        assert simulated_level2 is not None
        # 订阅000001是否在列表中
        assert "000001" in simulated_level2._callbacks or simulated_level2._running