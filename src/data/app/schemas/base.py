# app/schemas/base.py — Pydantic 响应/请求模型

from typing import Any, Optional
from pydantic import BaseModel


class ApiResponse(BaseModel):
    """统一 API 响应格式 — 与 Go 后端对齐"""
    code: int = 0
    message: str = "success"
    data: Any = None

    class Config:
        json_schema_extra = {
            "example": {
                "code": 0,
                "message": "success",
                "data": {"status": "ok"}
            }
        }


class ErrorResponse(BaseModel):
    """错误响应"""
    code: int
    message: str
    data: Optional[dict] = None


class HealthResponse(BaseModel):
    """健康检查响应"""
    status: str = "ok"
    service: str = "data-service"


class KlineRequest(BaseModel):
    """K 线查询参数"""
    symbol: str
    period: str = "daily"
    start_date: Optional[str] = None
    end_date: Optional[str] = None
    adjust: Optional[str] = ""


class FilterRequest(BaseModel):
    """多因子筛选请求"""
    factors: list[dict] = []
    logic: str = "and"  # and / or
    limit: int = 50