package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

/**
 * Database 封装数据库连接和通用查询操作
 * 提供配置读写等通用数据库能力，避免各 Service 直接依赖全局变量
 */
type Database struct {
	db *sql.DB
}

/**
 * 初始化数据库连接和表结构
 * 创建数据目录、打开连接、建表并执行迁移
 */
func InitDatabase() (*Database, error) {
	dataDir := GetDataDir()
	dbPath := filepath.Join(dataDir, "onewin.db")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	conn.SetMaxOpenConns(2)
	conn.SetMaxIdleConns(2)

	if err := createTables(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("创建表结构失败: %w", err)
	}

	return &Database{db: conn}, nil
}

/**
 * 获取底层 sql.DB 连接
 * 供需要直接操作数据库的场景使用
 */
func (d *Database) DB() *sql.DB {
	return d.db
}

/**
 * 获取配置项的值
 * 如果配置项不存在返回空字符串，查询出错则返回错误
 */
func (d *Database) GetConfig(key string) (string, error) {
	var value string
	err := d.db.QueryRow("SELECT config_value FROM app_config WHERE config_key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("查询配置失败: %w", err)
	}
	return value, nil
}

/**
 * 设置配置项的值
 * 使用 UPSERT 语义，如果 key 已存在则更新
 */
func (d *Database) SetConfig(key, value string) error {
	now := NowFormatted()
	_, err := d.db.Exec(
		"INSERT INTO app_config (config_key, config_value, updated_at) VALUES (?, ?, ?) ON CONFLICT(config_key) DO UPDATE SET config_value = ?, updated_at = ?",
		key, value, now, value, now,
	)
	if err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	return nil
}

/**
 * 批量获取配置项
 * 根据 keys 列表逐一查询，忽略不存在的键
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

	rows, err := d.db.Query("SELECT config_key, config_value FROM app_config WHERE config_key IN ("+placeholders+")", args...)
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
 * 关闭数据库连接
 */
func (d *Database) Close() {
	if d.db != nil {
		d.db.Close()
	}
}

/**
 * 创建所有数据库表和索引
 * 包含初始化建表和安全迁移逻辑
 */
func createTables(db *sql.DB) error {
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
			work_dir TEXT DEFAULT '',
			commands TEXT NOT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES shortcut_category(id) ON DELETE SET NULL
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
		`CREATE INDEX IF NOT EXISTS idx_shortcut_command_category_id ON shortcut_command(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sub_app_dir_name ON sub_app(dir_name)`,
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
		`CREATE INDEX IF NOT EXISTS idx_server_session_host_user ON server_session(host, user)`,
		`CREATE TABLE IF NOT EXISTS server_category (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	migrations := []string{
		`ALTER TABLE sub_app ADD COLUMN app_type TEXT NOT NULL DEFAULT 'static'`,
		`ALTER TABLE shortcut_cmd_group RENAME TO shortcut_cmd_category`,
		`ALTER TABLE shortcut_cmd RENAME COLUMN group_id TO category_id`,
		`ALTER TABLE ssh_server ADD COLUMN category_id INTEGER DEFAULT NULL`,
		`ALTER TABLE ssh_server ADD COLUMN session_name TEXT DEFAULT ''`,
		`ALTER TABLE ssh_server ADD COLUMN use_key_login INTEGER DEFAULT 0`,
		`ALTER TABLE ssh_session_category RENAME TO server_category`,
		`ALTER TABLE ssh_server RENAME TO server_session`,
	}

	for _, stmt := range migrations {
		db.Exec(stmt)
	}

	dropShortcutShellColumnIfExists(db)

	return nil
}

func dropShortcutShellColumnIfExists(db *sql.DB) {
	cols, err := getTableColumns(db, "shortcut_command")
	if err != nil {
		return
	}
	if !hasColumn(cols, "shell") {
		return
	}

	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS shortcut_command_new (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id INTEGER,
		name TEXT NOT NULL,
		work_dir TEXT DEFAULT '',
		commands TEXT NOT NULL,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES shortcut_category(id) ON DELETE SET NULL
	)`)
	_, _ = db.Exec(`INSERT INTO shortcut_command_new (id, category_id, name, work_dir, commands, sort_order, created_at, updated_at)
		SELECT id, category_id, name, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_command`)
	_, _ = db.Exec(`DROP TABLE shortcut_command`)
	_, _ = db.Exec(`ALTER TABLE shortcut_command_new RENAME TO shortcut_command`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_shortcut_command_category_id ON shortcut_command(category_id)`)
}

func getTableColumns(db *sql.DB, tableName string) ([]string, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}
	return columns, nil
}

func hasColumn(columns []string, target string) bool {
	target = strings.ToLower(target)
	for _, col := range columns {
		if strings.ToLower(col) == target {
			return true
		}
	}
	return false
}
