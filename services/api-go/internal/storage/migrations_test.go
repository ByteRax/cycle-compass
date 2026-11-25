package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) (*SQLiteStore, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test_cyclecompass.db")
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return store, dbPath
}

func TestMigrationsIdempotentAndTablesExist(t *testing.T) {
	store, dbPath := openTestStore(t)

	expectedTables := []string{
		"schema_migrations",
		"assets",
		"asset_categories",
		"asset_memberships",
		"universes",
		"universe_members",
		"daily_bars",
		"fx_rates",
		"algorithm_versions",
		"trend_snapshots",
		"trend_daily_metrics",
		"strength_daily_rankings",
		"trend_events",
		"asset_feature_tags",
		"data_sync_jobs",
	}

	for _, tbl := range expectedTables {
		var name string
		err := store.db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl,
		).Scan(&name)
		if err != nil {
			t.Fatalf("expected table %q to exist: %v", tbl, err)
		}
	}

	var idxName string
	if err := store.db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='index' AND name='idx_daily_bars_asset_date'",
	).Scan(&idxName); err != nil {
		t.Fatalf("expected index idx_daily_bars_asset_date to exist: %v", err)
	}

	// Re-running migrations must be idempotent
	if err := store.runMigrations(); err != nil {
		t.Fatalf("second runMigrations failed: %v", err)
	}

	var count int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != len(v4Migrations) {
		t.Fatalf("expected %d applied migrations, got %d", len(v4Migrations), count)
	}

	// Reopening store on the same file must also succeed idempotently
	store2, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("reopen NewSQLiteStore failed: %v", err)
	}
	defer store2.Close()
}

func TestMigrationRollbackOnFailure(t *testing.T) {
	store, _ := openTestStore(t)

	badMigrations := []migration{
		{
			version: 99,
			name:    "broken_migration",
			sql: `
				CREATE TABLE temp_should_rollback (id INTEGER PRIMARY KEY);
				INVALID SQL STATEMENT HERE;
			`,
		},
	}

	err := store.runMigrationsWithList(badMigrations)
	if err == nil {
		t.Fatal("expected error from broken migration, got nil")
	}

	var tbl string
	err = store.db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='temp_should_rollback'",
	).Scan(&tbl)
	if err != sql.ErrNoRows {
		t.Fatalf("expected temp_should_rollback to be rolled back, got err=%v name=%q", err, tbl)
	}

	var count int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 99").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected version 99 not recorded, got count=%d", count)
	}
}

func TestLegacyTablesAndSeedPreserved(t *testing.T) {
	store, _ := openTestStore(t)

	verifyLegacy := func() {
		t.Helper()
		acc, err := store.GetAccount("ACC_DEMO_001")
		if err != nil {
			t.Fatalf("GetAccount failed: %v", err)
		}
		if acc.AvailableCash != 5200.0 || acc.CashPool != 1000.0 {
			t.Fatalf("unexpected account balances: %+v", acc)
		}

		positions, err := store.GetPositions("ACC_DEMO_001")
		if err != nil {
			t.Fatalf("GetPositions failed: %v", err)
		}
		if len(positions) != 2 {
			t.Fatalf("expected 2 seeded positions, got %d", len(positions))
		}

		txs, err := store.GetTransactions("ACC_DEMO_001", 0)
		if err != nil {
			t.Fatalf("GetTransactions failed: %v", err)
		}
		if len(txs) != 2 {
			t.Fatalf("expected 2 seeded transactions, got %d", len(txs))
		}
	}

	verifyLegacy()
	if err := store.runMigrations(); err != nil {
		t.Fatalf("re-run migrations: %v", err)
	}
	verifyLegacy()
}

func TestUniqueConstraintsAndNullPreservation(t *testing.T) {
	store, _ := openTestStore(t)

	// D-01: Same symbol in different markets must not conflict
	err := store.UpsertAssets([]AssetEntity{
		{
			AssetID: "CN_IDX:000001", Market: "CN_IDX", Symbol: "000001",
			Name: "上证指数", AssetType: "index", MeasureType: "index", PriceType: "close",
			Currency: "CNY", Timezone: "Asia/Shanghai", IsActive: true,
			CreatedAt: "2026-11-24T00:00:00Z", UpdatedAt: "2026-11-24T00:00:00Z",
		},
		{
			AssetID: "CN_STK:000001", Market: "CN_STK", Symbol: "000001",
			Name: "平安银行", AssetType: "equity", MeasureType: "price", PriceType: "close",
			Currency: "CNY", Timezone: "Asia/Shanghai", IsActive: true,
			CreatedAt: "2026-11-24T00:00:00Z", UpdatedAt: "2026-11-24T00:00:00Z",
		},
	})
	if err != nil {
		t.Fatalf("UpsertAssets failed: %v", err)
	}

	// Same asset in multiple universes
	if err := store.UpsertUniverses([]UniverseEntity{
		{UniverseID: "u_core", Name: "核心池", IsPublic: true, SortOrder: 1},
		{UniverseID: "u_watch", Name: "观察池", IsPublic: true, SortOrder: 2},
	}); err != nil {
		t.Fatalf("UpsertUniverses failed: %v", err)
	}
	if err := store.UpsertUniverseMembers([]UniverseMemberEntity{
		{UniverseID: "u_core", AssetID: "CN_IDX:000001", SortOrder: 1},
		{UniverseID: "u_watch", AssetID: "CN_IDX:000001", SortOrder: 1},
	}); err != nil {
		t.Fatalf("UpsertUniverseMembers failed: %v", err)
	}

	members, err := store.ListUniverseMembers("u_core")
	if err != nil || len(members) != 1 || members[0].AssetID != "CN_IDX:000001" {
		t.Fatalf("ListUniverseMembers unexpected: members=%+v err=%v", members, err)
	}

	// asset_memberships unique(category_id, asset_id, valid_from)
	if _, err := store.db.Exec(
		`INSERT INTO asset_memberships (category_id, asset_id, weight, valid_from) VALUES ('cat1', 'CN_IDX:000001', 0.5, '2026-01-01')`,
	); err != nil {
		t.Fatalf("first insert asset_memberships failed: %v", err)
	}
	if _, err := store.db.Exec(
		`INSERT INTO asset_memberships (category_id, asset_id, weight, valid_from) VALUES ('cat1', 'CN_IDX:000001', 0.6, '2026-01-01')`,
	); err == nil {
		t.Fatal("expected UNIQUE constraint error on duplicate asset_memberships")
	}

	// daily_bars: revision 0 vs revision 1, and NULL != 0
	closeVal := 3200.5
	if err := store.UpsertDailyBars([]DailyBarEntity{
		{
			AssetID: "CN_IDX:000001", TradingDate: "2026-11-24", Provider: "akshare", Revision: 0,
			Close: &closeVal, Volume: nil, QualityStatus: "ok", UpdatedAt: "2026-11-24T08:00:00Z",
		},
		{
			AssetID: "CN_IDX:000001", TradingDate: "2026-11-24", Provider: "akshare", Revision: 1,
			Close: &closeVal, Volume: nil, QualityStatus: "revised", UpdatedAt: "2026-11-24T09:00:00Z",
		},
	}); err != nil {
		t.Fatalf("UpsertDailyBars failed: %v", err)
	}
	bars, err := store.GetDailyBars("CN_IDX:000001", "2026-11-24", "2026-11-24", 0)
	if err != nil {
		t.Fatalf("GetDailyBars failed: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("expected 2 revisions of daily_bars, got %d", len(bars))
	}
	if bars[0].Volume != nil || bars[0].Open != nil {
		t.Fatalf("expected nil for missing Volume/Open (null != 0), got Volume=%v Open=%v", bars[0].Volume, bars[0].Open)
	}

	// trend_events dedupe_key UNIQUE
	ev := TrendEventEntity{
		EventID: "ev1", AssetID: "CN_IDX:000001", EventType: "right_started",
		AsOf: "2026-11-24", DetectedAt: "2026-11-24T08:00:00Z",
		SnapshotID: "snap1", AlgorithmVersion: "v1", ScopeID: "u_core",
		DedupeKey: "snap1:CN_IDX:000001:right_started:2026-11-24",
	}
	if err := store.InsertTrendEvents([]TrendEventEntity{ev}); err != nil {
		t.Fatalf("InsertTrendEvents failed: %v", err)
	}
	// Idempotent insert via helper
	ev2 := ev
	ev2.EventID = "ev2"
	if err := store.InsertTrendEvents([]TrendEventEntity{ev2}); err != nil {
		t.Fatalf("idempotent InsertTrendEvents failed: %v", err)
	}
	// Raw duplicate must fail UNIQUE constraint
	if _, err := store.db.Exec(
		`INSERT INTO trend_events (event_id, asset_id, event_type, as_of, detected_at, snapshot_id, algorithm_version, scope_id, dedupe_key)
		 VALUES ('ev3', 'CN_IDX:000001', 'right_started', '2026-11-24', '2026-11-24T08:00:00Z', 'snap1', 'v1', 'u_core', ?)`,
		ev.DedupeKey,
	); err == nil {
		t.Fatal("expected UNIQUE constraint error on duplicate dedupe_key")
	}
}

func TestSnapshotAtomicPublishAndSwitch(t *testing.T) {
	store, _ := openTestStore(t)

	tempHot := "hot"
	score75 := 75.0
	rank1 := 1
	score92 := 92.0
	cov := 1.0

	snap1 := TrendSnapshotEntity{
		SnapshotID: "snap_v1", ScopeID: "u_core", Currency: "CNY", Mode: "latest_available",
		AsOf: "2026-11-23", ComputedAt: "2026-11-23T08:00:00Z", AlgorithmVersion: "v4.0",
		CandidateCount: 1, ValidCount: 1, Coverage: &cov, Status: "ok", IsCurrent: true,
	}
	if err := store.PublishTrendSnapshot(
		snap1,
		[]TrendDailyMetricEntity{{
			SnapshotID: "snap_v1", AssetID: "CN_IDX:000001", TradingDate: "2026-11-23",
			Temperature: &tempHot, TemperatureScore: &score75,
			Direction: "up", Confirmation: "confirmed", Side: "right",
		}},
		[]StrengthDailyRankingEntity{{
			SnapshotID: "snap_v1", ScopeID: "u_core", Currency: "CNY", AssetID: "CN_IDX:000001",
			Rank: &rank1, Score: &score92, CandidateCount: 1, ValidCount: 1,
		}},
		nil,
	); err != nil {
		t.Fatalf("PublishTrendSnapshot snap_v1 failed: %v", err)
	}

	snap2 := TrendSnapshotEntity{
		SnapshotID: "snap_v2", ScopeID: "u_core", Currency: "CNY", Mode: "latest_available",
		AsOf: "2026-11-24", ComputedAt: "2026-11-24T08:00:00Z", AlgorithmVersion: "v4.0",
		CandidateCount: 1, ValidCount: 1, Coverage: &cov, Status: "ok", IsCurrent: true,
	}
	if err := store.PublishTrendSnapshot(
		snap2,
		[]TrendDailyMetricEntity{{
			SnapshotID: "snap_v2", AssetID: "CN_IDX:000001", TradingDate: "2026-11-24",
			Temperature: &tempHot, TemperatureScore: &score75,
			Direction: "up", Confirmation: "confirmed", Side: "right",
		}},
		nil,
		nil,
	); err != nil {
		t.Fatalf("PublishTrendSnapshot snap_v2 failed: %v", err)
	}

	cur, err := store.GetCurrentSnapshot("u_core", "CNY", "latest_available")
	if err != nil {
		t.Fatalf("GetCurrentSnapshot failed: %v", err)
	}
	if cur.SnapshotID != "snap_v2" || !cur.IsCurrent {
		t.Fatalf("expected current snapshot snap_v2, got %+v", cur)
	}

	oldSnap, err := store.GetSnapshotByID("snap_v1")
	if err != nil {
		t.Fatalf("GetSnapshotByID snap_v1 failed: %v", err)
	}
	if oldSnap.IsCurrent {
		t.Fatal("expected snap_v1 is_current=false after snap_v2 published")
	}

	// Verify snap_v1 metrics are still preserved
	m1, err := store.ListTrendDailyMetrics("snap_v1", []string{"CN_IDX:000001"})
	if err != nil || len(m1) != 1 {
		t.Fatalf("expected snap_v1 metrics preserved, got len=%d err=%v", len(m1), err)
	}

	// Atomic switch back to snap_v1
	if err := store.SetCurrentSnapshot("snap_v1"); err != nil {
		t.Fatalf("SetCurrentSnapshot snap_v1 failed: %v", err)
	}
	cur, err = store.GetCurrentSnapshot("u_core", "CNY", "latest_available")
	if err != nil {
		t.Fatalf("GetCurrentSnapshot after switch failed: %v", err)
	}
	if cur.SnapshotID != "snap_v1" {
		t.Fatalf("expected snap_v1 to be current after switch, got %s", cur.SnapshotID)
	}
	s2, err := store.GetSnapshotByID("snap_v2")
	if err != nil || s2.IsCurrent {
		t.Fatalf("expected snap_v2 is_current=false after switch, got %+v err=%v", s2, err)
	}

	// Switching to non-existent snapshot must fail without changing current
	if err := store.SetCurrentSnapshot("non_existent"); err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows for non_existent snapshot, got %v", err)
	}
	cur, err = store.GetCurrentSnapshot("u_core", "CNY", "latest_available")
	if err != nil || cur.SnapshotID != "snap_v1" {
		t.Fatalf("expected snap_v1 to remain current after failed switch, got %+v err=%v", cur, err)
	}
}
