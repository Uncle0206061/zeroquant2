# app/api/v1/health.py — 健康检查接口（M1 交付物 #1）
# GET /data/v1/health → {code:0, data:{status:"ok"}}

from fastapi import APIRouter

router = APIRouter()


@router.get("/health")
async def health_check():
    """健康检查 — 三端联调第一枪"""
    return {
        "code": 0,
        "message": "success",
        "data": {
            "status": "ok",
            "service": "data-service",
        },
    }