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

const (
	ManifestVersion = "1.0"
	staticAppType   = "static"
	webAppType      = "web"
)

/**
 * AppService 我的应用服务。
 * 负责子应用的增删改查、导入导出、静态目录管理等业务逻辑。
 */
type AppService struct {
	db     *Database
	server *StaticServer
}

/**
 * ExportManifest 描述导出包清单。
 * Apps 保存静态应用目录名，WebApps 保存网页应用信息。
 */
type ExportManifest struct {
	Version string       `json:"version"`
	Apps    []string     `json:"apps"`
	WebApps []WebAppInfo `json:"webApps"`
}

/**
 * WebAppInfo 保存网页应用的导出信息。
 */
type WebAppInfo struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

/**
 * 创建 AppService 实例。
 */
func NewAppService(db *Database, server *StaticServer) *AppService {
	return &AppService{db: db, server: server}
}

/**
 * GetStaticDir 返回应用静态目录。
 * 固定使用可执行文件同级目录下的 apps 目录。
 */
func (a *AppService) GetStaticDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exePath), "apps"), nil
}

/**
 * GetServerStatus 返回静态服务器当前状态。
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
 * StartServer 使用静态目录启动服务。
 */
func (a *AppService) StartServer() (int, error) {
	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return 0, err
	}
	if a.server == nil {
		return 0, fmt.Errorf("静态服务器未初始化")
	}
	return a.server.Start(staticDir)
}

/**
 * StopServer 停止静态服务器。
 */
func (a *AppService) StopServer() error {
	if a.server == nil {
		return nil
	}
	return a.server.Stop()
}

/**
 * OpenApp 打开指定应用并返回前端需要的显示信息。
 * 静态应用会自动确保静态服务器已启动，网页应用直接返回原始 URL。
 */
func (a *AppService) OpenApp(appId int64) (map[string]interface{}, error) {
	app, err := a.getAppByID(appId)
	if err != nil {
		return nil, err
	}

	if app.AppType == webAppType {
		return map[string]interface{}{
			"url":      app.EntryUrl,
			"name":     app.DisplayName,
			"iconPath": "",
		}, nil
	}

	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return nil, err
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

	return map[string]interface{}{
		"url":      fmt.Sprintf("http://127.0.0.1:%d%s", port, app.EntryUrl),
		"name":     app.DisplayName,
		"iconPath": a.buildIconURL(app, port),
	}, nil
}

/**
 * ScanApps 扫描静态目录并同步应用列表。
 * 会自动更新新增、变更和已被移除的静态应用记录。
 */
func (a *AppService) ScanApps() ([]SubApp, error) {
	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(staticDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SubApp{}, nil
		}
		return nil, fmt.Errorf("读取应用目录失败: %w", err)
	}

	scannedDirs := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		subDir := filepath.Join(staticDir, dirName)
		if !a.hasAppEntry(subDir) {
			continue
		}

		scannedDirs[dirName] = true
		displayName := dirName
		iconPath := a.resolveIconPath(subDir)
		entryURL := a.buildStaticEntryURL(dirName)
		a.upsertStaticApp(dirName, displayName, iconPath, entryURL)
	}

	a.cleanupMissingStaticApps(scannedDirs)
	return a.GetApps()
}

/**
 * GetApps 获取所有应用列表。
 */
func (a *AppService) GetApps() ([]SubApp, error) {
	rows, err := a.db.DB().Query(
		"SELECT id, app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app ORDER BY sort_order, id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []SubApp
	for rows.Next() {
		var app SubApp
		if err := rows.Scan(
			&app.Id,
			&app.AppType,
			&app.DirName,
			&app.DisplayName,
			&app.IconPath,
			&app.EntryUrl,
			&app.SortOrder,
			&app.CreatedAt,
			&app.UpdatedAt,
		); err != nil {
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
 * CreateWebApp 创建网页应用。
 */
func (a *AppService) CreateWebApp(name string, url string) (*SubApp, error) {
	if err := validateWebAppInput(name, url); err != nil {
		return nil, err
	}

	now := NowFormatted()
	result, err := a.db.DB().Exec(
		"INSERT INTO sub_app (app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at) VALUES ('web', '', ?, '', ?, 0, ?, ?)",
		name,
		url,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("创建网页应用失败: %w", err)
	}

	id, _ := result.LastInsertId()
	return &SubApp{
		Id:          id,
		AppType:     webAppType,
		DisplayName: name,
		EntryUrl:    url,
		SortOrder:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

/**
 * UpdateWebApp 更新网页应用信息。
 */
func (a *AppService) UpdateWebApp(id int64, name string, url string) error {
	if err := validateWebAppInput(name, url); err != nil {
		return err
	}

	now := NowFormatted()
	_, err := a.db.DB().Exec(
		"UPDATE sub_app SET display_name = ?, entry_url = ?, updated_at = ? WHERE id = ? AND app_type = 'web'",
		name,
		url,
		now,
		id,
	)
	return err
}

/**
 * UpdateDisplayName 更新应用名称。
 * 静态应用还会同步调整目录名、入口路径和图标路径。
 */
func (a *AppService) UpdateDisplayName(id int64, name string) error {
	if name == "" {
		return fmt.Errorf("应用名称不能为空")
	}

	var appType string
	var oldDirName string
	var iconPath string
	err := a.db.DB().QueryRow(
		"SELECT app_type, dir_name, icon_path FROM sub_app WHERE id = ?",
		id,
	).Scan(&appType, &oldDirName, &iconPath)
	if err != nil {
		return fmt.Errorf("应用不存在")
	}

	now := NowFormatted()
	if appType != staticAppType {
		_, err = a.db.DB().Exec(
			"UPDATE sub_app SET display_name = ?, updated_at = ? WHERE id = ?",
			name,
			now,
			id,
		)
		return err
	}

	newDirName := SanitizeDirName(name)
	if newDirName == "" {
		return fmt.Errorf("应用名称无效")
	}

	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return err
	}

	oldPath := filepath.Join(staticDir, oldDirName)
	newPath := filepath.Join(staticDir, newDirName)
	updatedIconPath := iconPath
	if oldPath != newPath {
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("目录 %s 已存在", newDirName)
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return fmt.Errorf("重命名目录失败: %w", err)
		}
		updatedIconPath = a.resolveRenamedIconPath(newPath, iconPath)
	}

	entryURL := a.buildStaticEntryURL(newDirName)
	_, err = a.db.DB().Exec(
		"UPDATE sub_app SET display_name = ?, dir_name = ?, entry_url = ?, icon_path = ?, updated_at = ? WHERE id = ?",
		name,
		newDirName,
		entryURL,
		updatedIconPath,
		now,
		id,
	)
	return err
}

/**
 * DeleteApp 删除应用。
 * 对静态应用会同时删除磁盘目录。
 */
func (a *AppService) DeleteApp(id int64) error {
	var appType string
	var dirName string
	err := a.db.DB().QueryRow(
		"SELECT app_type, dir_name FROM sub_app WHERE id = ?",
		id,
	).Scan(&appType, &dirName)
	if err != nil {
		return err
	}

	if appType == staticAppType {
		staticDir, err := a.mustGetStaticDir()
		if err != nil {
			return err
		}
		if dirName != "" {
			_ = os.RemoveAll(filepath.Join(staticDir, dirName))
		}
	}

	_, err = a.db.DB().Exec("DELETE FROM sub_app WHERE id = ?", id)
	return err
}

/**
 * ImportHtml 将单个 HTML 文件导入为静态应用。
 */
func (a *AppService) ImportHtml(htmlPath string, appName string) error {
	if appName == "" {
		return fmt.Errorf("应用名称不能为空")
	}
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		return fmt.Errorf("HTML 文件不存在")
	}

	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return err
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
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("打开 HTML 文件失败: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(filepath.Join(destDir, "index.html"))
	if err != nil {
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("创建 index.html 失败: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		_ = os.RemoveAll(destDir)
		return fmt.Errorf("复制 HTML 文件失败: %w", err)
	}

	_, err = a.ScanApps()
	return err
}

/**
 * ImportDir 将整个目录导入为静态应用。
 */
func (a *AppService) ImportDir(srcDir string, appName string) error {
	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return err
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
 * ExportApp 导出单个应用。
 */
func (a *AppService) ExportApp(id int64, savePath string) error {
	app, err := a.getAppByID(id)
	if err != nil {
		return err
	}

	if savePath == "" {
		savePath = a.buildDefaultExportPath(app)
	}

	manifest := ExportManifest{Version: ManifestVersion}
	if app.AppType == webAppType {
		manifest.WebApps = []WebAppInfo{{Name: app.DisplayName, Url: app.EntryUrl}}
		return a.exportToZip(manifest, savePath)
	}

	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return err
	}

	appDir := filepath.Join(staticDir, app.DirName)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return fmt.Errorf("应用目录不存在")
	}
	manifest.Apps = []string{app.DirName}
	return a.exportToZip(manifest, savePath)
}

/**
 * BatchExportApps 批量导出多个应用。
 */
func (a *AppService) BatchExportApps(ids []int64, savePath string) error {
	if len(ids) == 0 {
		return fmt.Errorf("请至少选择一个应用")
	}

	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return err
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		"SELECT id, app_type, dir_name, display_name, entry_url FROM sub_app WHERE id IN (%s)",
		strings.Join(placeholders, ","),
	)
	rows, err := a.db.DB().Query(query, args...)
	if err != nil {
		return fmt.Errorf("查询应用失败: %w", err)
	}
	defer rows.Close()

	var dirNames []string
	var webApps []WebAppInfo
	for rows.Next() {
		var dirName string
		var displayName string
		var entryURL string
		var appType string
		var id int64
		if err := rows.Scan(&id, &appType, &dirName, &displayName, &entryURL); err != nil {
			return err
		}
		if appType == webAppType {
			webApps = append(webApps, WebAppInfo{Name: displayName, Url: entryURL})
			continue
		}
		if _, err := os.Stat(filepath.Join(staticDir, dirName)); err == nil {
			dirNames = append(dirNames, dirName)
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
 * ImportZip 导入 zip 压缩包应用。
 * 支持带 manifest 的导出包，也兼容单目录直接压缩的旧格式。
 * 返回值为被跳过的应用名称列表，使用逗号拼接供前端提示。
 */
func (a *AppService) ImportZip(zipPath string) (string, error) {
	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return "", err
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
 * exportToZip 将应用信息打包为 zip 文件。
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

	writer := zip.NewWriter(file)
	defer writer.Close()

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("生成清单文件失败: %w", err)
	}

	manifestWriter, err := writer.Create("manifest.json")
	if err != nil {
		return fmt.Errorf("创建清单条目失败: %w", err)
	}
	if _, err := manifestWriter.Write(manifestData); err != nil {
		return fmt.Errorf("写入清单文件失败: %w", err)
	}

	if len(manifest.Apps) == 0 {
		return nil
	}

	staticDir, err := a.mustGetStaticDir()
	if err != nil {
		return err
	}
	for _, dirName := range manifest.Apps {
		appDir := filepath.Join(staticDir, dirName)
		if err := a.addDirToZip(writer, appDir, dirName); err != nil {
			return fmt.Errorf("打包应用 %s 失败: %w", dirName, err)
		}
	}
	return nil
}

/**
 * addDirToZip 将目录递归写入 zip。
 */
func (a *AppService) addDirToZip(writer *zip.Writer, srcDir string, baseInZip string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		zipEntry := filepath.ToSlash(filepath.Join(baseInZip, relPath))
		if info.IsDir() {
			_, err := writer.Create(zipEntry + "/")
			return err
		}

		fileWriter, err := writer.Create(zipEntry)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(fileWriter, file)
		return err
	})
}

/**
 * mustGetStaticDir 获取并校验静态目录。
 */
func (a *AppService) mustGetStaticDir() (string, error) {
	staticDir, err := a.GetStaticDir()
	if err != nil {
		return "", err
	}
	if staticDir == "" {
		return "", fmt.Errorf("获取应用目录失败")
	}
	return staticDir, nil
}

/**
 * ensureServerRunning 确保静态服务器已启动。
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
 * getAppByID 根据 ID 查询应用。
 */
func (a *AppService) getAppByID(appId int64) (*SubApp, error) {
	var app SubApp
	err := a.db.DB().QueryRow(
		"SELECT id, app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app WHERE id = ?",
		appId,
	).Scan(
		&app.Id,
		&app.AppType,
		&app.DirName,
		&app.DisplayName,
		&app.IconPath,
		&app.EntryUrl,
		&app.SortOrder,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("应用不存在")
	}
	return &app, nil
}

/**
 * hasAppEntry 检查目录下是否存在 index.html 入口文件。
 */
func (a *AppService) hasAppEntry(subDir string) bool {
	indexFile := filepath.Join(subDir, "index.html")
	_, err := os.Stat(indexFile)
	return err == nil
}

/**
 * resolveIconPath 返回应用目录中的 icon.png 路径。
 */
func (a *AppService) resolveIconPath(subDir string) string {
	iconFile := filepath.Join(subDir, "icon.png")
	if _, err := os.Stat(iconFile); err == nil {
		return iconFile
	}
	return ""
}

/**
 * buildStaticEntryURL 生成静态应用入口路径。
 */
func (a *AppService) buildStaticEntryURL(dirName string) string {
	return fmt.Sprintf("/%s/index.html", dirName)
}

/**
 * buildIconURL 生成前端访问图标的 URL。
 */
func (a *AppService) buildIconURL(app *SubApp, port int) string {
	if app.IconPath == "" {
		return ""
	}
	dir := strings.TrimSuffix(app.EntryUrl, "index.html")
	return fmt.Sprintf("http://127.0.0.1:%d%sicon.png", port, dir)
}

/**
 * buildDefaultExportPath 生成默认导出文件路径。
 */
func (a *AppService) buildDefaultExportPath(app *SubApp) string {
	zipName := fmt.Sprintf("%s_%s.zip", app.DirName, NowCompactFormatted())
	if app.AppType == webAppType {
		zipName = fmt.Sprintf("%s_%s.zip", SanitizeDirName(app.DisplayName), NowCompactFormatted())
	}
	return filepath.Join(os.TempDir(), zipName)
}

/**
 * resolveRenamedIconPath 在静态应用目录重命名后重算图标路径。
 */
func (a *AppService) resolveRenamedIconPath(newPath string, oldIconPath string) string {
	if oldIconPath == "" {
		return ""
	}
	newIconPath := filepath.Join(newPath, "icon.png")
	if _, err := os.Stat(newIconPath); err == nil {
		return newIconPath
	}
	return oldIconPath
}

/**
 * upsertStaticApp 创建或更新静态应用记录。
 */
func (a *AppService) upsertStaticApp(dirName, displayName, iconPath, entryURL string) (SubApp, bool) {
	var existing SubApp
	err := a.db.DB().QueryRow(
		"SELECT id, app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at FROM sub_app WHERE dir_name = ? AND app_type = 'static'",
		dirName,
	).Scan(
		&existing.Id,
		&existing.AppType,
		&existing.DirName,
		&existing.DisplayName,
		&existing.IconPath,
		&existing.EntryUrl,
		&existing.SortOrder,
		&existing.CreatedAt,
		&existing.UpdatedAt,
	)
	if err != nil {
		now := NowFormatted()
		result, err := a.db.DB().Exec(
			"INSERT INTO sub_app (app_type, dir_name, display_name, icon_path, entry_url, sort_order, created_at, updated_at) VALUES ('static', ?, ?, ?, ?, 0, ?, ?)",
			dirName,
			displayName,
			iconPath,
			entryURL,
			now,
			now,
		)
		if err != nil {
			return SubApp{}, false
		}
		id, _ := result.LastInsertId()
		return SubApp{
			Id:          id,
			AppType:     staticAppType,
			DirName:     dirName,
			DisplayName: displayName,
			IconPath:    iconPath,
			EntryUrl:    entryURL,
			SortOrder:   0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, true
	}

	if existing.DisplayName == existing.DirName || existing.DisplayName != displayName {
		now := NowFormatted()
		_, _ = a.db.DB().Exec(
			"UPDATE sub_app SET display_name = ?, icon_path = ?, entry_url = ?, updated_at = ? WHERE id = ?",
			displayName,
			iconPath,
			entryURL,
			now,
			existing.Id,
		)
		existing.DisplayName = displayName
		existing.IconPath = iconPath
		existing.EntryUrl = entryURL
	}
	return existing, false
}

/**
 * cleanupMissingStaticApps 清理已不存在目录对应的静态应用。
 */
func (a *AppService) cleanupMissingStaticApps(scannedDirs map[string]bool) {
	rows, err := a.db.DB().Query("SELECT id, dir_name FROM sub_app WHERE app_type = 'static'")
	if err != nil {
		return
	}
	defer rows.Close()

	var toDelete []int64
	for rows.Next() {
		var id int64
		var dirName string
		if err := rows.Scan(&id, &dirName); err != nil {
			continue
		}
		if !scannedDirs[dirName] {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		_, _ = a.db.DB().Exec("DELETE FROM sub_app WHERE id = ?", id)
	}
}

/**
 * readManifestFromZip 读取压缩包中的 manifest.json。
 * 如果不存在或解析失败则返回 nil。
 */
func (a *AppService) readManifestFromZip(reader *zip.ReadCloser) *ExportManifest {
	for _, file := range reader.File {
		if file.Name != "manifest.json" || file.FileInfo().IsDir() {
			continue
		}

		rc, err := file.Open()
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
		return &manifest
	}
	return nil
}

/**
 * importZipWithManifest 导入包含 manifest 的导出包。
 * 静态应用按目录恢复，网页应用按清单写入数据库。
 */
func (a *AppService) importZipWithManifest(reader *zip.ReadCloser, manifest *ExportManifest, staticDir string) (string, error) {
	var skipped []string
	allowedStaticDirs := make(map[string]bool)
	for _, dirName := range manifest.Apps {
		allowedStaticDirs[dirName] = true
	}

	for _, dirName := range manifest.Apps {
		destDir := filepath.Join(staticDir, dirName)
		if _, err := os.Stat(destDir); err == nil {
			skipped = append(skipped, dirName)
			continue
		}
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return strings.Join(skipped, ", "), fmt.Errorf("创建应用目录失败: %w", err)
		}
	}

	for _, file := range reader.File {
		if file.Name == "manifest.json" {
			continue
		}
		topDir := a.extractTopLevelDir(file.Name)
		if topDir == "" || !allowedStaticDirs[topDir] || containsString(skipped, topDir) {
			continue
		}
		if err := a.extractZipFile(file, staticDir); err != nil {
			return strings.Join(skipped, ", "), err
		}
	}

	for _, webApp := range manifest.WebApps {
		_, err := a.CreateWebApp(webApp.Name, webApp.Url)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "已存在") {
				skipped = append(skipped, webApp.Name)
				continue
			}
			return strings.Join(skipped, ", "), err
		}
	}

	_, err := a.ScanApps()
	return strings.Join(skipped, ", "), err
}

/**
 * importZipWithoutManifest 兼容导入不带 manifest 的旧 zip。
 * 约定整个压缩包只代表一个静态应用目录。
 */
func (a *AppService) importZipWithoutManifest(reader *zip.ReadCloser, zipPath string, staticDir string) (string, error) {
	baseName := strings.TrimSuffix(filepath.Base(zipPath), filepath.Ext(zipPath))
	dirName := SanitizeDirName(baseName)
	if dirName == "" {
		return "", fmt.Errorf("压缩包名称无效")
	}

	destDir := filepath.Join(staticDir, dirName)
	if _, err := os.Stat(destDir); err == nil {
		return dirName, nil
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("创建应用目录失败: %w", err)
	}

	for _, file := range reader.File {
		if err := a.extractZipFile(file, destDir); err != nil {
			return "", err
		}
	}

	_, err := a.ScanApps()
	return "", err
}

/**
 * extractZipFile 将单个 zip 条目解压到目标根目录。
 */
func (a *AppService) extractZipFile(file *zip.File, targetRoot string) error {
	targetPath := filepath.Join(targetRoot, filepath.FromSlash(file.Name))
	cleanRoot := filepath.Clean(targetRoot)
	cleanTarget := filepath.Clean(targetPath)
	if !strings.HasPrefix(cleanTarget, cleanRoot) {
		return fmt.Errorf("压缩包包含非法路径: %s", file.Name)
	}

	if file.FileInfo().IsDir() {
		return os.MkdirAll(cleanTarget, 0755)
	}
	if err := os.MkdirAll(filepath.Dir(cleanTarget), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开压缩包文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(cleanTarget)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("解压文件失败: %w", err)
	}
	return nil
}

/**
 * extractTopLevelDir 提取 zip 条目的顶层目录名。
 */
func (a *AppService) extractTopLevelDir(name string) string {
	cleanName := strings.Trim(strings.ReplaceAll(name, "\\", "/"), "/")
	if cleanName == "" {
		return ""
	}
	parts := strings.Split(cleanName, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

/**
 * validateWebAppInput 校验网页应用的名称与地址。
 */
func validateWebAppInput(name string, url string) error {
	if name == "" {
		return fmt.Errorf("应用名称不能为空")
	}
	if url == "" {
		return fmt.Errorf("应用地址不能为空")
	}
	return nil
}

/**
 * containsString 判断切片中是否包含指定字符串。
 */
func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
