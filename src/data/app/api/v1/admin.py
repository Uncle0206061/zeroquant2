# app/api/v1/admin.py — 运维端点（备份/恢复）

import asyncio
from fastapi import APIRouter
from app.services.backup import backup_service
from app.services.recovery import recovery_service

router = APIRouter()


@router.post("/admin/backup")
async def trigger_backup():
    """触发手动备份"""
    from app.collectors.akshare_collector import collector as akc

    # 获取 K线样本数据
    kline_data = []
    quote_data = []
    try:
        for symbol in ["000001", "600519"]:
            df = akc.get_kline(symbol, "daily", "20260501", "20260516", use_cache=False)
            if df is not None:
                kline_data.extend(df.to_dict("records"))

        df_spot = akc.get_spot_all(use_cache=False)
        if df_spot is not None:
            quote_data = df_spot.head(100).to_dict("records")
    except Exception as e:
        return {"code": 500, "message": f"数据获取失败: {e}", "data": None}

    manifest = backup_service.backup_all(kline_data, quote_data)
    return {"code": 0, "message": "success", "data": manifest}


@router.get("/admin/backup/verify/{date}")
async def verify_backup(date: str):
    """验证指定日期备份完整性"""
    result = backup_service.verify(date)
    return {"code": 0, "message": "success", "data": result}


@router.get("/admin/backup/list")
async def list_backups():
    """列出可用备份"""
    from pathlib import Path
    root = Path(backup_service.backup_root)
    if not root.exists():
        return {"code": 0, "message": "success", "data": {"backups": [], "message": "备份目录不存在"}}

    backups = []
    for d in sorted(root.iterdir(), reverse=True):
        if d.is_dir():
            backups.append({
                "date": d.name,
                "has_manifest": (d / "backup_manifest.json").exists(),
            })
    return {"code": 0, "message": "success", "data": {"backups": backups[:30]}}


@router.post("/admin/recover")
async def trigger_recovery(symbols: list[str] = None):
    """触发数据恢复"""
    if symbols is None:
        symbols = ["000001", "600519", "300750"]

    loop = asyncio.get_event_loop()
    result = await loop.run_in_executor(
        None, recovery_service.auto_recover, symbols, "daily", 365
    )
    return {"code": 0, "message": "success", "data": result}


@router.get("/admin/recovery/log")
async def recovery_log(limit: int = 50):
    """恢复日志"""
    return {"code": 0, "message": "success", "data": {"log": recovery_service.recovery_log[-limit:]}}