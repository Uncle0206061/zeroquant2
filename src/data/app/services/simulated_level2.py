# app/services/simulated_level2.py — 模拟 Level-2 数据源
# 用 akshare 轮询模拟 WebSocket 推送，在线/离线双模式

import asyncio
import logging
import random
from datetime import datetime
from typing import Callable, Optional

from app.config import settings
from app.cache import CacheManager as cache

logger = logging.getLogger(__name__)

# akshare 五档盘口字段映射（东方财富格式）
ORDERBOOK_FIELDS = {
    "bid1": "买一",   "bid1_vol": "买一量",
    "bid2": "买二",   "bid2_vol": "买二量",
    "bid3": "买三",   "bid3_vol": "买三量",
    "bid4": "买四",   "bid4_vol": "买四量",
    "bid5": "买五",   "bid5_vol": "买五量",
    "ask1": "卖一",   "ask1_vol": "卖一量",
    "ask2": "卖二",   "ask2_vol": "卖二量",
    "ask3": "卖三",   "ask3_vol": "卖三量",
    "ask4": "卖四",   "ask4_vol": "卖四量",
    "ask5": "卖五",   "ask5_vol": "卖五量",
}


class SimulatedLevel2:
    """
    模拟 Level-2 盘口推送
    - 生产模式: 每 3 秒轮询 akshare 真实数据
    - 测试模式: 生成随机 mock 数据
    """

    def __init__(self, interval: float = 3.0):
        self.interval = interval
        self._callbacks: dict[str, list[Callable]] = {}
        self._running = False
        self._task: Optional[asyncio.Task] = None
        self._use_mock: bool = True  # sandbox 中默认 mock

    # ── 订阅 ──

    def subscribe(self, stock_code: str, callback: Callable):
        if stock_code not in self._callbacks:
            self._callbacks[stock_code] = []
        self._callbacks[stock_code].append(callback)

    def unsubscribe(self, stock_code: str, callback: Callable):
        if stock_code in self._callbacks:
            self._callbacks[stock_code] = [
                cb for cb in self._callbacks[stock_code] if cb is not callback
            ]

    # ── 启动/停止 ──

    async def start(self, stock_codes: Optional[list[str]] = None):
        if self._running:
            return
        self._running = True
        if stock_codes:
            for code in stock_codes:
                if code not in self._callbacks:
                    self._callbacks[code] = []
        self._task = asyncio.create_task(self._poll_loop())
        logger.info(f"[sim-level2] 启动, 模式={'mock' if self._use_mock else 'akshare'}, 股票: {len(self._callbacks)} 只")

    async def stop(self):
        self._running = False
        if self._task:
            self._task.cancel()
        logger.info("[sim-level2] 已停止")

    # ── 轮询循环 ──

    async def _poll_loop(self):
        while self._running:
            codes = list(self._callbacks.keys())
            for code in codes:
                try:
                    data = await self._fetch_orderbook(code)
                    if data:
                        # 写入 Redis 缓存
                        cache.set_orderbook(code, data)
                        # 通知订阅者
                        for cb in self._callbacks.get(code, []):
                            try:
                                if asyncio.iscoroutinefunction(cb):
                                    await cb(data)
                                else:
                                    cb(data)
                            except Exception:
                                pass
                except Exception as e:
                    logger.warning(f"[sim-level2] {code} 获取失败: {e}")

            await asyncio.sleep(self.interval)

    async def _fetch_orderbook(self, stock_code: str) -> Optional[dict]:
        if self._use_mock:
            return self._generate_mock_orderbook(stock_code)
        return await self._fetch_real_orderbook(stock_code)

    # ── Mock 数据生成 ──

    def _generate_mock_orderbook(self, stock_code: str) -> dict:
        """生成随机五档盘口数据"""
        base = 10.0 + hash(stock_code) % 100  # 基于代码的确定性基价
        price = base + random.uniform(-0.5, 0.5)
        spread = random.uniform(0.01, 0.05)

        return {
            "symbol": stock_code,
            "type": "orderbook",
            "timestamp": datetime.now().isoformat(),
            "price": round(price, 2),
            "bids": [
                [round(price - spread * i, 2), random.randint(100, 5000)]
                for i in range(1, 6)
            ],
            "asks": [
                [round(price + spread * i, 2), random.randint(100, 5000)]
                for i in range(1, 6)
            ],
        }

    # ── 真实数据获取 ──

    async def _fetch_real_orderbook(self, stock_code: str) -> Optional[dict]:
        """从 akshare 拉取真实五档盘口"""
        try:
            import akshare as ak
            loop = asyncio.get_event_loop()
            df = await loop.run_in_executor(None, ak.stock_zh_a_spot_em)
            if df is None:
                return None

            cols = df.columns.tolist()
            code_col = next((c for c in cols if '代码' in c), None)
            if code_col is None:
                return None

            row = df[df[code_col].astype(str) == str(stock_code)]
            if row.empty:
                return None

            r = row.iloc[0]

            # 提取五档
            def get_val(cn_name: str) -> float:
                col = next((c for c in cols if cn_name in c), None)
                return float(r[col]) if col and r[col] else 0.0

            bids = []
            asks = []
            for i in range(1, 6):
                bids.append([get_val(f"买{i}"), get_val(f"买{i}量")])
                asks.append([get_val(f"卖{i}"), get_val(f"卖{i}量")])

            # 去除非正价档位
            bids = [b for b in bids if b[0] > 0]
            asks = [a for a in asks if a[0] > 0]

            return {
                "symbol": stock_code,
                "type": "orderbook",
                "timestamp": datetime.now().isoformat(),
                "price": float(r.get("最新价", 0) if "最新价" in r.index else 0),
                "bids": bids,
                "asks": asks,
            }
        except Exception as e:
            logger.warning(f"[sim-level2] akshare 盘口获取失败: {e}")
            return None


# 全局单例
simulated_level2 = SimulatedLevel2()