package main

import (
	"fmt"
	"time"
)

type SettingService struct{}

func NewSettingService() *SettingService {
	return &SettingService{}
}

/**
 * 获取设置项
 */
func (s *SettingService) GetSetting(key string) (string, error) {
	var value string
	err := db.QueryRow("SELECT config_value FROM app_config WHERE config_key = ?", key).Scan(&value)
	if err != nil {
		return "", nil
	}
	return value, nil
}

/**
 * 设置设置项
 */
func (s *SettingService) SetSetting(key string, value string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		"INSERT INTO app_config (config_key, config_value, updated_at) VALUES (?, ?, ?) ON CONFLICT(config_key) DO UPDATE SET config_value = ?, updated_at = ?",
		key, value, now, value, now,
	)
	if err != nil {
		return fmt.Errorf("保存设置失败: %w", err)
	}
	return nil
}

/**
 * 批量获取设置项
 */
func (s *SettingService) GetSettings(keys []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, key := range keys {
		val, err := s.GetSetting(key)
		if err != nil {
			return nil, err
		}
		if val != "" {
			result[key] = val
		}
	}
	return result, nil
}
