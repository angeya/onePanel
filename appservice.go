package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SubApp struct {
	Id          int64  `json:"id"`
	DirName     string `json:"dirName"`
	DisplayName string `json:"displayName"`
	IconPath    string `json:"iconPath"`
	EntryUrl    string `json:"entryUrl"`
	SortOrder   int    `json:"sortOrder"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type AppConfig struct {
	Id          int64  `json:"id"`
	ConfigKey   string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	UpdatedAt   string `json:"updatedAt"`
}

type AppService struct{}

func NewAppService() *AppService {
	return &AppService{}
}

/**
 * 获取静态目录配置
 */
func (a *AppService) GetStaticDir() (string, error) {
	var value string
	err := db.QueryRow("SELECT config_value FROM app_config WHERE config_key = 'static_dir'").Scan(&value)
	if err != nil {
		return "", nil
	}
	return value, nil
}

/**
 * 设置静态目录配置
 */
func (a *AppService) SetStaticDir(dir string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(
		"INSERT INTO app_config (config_key, config_value, updated_at) VALUES ('static_dir', ?, ?) ON CONFLICT(config_key) DO UPDATE SET config_value = ?, updated_at = ?",
		dir, now, dir, now,
	)
	if err != nil {
		return fmt.Errorf("保存静态目录配置失败: %w", err)
	}

	if staticServer != nil && dir == "" {
		staticServer.Stop()
	}

	return nil
}

/**
 * 获取静态服务器状态
 */
func (a *AppService) GetServerStatus() map[string]interface{} {
	if staticServer == nil {
		return map[string]interface{}{
			"running": false,
			"port":    0,
			"dir":     "",
		}
	}
	return staticServer.GetStatus()
}

/**
 * 启动静态服务器
 */
func (a *AppService) StartServer() (int, error) {
	dir, err := a.GetStaticDir()
	if err != nil {
		return 0, err
	}
	if dir == "" {
		return 0, fmt.Errorf("请先设置静态目录")
	}
	if staticServer == nil {
		return 0, fmt.Errorf("静态服务器未初始化")
	}
	return staticServer.Start(dir)
}

/**
 * 停止静态服务器
 */
func (a *AppService) StopServer() error {
	if staticServer == nil {
		return nil
	}
	return staticServer.Stop()
}

/**
 * 打开应用 - 自动启动静态服务（如果未启动）并返回应用 URL
 */
func (a *AppService) OpenApp(appId int64) (map[string]interface{}, error) {
	var app SubApp
	err := db.QueryRow("SELECT id, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app WHERE id = ?", appId).Scan(
		&app.Id, &app.DirName, &app.DisplayName, &app.IconPath, &app.EntryUrl, &app.SortOrder, &app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("应用不存在")
	}

	staticDir, err := a.GetStaticDir()
	if err != nil || staticDir == "" {
		return nil, fmt.Errorf("请先设置静态目录")
	}

	appDir := filepath.Join(staticDir, app.DirName)
	indexFile := filepath.Join(appDir, "index.html")

	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("应用入口文件不存在")
	}

	port, err := a.ensureServerRunning(staticDir)
	if err != nil {
		return nil, err
	}

	iconUrl := ""
	if app.IconPath != "" {
		dir := app.EntryUrl[:len(app.EntryUrl)-len("index.html")]
		iconUrl = fmt.Sprintf("http://127.0.0.1:%d%sicon.png", port, dir)
	}

	return map[string]interface{}{
		"url":      fmt.Sprintf("http://127.0.0.1:%d%s", port, app.EntryUrl),
		"name":     app.DisplayName,
		"iconPath": iconUrl,
	}, nil
}

/**
 * 确保静态服务器已启动
 */
func (a *AppService) ensureServerRunning(dir string) (int, error) {
	if staticServer == nil {
		return 0, fmt.Errorf("静态服务器未初始化")
	}
	status := staticServer.GetStatus()
	if running, ok := status["running"].(bool); ok && running {
		if port, ok := status["port"].(int); ok {
			return port, nil
		}
	}
	return staticServer.Start(dir)
}

/**
 * 扫描静态目录并更新应用列表
 */
func (a *AppService) ScanApps() ([]SubApp, error) {
	dir, err := a.GetStaticDir()
	if err != nil {
		return nil, err
	}
	if dir == "" {
		return nil, fmt.Errorf("请先设置静态目录")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取静态目录失败: %w", err)
	}

	var apps []SubApp
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		subDir := filepath.Join(dir, dirName)

		indexFile := filepath.Join(subDir, "index.html")
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			continue
		}

		displayName := dirName
		nameFile := filepath.Join(subDir, dirName+".name")
		if data, err := os.ReadFile(nameFile); err == nil {
			name := strings.TrimSpace(string(data))
			if name != "" {
				displayName = name
			}
		}

		iconPath := ""
		iconFile := filepath.Join(subDir, "icon.png")
		if _, err := os.Stat(iconFile); err == nil {
			iconPath = iconFile
		}

		entryUrl := fmt.Sprintf("/%s/index.html", dirName)

		var existing SubApp
		err := db.QueryRow("SELECT id, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app WHERE dir_name = ?", dirName).Scan(
			&existing.Id, &existing.DirName, &existing.DisplayName, &existing.IconPath, &existing.EntryUrl, &existing.SortOrder, &existing.CreatedAt, &existing.UpdatedAt,
		)

		if err != nil {
			now := time.Now().Format("2006-01-02 15:04:05")
			result, err := db.Exec(
				"INSERT INTO sub_app (dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, 0, ?, ?)",
				dirName, displayName, iconPath, entryUrl, now, now,
			)
			if err != nil {
				continue
			}
			id, _ := result.LastInsertId()
			apps = append(apps, SubApp{
				Id:          id,
				DirName:     dirName,
				DisplayName: displayName,
				IconPath:    iconPath,
				EntryUrl:    entryUrl,
				SortOrder:   0,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		} else {
			if existing.DisplayName == existing.DirName || existing.DisplayName != displayName {
				now := time.Now().Format("2006-01-02 15:04:05")
				db.Exec("UPDATE sub_app SET display_name = ?, icon_path = ?, entry_url = ?, updated_at = ? WHERE id = ?",
					displayName, iconPath, entryUrl, now, existing.Id)
				existing.DisplayName = displayName
				existing.IconPath = iconPath
				existing.EntryUrl = entryUrl
			}
			apps = append(apps, existing)
		}
	}

	if apps == nil {
		apps = []SubApp{}
	}
	return apps, nil
}

/**
 * 获取所有应用列表
 */
func (a *AppService) GetApps() ([]SubApp, error) {
	rows, err := db.Query("SELECT id, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []SubApp
	for rows.Next() {
		var app SubApp
		if err := rows.Scan(&app.Id, &app.DirName, &app.DisplayName, &app.IconPath, &app.EntryUrl, &app.SortOrder, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	if apps == nil {
		apps = []SubApp{}
	}
	return apps, nil
}

/**
 * 更新应用显示名称
 */
func (a *AppService) UpdateDisplayName(id int64, name string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec("UPDATE sub_app SET display_name = ?, updated_at = ? WHERE id = ?", name, now, id)
	return err
}

/**
 * 更新应用目录名称
 */
func (a *AppService) UpdateDirName(id int64, newDirName string) error {
	var oldDirName string
	var iconPath string
	err := db.QueryRow("SELECT dir_name, icon_path FROM sub_app WHERE id = ?", id).Scan(&oldDirName, &iconPath)
	if err != nil {
		return err
	}

	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("静态目录未设置")
	}

	oldPath := filepath.Join(staticDir, oldDirName)
	newPath := filepath.Join(staticDir, newDirName)

	if oldPath != newPath {
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("目录 %s 已存在", newDirName)
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("重命名目录失败: %w", err)
		}

		if iconPath != "" {
			newIconPath := filepath.Join(newPath, "icon.png")
			if _, err := os.Stat(newIconPath); err == nil {
				iconPath = newIconPath
			}
		}
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	entryUrl := fmt.Sprintf("/%s/index.html", newDirName)
	_, err = db.Exec("UPDATE sub_app SET dir_name = ?, entry_url = ?, icon_path = ?, updated_at = ? WHERE id = ?",
		newDirName, entryUrl, iconPath, now, id)
	return err
}

/**
 * 上传应用图标
 */
func (a *AppService) UploadIcon(id int64, iconData []byte) error {
	var dirName string
	err := db.QueryRow("SELECT dir_name FROM sub_app WHERE id = ?", id).Scan(&dirName)
	if err != nil {
		return err
	}

	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("静态目录未设置")
	}

	iconPath := filepath.Join(staticDir, dirName, "icon.png")
	if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
		return fmt.Errorf("保存图标失败: %w", err)
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = db.Exec("UPDATE sub_app SET icon_path = ?, updated_at = ? WHERE id = ?", iconPath, now, id)
	return err
}

/**
 * 删除应用
 */
func (a *AppService) DeleteApp(id int64) error {
	var dirName string
	err := db.QueryRow("SELECT dir_name FROM sub_app WHERE id = ?", id).Scan(&dirName)
	if err != nil {
		return err
	}

	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir != "" {
		appDir := filepath.Join(staticDir, dirName)
		os.RemoveAll(appDir)
	}

	_, err = db.Exec("DELETE FROM sub_app WHERE id = ?", id)
	return err
}

/**
 * 导出应用为 zip 压缩包
 */
func (a *AppService) ExportApp(id int64) (string, error) {
	var dirName string
	err := db.QueryRow("SELECT dir_name FROM sub_app WHERE id = ?", id).Scan(&dirName)
	if err != nil {
		return "", err
	}

	staticDir, err := a.GetStaticDir()
	if err != nil {
		return "", err
	}
	if staticDir == "" {
		return "", fmt.Errorf("静态目录未设置")
	}

	appDir := filepath.Join(staticDir, dirName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return "", fmt.Errorf("应用目录不存在")
	}

	timestamp := time.Now().Format("20060102150405")
	zipName := fmt.Sprintf("%s_%s.zip", dirName, timestamp)
	zipPath := filepath.Join(os.TempDir(), zipName)

	if err := zipDirectory(appDir, zipPath); err != nil {
		return "", fmt.Errorf("压缩失败: %w", err)
	}

	return zipPath, nil
}

/**
 * 导入 zip 压缩包应用
 */
func (a *AppService) ImportZip(zipPath string) error {
	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("请先设置静态目录")
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开压缩包失败: %w", err)
	}
	defer reader.Close()

	var topDir string
	for _, f := range reader.File {
		parts := strings.SplitN(f.Name, "/", 2)
		if parts[0] != "" && (topDir == "" || topDir == parts[0]) {
			topDir = parts[0]
		}
		if topDir != "" && parts[0] != topDir {
			topDir = ""
			break
		}
	}

	if topDir == "" {
		baseName := strings.TrimSuffix(filepath.Base(zipPath), ".zip")
		if idx := strings.LastIndex(baseName, "_"); idx > 0 {
			topDir = baseName[:idx]
		} else {
			topDir = baseName
		}
	}

	destDir := filepath.Join(staticDir, topDir)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("应用目录 %s 已存在", topDir)
	}

	for _, f := range reader.File {
		fpath := filepath.Join(staticDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	_, err = a.ScanApps()
	return err
}

/**
 * 导入 HTML 目录作为应用
 */
func (a *AppService) ImportDir(srcDir string, appName string) error {
	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("请先设置静态目录")
	}

	dirName := appName
	if dirName == "" {
		dirName = filepath.Base(srcDir)
	}
	dirName = sanitizeDirName(dirName)

	destDir := filepath.Join(staticDir, dirName)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("应用目录 %s 已存在", dirName)
	}

	if err := copyDirectory(srcDir, destDir); err != nil {
		return fmt.Errorf("复制目录失败: %w", err)
	}

	_, err = a.ScanApps()
	return err
}

/**
 * 压缩目录为 zip 文件
 */
func zipDirectory(srcDir, zipPath string) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	baseDir := filepath.Base(srcDir)
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		zipPath := filepath.Join(baseDir, relPath)
		zipPath = filepath.ToSlash(zipPath)

		if info.IsDir() {
			_, err := w.Create(zipPath + "/")
			return err
		}

		fileWriter, err := w.Create(zipPath)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(fileWriter, f)
		return err
	})
}

/**
 * 复制目录
 */
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

/**
 * 清理目录名称中的非法字符
 */
func sanitizeDirName(name string) string {
	invalid := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, ch := range invalid {
		name = strings.ReplaceAll(name, ch, "_")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "app_" + time.Now().Format("20060102150405")
	}
	return name
}
