package main

import (
	"database/sql"
)

/**
 * ShortcutService 终端快捷命令服务
 * 负责终端侧边栏快捷命令的分类和命令管理
 * 通过依赖注入持有 Database 引用
 */
type ShortcutService struct {
	db *Database
}

/**
 * 创建 ShortcutService 实例
 * 注入 Database 依赖
 */
func NewShortcutService(db *Database) *ShortcutService {
	return &ShortcutService{db: db}
}

/**
 * 获取所有快捷命令分类
 * 按排序字段和 ID 升序排列
 */
func (s *ShortcutService) GetCategories() ([]ShortcutCategory, error) {
	rows, err := s.db.DB().Query("SELECT id, name, sort_order, created_at, updated_at FROM shortcut_category ORDER BY sort_order, id")
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
	now := NowFormatted()
	result, err := s.db.DB().Exec(
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
	now := NowFormatted()
	_, err := s.db.DB().Exec(
		"UPDATE shortcut_category SET name = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		name, sortOrder, now, id,
	)
	return err
}

/**
 * 删除快捷命令分类
 */
func (s *ShortcutService) DeleteCategory(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM shortcut_category WHERE id = ?", id)
	return err
}

/**
 * 获取所有快捷命令
 * 按排序字段和 ID 升序排列
 */
func (s *ShortcutService) GetCommands() ([]ShortcutCommand, error) {
	rows, err := s.db.DB().Query(
		"SELECT id, category_id, name, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_command ORDER BY sort_order, id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []ShortcutCommand
	for rows.Next() {
		var cmd ShortcutCommand
		var categoryId sql.NullInt64
		if err := rows.Scan(&cmd.Id, &categoryId, &cmd.Name, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt); err != nil {
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
	rows, err := s.db.DB().Query(
		"SELECT id, category_id, name, work_dir, commands, sort_order, created_at, updated_at FROM shortcut_command WHERE category_id = ? ORDER BY sort_order, id",
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
		if err := rows.Scan(&cmd.Id, &catId, &cmd.Name, &cmd.WorkDir, &cmd.Commands, &cmd.SortOrder, &cmd.CreatedAt, &cmd.UpdatedAt); err != nil {
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
func (s *ShortcutService) CreateCommand(categoryId *int64, name, workDir, commands string, sortOrder int) (*ShortcutCommand, error) {
	now := NowFormatted()

	var catId sql.NullInt64
	if categoryId != nil {
		catId = sql.NullInt64{Int64: *categoryId, Valid: true}
	}

	result, err := s.db.DB().Exec(
		"INSERT INTO shortcut_command (category_id, name, work_dir, commands, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		catId, name, workDir, commands, sortOrder, now, now,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &ShortcutCommand{
		Id:         id,
		CategoryId: categoryId,
		Name:       name,
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
func (s *ShortcutService) UpdateCommand(id int64, categoryId *int64, name, workDir, commands string, sortOrder int) error {
	now := NowFormatted()

	var catId sql.NullInt64
	if categoryId != nil {
		catId = sql.NullInt64{Int64: *categoryId, Valid: true}
	}

	_, err := s.db.DB().Exec(
		"UPDATE shortcut_command SET category_id = ?, name = ?, work_dir = ?, commands = ?, sort_order = ?, updated_at = ? WHERE id = ?",
		catId, name, workDir, commands, sortOrder, now, id,
	)
	return err
}

/**
 * 删除快捷命令
 */
func (s *ShortcutService) DeleteCommand(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM shortcut_command WHERE id = ?", id)
	return err
}
