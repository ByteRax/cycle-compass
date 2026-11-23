# -*- coding: utf-8 -*-
"""轮动系统演示：
A. 用笔记真实信号点验证区间判定
B. 用合成比值路径跑通完整回测（验证三区间、现金池、一次性回补机制）
"""
from datetime import date, timedelta
from rotation import (RotationConfig, zone_of, current_signal, backtest,
                      ratio_of, ZONE_LOW, ZONE_MID, ZONE_HIGH)

cfg = RotationConfig()

print("=" * 78)
print("A. 区间判定验证（阈值 0.16 / 0.28）")
print("=" * 78)
# (数据日期, 笔记披露比值, 笔记披露动作)
CASES = [
    ("2026-09-30", 0.2871, "高区买红利100"),
    ("2026-10-08", 0.2764, "中间区空仓"),
    ("2026-10-13", 0.2661, "中间区空仓"),
    ("2026-10-23", 0.2416, "中间区空仓"),
    ("2026-11-03", 0.2039, "中间区空仓"),
    ("2026-11-20", 0.1848, "中间区空仓"),
]
for ds, r, note in CASES:
    z = zone_of(r, cfg)
    ok = "✓" if (("高区" in note and z == ZONE_HIGH) or ("中间区" in note and z == ZONE_MID)) else "✗"
    print(f"{ds}  比值={r:.4f} → {z}  (笔记:{note}) {ok}")

print()
print("单点信号卡示例（2026-11-20，构造 hist 使分位≈24.4%）：")
hist = [0.10 + i * 0.0002 for i in range(1500)]  # 0.10~0.40 均匀，仅演示
sig = current_signal(date(2026, 11, 20), growth_close=1848.0, dividend_close=10000.0,
                     hist_ratios=hist, cfg=cfg)
print(f"  比值={sig.ratio:.4f}  区间={sig.zone}  动作={sig.action}")
print(f"  回高区需涨={sig.to_high_pct:+.1f}%  距低区需跌={sig.to_low_pct:.1f}%")
print(f"  (笔记同口径：回高区需涨+51.5%、距低区需跌13.4% → 验证比值距离公式)")

print()
print("=" * 78)
print("B. 合成数据回测演示（覆盖 高区→中间区→低区→中间区→高区 全路径）")
print("=" * 78)
# 构造 3 阶段比值路径，红利价格固定 100，创业板价格=比值*100
# 阶段1: 高区0.30 横走(买红利)  阶段2: 跌入中间区0.20(攒现金)
# 阶段3: 跌破低区0.15(回补+买创业板)  阶段4: 回中间区  阶段5: 升高区0.29(回补+买红利)
ratio_path = (
    [0.30] * 20 +                 # 高区
    [0.30 - i * 0.005 for i in range(1, 21)] +   # 平滑跌到 0.20 附近
    [0.20] * 20 +                 # 中间区
    [0.20 - i * 0.005 for i in range(1, 13)] +   # 跌到 0.14 进低区
    [0.15] * 15 +                 # 低区
    [0.15 + i * 0.004 for i in range(1, 16)] +   # 升回 0.21 中间区
    [0.21] * 10 +                 # 中间区
    [0.21 + i * 0.006 for i in range(1, 16)]     # 升到 0.30 高区
)
start = date(2024, 1, 1)
dates = [start + timedelta(days=i) for i in range(len(ratio_path))]
div_price = [100.0] * len(ratio_path)                 # 红利价格固定
grow_price = [r * 100.0 for r in ratio_path]          # 创业板价格随比值
res = backtest(dates, grow_price, div_price, cfg)

print(f"合成样本：{len(dates)} 天，{res.total_invest/cfg.weekly_amount:.0f} 个定投周")
print(f"总投入      = {res.total_invest:,.0f} 元")
print(f"期末总资产  = {res.final_value:,.0f} 元")
print(f"净盈利      = {res.net_profit:+,.0f} 元")
print(f"含息年化XIRR= {res.xirr*100:.2f}%  （注：合成样本仅127天且后期单边上涨，XIRR失真，仅用于验证机制）")
print(f"最大回撤    = {res.max_drawdown*100:.2f}%")
print(f"累计股息    = {res.dividends:,.0f} 元")
print(f"一次性回补次数 = {res.rebuy_count}")
print(f"三区间天数  = 低区{res.zone_days[ZONE_LOW]} / 中间区{res.zone_days[ZONE_MID]} / 高区{res.zone_days[ZONE_HIGH]}")

print("\n关键定投/回补流水（仅打印动作变化的前几条）：")
seen = set()
for d, act, amt in res.weekly_log:
    tag = act.split("→")[-1]
    if tag not in seen or "回补" in act:
        print(f"  {d}  {act}  {amt:,.0f}元")
        seen.add(tag)
