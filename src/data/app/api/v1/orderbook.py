# app/api/v1/orderbook.py — Level-2 五档盘口接口

from fastapi import APIRouter
from app.schemas.base import ApiResponse
from app.cache import CacheManager as cache

router = APIRouter()


@router.get("/orderbook/{stock_code}", response_model=ApiResponse)
async def get_orderbook(stock_code: str):
    """
    五档买卖盘口数据
    - 优先读 Redis 缓存（3s TTL）
    - 缓存 miss 返回空盘口
    """
    data = cache.get_orderbook(stock_code)

    if data:
        return {
            "code": 0,
            "message": "success",
            "data": {
                "stock_code": stock_code,
                "bids": data.get("bids", []),
                "asks": data.get("asks", []),
                "price": data.get("price"),
                "timestamp": data.get("timestamp"),
                "source": "level2-cache",
            },
        }

    return {
        "code": 0,
        "message": "success",
        "data": {
            "stock_code": stock_code,
            "bids": [],
            "asks": [],
            "price": None,
            "timestamp": None,
            "source": "cache-miss",
        },
    }


@router.post("/orderbook/subscribe")
async def subscribe_orderbook(stock_codes: list[str]):
    """订阅盘口推送 — 启动模拟 Level-2 轮询"""
    from app.services.simulated_level2 import simulated_level2

    for code in stock_codes:
        simulated_level2.subscribe(code, lambda d: None)

    if not simulated_level2._running:
        import asyncio
        asyncio.create_task(simulated_level2.start(stock_codes))

    return {"code": 0, "message": "success", "data": {"subscribed": stock_codes, "count": len(stock_codes)}}