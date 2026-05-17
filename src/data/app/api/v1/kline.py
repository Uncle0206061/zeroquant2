# app/api/v1/kline.py — K线数据接口（预置，TC-M2-01 实现）

from fastapi import APIRouter, Query
from app.schemas.base import ApiResponse
from app.collectors.akshare_collector import collector as ak_collector
from datetime import datetime, timedelta

router = APIRouter()


@router.get("/kline/{symbol}", response_model=ApiResponse)
async def get_kline(
    symbol: str,
    period: str = Query("daily", description="daily/weekly/monthly"),
    start_date: str = Query(None, description="起始日期 YYYYMMDD"),
    end_date: str = Query(None, description="结束日期 YYYYMMDD"),
    adjust: str = Query("", description="复权: ''/'qfq'/'hfq'"),
):
    """
    个股 K 线数据
    TODO: TC-M2-01 — 支持分钟K、数据落盘 PostgreSQL、缓存 Redis
    """
    if not start_date:
        start_date = (datetime.now() - timedelta(days=365)).strftime("%Y%m%d")
    if not end_date:
        end_date = datetime.now().strftime("%Y%m%d")

    df = ak_collector.get_kline(symbol, period, start_date, end_date, adjust)
    if df is None:
        return {"code": 50000, "message": "K线数据拉取失败", "data": None}

    return {
        "code": 0,
        "data": {
            "symbol": symbol,
            "period": period,
            "count": len(df),
            "rows": df.to_dict("records"),
        },
    }