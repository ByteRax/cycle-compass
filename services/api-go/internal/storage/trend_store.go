package storage

import (
	"database/sql"
	"fmt"
	"strings"
)

type AssetEntity struct {
	AssetID       string
	Market        string
	Symbol        string
	Name          string
	ShortName     string
	AssetType     string
	MeasureType   string
	PriceType     string
	Unit          string
	Currency      string
	Timezone      string
	IsActive      bool
	LicenseStatus string
	SourceName    string
	CreatedAt     string
	UpdatedAt     string
}

type AssetCategoryEntity struct {
	CategoryID   string
	ParentID     *string
	Name         string
	Level        int
	SortOrder    int
	HasIndex     bool
	IndexAssetID *string
}

type AssetMembershipEntity struct {
	ID         int64
	CategoryID string
	AssetID    string
	Weight     *float64
	ValidFrom  string
	ValidTo    *string
}

type UniverseEntity struct {
	UniverseID  string
	Name        string
	Description string
	IsPublic    bool
	SortOrder   int
}

type UniverseMemberEntity struct {
	UniverseID string
	AssetID    string
	SortOrder  int
}

type DailyBarEntity struct {
	AssetID          string
	TradingDate      string
	Provider         string
	Revision         int
	Open             *float64
	High             *float64
	Low              *float64
	Close            *float64
	AdjustedClose    *float64
	TotalReturnIndex *float64
	Volume           *float64
	Turnover         *float64
	FloatMarketCap   *float64
	FundAUM          *float64
	DividendYield    *float64
	YieldPct         *float64
	QualityStatus    string
	AsOf             string
	AvailableAt      string
	UpdatedAt        string
}

type TrendSnapshotEntity struct {
	SnapshotID       string
	ScopeID          string
	Currency         string
	Mode             string
	AsOf             string
	ComputedAt       string
	AlgorithmVersion string
	CandidateCount   int
	ValidCount       int
	Coverage         *float64
	Status           string
	IsCurrent        bool
	MetaJSON         string
}

type TrendDailyMetricEntity struct {
	SnapshotID           string
	AssetID              string
	TradingDate          string
	Temperature          *string
	TemperatureScore     *float64
	Direction            string
	Confirmation         string
	Side                 string
	StartedAt            *string
	ConfirmedAt          *string
	DurationTradingDays  *int
	DurationCalendarDays *int
	Phase                *string
	Tier                 *string
	Reason               *string
}

type StrengthDailyRankingEntity struct {
	SnapshotID     string
	ScopeID        string
	Currency       string
	AssetID        string
	RawFactor      *float64
	Rank           *int
	Score          *float64
	Delta          *float64
	CandidateCount int
	ValidCount     int
	Reason         *string
}

type TrendEventEntity struct {
	EventID          string
	AssetID          string
	EventType        string
	OldValue         *string
	NewValue         *string
	AsOf             string
	DetectedAt       string
	SnapshotID       string
	AlgorithmVersion string
	ScopeID          string
	DedupeKey        string
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullStr(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}

func strPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func nullFloat(p *float64) sql.NullFloat64 {
	if p == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *p, Valid: true}
}

func floatPtr(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	v := nf.Float64
	return &v
}

func nullInt(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}

func intPtr(ni sql.NullInt64) *int {
	if !ni.Valid {
		return nil
	}
	v := int(ni.Int64)
	return &v
}

func (s *SQLiteStore) execInTx(fn func(tx *sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLiteStore) UpsertAssets(assets []AssetEntity) error {
	if len(assets) == 0 {
		return nil
	}
	return s.execInTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
			INSERT INTO assets (
				asset_id, market, symbol, name, short_name, asset_type,
				measure_type, price_type, unit, currency, timezone, is_active,
				license_status, source_name, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(asset_id) DO UPDATE SET
				market = excluded.market,
				symbol = excluded.symbol,
				name = excluded.name,
				short_name = excluded.short_name,
				asset_type = excluded.asset_type,
				measure_type = excluded.measure_type,
				price_type = excluded.price_type,
				unit = excluded.unit,
				currency = excluded.currency,
				timezone = excluded.timezone,
				is_active = excluded.is_active,
				license_status = excluded.license_status,
				source_name = excluded.source_name,
				updated_at = excluded.updated_at
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, a := range assets {
			if _, err := stmt.Exec(
				a.AssetID, a.Market, a.Symbol, a.Name, a.ShortName, a.AssetType,
				a.MeasureType, a.PriceType, a.Unit, a.Currency, a.Timezone, boolToInt(a.IsActive),
				a.LicenseStatus, a.SourceName, a.CreatedAt, a.UpdatedAt,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLiteStore) UpsertAssetCategories(categories []AssetCategoryEntity) error {
	if len(categories) == 0 {
		return nil
	}
	return s.execInTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
			INSERT INTO asset_categories (
				category_id, parent_id, name, level, sort_order, has_index, index_asset_id
			) VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(category_id) DO UPDATE SET
				parent_id = excluded.parent_id,
				name = excluded.name,
				level = excluded.level,
				sort_order = excluded.sort_order,
				has_index = excluded.has_index,
				index_asset_id = excluded.index_asset_id
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, c := range categories {
			if _, err := stmt.Exec(
				c.CategoryID, nullStr(c.ParentID), c.Name, c.Level, c.SortOrder,
				boolToInt(c.HasIndex), nullStr(c.IndexAssetID),
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLiteStore) UpsertAssetMemberships(memberships []AssetMembershipEntity) error {
	if len(memberships) == 0 {
		return nil
	}
	return s.execInTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
			INSERT INTO asset_memberships (
				category_id, asset_id, weight, valid_from, valid_to
			) VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(category_id, asset_id, valid_from) DO UPDATE SET
				weight = excluded.weight,
				valid_to = excluded.valid_to
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, m := range memberships {
			if _, err := stmt.Exec(
				m.CategoryID, m.AssetID, nullFloat(m.Weight), m.ValidFrom, nullStr(m.ValidTo),
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLiteStore) UpsertUniverses(universes []UniverseEntity) error {
	if len(universes) == 0 {
		return nil
	}
	return s.execInTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
			INSERT INTO universes (
				universe_id, name, description, is_public, sort_order
			) VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(universe_id) DO UPDATE SET
				name = excluded.name,
				description = excluded.description,
				is_public = excluded.is_public,
				sort_order = excluded.sort_order
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, u := range universes {
			if _, err := stmt.Exec(
				u.UniverseID, u.Name, u.Description, boolToInt(u.IsPublic), u.SortOrder,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLiteStore) UpsertUniverseMembers(members []UniverseMemberEntity) error {
	if len(members) == 0 {
		return nil
	}
	return s.execInTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
			INSERT INTO universe_members (universe_id, asset_id, sort_order)
			VALUES (?, ?, ?)
			ON CONFLICT(universe_id, asset_id) DO UPDATE SET
				sort_order = excluded.sort_order
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, m := range members {
			if _, err := stmt.Exec(m.UniverseID, m.AssetID, m.SortOrder); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLiteStore) UpsertDailyBars(bars []DailyBarEntity) error {
	if len(bars) == 0 {
		return nil
	}
	return s.execInTx(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(`
			INSERT INTO daily_bars (
				asset_id, trading_date, provider, revision,
				open, high, low, close, adjusted_close, total_return_index,
				volume, turnover, float_market_cap, fund_aum, dividend_yield, yield_pct,
				quality_status, as_of, available_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(asset_id, trading_date, provider, revision) DO UPDATE SET
				open = excluded.open,
				high = excluded.high,
				low = excluded.low,
				close = excluded.close,
				adjusted_close = excluded.adjusted_close,
				total_return_index = excluded.total_return_index,
				volume = excluded.volume,
				turnover = excluded.turnover,
				float_market_cap = excluded.float_market_cap,
				fund_aum = excluded.fund_aum,
				dividend_yield = excluded.dividend_yield,
				yield_pct = excluded.yield_pct,
				quality_status = excluded.quality_status,
				as_of = excluded.as_of,
				available_at = excluded.available_at,
				updated_at = excluded.updated_at
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, b := range bars {
			if _, err := stmt.Exec(
				b.AssetID, b.TradingDate, b.Provider, b.Revision,
				nullFloat(b.Open), nullFloat(b.High), nullFloat(b.Low), nullFloat(b.Close),
				nullFloat(b.AdjustedClose), nullFloat(b.TotalReturnIndex),
				nullFloat(b.Volume), nullFloat(b.Turnover), nullFloat(b.FloatMarketCap),
				nullFloat(b.FundAUM), nullFloat(b.DividendYield), nullFloat(b.YieldPct),
				b.QualityStatus, b.AsOf, b.AvailableAt, b.UpdatedAt,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func upsertTrendDailyMetricsTx(tx *sql.Tx, metrics []TrendDailyMetricEntity) error {
	if len(metrics) == 0 {
		return nil
	}
	stmt, err := tx.Prepare(`
		INSERT INTO trend_daily_metrics (
			snapshot_id, asset_id, trading_date, temperature, temperature_score,
			direction, confirmation, side, started_at, confirmed_at,
			duration_trading_days, duration_calendar_days, phase, tier, reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(snapshot_id, asset_id) DO UPDATE SET
			trading_date = excluded.trading_date,
			temperature = excluded.temperature,
			temperature_score = excluded.temperature_score,
			direction = excluded.direction,
			confirmation = excluded.confirmation,
			side = excluded.side,
			started_at = excluded.started_at,
			confirmed_at = excluded.confirmed_at,
			duration_trading_days = excluded.duration_trading_days,
			duration_calendar_days = excluded.duration_calendar_days,
			phase = excluded.phase,
			tier = excluded.tier,
			reason = excluded.reason
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range metrics {
		if _, err := stmt.Exec(
			m.SnapshotID, m.AssetID, m.TradingDate,
			nullStr(m.Temperature), nullFloat(m.TemperatureScore),
			m.Direction, m.Confirmation, m.Side,
			nullStr(m.StartedAt), nullStr(m.ConfirmedAt),
			nullInt(m.DurationTradingDays), nullInt(m.DurationCalendarDays),
			nullStr(m.Phase), nullStr(m.Tier), nullStr(m.Reason),
		); err != nil {
			return err
		}
	}
	return nil
}

func upsertStrengthDailyRankingsTx(tx *sql.Tx, rankings []StrengthDailyRankingEntity) error {
	if len(rankings) == 0 {
		return nil
	}
	stmt, err := tx.Prepare(`
		INSERT INTO strength_daily_rankings (
			snapshot_id, scope_id, currency, asset_id,
			raw_factor, rank, score, delta, candidate_count, valid_count, reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(snapshot_id, scope_id, currency, asset_id) DO UPDATE SET
			raw_factor = excluded.raw_factor,
			rank = excluded.rank,
			score = excluded.score,
			delta = excluded.delta,
			candidate_count = excluded.candidate_count,
			valid_count = excluded.valid_count,
			reason = excluded.reason
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range rankings {
		if _, err := stmt.Exec(
			r.SnapshotID, r.ScopeID, r.Currency, r.AssetID,
			nullFloat(r.RawFactor), nullInt(r.Rank), nullFloat(r.Score), nullFloat(r.Delta),
			r.CandidateCount, r.ValidCount, nullStr(r.Reason),
		); err != nil {
			return err
		}
	}
	return nil
}

func insertTrendEventsTx(tx *sql.Tx, events []TrendEventEntity) error {
	if len(events) == 0 {
		return nil
	}
	stmt, err := tx.Prepare(`
		INSERT INTO trend_events (
			event_id, asset_id, event_type, old_value, new_value,
			as_of, detected_at, snapshot_id, algorithm_version, scope_id, dedupe_key
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		if _, err := stmt.Exec(
			e.EventID, e.AssetID, e.EventType, nullStr(e.OldValue), nullStr(e.NewValue),
			e.AsOf, e.DetectedAt, e.SnapshotID, e.AlgorithmVersion, e.ScopeID, e.DedupeKey,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) UpsertTrendDailyMetrics(metrics []TrendDailyMetricEntity) error {
	if len(metrics) == 0 {
		return nil
	}
	return s.execInTx(func(tx *sql.Tx) error {
		return upsertTrendDailyMetricsTx(tx, metrics)
	})
}

func (s *SQLiteStore) UpsertStrengthDailyRankings(rankings []StrengthDailyRankingEntity) error {
	if len(rankings) == 0 {
		return nil
	}
	return s.execInTx(func(tx *sql.Tx) error {
		return upsertStrengthDailyRankingsTx(tx, rankings)
	})
}

func (s *SQLiteStore) InsertTrendEvents(events []TrendEventEntity) error {
	if len(events) == 0 {
		return nil
	}
	if err := s.execInTx(func(tx *sql.Tx) error {
		return insertTrendEventsTx(tx, events)
	}); err != nil {
		return err
	}
	return nil
}

// PublishTrendSnapshotWithFrozen atomically publishes snapshot + frozen per-asset bar revisions.
func (s *SQLiteStore) PublishTrendSnapshotWithFrozen(
	snapshot TrendSnapshotEntity,
	metrics []TrendDailyMetricEntity,
	rankings []StrengthDailyRankingEntity,
	events []TrendEventEntity,
	barsByAsset map[string][]DailyBarEntity,
) error {
	return s.execInTx(func(tx *sql.Tx) error {
		if err := publishSnapshotCoreTx(tx, snapshot, metrics, rankings, events); err != nil {
			return err
		}
		if len(barsByAsset) == 0 {
			return nil
		}
		stmt, err := tx.Prepare(`
			INSERT INTO trend_snapshot_frozen_bars (
				snapshot_id, asset_id, trading_date, provider, revision,
				open, high, low, close, adjusted_close, total_return_index,
				volume, turnover, dividend_yield, yield_pct, quality_status, as_of, available_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(snapshot_id, asset_id, trading_date) DO UPDATE SET
				provider = excluded.provider,
				revision = excluded.revision,
				open = excluded.open,
				high = excluded.high,
				low = excluded.low,
				close = excluded.close,
				adjusted_close = excluded.adjusted_close,
				total_return_index = excluded.total_return_index,
				volume = excluded.volume,
				turnover = excluded.turnover,
				dividend_yield = excluded.dividend_yield,
				yield_pct = excluded.yield_pct,
				quality_status = excluded.quality_status,
				as_of = excluded.as_of,
				available_at = excluded.available_at,
				updated_at = excluded.updated_at
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for assetID, list := range barsByAsset {
			for _, b := range list {
				if _, err := stmt.Exec(
					snapshot.SnapshotID, assetID, b.TradingDate, b.Provider, b.Revision,
					nullFloat(b.Open), nullFloat(b.High), nullFloat(b.Low), nullFloat(b.Close),
					nullFloat(b.AdjustedClose), nullFloat(b.TotalReturnIndex),
					nullFloat(b.Volume), nullFloat(b.Turnover), nullFloat(b.DividendYield), nullFloat(b.YieldPct),
					b.QualityStatus, b.AsOf, b.AvailableAt, b.UpdatedAt,
				); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func publishSnapshotCoreTx(
	tx *sql.Tx,
	snapshot TrendSnapshotEntity,
	metrics []TrendDailyMetricEntity,
	rankings []StrengthDailyRankingEntity,
	events []TrendEventEntity,
) error {
	if snapshot.IsCurrent {
		if _, err := tx.Exec(
			`UPDATE trend_snapshots SET is_current = 0
			 WHERE scope_id = ? AND currency = ? AND mode = ? AND is_current = 1`,
			snapshot.ScopeID, snapshot.Currency, snapshot.Mode,
		); err != nil {
			return err
		}
	}

	meta := snapshot.MetaJSON
	if meta == "" {
		meta = "{}"
	}
	if _, err := tx.Exec(`
		INSERT INTO trend_snapshots (
			snapshot_id, scope_id, currency, mode, as_of, computed_at,
			algorithm_version, candidate_count, valid_count, coverage,
			status, is_current, meta_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(snapshot_id) DO UPDATE SET
			scope_id = excluded.scope_id,
			currency = excluded.currency,
			mode = excluded.mode,
			as_of = excluded.as_of,
			computed_at = excluded.computed_at,
			algorithm_version = excluded.algorithm_version,
			candidate_count = excluded.candidate_count,
			valid_count = excluded.valid_count,
			coverage = excluded.coverage,
			status = excluded.status,
			is_current = excluded.is_current,
			meta_json = excluded.meta_json
	`,
		snapshot.SnapshotID, snapshot.ScopeID, snapshot.Currency, snapshot.Mode,
		snapshot.AsOf, snapshot.ComputedAt, snapshot.AlgorithmVersion,
		snapshot.CandidateCount, snapshot.ValidCount, nullFloat(snapshot.Coverage),
		snapshot.Status, boolToInt(snapshot.IsCurrent), meta,
	); err != nil {
		return err
	}

	if err := upsertTrendDailyMetricsTx(tx, metrics); err != nil {
		return err
	}
	if err := upsertStrengthDailyRankingsTx(tx, rankings); err != nil {
		return err
	}
	if err := insertTrendEventsTx(tx, events); err != nil {
		return err
	}
	return nil
}

// GetFrozenSnapshotBars returns frozen daily bars captured when snapshotID was published.
func (s *SQLiteStore) GetFrozenSnapshotBars(snapshotID, assetID string) ([]DailyBarEntity, error) {
	rows, err := s.db.Query(`
		SELECT asset_id, trading_date, provider, revision,
		       open, high, low, close, adjusted_close, total_return_index,
		       volume, turnover, dividend_yield, yield_pct,
		       quality_status, as_of, available_at, updated_at
		FROM trend_snapshot_frozen_bars
		WHERE snapshot_id = ? AND asset_id = ?
		ORDER BY trading_date ASC`, snapshotID, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DailyBarEntity
	for rows.Next() {
		var b DailyBarEntity
		var open, high, low, closeP, adjClose, tri sql.NullFloat64
		var vol, turnover, divYield, yieldPct sql.NullFloat64
		if err := rows.Scan(
			&b.AssetID, &b.TradingDate, &b.Provider, &b.Revision,
			&open, &high, &low, &closeP, &adjClose, &tri,
			&vol, &turnover, &divYield, &yieldPct,
			&b.QualityStatus, &b.AsOf, &b.AvailableAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		b.Open = floatPtr(open)
		b.High = floatPtr(high)
		b.Low = floatPtr(low)
		b.Close = floatPtr(closeP)
		b.AdjustedClose = floatPtr(adjClose)
		b.TotalReturnIndex = floatPtr(tri)
		b.Volume = floatPtr(vol)
		b.Turnover = floatPtr(turnover)
		b.DividendYield = floatPtr(divYield)
		b.YieldPct = floatPtr(yieldPct)
		list = append(list, b)
	}
	return list, rows.Err()
}

// PublishTrendSnapshot atomically writes a snapshot, its metrics/rankings/events,
// and switches the is_current pointer for (scope_id, currency, mode) in a single transaction.
func (s *SQLiteStore) PublishTrendSnapshot(
	snapshot TrendSnapshotEntity,
	metrics []TrendDailyMetricEntity,
	rankings []StrengthDailyRankingEntity,
	events []TrendEventEntity,
) error {
	return s.execInTx(func(tx *sql.Tx) error {
		return publishSnapshotCoreTx(tx, snapshot, metrics, rankings, events)
	})
}

// SetCurrentSnapshot atomically switches is_current to the given snapshotID
// within its (scope_id, currency, mode) partition.
func (s *SQLiteStore) SetCurrentSnapshot(snapshotID string) error {
	return s.execInTx(func(tx *sql.Tx) error {
		var scopeID, currency, mode string
		err := tx.QueryRow(
			`SELECT scope_id, currency, mode FROM trend_snapshots WHERE snapshot_id = ?`,
			snapshotID,
		).Scan(&scopeID, &currency, &mode)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			`UPDATE trend_snapshots SET is_current = 0
			 WHERE scope_id = ? AND currency = ? AND mode = ? AND is_current = 1`,
			scopeID, currency, mode,
		); err != nil {
			return err
		}

		res, err := tx.Exec(
			`UPDATE trend_snapshots SET is_current = 1 WHERE snapshot_id = ?`,
			snapshotID,
		)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
}

func scanAsset(scanner interface{ Scan(...any) error }) (AssetEntity, error) {
	var a AssetEntity
	var isActive int
	err := scanner.Scan(
		&a.AssetID, &a.Market, &a.Symbol, &a.Name, &a.ShortName, &a.AssetType,
		&a.MeasureType, &a.PriceType, &a.Unit, &a.Currency, &a.Timezone, &isActive,
		&a.LicenseStatus, &a.SourceName, &a.CreatedAt, &a.UpdatedAt,
	)
	a.IsActive = isActive != 0
	return a, err
}

func (s *SQLiteStore) GetAssetByID(assetID string) (*AssetEntity, error) {
	row := s.db.QueryRow(`
		SELECT asset_id, market, symbol, name, short_name, asset_type,
		       measure_type, price_type, unit, currency, timezone, is_active,
		       license_status, source_name, created_at, updated_at
		FROM assets WHERE asset_id = ?`, assetID,
	)
	a, err := scanAsset(row)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *SQLiteStore) ListAssets(activeOnly bool) ([]AssetEntity, error) {
	query := `
		SELECT asset_id, market, symbol, name, short_name, asset_type,
		       measure_type, price_type, unit, currency, timezone, is_active,
		       license_status, source_name, created_at, updated_at
		FROM assets`
	if activeOnly {
		query += ` WHERE is_active = 1`
	}
	query += ` ORDER BY asset_id ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AssetEntity
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *SQLiteStore) ListAssetCategories() ([]AssetCategoryEntity, error) {
	rows, err := s.db.Query(`
		SELECT category_id, parent_id, name, level, sort_order, has_index, index_asset_id
		FROM asset_categories
		ORDER BY level ASC, sort_order ASC, category_id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AssetCategoryEntity
	for rows.Next() {
		var c AssetCategoryEntity
		var parentID, indexAssetID sql.NullString
		var hasIndex int
		if err := rows.Scan(
			&c.CategoryID, &parentID, &c.Name, &c.Level, &c.SortOrder, &hasIndex, &indexAssetID,
		); err != nil {
			return nil, err
		}
		c.ParentID = strPtr(parentID)
		c.HasIndex = hasIndex != 0
		c.IndexAssetID = strPtr(indexAssetID)
		list = append(list, c)
	}
	return list, rows.Err()
}

func (s *SQLiteStore) ListUniverses(publicOnly bool) ([]UniverseEntity, error) {
	query := `SELECT universe_id, name, description, is_public, sort_order FROM universes`
	if publicOnly {
		query += ` WHERE is_public = 1`
	}
	query += ` ORDER BY sort_order ASC, universe_id ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []UniverseEntity
	for rows.Next() {
		var u UniverseEntity
		var isPublic int
		if err := rows.Scan(&u.UniverseID, &u.Name, &u.Description, &isPublic, &u.SortOrder); err != nil {
			return nil, err
		}
		u.IsPublic = isPublic != 0
		list = append(list, u)
	}
	return list, rows.Err()
}

func (s *SQLiteStore) ListUniverseMembers(universeID string) ([]AssetEntity, error) {
	rows, err := s.db.Query(`
		SELECT a.asset_id, a.market, a.symbol, a.name, a.short_name, a.asset_type,
		       a.measure_type, a.price_type, a.unit, a.currency, a.timezone, a.is_active,
		       a.license_status, a.source_name, a.created_at, a.updated_at
		FROM universe_members um
		INNER JOIN assets a ON a.asset_id = um.asset_id
		WHERE um.universe_id = ?
		ORDER BY um.sort_order ASC, a.asset_id ASC
	`, universeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AssetEntity
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *SQLiteStore) GetDailyBars(assetID, fromDate, toDate string, limit int) ([]DailyBarEntity, error) {
	var args []any
	var conds []string
	conds = append(conds, "asset_id = ?")
	args = append(args, assetID)

	if fromDate != "" {
		conds = append(conds, "trading_date >= ?")
		args = append(args, fromDate)
	}
	if toDate != "" {
		conds = append(conds, "trading_date <= ?")
		args = append(args, toDate)
	}

	query := fmt.Sprintf(`
		SELECT asset_id, trading_date, provider, revision,
		       open, high, low, close, adjusted_close, total_return_index,
		       volume, turnover, float_market_cap, fund_aum, dividend_yield, yield_pct,
		       quality_status, as_of, available_at, updated_at
		FROM daily_bars
		WHERE %s
		ORDER BY trading_date ASC, provider ASC, revision DESC`, strings.Join(conds, " AND "))

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DailyBarEntity
	for rows.Next() {
		var b DailyBarEntity
		var open, high, low, closeP, adjClose, tri sql.NullFloat64
		var vol, turnover, fmc, aum, divYield, yieldPct sql.NullFloat64
		if err := rows.Scan(
			&b.AssetID, &b.TradingDate, &b.Provider, &b.Revision,
			&open, &high, &low, &closeP, &adjClose, &tri,
			&vol, &turnover, &fmc, &aum, &divYield, &yieldPct,
			&b.QualityStatus, &b.AsOf, &b.AvailableAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		b.Open = floatPtr(open)
		b.High = floatPtr(high)
		b.Low = floatPtr(low)
		b.Close = floatPtr(closeP)
		b.AdjustedClose = floatPtr(adjClose)
		b.TotalReturnIndex = floatPtr(tri)
		b.Volume = floatPtr(vol)
		b.Turnover = floatPtr(turnover)
		b.FloatMarketCap = floatPtr(fmc)
		b.FundAUM = floatPtr(aum)
		b.DividendYield = floatPtr(divYield)
		b.YieldPct = floatPtr(yieldPct)
		list = append(list, b)
	}
	return list, rows.Err()
}

// ListDailyBarsForAssets 批量按资产 ID 列表与截止日查询日线，按资产与交易日升序去重，避免 N+1 查询。
func (s *SQLiteStore) ListDailyBarsForAssets(assetIDs []string, toDate string) (map[string][]DailyBarEntity, error) {
	result := make(map[string][]DailyBarEntity, len(assetIDs))
	if len(assetIDs) == 0 {
		return result, nil
	}

	args := make([]any, 0, len(assetIDs)+1)
	placeholders := make([]string, len(assetIDs))
	for i, id := range assetIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	conds := []string{fmt.Sprintf("asset_id IN (%s)", strings.Join(placeholders, ","))}
	if toDate != "" {
		conds = append(conds, "trading_date <= ?")
		args = append(args, toDate)
	}

	query := fmt.Sprintf(`
		SELECT asset_id, trading_date, provider, revision,
		       open, high, low, close, adjusted_close, total_return_index,
		       volume, turnover, float_market_cap, fund_aum, dividend_yield, yield_pct,
		       quality_status, as_of, available_at, updated_at
		FROM daily_bars
		WHERE %s
		ORDER BY asset_id ASC, trading_date ASC, provider ASC, revision DESC`, strings.Join(conds, " AND "))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lastDateByAsset := make(map[string]string, len(assetIDs))
	for rows.Next() {
		var b DailyBarEntity
		var open, high, low, closeP, adjClose, tri sql.NullFloat64
		var vol, turnover, fmc, aum, divYield, yieldPct sql.NullFloat64
		if err := rows.Scan(
			&b.AssetID, &b.TradingDate, &b.Provider, &b.Revision,
			&open, &high, &low, &closeP, &adjClose, &tri,
			&vol, &turnover, &fmc, &aum, &divYield, &yieldPct,
			&b.QualityStatus, &b.AsOf, &b.AvailableAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if lastDateByAsset[b.AssetID] == b.TradingDate {
			continue
		}
		lastDateByAsset[b.AssetID] = b.TradingDate
		b.Open = floatPtr(open)
		b.High = floatPtr(high)
		b.Low = floatPtr(low)
		b.Close = floatPtr(closeP)
		b.AdjustedClose = floatPtr(adjClose)
		b.TotalReturnIndex = floatPtr(tri)
		b.Volume = floatPtr(vol)
		b.Turnover = floatPtr(turnover)
		b.FloatMarketCap = floatPtr(fmc)
		b.FundAUM = floatPtr(aum)
		b.DividendYield = floatPtr(divYield)
		b.YieldPct = floatPtr(yieldPct)
		result[b.AssetID] = append(result[b.AssetID], b)
	}
	return result, rows.Err()
}

type FXRateEntity struct {
	Pair        string
	TradingDate string
	Rate        float64
	Provider    string
	UpdatedAt   string
}

// ListFXRates 批量查询汇率表序列（按 pair 与交易日升序）。
func (s *SQLiteStore) ListFXRates(pairs []string, toDate string) (map[string][]FXRateEntity, error) {
	result := make(map[string][]FXRateEntity)
	var conds []string
	var args []any
	if len(pairs) > 0 {
		placeholders := make([]string, len(pairs))
		for i, p := range pairs {
			placeholders[i] = "?"
			args = append(args, p)
		}
		conds = append(conds, fmt.Sprintf("pair IN (%s)", strings.Join(placeholders, ",")))
	}
	if toDate != "" {
		conds = append(conds, "trading_date <= ?")
		args = append(args, toDate)
	}
	query := `SELECT pair, trading_date, rate, provider, updated_at FROM fx_rates`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY pair ASC, trading_date ASC, provider ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lastDate := make(map[string]string)
	for rows.Next() {
		var f FXRateEntity
		if err := rows.Scan(&f.Pair, &f.TradingDate, &f.Rate, &f.Provider, &f.UpdatedAt); err != nil {
			return nil, err
		}
		if lastDate[f.Pair] == f.TradingDate {
			continue
		}
		lastDate[f.Pair] = f.TradingDate
		result[f.Pair] = append(result[f.Pair], f)
	}
	return result, rows.Err()
}

func scanSnapshot(scanner interface{ Scan(...any) error }) (*TrendSnapshotEntity, error) {
	var sn TrendSnapshotEntity
	var cov sql.NullFloat64
	var isCurrent int
	if err := scanner.Scan(
		&sn.SnapshotID, &sn.ScopeID, &sn.Currency, &sn.Mode, &sn.AsOf, &sn.ComputedAt,
		&sn.AlgorithmVersion, &sn.CandidateCount, &sn.ValidCount, &cov,
		&sn.Status, &isCurrent, &sn.MetaJSON,
	); err != nil {
		return nil, err
	}
	sn.Coverage = floatPtr(cov)
	sn.IsCurrent = isCurrent != 0
	return &sn, nil
}

func (s *SQLiteStore) GetCurrentSnapshot(scopeID, currency, mode string) (*TrendSnapshotEntity, error) {
	row := s.db.QueryRow(`
		SELECT snapshot_id, scope_id, currency, mode, as_of, computed_at,
		       algorithm_version, candidate_count, valid_count, coverage,
		       status, is_current, meta_json
		FROM trend_snapshots
		WHERE scope_id = ? AND currency = ? AND mode = ? AND is_current = 1
		LIMIT 1`, scopeID, currency, mode,
	)
	return scanSnapshot(row)
}

func (s *SQLiteStore) GetSnapshotByID(snapshotID string) (*TrendSnapshotEntity, error) {
	row := s.db.QueryRow(`
		SELECT snapshot_id, scope_id, currency, mode, as_of, computed_at,
		       algorithm_version, candidate_count, valid_count, coverage,
		       status, is_current, meta_json
		FROM trend_snapshots
		WHERE snapshot_id = ?`, snapshotID,
	)
	return scanSnapshot(row)
}

func (s *SQLiteStore) GetSnapshotByAsOf(scopeID, currency, mode, asOf string) (*TrendSnapshotEntity, error) {
	row := s.db.QueryRow(`
		SELECT snapshot_id, scope_id, currency, mode, as_of, computed_at,
		       algorithm_version, candidate_count, valid_count, coverage,
		       status, is_current, meta_json
		FROM trend_snapshots
		WHERE scope_id = ? AND currency = ? AND mode = ? AND as_of = ?
		ORDER BY is_current DESC, computed_at DESC
		LIMIT 1`, scopeID, currency, mode, asOf,
	)
	return scanSnapshot(row)
}

// PruneOldSnapshots keeps the latest keepCount snapshots per (scope_id, currency, mode) and removes older non-current ones.
func (s *SQLiteStore) PruneOldSnapshots(keepCount int) error {
	if keepCount <= 0 {
		return nil
	}
	return s.execInTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(`
			SELECT snapshot_id FROM trend_snapshots
			WHERE is_current = 0
			ORDER BY as_of DESC, computed_at DESC
			LIMIT -1 OFFSET ?`, keepCount)
		if err != nil {
			return err
		}
		var staleIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err == nil && id != "" {
				staleIDs = append(staleIDs, id)
			}
		}
		rows.Close()
		for _, id := range staleIDs {
			_, _ = tx.Exec(`DELETE FROM trend_snapshot_frozen_bars WHERE snapshot_id = ?`, id)
			_, _ = tx.Exec(`DELETE FROM trend_daily_metrics WHERE snapshot_id = ?`, id)
			_, _ = tx.Exec(`DELETE FROM strength_daily_rankings WHERE snapshot_id = ?`, id)
			_, _ = tx.Exec(`DELETE FROM trend_snapshots WHERE snapshot_id = ? AND is_current = 0`, id)
		}
		return nil
	})
}

func (s *SQLiteStore) ListTrendDailyMetrics(snapshotID string, assetIDs []string) ([]TrendDailyMetricEntity, error) {
	args := []any{snapshotID}
	query := `
		SELECT snapshot_id, asset_id, trading_date, temperature, temperature_score,
		       direction, confirmation, side, started_at, confirmed_at,
		       duration_trading_days, duration_calendar_days, phase, tier, reason
		FROM trend_daily_metrics
		WHERE snapshot_id = ?`

	if len(assetIDs) > 0 {
		placeholders := make([]string, len(assetIDs))
		for i, id := range assetIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += fmt.Sprintf(" AND asset_id IN (%s)", strings.Join(placeholders, ","))
	}
	query += " ORDER BY asset_id ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []TrendDailyMetricEntity
	for rows.Next() {
		var m TrendDailyMetricEntity
		var temp, startedAt, confirmedAt, phase, tier, reason sql.NullString
		var tempScore sql.NullFloat64
		var durTrading, durCal sql.NullInt64
		if err := rows.Scan(
			&m.SnapshotID, &m.AssetID, &m.TradingDate, &temp, &tempScore,
			&m.Direction, &m.Confirmation, &m.Side, &startedAt, &confirmedAt,
			&durTrading, &durCal, &phase, &tier, &reason,
		); err != nil {
			return nil, err
		}
		m.Temperature = strPtr(temp)
		m.TemperatureScore = floatPtr(tempScore)
		m.StartedAt = strPtr(startedAt)
		m.ConfirmedAt = strPtr(confirmedAt)
		m.DurationTradingDays = intPtr(durTrading)
		m.DurationCalendarDays = intPtr(durCal)
		m.Phase = strPtr(phase)
		m.Tier = strPtr(tier)
		m.Reason = strPtr(reason)
		list = append(list, m)
	}
	return list, rows.Err()
}

func (s *SQLiteStore) ListStrengthDailyRankings(snapshotID, scopeID, currency string) ([]StrengthDailyRankingEntity, error) {
	rows, err := s.db.Query(`
		SELECT snapshot_id, scope_id, currency, asset_id,
		       raw_factor, rank, score, delta, candidate_count, valid_count, reason
		FROM strength_daily_rankings
		WHERE snapshot_id = ? AND scope_id = ? AND currency = ?
		ORDER BY COALESCE(rank, 999999) ASC, asset_id ASC
	`, snapshotID, scopeID, currency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []StrengthDailyRankingEntity
	for rows.Next() {
		var r StrengthDailyRankingEntity
		var rawFactor, score, delta sql.NullFloat64
		var rank sql.NullInt64
		var reason sql.NullString
		if err := rows.Scan(
			&r.SnapshotID, &r.ScopeID, &r.Currency, &r.AssetID,
			&rawFactor, &rank, &score, &delta, &r.CandidateCount, &r.ValidCount, &reason,
		); err != nil {
			return nil, err
		}
		r.RawFactor = floatPtr(rawFactor)
		r.Rank = intPtr(rank)
		r.Score = floatPtr(score)
		r.Delta = floatPtr(delta)
		r.Reason = strPtr(reason)
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *SQLiteStore) ListTrendEvents(snapshotID, assetID string, eventTypes []string, limit, offset int) ([]TrendEventEntity, error) {
	var conds []string
	var args []any

	if snapshotID != "" {
		conds = append(conds, "snapshot_id = ?")
		args = append(args, snapshotID)
	}
	if assetID != "" {
		conds = append(conds, "asset_id = ?")
		args = append(args, assetID)
	}
	if len(eventTypes) > 0 {
		placeholders := make([]string, len(eventTypes))
		for i, et := range eventTypes {
			placeholders[i] = "?"
			args = append(args, et)
		}
		conds = append(conds, fmt.Sprintf("event_type IN (%s)", strings.Join(placeholders, ",")))
	}

	query := `
		SELECT event_id, asset_id, event_type, old_value, new_value,
		       as_of, detected_at, snapshot_id, algorithm_version, scope_id, dedupe_key
		FROM trend_events`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY as_of DESC, detected_at DESC, event_id ASC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []TrendEventEntity
	for rows.Next() {
		var e TrendEventEntity
		var oldVal, newVal sql.NullString
		if err := rows.Scan(
			&e.EventID, &e.AssetID, &e.EventType, &oldVal, &newVal,
			&e.AsOf, &e.DetectedAt, &e.SnapshotID, &e.AlgorithmVersion, &e.ScopeID, &e.DedupeKey,
		); err != nil {
			return nil, err
		}
		e.OldValue = strPtr(oldVal)
		e.NewValue = strPtr(newVal)
		list = append(list, e)
	}
	return list, rows.Err()
}
