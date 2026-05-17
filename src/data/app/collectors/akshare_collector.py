# app/collectors/akshare_collector.py
# akshare 数据采集器 + 缓存/数据库双写

import logging
from typing import Optional
from datetime import datetime

import akshare as ak
import pandas as pd

from app.config import settings
from app.cache import CacheManager as cache

logger = logging.getLogger(__name__)


class AkshareCollector:

    # ── 行情 ──

    @staticmethod
    def get_spot_all(use_cache: bool = True) -> Optional[pd.DataFrame]:
        if use_cache:
            cached = cache.get_spot("all")
            if cached and "rows" in cached:
                return pd.DataFrame(cached["rows"])
        try:
            df = ak.stock_zh_a_spot_em()
            logger.info(f"[akshare] spot: {len(df)} rows")
            if use_cache:
                cache.set_spot("all", {"rows": df.head(500).to_dict("records"), "count": len(df), "ts": str(datetime.now())})
            return df
        except Exception as e:
            logger.error(f"[akshare] spot failed: {e}")
            return None

    @staticmethod
    def get_spot(stock_code: str) -> Optional[dict]:
        cached = cache.get_spot(stock_code)
        if cached:
            return cached
        df = AkshareCollector.get_spot_all()
        if df is None:
            return None
        cols = df.columns.tolist()
        code_col = next((c for c in cols if '代码' in c), None)
        if code_col is None:
            return None
        row = df[df[code_col].astype(str) == str(stock_code)]
        if row.empty:
            return None
        result = row.iloc[0].to_dict()
        cache.set_spot(stock_code, {k: str(v) for k, v in result.items()})
        return result

    # ── K线 ──

    @staticmethod
    def get_kline(
        symbol: str, period: str = "daily",
        start_date: str = "20250101", end_date: str = "20260514",
        adjust: str = "", use_cache: bool = True
    ) -> Optional[pd.DataFrame]:
        if use_cache:
            cached = cache.get_kline(symbol, period, start_date, end_date)
            if cached and "rows" in cached:
                return pd.DataFrame(cached["rows"])
        try:
            df = ak.stock_zh_a_hist(
                symbol=symbol, period=period,
                start_date=start_date, end_date=end_date, adjust=adjust
            )
            logger.info(f"[akshare] kline {symbol} {period}: {len(df)} rows")
            if use_cache and df is not None and len(df) > 0:
                cache.set_kline(symbol, period, start_date, end_date, {
                    "rows": df.to_dict("records"), "count": len(df)
                })
            return df
        except Exception as e:
            logger.error(f"[akshare] kline failed {symbol}: {e}")
            return None

    # ── 板块 ──

    @staticmethod
    def get_sector_industry(use_cache: bool = True) -> Optional[pd.DataFrame]:
        if use_cache:
            cached = cache.get_sector("industry")
            if cached and "rows" in cached:
                return pd.DataFrame(cached["rows"])
        try:
            df = ak.stock_board_industry_name_em()
            logger.info(f"[akshare] industry: {len(df)} sectors")
            if use_cache:
                cache.set_sector("industry", {"rows": df.to_dict("records"), "count": len(df)})
            return df
        except Exception as e:
            logger.error(f"[akshare] sector failed: {e}")
            return None

    @staticmethod
    def get_sector_concept(use_cache: bool = True) -> Optional[pd.DataFrame]:
        if use_cache:
            cached = cache.get_sector("concept")
            if cached and "rows" in cached:
                return pd.DataFrame(cached["rows"])
        try:
            df = ak.stock_board_concept_name_em()
            logger.info(f"[akshare] concept: {len(df)} sectors")
            if use_cache:
                cache.set_sector("concept", {"rows": df.to_dict("records"), "count": len(df)})
            return df
        except Exception as e:
            logger.error(f"[akshare] concept failed: {e}")
            return None

    @staticmethod
    def get_fund_flow(stock_code: str) -> Optional[pd.DataFrame]:
        try:
            mkt = "sh" if str(stock_code).startswith(("6", "9")) else "sz"
            df = ak.stock_individual_fund_flow(stock=stock_code, market=mkt)
            return df
        except Exception as e:
            logger.error(f"[akshare] fund_flow failed {stock_code}: {e}")
            return None

    # ── 健康自检 ──

    @staticmethod
    def test_connectivity() -> dict:
        result = {"source": "akshare", "status": "unknown", "tests": {}}
        try:
            df = AkshareCollector.get_kline("000001", "daily", "20260101", "20260514", use_cache=False)
            if df is not None and len(df) > 0:
                result["status"] = "ok"
                result["tests"]["kline"] = f"{len(df)} rows"
                result["last_check"] = str(datetime.now())
            else:
                result["status"] = "degraded"
        except Exception as e:
            result["status"] = "error"
            result["error"] = str(e)[:100]
        return result


collector = AkshareCollector()