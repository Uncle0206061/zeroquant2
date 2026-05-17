# app/config.py — ZeroQuant 数据服务配置中心
# 所有配置从环境变量读取，提供合理默认值用于本地开发

import os
from dotenv import load_dotenv

load_dotenv()


class Settings:
    """全局配置单例"""

    # ---- 服务 ----
    APP_NAME: str = "ZeroQuant Data Service"
    APP_VERSION: str = "0.1.0"
    PORT: int = int(os.getenv("DATA_SERVICE_PORT", "8081"))
    DEBUG: bool = os.getenv("DEBUG", "false").lower() == "true"

    # ---- PostgreSQL ----
    DATABASE_URL: str = os.getenv(
        "DATABASE_URL",
        "postgresql://postgres:password@localhost:5432/data"
    )
    DB_TABLE_PREFIX: str = "data_"  # 所有表统一前缀，与 Go 后端 biz_ 隔离

    # ---- Redis ----
    REDIS_URL: str = os.getenv("REDIS_URL", "redis://localhost:6379/1")
    REDIS_KEY_PREFIX: str = "data:"

    # ---- 数据源开关 ----
    AKSHARE_ENABLED: bool = os.getenv("AKSHARE_ENABLED", "true").lower() == "true"
    TUSHARE_ENABLED: bool = os.getenv("TUSHARE_ENABLED", "false").lower() == "true"
    LEVEL2_ENABLED: bool = os.getenv("LEVEL2_ENABLED", "false").lower() == "true"

    # ---- Tushare ----
    TUSHARE_TOKEN: str = os.getenv("TUSHARE_TOKEN", "")

    # ---- Level-2 WebSocket ----
    LEVEL2_WS_URL: str = os.getenv("LEVEL2_WS_URL", "")
    LEVEL2_HEARTBEAT: int = int(os.getenv("LEVEL2_HEARTBEAT", "10"))
    LEVEL2_MAX_RECONNECT: int = int(os.getenv("LEVEL2_MAX_RECONNECT", "10"))
    LEVEL2_RECONNECT_INTERVAL: int = int(os.getenv("LEVEL2_RECONNECT_INTERVAL", "2"))

    # ---- 数据源容错 ----
    SOURCE_SWITCH_TIMEOUT: int = 1       # 主源断开后 1 秒内切换
    DATA_RETRY_MAX: int = 3              # 最大重试次数
    DATA_RETRY_BACKOFF: float = 0.5      # 重试退避因子

    # ---- 缓存 TTL ----
    CACHE_QUOTE_TTL: int = 5             # 行情缓存 5 秒
    CACHE_ORDERBOOK_TTL: int = 3         # 盘口缓存 3 秒
    CACHE_SECTOR_TTL: int = 60           # 板块缓存 60 秒

    # ---- 数据保留 ----
    TICK_RETENTION_DAYS: int = 90        # 逐笔保留 90 天
    ORDERBOOK_CACHE_DAYS: int = 3        # 盘口缓存 3 天


settings = Settings()