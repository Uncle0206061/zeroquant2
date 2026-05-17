from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.api.v1 import health, market, kline, orderbook, sector, filter_routes, monitor, admin
from app.config import settings

@asynccontextmanager
async def lifespan(app: FastAPI):
    print(f"[数据服务] {settings.APP_NAME} v{settings.APP_VERSION} :{settings.PORT}")
    import asyncio
    from app.services.simulated_level2 import simulated_level2
    asyncio.create_task(simulated_level2.start(["000001", "600519", "300750"]))
    print("[数据服务] 模拟Level-2已启动")
    yield
    await simulated_level2.stop()

app = FastAPI(title=settings.APP_NAME, version=settings.APP_VERSION, docs_url="/data/docs", redoc_url="/data/redoc", lifespan=lifespan)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://localhost:8080"],
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["Content-Type", "Authorization"],
)
app.include_router(health.router, prefix="/data/v1", tags=["health"])
app.include_router(market.router, prefix="/data/v1", tags=["market"])
app.include_router(kline.router, prefix="/data/v1", tags=["kline"])
app.include_router(orderbook.router, prefix="/data/v1", tags=["orderbook"])
app.include_router(sector.router, prefix="/data/v1", tags=["sector"])
app.include_router(filter_routes.router, prefix="/data/v1", tags=["filter"])
app.include_router(monitor.router, prefix="/data/v1", tags=["monitor"])
app.include_router(admin.router, prefix="/data/v1", tags=["admin"])

@app.get("/")
async def root():
    return {"service": settings.APP_NAME, "version": settings.APP_VERSION, "docs": "/data/docs"}