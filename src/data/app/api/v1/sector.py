# app/api/v1/sector.py — 板块数据接口（预置，TC-M2-01 实现）

from fastapi import APIRouter, Query
from app.schemas.base import ApiResponse
from app.collectors.akshare_collector import collector as ak_collector

router = APIRouter()


@router.get("/sector/industry", response_model=ApiResponse)
async def get_sector_industry():
    """行业板块列表
    TODO: TC-M2-01 — Redis 缓存 60s
    """
    df = ak_collector.get_sector_industry()
    if df is None:
        return {"code": 50000, "message": "板块数据拉取失败", "data": None}
    return {"code": 0, "message": "success", "data": {"count": len(df), "rows": df.to_dict("records")}}


@router.get("/sector/concept", response_model=ApiResponse)
async def get_sector_concept():
    """概念板块列表
    TODO: TC-M2-01 — Redis 缓存 60s
    """
    df = ak_collector.get_sector_concept()
    if df is None:
        return {"code": 50000, "message": "板块数据拉取失败", "data": None}
    return {"code": 0, "message": "success", "data": {"count": len(df), "rows": df.to_dict("records")}}