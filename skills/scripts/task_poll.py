#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
ZeroQuant 2.0 任务轮询脚本

功能：
1. 检测 tasks/{终端}/inbox/ 下的新任务卡
2. 执行休息时间规则
3. 输出检测结果

使用：
    $env:TERMINAL="niuniu"  # 或 dacongming / dataofficer
    python task_poll.py
"""

import os
import sys
from datetime import datetime
from pathlib import Path

# 配置
NAS_PATH = r"\\100.65.205.77\homes\zero\ZeroQuant2"
REST_START = 0   # 00:01
REST_END = 12    # 12:01

def is_rest_time():
    """检查是否在休息时间"""
    now = datetime.now()
    hour = now.hour
    minute = now.minute
    
    # 00:01 - 12:01 为休息时间
    if hour == 0 and minute >= 1:
        return True
    if 1 <= hour < 12:
        return True
    if hour == 12 and minute < 1:
        return True
    return False

def get_terminal():
    """获取终端标识"""
    terminal = os.environ.get("TERMINAL", "").lower()
    if terminal not in ["niuniu", "dacongming", "dataofficer"]:
        print("❌ 错误：未设置 TERMINAL 环境变量")
        print("   请执行：$env:TERMINAL=\"niuniu\"  # 或 dacongming / dataofficer")
        sys.exit(1)
    return terminal

def scan_inbox(terminal):
    """扫描收件箱"""
    inbox_path = Path(NAS_PATH) / "tasks" / terminal / "inbox"
    all_inbox_path = Path(NAS_PATH) / "tasks" / "all" / "inbox"
    
    tasks = []
    
    # 扫描个人收件箱
    if inbox_path.exists():
        for f in inbox_path.glob("*.md"):
            if f.name.startswith("task-") or f.name.startswith("REQ-"):
                tasks.append(("personal", f))
    
    # 扫描全体收件箱
    if all_inbox_path.exists():
        for f in all_inbox_path.glob("*.md"):
            if f.name.startswith("task-") or f.name.startswith("REQ-"):
                tasks.append(("all", f))
    
    return tasks

def main():
    print("=" * 50)
    print("ZeroQuant 2.0 任务轮询")
    print("=" * 50)
    
    # 检查休息时间
    if is_rest_time():
        now = datetime.now().strftime("%H:%M:%S")
        print(f"[{now}] 休息时间，跳过轮询")
        return
    
    # 获取终端
    terminal = get_terminal()
    print(f"终端：{terminal}")
    print(f"时间：{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("-" * 50)
    
    # 扫描任务
    tasks = scan_inbox(terminal)
    
    if not tasks:
        print("✅ 无新任务")
    else:
        print(f"📬 发现 {len(tasks)} 个任务：")
        for source, task_file in tasks:
            source_label = "全体" if source == "all" else "个人"
            print(f"   [{source_label}] {task_file.name}")
    
    print("=" * 50)

if __name__ == "__main__":
    main()
