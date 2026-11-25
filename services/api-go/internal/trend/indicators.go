package trend

import (
	"fmt"
	"math"
	"time"
)

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func ptrFloat(v float64) *float64 {
	return &v
}

func ptrInt(v int) *int {
	return &v
}

func isValidPositive(v *float64) bool {
	return v != nil && !math.IsNaN(*v) && !math.IsInf(*v, 0) && *v > 0
}

func hasInvalidNonNull(v *float64) bool {
	return v != nil && (math.IsNaN(*v) || math.IsInf(*v, 0) || *v <= 0)
}

// computeEMA computes an exponential moving average with explicit SMA initialization.
// Returns a slice of length len(values); indices < period-1 are 0 and unused.
func computeEMA(values []float64, period int) []float64 {
	n := len(values)
	ema := make([]float64, n)
	if period <= 0 || n < period {
		return ema
	}
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	ema[period-1] = sum / float64(period)
	k := 2.0 / float64(period+1)
	for i := period; i < n; i++ {
		ema[i] = values[i]*k + ema[i-1]*(1.0-k)
	}
	return ema
}

// computeSampleStdDev returns the sample standard deviation (N-1 denominator) of values.
func computeSampleStdDev(values []float64) float64 {
	n := len(values)
	if n < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(n)
	sq := 0.0
	for _, v := range values {
		d := v - mean
		sq += d * d
	}
	return math.Sqrt(sq / float64(n-1))
}

// classifyTempLevel maps an unrounded temperature score in [0, 100] to TempLevel.
func classifyTempLevel(score float64) TempLevel {
	switch {
	case score < 15:
		return TempFreezing
	case score < 29:
		return TempCold
	case score < 43:
		return TempCool
	case score < 58:
		return TempFlat
	case score < 72:
		return TempWarm
	case score < 86:
		return TempHot
	default:
		return TempBoiling
	}
}

type coreMetrics struct {
	score   float64
	q       float64
	z20     float64
	z60     float64
	z120    float64
	sigma60 float64
	ema20   float64
	ema60   float64
	slope20 float64
	slope60 float64
	w       float64
}

// computeCoreAtLast computes all temperature & state indicators at the last element of validValues.
// validValues must have length >= cfg.TempWindow (default 121) and all elements > 0.
func computeCoreAtLast(validValues []float64, cfg AlgorithmConfig) coreMetrics {
	n := len(validValues)
	last := n - 1
	w := validValues[last]

	// Past cfg.SigmaWindow (60) single-day log returns ending at last.
	logRets := make([]float64, cfg.SigmaWindow)
	start := last - cfg.SigmaWindow + 1
	for i := 0; i < cfg.SigmaWindow; i++ {
		idx := start + i
		logRets[i] = math.Log(validValues[idx] / validValues[idx-1])
	}
	rawSigma := computeSampleStdDev(logRets)
	sigmaEff := math.Max(rawSigma, cfg.Epsilon)

	r20 := math.Log(w / validValues[last-20])
	r60 := math.Log(w / validValues[last-60])
	r120 := math.Log(w / validValues[last-120])

	z := func(r float64, horizon int) float64 {
		if rawSigma == 0 && r == 0 {
			return 0
		}
		return clamp(r/(sigmaEff*math.Sqrt(float64(horizon))), -4, 4)
	}
	z20 := z(r20, 20)
	z60 := z(r60, 60)
	z120 := z(r120, 120)

	ema20Series := computeEMA(validValues, cfg.EMAFast)
	ema60Series := computeEMA(validValues, cfg.EMASlow)
	ema20 := ema20Series[last]
	ema60 := ema60Series[last]
	ema20Prev := ema20Series[last-cfg.SlopeLag]
	ema60Prev := ema60Series[last-cfg.SlopeLag]

	slope20 := (ema20 - ema20Prev) / math.Max(math.Abs(ema20Prev), cfg.Epsilon)
	slope60 := (ema60 - ema60Prev) / math.Max(math.Abs(ema60Prev), cfg.Epsilon)

	a := 0.0
	logEMA := math.Log(ema20 / ema60)
	if !(rawSigma == 0 && logEMA == 0) {
		a = clamp(logEMA/(sigmaEff*math.Sqrt(20)), -4, 4)
	}
	b := 0.0
	if !(rawSigma == 0 && slope20 == 0) {
		b = clamp(slope20/(sigmaEff*math.Sqrt(5)), -4, 4)
	}
	c := 0.0
	if !(rawSigma == 0 && slope60 == 0) {
		c = clamp(slope60/(sigmaEff*math.Sqrt(5)), -4, 4)
	}

	q := 0.25*z20 + 0.25*z60 + 0.20*z120 + 0.15*a + 0.10*b + 0.05*c
	score := clamp(50.0+50.0*math.Tanh(q/2.0), 0, 100)

	return coreMetrics{
		score:   score,
		q:       q,
		z20:     z20,
		z60:     z60,
		z120:    z120,
		sigma60: rawSigma,
		ema20:   ema20,
		ema60:   ema60,
		slope20: slope20,
		slope60: slope60,
		w:       w,
	}
}

// ComputeTemperature computes the seven-tier absolute trend temperature for the latest bar.
func ComputeTemperature(bars []DailyBar, cfg AlgorithmConfig) TempResult {
	if cfg.TempWindow <= 0 {
		cfg = DefaultAlgorithmConfig()
	}
	if len(bars) == 0 {
		return TempResult{Level: TempUnavailable, Reason: "empty bar series"}
	}
	latestDate := bars[len(bars)-1].TradingDate

	// Check coverage over the most recent TempWindow expected trading days.
	windowStart := len(bars) - cfg.TempWindow
	if windowStart < 0 {
		windowStart = 0
	}
	windowSlice := bars[windowStart:]
	validInWindow := 0
	for _, b := range windowSlice {
		if hasInvalidNonNull(b.Value) {
			return TempResult{
				TradingDate: latestDate,
				Level:       TempUnavailable,
				Reason:      "non-positive or NaN value in window",
			}
		}
		if isValidPositive(b.Value) {
			validInWindow++
		}
	}
	coverage := float64(validInWindow) / float64(cfg.TempWindow)

	if len(bars) < cfg.TempWindow {
		return TempResult{
			TradingDate: latestDate,
			Level:       TempUnavailable,
			ValidCount:  validInWindow,
			Coverage:    coverage,
			Reason:      fmt.Sprintf("insufficient history: got %d bars, need %d", len(bars), cfg.TempWindow),
		}
	}
	if !isValidPositive(bars[len(bars)-1].Value) {
		return TempResult{
			TradingDate: latestDate,
			Level:       TempUnavailable,
			ValidCount:  validInWindow,
			Coverage:    coverage,
			Reason:      "latest bar value is missing or invalid",
		}
	}
	if coverage < cfg.MinCoverage {
		return TempResult{
			TradingDate: latestDate,
			Level:       TempUnavailable,
			ValidCount:  validInWindow,
			Coverage:    coverage,
			Reason:      fmt.Sprintf("insufficient coverage: %.2f%% < %.2f%%", coverage*100, cfg.MinCoverage*100),
		}
	}

	validValues := make([]float64, 0, len(bars))
	for _, b := range bars {
		if hasInvalidNonNull(b.Value) {
			return TempResult{
				TradingDate: latestDate,
				Level:       TempUnavailable,
				ValidCount:  len(validValues),
				Coverage:    coverage,
				Reason:      "non-positive or NaN value in series",
			}
		}
		if isValidPositive(b.Value) {
			validValues = append(validValues, *b.Value)
		}
	}

	if len(validValues) < cfg.TempWindow {
		return TempResult{
			TradingDate: latestDate,
			Level:       TempUnavailable,
			ValidCount:  len(validValues),
			Coverage:    coverage,
			Reason:      fmt.Sprintf("insufficient valid points: got %d, need %d", len(validValues), cfg.TempWindow),
		}
	}

	m := computeCoreAtLast(validValues, cfg)
	disp := int(math.Round(m.score))
	return TempResult{
		TradingDate:  latestDate,
		Score:        ptrFloat(m.score),
		DisplayScore: ptrInt(disp),
		Level:        classifyTempLevel(m.score),
		Q:            ptrFloat(m.q),
		Z20:          ptrFloat(m.z20),
		Z60:          ptrFloat(m.z60),
		Z120:         ptrFloat(m.z120),
		Sigma60:      ptrFloat(m.sigma60),
		EMA20:        ptrFloat(m.ema20),
		EMA60:        ptrFloat(m.ema60),
		Slope20:      ptrFloat(m.slope20),
		Slope60:      ptrFloat(m.slope60),
		ValidCount:   len(validValues),
		Coverage:     coverage,
	}
}

func calendarDaysInclusive(startDate, endDate string) int {
	t1, err1 := time.Parse("2006-01-02", startDate)
	t2, err2 := time.Parse("2006-01-02", endDate)
	if err1 != nil || err2 != nil {
		return 0
	}
	diff := int(t2.Sub(t1).Hours() / 24)
	if diff < 0 {
		return 0
	}
	return diff + 1
}

// ComputeStateSeries evaluates the direction/confirmation/duration state machine across bars
// and emits idempotent TrendEvents on transitions.
func ComputeStateSeries(bars []DailyBar, cfg AlgorithmConfig) ([]StateResult, []TrendEvent) {
	if cfg.TempWindow <= 0 {
		cfg = DefaultAlgorithmConfig()
	}
	results := make([]StateResult, len(bars))
	var events []TrendEvent

	// Persistent state across valid trading days.
	dir := DirectionNeutral
	conf := ConfirmationNone
	side := SideNeutral
	startedAt := ""
	confirmedAt := ""
	durationTradingDays := 0
	formingDays := 0
	exitStreak := 0
	initialized := false

	validValues := make([]float64, 0, len(bars))

	emitEvent := func(assetID, eventType, oldVal, newVal, asOf string) {
		dedupeKey := fmt.Sprintf("%s:%s:%s:%s", assetID, eventType, asOf, cfg.AlgorithmVersion)
		events = append(events, TrendEvent{
			EventID:          dedupeKey,
			AssetID:          assetID,
			EventType:        eventType,
			OldValue:         oldVal,
			NewValue:         newVal,
			AsOf:             asOf,
			DetectedAt:       asOf,
			AlgorithmVersion: cfg.AlgorithmVersion,
			DedupeKey:        dedupeKey,
		})
	}

	for i, b := range bars {
		if !isValidPositive(b.Value) {
			results[i] = StateResult{
				TradingDate:  b.TradingDate,
				Direction:    DirectionUnknown,
				Confirmation: ConfirmationUnknown,
				Side:         SideUnknown,
				StateVersion: cfg.AlgorithmVersion,
				Stale:        true,
				Reason:       "missing or invalid bar",
			}
			continue
		}

		validValues = append(validValues, *b.Value)

		// Check recent window coverage.
		winStart := i - cfg.TempWindow + 1
		if winStart < 0 {
			winStart = 0
		}
		validInWin := 0
		for j := winStart; j <= i; j++ {
			if isValidPositive(bars[j].Value) {
				validInWin++
			}
		}
		coverage := float64(validInWin) / float64(cfg.TempWindow)

		if len(validValues) < cfg.TempWindow || coverage < cfg.MinCoverage {
			results[i] = StateResult{
				TradingDate:  b.TradingDate,
				Direction:    DirectionUnknown,
				Confirmation: ConfirmationUnknown,
				Side:         SideUnknown,
				StateVersion: cfg.AlgorithmVersion,
				Stale:        false,
				Reason:       "insufficient history or coverage",
			}
			continue
		}

		m := computeCoreAtLast(validValues, cfg)
		upCond := m.w > m.ema60 && m.ema20 > m.ema60 && m.slope20 > 0 && m.q >= cfg.UpEntryQ
		downCond := m.w < m.ema60 && m.ema20 < m.ema60 && m.slope20 < 0 && m.q <= cfg.DownEntryQ
		upExitCond := m.w < m.ema60 || m.q < cfg.UpExitQ
		downExitCond := m.w > m.ema60 || m.q > cfg.DownExitQ

		if !initialized {
			initialized = true
		}

		switch {
		case conf == ConfirmationConfirmed && dir == DirectionUp:
			if upExitCond {
				exitStreak++
				if exitStreak >= cfg.ExitDays {
					emitEvent(b.AssetID, "right_ended", string(SideRight), string(SideNeutral), b.TradingDate)
					dir = DirectionNeutral
					conf = ConfirmationNone
					side = SideNeutral
					startedAt = ""
					confirmedAt = ""
					durationTradingDays = 0
					formingDays = 0
					exitStreak = 0

					// Exit takes priority; if reverse condition holds on exit day, start reverse forming count at 1.
					if downCond {
						dir = DirectionDown
						conf = ConfirmationForming
						startedAt = b.TradingDate
						formingDays = 1
					}
				} else {
					durationTradingDays++
				}
			} else {
				exitStreak = 0
				durationTradingDays++
			}

		case conf == ConfirmationConfirmed && dir == DirectionDown:
			if downExitCond {
				exitStreak++
				if exitStreak >= cfg.ExitDays {
					emitEvent(b.AssetID, "down_ended", string(SideLeft), string(SideNeutral), b.TradingDate)
					dir = DirectionNeutral
					conf = ConfirmationNone
					side = SideNeutral
					startedAt = ""
					confirmedAt = ""
					durationTradingDays = 0
					formingDays = 0
					exitStreak = 0

					if upCond {
						dir = DirectionUp
						conf = ConfirmationForming
						startedAt = b.TradingDate
						formingDays = 1
					}
				} else {
					durationTradingDays++
				}
			} else {
				exitStreak = 0
				durationTradingDays++
			}

		case conf == ConfirmationForming && dir == DirectionUp:
			if upCond {
				formingDays++
				if formingDays >= cfg.ConfirmDays {
					conf = ConfirmationConfirmed
					side = SideRight
					confirmedAt = b.TradingDate
					durationTradingDays = 1
					exitStreak = 0
					emitEvent(b.AssetID, "right_started", string(SideNeutral), string(SideRight), b.TradingDate)
				}
			} else if downCond {
				dir = DirectionDown
				conf = ConfirmationForming
				side = SideNeutral
				startedAt = b.TradingDate
				formingDays = 1
			} else {
				dir = DirectionNeutral
				conf = ConfirmationNone
				side = SideNeutral
				startedAt = ""
				formingDays = 0
			}

		case conf == ConfirmationForming && dir == DirectionDown:
			if downCond {
				formingDays++
				if formingDays >= cfg.ConfirmDays {
					conf = ConfirmationConfirmed
					side = SideLeft
					confirmedAt = b.TradingDate
					durationTradingDays = 1
					exitStreak = 0
					emitEvent(b.AssetID, "down_started", string(SideNeutral), string(SideLeft), b.TradingDate)
				}
			} else if upCond {
				dir = DirectionUp
				conf = ConfirmationForming
				side = SideNeutral
				startedAt = b.TradingDate
				formingDays = 1
			} else {
				dir = DirectionNeutral
				conf = ConfirmationNone
				side = SideNeutral
				startedAt = ""
				formingDays = 0
			}

		default: // Neutral / None
			if upCond {
				dir = DirectionUp
				conf = ConfirmationForming
				side = SideNeutral
				startedAt = b.TradingDate
				formingDays = 1
			} else if downCond {
				dir = DirectionDown
				conf = ConfirmationForming
				side = SideNeutral
				startedAt = b.TradingDate
				formingDays = 1
			} else {
				dir = DirectionNeutral
				conf = ConfirmationNone
				side = SideNeutral
				startedAt = ""
				formingDays = 0
			}
		}

		calDays := 0
		if conf == ConfirmationConfirmed && confirmedAt != "" {
			calDays = calendarDaysInclusive(confirmedAt, b.TradingDate)
		}

		results[i] = StateResult{
			TradingDate:          b.TradingDate,
			Direction:            dir,
			Confirmation:         conf,
			Side:                 side,
			StartedAt:            startedAt,
			ConfirmedAt:          confirmedAt,
			DurationTradingDays:  durationTradingDays,
			DurationCalendarDays: calDays,
			FormingDays:          formingDays,
			ExitStreak:           exitStreak,
			StateVersion:         cfg.AlgorithmVersion,
			Stale:                false,
		}
	}

	return results, events
}

// ComputePhase maps a StateResult to its trend solar-term lifecycle phase, independent of calendar dates.
func ComputePhase(state StateResult) PhaseResult {
	base := PhaseResult{
		Direction:           state.Direction,
		Confirmation:        state.Confirmation,
		DurationTradingDays: state.DurationTradingDays,
	}

	if state.Direction == DirectionUnknown || state.Confirmation == ConfirmationUnknown {
		base.PhaseCode = "unknown"
		base.PhaseName = "未知"
		base.Description = "状态未知或数据不足"
		return base
	}

	if state.Direction == DirectionNeutral || state.Confirmation == ConfirmationNone {
		base.PhaseCode = "oscillation"
		base.PhaseName = "震荡"
		base.Description = "趋势中性震荡"
		return base
	}

	if state.Direction == DirectionUp {
		if state.Confirmation == ConfirmationForming {
			base.PhaseCode = "guyu"
			base.PhaseName = "谷雨"
			base.Description = "上行形成中（未确认）"
			return base
		}
		d := state.DurationTradingDays
		switch {
		case d <= 10:
			base.PhaseCode = "lixia"
			base.PhaseName = "立夏"
			base.Description = "已确认上行 1–10 交易日"
		case d <= 40:
			base.PhaseCode = "xiazhi"
			base.PhaseName = "夏至"
			base.Description = "已确认上行 11–40 交易日"
		case d <= 90:
			base.PhaseCode = "xiaoshu"
			base.PhaseName = "小暑"
			base.Description = "已确认上行 41–90 交易日"
		default:
			base.PhaseCode = "dashu"
			base.PhaseName = "大暑"
			base.Description = "已确认上行 >90 交易日"
		}
		return base
	}

	if state.Direction == DirectionDown {
		if state.Confirmation == ConfirmationForming {
			base.PhaseCode = "hanlu"
			base.PhaseName = "寒露"
			base.Description = "下行形成中（未确认）"
			return base
		}
		d := state.DurationTradingDays
		switch {
		case d <= 10:
			base.PhaseCode = "lidong"
			base.PhaseName = "立冬"
			base.Description = "已确认下行 1–10 交易日"
		case d <= 40:
			base.PhaseCode = "dongzhi"
			base.PhaseName = "冬至"
			base.Description = "已确认下行 11–40 交易日"
		case d <= 90:
			base.PhaseCode = "xiaohan"
			base.PhaseName = "小寒"
			base.Description = "已确认下行 41–90 交易日"
		default:
			base.PhaseCode = "dahan"
			base.PhaseName = "大寒"
			base.Description = "已确认下行 >90 交易日"
		}
		return base
	}

	base.PhaseCode = "unknown"
	base.PhaseName = "未知"
	base.Description = "状态未知"
	return base
}
