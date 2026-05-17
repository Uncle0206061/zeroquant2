# app/services/recovery.py — 数据恢复服务  TC-M2-02

import logging
from datetime import datetime, timedelta
from typing import Optional
import pandas as pd

from app.collectors.akshare_collector import collector as akc

logger = logging.getLogger(__name__)


class RecoveryService:
    """
    数据恢复服务
    - 启动时检测 K线数据缺口
    - 自动从 akshare 拉取缺失区间
    - 恢复完成写日志
    """

    def __init__(self):
        self.recovery_log: list[dict] = []

    def detect_kline_gaps(self, symbol: str, existing_dates: list[str],
                          start: str, end: str) -> list[tuple[str, str]]:
        """
        检测 K线缺口区间
        返回: [(gap_start, gap_end), ...]
        """
        if not existing_dates:
            return [(start, end)]

        # 生成应有交易日列表（简化：用自然日）
        s = datetime.strptime(start, "%Y%m%d")
        e = datetime.strptime(end, "%Y%m%d")
        expected = set()
        current = s
        while current <= e:
            if current.weekday() < 5:  # 周一到周五
                expected.add(current.strftime("%Y%m%d"))
            current += timedelta(days=1)

        existing_set = set(existing_dates)
        missing = sorted(expected - existing_set)

        if not missing:
            return []

        # 合并连续缺失日期为区间
        gaps = []
        gap_start = missing[0]
        gap_end = missing[0]
        for d in missing[1:]:
            d_prev = datetime.strptime(gap_end, "%Y%m%d") + timedelta(days=1)
            if d == d_prev.strftime("%Y%m%d"):
                gap_end = d
            else:
                gaps.append((gap_start, gap_end))
                gap_start = gap_end = d
        gaps.append((gap_start, gap_end))
        return gaps

    def fill_kline_gaps(self, symbol: str, gaps: list[tuple[str, str]],
                        period: str = "daily") -> dict:
        """
        补全 K线缺口数据
        返回: {filled_count, errors}
        """
        filled = 0
        errors = []

        for gap_start, gap_end in gaps:
            try:
                df = akc.get_kline(symbol, period, gap_start, gap_end, use_cache=False)
                if df is not None and len(df) > 0:
                    filled += len(df)
                    logger.info(f"[recovery] {symbol} 补全 {gap_start}-{gap_end}: {len(df)} 条")
                    self.recovery_log.append({
                        "symbol": symbol, "period": period,
                        "gap": f"{gap_start}-{gap_end}",
                        "filled": len(df),
                        "timestamp": datetime.now().isoformat(),
                    })
                else:
                    errors.append(f"{gap_start}-{gap_end}: 无数据")
            except Exception as e:
                err = f"{gap_start}-{gap_end}: {e}"
                errors.append(err)
                logger.error(f"[recovery] {symbol} 补全失败: {e}")

        return {"filled_count": filled, "errors": errors}

    def auto_recover(self, symbols: list[str], period: str = "daily",
                     lookback_days: int = 365) -> dict:
        """
        自动恢复：检测缺口 → 拉取数据
        """
        end = datetime.now().strftime("%Y%m%d")
        start = (datetime.now() - timedelta(days=lookback_days)).strftime("%Y%m%d")

        total_filled = 0
        all_results = {}

        for symbol in symbols:
            try:
                # 拉取已有数据
                existing = akc.get_kline(symbol, period, start, end, use_cache=False)
                existing_dates = []
                if existing is not None and len(existing) > 0:
                    date_col = next((c for c in existing.columns if '日期' in str(c)), None)
                    if date_col:
                        existing_dates = existing[date_col].astype(str).tolist()

                gaps = self.detect_kline_gaps(symbol, existing_dates, start, end)
                if gaps:
                    result = self.fill_kline_gaps(symbol, gaps, period)
                    total_filled += result["filled_count"]
                    all_results[symbol] = result
            except Exception as e:
                logger.error(f"[recovery] {symbol} 恢复失败: {e}")
                all_results[symbol] = {"filled_count": 0, "errors": [str(e)]}

        logger.info(f"[recovery] 完成: {total_filled} 条补全, {len(symbols)} 只股票")
        return {
            "total_filled": total_filled,
            "symbols_checked": len(symbols),
            "details": all_results,
            "timestamp": datetime.now().isoformat(),
        }


recovery_service = RecoveryService()