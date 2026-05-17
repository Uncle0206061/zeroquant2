# app/services/filter_engine.py — 多因子筛选引擎核心
# AND/OR 逻辑 + 8 种因子筛选

import time
import logging
from typing import Optional
import pandas as pd
import numpy as np

from app.collectors.akshare_collector import collector as akc
from app.services.indicators import calc_rsi, calc_macd, check_ma_cross
from app.services.factor_registry import get_factor

logger = logging.getLogger(__name__)


class FilterEngine:
    """多因子筛选引擎"""

    # ── 主入口 ──

    def execute(
        self,
        rules: list[dict],
        logic: str = "AND",
        stocks: Optional[list[str]] = None,
        timeout: float = 4.5,
    ) -> dict:
        """
        执行多因子筛选
        rules: [{type, factor, ...params}]
        logic: AND 全部满足 / OR 任一满足
        stocks: 指定股票列表, None=全市场
        timeout: 超时秒数
        """
        start = time.time()

        # 1. 获取行情数据
        spot_df = akc.get_spot_all()
        if spot_df is None:
            return {"code": 500, "message": "行情数据不可用", "data": None}

        total_scanned = len(spot_df)

        # 2. 按指定股票过滤
        if stocks:
            code_col = self._find_code_col(spot_df)
            if code_col:
                spot_df = spot_df[spot_df[code_col].astype(str).isin(stocks)]
            if spot_df.empty:
                return {"code": 0, "data": {"matched": [], "total_scanned": total_scanned,
                        "match_rate": 0, "filters_applied": len(rules)}}

        # 3. 逐规则筛选
        result_df = spot_df.copy()
        applied = 0

        for rule in rules:
            if time.time() - start > timeout:
                logger.warning(f"[筛选] 超时 {timeout}s")
                break

            factor_name = rule.get("factor", "")
            factor_def = get_factor(factor_name)
            if factor_def is None:
                continue

            mask = self._apply_rule(result_df, rule, factor_def)
            if mask is None:
                continue

            if logic.upper() == "AND":
                result_df = result_df[mask]
            else:  # OR
                # OR 模式：收集所有匹配结果
                pass  # 在下面处理

            applied += 1

        # OR 逻辑：收集所有命中
        if logic.upper() == "OR" and applied > 0:
            all_matched = set()
            code_col = self._find_code_col(spot_df)
            for rule in rules:
                factor_def = get_factor(rule.get("factor", ""))
                if factor_def is None:
                    continue
                mask = self._apply_rule(spot_df, rule, factor_def)
                if mask is not None and code_col:
                    matched_codes = spot_df.loc[mask, code_col].astype(str).tolist()
                    all_matched.update(matched_codes)
            code_col = self._find_code_col(result_df)
            if code_col:
                result_df = result_df[result_df[code_col].astype(str).isin(all_matched)]

        # 4. 组装结果
        code_col = self._find_code_col(result_df)
        matched = result_df[code_col].astype(str).tolist() if code_col else []
        elapsed = round((time.time() - start) * 1000, 1)

        return {
            "code": 0,
            "data": {
                "matched": matched[:200],
                "total_scanned": total_scanned,
                "matched_count": len(matched),
                "match_rate": round(len(matched) / total_scanned, 4) if total_scanned > 0 else 0,
                "filters_applied": applied,
                "elapsed_ms": elapsed,
            },
        }

    # ── 规则适配器 ──

    def _apply_rule(self, df: pd.DataFrame, rule: dict, factor_def) -> Optional[pd.Series]:
        """对 DataFrame 应用单条筛选规则，返回布尔 mask"""
        name = factor_def.name
        category = factor_def.category

        if category == "financial":
            return self._filter_financial(df, name, rule)
        elif category == "sector":
            return self._filter_sector(df, name, rule)
        elif category == "technical":
            return self._filter_technical(df, name, rule)
        return None

    # ── 财务指标筛选 ──

    def _filter_financial(self, df: pd.DataFrame, factor: str, rule: dict) -> Optional[pd.Series]:
        """PE/PB/市值 范围筛选"""
        min_val = rule.get("min", 0)
        max_val = rule.get("max", 999999)

        col_map = {"pe": "市盈率-动态", "pb": "市净率", "market_cap": "总市值"}
        col = col_map.get(factor, "")
        if not col:
            return None

        # 列名模糊匹配
        matched_col = next((c for c in df.columns if col in str(c)), None)
        if matched_col is None:
            logger.warning(f"[筛选] 列 {col} 不存在")
            return None

        series = pd.to_numeric(df[matched_col], errors="coerce")
        return (series >= min_val) & (series <= max_val)

    # ── 板块筛选 ──

    def _filter_sector(self, df: pd.DataFrame, factor: str, rule: dict) -> Optional[pd.Series]:
        """板块/概念 包含筛选"""
        target_sectors = rule.get("value", [])
        if not target_sectors:
            return pd.Series([True] * len(df), index=df.index)

        # 获取板块成分股
        sector_stocks = self._get_sector_stocks(factor, target_sectors)
        if not sector_stocks:
            return pd.Series([False] * len(df), index=df.index)

        code_col = self._find_code_col(df)
        if code_col is None:
            return None

        return df[code_col].astype(str).isin(sector_stocks)

    def _get_sector_stocks(self, factor: str, sector_names: list[str]) -> set[str]:
        """获取板块包含的股票代码集合"""
        stocks = set()
        try:
            if factor == "sector":
                df = akc.get_sector_industry()
            else:
                df = akc.get_sector_concept()

            if df is None:
                return stocks

            # 查找板块名称列和代码列
            cols = df.columns.tolist()
            name_col = next((c for c in cols if '板块' in c or '名称' in c), None)

            if name_col:
                matched = df[df[name_col].astype(str).isin(sector_names)]
                # 如果接口直接返回代码列
                code_col = next((c for c in cols if '代码' in c), None)
                if code_col:
                    stocks.update(matched[code_col].astype(str).tolist())
        except Exception as e:
            logger.warning(f"[筛选] 板块数据获取失败: {e}")
        return stocks

    # ── 技术指标筛选 ──

    def _filter_technical(self, df: pd.DataFrame, factor: str, rule: dict) -> Optional[pd.Series]:
        """MA/MACD/RSI 技术指标筛选"""
        result = []
        code_col = self._find_code_col(df)
        if code_col is None:
            return None

        codes = df[code_col].astype(str).tolist()

        for code in codes:
            kline = akc.get_kline(code, "daily", "20250901", "20260514")
            if kline is None or len(kline) < 30:
                result.append(False)
                continue

            close_col = next((c for c in kline.columns if '收盘' in str(c)), None)
            if close_col is None:
                result.append(False)
                continue

            closes = pd.to_numeric(kline[close_col], errors="coerce").dropna()

            if factor == "rsi":
                period = rule.get("period", 14)
                threshold = rule.get("threshold", 30)
                rsi_val = calc_rsi(closes, period)
                # RSI < threshold = 超卖信号
                result.append(rsi_val < threshold)

            elif factor == "macd":
                fast, slow, signal = rule.get("fast", 12), rule.get("slow", 26), rule.get("signal", 9)
                macd_result = calc_macd(closes, fast, slow, signal)
                result.append(macd_result["golden_cross"])

            elif factor == "ma":
                period = rule.get("period", 20)
                cross_type = rule.get("cross_type", "golden")
                cross = check_ma_cross(closes, [5, period])
                result.append(cross.get(f"{cross_type}_cross", False))

            else:
                result.append(False)

        return pd.Series(result, index=df.index)

    # ── 辅助 ──

    @staticmethod
    def _find_code_col(df: pd.DataFrame) -> Optional[str]:
        for kw in ["代码", "code", "symbol"]:
            for c in df.columns:
                if kw in str(c).lower():
                    return c
        return None


# 单例
engine = FilterEngine()