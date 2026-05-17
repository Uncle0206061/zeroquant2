# app/collectors/level2_collector.py
# Level-2 盘口数据采集器 — WebSocket 实时流

import asyncio
import json
import logging
from datetime import datetime
from typing import Optional, Callable

from app.config import settings
from app.cache import CacheManager as cache

logger = logging.getLogger(__name__)


class Level2Collector:
    """
    Level-2 实时数据采集器
    - 协议: WebSocket 长连接
    - 数据: 五档盘口、逐笔委托、逐笔成交
    - 特性: 心跳10s、自动重连(最多10次,间隔2s)、断线告警
    """

    def __init__(self):
        self.ws_url: str = settings.LEVEL2_WS_URL
        self.connected: bool = False
        self._ws = None
        self._subscribers: dict[str, list[Callable]] = {}  # stock_code → [callbacks]
        self._heartbeat_task = None
        self._reconnect_count: int = 0
        self._last_heartbeat: Optional[datetime] = None

    # ── 连接管理 ──

    async def connect(self) -> bool:
        if not self.ws_url:
            logger.warning("[level2] WS URL 未配置，跳过连接")
            return False

        try:
            import websockets
            self._ws = await websockets.connect(
                self.ws_url,
                ping_interval=settings.LEVEL2_HEARTBEAT,
                ping_timeout=5,
                close_timeout=3,
            )
            self.connected = True
            self._reconnect_count = 0
            self._last_heartbeat = datetime.now()
            logger.info(f"[level2] 连接成功: {self.ws_url}")

            # 启动心跳任务
            self._heartbeat_task = asyncio.create_task(self._heartbeat_loop())

            # 重新订阅所有已注册股票
            for code in list(self._subscribers.keys()):
                await self._send_subscribe(code)

            return True
        except Exception as e:
            logger.error(f"[level2] 连接失败: {e}")
            self.connected = False
            return False

    async def _heartbeat_loop(self):
        while self.connected:
            await asyncio.sleep(settings.LEVEL2_HEARTBEAT)
            if self._ws:
                try:
                    pong = await self._ws.ping()
                    await asyncio.wait_for(pong, timeout=3)
                    self._last_heartbeat = datetime.now()
                except Exception:
                    logger.warning("[level2] 心跳超时，触发重连")
                    await self._reconnect()

    async def _reconnect(self):
        self._reconnect_count += 1
        if self._reconnect_count > settings.LEVEL2_MAX_RECONNECT:
            logger.error(f"[level2] 重连次数超限({settings.LEVEL2_MAX_RECONNECT})，停止重连")
            self.connected = False
            return

        interval = settings.LEVEL2_RECONNECT_INTERVAL
        logger.info(f"[level2] 第{self._reconnect_count}次重连，等待{interval}s")
        await asyncio.sleep(interval)
        await self.connect()

    async def _send_subscribe(self, stock_code: str):
        if self._ws and self.connected:
            try:
                msg = json.dumps({"action": "subscribe", "symbol": stock_code})
                await self._ws.send(msg)
            except Exception as e:
                logger.error(f"[level2] 订阅失败 {stock_code}: {e}")

    # ── 订阅 ──

    async def subscribe_orderbook(self, stock_code: str, callback: Optional[Callable] = None):
        if stock_code not in self._subscribers:
            self._subscribers[stock_code] = []
        if callback and callback not in self._subscribers[stock_code]:
            self._subscribers[stock_code].append(callback)
        if self.connected:
            await self._send_subscribe(stock_code)

    async def subscribe_tick(self, stock_code: str, callback: Optional[Callable] = None):
        await self.subscribe_orderbook(stock_code, callback)

    # ── 数据接收 ──

    async def listen(self):
        """持续监听 WebSocket 消息，分发到回调并写入 Redis"""
        if not self._ws or not self.connected:
            logger.warning("[level2] 未连接，无法监听")
            return

        try:
            async for raw in self._ws:
                try:
                    data = json.loads(raw)
                    msg_type = data.get("type", "")
                    symbol = data.get("symbol", "")

                    # 写入 Redis 缓存（3秒TTL）
                    if msg_type == "orderbook":
                        cache.set_orderbook(symbol, data)
                    elif msg_type == "tick":
                        # 逐笔数据量大，先缓存再批量写入DB
                        cache.set("tick", f"{symbol}:{datetime.now().timestamp()}", data, ttl=90*86400)

                    # 分发给订阅回调
                    for cb in self._subscribers.get(symbol, []):
                        try:
                            cb(data)
                        except Exception:
                            pass

                except json.JSONDecodeError:
                    continue
        except Exception as e:
            logger.error(f"[level2] 监听异常: {e}")
            await self._reconnect()

    # ── 断开 ──

    async def disconnect(self):
        self.connected = False
        if self._heartbeat_task:
            self._heartbeat_task.cancel()
        if self._ws:
            await self._ws.close()
        logger.info("[level2] 已断开")

    # ── 健康 ──

    @property
    def health(self) -> dict:
        return {
            "connected": self.connected,
            "reconnect_count": self._reconnect_count,
            "last_heartbeat": str(self._last_heartbeat) if self._last_heartbeat else None,
            "subscribers": len(self._subscribers),
        }


# 全局单例
level2 = Level2Collector()