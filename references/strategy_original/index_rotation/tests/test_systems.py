# -*- coding: utf-8 -*-
import json
import sys
import unittest
from datetime import date, timedelta
from pathlib import Path

import pandas as pd

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from daily import calculate
from data_provider import MarketData
from execution import DEFAULT_STATE, apply_rotation, apply_scoring
from indicators import rsi_wilder
from rotation import RotationConfig, ZONE_HIGH, ZONE_MID, backtest, current_signal
from scoring import ScoreInput, evaluate


class ScoringTests(unittest.TestCase):
    def test_note_calibration(self):
        cases = [
            (ScoreInput(0.846, 4.41, 1.67, 50, 0.64, 0.45), 9.20),
            (ScoreInput(0.870, 4.58, 1.62, 53, 0.99, 0.62), 7.00),
            (ScoreInput(0.900, 4.49, 1.58, 54, 0.96, 0.75), 6.72),
        ]
        for inp, expected in cases:
            self.assertAlmostEqual(evaluate(inp).total, expected, delta=0.03)

    def test_reduction_is_not_repeated_every_day(self):
        state = json.loads(json.dumps(DEFAULT_STATE))
        first = apply_scoring(state, date(2026, 11, 18), 5.5)
        same = apply_scoring(state, date(2026, 11, 18), 5.5)
        next_day = apply_scoring(state, date(2026, 11, 19), 5.4)
        self.assertEqual(first["type"], "reduce")
        self.assertEqual(first["reduce_ratio"], 0.25)
        self.assertEqual(same["type"], "none")
        self.assertEqual(next_day["type"], "none")

    def test_sell_pool_buyback_only_once(self):
        state = json.loads(json.dumps(DEFAULT_STATE))
        apply_scoring(state, date(2026, 11, 18), 5.5)
        buyback = apply_scoring(state, date(2026, 11, 19), 12.2)
        repeated = apply_scoring(state, date(2026, 11, 20), 12.3)
        self.assertEqual(buyback["type"], "sell_pool_buyback")
        self.assertEqual(repeated["type"], "none")


class RotationTests(unittest.TestCase):
    def test_distance_matches_note(self):
        sig = current_signal(date(2026, 11, 20), 1848, 10000, [0.1, 0.2, 0.3])
        self.assertEqual(sig.zone, ZONE_MID)
        self.assertAlmostEqual(sig.to_high_pct, 51.5, delta=0.1)
        self.assertAlmostEqual(sig.to_low_pct, 13.4, delta=0.1)

    def test_rebuy_waits_until_next_week(self):
        start = date(2026, 11, 2)  # 周一
        dates = [d.date() for d in pd.bdate_range(start, periods=15)]
        ratios = [0.20] * 5 + [0.30] * 10
        dividend = [100.0] * len(dates)
        growth = [r * 100 for r in ratios]
        res = backtest(dates, growth, dividend, RotationConfig(div_yield_dividend=0, div_yield_growth=0))
        rebuys = [row for row in res.weekly_log if "回补" in row[1]]
        self.assertEqual(len(rebuys), 1)
        self.assertEqual(rebuys[0][0], dates[10])

    def test_rotation_state_idempotent(self):
        state = json.loads(json.dumps(DEFAULT_STATE))
        first = apply_rotation(state, date(2026, 11, 2), ZONE_MID, True, 1000)
        repeated = apply_rotation(state, date(2026, 11, 2), ZONE_MID, True, 1000)
        self.assertEqual(first["cash_pool"], 1000)
        self.assertEqual(repeated["cash_pool"], 1000)


class IndicatorAndPipelineTests(unittest.TestCase):
    def test_wilder_rsi_extremes(self):
        up = pd.Series(range(1, 102), dtype=float)
        flat = pd.Series([100.0] * 101)
        self.assertEqual(float(rsi_wilder(up, 70).iloc[-1]), 100.0)
        self.assertEqual(float(rsi_wilder(flat, 70).iloc[-1]), 50.0)

    def test_synthetic_end_to_end(self):
        dates = pd.bdate_range(end="2026-11-20", periods=300)
        red = pd.DataFrame({
            "date": dates,
            "close": [11000 + i + 30 * ((i % 7) - 3) for i in range(300)],
            "pe_ttm": [8 + i * 0.003 for i in range(300)],
        })
        growth = pd.DataFrame({"date": dates, "close": red["close"] * 0.20})
        valuation = pd.DataFrame({"date": [dates[-1]], "pe1": [9.0], "pe2": [9.5],
                                  "dividend_yield_1": [4.5], "dividend_yield_2": [4.8]})
        bond = pd.DataFrame({"date": [dates[-1]], "bond10y": [1.6]})
        market = MarketData(red, growth, valuation, bond, {"test": "synthetic"}, [])
        config = json.loads((ROOT / "config.json").read_text(encoding="utf-8"))
        result, report = calculate(config, market, explicit_as_of=True)
        self.assertIn("rotation", result)
        self.assertIn("总分", report)
        self.assertGreaterEqual(result["scoring"]["total"], 0)
        self.assertLessEqual(result["scoring"]["total"], 20)


if __name__ == "__main__":
    unittest.main()
