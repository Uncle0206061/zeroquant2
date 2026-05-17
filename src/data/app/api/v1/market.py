# app/api/v1/market.py — 实时行情接口（预置，TC-M2-01 实现）

from fastapi import APIRouter, Query
from app.schemas.base import ApiResponse
from app.collectors.akshare_collector import collector as ak_collector

router = APIRouter()


@router.get("/market/spot", response_model=ApiResponse)
async def get_market_spot():
    """全量 A 股实时行情
    TODO: TC-M2-01 — 接入 Redis 缓存 + 数据格式化
    """
    df = ak_collector.get_spot_all()
    if df is None:
        return {"code": 50000, "message": "数据源不可用", "data": None}
    return {"code": 0, "message": "success", "data": {"count": len(df), "rows": df.head(20).to_dict("records")}}


@router.get("/market/{stock_code}", response_model=ApiResponse)
async def get_market_by_code(stock_code: str):
    """单只股票实时行情
    TODO: TC-M2-01 — 接入 Redis 缓存
    """
    row = ak_collector.get_spot(stock_code)
    if row is None:
        return {"code": 40400, "message": f"未找到股票 {stock_code}", "data": None}
    return {"code": 0, "message": "success", "data": row}