package trend

import (
	"math"
	"testing"
	"time"
)

func makeConstBars(assetID string, start string, count int, val float64) []DailyBar {
	t, _ := time.Parse("2006-01-02", start)
	bars := make([]DailyBar, count)
	for i := 0; i < count; i++ {
		v := val
		bars[i] = DailyBar{
			AssetID:     assetID,
			TradingDate: t.AddDate(0, 0, i).Format("2006-01-02"),
			Value:       &v,
		}
	}
	return bars
}

func makeTrendBars(assetID string, start string, vals []float64) []DailyBar {
	t, _ := time.Parse("2006-01-02", start)
	bars := make([]DailyBar, len(vals))
	for i, val := range vals {
		v := val
		bars[i] = DailyBar{
			AssetID:     assetID,
			TradingDate: t.AddDate(0, 0, i).Format("2006-01-02"),
			Value:       &v,
		}
	}
	return bars
}

// A-01: Temperature boundaries, null, NaN, insufficient history, coverage, zero-vol.
func TestTemperatureBoundariesAndValidation(t *testing.T) {
	cfg := DefaultAlgorithmConfig()

	t.Run("classifyTempLevel boundaries", func(t *testing.T) {
		cases := []struct {
			score float64
			want  TempLevel
		}{
			{0, TempFreezing},
			{14.999, TempFreezing},
			{15.0, TempCold},
			{28.999, TempCold},
			{29.0, TempCool},
			{42.999, TempCool},
			{43.0, TempFlat},
			{57.999, TempFlat},
			{58.0, TempWarm},
			{71.999, TempWarm},
			{72.0, TempHot},
			{85.999, TempHot},
			{86.0, TempBoiling},
			{100.0, TempBoiling},
		}
		for _, tc := range cases {
			got := classifyTempLevel(tc.score)
			if got != tc.want {
				t.Errorf("classifyTempLevel(%v) = %s, want %s", tc.score, got, tc.want)
			}
		}
	})

	t.Run("zero volatility constant series", func(t *testing.T) {
		bars := makeConstBars("A", "2025-01-01", 121, 100.0)
		res := ComputeTemperature(bars, cfg)
		if res.Level != TempFlat || res.Score == nil || math.Abs(*res.Score-50.0) > 1e-9 {
			t.Fatalf("expected flat 50.0, got level=%s score=%v reason=%s", res.Level, res.Score, res.Reason)
		}
	})

	t.Run("insufficient history", func(t *testing.T) {
		bars := makeConstBars("A", "2025-01-01", 120, 100.0)
		res := ComputeTemperature(bars, cfg)
		if res.Level != TempUnavailable || res.Score != nil || res.Reason == "" {
			t.Fatalf("expected unavailable for 120 bars, got %+v", res)
		}
	})

	t.Run("insufficient coverage < 90%", func(t *testing.T) {
		bars := makeConstBars("A", "2025-01-01", 140, 100.0)
		// Null out 15 bars in the last 121 -> 106/121 = 87.6% < 90%
		for i := 20; i < 35; i++ {
			bars[i].Value = nil
		}
		res := ComputeTemperature(bars, cfg)
		if res.Level != TempUnavailable || res.Reason == "" {
			t.Fatalf("expected unavailable for <90%% coverage, got %+v", res)
		}
	})

	t.Run("NaN and non-positive values", func(t *testing.T) {
		bars := makeConstBars("A", "2025-01-01", 121, 100.0)
		nanVal := math.NaN()
		bars[60].Value = &nanVal
		res := ComputeTemperature(bars, cfg)
		if res.Level != TempUnavailable || res.Reason == "" {
			t.Fatalf("expected unavailable for NaN, got %+v", res)
		}

		zeroVal := 0.0
		bars[60].Value = &zeroVal
		res2 := ComputeTemperature(bars, cfg)
		if res2.Level != TempUnavailable || res2.Reason == "" {
			t.Fatalf("expected unavailable for 0, got %+v", res2)
		}
	})

	t.Run("latest bar null", func(t *testing.T) {
		bars := makeConstBars("A", "2025-01-01", 125, 100.0)
		bars[124].Value = nil
		res := ComputeTemperature(bars, cfg)
		if res.Level != TempUnavailable {
			t.Fatalf("expected unavailable when latest bar is nil, got %+v", res)
		}
	})
}

// A-02 & A-08: State machine 3-day confirm, 2-day exit, gap handling, same-day exit+reverse, idempotent events.
func TestStateMachineAndEvents(t *testing.T) {
	cfg := DefaultAlgorithmConfig()

	// Build 125 flat bars, then strong uptrend, then gap, then sharp reversal.
	vals := make([]float64, 0, 160)
	for i := 0; i < 125; i++ {
		// slight oscillation so sigma60 > 0
		vals = append(vals, 100.0+0.1*float64(i%2))
	}
	// Uptrend for 10 days
	cur := 100.0
	for i := 0; i < 10; i++ {
		cur *= 1.02
		vals = append(vals, cur)
	}
	bars := makeTrendBars("ASSET1", "2025-01-01", vals)

	// Insert a missing/holiday bar after confirmation (e.g. index 132)
	bars[132].Value = nil

	// Append sharp downtrend for 8 days to trigger 2-day UP exit and reverse DOWN forming
	lastVal := *bars[len(bars)-1].Value
	lastDate, _ := time.Parse("2006-01-02", bars[len(bars)-1].TradingDate)
	for i := 1; i <= 8; i++ {
		lastVal *= 0.93
		v := lastVal
		bars = append(bars, DailyBar{
			AssetID:     "ASSET1",
			TradingDate: lastDate.AddDate(0, 0, i).Format("2006-01-02"),
			Value:       &v,
		})
	}

	states, events := ComputeStateSeries(bars, cfg)
	if len(states) != len(bars) {
		t.Fatalf("expected %d states, got %d", len(bars), len(states))
	}

	// Verify gap day is stale/unknown and did not increment duration
	if !states[132].Stale || states[132].Direction != DirectionUnknown {
		t.Errorf("expected gap day 132 to be stale/unknown, got %+v", states[132])
	}

	// Verify right_started, right_ended, down_started events
	var hasRightStart, hasRightEnd, hasDownStart bool
	seenKeys := make(map[string]bool)
	for _, ev := range events {
		if seenKeys[ev.DedupeKey] {
			t.Errorf("duplicate dedupe key: %s", ev.DedupeKey)
		}
		seenKeys[ev.DedupeKey] = true
		switch ev.EventType {
		case "right_started":
			hasRightStart = true
		case "right_ended":
			hasRightEnd = true
		case "down_started":
			hasDownStart = true
		}
	}
	if !hasRightStart || !hasRightEnd || !hasDownStart {
		t.Errorf("expected right_started, right_ended, down_started events; got %+v", events)
	}

	// Idempotency check
	states2, events2 := ComputeStateSeries(bars, cfg)
	if len(events) != len(events2) {
		t.Fatalf("non-idempotent event count: %d vs %d", len(events), len(events2))
	}
	for i := range events {
		if events[i].DedupeKey != events2[i].DedupeKey {
			t.Errorf("non-idempotent event at %d", i)
		}
	}
	_ = states2
}

// A-05: Phase mapping independent of calendar date.
func TestComputePhase(t *testing.T) {
	cases := []struct {
		name     string
		state    StateResult
		wantCode string
		wantName string
	}{
		{"unknown", StateResult{Direction: DirectionUnknown, Confirmation: ConfirmationUnknown}, "unknown", "未知"},
		{"neutral", StateResult{Direction: DirectionNeutral, Confirmation: ConfirmationNone}, "oscillation", "震荡"},
		{"up forming", StateResult{Direction: DirectionUp, Confirmation: ConfirmationForming}, "guyu", "谷雨"},
		{"up 1", StateResult{Direction: DirectionUp, Confirmation: ConfirmationConfirmed, DurationTradingDays: 1}, "lixia", "立夏"},
		{"up 10", StateResult{Direction: DirectionUp, Confirmation: ConfirmationConfirmed, DurationTradingDays: 10}, "lixia", "立夏"},
		{"up 11", StateResult{Direction: DirectionUp, Confirmation: ConfirmationConfirmed, DurationTradingDays: 11}, "xiazhi", "夏至"},
		{"up 40", StateResult{Direction: DirectionUp, Confirmation: ConfirmationConfirmed, DurationTradingDays: 40}, "xiazhi", "夏至"},
		{"up 41", StateResult{Direction: DirectionUp, Confirmation: ConfirmationConfirmed, DurationTradingDays: 41}, "xiaoshu", "小暑"},
		{"up 90", StateResult{Direction: DirectionUp, Confirmation: ConfirmationConfirmed, DurationTradingDays: 90}, "xiaoshu", "小暑"},
		{"up 91", StateResult{Direction: DirectionUp, Confirmation: ConfirmationConfirmed, DurationTradingDays: 91}, "dashu", "大暑"},
		{"down forming", StateResult{Direction: DirectionDown, Confirmation: ConfirmationForming}, "hanlu", "寒露"},
		{"down 5", StateResult{Direction: DirectionDown, Confirmation: ConfirmationConfirmed, DurationTradingDays: 5}, "lidong", "立冬"},
		{"down 20", StateResult{Direction: DirectionDown, Confirmation: ConfirmationConfirmed, DurationTradingDays: 20}, "dongzhi", "冬至"},
		{"down 50", StateResult{Direction: DirectionDown, Confirmation: ConfirmationConfirmed, DurationTradingDays: 50}, "xiaohan", "小寒"},
		{"down 100", StateResult{Direction: DirectionDown, Confirmation: ConfirmationConfirmed, DurationTradingDays: 100}, "dahan", "大寒"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputePhase(tc.state)
			if got.PhaseCode != tc.wantCode || got.PhaseName != tc.wantName {
				t.Errorf("got (%s, %s), want (%s, %s)", got.PhaseCode, got.PhaseName, tc.wantCode, tc.wantName)
			}
		})
	}
}

// A-03: Relative strength ranking, N=0/1/2, ties, 0/100, insufficient history.
func TestComputeRelativeStrength(t *testing.T) {
	// N < 2 -> nil
	if got := ComputeRelativeStrength(nil, "s", "S", "USD", "trend-v1.0.0"); got != nil {
		t.Errorf("expected nil for N=0")
	}
	one := []StrengthInput{{AssetID: "A", Bars: makeConstBars("A", "2025-01-01", 61, 100)}}
	if got := ComputeRelativeStrength(one, "s", "S", "USD", "trend-v1.0.0"); got != nil {
		t.Errorf("expected nil for N=1")
	}

	// Tie + winner + loser + short
	upVals := make([]float64, 65)
	flatVals1 := make([]float64, 65)
	flatVals2 := make([]float64, 65)
	downVals := make([]float64, 65)
	for i := 0; i < 65; i++ {
		upVals[i] = 100 * math.Pow(1.01, float64(i))
		flatVals1[i] = 100
		flatVals2[i] = 100
		downVals[i] = 100 * math.Pow(0.99, float64(i))
	}

	inputs := []StrengthInput{
		{AssetID: "UP", Bars: makeTrendBars("UP", "2025-01-01", upVals)},
		{AssetID: "FLAT1", Bars: makeTrendBars("FLAT1", "2025-01-01", flatVals1)},
		{AssetID: "FLAT2", Bars: makeTrendBars("FLAT2", "2025-01-01", flatVals2)},
		{AssetID: "DOWN", Bars: makeTrendBars("DOWN", "2025-01-01", downVals)},
		{AssetID: "SHORT", Bars: makeConstBars("SHORT", "2025-01-01", 30, 100)},
	}

	res := ComputeRelativeStrength(inputs, "pool1", "Pool 1", "USD", "trend-v1.0.0")
	if len(res) != 5 {
		t.Fatalf("expected 5 results, got %d", len(res))
	}

	byID := make(map[string]StrengthResult)
	for _, r := range res {
		byID[r.AssetID] = r
	}

	if *byID["UP"].Score != 100.0 || *byID["UP"].Rank != 1.0 {
		t.Errorf("UP expected score=100 rank=1, got score=%v rank=%v", *byID["UP"].Score, *byID["UP"].Rank)
	}
	if *byID["FLAT1"].Rank != 2.5 || *byID["FLAT2"].Rank != 2.5 || *byID["FLAT1"].Score != 50.0 {
		t.Errorf("FLAT tie expected rank=2.5 score=50, got rank=%v score=%v", *byID["FLAT1"].Rank, *byID["FLAT1"].Score)
	}
	if *byID["DOWN"].Score != 0.0 || *byID["DOWN"].Rank != 4.0 {
		t.Errorf("DOWN expected score=0 rank=4, got score=%v rank=%v", *byID["DOWN"].Score, *byID["DOWN"].Rank)
	}
	if byID["SHORT"].Score != nil || byID["SHORT"].Reason == "" {
		t.Errorf("SHORT expected nil score + reason, got %+v", byID["SHORT"])
	}
}

// A-04: Tier boundary rules.
func TestComputeTier(t *testing.T) {
	cases := []struct {
		name     string
		score    *float64
		side     Side
		wantTier *Tier
	}{
		{"nil score", nil, SideRight, nil},
		{"unknown side", ptrFloat(95), SideUnknown, nil},
		{"90 right -> S", ptrFloat(90), SideRight, ptrTier(TierS)},
		{"95 neutral -> A", ptrFloat(95), SideNeutral, ptrTier(TierA)},
		{"95 left -> A", ptrFloat(95), SideLeft, ptrTier(TierA)},
		{"70 neutral -> A", ptrFloat(70), SideNeutral, ptrTier(TierA)},
		{"69.9 right -> B", ptrFloat(69.9), SideRight, ptrTier(TierB)},
		{"30 left -> B", ptrFloat(30), SideLeft, ptrTier(TierB)},
		{"29.9 right -> C", ptrFloat(29.9), SideRight, ptrTier(TierC)},
		{"10 left -> C", ptrFloat(10), SideLeft, ptrTier(TierC)},
		{"9.9 left -> D", ptrFloat(9.9), SideLeft, ptrTier(TierD)},
		{"9.9 neutral -> C", ptrFloat(9.9), SideNeutral, ptrTier(TierC)},
		{"0 right -> C", ptrFloat(0), SideRight, ptrTier(TierC)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, reason := ComputeTier(tc.score, tc.side)
			if tc.wantTier == nil {
				if got != nil || reason == "" {
					t.Errorf("expected nil + reason, got %v (%s)", got, reason)
				}
			} else if got == nil || *got != *tc.wantTier {
				t.Errorf("got %v, want %v", got, *tc.wantTier)
			}
		})
	}
}

func ptrTier(t Tier) *Tier { return &t }

// A-06: Extremum O(n) scan.
func TestComputeExtremum(t *testing.T) {
	t.Run("multi-peak drawdown and gain", func(t *testing.T) {
		// 100 -> 120 -> 90 (-25% from 120) -> 150 -> 120 (-20%) -> 180 (+100% from 90)
		bars := makeTrendBars("X", "2025-01-01", []float64{100, 120, 90, 150, 120, 180})
		res := ComputeExtremum(bars)
		if !res.Valid {
			t.Fatalf("expected valid, got %s", res.Reason)
		}
		if math.Abs(res.MaxDrawdown-(-0.25)) > 1e-9 {
			t.Errorf("MaxDrawdown = %v, want -0.25", res.MaxDrawdown)
		}
		if res.DrawdownPeakValue != 120 || res.DrawdownTroughValue != 90 {
			t.Errorf("Drawdown peak/trough mismatch: %+v", res)
		}
		if math.Abs(res.MaxGain-1.0) > 1e-9 {
			t.Errorf("MaxGain = %v, want 1.0", res.MaxGain)
		}
		if res.GainTroughValue != 90 || res.GainPeakValue != 180 {
			t.Errorf("Gain trough/peak mismatch: %+v", res)
		}
	})

	t.Run("flat series", func(t *testing.T) {
		res := ComputeExtremum(makeConstBars("F", "2025-01-01", 5, 100))
		if !res.Valid || res.MaxDrawdown != 0 || res.MaxGain != 0 {
			t.Errorf("expected 0/0, got %+v", res)
		}
	})

	t.Run("monotonic up and down", func(t *testing.T) {
		up := ComputeExtremum(makeTrendBars("U", "2025-01-01", []float64{100, 110, 125}))
		if up.MaxDrawdown != 0 || math.Abs(up.MaxGain-0.25) > 1e-9 {
			t.Errorf("up mismatch: %+v", up)
		}
		down := ComputeExtremum(makeTrendBars("D", "2025-01-01", []float64{100, 90, 80}))
		if down.MaxGain != 0 || math.Abs(down.MaxDrawdown-(-0.20)) > 1e-9 {
			t.Errorf("down mismatch: %+v", down)
		}
	})

	t.Run("non-positive invalid", func(t *testing.T) {
		res := ComputeExtremum(makeTrendBars("BAD", "2025-01-01", []float64{100, -5, 110}))
		if res.Valid {
			t.Errorf("expected invalid for negative value")
		}
	})
}

// A-07: Annual comparison across 4 years, 53rd week, listing year exclusion, asOf truncation.
func TestComputeAnnualComparison(t *testing.T) {
	var bars []DailyBar
	addBar := func(date string, val float64) {
		v := val
		bars = append(bars, DailyBar{AssetID: "IDX", TradingDate: date, Value: &v})
	}

	// 2022: listing year (incomplete, provides 2022-12-30 base = 100)
	addBar("2022-06-01", 90)
	addBar("2022-12-30", 100)

	// 2023: complete year (base 100 -> week 2 = 110 (+10%), week 53 Dec 31 = 120 (+20%))
	addBar("2023-01-10", 110) // week 2 (day 10 -> (10-1)/7+1 = 2)
	addBar("2023-12-31", 120) // week 53

	// 2024: complete leap year (base 120 -> week 2 = 132 (+10%), Dec 31 = 108 (-10%))
	addBar("2024-01-09", 132) // week 2
	addBar("2024-12-31", 108) // week 53

	// 2025: current year truncated at 2025-01-12 (week 2), base = 108
	addBar("2025-01-10", 118.8) // +10%
	addBar("2025-01-20", 999.0) // future after asOf, must be excluded!

	res := ComputeAnnualComparison(bars, "2025-01-12")
	if res.CurrentYear != 2025 || res.CurrentWeek != 2 {
		t.Fatalf("expected 2025 week 2, got %d week %d", res.CurrentYear, res.CurrentWeek)
	}
	if res.CompleteYearCount != 2 {
		t.Fatalf("expected 2 complete historical years (2023, 2024), got %d", res.CompleteYearCount)
	}
	if res.CurrentReturn == nil || math.Abs(*res.CurrentReturn-0.10) > 1e-9 {
		t.Errorf("expected current return 0.10, got %v", res.CurrentReturn)
	}
	if res.SameWeekMedian == nil || math.Abs(*res.SameWeekMedian-0.10) > 1e-9 {
		t.Errorf("expected same-week median 0.10, got %v", res.SameWeekMedian)
	}
	if res.FullYearMean == nil || math.Abs(*res.FullYearMean-0.05) > 1e-9 {
		t.Errorf("expected full-year mean 0.05 ((0.20 + -0.10)/2), got %v", res.FullYearMean)
	}
	if res.FullYearUpRatio == nil || math.Abs(*res.FullYearUpRatio-0.5) > 1e-9 {
		t.Errorf("expected full-year up ratio 0.5, got %v", res.FullYearUpRatio)
	}
	// Verify 2023 has 53rd week
	for _, y := range res.Years {
		if y.Year == 2023 {
			lastW := y.Weeks[len(y.Weeks)-1]
			if lastW.Week != 53 {
				t.Errorf("expected 2023 last week = 53, got %d", lastW.Week)
			}
		}
	}
}
