package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

/**
 * AppService 我的应用服务
 * 负责子应用的增删改查、导入导出、静态目录管理等业务逻辑
 * 通过依赖注入持有 Database 和 StaticServer 引用
 */
type AppService struct {
	db     *Database
	server *StaticServer
}

/**
 * 创建 AppService 实例
 * 注入 Database 和 StaticServer 依赖
 */
func NewAppService(db *Database, server *StaticServer) *AppService {
	return &AppService{db: db, server: server}
}

/**
 * 获取静态目录配置
 * 如果未配置自定义目录，则使用可执行文件同级目录下的 apps 目录作为默认值
 */
func (a *AppService) GetStaticDir() (string, error) {
	dir, err := a.db.GetConfig("static_dir")
	if err != nil {
		return "", err
	}
	if dir != "" {
		return dir, nil
	}
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exePath), "apps"), nil
}

/**
 * 设置自定义应用目录
 * 清空目录表示恢复使用默认目录（exe 同级 apps 目录）
 * 如果目录发生变化，自动停止静态服务器以便下次使用新目录启动
 */
func (a *AppService) SetStaticDir(dir string) error {
	if err := a.db.SetConfig("static_dir", dir); err != nil {
		return fmt.Errorf("保存静态目录配置失败: %w", err)
	}

	if a.server != nil {
		a.server.Stop()
	}

	return nil
}

/**
 * 获取静态服务器状态
 */
func (a *AppService) GetServerStatus() map[string]interface{} {
	if a.server == nil {
		return map[string]interface{}{
			"running": false,
			"port":    0,
			"dir":     "",
		}
	}
	return a.server.GetStatus()
}

/**
 * 启动静态服务器
 * 使用已配置的静态目录启动服务
 */
func (a *AppService) StartServer() (int, error) {
	dir, err := a.GetStaticDir()
	if err != nil {
		return 0, err
	}
	if dir == "" {
		return 0, fmt.Errorf("获取应用目录失败")
	}
	if a.server == nil {
		return 0, fmt.Errorf("静态服务器未初始化")
	}
	return a.server.Start(dir)
}

/**
 * 停止静态服务器
 */
func (a *AppService) StopServer() error {
	if a.server == nil {
		return nil
	}
	return a.server.Stop()
}

/**
 * 打开应用
 * 自动启动静态服务（如果未启动）并返回应用 URL
 * 对于 web 类型应用直接返回 URL
 */
func (a *AppService) OpenApp(appId int64) (map[string]interface{}, error) {
	var app SubApp
	err := a.db.DB().QueryRow(
		"SELECT id, app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app WHERE id = ?",
		appId,
	).Scan(&app.Id, &app.AppType, &app.DirName, &app.DisplayName, &app.IconPath, &app.EntryUrl, &app.SortOrder, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("应用不存在")
	}

	if app.AppType == "web" {
		return map[string]interface{}{
			"url":      app.EntryUrl,
			"name":     app.DisplayName,
			"iconPath": "",
		}, nil
	}

	staticDir, err := a.GetStaticDir()
	if err != nil || staticDir == "" {
		return nil, fmt.Errorf("获取应用目录失败")
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
 * 如果已运行则复用现有端口，否则启动新实例
 */
func (a *AppService) ensureServerRunning(dir string) (int, error) {
	if a.server == nil {
		return 0, fmt.Errorf("静态服务器未初始化")
	}
	if a.server.IsRunning() {
		return a.server.Port(), nil
	}
	return a.server.Start(dir)
}

/**
 * 扫描静态目录并更新应用列表
 * 读取静态目录下的子目录，识别包含 index.html 的应用
 * 自动创建或更新数据库中的应用记录
 */
func (a *AppService) ScanApps() ([]SubApp, error) {
	dir, err := a.GetStaticDir()
	if err != nil {
		return nil, err
	}
	if dir == "" {
		return nil, fmt.Errorf("获取应用目录失败")
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

		displayName := a.readAppName(subDir, dirName)
		iconPath := a.resolveIconPath(subDir)
		entryUrl := fmt.Sprintf("/%s/index.html", dirName)

		app, created := a.upsertStaticApp(dirName, displayName, iconPath, entryUrl)
		if created || app.Id > 0 {
			apps = append(apps, app)
		}
	}

	if apps == nil {
		apps = []SubApp{}
	}
	return apps, nil
}

/**
 * 读取应用名称
 * 优先从 xxx.name 文件读取，否则使用目录名
 */
func (a *AppService) readAppName(subDir, dirName string) string {
	nameFile := filepath.Join(subDir, dirName+".name")
	if data, err := os.ReadFile(nameFile); err == nil {
		name := strings.TrimSpace(string(data))
		if name != "" {
			return name
		}
	}
	return dirName
}

/**
 * 解析应用图标路径
 * 检查子目录下是否存在 icon.png
 */
func (a *AppService) resolveIconPath(subDir string) string {
	iconFile := filepath.Join(subDir, "icon.png")
	if _, err := os.Stat(iconFile); err == nil {
		return iconFile
	}
	return ""
}

/**
 * 创建或更新静态应用记录
 * 如果目录名对应的记录已存在则更新，否则新建
 */
func (a *AppService) upsertStaticApp(dirName, displayName, iconPath, entryUrl string) (SubApp, bool) {
	var existing SubApp
	err := a.db.DB().QueryRow(
		"SELECT id, app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app WHERE dir_name = ? AND app_type = 'static'",
		dirName,
	).Scan(&existing.Id, &existing.AppType, &existing.DirName, &existing.DisplayName, &existing.IconPath, &existing.EntryUrl, &existing.SortOrder, &existing.CreatedAt, &existing.UpdatedAt)

	if err != nil {
		now := NowFormatted()
		result, err := a.db.DB().Exec(
			"INSERT INTO sub_app (app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at) VALUES ('static', ?, ?, ?, ?, 0, ?, ?)",
			dirName, displayName, iconPath, entryUrl, now, now,
		)
		if err != nil {
			return SubApp{}, false
		}
		id, _ := result.LastInsertId()
		return SubApp{
			Id:          id,
			AppType:     "static",
			DirName:     dirName,
			DisplayName: displayName,
			IconPath:    iconPath,
			EntryUrl:    entryUrl,
			SortOrder:   0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, true
	}

	if existing.DisplayName == existing.DirName || existing.DisplayName != displayName {
		now := NowFormatted()
		a.db.DB().Exec(
			"UPDATE sub_app SET display_name = ?, icon_path = ?, entry_url = ?, updated_at = ? WHERE id = ?",
			displayName, iconPath, entryUrl, now, existing.Id,
		)
		existing.DisplayName = displayName
		existing.IconPath = iconPath
		existing.EntryUrl = entryUrl
	}

	return existing, false
}

/**
 * 获取所有应用列表
 * 按排序字段和 ID 升序排列
 */
func (a *AppService) GetApps() ([]SubApp, error) {
	rows, err := a.db.DB().Query("SELECT id, app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []SubApp
	for rows.Next() {
		var app SubApp
		if err := rows.Scan(&app.Id, &app.AppType, &app.DirName, &app.DisplayName, &app.IconPath, &app.EntryUrl, &app.SortOrder, &app.CreatedAt, &app.UpdatedAt); err != nil {
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
 * 创建网页应用
 * 验证名称和 URL 非空后插入数据库
 */
func (a *AppService) CreateWebApp(name string, url string) (*SubApp, error) {
	if name == "" {
		return nil, fmt.Errorf("应用名称不能为空")
	}
	if url == "" {
		return nil, fmt.Errorf("应用地址不能为空")
	}

	now := NowFormatted()
	result, err := a.db.DB().Exec(
		"INSERT INTO sub_app (app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at) VALUES ('web', '', ?, '', ?, 0, ?, ?)",
		name, url, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("创建网页应用失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return &SubApp{
		Id:          id,
		AppType:     "web",
		DirName:     "",
		DisplayName: name,
		IconPath:    "",
		EntryUrl:    url,
		SortOrder:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

/**
 * 更新网页应用
 * 仅允许更新 web 类型的应用
 */
func (a *AppService) UpdateWebApp(id int64, name string, url string) error {
	if name == "" {
		return fmt.Errorf("应用名称不能为空")
	}
	if url == "" {
		return fmt.Errorf("应用地址不能为空")
	}

	now := NowFormatted()
	_, err := a.db.DB().Exec(
		"UPDATE sub_app SET display_name = ?, entry_url = ?, updated_at = ? WHERE id = ? AND app_type = 'web'",
		name, url, now, id,
	)
	return err
}

/**
 * 更新应用显示名称
 */
func (a *AppService) UpdateDisplayName(id int64, name string) error {
	now := NowFormatted()
	_, err := a.db.DB().Exec("UPDATE sub_app SET display_name = ?, updated_at = ? WHERE id = ?", name, now, id)
	return err
}

/**
 * 更新应用目录名称
 * 同时重命名文件系统中的目录，并更新图标路径和入口 URL
 */
func (a *AppService) UpdateDirName(id int64, newDirName string) error {
	var oldDirName string
	var iconPath string
	err := a.db.DB().QueryRow("SELECT dir_name, icon_path FROM sub_app WHERE id = ?", id).Scan(&oldDirName, &iconPath)
	if err != nil {
		return err
	}

	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("获取应用目录失败")
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

	now := NowFormatted()
	entryUrl := fmt.Sprintf("/%s/index.html", newDirName)
	_, err = a.db.DB().Exec(
		"UPDATE sub_app SET dir_name = ?, entry_url = ?, icon_path = ?, updated_at = ? WHERE id = ?",
		newDirName, entryUrl, iconPath, now, id,
	)
	return err
}

/**
 * 上传应用图标
 * 将图标数据写入应用目录下的 icon.png 文件
 */
func (a *AppService) UploadIcon(id int64, iconData []byte) error {
	var dirName string
	err := a.db.DB().QueryRow("SELECT dir_name FROM sub_app WHERE id = ?", id).Scan(&dirName)
	if err != nil {
		return err
	}

	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("获取应用目录失败")
	}

	iconPath := filepath.Join(staticDir, dirName, "icon.png")
	if err := os.WriteFile(iconPath, iconData, 0644); err != nil {
		return fmt.Errorf("保存图标失败: %w", err)
	}

	now := NowFormatted()
	_, err = a.db.DB().Exec("UPDATE sub_app SET icon_path = ?, updated_at = ? WHERE id = ?", iconPath, now, id)
	return err
}

/**
 * 删除应用
 * 对于静态应用，同时删除文件系统中的应用目录
 */
func (a *AppService) DeleteApp(id int64) error {
	var appType string
	var dirName string
	err := a.db.DB().QueryRow("SELECT app_type, dir_name FROM sub_app WHERE id = ?", id).Scan(&appType, &dirName)
	if err != nil {
		return err
	}

	if appType == "static" {
		staticDir, err := a.GetStaticDir()
		if err != nil {
			return err
		}
		if staticDir != "" && dirName != "" {
			appDir := filepath.Join(staticDir, dirName)
			os.RemoveAll(appDir)
		}
	}

	_, err = a.db.DB().Exec("DELETE FROM sub_app WHERE id = ?", id)
	return err
}

/**
 * 导出应用为 zip 压缩包
 * 生成 "目录名_时间戳.zip" 格式的压缩文件到临时目录
 */
func (a *AppService) ExportApp(id int64) (string, error) {
	var dirName string
	err := a.db.DB().QueryRow("SELECT dir_name FROM sub_app WHERE id = ?", id).Scan(&dirName)
	if err != nil {
		return "", err
	}

	staticDir, err := a.GetStaticDir()
	if err != nil {
		return "", err
	}
	if staticDir == "" {
		return "", fmt.Errorf("获取应用目录失败")
	}

	appDir := filepath.Join(staticDir, dirName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return "", fmt.Errorf("应用目录不存在")
	}

	zipName := fmt.Sprintf("%s_%s.zip", dirName, NowCompactFormatted())
	zipPath := filepath.Join(os.TempDir(), zipName)

	if err := ZipDirectory(appDir, zipPath); err != nil {
		return "", fmt.Errorf("压缩失败: %w", err)
	}

	return zipPath, nil
}

/**
 * 导入 zip 压缩包应用
 * 解压到静态目录并扫描更新应用列表
 */
func (a *AppService) ImportZip(zipPath string) error {
	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("获取应用目录失败")
	}

	if err := a.extractZipToStaticDir(zipPath, staticDir); err != nil {
		return err
	}

	_, err = a.ScanApps()
	return err
}

/**
 * 导入 HTML 目录作为应用
 * 复制源目录到静态目录并扫描更新应用列表
 */
func (a *AppService) ImportDir(srcDir string, appName string) error {
	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("获取应用目录失败")
	}

	dirName := appName
	if dirName == "" {
		dirName = filepath.Base(srcDir)
	}
	dirName = SanitizeDirName(dirName)

	destDir := filepath.Join(staticDir, dirName)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("应用目录 %s 已存在", dirName)
	}

	if err := CopyDirectory(srcDir, destDir); err != nil {
		return fmt.Errorf("复制目录失败: %w", err)
	}

	_, err = a.ScanApps()
	return err
}

/**
 * 解压 ZIP 文件到静态目录
 * 自动识别 ZIP 内的顶层目录结构
 */
func (a *AppService) extractZipToStaticDir(zipPath, staticDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开压缩包失败: %w", err)
	}
	defer reader.Close()

	topDir := a.detectTopDir(reader, zipPath)

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

	return nil
}

/**
 * 检测 ZIP 文件中的顶层目录名
 * 优先从 ZIP 结构推断，否则从文件名提取
 */
func (a *AppService) detectTopDir(reader *zip.ReadCloser, zipPath string) string {
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

	return topDir
}
