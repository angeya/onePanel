package main

import (
	"database/sql"
	"fmt"
	"os/exec"
	"runtime"
)

/**
 * ShortcutCmdService 快速启动命令服务
 * 负责快速启动功能的分组管理、命令 CRUD 和命令执行
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
 * 获取所有快捷命令分组
 * 按排序字段和 ID 升序排列
 */
func (s *ShortcutCmdService) GetGroups() ([]ShortcutCmdGroup, error) {
	rows, err := s.db.DB().Query("SELECT id, name, sort_order, created_at, updated_at FROM shortcut_cmd_group ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []ShortcutCmdGroup
	for rows.Next() {
		var g ShortcutCmdGroup
		if err := rows.Scan(&g.Id, &g.Name, &g.SortOrder, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	if groups == nil {
		groups = []ShortcutCmdGroup{}
	}
	return groups, nil
}

/**
 * 创建快捷命令分组
 */
func (s *ShortcutCmdService) CreateGroup(name string, sortOrder int) (*ShortcutCmdGroup, error) {
	now := NowFormatted()
	result, err := s.db.DB().Exec(
		"INSERT INTO shortcut_cmd_group (name, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?)",
		name, sortOrder, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("创建分组失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return &ShortcutCmdGroup{
		Id:        id,
		Name:      name,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

/**
 * 更新快捷命令分组
 */
func (s *ShortcutCmdService) UpdateGroup(id int64, name string, sortOrder int) error {
	now := NowFormatted()
	_, err := s.db.DB().Exec(
		"UPDATE shortcut_cmd_group SET name = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		name, sortOrder, now, id,
	)
	return err
}

/**
 * 删除快捷命令分组
 */
func (s *ShortcutCmdService) DeleteGroup(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM shortcut_cmd_group WHERE id = ?", id)
	return err
}

/**
 * 获取所有快捷命令
 * 按排序字段和 ID 升序排列
 */
func (s *ShortcutCmdService) GetCommands() ([]ShortcutCmd, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, group_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_cmd ORDER BY sort_order, id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []ShortcutCmd
	for rows.Next() {
		var cmd ShortcutCmd
		var groupId sql.NullInt64
		if err := rows.Scan(&cmd.Id, &groupId, &cmd.Name, &cmd.Shell, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt); err != nil {
			return nil, err
		}
		if groupId.Valid {
			cmd.GroupId = &groupId.Int64
		}
		commands = append(commands, cmd)
	}
	if commands == nil {
		commands = []ShortcutCmd{}
	}
	return commands, nil
}

/**
 * 根据分组获取快捷命令
 */
func (s *ShortcutCmdService) GetCommandsByGroup(groupId int64) ([]ShortcutCmd, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, group_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_cmd WHERE group_id = ? ORDER BY sort_order, id",
		groupId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []ShortcutCmd
	for rows.Next() {
		var cmd ShortcutCmd
		var gId sql.NullInt64
		if err := rows.Scan(&cmd.Id, &gId, &cmd.Name, &cmd.Shell, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt); err != nil {
			return nil, err
		}
		if gId.Valid {
			cmd.GroupId = &gId.Int64
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
func (s *ShortcutCmdService) CreateCommand(groupId *int64, name, shell, workDir, commands string, sortOrder int) (*ShortcutCmd, error) {
	now := NowFormatted()
	shell = DefaultShell(shell)

	var gId sql.NullInt64
	if groupId != nil {
		gId = sql.NullInt64{Int64: *groupId, Valid: true}
	}

	result, err := s.db.DB().Exec(
		"INSERT INTO shortcut_cmd (group_id, name, shell, work_dir, commands, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		gId, name, shell, workDir, commands, sortOrder, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("创建快捷命令失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return &ShortcutCmd{
		Id:        id,
		GroupId:   groupId,
		Name:      name,
		Shell:     shell,
		WorkDir:   workDir,
		Commands:  commands,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

/**
 * 更新快捷命令
 */
func (s *ShortcutCmdService) UpdateCommand(id int64, groupId *int64, name, shell, workDir, commands string, sortOrder int) error {
	now := NowFormatted()
	shell = DefaultShell(shell)

	var gId sql.NullInt64
	if groupId != nil {
		gId = sql.NullInt64{Int64: *groupId, Valid: true}
	}

	_, err := s.db.DB().Exec(
		"UPDATE shortcut_cmd SET group_id = ?, name = ?, shell = ?, work_dir = ?, commands = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		gId, name, shell, workDir, commands, sortOrder, now, id,
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
	var groupId sql.NullInt64
	err := s.db.DB().QueryRow(
		"SELECT id, group_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_cmd WHERE id = ?",
		id,
	).Scan(&cmd.Id, &groupId, &cmd.Name, &cmd.Shell, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt)
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
