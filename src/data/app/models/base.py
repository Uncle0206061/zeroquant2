# app/models/base.py — SQLAlchemy 基础模型（延迟连接）

from sqlalchemy import create_engine
from sqlalchemy.orm import declarative_base, sessionmaker
from app.config import settings

Base = declarative_base()

_engine = None
_SessionLocal = None


def get_engine():
    global _engine
    if _engine is None:
        _engine = create_engine(
            settings.DATABASE_URL,
            pool_size=10,
            max_overflow=20,
            pool_pre_ping=True,
            echo=settings.DEBUG,
            connect_args={"connect_timeout": 2} if "postgresql" in settings.DATABASE_URL else {},
        )
    return _engine


def get_session():
    global _SessionLocal
    if _SessionLocal is None:
        _SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=get_engine())
    return _SessionLocal()


def get_db():
    db = get_session()
    try:
        yield db
    finally:
        db.close()


# 向后兼容别名（延迟获取）
@property
def engine():
    return get_engine()