package trend

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// ComputeExtremum performs an O(n) scan over positive wealth index bars to find
// maximum drawdown (peak before trough) and maximum gain (trough before peak).
func ComputeExtremum(bars []DailyBar) ExtremumResult {
	if len(bars) == 0 {
		return ExtremumResult{Valid: false, Reason: "empty bar series"}
	}

	var validBars []struct {
		date string
		val  float64
	}
	for _, b := range bars {
		if hasInvalidNonNull(b.Value) {
			return ExtremumResult{Valid: false, Reason: "non-positive or NaN value in series"}
		}
		if isValidPositive(b.Value) {
			validBars = append(validBars, struct {
				date string
				val  float64
			}{date: b.TradingDate, val: *b.Value})
		}
	}

	if len(validBars) == 0 {
		return ExtremumResult{Valid: false, Reason: "no valid positive values"}
	}

	first := validBars[0]
	runMaxVal, runMaxDate := first.val, first.date
	runMinVal, runMinDate := first.val, first.date

	res := ExtremumResult{
		MaxDrawdown:         0,
		DrawdownPeakDate:    first.date,
		DrawdownPeakValue:   first.val,
		DrawdownTroughDate:  first.date,
		DrawdownTroughValue: first.val,
		MaxGain:             0,
		GainTroughDate:      first.date,
		GainTroughValue:     first.val,
		GainPeakDate:        first.date,
		GainPeakValue:       first.val,
		Valid:               true,
	}

	for _, pt := range validBars[1:] {
		if pt.val > runMaxVal {
			runMaxVal = pt.val
			runMaxDate = pt.date
		}
		dd := pt.val/runMaxVal - 1.0
		if dd < res.MaxDrawdown {
			res.MaxDrawdown = dd
			res.DrawdownPeakDate = runMaxDate
			res.DrawdownPeakValue = runMaxVal
			res.DrawdownTroughDate = pt.date
			res.DrawdownTroughValue = pt.val
		}

		if pt.val < runMinVal {
			runMinVal = pt.val
			runMinDate = pt.date
		}
		gain := pt.val/runMinVal - 1.0
		if gain > res.MaxGain {
			res.MaxGain = gain
			res.GainTroughDate = runMinDate
			res.GainTroughValue = runMinVal
			res.GainPeakDate = pt.date
			res.GainPeakValue = pt.val
		}
	}

	return res
}

func daysInYear(year int) int {
	if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
		return 366
	}
	return 365
}

func naturalWeekOf(t time.Time) int {
	return (t.YearDay()-1)/7 + 1
}

func naturalWeekEndDate(year, week int) time.Time {
	endDay := week * 7
	maxDay := daysInYear(year)
	if endDay > maxDay {
		endDay = maxDay
	}
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	return jan1.AddDate(0, 0, endDay-1)
}

func computeMean(vals []float64) float64 {
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func computeMedian(vals []float64) float64 {
	cp := make([]float64, len(vals))
	copy(cp, vals)
	sort.Float64s(cp)
	n := len(cp)
	if n%2 == 1 {
		return cp[n/2]
	}
	return (cp[n/2-1] + cp[n/2]) / 2.0
}

// ComputeAnnualComparison computes natural-year 53-week annual comparisons anchored to the
// prior year's last valid close, truncated to asOf for the current year.
func ComputeAnnualComparison(bars []DailyBar, asOf string) AnnualComparisonResult {
	type parsedBar struct {
		t    time.Time
		date string
		val  float64
	}
	var valid []parsedBar
	for _, b := range bars {
		if hasInvalidNonNull(b.Value) {
			return AnnualComparisonResult{AsOf: asOf, Reason: "non-positive or NaN value in series"}
		}
		if !isValidPositive(b.Value) {
			continue
		}
		t, err := time.Parse("2006-01-02", b.TradingDate)
		if err != nil {
			continue
		}
		valid = append(valid, parsedBar{t: t, date: b.TradingDate, val: *b.Value})
	}

	if len(valid) == 0 {
		return AnnualComparisonResult{AsOf: asOf, Reason: "no valid bars"}
	}

	sort.Slice(valid, func(i, j int) bool {
		return valid[i].date < valid[j].date
	})

	var asOfTime time.Time
	var err error
	if asOf == "" {
		asOfTime = valid[len(valid)-1].t
		asOf = valid[len(valid)-1].date
	} else {
		asOfTime, err = time.Parse("2006-01-02", asOf)
		if err != nil {
			return AnnualComparisonResult{AsOf: asOf, Reason: fmt.Sprintf("invalid as_of date: %s", asOf)}
		}
	}

	// Filter bars <= asOf.
	filtered := valid[:0]
	for _, pb := range valid {
		if !pb.t.After(asOfTime) {
			filtered = append(filtered, pb)
		}
	}
	if len(filtered) == 0 {
		return AnnualComparisonResult{AsOf: asOf, Reason: "no valid bars on or before as_of"}
	}

	// Group bars by year.
	byYear := make(map[int][]parsedBar)
	var yearsList []int
	for _, pb := range filtered {
		y := pb.t.Year()
		if _, exists := byYear[y]; !exists {
			yearsList = append(yearsList, y)
		}
		byYear[y] = append(byYear[y], pb)
	}
	sort.Ints(yearsList)

	currentYear := asOfTime.Year()
	currentWeek := naturalWeekOf(asOfTime)

	var summaries []AnnualYearSummary
	var sameWeekHist []float64
	var fullYearHist []float64
	var currentReturn *float64

	for _, y := range yearsList {
		prevBars, hasPrev := byYear[y-1]
		yBars := byYear[y]
		if !hasPrev || len(prevBars) == 0 {
			// Listing year or missing prior year close: cannot compute YTD from prior year end.
			summaries = append(summaries, AnnualYearSummary{
				Year:       y,
				IsComplete: false,
			})
			continue
		}

		baseBar := prevBars[len(prevBars)-1]
		baseClose := baseBar.val

		maxWeek := 53
		cutoff := naturalWeekEndDate(y, 53)
		if y == currentYear {
			maxWeek = currentWeek
			cutoff = asOfTime
		}

		weeks := make([]AnnualWeekResult, 0, maxWeek)
		bIdx := 0
		var lastSeen *parsedBar
		var sameWeekRet *float64
		var sameWeekDate string

		for w := 1; w <= maxWeek; w++ {
			wEnd := naturalWeekEndDate(y, w)
			effectiveEnd := wEnd
			if y == currentYear && effectiveEnd.After(cutoff) {
				effectiveEnd = cutoff
			}

			for bIdx < len(yBars) && !yBars[bIdx].t.After(effectiveEnd) {
				pb := yBars[bIdx]
				lastSeen = &pb
				bIdx++
			}

			if lastSeen != nil {
				ret := lastSeen.val/baseClose - 1.0
				weeks = append(weeks, AnnualWeekResult{
					Year:        y,
					Week:        w,
					WeekEndDate: wEnd.Format("2006-01-02"),
					ActualDate:  lastSeen.date,
					Close:       lastSeen.val,
					Return:      ret,
				})
				if w == currentWeek {
					sameWeekRet = ptrFloat(ret)
					sameWeekDate = lastSeen.date
				}
			}
		}

		var fullYearRet *float64
		var fullYearDate string
		isComplete := false
		if y < currentYear && len(yBars) > 0 {
			isComplete = true
			lastBar := yBars[len(yBars)-1]
			fullYearRet = ptrFloat(lastBar.val/baseClose - 1.0)
			fullYearDate = lastBar.date
			fullYearHist = append(fullYearHist, *fullYearRet)
			if sameWeekRet != nil {
				sameWeekHist = append(sameWeekHist, *sameWeekRet)
			}
		}
		if y == currentYear {
			currentReturn = sameWeekRet
		}

		summaries = append(summaries, AnnualYearSummary{
			Year:           y,
			BaseDate:       baseBar.date,
			BaseClose:      baseClose,
			SameWeekReturn: sameWeekRet,
			SameWeekDate:   sameWeekDate,
			FullYearReturn: fullYearRet,
			FullYearDate:   fullYearDate,
			IsComplete:     isComplete,
			Weeks:          weeks,
		})
	}

	res := AnnualComparisonResult{
		AsOf:              asOf,
		CurrentYear:       currentYear,
		CurrentWeek:       currentWeek,
		CurrentReturn:     currentReturn,
		Years:             summaries,
		CompleteYearCount: len(fullYearHist),
	}

	if len(sameWeekHist) > 0 {
		res.SameWeekMedian = ptrFloat(computeMedian(sameWeekHist))
		res.SameWeekMean = ptrFloat(computeMean(sameWeekHist))
		up := 0
		for _, v := range sameWeekHist {
			if v > 0 {
				up++
			}
		}
		res.SameWeekUpRatio = ptrFloat(float64(up) / float64(len(sameWeekHist)))
		if currentReturn != nil && !math.IsNaN(*currentReturn) {
			leCount := 0
			for _, v := range sameWeekHist {
				if v <= *currentReturn {
					leCount++
				}
			}
			res.SameWeekPercentile = ptrFloat(float64(leCount) / float64(len(sameWeekHist)))
		}
	}

	if len(fullYearHist) > 0 {
		res.FullYearMedian = ptrFloat(computeMedian(fullYearHist))
		res.FullYearMean = ptrFloat(computeMean(fullYearHist))
		up := 0
		for _, v := range fullYearHist {
			if v > 0 {
				up++
			}
		}
		res.FullYearUpRatio = ptrFloat(float64(up) / float64(len(fullYearHist)))
	}

	return res
}
