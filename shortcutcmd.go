package main

import (
	"database/sql"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

/**
 * ShortcutCmdService 快速启动命令服务。
 * 负责快速启动功能的分类管理、命令 CRUD 和命令执行。
 */
type ShortcutCmdService struct {
	db *Database
}

/**
 * 创建 ShortcutCmdService 实例。
 */
func NewShortcutCmdService(db *Database) *ShortcutCmdService {
	return &ShortcutCmdService{db: db}
}

/**
 * GetCategories 获取所有快捷命令分类。
 */
func (s *ShortcutCmdService) GetCategories() ([]ShortcutCmdCategory, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, name, sort_order, created_at, updated_at FROM shortcut_cmd_category ORDER BY sort_order, id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []ShortcutCmdCategory
	for rows.Next() {
		var category ShortcutCmdCategory
		if err := rows.Scan(
			&category.Id,
			&category.Name,
			&category.SortOrder,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if categories == nil {
		categories = []ShortcutCmdCategory{}
	}
	return categories, nil
}

/**
 * CreateCategory 创建快捷命令分类。
 */
func (s *ShortcutCmdService) CreateCategory(name string, sortOrder int) (*ShortcutCmdCategory, error) {
	if err := validateShortcutCategoryName(name); err != nil {
		return nil, err
	}

	now := NowFormatted()
	result, err := s.db.DB().Exec(
		"INSERT INTO shortcut_cmd_category (name, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?)",
		name,
		sortOrder,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("创建分类失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return &ShortcutCmdCategory{
		Id:        id,
		Name:      name,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

/**
 * UpdateCategory 更新快捷命令分类。
 */
func (s *ShortcutCmdService) UpdateCategory(id int64, name string, sortOrder int) error {
	if err := validateShortcutCategoryName(name); err != nil {
		return err
	}

	now := NowFormatted()
	_, err := s.db.DB().Exec(
		"UPDATE shortcut_cmd_category SET name = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		name,
		sortOrder,
		now,
		id,
	)
	return err
}

/**
 * DeleteCategory 删除快捷命令分类。
 */
func (s *ShortcutCmdService) DeleteCategory(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM shortcut_cmd_category WHERE id = ?", id)
	return err
}

/**
 * GetCommands 获取所有快捷命令。
 */
func (s *ShortcutCmdService) GetCommands() ([]ShortcutCmd, error) {
	return s.queryCommands(
		"SELECT id, category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_cmd ORDER BY sort_order, id",
	)
}

/**
 * GetCommandsByCategory 根据分类获取快捷命令。
 */
func (s *ShortcutCmdService) GetCommandsByCategory(categoryID int64) ([]ShortcutCmd, error) {
	return s.queryCommands(
		"SELECT id, category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_cmd WHERE category_id = ? ORDER BY sort_order, id",
		categoryID,
	)
}

/**
 * CreateCommand 创建快捷命令。
 */
func (s *ShortcutCmdService) CreateCommand(categoryID *int64, name, shell, workDir, commands string, sortOrder int) (*ShortcutCmd, error) {
	normalizedShell, err := normalizeShortcutShell(shell)
	if err != nil {
		return nil, err
	}
	if err := validateShortcutCommandInput(name, commands); err != nil {
		return nil, err
	}

	now := NowFormatted()
	result, err := s.db.DB().Exec(
		"INSERT INTO shortcut_cmd (category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		shortcutCategoryIDToNull(categoryID),
		name,
		normalizedShell,
		workDir,
		commands,
		sortOrder,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("创建快捷命令失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return &ShortcutCmd{
		Id:         id,
		CategoryId: categoryID,
		Name:       name,
		Shell:      normalizedShell,
		WorkDir:    workDir,
		Commands:   commands,
		SortOrder:  sortOrder,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

/**
 * UpdateCommand 更新快捷命令。
 */
func (s *ShortcutCmdService) UpdateCommand(id int64, categoryID *int64, name, shell, workDir, commands string, sortOrder int) error {
	normalizedShell, err := normalizeShortcutShell(shell)
	if err != nil {
		return err
	}
	if err := validateShortcutCommandInput(name, commands); err != nil {
		return err
	}

	now := NowFormatted()
	_, err = s.db.DB().Exec(
		"UPDATE shortcut_cmd SET category_id = ?, name = ?, shell = ?, work_dir = ?, commands = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		shortcutCategoryIDToNull(categoryID),
		name,
		normalizedShell,
		workDir,
		commands,
		sortOrder,
		now,
		id,
	)
	return err
}

/**
 * DeleteCommand 删除快捷命令。
 */
func (s *ShortcutCmdService) DeleteCommand(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM shortcut_cmd WHERE id = ?", id)
	return err
}

/**
 * ExecuteCommand 执行快捷命令。
 */
func (s *ShortcutCmdService) ExecuteCommand(id int64) error {
	command, err := s.getCommandByID(id)
	if err != nil {
		return fmt.Errorf("快捷命令不存在: %w", err)
	}
	return executeShellCommands(command.Shell, command.WorkDir, command.Commands)
}

/**
 * queryCommands 统一执行快捷命令查询并完成模型映射。
 */
func (s *ShortcutCmdService) queryCommands(query string, args ...interface{}) ([]ShortcutCmd, error) {
	rows, err := s.db.DB().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []ShortcutCmd
	for rows.Next() {
		command, err := scanShortcutCommand(rows.Scan)
		if err != nil {
			return nil, err
		}
		commands = append(commands, *command)
	}
	if commands == nil {
		commands = []ShortcutCmd{}
	}
	return commands, nil
}

/**
 * getCommandByID 根据 ID 查询单个快捷命令。
 */
func (s *ShortcutCmdService) getCommandByID(id int64) (*ShortcutCmd, error) {
	row := s.db.DB().QueryRow(
		"SELECT id, category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_cmd WHERE id = ?",
		id,
	)
	return scanShortcutCommand(row.Scan)
}

/**
 * scanShortcutCommand 统一完成 shortcut_cmd 记录到模型的扫描与转换。
 */
func scanShortcutCommand(scan func(dest ...interface{}) error) (*ShortcutCmd, error) {
	var command ShortcutCmd
	var categoryID sql.NullInt64
	if err := scan(
		&command.Id,
		&categoryID,
		&command.Name,
		&command.Shell,
		&command.WorkDir,
		&command.Commands,
		&command.SortOrder,
		&command.CreatedAt,
		&command.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if categoryID.Valid {
		command.CategoryId = &categoryID.Int64
	}
	return &command, nil
}

/**
 * shortcutCategoryIDToNull 将可选分类 ID 转为数据库可识别的 NullInt64。
 */
func shortcutCategoryIDToNull(categoryID *int64) sql.NullInt64 {
	if categoryID == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *categoryID, Valid: true}
}

/**
 * normalizeShortcutShell 归一化快捷命令使用的 shell。
 */
func normalizeShortcutShell(shell string) (string, error) {
	normalized := strings.TrimSpace(DefaultShell(shell))
	if normalized == "" {
		return "", fmt.Errorf("Shell 不能为空")
	}
	return normalized, nil
}

/**
 * validateShortcutCategoryName 校验分类名称。
 */
func validateShortcutCategoryName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("分类名称不能为空")
	}
	return nil
}

/**
 * validateShortcutCommandInput 校验快捷命令的关键输入。
 */
func validateShortcutCommandInput(name string, commands string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("命令名称不能为空")
	}
	if strings.TrimSpace(commands) == "" {
		return fmt.Errorf("命令内容不能为空")
	}
	return nil
}

/**
 * executeShellCommands 执行 shell 命令列表。
 * 根据 shell 类型使用 cmd /k 或 powershell -NoExit -Command 启动新终端窗口。
 */
func executeShellCommands(shell, workDir, commands string) error {
	normalizedShell, err := normalizeShortcutShell(shell)
	if err != nil {
		return err
	}
	shellPath := ResolveShellPath(normalizedShell)

	var cmd *exec.Cmd
	if isPowerShell(normalizedShell) {
		cmd = exec.Command(shellPath, "-NoExit", "-Command", commands)
	} else {
		cmd = exec.Command(shellPath, "/k", commands)
	}

	if workDir != "" {
		cmd.Dir = workDir
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("执行命令失败: %w", err)
	}

	runtime.KeepAlive(cmd)
	return nil
}

/**
 * isPowerShell 判断是否为 PowerShell 类型的 shell。
 */
func isPowerShell(shell string) bool {
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "powershell", "powershell.exe":
		return true
	default:
		return false
	}
}
