# app/collectors/tushare_collector.py
# Tushare Pro 采集器 — 备用数据源 / 高精度校验
# TODO: TC-M2-01 — 实现日线、复权因子、财务数据采集

import logging
from typing import Optional
import pandas as pd

from app.config import settings

logger = logging.getLogger(__name__)


class TushareCollector:
    """Tushare Pro 数据采集器（备用源）"""

    _pro = None

    @classmethod
    def _get_pro(cls):
        if cls._pro is None and settings.TUSHARE_TOKEN:
            import tushare as ts
            ts.set_token(settings.TUSHARE_TOKEN)
            cls._pro = ts.pro_api()
        return cls._pro

    @classmethod
    def get_kline(
        cls, ts_code: str, start_date: str, end_date: str
    ) -> Optional[pd.DataFrame]:
        """Tushare 日 K 线"""
        pro = cls._get_pro()
        if pro is None:
            return None
        try:
            df = pro.daily(ts_code=ts_code, start_date=start_date, end_date=end_date)
            logger.info(f"[tushare] K线 {ts_code}: {len(df)} 行")
            return df
        except Exception as e:
            logger.error(f"[tushare] K线拉取失败 {ts_code}: {e}")
            return None

    @classmethod
    def get_stock_basic(cls) -> Optional[pd.DataFrame]:
        """A 股股票列表"""
        pro = cls._get_pro()
        if pro is None:
            return None
        try:
            df = pro.stock_basic(exchange='', list_status='L',
                                 fields='ts_code,symbol,name,area,industry,list_date')
            return df
        except Exception as e:
            logger.error(f"[tushare] 股票列表拉取失败: {e}")
            return None

    @classmethod
    def test_connectivity(cls) -> dict:
        """连通性自检"""
        pro = cls._get_pro()
        if pro is None:
            return {"source": "tushare", "status": "disabled", "reason": "token 未配置"}
        try:
            df = pro.stock_basic(exchange='', list_status='L', limit=1)
            return {"source": "tushare", "status": "ok", "tests": {"stock_basic": "ok"}}
        except Exception as e:
            return {"source": "tushare", "status": "error", "reason": str(e)}