# app/models/models.py — 数据表模型（data_ 前缀）

from datetime import datetime
from sqlalchemy import Column, Integer, String, Float, DateTime, Date, Index, Text
from app.models.base import Base


class Kline(Base):
    """K线数据 — data_kline"""
    __tablename__ = "data_kline"

    id = Column(Integer, primary_key=True, autoincrement=True)
    symbol = Column(String(10), nullable=False, index=True, comment="股票代码")
    period = Column(String(10), nullable=False, comment="周期: daily/weekly/monthly")
    trade_date = Column(Date, nullable=False, comment="交易日")
    open = Column(Float, comment="开盘价")
    high = Column(Float, comment="最高价")
    low = Column(Float, comment="最低价")
    close = Column(Float, comment="收盘价")
    volume = Column(Float, comment="成交量(手)")
    amount = Column(Float, comment="成交额(元)")
    amplitude = Column(Float, comment="振幅")
    pct_change = Column(Float, comment="涨跌幅")
    turnover = Column(Float, comment="换手率")
    created_at = Column(DateTime, default=datetime.utcnow)

    __table_args__ = (
        Index("ix_data_kline_symbol_period", "symbol", "period"),
        Index("ix_data_kline_symbol_date", "symbol", "trade_date"),
        {"comment": "K线数据表 — 永久存储"},
    )


class Quote(Base):
    """实时行情快照 — data_quote"""
    __tablename__ = "data_quote"

    id = Column(Integer, primary_key=True, autoincrement=True)
    symbol = Column(String(10), nullable=False, index=True, comment="股票代码")
    name = Column(String(20), comment="股票名称")
    price = Column(Float, comment="最新价")
    change = Column(Float, comment="涨跌额")
    pct_change = Column(Float, comment="涨跌幅")
    volume = Column(Float, comment="成交量(手)")
    amount = Column(Float, comment="成交额(元)")
    high = Column(Float)
    low = Column(Float)
    open = Column(Float)
    pre_close = Column(Float)
    pe = Column(Float, comment="市盈率")
    pb = Column(Float, comment="市净率")
    market_cap = Column(Float, comment="总市值")
    snapshot_time = Column(DateTime, default=datetime.utcnow, comment="快照时间")

    __table_args__ = (
        Index("ix_data_quote_symbol_time", "symbol", "snapshot_time"),
        {"comment": "实时行情快照 — 保留1年"},
    )


class Tick(Base):
    """逐笔成交 — data_tick"""
    __tablename__ = "data_tick"

    id = Column(Integer, primary_key=True, autoincrement=True)
    symbol = Column(String(10), nullable=False, index=True, comment="股票代码")
    trade_time = Column(DateTime, nullable=False, comment="成交时间")
    price = Column(Float, nullable=False, comment="成交价")
    volume = Column(Float, comment="成交量(手)")
    amount = Column(Float, comment="成交额(元)")
    direction = Column(String(2), comment="买卖方向: B/S")
    created_at = Column(DateTime, default=datetime.utcnow)

    __table_args__ = (
        Index("ix_data_tick_symbol_time", "symbol", "trade_time"),
        {"comment": "逐笔成交 — 保留90天"},
    )


class Sector(Base):
    """板块分类 — data_sector"""
    __tablename__ = "data_sector"

    id = Column(Integer, primary_key=True, autoincrement=True)
    sector_name = Column(String(50), nullable=False, index=True, comment="板块名称")
    sector_type = Column(String(20), comment="类型: industry/concept")
    stock_code = Column(String(10), nullable=False, index=True, comment="成分股代码")
    stock_name = Column(String(20), comment="成分股名称")

    __table_args__ = (
        Index("ix_data_sector_type_name", "sector_type", "sector_name"),
        {"comment": "板块成分股"},
    )