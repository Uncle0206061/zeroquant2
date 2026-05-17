# app/api/v1/monitor.py — 监控/告警/指标  TC-M2-02

import json
import time
import logging
from datetime import datetime, timedelta
from typing import Optional
from fastapi import APIRouter, WebSocket, WebSocketDisconnect

from app.collectors.akshare_collector import collector as akc
from app.collectors.tushare_collector import TushareCollector
from app.collectors.level2_collector import level2
from app.cache import CacheManager as cache
from app.models.base import get_engine

logger = logging.getLogger(__name__)
router = APIRouter()

# ── 告警状态机 ──

class AlertManager:
    """
    告警管理器
    - 记录各数据源失败起始时间
    - 根据持续时间判定告警级别
    - 通过 WebSocket 广播 alarm 事件
    """

    def __init__(self):
        self._fail_since: dict[str, Optional[float]] = {}  # source → 失败起始时间戳
        self._alerts: list[dict] = []  # 告警历史（最近100条）
        self._ws_clients: list[WebSocket] = []
        self._metrics: dict = {
            "kline_fetches": 0, "kline_failures": 0,
            "spot_fetches": 0, "spot_failures": 0,
            "orderbook_pushes": 0,
            "last_minute_throughput": 0,
        }
        self._throughput_window: list[tuple[float, int]] = []  # (timestamp, count)

    # ── 状态上报 ──

    def report(self, source: str, status: str):
        """数据源状态上报"""
        now = time.time()
        if status == "ok":
            if source in self._fail_since and self._fail_since[source] is not None:
                duration = now - self._fail_since[source]
                if duration > 0:
                    self._add_alert(source, "recovered", f"恢复, 中断{duration:.0f}s")
                self._fail_since[source] = None
        else:
            if source not in self._fail_since or self._fail_since[source] is None:
                self._fail_since[source] = now
                self._add_alert(source, "warning", f"{source} 数据源异常")

            if self._fail_since[source] is not None:
                duration = now - self._fail_since[source]
                if duration > 300:  # 5分钟
                    self._add_alert(source, "critical", f"{source} 已中断 {duration:.0f}s")
                elif duration > 30:  # 30秒
                    self._add_alert(source, "warning", f"{source} 已中断 {duration:.0f}s")

    def _add_alert(self, source: str, level: str, message: str):
        alert = {
            "source": source, "level": level, "message": message,
            "timestamp": datetime.now().isoformat(),
        }
        self._alerts.insert(0, alert)
        if len(self._alerts) > 100:
            self._alerts = self._alerts[:100]
        logger.warning(f"[ALERT] [{level.upper()}] {message}")

        # WebSocket 广播（异步）
        import asyncio
        for ws in self._ws_clients:
            try:
                asyncio.create_task(ws.send_json(alert))
            except Exception:
                pass

    # ── WebSocket ──

    def ws_connect(self, ws: WebSocket):
        self._ws_clients.append(ws)

    def ws_disconnect(self, ws: WebSocket):
        if ws in self._ws_clients:
            self._ws_clients.remove(ws)

    # ── 吞吐量计数 ──

    def count_fetch(self, source: str, success: bool):
        if source == "kline":
            self._metrics["kline_fetches"] += 1
            if not success:
                self._metrics["kline_failures"] += 1
        elif source == "spot":
            self._metrics["spot_fetches"] += 1
            if not success:
                self._metrics["spot_failures"] += 1
        elif source == "orderbook":
            self._metrics["orderbook_pushes"] += 1

        # 滑动窗口统计
        now = time.time()
        self._throughput_window.append((now, 1))
        self._throughput_window = [
            (t, c) for t, c in self._throughput_window if now - t < 60
        ]
        self._metrics["last_minute_throughput"] = sum(c for _, c in self._throughput_window)

    # ── 指标计算 ──

    def get_metrics(self) -> dict:
        """数据质量指标"""
        kf = self._metrics["kline_fetches"]
        ke = self._metrics["kline_failures"]
        completeness = round(1 - (ke / kf), 4) if kf > 0 else 1.0

        # 数据源可用率
        ak_health = akc.test_connectivity()
        ak_uptime = 1.0 if ak_health["status"] == "ok" else (0.5 if ak_health["status"] == "degraded" else 0.0)

        return {
            "kline_completeness": completeness,
            "orderbook_latency_ms": None,  # 需要真实 Level-2 才能测量
            "datasource_uptime": {
                "akshare": ak_uptime,
                "tushare": 0.0,  # 未启用
            },
            "collector_throughput_per_min": self._metrics["last_minute_throughput"],
            "total_fetches": kf + self._metrics["spot_fetches"],
            "total_failures": ke + self._metrics["spot_failures"],
            "orderbook_pushes": self._metrics["orderbook_pushes"],
        }

    def get_alerts(self, limit: int = 20) -> list[dict]:
        return self._alerts[:limit]


# 全局单例
alerts = AlertManager()


# ── REST 端点 ──

@router.get("/monitor/status")
async def monitor_status():
    ak_health = akc.test_connectivity()
    ts_health = TushareCollector.test_connectivity()
    l2_health = level2.health
    redis_ok = cache.is_available()
    pg_ok = False
    try:
        with get_engine().connect() as conn:
            conn.execute(__import__('sqlalchemy').text("SELECT 1"))
        pg_ok = True
    except Exception:
        pass

    # 上报状态
    alerts.report("akshare", ak_health["status"])
    alerts.report("tushare", ts_health.get("status", "disabled"))

    sources = [ak_health["status"], ts_health.get("status", "ok")]
    ok_count = sum(1 for s in sources if s == "ok")
    alert_level = "critical" if ok_count == 0 else ("warning" if ok_count < len(sources) else "ok")

    return {
        "code": 0,
        "message": "success",
        "data": {
            "data_sources": {"akshare": ak_health["status"], "tushare": ts_health.get("status", "disabled")},
            "level2": l2_health,
            "redis_health": redis_ok,
            "postgres_health": pg_ok,
            "collector_running": {
                "orderbook": l2_health["connected"],
                "kline": ak_health["status"] == "ok",
                "sector": ak_health["status"] == "ok",
            },
            "alert_level": alert_level,
            "last_heartbeat": l2_health.get("last_heartbeat"),
            "checked_at": str(datetime.now()),
        },
    }


@router.get("/monitor/metrics")
async def monitor_metrics():
    return {"code": 0, "message": "success", "data": alerts.get_metrics()}


@router.get("/monitor/alerts")
async def monitor_alerts(limit: int = 20):
    return {"code": 0, "message": "success", "data": {"alerts": alerts.get_alerts(limit), "total": len(alerts._alerts)}}


@router.websocket("/monitor/ws")
async def monitor_websocket(ws: WebSocket):
    await ws.accept()
    alerts.ws_connect(ws)
    try:
        while True:
            await ws.receive_text()  # keep-alive
    except WebSocketDisconnect:
        alerts.ws_disconnect(ws)
    except Exception:
        alerts.ws_disconnect(ws)