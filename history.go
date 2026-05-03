package main

import (
	"fmt"
	"time"
)

type CommandHistory struct {
	Id         int64  `json:"id"`
	Command    string `json:"command"`
	Shell      string `json:"shell"`
	WorkDir    string `json:"workDir"`
	ExecutedAt string `json:"executedAt"`
}

type HistoryResult struct {
	Histories []CommandHistory `json:"histories"`
	Total     int64            `json:"total"`
}

type HistoryService struct{}

func NewHistoryService() *HistoryService {
	return &HistoryService{}
}

/**
 * 记录命令历史
 */
func (h *HistoryService) AddHistory(command, shell, workDir string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	if shell == "" {
		shell = "cmd.exe"
	}
	_, err := db.Exec(
		"INSERT INTO command_history (command, shell, work_dir, executed_at) VALUES (?, ?, ?, ?)",
		command, shell, workDir, now,
	)
	return err
}

/**
 * 获取命令历史列表
 */
func (h *HistoryService) GetHistory(page, pageSize int) (*HistoryResult, error) {
	var total int64
	err := db.QueryRow("SELECT COUNT(*) FROM command_history").Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	rows, err := db.Query(
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
 */
func (h *HistoryService) SearchHistory(keyword string, page, pageSize int) (*HistoryResult, error) {
	pattern := fmt.Sprintf("%%%s%%", keyword)

	var total int64
	err := db.QueryRow("SELECT COUNT(*) FROM command_history WHERE command LIKE ?", pattern).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	rows, err := db.Query(
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
	_, err := db.Exec("DELETE FROM command_history")
	return err
}

/**
 * 删除单条命令历史
 */
func (h *HistoryService) DeleteHistory(id int64) error {
	_, err := db.Exec("DELETE FROM command_history WHERE id = ?", id)
	return err
}
