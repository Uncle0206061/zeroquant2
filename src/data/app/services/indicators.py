# app/services/indicators.py — 技术指标计算
# MA / MACD / RSI — 纯 Python 实现，基于收盘价序列

import pandas as pd
import numpy as np
from typing import Optional


def calc_ma(closes: pd.Series, period: int = 20) -> pd.Series:
    """简单移动均线"""
    return closes.rolling(window=period).mean()


def calc_ema(closes: pd.Series, period: int = 12) -> pd.Series:
    """指数移动均线"""
    return closes.ewm(span=period, adjust=False).mean()


def calc_macd(
    closes: pd.Series, fast: int = 12, slow: int = 26, signal: int = 9
) -> dict:
    """
    MACD 指标
    返回: {diff, dea, macd, golden_cross, death_cross}
    """
    ema_fast = calc_ema(closes, fast)
    ema_slow = calc_ema(closes, slow)
    diff = ema_fast - ema_slow
    dea = calc_ema(diff.dropna(), signal)
    macd_bar = 2 * (diff - dea)

    # 金叉/死叉
    golden = (diff > dea) & (diff.shift(1) <= dea.shift(1))
    death  = (diff < dea) & (diff.shift(1) >= dea.shift(1))

    return {
        "diff": diff.values.tolist(),
        "dea": dea.values.tolist(),
        "macd": macd_bar.values.tolist(),
        "golden_cross": bool(golden.iloc[-1]) if len(golden) > 0 else False,
        "death_cross": bool(death.iloc[-1]) if len(death) > 0 else False,
    }


def calc_rsi(closes: pd.Series, period: int = 14) -> float:
    """
    RSI 相对强弱指标
    返回最新 RSI 值 (0-100)
    """
    delta = closes.diff()
    gain = delta.where(delta > 0, 0.0)
    loss = -delta.where(delta < 0, 0.0)

    avg_gain = gain.rolling(window=period).mean()
    avg_loss = loss.rolling(window=period).mean()

    rs = avg_gain / avg_loss.replace(0, np.nan)
    rsi = 100 - (100 / (1 + rs))

    last = rsi.dropna().iloc[-1] if len(rsi.dropna()) > 0 else 50.0
    return round(float(last), 2)


def check_ma_cross(
    closes: pd.Series, ma_periods: list[int]
) -> dict:
    """
    检查多周期均线交叉
    返回: {golden_cross, death_cross} 布尔值
    """
    if len(ma_periods) < 2:
        return {"golden_cross": False, "death_cross": False}

    short_p, long_p = min(ma_periods), max(ma_periods)
    ma_short = calc_ma(closes, short_p)
    ma_long  = calc_ma(closes, long_p)

    golden = (ma_short > ma_long) & (ma_short.shift(1) <= ma_long.shift(1))
    death  = (ma_short < ma_long) & (ma_short.shift(1) >= ma_long.shift(1))

    return {
        "golden_cross": bool(golden.iloc[-1]) if len(golden) > 0 else False,
        "death_cross": bool(death.iloc[-1]) if len(death) > 0 else False,
    }