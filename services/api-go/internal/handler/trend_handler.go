package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"cyclecompass/api/internal/service"
)

var (
	allowedSortFields = map[string]bool{
		"":                      true,
		"temperature_score":     true,
		"temperature":           true,
		"strength_score":        true,
		"strength":              true,
		"day_change_pct":        true,
		"change":                true,
		"duration_trading_days": true,
		"duration":              true,
		"days":                  true,
		"latest_value":          true,
		"turnover":              true,
		"name":                  true,
		"asset_id":              true,
		"wtd":                   true,
		"mtd":                   true,
		"qtd":                   true,
		"ytd":                   true,
		"prev_year":             true,
		"rolling_1m":            true,
	}
	allowedOrders = map[string]bool{
		"":     true,
		"asc":  true,
		"desc": true,
	}
	allowedRanges = map[string]bool{
		"":    true,
		"1M":  true,
		"3M":  true,
		"6M":  true,
		"1Y":  true,
		"3Y":  true,
		"5Y":  true,
		"ALL": true,
	}
	allowedSeasonDirs = map[string]bool{
		"":     true,
		"all":  true,
		"up":   true,
		"down": true,
	}
	allowedWeightBy = map[string]bool{
		"":                 true,
		"float_market_cap": true,
		"turnover":         true,
		"equal":            true,
	}
	allowedColorBy = map[string]bool{
		"":            true,
		"side":        true,
		"temperature": true,
		"change":      true,
	}
)

type TrendHandler struct {
	svc *service.TrendService
}

func NewTrendHandler(svc *service.TrendService) *TrendHandler {
	return &TrendHandler{svc: svc}
}

func parseLimitParam(r *http.Request, defaultVal int) (int, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return defaultVal, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 || v > 200 {
		return 0, false
	}
	return v, true
}

func handleServiceErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeFail(w, http.StatusNotFound, 404, "未找到对应的资产或观察池")
	case errors.Is(err, service.ErrSnapshotExpired):
		writeFail(w, http.StatusGone, 410, "请求的快照已过期或不存在，请刷新获取最新快照")
	case errors.Is(err, service.ErrStaleCursor):
		writeFail(w, http.StatusConflict, 409, "数据快照已更新，请刷新列表后重试")
	case errors.Is(err, service.ErrBadRequest):
		writeFail(w, http.StatusBadRequest, 400, err.Error())
	default:
		writeFail(w, http.StatusInternalServerError, 500, err.Error())
	}
}

func (h *TrendHandler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp, err := h.svc.GetCatalog(r.Context(), q.Get("scope_id"), q.Get("currency"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetUniverses(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetUniverses(r.Context())
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetEnums(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, h.svc.GetEnums())
}

func (h *TrendHandler) SearchAssets(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseLimitParam(r, 50)
	if !ok {
		writeFail(w, http.StatusBadRequest, 400, "limit 必须为 1 到 200 之间的整数")
		return
	}
	q := r.URL.Query()
	resp, err := h.svc.SearchAssets(r.Context(), q.Get("q"), q.Get("type"), q.Get("cursor"), limit)
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetAssetDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeFail(w, http.StatusBadRequest, 400, "缺少资产 ID")
		return
	}
	q := r.URL.Query()
	resp, err := h.svc.GetAssetDetail(r.Context(), id, q.Get("scope_id"), q.Get("currency"), q.Get("snapshot_id"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetAssetMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeFail(w, http.StatusBadRequest, 400, "缺少观察池或分类 ID")
		return
	}
	q := r.URL.Query()
	sortField := q.Get("sort")
	if !allowedSortFields[sortField] {
		writeFail(w, http.StatusBadRequest, 400, "不支持的 sort 排序字段")
		return
	}
	order := strings.ToLower(q.Get("order"))
	if !allowedOrders[order] {
		writeFail(w, http.StatusBadRequest, 400, "order 仅支持 asc 或 desc")
		return
	}
	limit, ok := parseLimitParam(r, 50)
	if !ok {
		writeFail(w, http.StatusBadRequest, 400, "limit 必须为 1 到 200 之间的整数")
		return
	}
	resp, err := h.svc.GetAssetMembers(r.Context(), service.MembersQuery{
		PoolID:          id,
		ScopeID:         q.Get("scope_id"),
		SortField:       sortField,
		Order:           order,
		Cursor:          q.Get("cursor"),
		Limit:           limit,
		Currency:        q.Get("currency"),
		SnapshotID:      q.Get("snapshot_id"),
		MarketFilter:    q.Get("market"),
		AssetTypeFilter: q.Get("asset_type"),
		StatusFilter:    q.Get("status"),
		QueryFilter:     q.Get("q"),
	})
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetLatestSnapshot(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	resp, err := h.svc.GetLatestSnapshot(r.Context(), q.Get("scope_id"), q.Get("currency"), q.Get("mode"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetTrendCurve(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := r.URL.Query()
	rangeStr := strings.ToUpper(q.Get("range"))
	if !allowedRanges[rangeStr] {
		writeFail(w, http.StatusBadRequest, 400, "不支持的 range 区间参数")
		return
	}
	resp, err := h.svc.GetTrendCurve(r.Context(), id, rangeStr, q.Get("freq"), q.Get("currency"), q.Get("benchmark_id"), q.Get("snapshot_id"), q.Get("start_date"), q.Get("end_date"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetTrendAnnual(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := r.URL.Query()
	resp, err := h.svc.GetTrendAnnual(r.Context(), id, q.Get("years"), q.Get("currency"), q.Get("snapshot_id"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetTrendComponents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := r.URL.Query()
	rangeStr := strings.ToUpper(q.Get("range"))
	if !allowedRanges[rangeStr] {
		writeFail(w, http.StatusBadRequest, 400, "不支持的 range 区间参数")
		return
	}
	sortField := q.Get("sort")
	if !allowedSortFields[sortField] {
		writeFail(w, http.StatusBadRequest, 400, "不支持的 sort 排序字段")
		return
	}
	topN := 0
	if rawTopN := strings.TrimSpace(q.Get("top_n")); rawTopN != "" {
		v, err := strconv.Atoi(rawTopN)
		if err != nil || v <= 0 || v > 200 {
			writeFail(w, http.StatusBadRequest, 400, "top_n 必须为 1 到 200 之间的整数")
			return
		}
		topN = v
	}
	resp, err := h.svc.GetTrendComponents(r.Context(), id, q.Get("scope_id"), rangeStr, topN, sortField, q.Get("currency"), q.Get("snapshot_id"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetTrendDashboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := r.URL.Query()
	sortField := q.Get("sort")
	if !allowedSortFields[sortField] {
		writeFail(w, http.StatusBadRequest, 400, "不支持的 sort 排序字段")
		return
	}
	order := strings.ToLower(q.Get("order"))
	if !allowedOrders[order] {
		writeFail(w, http.StatusBadRequest, 400, "order 仅支持 asc 或 desc")
		return
	}
	sparkRange := strings.ToUpper(q.Get("spark_range"))
	if !allowedRanges[sparkRange] {
		writeFail(w, http.StatusBadRequest, 400, "不支持的 spark_range 区间参数")
		return
	}
	resp, err := h.svc.GetTrendDashboard(r.Context(), id, q.Get("change_period"), sparkRange, sortField, order, q.Get("scope_id"), q.Get("group"), q.Get("currency"), q.Get("snapshot_id"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetTrendSeason(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := r.URL.Query()
	dir := strings.ToLower(q.Get("direction"))
	if !allowedSeasonDirs[dir] {
		writeFail(w, http.StatusBadRequest, 400, "direction 仅支持 all、up 或 down")
		return
	}
	resp, err := h.svc.GetTrendSeason(r.Context(), id, dir, q.Get("snapshot_id"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetTrendDistribution(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := r.URL.Query()
	weightBy := q.Get("weight_by")
	if !allowedWeightBy[weightBy] {
		writeFail(w, http.StatusBadRequest, 400, "不支持的 weight_by 参数")
		return
	}
	colorBy := q.Get("color_by")
	if !allowedColorBy[colorBy] {
		writeFail(w, http.StatusBadRequest, 400, "不支持的 color_by 参数")
		return
	}
	resp, err := h.svc.GetTrendDistribution(r.Context(), id, weightBy, colorBy, q.Get("snapshot_id"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetTrendTier(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	q := r.URL.Query()
	resp, err := h.svc.GetTrendTier(r.Context(), id, q.Get("scope_id"), q.Get("snapshot_id"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetTrendEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit, ok := parseLimitParam(r, 50)
	if !ok {
		writeFail(w, http.StatusBadRequest, 400, "limit 必须为 1 到 200 之间的整数")
		return
	}
	q := r.URL.Query()
	resp, err := h.svc.GetTrendEvents(r.Context(), id, q.Get("types"), limit, q.Get("cursor"), q.Get("snapshot_id"))
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetMethodology(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, h.svc.GetMethodology(r.Context()))
}

func (h *TrendHandler) GetDataStatus(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetDataStatus(r.Context())
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}

func (h *TrendHandler) GetPeriodMatrix(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeFail(w, http.StatusBadRequest, 400, "缺少观察池或分类 ID")
		return
	}
	q := r.URL.Query()
	sortField := q.Get("sort")
	if !allowedSortFields[sortField] {
		writeFail(w, http.StatusBadRequest, 400, "不支持的 sort 排序字段")
		return
	}
	order := strings.ToLower(q.Get("order"))
	if !allowedOrders[order] {
		writeFail(w, http.StatusBadRequest, 400, "order 仅支持 asc 或 desc")
		return
	}
	limit, ok := parseLimitParam(r, 50)
	if !ok {
		writeFail(w, http.StatusBadRequest, 400, "limit 必须为 1 到 200 之间的整数")
		return
	}
	resp, err := h.svc.GetPeriodMatrix(r.Context(), service.PeriodMatrixQuery{
		PoolID:          id,
		ScopeID:         q.Get("scope_id"),
		SnapshotID:      q.Get("snapshot_id"),
		Currency:        q.Get("currency"),
		SortField:       sortField,
		Order:           order,
		Cursor:          q.Get("cursor"),
		Limit:           limit,
		MarketFilter:    q.Get("market"),
		AssetTypeFilter: q.Get("asset_type"),
		StatusFilter:    q.Get("status"),
		QueryFilter:     q.Get("q"),
	})
	if err != nil {
		handleServiceErr(w, err)
		return
	}
	writeSuccess(w, resp)
}
