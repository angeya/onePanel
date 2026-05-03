package main

import (
	"database/sql"
	"time"
)

type ShortcutCategory struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ShortcutCommand struct {
	Id         int64  `json:"id"`
	CategoryId *int64 `json:"categoryId"`
	Name       string `json:"name"`
	Shell      string `json:"shell"`
	WorkDir    string `json:"workDir"`
	Commands   string `json:"commands"`
	SortOrder  int    `json:"sortOrder"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type ShortcutService struct{}

func NewShortcutService() *ShortcutService {
	return &ShortcutService{}
}

/**
 * 获取所有快捷命令分类
 */
func (s *ShortcutService) GetCategories() ([]ShortcutCategory, error) {
	rows, err := db.Query("SELECT id, name, sort_order, created_at, updated_at FROM shortcut_category ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []ShortcutCategory
	for rows.Next() {
		var c ShortcutCategory
		if err := rows.Scan(&c.Id, &c.Name, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

/**
 * 创建快捷命令分类
 */
func (s *ShortcutService) CreateCategory(name string, sortOrder int) (*ShortcutCategory, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	result, err := db.Exec(
		"INSERT INTO shortcut_category (name, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?)",
		name, sortOrder, now, now,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &ShortcutCategory{
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
func (s *ShortcutService) UpdateCategory(id int64, name string, sortOrder int) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		"UPDATE shortcut_category SET name = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		name, sortOrder, now, id,
	)
	return err
}

/**
 * 删除快捷命令分类
 */
func (s *ShortcutService) DeleteCategory(id int64) error {
	_, err := db.Exec("DELETE FROM shortcut_category WHERE id = ?", id)
	return err
}

/**
 * 获取所有快捷命令
 */
func (s *ShortcutService) GetCommands() ([]ShortcutCommand, error) {
	rows, err := db.Query(
		"SELECT id, category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_command ORDER BY sort_order, id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []ShortcutCommand
	for rows.Next() {
		var cmd ShortcutCommand
		var categoryId sql.NullInt64
		if err := rows.Scan(&cmd.Id, &categoryId, &cmd.Name, &cmd.Shell, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt); err != nil {
			return nil, err
		}
		if categoryId.Valid {
			cmd.CategoryId = &categoryId.Int64
		}
		commands = append(commands, cmd)
	}
	return commands, nil
}

/**
 * 根据分类 ID 获取快捷命令
 */
func (s *ShortcutService) GetCommandsByCategory(categoryId int64) ([]ShortcutCommand, error) {
	rows, err := db.Query(
		"SELECT id, category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_command WHERE category_id = ? ORDER BY sort_order, id",
		categoryId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []ShortcutCommand
	for rows.Next() {
		var cmd ShortcutCommand
		var catId sql.NullInt64
		if err := rows.Scan(&cmd.Id, &catId, &cmd.Name, &cmd.Shell, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt); err != nil {
			return nil, err
		}
		if catId.Valid {
			cmd.CategoryId = &catId.Int64
		}
		commands = append(commands, cmd)
	}
	return commands, nil
}

/**
 * 创建快捷命令
 */
func (s *ShortcutService) CreateCommand(categoryId *int64, name, shell, workDir, commands string, sortOrder int) (*ShortcutCommand, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	var catId sql.NullInt64
	if categoryId != nil {
		catId = sql.NullInt64{Int64: *categoryId, Valid: true}
	}

	if shell == "" {
		shell = "cmd.exe"
	}

	result, err := db.Exec(
		"INSERT INTO shortcut_command (category_id, name, shell, work_dir, commands, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		catId, name, shell, workDir, commands, sortOrder, now, now,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &ShortcutCommand{
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
func (s *ShortcutService) UpdateCommand(id int64, categoryId *int64, name, shell, workDir, commands string, sortOrder int) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	var catId sql.NullInt64
	if categoryId != nil {
		catId = sql.NullInt64{Int64: *categoryId, Valid: true}
	}

	if shell == "" {
		shell = "cmd.exe"
	}

	_, err := db.Exec(
		"UPDATE shortcut_command SET category_id = ?, name = ?, shell = ?, work_dir = ?, commands = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		catId, name, shell, workDir, commands, sortOrder, now, id,
	)
	return err
}

/**
 * 删除快捷命令
 */
func (s *ShortcutService) DeleteCommand(id int64) error {
	_, err := db.Exec("DELETE FROM shortcut_command WHERE id = ?", id)
	return err
}
