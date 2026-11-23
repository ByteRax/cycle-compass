#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""生成每日调度配置；macOS可显式 --install 安装 launchd，其他系统输出 cron 行。"""
from __future__ import annotations

import argparse
import os
from pathlib import Path
import plistlib
import shlex
import subprocess
import sys

PROJECT = Path(__file__).resolve().parent
LABEL = "com.xcnbeta.indexrotation.daily"


def parse_time(value: str) -> tuple[int, int]:
    try:
        hour_s, minute_s = value.split(":", 1)
        hour, minute = int(hour_s), int(minute_s)
        if not (0 <= hour <= 23 and 0 <= minute <= 59):
            raise ValueError
        return hour, minute
    except ValueError as exc:
        raise argparse.ArgumentTypeError("时间格式必须为 HH:MM，例如 20:30") from exc


def main() -> int:
    parser = argparse.ArgumentParser(description="安装或显示双策略每日调度")
    parser.add_argument("--time", default="20:30", type=parse_time, help="每天运行时间，默认20:30")
    parser.add_argument("--mode", choices=["launchd", "cron"],
                        default="launchd" if sys.platform == "darwin" else "cron")
    parser.add_argument("--install", action="store_true", help="实际安装launchd；不加则只生成预览")
    args = parser.parse_args()
    hour, minute = args.time
    python = Path(sys.executable).resolve()
    daily = PROJECT / "daily.py"
    log = PROJECT / "output" / "scheduler.log"

    if args.mode == "launchd":
        plist = {
            "Label": LABEL,
            "ProgramArguments": [str(python), str(daily), "run"],
            "WorkingDirectory": str(PROJECT),
            "StartCalendarInterval": {"Hour": hour, "Minute": minute},
            "RunAtLoad": False,
            "StandardOutPath": str(log),
            "StandardErrorPath": str(log),
            "EnvironmentVariables": {
                "PATH": "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin",
                "PYTHONUNBUFFERED": "1",
                "TZ": "Asia/Shanghai",
            },
        }
        target = Path.home() / "Library" / "LaunchAgents" / f"{LABEL}.plist"
        print(f"调度：每天 {hour:02d}:{minute:02d}（系统本地时间）")
        print(f"配置：{target}")
        print(f"命令：{python} {daily} run")
        if args.install:
            target.parent.mkdir(parents=True, exist_ok=True)
            log.parent.mkdir(parents=True, exist_ok=True)
            with target.open("wb") as f:
                plistlib.dump(plist, f)
            uid = os.getuid()
            subprocess.run(["launchctl", "bootout", f"gui/{uid}", str(target)],
                           stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=False)
            subprocess.run(["launchctl", "bootstrap", f"gui/{uid}", str(target)], check=True)
            print("已安装并加载。验证：launchctl list | grep indexrotation")
        else:
            print("这是预览，尚未安装。确认后执行：")
            print(f"  {shlex.quote(str(python))} {shlex.quote(str(Path(__file__).resolve()))} --time {hour:02d}:{minute:02d} --mode launchd --install")
    else:
        command = f"cd {shlex.quote(str(PROJECT))} && {shlex.quote(str(python))} {shlex.quote(str(daily))} run >> {shlex.quote(str(log))} 2>&1"
        print("将下面一行加入 crontab -e：")
        print(f"{minute} {hour} * * * {command}")
        print("提示：请把服务器时区设为 Asia/Shanghai；节假日重复运行不会重复记账。")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
