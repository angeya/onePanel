package main

import (
	"encoding/json"
	"fmt"
)

/**
 * SettingService 系统设置服务
 * 负责系统级配置项的读写操作
 * 通过依赖注入持有 Database 引用，委托 Database 实现 CRUD
 */
type SettingService struct {
	db *Database
}

/**
 * 创建 SettingService 实例
 * 注入 Database 依赖
 */
func NewSettingService(db *Database) *SettingService {
	return &SettingService{db: db}
}

/**
 * 获取设置项
 * 委托给 Database.GetConfig 实现
 */
func (s *SettingService) GetSetting(key string) (string, error) {
	return s.db.GetConfig(key)
}

/**
 * 设置设置项
 * 委托给 Database.SetConfig 实现
 */
func (s *SettingService) SetSetting(key string, value string) error {
	return s.db.SetConfig(key, value)
}

/**
 * 批量获取设置项
 * 委托给 Database.GetConfigs 实现
 */
func (s *SettingService) GetBootstrapSettings() (map[string]string, error) {
	return s.db.GetConfigs([]string{"theme", "default_shell", "close_action", "global_hotkey"})
}

/**
 * GetGlobalHotkey 读取全局快捷键配置。
 * 未设置时返回默认配置。
 */
func (s *SettingService) GetGlobalHotkey() (HotkeyConfig, error) {
	value, err := s.db.GetConfig("global_hotkey")
	if err != nil {
		return DefaultHotkeyConfig(), err
	}
	if value == "" {
		return DefaultHotkeyConfig(), nil
	}

	var config HotkeyConfig
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return DefaultHotkeyConfig(), fmt.Errorf("解析全局快捷键配置失败: %w", err)
	}
	if config.Key == "" {
		return DefaultHotkeyConfig(), nil
	}
	return config, nil
}

/**
 * SetGlobalHotkey 保存全局快捷键配置。
 */
func (s *SettingService) SetGlobalHotkey(config HotkeyConfig) error {
	if config.Key == "" {
		config = DefaultHotkeyConfig()
	}
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化全局快捷键配置失败: %w", err)
	}
	return s.db.SetConfig("global_hotkey", string(data))
}
