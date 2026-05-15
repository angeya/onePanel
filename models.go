package main

/**
 * 应用配置模型
 * 对应 app_config 表，用于存储全局键值对配置
 */
type AppConfig struct {
	Id          int64  `json:"id"`
	ConfigKey   string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	UpdatedAt   string `json:"updatedAt"`
}

/**
 * 子应用模型
 * 对应 sub_app 表，支持静态应用和网页应用两种类型
 */
type SubApp struct {
	Id          int64  `json:"id"`
	AppType     string `json:"appType"`
	DirName     string `json:"dirName"`
	DisplayName string `json:"displayName"`
	IconPath    string `json:"iconPath"`
	EntryUrl    string `json:"entryUrl"`
	SortOrder   int    `json:"sortOrder"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

/**
 * 快捷命令分类模型
 * 对应 shortcut_category 表，用于终端侧边栏的快捷命令分类
 */
type ShortcutCategory struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

/**
 * 快捷命令模型
 * 对应 shortcut_command 表，终端侧边栏的快捷命令
 */
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

/**
 * 快速启动命令分类模型
 * 对应 shortcut_cmd_category 表，用于快速启动功能的命令分类
 */
type ShortcutCmdCategory struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

/**
 * 快速启动命令模型
 * 对应 shortcut_cmd 表，快速启动功能的具体命令
 */
type ShortcutCmd struct {
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

/**
 * 命令历史记录模型
 * 对应 command_history 表，记录终端中执行过的命令
 */
type CommandHistory struct {
	Id         int64  `json:"id"`
	Command    string `json:"command"`
	Shell      string `json:"shell"`
	WorkDir    string `json:"workDir"`
	ExecutedAt string `json:"executedAt"`
}

/**
 * 命令历史查询结果
 * 包含分页后的历史记录列表和总数
 */
type HistoryResult struct {
	Histories []CommandHistory `json:"histories"`
	Total     int64            `json:"total"`
}

/**
 * 网络端口信息模型
 * 对应 netstat 命令输出解析后的结构
 */
type PortInfo struct {
	Protocol    string `json:"protocol"`
	LocalAddr   string `json:"localAddr"`
	LocalPort   int    `json:"localPort"`
	ForeignAddr string `json:"foreignAddr"`
	ForeignPort int    `json:"foreignPort"`
	State       string `json:"state"`
	Pid         int    `json:"pid"`
	ProcessName string `json:"processName"`
}

/**
 * PTY 启动请求参数
 * 前端启动终端时传入的参数
 */
type StartRequest struct {
	Shell string `json:"shell"`
	Cols  int    `json:"cols"`
	Rows  int    `json:"rows"`
}
