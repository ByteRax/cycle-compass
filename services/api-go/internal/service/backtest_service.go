package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"cyclecompass/api/internal/dto"
	"cyclecompass/api/internal/market"
)

type BacktestCacheKey struct {
	LowLine      int64
	HighLine     int64
	WeeklyAmount int64
	Days         int
}

type cachedBacktestResult struct {
	output    dto.BacktestOutput
	jsonBytes []byte
}

type inFlightBacktest struct {
	wg  sync.WaitGroup
	res cachedBacktestResult
	err error
}

type BacktestService struct {
	seedProvider *market.SeedProvider
	cache        map[BacktestCacheKey]cachedBacktestResult
	cacheOrder   []BacktestCacheKey
	inFlight     map[BacktestCacheKey]*inFlightBacktest
	cacheMu      sync.RWMutex
}

func NewBacktestService(seedProvider *market.SeedProvider) *BacktestService {
	return &BacktestService{
		seedProvider: seedProvider,
		cache:        make(map[BacktestCacheKey]cachedBacktestResult),
		cacheOrder:   make([]BacktestCacheKey, 0, 64),
		inFlight:     make(map[BacktestCacheKey]*inFlightBacktest),
	}
}

type pendingRebuyInfo struct {
	Zone      string
	SinceYear int
	SinceWeek int
}

func normalizeAndValidateBacktestReq(req *dto.BacktestRequest) error {
	if req.LowLine == 0 {
		req.LowLine = 0.16
	}
	if req.HighLine == 0 {
		req.HighLine = 0.28
	}
	if req.WeeklyAmount == 0 {
		req.WeeklyAmount = 1000.0
	}
	if req.Days == 0 {
		req.Days = 1500
	}

	if req.LowLine < 0.05 || req.LowLine > 0.50 {
		return fmt.Errorf("轮动低轨 (low_line) 需在 0.05 ~ 0.50 之间")
	}
	if req.HighLine < 0.10 || req.HighLine > 0.80 {
		return fmt.Errorf("轮动高轨 (high_line) 需在 0.10 ~ 0.80 之间")
	}
	if req.LowLine >= req.HighLine {
		return fmt.Errorf("轮动低轨 (low_line) 必须严格小于轮动高轨 (high_line)")
	}
	if req.WeeklyAmount < 100.0 {
		return fmt.Errorf("每周定投金额 (weekly_amount) 需不少于 100 元")
	}
	if req.Days < 100 || req.Days > 3000 {
		return fmt.Errorf("历史回测天数 (days) 需在 100 ~ 3000 之间")
	}
	return nil
}

func (s *BacktestService) runOrJoin(req dto.BacktestRequest) (cachedBacktestResult, error) {
	if err := normalizeAndValidateBacktestReq(&req); err != nil {
		return cachedBacktestResult{}, err
	}

	key := BacktestCacheKey{
		LowLine:      int64(math.Round(req.LowLine * 10000)),
		HighLine:     int64(math.Round(req.HighLine * 10000)),
		WeeklyAmount: int64(math.Round(req.WeeklyAmount * 100)),
		Days:         req.Days,
	}

	s.cacheMu.RLock()
	if val, ok := s.cache[key]; ok {
		s.cacheMu.RUnlock()
		return val, nil
	}
	s.cacheMu.RUnlock()

	// Singleflight 并发防击穿：合并同参数并发计算
	s.cacheMu.Lock()
	if val, ok := s.cache[key]; ok {
		s.cacheMu.Unlock()
		return val, nil
	}
	if call, ok := s.inFlight[key]; ok {
		s.cacheMu.Unlock()
		call.wg.Wait()
		return call.res, call.err
	}
	call := &inFlightBacktest{}
	call.wg.Add(1)
	s.inFlight[key] = call
	s.cacheMu.Unlock()

	var (
		res cachedBacktestResult
		err error
	)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("backtest panic: %v", r)
		}
		s.cacheMu.Lock()
		if err == nil {
			if _, exists := s.cache[key]; !exists {
				if len(s.cache) >= 64 && len(s.cacheOrder) > 0 {
					oldest := s.cacheOrder[0]
					s.cacheOrder = s.cacheOrder[1:]
					delete(s.cache, oldest)
				}
				s.cacheOrder = append(s.cacheOrder, key)
			}
			s.cache[key] = res
		}
		delete(s.inFlight, key)
		call.res = res
		call.err = err
		call.wg.Done()
		s.cacheMu.Unlock()
	}()

	res, err = s.computeBacktest(req)
	return res, err
}

func (s *BacktestService) RunBacktest(req dto.BacktestRequest) (*dto.BacktestOutput, error) {
	res, err := s.runOrJoin(req)
	if err != nil {
		return nil, err
	}
	out := res.output
	return &out, nil
}

func (s *BacktestService) computeBacktest(req dto.BacktestRequest) (cachedBacktestResult, error) {
	gKlines := s.seedProvider.GetHistoricalKlines("399673", req.Days+100)
	dKlines := s.seedProvider.GetHistoricalKlines("930955", req.Days+100)

	if len(gKlines) == 0 || len(dKlines) == 0 {
		return cachedBacktestResult{}, fmt.Errorf("历史行情数据不足，无法执行回测")
	}

	dMap := make(map[string]float64, len(dKlines))
	for _, p := range dKlines {
		dMap[p.Date] = p.Close
	}

	type alignedDay struct {
		dateStr string
		date    time.Time
		gPrice  float64
		dPrice  float64
	}

	aligned := make([]alignedDay, 0, len(gKlines))
	for _, g := range gKlines {
		if dPrice, ok := dMap[g.Date]; ok && dPrice > 0 && g.Close > 0 {
			aligned = append(aligned, alignedDay{
				dateStr: g.Date,
				date:    g.ParsedDate,
				gPrice:  g.Close,
				dPrice:  dPrice,
			})
		}
	}

	if len(aligned) > req.Days {
		aligned = aligned[len(aligned)-req.Days:]
	}

	n := len(aligned)
	if n < 2 {
		return cachedBacktestResult{}, fmt.Errorf("对齐后的有效交易日不足")
	}

	var cashPool float64
	var pendingRebuy *pendingRebuyInfo
	var lastZone string
	var lastWeeklyDate string

	var growthShares float64
	var dividendShares float64
	var totalInvest float64
	var rebuyCount int
	zoneCounts := map[string]int{"低区": 0, "中间区": 0, "高区": 0}

	dailyRecords := make([]dto.DailyRecord, 0, n)
	estWeekly := n/5 + 16
	allLedgers := make([]dto.LedgerEntry, 0, estWeekly*3)
	cashFlows := make([]cfPoint, 0, estWeekly*2)

	// 时间加权净值（基金份额法 TWR）：隔离每周追加本金对回撤与夏普比率的失真影响，与 Python 引擎对齐
	twrNavs := make([]float64, 0, n)
	unitNav := 1.0
	units := 0.0

	for i := 0; i < n; i++ {
		day := aligned[i]
		r := day.gPrice / day.dPrice

		var z string
		if r < req.LowLine {
			z = "低区"
		} else if r >= req.HighLine {
			z = "高区"
		} else {
			z = "中间区"
		}
		zoneCounts[z]++

		isWeeklyDay := (day.date.Weekday() == time.Monday) || (i > 0 && day.date.Weekday() < aligned[i-1].date.Weekday())

		year, week := day.date.ISOWeek()
		if z == "中间区" {
			pendingRebuy = nil
		} else if lastZone == "" || lastZone == "中间区" {
			pendingRebuy = &pendingRebuyInfo{
				Zone:      z,
				SinceYear: year,
				SinceWeek: week,
			}
		} else if pendingRebuy != nil && pendingRebuy.Zone != z {
			pendingRebuy.Zone = z
		}

		dayInflow := 0.0
		if isWeeklyDay && lastWeeklyDate != day.dateStr {
			hasPendingRebuy := pendingRebuy != nil &&
				(pendingRebuy.SinceYear != year || pendingRebuy.SinceWeek != week) &&
				z != "中间区" &&
				cashPool > 0

			targetName := "创业板50(159949)"
			targetSymbol := "159949"
			targetPrice := day.gPrice
			if z == "高区" {
				targetName = "红利低波100(515100)"
				targetSymbol = "515100"
				targetPrice = day.dPrice
			}

			if hasPendingRebuy {
				poolAmount := math.Round(cashPool*100) / 100
				shares := 0.0
				if targetPrice > 0 {
					shares = math.Round(poolAmount/targetPrice*10000) / 10000
				}
				rebuyCount++
				sym := targetSymbol
				allLedgers = append(allLedgers, dto.LedgerEntry{
					Date:             day.dateStr,
					EventType:        "cash_pool_rebuy",
					Symbol:           &sym,
					TargetName:       targetName,
					Shares:           shares,
					Price:            targetPrice,
					Amount:           poolAmount,
					CashInflow:       0.0,
					CashOutflow:      poolAmount,
					CashPoolDelta:    -poolAmount,
					CashBalanceAfter: 0.0,
					CashPoolAfter:    0.0,
					Description:      fmt.Sprintf("次周一次性回补：释放现金池 %.2f 元全额买入 %s", poolAmount, targetName),
				})
				if targetSymbol == "159949" {
					growthShares += shares
				} else {
					dividendShares += shares
				}
				cashPool = 0.0
				pendingRebuy = nil
			}

			if z == "中间区" {
				cashPool += req.WeeklyAmount
				cashPool = math.Round(cashPool*100) / 100
				totalInvest += req.WeeklyAmount
				dayInflow += req.WeeklyAmount
				cashFlows = append(cashFlows, cfPoint{t: day.date, amt: -req.WeeklyAmount})
				allLedgers = append(allLedgers, dto.LedgerEntry{
					Date:             day.dateStr,
					EventType:        "cash_deposit",
					Symbol:           nil,
					TargetName:       "计划现金池",
					Shares:           0.0,
					Price:            1.0,
					Amount:           req.WeeklyAmount,
					CashInflow:       req.WeeklyAmount,
					CashOutflow:      req.WeeklyAmount,
					CashPoolDelta:    req.WeeklyAmount,
					CashBalanceAfter: 0.0,
					CashPoolAfter:    cashPool,
					Description:      fmt.Sprintf("中间区：本周定投资金 %.0f 元存入计划现金池", req.WeeklyAmount),
				})
			} else {
				totalInvest += req.WeeklyAmount
				dayInflow += req.WeeklyAmount
				cashFlows = append(cashFlows, cfPoint{t: day.date, amt: -req.WeeklyAmount})
				shares := 0.0
				if targetPrice > 0 {
					shares = math.Round(req.WeeklyAmount/targetPrice*10000) / 10000
				}
				if targetSymbol == "159949" {
					growthShares += shares
				} else {
					dividendShares += shares
				}
				sym := targetSymbol
				allLedgers = append(allLedgers, dto.LedgerEntry{
					Date:             day.dateStr,
					EventType:        "weekly_invest",
					Symbol:           &sym,
					TargetName:       targetName,
					Shares:           shares,
					Price:            targetPrice,
					Amount:           req.WeeklyAmount,
					CashInflow:       req.WeeklyAmount,
					CashOutflow:      req.WeeklyAmount,
					CashPoolDelta:    0.0,
					CashBalanceAfter: 0.0,
					CashPoolAfter:    cashPool,
					Description:      fmt.Sprintf("本周定投 %.0f 元买入 %s", req.WeeklyAmount, targetName),
				})
			}
			lastWeeklyDate = day.dateStr
		}

		lastZone = z

		marketVal := growthShares*day.gPrice + dividendShares*day.dPrice
		totalAsset := marketVal + cashPool
		netProfit := totalAsset - totalInvest
		retPct := 0.0
		if totalInvest > 0 {
			retPct = (netProfit / totalInvest) * 100.0
		}

		if dayInflow > 0 {
			preInflowAsset := math.Max(totalAsset-dayInflow, 0.0)
			if units > 0 && preInflowAsset > 0 {
				unitNav = preInflowAsset / units
			}
			navDivisor := math.Max(unitNav, 1e-6)
			units += dayInflow / navDivisor
		}
		if units > 0 {
			unitNav = totalAsset / units
		}
		twrNavs = append(twrNavs, unitNav)

		dailyRecords = append(dailyRecords, dto.DailyRecord{
			Date:           day.dateStr,
			Ratio:          math.Round(r*10000) / 10000,
			Zone:           z,
			GrowthPrice:    math.Round(day.gPrice*10000) / 10000,
			DividendPrice:  math.Round(day.dPrice*10000) / 10000,
			GrowthShares:   math.Round(growthShares*10000) / 10000,
			DividendShares: math.Round(dividendShares*10000) / 10000,
			MarketValue:    math.Round(marketVal*100) / 100,
			CashPool:       math.Round(cashPool*100) / 100,
			TotalAsset:     math.Round(totalAsset*100) / 100,
			TotalInvest:    math.Round(totalInvest*100) / 100,
			NetProfit:      math.Round(netProfit*100) / 100,
			ReturnPct:      math.Round(retPct*100) / 100,
		})
	}

	finalAsset := dailyRecords[len(dailyRecords)-1].TotalAsset
	if finalAsset > 0 {
		cashFlows = append(cashFlows, cfPoint{t: aligned[len(aligned)-1].date, amt: finalAsset})
	}

	xirrVal := calcXIRR(cashFlows)
	mdd := calcMaxDrawdown(twrNavs)

	dailyReturns := make([]float64, 0, len(twrNavs)-1)
	for i := 1; i < len(twrNavs); i++ {
		prev := twrNavs[i-1]
		if prev > 0 {
			dailyReturns = append(dailyReturns, (twrNavs[i]-prev)/prev)
		} else {
			dailyReturns = append(dailyReturns, 0.0)
		}
	}

	sharpe := calcSharpeRatio(dailyReturns)

	daysSpan := aligned[len(aligned)-1].date.Sub(aligned[0].date).Hours() / 24.0
	years := math.Max(daysSpan/365.25, 0.01)
	totRet := 0.0
	if totalInvest > 0 {
		totRet = (finalAsset - totalInvest) / totalInvest
	}
	annualized := 0.0
	if totRet > -1.0 {
		annualized = math.Pow(1.0+totRet, 1.0/years) - 1.0
	}
	// Calmar Ratio: 分子采用与 TWR 最大回撤口径一致的 TWR 年化收益率，保证风险调整指标口径统一
	twrTotRet := totRet
	if len(twrNavs) > 0 && twrNavs[0] > 0 {
		twrTotRet = twrNavs[len(twrNavs)-1]/twrNavs[0] - 1.0
	}
	annualizedTWR := -1.0
	if twrTotRet > -1.0 {
		annualizedTWR = math.Pow(1.0+twrTotRet, 1.0/years) - 1.0
	}
	calmar := 0.0
	if mdd > 1e-6 && annualizedTWR > -1.0 {
		calmar = annualizedTWR / mdd
	}

	summary := dto.BacktestSummary{
		StartDate:          aligned[0].dateStr,
		EndDate:            aligned[len(aligned)-1].dateStr,
		TradingDays:        n,
		TotalInvest:        math.Round(totalInvest*100) / 100,
		FinalAsset:         math.Round(finalAsset*100) / 100,
		NetProfit:          math.Round((finalAsset-totalInvest)*100) / 100,
		TotalReturnPct:     math.Round(totRet*10000) / 100,
		AnnualizedReturn:   math.Round(annualized*10000) / 100,
		XIRR:               math.Round(xirrVal*10000) / 100,
		MaxDrawdown:        math.Round(mdd*10000) / 100,
		SharpeRatio:        math.Round(sharpe*100) / 100,
		CalmarRatio:        math.Round(calmar*100) / 100,
		ZoneDistribution:   zoneCounts,
		TotalRebuyCount:    rebuyCount,
		LedgerEntriesCount: len(allLedgers),
		RuleVersion:        "rotation-v1.0.0",
		DataSource:         "local_csv(399673,930955)",
		ExecutionTiming:    "ISO业务周首个交易日盘后判断；离开中间区于次周执行日全额回补",
		ClipReason:         fmt.Sprintf("按两指数可用交集裁剪，实际有效 %d 交易日（%s ~ %s）", n, aligned[0].dateStr, aligned[len(aligned)-1].dateStr),
	}

	out := dto.BacktestOutput{
		Summary:       summary,
		DailyRecords:  dailyRecords,
		LedgerEntries: allLedgers,
	}

	resp := dto.Success(out)
	jsonBytes, _ := json.Marshal(resp)

	return cachedBacktestResult{
		output:    out,
		jsonBytes: jsonBytes,
	}, nil
}

func (s *BacktestService) RunBacktestRaw(req dto.BacktestRequest) ([]byte, error) {
	res, err := s.runOrJoin(req)
	if err != nil {
		return nil, err
	}
	return res.jsonBytes, nil
}

func calcMaxDrawdown(navs []float64) float64 {
	if len(navs) < 2 {
		return 0.0
	}
	currentPeak := navs[0]
	maxDD := 0.0

	for _, val := range navs {
		if val > currentPeak {
			currentPeak = val
		} else if currentPeak > 0 {
			dd := (currentPeak - val) / currentPeak
			if dd > maxDD {
				maxDD = dd
			}
		}
	}
	return maxDD
}

func calcSharpeRatio(returns []float64) float64 {
	if len(returns) < 5 {
		return 0.0
	}
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	var varianceSum float64
	for _, r := range returns {
		varianceSum += (r - mean) * (r - mean)
	}
	std := math.Sqrt(varianceSum / float64(len(returns)-1))
	if std < 1e-8 {
		return 0.0
	}

	rfDaily := math.Pow(1.0+0.02, 1.0/252.0) - 1.0
	return (mean - rfDaily) / std * math.Sqrt(252.0)
}

type cfPoint struct {
	t   time.Time
	amt float64
}

func calcXIRR(cfs []cfPoint) float64 {
	if len(cfs) < 2 {
		return 0.0
	}

	hasPos, hasNeg := false, false
	for _, c := range cfs {
		if c.amt > 0 {
			hasPos = true
		}
		if c.amt < 0 {
			hasNeg = true
		}
	}
	if !hasPos || !hasNeg {
		return 0.0
	}

	t0 := cfs[0].t
	years := make([]float64, len(cfs))
	amts := make([]float64, len(cfs))
	for i, c := range cfs {
		years[i] = c.t.Sub(t0).Hours() / 24.0 / 365.0
		amts[i] = c.amt
	}

	npv := func(rate float64) float64 {
		if rate <= -0.9999 {
			return math.Inf(1)
		}
		val := 0.0
		for i := 0; i < len(amts); i++ {
			val += amts[i] / math.Pow(1.0+rate, years[i])
		}
		return val
	}

	npvPrime := func(rate float64) float64 {
		if rate <= -0.9999 {
			return math.Inf(-1)
		}
		val := 0.0
		for i := 0; i < len(amts); i++ {
			val += -years[i] * amts[i] / math.Pow(1.0+rate, years[i]+1.0)
		}
		return val
	}

	rate := 0.10
	tol := 1e-6
	for iter := 0; iter < 200; iter++ {
		fVal := npv(rate)
		if math.Abs(fVal) < tol {
			return rate
		}
		fPrime := npvPrime(rate)
		if math.Abs(fPrime) < 1e-12 {
			break
		}
		newRate := rate - fVal/fPrime
		if newRate <= -0.999 {
			newRate = (rate - 0.999) / 2.0
		}
		if math.Abs(newRate-rate) < tol {
			return newRate
		}
		rate = newRate
	}

	// 二分法 fallback
	low, high := -0.99, 10.0
	fLow := npv(low)
	fHigh := npv(high)
	if fLow*fHigh > 0 {
		return 0.0
	}

	for iter := 0; iter < 100; iter++ {
		mid := (low + high) / 2.0
		fMid := npv(mid)
		if math.Abs(fMid) < tol || (high-low)/2.0 < tol {
			return mid
		}
		if fLow*fMid < 0 {
			high = mid
			fHigh = fMid
		} else {
			low = mid
			fLow = fMid
		}
	}

	return rate
}
