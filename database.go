package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const sqliteDSNOptions = "?_journal_mode=WAL&_busy_timeout=5000"

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS app_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_key TEXT NOT NULL UNIQUE,
		config_value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS sub_app (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_type TEXT NOT NULL DEFAULT 'static',
		name TEXT NOT NULL,
		entry_url TEXT DEFAULT '',
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
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
		work_dir TEXT DEFAULT '',
		commands TEXT NOT NULL,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES shortcut_category(id) ON DELETE SET NULL
	)`,
	`CREATE TABLE IF NOT EXISTS shortcut_cmd_category (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS shortcut_cmd (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id INTEGER,
		name TEXT NOT NULL,
		shell TEXT DEFAULT 'cmd.exe',
		work_dir TEXT DEFAULT '',
		commands TEXT NOT NULL,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES shortcut_cmd_category(id) ON DELETE SET NULL
	)`,
	`CREATE TABLE IF NOT EXISTS server_category (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS server_session (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id INTEGER,
		session_name TEXT DEFAULT '',
		host TEXT NOT NULL,
		port INTEGER DEFAULT 22,
		user TEXT NOT NULL,
		use_key_login INTEGER DEFAULT 0,
		key_deployed INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES server_category(id) ON DELETE SET NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_sub_app_name ON sub_app(name)`,
	`CREATE INDEX IF NOT EXISTS idx_shortcut_command_category_id ON shortcut_command(category_id)`,
	`CREATE INDEX IF NOT EXISTS idx_shortcut_cmd_category_id ON shortcut_cmd(category_id)`,
	`CREATE INDEX IF NOT EXISTS idx_server_session_category_id ON server_session(category_id)`,
	`CREATE INDEX IF NOT EXISTS idx_server_session_host_user ON server_session(host, user)`,
}

/**
 * Database 封装数据库连接和通用查询操作。
 * 负责统一管理 SQLite 连接、配置读写和基础 Schema 初始化。
 */
type Database struct {
	db *sql.DB
}

/**
 * InitDatabase 初始化数据库连接和当前版本所需的表结构。
 * 当前版本直接以最新 Schema 建库，不再兼容旧版本迁移。
 */
func InitDatabase() (*Database, error) {
	dataDir := GetDataDir()
	dbPath := filepath.Join(dataDir, "onewin.db")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath+sqliteDSNOptions)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	conn.SetMaxOpenConns(2)
	conn.SetMaxIdleConns(2)

	if err := initializeSchema(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("初始化数据库结构失败: %w", err)
	}

	return &Database{db: conn}, nil
}

/**
 * DB 返回底层 sql.DB 连接，供需要直接执行 SQL 的服务层复用。
 */
func (d *Database) DB() *sql.DB {
	return d.db
}

/**
 * GetConfig 获取单个配置项。
 * 当配置不存在时返回空字符串，不视为错误。
 */
func (d *Database) GetConfig(key string) (string, error) {
	var value string
	err := d.db.QueryRow(
		"SELECT config_value FROM app_config WHERE config_key = ?",
		key,
	).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("查询配置失败: %w", err)
	}
	return value, nil
}

/**
 * SetConfig 保存单个配置项。
 * 使用 UPSERT 保证同一 key 可以重复写入而无需先查再改。
 */
func (d *Database) SetConfig(key, value string) error {
	now := NowFormatted()
	_, err := d.db.Exec(
		"INSERT INTO app_config (config_key, config_value, updated_at) VALUES (?, ?, ?) ON CONFLICT(config_key) DO UPDATE SET config_value = ?, updated_at = ?",
		key,
		value,
		now,
		value,
		now,
	)
	if err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	return nil
}

/**
 * GetConfigs 批量获取多个配置项。
 * 通过单次 IN 查询减少启动阶段的数据库往返次数。
 */
func (d *Database) GetConfigs(keys []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(keys) == 0 {
		return result, nil
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(keys)), ",")
	args := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		args = append(args, key)
	}

	rows, err := d.db.Query(
		"SELECT config_key, config_value FROM app_config WHERE config_key IN ("+placeholders+")",
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("批量查询配置失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("读取配置失败: %w", err)
		}
		result[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历配置结果失败: %w", err)
	}

	return result, nil
}

/**
 * Close 关闭数据库连接。
 * 应用退出时调用，释放 SQLite 句柄。
 */
func (d *Database) Close() {
	if d.db != nil {
		d.db.Close()
	}
}

/**
 * initializeSchema 初始化当前版本所需的完整 Schema。
 * 由于当前版本不兼容历史数据库，因此仅执行最新建表和索引语句。
 */
func initializeSchema(db *sql.DB) error {
	for _, stmt := range schemaStatements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
