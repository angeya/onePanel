package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB

/**
 * 初始化数据库连接和表结构
 */
func InitDatabase() error {
	dataDir := GetDataDir()
	dbPath := filepath.Join(dataDir, "onewin.db")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("创建表结构失败: %w", err)
	}

	return nil
}

/**
 * 创建所有数据库表
 */
func createTables() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS shortcut_category (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS shortcut_command (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category_id INTEGER,
			name TEXT NOT NULL,
			shell TEXT DEFAULT 'cmd.exe',
			work_dir TEXT DEFAULT '',
			commands TEXT NOT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES shortcut_category(id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS command_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command TEXT NOT NULL,
			shell TEXT DEFAULT 'cmd.exe',
			work_dir TEXT DEFAULT '',
			executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS app_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			config_key TEXT NOT NULL UNIQUE,
			config_value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sub_app (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_type TEXT NOT NULL DEFAULT 'static',
			dir_name TEXT NOT NULL DEFAULT '',
			display_name TEXT NOT NULL,
			icon_path TEXT DEFAULT '',
			entry_url TEXT DEFAULT '',
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS shortcut_cmd_group (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS shortcut_cmd (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER,
			name TEXT NOT NULL,
			shell TEXT DEFAULT 'cmd.exe',
			work_dir TEXT DEFAULT '',
			commands TEXT NOT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (group_id) REFERENCES shortcut_cmd_group(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_command_history_executed_at ON command_history(executed_at)`,
		`CREATE INDEX IF NOT EXISTS idx_shortcut_command_category_id ON shortcut_command(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sub_app_dir_name ON sub_app(dir_name)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	migrations := []string{
		`ALTER TABLE sub_app ADD COLUMN app_type TEXT NOT NULL DEFAULT 'static'`,
	}

	for _, stmt := range migrations {
		db.Exec(stmt)
	}

	return nil
}

/**
 * 获取应用数据存储目录
 */
func GetDataDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "./data"
	}
	return filepath.Join(filepath.Dir(exePath), "data")
}

/**
 * 关闭数据库连接
 */
func CloseDatabase() {
	if db != nil {
		db.Close()
	}
}
