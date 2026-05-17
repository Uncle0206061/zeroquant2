# app/cache.py — Redis 缓存管理器（带优雅降级）

import json
import logging
from typing import Optional, Any
from datetime import timedelta

from app.config import settings

logger = logging.getLogger(__name__)


class CacheManager:
    """
    Redis 缓存管理器
    - Redis 不可用时优雅降级（返回 None，不崩溃）
    - 统一 key 前缀: data:
    """

    _client = None
    _available = None  # None=未检测, True=可用, False=不可用

    @classmethod
    def _get_client(cls):
        if cls._client is not None:
            return cls._client
        try:
            import redis as redis_lib
            cls._client = redis_lib.from_url(
                settings.REDIS_URL,
                socket_connect_timeout=1,
                socket_timeout=1,
                decode_responses=True,
            )
            cls._client.ping()
            cls._available = True
            logger.info("[cache] Redis 连接成功")
            return cls._client
        except Exception as e:
            cls._available = False
            logger.warning(f"[cache] Redis 不可用 ({e})，使用降级模式")
            return None

    @classmethod
    def is_available(cls) -> bool:
        if cls._available is None:
            cls._get_client()
        return cls._available or False

    @classmethod
    def _key(cls, namespace: str, identifier: str) -> str:
        return f"{settings.REDIS_KEY_PREFIX}{namespace}:{identifier}"

    # ── 基础操作 ──

    @classmethod
    def get(cls, namespace: str, identifier: str) -> Optional[dict]:
        client = cls._get_client()
        if client is None:
            return None
        try:
            raw = client.get(cls._key(namespace, identifier))
            return json.loads(raw) if raw else None
        except Exception as e:
            logger.error(f"[cache] GET 失败: {e}")
            return None

    @classmethod
    def set(cls, namespace: str, identifier: str, data: dict, ttl: int = 60):
        client = cls._get_client()
        if client is None:
            return
        try:
            key = cls._key(namespace, identifier)
            client.setex(key, ttl, json.dumps(data, ensure_ascii=False, default=str))
        except Exception as e:
            logger.warning(f"[cache] SET 失败: {e}")

    @classmethod
    def delete(cls, namespace: str, identifier: str):
        client = cls._get_client()
        if client is None:
            return
        try:
            client.delete(cls._key(namespace, identifier))
        except Exception:
            pass

    # ── 业务方法 ──

    @classmethod
    def get_kline(cls, symbol: str, period: str, start: str, end: str) -> Optional[dict]:
        key = f"{symbol}:{period}:{start}:{end}"
        return cls.get("kline", key)

    @classmethod
    def set_kline(cls, symbol: str, period: str, start: str, end: str, data: dict):
        key = f"{symbol}:{period}:{start}:{end}"
        cls.set("kline", key, data, ttl=300)  # 5分钟缓存

    @classmethod
    def get_spot(cls, stock_code: str) -> Optional[dict]:
        return cls.get("spot", stock_code)

    @classmethod
    def set_spot(cls, stock_code: str, data: dict):
        cls.set("spot", stock_code, data, ttl=5)  # 5秒TTL

    @classmethod
    def get_orderbook(cls, stock_code: str) -> Optional[dict]:
        return cls.get("orderbook", stock_code)

    @classmethod
    def set_orderbook(cls, stock_code: str, data: dict):
        cls.set("orderbook", stock_code, data, ttl=3)  # 3秒TTL

    @classmethod
    def get_sector(cls, sector_type: str) -> Optional[dict]:
        return cls.get("sector", sector_type)

    @classmethod
    def set_sector(cls, sector_type: str, data: dict):
        cls.set("sector", sector_type, data, ttl=60)  # 60秒TTL

    @classmethod
    def health(cls) -> dict:
        try:
            client = cls._get_client()
            if client is None:
                return {"redis": False, "error": "连接失败"}
            client.ping()
            info = client.info("memory")
            return {
                "redis": True,
                "used_memory_mb": round(info.get("used_memory_human", 0), 2),
            }
        except Exception as e:
            return {"redis": False, "error": str(e)[:80]}