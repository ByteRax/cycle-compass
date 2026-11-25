package storage

import (
	"fmt"
	"sort"
	"time"
)

type migration struct {
	version int
	name    string
	sql     string
}

var v4Migrations = []migration{
	{
		version: 1,
		name:    "init_v4_trend_schema",
		sql: `
CREATE TABLE IF NOT EXISTS assets (
	asset_id TEXT PRIMARY KEY,
	market TEXT NOT NULL,
	symbol TEXT NOT NULL,
	name TEXT NOT NULL,
	short_name TEXT NOT NULL DEFAULT '',
	asset_type TEXT NOT NULL,
	measure_type TEXT NOT NULL DEFAULT 'price',
	price_type TEXT NOT NULL DEFAULT 'close',
	unit TEXT NOT NULL DEFAULT '',
	currency TEXT NOT NULL DEFAULT 'CNY',
	timezone TEXT NOT NULL DEFAULT 'Asia/Shanghai',
	is_active INTEGER NOT NULL DEFAULT 1,
	license_status TEXT NOT NULL DEFAULT 'unknown',
	source_name TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_assets_market_symbol ON assets(market, symbol);

CREATE TABLE IF NOT EXISTS asset_categories (
	category_id TEXT PRIMARY KEY,
	parent_id TEXT,
	name TEXT NOT NULL,
	level INTEGER NOT NULL DEFAULT 0,
	sort_order INTEGER NOT NULL DEFAULT 0,
	has_index INTEGER NOT NULL DEFAULT 0,
	index_asset_id TEXT
);
CREATE INDEX IF NOT EXISTS idx_asset_categories_parent ON asset_categories(parent_id, sort_order);

CREATE TABLE IF NOT EXISTS asset_memberships (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	category_id TEXT NOT NULL,
	asset_id TEXT NOT NULL,
	weight REAL,
	valid_from TEXT NOT NULL,
	valid_to TEXT,
	UNIQUE(category_id, asset_id, valid_from)
);
CREATE INDEX IF NOT EXISTS idx_asset_memberships_cat_valid ON asset_memberships(category_id, valid_from, valid_to);
CREATE INDEX IF NOT EXISTS idx_asset_memberships_asset ON asset_memberships(asset_id);

CREATE TABLE IF NOT EXISTS universes (
	universe_id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	is_public INTEGER NOT NULL DEFAULT 1,
	sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS universe_members (
	universe_id TEXT NOT NULL,
	asset_id TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY(universe_id, asset_id)
);

CREATE TABLE IF NOT EXISTS daily_bars (
	asset_id TEXT NOT NULL,
	trading_date TEXT NOT NULL,
	provider TEXT NOT NULL,
	revision INTEGER NOT NULL DEFAULT 0,
	open REAL,
	high REAL,
	low REAL,
	close REAL,
	adjusted_close REAL,
	total_return_index REAL,
	volume REAL,
	turnover REAL,
	float_market_cap REAL,
	fund_aum REAL,
	dividend_yield REAL,
	yield_pct REAL,
	quality_status TEXT NOT NULL DEFAULT 'ok',
	as_of TEXT NOT NULL DEFAULT '',
	available_at TEXT NOT NULL DEFAULT '',
	updated_at TEXT NOT NULL,
	PRIMARY KEY(asset_id, trading_date, provider, revision)
);
CREATE INDEX IF NOT EXISTS idx_daily_bars_asset_date ON daily_bars(asset_id, trading_date DESC);

CREATE TABLE IF NOT EXISTS fx_rates (
	pair TEXT NOT NULL,
	trading_date TEXT NOT NULL,
	rate REAL NOT NULL,
	provider TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	PRIMARY KEY(pair, trading_date, provider)
);

CREATE TABLE IF NOT EXISTS algorithm_versions (
	version TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	config_json TEXT NOT NULL DEFAULT '{}',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS trend_snapshots (
	snapshot_id TEXT PRIMARY KEY,
	scope_id TEXT NOT NULL,
	currency TEXT NOT NULL,
	mode TEXT NOT NULL,
	as_of TEXT NOT NULL,
	computed_at TEXT NOT NULL,
	algorithm_version TEXT NOT NULL,
	candidate_count INTEGER NOT NULL DEFAULT 0,
	valid_count INTEGER NOT NULL DEFAULT 0,
	coverage REAL,
	status TEXT NOT NULL,
	is_current INTEGER NOT NULL DEFAULT 0,
	meta_json TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_trend_snapshots_current ON trend_snapshots(scope_id, currency, mode, is_current);

CREATE TABLE IF NOT EXISTS trend_daily_metrics (
	snapshot_id TEXT NOT NULL,
	asset_id TEXT NOT NULL,
	trading_date TEXT NOT NULL,
	temperature TEXT,
	temperature_score REAL,
	direction TEXT NOT NULL DEFAULT 'unknown',
	confirmation TEXT NOT NULL DEFAULT 'none',
	side TEXT NOT NULL DEFAULT 'unknown',
	started_at TEXT,
	confirmed_at TEXT,
	duration_trading_days INTEGER,
	duration_calendar_days INTEGER,
	phase TEXT,
	tier TEXT,
	reason TEXT,
	PRIMARY KEY(snapshot_id, asset_id)
);

CREATE TABLE IF NOT EXISTS strength_daily_rankings (
	snapshot_id TEXT NOT NULL,
	scope_id TEXT NOT NULL,
	currency TEXT NOT NULL,
	asset_id TEXT NOT NULL,
	raw_factor REAL,
	rank INTEGER,
	score REAL,
	delta REAL,
	candidate_count INTEGER NOT NULL DEFAULT 0,
	valid_count INTEGER NOT NULL DEFAULT 0,
	reason TEXT,
	PRIMARY KEY(snapshot_id, scope_id, currency, asset_id)
);

CREATE TABLE IF NOT EXISTS trend_events (
	event_id TEXT PRIMARY KEY,
	asset_id TEXT NOT NULL,
	event_type TEXT NOT NULL,
	old_value TEXT,
	new_value TEXT,
	as_of TEXT NOT NULL,
	detected_at TEXT NOT NULL,
	snapshot_id TEXT NOT NULL,
	algorithm_version TEXT NOT NULL,
	scope_id TEXT NOT NULL,
	dedupe_key TEXT NOT NULL UNIQUE
);
CREATE INDEX IF NOT EXISTS idx_trend_events_asset_asof ON trend_events(asset_id, as_of DESC);

CREATE TABLE IF NOT EXISTS asset_feature_tags (
	snapshot_id TEXT NOT NULL,
	asset_id TEXT NOT NULL,
	tag_key TEXT NOT NULL,
	tag_label TEXT NOT NULL,
	scope_id TEXT NOT NULL,
	PRIMARY KEY(snapshot_id, asset_id, tag_key, scope_id)
);

CREATE TABLE IF NOT EXISTS data_sync_jobs (
	job_id TEXT PRIMARY KEY,
	job_type TEXT NOT NULL,
	status TEXT NOT NULL,
	started_at TEXT NOT NULL,
	finished_at TEXT,
	detail_json TEXT NOT NULL DEFAULT '{}'
);
`,
	},
}

func (s *SQLiteStore) runMigrations() error {
	return s.runMigrationsWithList(v4Migrations)
}

func (s *SQLiteStore) runMigrationsWithList(migrations []migration) error {
	const createMigrationsTable = `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TEXT NOT NULL
	);`
	if _, err := s.db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	sorted := make([]migration, len(migrations))
	copy(sorted, migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].version < sorted[j].version
	})

	for _, m := range sorted {
		var count int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&count); err != nil {
			return fmt.Errorf("check migration %d (%s): %w", m.version, m.name, err)
		}
		if count > 0 {
			continue
		}

		if err := s.applyMigration(m); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) applyMigration(m migration) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %d (%s): %w", m.version, m.name, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(m.sql); err != nil {
		return fmt.Errorf("exec migration %d (%s): %w", m.version, m.name, err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(
		"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
		m.version, m.name, now,
	); err != nil {
		return fmt.Errorf("record migration %d (%s): %w", m.version, m.name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d (%s): %w", m.version, m.name, err)
	}
	return nil
}
