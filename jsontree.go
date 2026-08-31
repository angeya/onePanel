package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
)

/**
 * JsonTreeService JSON 树工具服务。
 * 负责 JSON 树配置的增删改查，以及 JSON 数据内容的读取与校验。
 */
type JsonTreeService struct {
	db *Database
}

/**
 * 创建 JsonTreeService 实例。
 */
func NewJsonTreeService(db *Database) *JsonTreeService {
	return &JsonTreeService{db: db}
}

/**
 * GetJsonTrees 获取所有 JSON 树配置列表。
 * 列表不含大文本内容，避免不必要的传输。
 */
func (s *JsonTreeService) GetJsonTrees() ([]JsonTree, error) {
	rows, err := s.db.DB().Query(`
		SELECT id, name, source_type, file_path, tree_type, id_field, parent_field, children_field, label_field, created_at, updated_at
		FROM json_tree
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("查询 JSON 树列表失败: %w", err)
	}
	defer rows.Close()

	var trees []JsonTree
	for rows.Next() {
		var tree JsonTree
		if err := rows.Scan(
			&tree.Id,
			&tree.Name,
			&tree.SourceType,
			&tree.FilePath,
			&tree.TreeType,
			&tree.IdField,
			&tree.ParentField,
			&tree.ChildrenField,
			&tree.LabelField,
			&tree.CreatedAt,
			&tree.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("读取 JSON 树记录失败: %w", err)
		}
		trees = append(trees, tree)
	}
	if trees == nil {
		trees = []JsonTree{}
	}
	return trees, nil
}

/**
 * GetJsonTree 根据 ID 获取 JSON 树配置，包含完整内容。
 */
func (s *JsonTreeService) GetJsonTree(id int64) (*JsonTree, error) {
	var tree JsonTree
	err := s.db.DB().QueryRow(`
		SELECT id, name, source_type, content, file_path, tree_type, id_field, parent_field, children_field, label_field, created_at, updated_at
		FROM json_tree WHERE id = ?
	`, id).Scan(
		&tree.Id,
		&tree.Name,
		&tree.SourceType,
		&tree.Content,
		&tree.FilePath,
		&tree.TreeType,
		&tree.IdField,
		&tree.ParentField,
		&tree.ChildrenField,
		&tree.LabelField,
		&tree.CreatedAt,
		&tree.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("JSON 树配置不存在")
		}
		return nil, fmt.Errorf("查询 JSON 树失败: %w", err)
	}
	return &tree, nil
}

/**
 * SaveJsonTree 保存 JSON 树配置。
 * id 为 0 时新增，否则更新。保存前会校验配置与 JSON 数据的合法性。
 */
func (s *JsonTreeService) SaveJsonTree(tree *JsonTree) (*JsonTree, error) {
	if tree.Name == "" {
		return nil, fmt.Errorf("树名称不能为空")
	}
	if tree.SourceType != "text" && tree.SourceType != "file" {
		return nil, fmt.Errorf("数据来源类型不合法")
	}
	if tree.TreeType != "logic" && tree.TreeType != "structure" {
		return nil, fmt.Errorf("树结构类型不合法")
	}
	if tree.SourceType == "file" && tree.FilePath == "" {
		return nil, fmt.Errorf("文件路径不能为空")
	}
	if tree.SourceType == "text" && tree.Content == "" {
		return nil, fmt.Errorf("JSON 文本内容不能为空")
	}

	// 校验 JSON 数据与字段配置是否匹配
	if _, _, err := s.loadJsonData(tree); err != nil {
		return nil, err
	}

	now := NowFormatted()
	if tree.Id == 0 {
		result, err := s.db.DB().Exec(`
			INSERT INTO json_tree (name, source_type, content, file_path, tree_type, id_field, parent_field, children_field, label_field, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			tree.Name,
			tree.SourceType,
			tree.Content,
			tree.FilePath,
			tree.TreeType,
			tree.IdField,
			tree.ParentField,
			tree.ChildrenField,
			tree.LabelField,
			now,
			now,
		)
		if err != nil {
			return nil, fmt.Errorf("新增 JSON 树失败: %w", err)
		}
		id, _ := result.LastInsertId()
		tree.Id = id
		tree.CreatedAt = now
		tree.UpdatedAt = now
		return tree, nil
	}

	_, err := s.db.DB().Exec(`
		UPDATE json_tree
		SET name = ?, source_type = ?, content = ?, file_path = ?, tree_type = ?, id_field = ?, parent_field = ?, children_field = ?, label_field = ?, updated_at = ?
		WHERE id = ?
	`,
		tree.Name,
		tree.SourceType,
		tree.Content,
		tree.FilePath,
		tree.TreeType,
		tree.IdField,
		tree.ParentField,
		tree.ChildrenField,
		tree.LabelField,
		now,
		tree.Id,
	)
	if err != nil {
		return nil, fmt.Errorf("更新 JSON 树失败: %w", err)
	}
	tree.UpdatedAt = now
	return tree, nil
}

/**
 * DeleteJsonTree 删除指定 JSON 树配置。
 */
func (s *JsonTreeService) DeleteJsonTree(id int64) error {
	_, err := s.db.DB().Exec("DELETE FROM json_tree WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("删除 JSON 树失败: %w", err)
	}
	return nil
}

/**
 * GetJsonTreeData 读取 JSON 树的原始 JSON 文本。
 * 文本来源直接返回存储内容，文件来源读取指定文件。
 */
func (s *JsonTreeService) GetJsonTreeData(id int64) (string, error) {
	tree, err := s.GetJsonTree(id)
	if err != nil {
		return "", err
	}
	content, _, err := s.loadJsonData(tree)
	if err != nil {
		return "", err
	}
	return content, nil
}

/**
 * loadJsonData 加载并校验 JSON 数据。
 * 返回 JSON 文本与解析后的任意结构数据，同时根据树类型校验字段配置。
 */
func (s *JsonTreeService) loadJsonData(tree *JsonTree) (string, interface{}, error) {
	var content string
	if tree.SourceType == "file" {
		data, err := os.ReadFile(tree.FilePath)
		if err != nil {
			return "", nil, fmt.Errorf("读取 JSON 文件失败: %w", err)
		}
		content = string(data)
	} else {
		content = tree.Content
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return "", nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	if tree.TreeType == "structure" {
		if err := validateStructureTree(parsed); err != nil {
			return "", nil, err
		}
	} else if err := validateLogicTree(parsed); err != nil {
		return "", nil, err
	}

	return content, parsed, nil
}

/**
 * validateStructureTree 校验结构树数据。
 * 要求根数据为数组，数组元素为对象；子节点字段存在时也必须为数组。
 */
func validateStructureTree(root interface{}) error {
	items, ok := root.([]interface{})
	if !ok {
		return fmt.Errorf("结构树数据根节点必须是 JSON 数组")
	}
	var walk func(node interface{}, path string) error
	walk = func(node interface{}, path string) error {
		obj, ok := node.(map[string]interface{})
		if !ok {
			return fmt.Errorf("节点 %s 必须是 JSON 对象", path)
		}
		for key, value := range obj {
			child, isObj := value.([]interface{})
			if isObj {
				for i, c := range child {
					if err := walk(c, fmt.Sprintf("%s.%s[%d]", path, key, i)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	for i, item := range items {
		if err := walk(item, fmt.Sprintf("root[%d]", i)); err != nil {
			return err
		}
	}
	return nil
}

/**
 * validateLogicTree 校验逻辑树数据。
 * 要求根数据为平铺的节点对象数组，数组元素为对象。
 */
func validateLogicTree(root interface{}) error {
	items, ok := root.([]interface{})
	if !ok {
		return fmt.Errorf("逻辑树数据根节点必须是 JSON 数组（平铺的节点列表）")
	}
	for i, item := range items {
		if _, ok := item.(map[string]interface{}); !ok {
			return fmt.Errorf("逻辑树节点 root[%d] 必须是 JSON 对象", i)
		}
	}
	return nil
}
