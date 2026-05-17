# tests/test_tc_m2_01.py — TC-M2-01 补全测试
# 覆盖: 缓存读写 / 监控端点 / Level-2健康 / 数据库模型

import pytest
import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)


class TestCache:
    """Redis 缓存层测试"""

    def test_cache_import(self):
        from app.cache import CacheManager
        assert CacheManager is not None

    def test_cache_set_get(self):
        from app.cache import CacheManager as cache
        cache.set("test", "key1", {"val": 123}, ttl=10)
        result = cache.get("test", "key1")
        if cache.is_available():
            assert result is not None
            assert result["val"] == 123
        else:
            assert result is None  # Redis 不可用时返回 None

    def test_cache_spot_ops(self):
        from app.cache import CacheManager as cache
        cache.set_spot("000001", {"price": 10.5, "name": "平安银行"})
        got = cache.get_spot("000001")
        if cache.is_available():
            assert got is not None
            assert got["price"] == 10.5

    def test_cache_kline_ops(self):
        from app.cache import CacheManager as cache
        cache.set_kline("000001", "daily", "20260101", "20260514", {"count": 84})
        got = cache.get_kline("000001", "daily", "20260101", "20260514")
        if cache.is_available():
            assert got is not None
            assert got["count"] == 84

    def test_cache_graceful_degradation(self):
        """Redis 不可用时不应崩溃"""
        from app.cache import CacheManager as cache
        # 即使 Redis 不可用，get/set 都不应抛异常
        try:
            cache.set("test", "graceful", {"x": 1})
            cache.get("test", "graceful")
        except Exception as e:
            pytest.fail(f"缓存操作异常: {e}")


class TestMonitor:
    """监控端点"""

    def test_monitor_status_200(self):
        r = client.get("/data/v1/monitor/status")
        assert r.status_code == 200

    def test_monitor_returns_data(self):
        r = client.get("/data/v1/monitor/status")
        body = r.json()
        assert body["code"] == 0
        assert "data_sources" in body["data"]
        assert "redis_health" in body["data"]
        assert "postgres_health" in body["data"]
        assert "collector_running" in body["data"]
        assert "alert_level" in body["data"]

    def test_monitor_akshare_status(self):
        r = client.get("/data/v1/monitor/status")
        sources = r.json()["data"]["data_sources"]
        assert "akshare" in sources
        assert sources["akshare"] in ["ok", "degraded", "error"]


class TestLevel2:
    """Level-2 采集器"""

    def test_level2_import(self):
        from app.collectors.level2_collector import level2
        assert level2 is not None
        assert level2.connected is False  # 未配置时不连接

    def test_level2_health(self):
        from app.collectors.level2_collector import level2
        h = level2.health
        assert "connected" in h
        assert "subscribers" in h
        assert h["connected"] is False  # 本地无 Level-2 服务器


class TestModels:
    """数据库模型"""

    def test_models_import(self):
        from app.models.models import Kline, Quote, Tick, Sector
        assert Kline.__tablename__ == "data_kline"
        assert Quote.__tablename__ == "data_quote"
        assert Tick.__tablename__ == "data_tick"
        assert Sector.__tablename__ == "data_sector"

    def test_models_have_data_prefix(self):
        from app.models.models import Kline, Quote, Tick, Sector
        for model in [Kline, Quote, Tick, Sector]:
            assert model.__tablename__.startswith("data_"), f"{model.__name__} 缺少 data_ 前缀"


class TestRoutesComplete:
    """全路由验证"""

    def test_monitor_route_registered(self):
        routes = [r.path for r in app.routes if hasattr(r, 'path')]
        assert "/data/v1/monitor/status" in routes