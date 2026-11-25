package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cyclecompass/api/internal/dto"

	_ "modernc.org/sqlite"
)

type AccountEntity struct {
	AccountID     string
	AvailableCash float64
	CashPool      float64
}

type PositionEntity struct {
	Symbol          string
	SignalSymbol    string
	Name            string
	Shares          float64
	CostPrice       float64
	LastPrice       float64
	LastPriceDate   string
	LastPriceSource string
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite 单写连接池，避免高并发写入触发 SQLITE_BUSY
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	store := &SQLiteStore{db: db}
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	if err := store.seedInitialData(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) initSchema() error {
	schema := `
	PRAGMA journal_mode = WAL;
	PRAGMA busy_timeout = 5000;
	PRAGMA synchronous = NORMAL;
	PRAGMA cache_size = -2000;
	PRAGMA mmap_size = 67108864;
	PRAGMA optimize;

	CREATE TABLE IF NOT EXISTS accounts (
		account_id TEXT PRIMARY KEY,
		available_cash REAL NOT NULL DEFAULT 0.0,
		cash_pool REAL NOT NULL DEFAULT 0.0,
		updated_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS positions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		account_id TEXT NOT NULL,
		symbol TEXT NOT NULL,
		name TEXT NOT NULL,
		shares REAL NOT NULL DEFAULT 0.0,
		cost_price REAL NOT NULL DEFAULT 0.0,
		FOREIGN KEY (account_id) REFERENCES accounts(account_id)
	);
	CREATE INDEX IF NOT EXISTS idx_positions_account ON positions(account_id, id);

	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		account_id TEXT NOT NULL,
		date TEXT NOT NULL,
		event_type TEXT NOT NULL,
		symbol TEXT,
		target_name TEXT NOT NULL,
		shares REAL NOT NULL DEFAULT 0.0,
		price REAL NOT NULL DEFAULT 0.0,
		amount REAL NOT NULL DEFAULT 0.0,
		cash_inflow REAL NOT NULL DEFAULT 0.0,
		cash_outflow REAL NOT NULL DEFAULT 0.0,
		cash_pool_delta REAL NOT NULL DEFAULT 0.0,
		cash_balance_after REAL NOT NULL DEFAULT 0.0,
		cash_pool_after REAL NOT NULL DEFAULT 0.0,
		description TEXT NOT NULL,
		FOREIGN KEY (account_id) REFERENCES accounts(account_id)
	);
	CREATE INDEX IF NOT EXISTS idx_transactions_account ON transactions(account_id, id ASC);
	`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}
	return s.runMigrations()
}

func (s *SQLiteStore) seedInitialData() error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM accounts WHERE account_id = 'ACC_DEMO_001'").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Format(time.RFC3339)
	_, err = tx.Exec(
		"INSERT INTO accounts (account_id, available_cash, cash_pool, updated_at) VALUES (?, ?, ?, ?)",
		"ACC_DEMO_001", 5200.0, 1000.0, now,
	)
	if err != nil {
		return err
	}

	// Positions（分离真实交易 ETF 与信号指数，绝不用指数代码冒充 ETF 持仓）
	_, err = tx.Exec(
		"INSERT INTO positions (account_id, symbol, signal_symbol, name, shares, cost_price, last_price, last_price_date, last_price_source) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"ACC_DEMO_001", "515100", "930955", "红利低波100ETF(515100)", 12500.0, 1.12, 1.136, "2026-11-09", "示范样本历史成交价",
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		"INSERT INTO positions (account_id, symbol, signal_symbol, name, shares, cost_price, last_price, last_price_date, last_price_source) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"ACC_DEMO_001", "159949", "399673", "创业板50ETF(159949)", 8000.0, 0.85, 0.85, "2026-11-09", "示范样本建仓参考价",
	)
	if err != nil {
		return err
	}

	// Transactions
	_, err = tx.Exec(`
		INSERT INTO transactions (
			account_id, date, event_type, symbol, target_name, shares, price, amount,
			cash_inflow, cash_outflow, cash_pool_delta, cash_balance_after, cash_pool_after, description
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"ACC_DEMO_001", "2026-11-16", "cash_deposit", nil, "计划现金池", 0.0, 0.0, 1000.0,
		1000.0, 1000.0, 1000.0, 0.0, 1000.0, "中间区：本周定投资金 1000 元存入计划现金池",
	)
	if err != nil {
		return err
	}

	sym515100 := "515100"
	_, err = tx.Exec(`
		INSERT INTO transactions (
			account_id, date, event_type, symbol, target_name, shares, price, amount,
			cash_inflow, cash_outflow, cash_pool_delta, cash_balance_after, cash_pool_after, description
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"ACC_DEMO_001", "2026-11-09", "weekly_invest", sym515100, "红利低波100(515100)", 880.0, 1.136, 1000.0,
		1000.0, 1000.0, 0.0, 0.0, 0.0, "高区：周定投 1000 元买入红利低波100",
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetAccount(accountID string) (*AccountEntity, error) {
	row := s.db.QueryRow("SELECT account_id, available_cash, cash_pool FROM accounts WHERE account_id = ?", accountID)
	var acc AccountEntity
	if err := row.Scan(&acc.AccountID, &acc.AvailableCash, &acc.CashPool); err != nil {
		return nil, err
	}
	return &acc, nil
}

func (s *SQLiteStore) GetPositions(accountID string) ([]PositionEntity, error) {
	rows, err := s.db.Query(
		"SELECT symbol, signal_symbol, name, shares, cost_price, last_price, last_price_date, last_price_source FROM positions WHERE account_id = ? ORDER BY id ASC",
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PositionEntity
	for rows.Next() {
		var p PositionEntity
		if err := rows.Scan(&p.Symbol, &p.SignalSymbol, &p.Name, &p.Shares, &p.CostPrice, &p.LastPrice, &p.LastPriceDate, &p.LastPriceSource); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (s *SQLiteStore) UpdatePositionLastPrice(accountID, symbol string, price float64, priceDate, source string) error {
	if price <= 0 {
		return nil
	}
	_, err := s.db.Exec(
		"UPDATE positions SET last_price = ?, last_price_date = ?, last_price_source = ? WHERE account_id = ? AND symbol = ?",
		price, priceDate, source, accountID, symbol,
	)
	return err
}

func (s *SQLiteStore) GetTransactions(accountID string, limit int) ([]dto.LedgerEntry, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if limit > 0 {
		rows, err = s.db.Query(
			"SELECT date, event_type, symbol, target_name, shares, price, amount, cash_inflow, cash_outflow, cash_pool_delta, cash_balance_after, cash_pool_after, description FROM transactions WHERE account_id = ? ORDER BY id DESC LIMIT ?",
			accountID, limit,
		)
	} else {
		rows, err = s.db.Query(
			"SELECT date, event_type, symbol, target_name, shares, price, amount, cash_inflow, cash_outflow, cash_pool_delta, cash_balance_after, cash_pool_after, description FROM transactions WHERE account_id = ? ORDER BY id DESC",
			accountID,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []dto.LedgerEntry
	for rows.Next() {
		var e dto.LedgerEntry
		var sym sql.NullString
		if err := rows.Scan(
			&e.Date, &e.EventType, &sym, &e.TargetName,
			&e.Shares, &e.Price, &e.Amount, &e.CashInflow, &e.CashOutflow,
			&e.CashPoolDelta, &e.CashBalanceAfter, &e.CashPoolAfter, &e.Description,
		); err != nil {
			return nil, err
		}
		if sym.Valid {
			e.Symbol = &sym.String
		} else {
			e.Symbol = nil
		}
		list = append(list, e)
	}
	return list, nil
}
