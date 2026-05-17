# app/services/backup.py — 数据自动备份服务

import os, json, hashlib, logging, csv, shutil
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional

from app.config import settings

logger = logging.getLogger(__name__)


class BackupService:
    def __init__(self, backup_root: str = None):
        self.backup_root = backup_root or r"Z:\ZeroQuant2\backup\data"
        self.retention_days = 30

    def backup_all(self, kline_data=None, quote_data=None) -> dict:
        today = datetime.now().strftime("%Y%m%d")
        backup_dir = Path(self.backup_root) / today
        backup_dir.mkdir(parents=True, exist_ok=True)

        manifest = {"backup_time": datetime.now().isoformat(), "date": today, "files": [], "md5_checksums": {}}

        if kline_data:
            fp = backup_dir / "kline.csv"
            self._write_csv(fp, kline_data)
            cs = self._md5(fp)
            manifest["files"].append(fp.name)
            manifest["md5_checksums"]["kline.csv"] = cs
            logger.info(f"[backup] K线: {len(kline_data)}条, MD5={cs}")

        if quote_data:
            fp = backup_dir / "quote.csv"
            self._write_csv(fp, quote_data)
            cs = self._md5(fp)
            manifest["files"].append(fp.name)
            manifest["md5_checksums"]["quote.csv"] = cs
            logger.info(f"[backup] 行情: {len(quote_data)}条, MD5={cs}")

        mf = backup_dir / "backup_manifest.json"
        with open(mf, "w", encoding="utf-8") as f:
            json.dump(manifest, f, ensure_ascii=False, indent=2)

        self._cleanup_old()
        logger.info(f"[backup] 完成: {backup_dir}")
        return manifest

    @staticmethod
    def _write_csv(fp: Path, data: list[dict]):
        if not data:
            return
        with open(fp, "w", newline="", encoding="utf-8-sig") as f:
            w = csv.DictWriter(f, fieldnames=data[0].keys())
            w.writeheader()
            w.writerows(data)

    @staticmethod
    def _md5(fp: Path) -> str:
        h = hashlib.md5()
        with open(fp, "rb") as f:
            for chunk in iter(lambda: f.read(8192), b""):
                h.update(chunk)
        return h.hexdigest()

    def _cleanup_old(self):
        root = Path(self.backup_root)
        if not root.exists():
            return
        cutoff = datetime.now() - timedelta(days=self.retention_days)
        for d in root.iterdir():
            if d.is_dir():
                try:
                    dt = datetime.strptime(d.name, "%Y%m%d")
                    if dt < cutoff:
                        shutil.rmtree(d)
                        logger.info(f"[backup] 清理过期: {d.name}")
                except ValueError:
                    pass

    def verify(self, date_str: str) -> dict:
        backup_dir = Path(self.backup_root) / date_str
        if not backup_dir.exists():
            return {"valid": False, "reason": "备份目录不存在"}
        mf = backup_dir / "backup_manifest.json"
        if not mf.exists():
            return {"valid": False, "reason": "manifest 不存在"}
        with open(mf, "r", encoding="utf-8") as f:
            manifest = json.load(f)
        results = {}
        for fn, expected_md5 in manifest.get("md5_checksums", {}).items():
            fp = backup_dir / fn
            if not fp.exists():
                results[fn] = "missing"
            else:
                actual = self._md5(fp)
                results[fn] = "ok" if actual == expected_md5 else "mismatch"
        all_ok = all(v == "ok" for v in results.values())
        return {"valid": all_ok, "results": results}


backup_service = BackupService()