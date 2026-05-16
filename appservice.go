package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

/**
 * 导出清单文件版本号
 */
const ManifestVersion = "1.0"

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
 * ExportManifest 导出清单结构
 * 记录导出包的版本号、包含的本地应用目录名列表和网页应用信息列表
 */
type ExportManifest struct {
	Version string       `json:"version"`
	Apps    []string     `json:"apps"`
	WebApps []WebAppInfo `json:"webApps"`
}

/**
 * WebAppInfo 网页应用导出信息
 * 用于在 zip 中保存网页应用的名称和地址
 */
type WebAppInfo struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

/**
 * 创建 AppService 实例
 * 注入 Database 和 StaticServer 依赖
 */
func NewAppService(db *Database, server *StaticServer) *AppService {
	return &AppService{db: db, server: server}
}

/**
 * 获取应用目录
 * 固定使用可执行文件同级目录下的 apps 目录
 */
func (a *AppService) GetStaticDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exePath), "apps"), nil
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

	a.db.DB().Exec("DELETE FROM sub_app WHERE app_type = 'static'")

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SubApp{}, nil
		}
		return nil, fmt.Errorf("读取应用目录失败: %w", err)
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

	allApps, err := a.GetApps()
	if err != nil {
		return apps, nil
	}
	return allApps, nil
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
 * 对于静态应用，同时重命名文件系统中的目录，并更新目录名、图标路径和入口 URL
 * 对于网页应用，仅更新显示名称
 */
func (a *AppService) UpdateDisplayName(id int64, name string) error {
	if name == "" {
		return fmt.Errorf("应用名称不能为空")
	}

	var appType string
	var oldDirName string
	var iconPath string
	err := a.db.DB().QueryRow("SELECT app_type, dir_name, icon_path FROM sub_app WHERE id = ?", id).Scan(&appType, &oldDirName, &iconPath)
	if err != nil {
		return fmt.Errorf("应用不存在")
	}

	now := NowFormatted()

	if appType == "static" {
		newDirName := SanitizeDirName(name)
		if newDirName == "" {
			return fmt.Errorf("应用名称无效")
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

		entryUrl := fmt.Sprintf("/%s/index.html", newDirName)
		_, err = a.db.DB().Exec(
			"UPDATE sub_app SET display_name = ?, dir_name = ?, entry_url = ?, icon_path = ?, updated_at = ? WHERE id = ?",
			name, newDirName, entryUrl, iconPath, now, id,
		)
		return err
	}

	_, err = a.db.DB().Exec("UPDATE sub_app SET display_name = ?, updated_at = ? WHERE id = ?", name, now, id)
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
 * 导入 HTML 文件作为应用
 * 名称必填，在静态目录下创建以名称命名的子目录，将 HTML 文件复制为 index.html
 */
func (a *AppService) ImportHtml(htmlPath string, appName string) error {
	if appName == "" {
		return fmt.Errorf("应用名称不能为空")
	}

	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("获取应用目录失败")
	}

	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		return fmt.Errorf("HTML 文件不存在")
	}

	dirName := SanitizeDirName(appName)
	destDir := filepath.Join(staticDir, dirName)
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("应用目录 %s 已存在", dirName)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建应用目录失败: %w", err)
	}

	srcFile, err := os.Open(htmlPath)
	if err != nil {
		os.RemoveAll(destDir)
		return fmt.Errorf("打开 HTML 文件失败: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(filepath.Join(destDir, "index.html"))
	if err != nil {
		os.RemoveAll(destDir)
		return fmt.Errorf("创建 index.html 失败: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		os.RemoveAll(destDir)
		return fmt.Errorf("复制 HTML 文件失败: %w", err)
	}

	_, err = a.ScanApps()
	return err
}

/**
 * 导入 HTML 目录作为应用
 * 名称非必填，不填则使用源目录名称
 * 如果填写了名称，则目录名使用该名称
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
 * 导出单个应用为 zip 压缩包
 * 支持静态应用和网页应用，在 zip 中包含 manifest.json
 * 静态应用会打包目录内容，网页应用信息保存在 manifest 的 webApps 中
 */
func (a *AppService) ExportApp(id int64, savePath string) error {
	var app SubApp
	err := a.db.DB().QueryRow(
		"SELECT id, app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app WHERE id = ?",
		id,
	).Scan(&app.Id, &app.AppType, &app.DirName, &app.DisplayName, &app.IconPath, &app.EntryUrl, &app.SortOrder, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return fmt.Errorf("应用不存在")
	}

	if savePath == "" {
		zipName := fmt.Sprintf("%s_%s.zip", app.DirName, NowCompactFormatted())
		if app.AppType == "web" {
			zipName = fmt.Sprintf("%s_%s.zip", SanitizeDirName(app.DisplayName), NowCompactFormatted())
		}
		savePath = filepath.Join(os.TempDir(), zipName)
	}

	manifest := ExportManifest{
		Version: ManifestVersion,
	}

	if app.AppType == "web" {
		manifest.WebApps = []WebAppInfo{{Name: app.DisplayName, Url: app.EntryUrl}}
	} else {
		staticDir, err := a.GetStaticDir()
		if err != nil {
			return err
		}
		if staticDir == "" {
			return fmt.Errorf("获取应用目录失败")
		}

		appDir := filepath.Join(staticDir, app.DirName)
		if _, err := os.Stat(appDir); os.IsNotExist(err) {
			return fmt.Errorf("应用目录不存在")
		}
		manifest.Apps = []string{app.DirName}
	}

	return a.exportToZip(manifest, savePath)
}

/**
 * 批量导出多个应用为 zip 压缩包
 * 支持静态应用和网页应用，在 zip 中包含 manifest.json
 * 静态应用打包目录内容，网页应用信息保存在 manifest 的 webApps 中
 */
func (a *AppService) BatchExportApps(ids []int64, savePath string) error {
	if len(ids) == 0 {
		return fmt.Errorf("请至少选择一个应用")
	}

	staticDir, err := a.GetStaticDir()
	if err != nil {
		return err
	}
	if staticDir == "" {
		return fmt.Errorf("获取应用目录失败")
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT id, app_type, dir_name, display_name, entry_url FROM sub_app WHERE id IN (%s)", strings.Join(placeholders, ","))
	rows, err := a.db.DB().Query(query, args...)
	if err != nil {
		return fmt.Errorf("查询应用失败: %w", err)
	}
	defer rows.Close()

	var dirNames []string
	var webApps []WebAppInfo

	for rows.Next() {
		var id int64
		var appType string
		var dirName string
		var displayName string
		var entryUrl string
		if err := rows.Scan(&id, &appType, &dirName, &displayName, &entryUrl); err != nil {
			return err
		}
		if appType == "web" {
			webApps = append(webApps, WebAppInfo{Name: displayName, Url: entryUrl})
		} else {
			appDir := filepath.Join(staticDir, dirName)
			if _, err := os.Stat(appDir); err == nil {
				dirNames = append(dirNames, dirName)
			}
		}
	}

	if len(dirNames) == 0 && len(webApps) == 0 {
		return fmt.Errorf("没有可导出的应用")
	}

	if savePath == "" {
		savePath = filepath.Join(os.TempDir(), fmt.Sprintf("apps_export_%s.zip", NowCompactFormatted()))
	}

	manifest := ExportManifest{
		Version: ManifestVersion,
		Apps:    dirNames,
		WebApps: webApps,
	}

	return a.exportToZip(manifest, savePath)
}

/**
 * 将应用信息打包为 zip 文件
 * zip 内包含 manifest.json，静态应用的目录内容，以及网页应用信息
 */
func (a *AppService) exportToZip(manifest ExportManifest, zipPath string) error {
	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	file, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("创建压缩文件失败: %w", err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("生成清单文件失败: %w", err)
	}

	manifestWriter, err := w.Create("manifest.json")
	if err != nil {
		return fmt.Errorf("创建清单条目失败: %w", err)
	}
	if _, err := manifestWriter.Write(manifestData); err != nil {
		return fmt.Errorf("写入清单文件失败: %w", err)
	}

	if len(manifest.Apps) > 0 {
		staticDir, err := a.GetStaticDir()
		if err != nil {
			return fmt.Errorf("获取应用目录失败: %w", err)
		}
		if staticDir == "" {
			return fmt.Errorf("获取应用目录失败")
		}

		for _, dirName := range manifest.Apps {
			appDir := filepath.Join(staticDir, dirName)
			if err := a.addDirToZip(w, appDir, dirName); err != nil {
				return fmt.Errorf("打包应用 %s 失败: %w", dirName, err)
			}
		}
	}

	return nil
}

/**
 * 将一个目录添加到 zip 写入器中
 * baseInZip 为 zip 内的根路径前缀
 */
func (a *AppService) addDirToZip(w *zip.Writer, srcDir string, baseInZip string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		zipEntry := filepath.Join(baseInZip, relPath)
		zipEntry = filepath.ToSlash(zipEntry)

		if info.IsDir() {
			_, err := w.Create(zipEntry + "/")
			return err
		}

		fileWriter, err := w.Create(zipEntry)
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
 * 导入 zip 压缩包应用
 * 支持包含 manifest.json 的批量导出格式和不含 manifest 的单应用格式
 * 如果 manifest 中声明的应用目录已存在，则跳过该应用
 * 返回跳过的应用名称列表供前端提示
 */
func (a *AppService) ImportZip(zipPath string) (string, error) {
	staticDir, err := a.GetStaticDir()
	if err != nil {
		return "", err
	}
	if staticDir == "" {
		return "", fmt.Errorf("获取应用目录失败")
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("打开压缩包失败: %w", err)
	}
	defer reader.Close()

	manifestData := a.readManifestFromZip(reader)
	if manifestData != nil {
		return a.importZipWithManifest(reader, manifestData, staticDir)
	}

	return a.importZipWithoutManifest(reader, zipPath, staticDir)
}

/**
 * 从 zip 中读取 manifest.json 内容
 * 如果不存在或解析失败则返回 nil
 */
func (a *AppService) readManifestFromZip(reader *zip.ReadCloser) *ExportManifest {
	for _, f := range reader.File {
		if f.Name == "manifest.json" && !f.FileInfo().IsDir() {
			rc, err := f.Open()
			if err != nil {
				return nil
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil
			}

			var manifest ExportManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				return nil
			}
			if manifest.Version == "" {
				return nil
			}
			if len(manifest.Apps) == 0 && len(manifest.WebApps) == 0 {
				return nil
			}
			return &manifest
		}
	}
	return nil
}

/**
 * 处理包含 manifest 的 zip 导入
 * 校验 manifest 中声明的本地应用目录是否在 zip 中存在
 * 已存在的目录跳过，返回跳过的目录名
 * 同时导入 manifest 中的网页应用信息
 */
func (a *AppService) importZipWithManifest(reader *zip.ReadCloser, manifest *ExportManifest, staticDir string) (string, error) {
	zipDirs := a.collectZipTopDirs(reader)

	var missingDirs []string
	for _, appDir := range manifest.Apps {
		if !zipDirs[appDir] {
			missingDirs = append(missingDirs, appDir)
		}
	}
	if len(missingDirs) > 0 {
		return "", fmt.Errorf("清单中声明的应用目录在压缩包中不存在: %s", strings.Join(missingDirs, ", "))
	}

	var skipped []string
	for _, f := range reader.File {
		if f.Name == "manifest.json" {
			continue
		}

		topDir := a.extractTopDirFromPath(f.Name)
		if topDir == "" {
			continue
		}

		destDir := filepath.Join(staticDir, topDir)
		if _, err := os.Stat(destDir); err == nil {
			if !containsStr(skipped, topDir) {
				skipped = append(skipped, topDir)
			}
			continue
		}

		fpath := filepath.Join(staticDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return "", err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return "", err
		}
	}

	for _, webApp := range manifest.WebApps {
		existingName := ""
		a.db.DB().QueryRow("SELECT display_name FROM sub_app WHERE entry_url = ? AND app_type = 'web'", webApp.Url).Scan(&existingName)
		if existingName != "" {
			if !containsStr(skipped, webApp.Name) {
				skipped = append(skipped, webApp.Name)
			}
			continue
		}
		if _, err := a.CreateWebApp(webApp.Name, webApp.Url); err != nil {
			LogError("导入网页应用 %s 失败: %v", webApp.Name, err)
		}
	}

	_, err := a.ScanApps()

	skippedStr := strings.Join(skipped, ", ")
	return skippedStr, err
}

/**
 * 处理不含 manifest 的 zip 导入（兼容旧格式）
 * 自动识别 zip 内的顶层目录结构
 */
func (a *AppService) importZipWithoutManifest(reader *zip.ReadCloser, zipPath string, staticDir string) (string, error) {
	topDir := a.detectTopDir(reader, zipPath)

	destDir := filepath.Join(staticDir, topDir)
	if _, err := os.Stat(destDir); err == nil {
		return "", fmt.Errorf("应用目录 %s 已存在", topDir)
	}

	for _, f := range reader.File {
		fpath := filepath.Join(staticDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return "", err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return "", err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return "", err
		}
	}

	_, err := a.ScanApps()
	return "", err
}

/**
 * 收集 zip 中所有顶层目录名
 * 返回以目录名为键的 map
 */
func (a *AppService) collectZipTopDirs(reader *zip.ReadCloser) map[string]bool {
	dirs := make(map[string]bool)
	for _, f := range reader.File {
		topDir := a.extractTopDirFromPath(f.Name)
		if topDir != "" {
			dirs[topDir] = true
		}
	}
	return dirs
}

/**
 * 从 zip 条目路径中提取顶层目录名
 * 例如 "app1/index.html" 返回 "app1"
 */
func (a *AppService) extractTopDirFromPath(zipPath string) string {
	parts := strings.SplitN(zipPath, "/", 2)
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
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

/**
 * 判断字符串切片中是否包含指定值
 */
func containsStr(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
