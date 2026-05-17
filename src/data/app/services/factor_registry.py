# app/services/factor_registry.py — 因子注册表
# 定义全部 8 种因子：pe/pb/market_cap/sector/concept/ma/macd/rsi

from typing import Any, Callable
from dataclasses import dataclass, field


@dataclass
class FactorDef:
    """因子定义"""
    name: str
    category: str           # financial / sector / technical
    display: str            # 显示名
    params: dict = field(default_factory=dict)  # 参数定义
    data_source: str = ""   # akshare / indicator / computed


# ── 因子注册表 ──

FACTOR_REGISTRY: dict[str, FactorDef] = {
    # 财务指标
    "pe": FactorDef(
        name="pe", category="financial", display="市盈率",
        params={"min": {"type": "float", "default": 0},
                "max": {"type": "float", "default": 9999}},
        data_source="akshare",
    ),
    "pb": FactorDef(
        name="pb", category="financial", display="市净率",
        params={"min": {"type": "float", "default": 0},
                "max": {"type": "float", "default": 9999}},
        data_source="akshare",
    ),
    "market_cap": FactorDef(
        name="market_cap", category="financial", display="总市值",
        params={"min": {"type": "float", "default": 0},
                "max": {"type": "float", "default": 1e15}},
        data_source="akshare",
    ),
    # 板块
    "sector": FactorDef(
        name="sector", category="sector", display="行业板块",
        params={"value": {"type": "list", "default": []}},
        data_source="akshare",
    ),
    "concept": FactorDef(
        name="concept", category="sector", display="概念题材",
        params={"value": {"type": "list", "default": []}},
        data_source="akshare",
    ),
    # 技术指标
    "ma": FactorDef(
        name="ma", category="technical", display="移动均线",
        params={"period": {"type": "int", "default": 20},
                "cross_type": {"type": "str", "default": "golden"}},
        data_source="computed",
    ),
    "macd": FactorDef(
        name="macd", category="technical", display="MACD",
        params={"fast": {"type": "int", "default": 12},
                "slow": {"type": "int", "default": 26},
                "signal": {"type": "int", "default": 9}},
        data_source="computed",
    ),
    "rsi": FactorDef(
        name="rsi", category="technical", display="相对强弱",
        params={"period": {"type": "int", "default": 14},
                "threshold": {"type": "float", "default": 30}},
        data_source="computed",
    ),
}


def get_factor(name: str) -> FactorDef | None:
    return FACTOR_REGISTRY.get(name)


def list_factors(category: str = "") -> list[FactorDef]:
    if category:
        return [f for f in FACTOR_REGISTRY.values() if f.category == category]
    return list(FACTOR_REGISTRY.values())