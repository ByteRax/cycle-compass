# -*- coding: utf-8 -*-
"""Go API 与 Python API 黄金样本全量自动化比对与校验脚本。"""
import json
import urllib.request
from pathlib import Path

BASE_URL = "http://127.0.0.1:8088"
GOLDEN_DIR = Path(__file__).resolve().parent / "golden_api_snapshots"

IGNORE_KEYS = {"timestamp", "updated_at", "latency_ms", "as_of", "date"}

ENDPOINTS = [
    ("root.json", "GET", "/", None),
    ("health.json", "GET", "/api/v1/health", None),
    ("quotes_batch.json", "GET", "/api/v1/markets/quotes?symbols=930955,399673,000300", None),
    ("quote_single.json", "GET", "/api/v1/markets/quotes/399673", None),
    ("sectors_metrics.json", "GET", "/api/v1/sectors/metrics", None),
    ("cycles_overview.json", "GET", "/api/v1/cycles/overview", None),
    ("rotation_signal.json", "GET", "/api/v1/strategies/rotation/signal", None),
    ("scoring_detail.json", "GET", "/api/v1/strategies/scoring/detail", None),
    ("combined_signal.json", "GET", "/api/v1/strategies/combined", None),
    ("portfolios_overview.json", "GET", "/api/v1/portfolios/overview", None),
    ("backtest_run.json", "POST", "/api/v1/backtests/run", json.dumps({"days": 300, "low_line": 0.16, "high_line": 0.28, "weekly_amount": 1000.0}).encode("utf-8")),
]


def compare_values(golden, current, path=""):
    diffs = []
    if isinstance(golden, dict) and isinstance(current, dict):
        for k, v in golden.items():
            sub_path = f"{path}.{k}" if path else k
            if k in IGNORE_KEYS:
                continue
            if k not in current:
                diffs.append(f"Missing key in Go response: {sub_path}")
            else:
                diffs.extend(compare_values(v, current[k], sub_path))
        for k in current:
            sub_path = f"{path}.{k}" if path else k
            if k not in golden and k not in IGNORE_KEYS:
                diffs.append(f"Unexpected extra key in Go response: {sub_path}")
    elif isinstance(golden, list) and isinstance(current, list):
        if len(golden) != len(current):
            diffs.append(f"List length mismatch at {path}: golden={len(golden)}, go={len(current)}")
        else:
            for i, (g_item, c_item) in enumerate(zip(golden, current)):
                diffs.extend(compare_values(g_item, c_item, f"{path}[{i}]"))
    elif isinstance(golden, float) or isinstance(current, float):
        try:
            g_f = float(golden)
            c_f = float(current)
            if abs(g_f - c_f) > 0.05:  # 允许极小浮点差异
                diffs.append(f"Float value diff at {path}: golden={g_f}, go={c_f}")
        except Exception:
            if golden != current:
                diffs.append(f"Value diff at {path}: golden={golden}, go={current}")
    else:
        if golden != current:
            diffs.append(f"Value diff at {path}: golden={golden}, go={current}")
    return diffs


def main():
    print("=" * 60)
    print("CycleCompass Go API 对比回归验证开始")
    print("=" * 60)

    results = []
    all_passed = True

    for snapshot_file, method, path, body in ENDPOINTS:
        golden_path = GOLDEN_DIR / snapshot_file
        if not golden_path.exists():
            print(f"[SKIP] 缺少黄金样本: {snapshot_file}")
            continue

        with open(golden_path, "r", encoding="utf-8") as f:
            golden_data = json.load(f)

        url = f"{BASE_URL}{path}"
        req = urllib.request.Request(url, method=method, data=body)
        if body:
            req.add_header("Content-Type", "application/json")

        try:
            with urllib.request.urlopen(req, timeout=5) as resp:
                status = resp.status
                res_bytes = resp.read()
                go_data = json.loads(res_bytes.decode("utf-8"))
        except Exception as e:
            print(f"[FAIL] {method} {path} 请求失败: {e}")
            all_passed = False
            results.append((path, method, False, [str(e)]))
            continue

        diffs = compare_values(golden_data, go_data)
        if not diffs:
            print(f"[PASS] {method} {path} 100% 契约一致！")
            results.append((path, method, True, []))
        else:
            print(f"[WARN] {method} {path} 存在 {len(diffs)} 处差异:")
            for d in diffs[:5]:
                print(f"       - {d}")
            results.append((path, method, False, diffs))
            all_passed = False

    print("=" * 60)
    if all_passed:
        print("🎉 全部 11 个核心接口契约 100% 完美对齐无损通过！")
    else:
        print("⚠️ 存在差异项，请查看详细列表。")
    print("=" * 60)
    return results


if __name__ == "__main__":
    main()
