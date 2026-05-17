# app/api/v1/filter_routes.py — 多因子筛选接口 TC-M2-03
# POST /data/v1/filter     筛选执行
# POST /data/v1/filter/parse  策略解析

import time
import logging
from typing import Optional
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

from app.services.filter_engine import engine
from app.services.factor_registry import FACTOR_REGISTRY, get_factor
from app.config import settings

logger = logging.getLogger(__name__)
router = APIRouter()


# ── 请求模型 ──

class FilterRule(BaseModel):
    type: str = ""                  # financial / sector / technical
    factor: str                     # pe/pb/market_cap/sector/concept/ma/macd/rsi
    min: Optional[float] = None
    max: Optional[float] = None
    value: Optional[list[str]] = None
    period: Optional[int] = None
    threshold: Optional[float] = None
    cross_type: Optional[str] = None
    fast: Optional[int] = None
    slow: Optional[int] = None
    signal: Optional[int] = None


class FilterRequest(BaseModel):
    rules: list[FilterRule] = Field(..., min_length=1, max_length=10)
    logic: str = Field("AND", pattern="^(AND|OR)$")
    stocks: Optional[list[str]] = None
    date: Optional[str] = None

    class Config:
        json_schema_extra = {
            "example": {
                "rules": [
                    {"type": "financial", "factor": "pe", "min": 0, "max": 30},
                    {"type": "sector", "factor": "sector", "value": ["银行"]}
                ],
                "logic": "AND",
                "stocks": None
            }
        }


class StrategyJSON(BaseModel):
    stock_filter: dict = Field(default_factory=dict)
    timing: dict = Field(default_factory=dict)
    risk: dict = Field(default_factory=dict)

    class Config:
        json_schema_extra = {
            "example": {
                "stock_filter": {"pe": {"min": 0, "max": 30}, "sector": ["银行"]},
                "timing": {"type": "ma_cross", "period": 20},
                "risk": {"max_position": 0.3, "stop_loss": 0.05}
            }
        }


# ── POST /filter ──

@router.post("/filter")
async def filter_stocks(req: FilterRequest):
    """
    多因子股票筛选
    接收 Go 后端转发的策略筛选请求
    超时 5 秒返回 code=50001
    """
    start = time.time()

    # 校验因子合法性
    for rule in req.rules:
        if get_factor(rule.factor) is None:
            return {
                "code": 400,
                "message": f"不支持的因子: {rule.factor}",
                "data": {"supported_factors": list(FACTOR_REGISTRY.keys())}
            }

    # 转换规则
    rules_dict = []
    for r in req.rules:
        d = {"type": r.type, "factor": r.factor}
        for field in ["min", "max", "value", "period", "threshold", "cross_type", "fast", "slow", "signal"]:
            v = getattr(r, field, None)
            if v is not None:
                d[field] = v
        rules_dict.append(d)

    try:
        result = engine.execute(
            rules=rules_dict,
            logic=req.logic,
            stocks=req.stocks,
            timeout=4.5,
        )
    except Exception as e:
        elapsed = time.time() - start
        if elapsed > 5:
            return {"code": 50001, "message": "筛选超时", "data": None}
        logger.error(f"[filter] 执行异常: {e}")
        return {"code": 500, "message": f"筛选异常: {str(e)[:100]}", "data": None}

    if result.get("code") != 0:
        return result

    data = result.get("data") or {}
    elapsed = round((time.time() - start) * 1000, 1)
    logger.info(f"[filter] {len(req.rules)} 规则, {req.logic}, {data.get('matched_count',0)} 命中, {elapsed}ms")
    return result


# ── POST /filter/parse ──

@router.post("/filter/parse")
async def parse_strategy(req: StrategyJSON):
    """
    解析前端策略编辑器 JSON
    转换为标准化的 rules 数组
    """
    parsed_rules = []
    warnings = []

    # stock_filter 解析
    sf = req.stock_filter
    for factor_name, params in sf.items():
        factor_def = get_factor(factor_name)
        if factor_def is None:
            warnings.append(f"未知因子: {factor_name}")
            continue

        rule = {"type": factor_def.category, "factor": factor_name}

        if factor_def.category == "financial":
            if isinstance(params, dict):
                rule["min"] = params.get("min", 0)
                rule["max"] = params.get("max", 999999)
        elif factor_def.category == "sector":
            if isinstance(params, list):
                rule["value"] = params
            elif isinstance(params, dict):
                rule["value"] = params.get("value", [])
        elif factor_def.category == "technical":
            if isinstance(params, dict):
                for k, v in params.items():
                    rule[k] = v

        parsed_rules.append(rule)

    # timing 解析
    timing = req.timing
    t_type = timing.get("type", "")
    if t_type == "ma_cross":
        parsed_rules.append({
            "type": "technical",
            "factor": "ma",
            "period": timing.get("period", 20),
            "cross_type": "golden",
        })
    elif t_type == "macd_golden":
        parsed_rules.append({
            "type": "technical",
            "factor": "macd",
            "fast": timing.get("fast", 12),
            "slow": timing.get("slow", 26),
            "signal": timing.get("signal", 9),
        })
    elif t_type == "rsi_oversold":
        parsed_rules.append({
            "type": "technical",
            "factor": "rsi",
            "period": timing.get("period", 14),
            "threshold": timing.get("threshold", 30),
        })

    # risk 仅做验证，不参与筛选
    risk_valid = True
    if req.risk:
        mp = req.risk.get("max_position", 0)
        sl = req.risk.get("stop_loss", 0)
        if not (0 < mp <= 1):
            warnings.append(f"max_position 应在 (0,1] 范围, 当前: {mp}")
            risk_valid = False
        if not (0 < sl <= 1):
            warnings.append(f"stop_loss 应在 (0,1] 范围, 当前: {sl}")
            risk_valid = False

    return {
        "code": 0,
        "data": {
            "parsed_rules": parsed_rules,
            "rules_count": len(parsed_rules),
            "validation": "ok" if (risk_valid and not warnings) else "warning",
            "warnings": warnings,
            "estimated_stocks": None,  # 需实际执行筛选后才知道
        }
    }