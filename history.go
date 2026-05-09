package main

import (
	"fmt"
)

/**
 * HistoryService 命令历史服务
 * 负责命令执行历史的记录、查询、搜索和删除
 * 通过依赖注入持有 Database 引用
 */
type HistoryService struct {
	db *Database
}

/**
 * 创建 HistoryService 实例
 * 注入 Database 依赖
 */
func NewHistoryService(db *Database) *HistoryService {
	return &HistoryService{db: db}
}

/**
 * 记录命令历史
 * shell 为空时默认使用 cmd.exe
 */
func (h *HistoryService) AddHistory(command, shell, workDir string) error {
	now := NowFormatted()
	shell = DefaultShell(shell)
	_, err := h.db.DB().Exec(
		"INSERT INTO command_history (command, shell, work_dir, executed_at) VALUES (?, ?, ?, ?)",
		command, shell, workDir, now,
	)
	return err
}

/**
 * 获取命令历史列表
 * 按执行时间倒序分页查询
 */
func (h *HistoryService) GetHistory(page, pageSize int) (*HistoryResult, error) {
	var total int64
	err := h.db.DB().QueryRow("SELECT COUNT(*) FROM command_history").Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	rows, err := h.db.DB().Query(
		"SELECT id, command, shell, work_dir, executed_at FROM command_history ORDER BY executed_at DESC LIMIT ? OFFSET ?",
		pageSize, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []CommandHistory
	for rows.Next() {
		var ch CommandHistory
		if err := rows.Scan(&ch.Id, &ch.Command, &ch.Shell, &ch.WorkDir, &ch.ExecutedAt); err != nil {
			return nil, err
		}
		histories = append(histories, ch)
	}

	if histories == nil {
		histories = []CommandHistory{}
	}

	return &HistoryResult{
		Histories: histories,
		Total:     total,
	}, nil
}

/**
 * 搜索命令历史
 * 支持按关键字模糊搜索，按执行时间倒序分页
 */
func (h *HistoryService) SearchHistory(keyword string, page, pageSize int) (*HistoryResult, error) {
	pattern := fmt.Sprintf("%%%s%%", keyword)

	var total int64
	err := h.db.DB().QueryRow("SELECT COUNT(*) FROM command_history WHERE command LIKE ?", pattern).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	rows, err := h.db.DB().Query(
		"SELECT id, command, shell, work_dir, executed_at FROM command_history WHERE command LIKE ? ORDER BY executed_at DESC LIMIT ? OFFSET ?",
		pattern, pageSize, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []CommandHistory
	for rows.Next() {
		var ch CommandHistory
		if err := rows.Scan(&ch.Id, &ch.Command, &ch.Shell, &ch.WorkDir, &ch.ExecutedAt); err != nil {
			return nil, err
		}
		histories = append(histories, ch)
	}

	if histories == nil {
		histories = []CommandHistory{}
	}

	return &HistoryResult{
		Histories: histories,
		Total:     total,
	}, nil
}

/**
 * 清空命令历史
 */
func (h *HistoryService) ClearHistory() error {
	_, err := h.db.DB().Exec("DELETE FROM command_history")
	return err
}

/**
 * 删除单条命令历史
 */
func (h *HistoryService) DeleteHistory(id int64) error {
	_, err := h.db.DB().Exec("DELETE FROM command_history WHERE id = ?", id)
	return err
}
