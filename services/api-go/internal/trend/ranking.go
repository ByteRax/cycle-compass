package trend

import (
	"fmt"
	"math"
	"sort"
)

// computeMomentumFactor computes M = 0.55*z20 + 0.45*z60 using sigma60 over at least 61 valid points.
func computeMomentumFactor(bars []DailyBar, cfg AlgorithmConfig) (*float64, string, string) {
	if len(bars) == 0 {
		return nil, "", "empty bar series"
	}
	latestDate := bars[len(bars)-1].TradingDate
	if !isValidPositive(bars[len(bars)-1].Value) {
		return nil, latestDate, "latest bar value is missing or invalid"
	}

	validValues := make([]float64, 0, len(bars))
	for _, b := range bars {
		if hasInvalidNonNull(b.Value) {
			return nil, latestDate, "non-positive or NaN value in series"
		}
		if isValidPositive(b.Value) {
			validValues = append(validValues, *b.Value)
		}
	}

	minPts := cfg.StrengthMinPoints
	if minPts <= 0 {
		minPts = 61
	}
	if len(validValues) < minPts {
		return nil, latestDate, fmt.Sprintf("insufficient valid history: got %d, need %d", len(validValues), minPts)
	}

	last := len(validValues) - 1
	w := validValues[last]

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

	z20 := 0.0
	if !(rawSigma == 0 && r20 == 0) {
		z20 = clamp(r20/(sigmaEff*math.Sqrt(20)), -4, 4)
	}
	z60 := 0.0
	if !(rawSigma == 0 && r60 == 0) {
		z60 = clamp(r60/(sigmaEff*math.Sqrt(60)), -4, 4)
	}

	m := 0.55*z20 + 0.45*z60
	return ptrFloat(m), latestDate, ""
}

// ComputeRelativeStrength computes cross-sectional 0..100 relative strength within a fixed candidate pool.
// Returns nil if inputs has fewer than 2 candidates or if no valid factors can be computed.
func ComputeRelativeStrength(inputs []StrengthInput, scopeID, scopeLabel, currency, algoVer string) []StrengthResult {
	nCandidates := len(inputs)
	if nCandidates < 2 {
		return nil
	}

	cfg := DefaultAlgorithmConfig()
	if algoVer != "" {
		cfg.AlgorithmVersion = algoVer
	}

	results := make([]StrengthResult, nCandidates)
	type validEntry struct {
		idx     int
		assetID string
		m       float64
	}
	var validEntries []validEntry

	for i, in := range inputs {
		m, srcDate, reason := computeMomentumFactor(in.Bars, cfg)
		results[i] = StrengthResult{
			AssetID:          in.AssetID,
			RawFactor:        m,
			CandidateCount:   nCandidates,
			ScopeID:          scopeID,
			ScopeLabel:       scopeLabel,
			Currency:         currency,
			AlgorithmVersion: cfg.AlgorithmVersion,
			SourceDate:       srcDate,
			Reason:           reason,
		}
		if m != nil {
			validEntries = append(validEntries, validEntry{
				idx:     i,
				assetID: in.AssetID,
				m:       *m,
			})
		}
	}

	validCount := len(validEntries)
	for i := range results {
		results[i].ValidCount = validCount
	}

	if validCount == 0 {
		return nil
	}
	if validCount < 2 {
		for _, ve := range validEntries {
			results[ve.idx].Reason = "fewer than 2 valid candidates in pool"
		}
		return results
	}

	// Sort descending by M, tie-break by AssetID ascending for stability.
	sort.Slice(validEntries, func(i, j int) bool {
		if validEntries[i].m != validEntries[j].m {
			return validEntries[i].m > validEntries[j].m
		}
		return validEntries[i].assetID < validEntries[j].assetID
	})

	// Assign 1-based average ranks for ties.
	i := 0
	for i < validCount {
		j := i + 1
		for j < validCount && validEntries[j].m == validEntries[i].m {
			j++
		}
		// 1-based positions (i+1) .. j; average rank = (float64(i+1) + float64(j)) / 2.0
		avgRank := (float64(i+1) + float64(j)) / 2.0
		rs := 100.0 * (float64(validCount) - avgRank) / float64(validCount-1)
		rs = clamp(rs, 0, 100)
		disp := math.Round(rs*10.0) / 10.0

		for k := i; k < j; k++ {
			resIdx := validEntries[k].idx
			results[resIdx].Rank = ptrFloat(avgRank)
			results[resIdx].Score = ptrFloat(rs)
			results[resIdx].DisplayScore = ptrFloat(disp)
		}
		i = j
	}

	return results
}

// ComputeTier maps an unrounded relative strength score and trend side to a five-tier echelon.
func ComputeTier(score *float64, side Side) (*Tier, string) {
	if score == nil || math.IsNaN(*score) || math.IsInf(*score, 0) {
		return nil, "missing or invalid relative strength score"
	}
	if side == SideUnknown || side == "" {
		return nil, "unknown trend side"
	}

	s := *score
	var t Tier
	switch {
	case s >= 90 && side == SideRight:
		t = TierS
	case s >= 90:
		t = TierA
	case s >= 70:
		t = TierA
	case s >= 30:
		t = TierB
	case s >= 10:
		t = TierC
	case s < 10 && side == SideLeft:
		t = TierD
	default:
		t = TierC
	}
	return &t, ""
}

// BuildTierResult returns a full TierResult with professional label and optional alias.
func BuildTierResult(score *float64, side Side) TierResult {
	t, reason := ComputeTier(score, side)
	if t == nil {
		return TierResult{Reason: reason}
	}
	var label, alias string
	switch *t {
	case TierS:
		label, alias = "S领跑", "夯"
	case TierA:
		label, alias = "A强势", "顶级"
	case TierB:
		label, alias = "B中坚", "人上人"
	case TierC:
		label, alias = "C偏弱", "NPC"
	case TierD:
		label, alias = "D落后", "拉完了"
	}
	return TierResult{
		Tier:  t,
		Label: label,
		Alias: alias,
	}
}
