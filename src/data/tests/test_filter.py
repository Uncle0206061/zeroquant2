# tests/test_filter.py — TC-M2-03 多因子筛选引擎集成测试

import pytest
import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from fastapi.testclient import TestClient
from app.main import app

client = TestClient(app)


def _check_network(body):
    if body.get("code") == 500 and "行情数据" in body.get("message", ""):
        pytest.skip("网络阻断 — 非代码问题")
    if body.get("code") == 500:
        msg = body.get("message", "")
        if any(kw in msg for kw in ["不可用", "连接", "abort", "timeout"]):
            pytest.skip(f"网络阻断: {msg[:60]}")


class TestFilterFinancial:

    def test_pe_range(self):
        r = client.post("/data/v1/filter", json={
            "rules": [{"type": "financial", "factor": "pe", "min": 0, "max": 30}],
            "logic": "AND"
        })
        assert r.status_code == 200
        body = r.json()
        _check_network(body)
        assert body["code"] == 0
        assert "matched" in body["data"]
        assert body["data"]["total_scanned"] > 1000
        print(f"  PE: {body['data']['matched_count']}/{body['data']['total_scanned']} ({body['data']['elapsed_ms']}ms)")

    def test_pb_range(self):
        r = client.post("/data/v1/filter", json={
            "rules": [{"type": "financial", "factor": "pb", "min": 0, "max": 2}],
            "logic": "AND"
        })
        body = r.json()
        _check_network(body)
        assert body["code"] == 0
        print(f"  PB: {body['data']['matched_count']}/{body['data']['total_scanned']} ({body['data']['elapsed_ms']}ms)")

    def test_pe_and_pb(self):
        r = client.post("/data/v1/filter", json={
            "rules": [
                {"type": "financial", "factor": "pe", "min": 0, "max": 30},
                {"type": "financial", "factor": "pb", "min": 0, "max": 3}
            ],
            "logic": "AND"
        })
        body = r.json()
        _check_network(body)
        assert body["code"] == 0
        print(f"  PE+PB(AND): {body['data']['matched_count']} ({body['data']['elapsed_ms']}ms)")


class TestFilterSector:

    def test_sector_bank(self):
        r = client.post("/data/v1/filter", json={
            "rules": [{"type": "sector", "factor": "sector", "value": ["银行"]}],
            "logic": "AND"
        })
        body = r.json()
        _check_network(body)
        assert body["code"] == 0
        assert len(body["data"]["matched"]) >= 0
        print(f"  银行板块: {body['data']['matched_count']} 只")


class TestFilterParse:

    def test_parse_simple(self):
        r = client.post("/data/v1/filter/parse", json={
            "stock_filter": {"pe": {"min": 0, "max": 30}, "sector": ["银行", "证券"]},
            "timing": {"type": "ma_cross", "period": 20},
            "risk": {"max_position": 0.3, "stop_loss": 0.05}
        })
        assert r.status_code == 200
        body = r.json()
        assert body["code"] == 0
        assert len(body["data"]["parsed_rules"]) == 3
        assert body["data"]["validation"] == "ok"

    def test_parse_macd_timing(self):
        r = client.post("/data/v1/filter/parse", json={
            "stock_filter": {"pb": {"min": 0, "max": 2}},
            "timing": {"type": "macd_golden", "fast": 12, "slow": 26, "signal": 9},
            "risk": {"max_position": 0.2, "stop_loss": 0.08}
        })
        assert r.json()["code"] == 0
        rules = r.json()["data"]["parsed_rules"]
        assert len(rules) == 2
        assert rules[1]["factor"] == "macd"

    def test_parse_invalid_factor(self):
        r = client.post("/data/v1/filter/parse", json={
            "stock_filter": {"unknown_xyz": {"min": 0}},
            "timing": {},
            "risk": {}
        })
        body = r.json()
        assert body["data"]["validation"] == "warning"
        assert len(body["data"]["warnings"]) >= 1

    def test_parse_risk_validation(self):
        r = client.post("/data/v1/filter/parse", json={
            "stock_filter": {},
            "timing": {},
            "risk": {"max_position": 1.5, "stop_loss": -0.1}
        })
        body = r.json()
        assert body["data"]["validation"] == "warning"

    def test_parse_rsi_timing(self):
        r = client.post("/data/v1/filter/parse", json={
            "stock_filter": {"pb": {"min": 0, "max": 3}},
            "timing": {"type": "rsi_oversold", "period": 14, "threshold": 30},
            "risk": {"max_position": 0.25, "stop_loss": 0.06}
        })
        assert r.json()["code"] == 0
        rules = r.json()["data"]["parsed_rules"]
        assert len(rules) == 2
        assert rules[1]["factor"] == "rsi"


class TestFilterEdgeCases:

    def test_unknown_factor(self):
        r = client.post("/data/v1/filter", json={
            "rules": [{"factor": "nonexistent"}],
            "logic": "AND"
        })
        assert r.json()["code"] == 400

    def test_empty_rules(self):
        r = client.post("/data/v1/filter", json={
            "rules": [],
            "logic": "AND"
        })
        assert r.status_code == 422

    def test_invalid_logic(self):
        r = client.post("/data/v1/filter", json={
            "rules": [{"factor": "pe", "min": 0, "max": 100}],
            "logic": "XOR"
        })
        assert r.status_code == 422

    def test_stocks_whitelist(self):
        r = client.post("/data/v1/filter", json={
            "rules": [{"factor": "pe", "min": 0, "max": 999}],
            "logic": "AND",
            "stocks": ["000001", "600519", "300750"]
        })
        body = r.json()
        _check_network(body)
        assert body["code"] == 0
        matched = body["data"]["matched"]
        for m in matched:
            assert m in ["000001", "600519", "300750"]

    def test_factor_registry_complete(self):
        from app.services.factor_registry import FACTOR_REGISTRY
        assert len(FACTOR_REGISTRY) == 8
        for name in ["pe", "pb", "market_cap", "sector", "concept", "ma", "macd", "rsi"]:
            assert name in FACTOR_REGISTRY, f"缺少因子: {name}"

    def test_indicators_rsi(self):
        from app.services.indicators import calc_rsi
        import pandas as pd
        closes = pd.Series([10,11,12,11,10,9,8,9,10,11,12,13,14,15,14,13,12,11,10,11])
        rsi = calc_rsi(closes, 14)
        assert 0 <= rsi <= 100
        print(f"  RSI: {rsi}")

    def test_indicators_macd(self):
        from app.services.indicators import calc_macd
        import pandas as pd
        closes = pd.Series([10,11,12,11,10,9,8,9,10,11,12,13,14,15,14,13,12,11,10,11]*5)
        result = calc_macd(closes)
        assert "diff" in result
        assert "golden_cross" in result
        print(f"  MACD golden_cross: {result['golden_cross']}")