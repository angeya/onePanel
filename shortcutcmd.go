package main

import (
	"database/sql"
	"fmt"
	"os/exec"
	"runtime"
)

/**
 * ShortcutCmdService 快速启动命令服务
 * 负责快速启动功能的分类管理、命令 CRUD 和命令执行
 * 通过依赖注入持有 Database 引用
 */
type ShortcutCmdService struct {
	db *Database
}

/**
 * 创建 ShortcutCmdService 实例
 * 注入 Database 依赖
 */
func NewShortcutCmdService(db *Database) *ShortcutCmdService {
	return &ShortcutCmdService{db: db}
}

/**
 * 获取所有快捷命令分类
 * 按排序字段和 ID 升序排列
 */
func (s *ShortcutCmdService) GetCategories() ([]ShortcutCmdCategory, error) {
	rows, err := s.db.DB().Query("SELECT id, name, sort_order, created_at, updated_at FROM shortcut_cmd_category ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []ShortcutCmdCategory
	for rows.Next() {
		var c ShortcutCmdCategory
		if err := rows.Scan(&c.Id, &c.Name, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	if categories == nil {
		categories = []ShortcutCmdCategory{}
	}
	return categories, nil
}

/**
 * 创建快捷命令分类
 */
func (s *ShortcutCmdService) CreateCategory(name string, sortOrder int) (*ShortcutCmdCategory, error) {
	now := NowFormatted()
	result, err := s.db.DB().Exec(
		"INSERT INTO shortcut_cmd_category (name, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?)",
		name, sortOrder, now, now,
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
 * 更新快捷命令分类
 */
func (s *ShortcutCmdService) UpdateCategory(id int64, name string, sortOrder int) error {
	now := NowFormatted()
	_, err := s.db.DB().Exec(
		"UPDATE shortcut_cmd_category SET name = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		name, sortOrder, now, id,
	)
	return err
}

/**
 * 删除快捷命令分类
 */
func (s *ShortcutCmdService) DeleteCategory(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM shortcut_cmd_category WHERE id = ?", id)
	return err
}

/**
 * 获取所有快捷命令
 * 按排序字段和 ID 升序排列
 */
func (s *ShortcutCmdService) GetCommands() ([]ShortcutCmd, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_cmd ORDER BY sort_order, id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []ShortcutCmd
	for rows.Next() {
		var cmd ShortcutCmd
		var categoryId sql.NullInt64
		if err := rows.Scan(&cmd.Id, &categoryId, &cmd.Name, &cmd.Shell, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt); err != nil {
			return nil, err
		}
		if categoryId.Valid {
			cmd.CategoryId = &categoryId.Int64
		}
		commands = append(commands, cmd)
	}
	if commands == nil {
		commands = []ShortcutCmd{}
	}
	return commands, nil
}

/**
 * 根据分类获取快捷命令
 */
func (s *ShortcutCmdService) GetCommandsByCategory(categoryId int64) ([]ShortcutCmd, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_cmd WHERE category_id = ? ORDER BY sort_order, id",
		categoryId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []ShortcutCmd
	for rows.Next() {
		var cmd ShortcutCmd
		var cId sql.NullInt64
		if err := rows.Scan(&cmd.Id, &cId, &cmd.Name, &cmd.Shell, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt); err != nil {
			return nil, err
		}
		if cId.Valid {
			cmd.CategoryId = &cId.Int64
		}
		commands = append(commands, cmd)
	}
	if commands == nil {
		commands = []ShortcutCmd{}
	}
	return commands, nil
}

/**
 * 创建快捷命令
 * shell 为空时默认使用 cmd.exe
 */
func (s *ShortcutCmdService) CreateCommand(categoryId *int64, name, shell, workDir, commands string, sortOrder int) (*ShortcutCmd, error) {
	now := NowFormatted()
	shell = DefaultShell(shell)

	var cId sql.NullInt64
	if categoryId != nil {
		cId = sql.NullInt64{Int64: *categoryId, Valid: true}
	}

	result, err := s.db.DB().Exec(
		"INSERT INTO shortcut_cmd (category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		cId, name, shell, workDir, commands, sortOrder, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("创建快捷命令失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return &ShortcutCmd{
		Id:         id,
		CategoryId: categoryId,
		Name:       name,
		Shell:      shell,
		WorkDir:    workDir,
		Commands:   commands,
		SortOrder:  sortOrder,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

/**
 * 更新快捷命令
 */
func (s *ShortcutCmdService) UpdateCommand(id int64, categoryId *int64, name, shell, workDir, commands string, sortOrder int) error {
	now := NowFormatted()
	shell = DefaultShell(shell)

	var cId sql.NullInt64
	if categoryId != nil {
		cId = sql.NullInt64{Int64: *categoryId, Valid: true}
	}

	_, err := s.db.DB().Exec(
		"UPDATE shortcut_cmd SET category_id = ?, name = ?, shell = ?, work_dir = ?, commands = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		cId, name, shell, workDir, commands, sortOrder, now, id,
	)
	return err
}

/**
 * 删除快捷命令
 */
func (s *ShortcutCmdService) DeleteCommand(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM shortcut_cmd WHERE id = ?", id)
	return err
}

/**
 * 执行快捷命令
 * 从数据库读取命令信息，根据 shell 类型启动对应终端进程
 */
func (s *ShortcutCmdService) ExecuteCommand(id int64) error {
	var cmd ShortcutCmd
	var categoryId sql.NullInt64
	err := s.db.DB().QueryRow(
		"SELECT id, category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_cmd WHERE id = ?",
		id,
	).Scan(&cmd.Id, &categoryId, &cmd.Name, &cmd.Shell, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt)
	if err != nil {
		return fmt.Errorf("快捷命令不存在: %w", err)
	}

	return executeShellCommands(cmd.Shell, cmd.WorkDir, cmd.Commands)
}

/**
 * 执行 shell 命令列表
 * 根据 shell 类型使用 cmd /k 或 powershell -NoExit -Command 启动新终端窗口
 */
func executeShellCommands(shell, workDir, commands string) error {
	shell = DefaultShell(shell)
	shellPath := ResolveShellPath(shell)

	var cmd *exec.Cmd
	if shell == "powershell" || shell == "powershell.exe" {
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
