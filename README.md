# oneWin

一款基于 Wails + Vue3 + Element Plus + SQLite 开发的 Windows 桌面程序，旨在一个应用窗口中解决更多的问题。

## 项目简介

oneWin 将终端、快捷命令、子应用加载、快速启动和系统工具整合到同一窗口中，通过左侧导航 + 右侧多 Tab 页的布局方式，实现高效的多任务操作体验。

### 核心功能

| 功能模块 | 说明 |
|---------|------|
| 终端 | 支持多标签 PTY（基于 conpty），效果与 cmd/powershell 一致；支持快捷命令执行、历史命令查看、终端内搜索 |
| 快捷命令 | 终端侧边栏的快捷命令面板，支持分类管理、点击自动输入到终端并执行 |
| 我的应用 | 将 HTML 页面作为子应用加载，通过内置 HTTP 静态服务器在 iframe 中运行；支持应用的扫描、导入、导出、增删改查 |
| 快速启动 | 双击快速执行命令（如 JMeter、Arthas 等），支持分组管理，可配置 Shell 类型和工作目录 |
| 实用工具 | 提供端口查看、进程管理（taskkill）等 Windows 系统级工具 |
| 系统设置 | 主题切换（深色/常规）、默认 Shell 配置、应用目录配置、版本信息 |

### 布局方式

- **左右布局**：左侧为功能导航面板，右侧为终端/功能界面
- **左侧面板**：分两列，左列为主功能列表（终端、我的应用、快速启动、实用工具），右列为当前功能的配置信息
- **右侧面板**：统一使用 Tab 页展示，终端以"终端1、终端2"命名，应用/工具以名称命名，Tab 不可重复

## 技术栈

### 后端

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.25.0 | 后端开发语言 |
| Wails | v2.12.0 | Go 桌面应用框架，提供 Go 与前端的桥接能力 |
| modernc.org/sqlite | v1.50.0 | 纯 Go 实现的 SQLite 驱动，无需 CGO |
| conpty | v0.1.4 | Windows 伪终端（ConPTY）的 Go 封装 |

### 前端

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | ^3.5.33 | 前端框架，使用 Composition API |
| Element Plus | ^2.13.7 | UI 组件库 |
| @element-plus/icons-vue | ^2.3.2 | Element Plus 图标库 |
| xterm.js | ^5.3.0 | 终端渲染器 |
| xterm-addon-fit | ^0.8.0 | xterm.js 自适应尺寸插件 |
| xterm-addon-search | ^0.13.0 | xterm.js 终端搜索插件 |
| xterm-addon-web-links | ^0.9.0 | xterm.js 网页链接识别插件 |
| Vite | ^4.5.14 | 前端构建工具 |
| @vitejs/plugin-vue | ^4.6.2 | Vite Vue 插件 |

## 项目结构

```
onePanel/
├── main.go                 # 程序入口，初始化各 Service 并启动 Wails 应用
├── app.go                  # App 基础结构，提供文件/目录选择对话框（通过 PowerShell 调用）
├── appservice.go           # 我的应用服务，子应用增删改查、导入导出、静态目录管理
├── database.go             # 数据库封装，SQLite 连接管理、建表迁移、配置项读写
├── models.go               # 数据模型定义（AppConfig, SubApp, ShortcutCategory 等）
├── pty.go                  # 伪终端服务，多实例管理、输出读取与事件发射
├── shortcut.go             # 终端快捷命令服务，分类和命令的 CRUD
├── shortcutcmd.go          # 快速启动服务，分组管理、命令 CRUD 与执行
├── history.go              # 命令历史服务，记录、查询、搜索与删除
├── staticserver.go         # HTTP 静态文件服务器，为子应用提供静态资源服务
├── toolservice.go          # 实用工具服务，端口查询（netstat）、进程管理（taskkill）
├── settingservice.go       # 系统设置服务，主题/Shell/应用目录等配置项的读写
├── util.go                 # 工具函数（时间格式化、Shell 路径解析、目录操作等）
├── go.mod                  # Go 模块定义
├── go.sum                  # Go 依赖校验
├── wails.json              # Wails 项目配置
├── 软件需求说明.md           # 项目需求文档
└── frontend/               # 前端工程
    ├── index.html          # SPA 入口 HTML
    ├── package.json        # 前端依赖定义
    ├── vite.config.js      # Vite 构建配置
    └── src/
        ├── main.js         # 前端入口，创建 Vue 实例并挂载 Element Plus
        ├── App.vue         # 根组件，主布局与各模块的协调中心
        ├── style.css       # 全局样式（CSS 变量定义主题色等）
        ├── components/
        │   └── SearchBar.vue        # 全局搜索栏组件（Ctrl+F 触发）
        ├── composables/
        │   ├── useAppTabs.js        # Tab 页管理（终端/应用/工具/快速启动 Tab 的增删切换）
        │   ├── useAppService.js     # 我的应用模块业务逻辑
        │   ├── useQuickLaunch.js    # 快速启动模块业务逻辑
        │   ├── useSettings.js       # 系统设置模块业务逻辑
        │   ├── useTheme.js          # 主题切换逻辑
        │   └── useTerminalEvent.js  # 终端事件处理（发送命令、记录历史）
        ├── utils/
        │   └── domSearch.js         # DOM 内容搜索工具（高亮、上下查找）
        └── views/
            ├── app/
            │   ├── AppSidebar.vue   # 左侧导航侧边栏
            │   └── AppDialogs.vue   # 我的应用相关弹窗集合
            ├── quicklaunch/
            │   ├── QuickLaunchTab.vue   # 快速启动 Tab 页
            │   └── ShortcutExecTab.vue  # 快速启动命令执行结果 Tab 页
            ├── settings/
            │   └── SettingsDialog.vue   # 系统设置弹窗
            ├── terminal/
            │   ├── TerminalTab.vue      # 终端 Tab 页（集成 xterm.js）
            │   ├── ShortcutPanel.vue    # 终端侧边栏快捷命令面板
            │   └── HistoryPanel.vue     # 终端侧边栏历史命令面板
            └── tools/
                └── ToolsPage.vue        # 实用工具页面（端口查看、进程管理）
```

## 后端架构

### 服务层设计

程序采用服务层架构，各 Service 在 `main.go` 中初始化并通过 Wails 的 `Bind` 机制暴露给前端：

```
main.go
  ├── App              → 基础对话框（文件选择、目录选择、保存文件）
  ├── PtyService       → 伪终端管理（启动/写入/调整大小/停止）
  ├── ShortcutService  → 终端快捷命令（分类 CRUD、命令 CRUD）
  ├── HistoryService   → 命令历史（记录/查询/搜索/删除）
  ├── StaticServer     → 静态文件服务器（启动/停止/状态查询）
  ├── AppService       → 我的应用（扫描/导入/导出/增删改查）
  ├── ShortcutCmdService → 快速启动（分组 CRUD、命令 CRUD、命令执行）
  ├── ToolService      → 实用工具（端口列表/进程终止）
  └── SettingService   → 系统设置（配置项读写）
```

### 数据库

使用 SQLite 存储数据，数据库文件位于可执行文件同级的 `data/onewin.db`，使用 WAL 模式提升并发性能。

| 表名 | 说明 |
|------|------|
| shortcut_category | 终端快捷命令分类 |
| shortcut_command | 终端快捷命令 |
| command_history | 命令执行历史 |
| app_config | 全局键值对配置（主题、Shell、应用目录等） |
| sub_app | 子应用信息（静态应用和网页应用） |
| shortcut_cmd_group | 快速启动命令分组 |
| shortcut_cmd | 快速启动命令 |

### 前后端通信

- **Wails Bind**：Go 的 Service 方法通过 `Bind` 暴露为前端可调用的 JavaScript 函数，前端通过 `window.go.main.ServiceName.MethodName()` 调用
- **Wails Events**：后端通过 `runtime.EventsEmit` 向前端发送事件（如 PTY 输出 `pty-output-{id}`、PTY 退出 `pty-exit-{id}`），前端通过 `runtime.EventsOn` 监听

## 前端架构

### Composables 模式

前端逻辑按功能模块拆分为 Composables（组合式函数），每个 Composable 封装一个功能领域的状态和方法：

| Composable | 职责 |
|-----------|------|
| useAppTabs | Tab 页生命周期管理（创建、切换、关闭） |
| useAppService | 我的应用模块所有业务逻辑 |
| useQuickLaunch | 快速启动模块所有业务逻辑 |
| useSettings | 系统设置读写 |
| useTheme | 主题加载与切换 |
| useTerminalEvent | 终端命令发送与历史记录 |

### 组件层次

```
App.vue（根组件，协调各模块）
├── AppSidebar.vue（左侧导航）
├── AppDialogs.vue（应用相关弹窗）
├── SettingsDialog.vue（设置弹窗）
├── SearchBar.vue（全局搜索栏）
├── TerminalTab.vue（终端 Tab，使用 xterm.js）
├── QuickLaunchTab.vue（快速启动 Tab）
├── ShortcutExecTab.vue（命令执行结果 Tab）
└── ToolsPage.vue（实用工具 Tab）
```

## 开发环境搭建

### 前置条件

1. **Go** >= 1.25.0 — [下载地址](https://go.dev/dl/)
2. **Node.js** >= 16（推荐 LTS 版本）— [下载地址](https://nodejs.org/)
3. **Wails CLI** v2 — 安装命令：
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```
4. **Windows** 操作系统（本项目仅支持 Windows）

### 克隆项目

```bash
git clone <repository-url>
cd onePanel
```

### 安装前端依赖

```bash
cd frontend
npm install
cd ..
```

### 开发模式运行

开发模式支持前端热更新，修改前端代码后浏览器端自动刷新：

```bash
wails dev
```

如果需要在浏览器中调试前端（可访问 Go 方法），开发服务器会运行在 `http://localhost:34115`。

### 构建生产版本

```bash
wails build
```

构建产物位于 `build/bin/oneWin.exe`。

## 数据存储

| 数据 | 路径 |
|------|------|
| SQLite 数据库 | `{exe所在目录}/data/onewin.db` |
| 子应用静态文件 | `{exe所在目录}/apps/`（默认，可在设置中自定义） |

## 关键设计说明

### PTY 输出优化

PTY 输出采用读 goroutine + flush goroutine 分离架构：
- 读 goroutine 持续读取 PTY 输出并写入缓冲区
- flush goroutine 以约 60fps（16ms 间隔）将缓冲区内容批量发送到前端
- 缓冲区超过 64KB 时立即发送，避免内存膨胀
- 该设计避免了 `EventsEmit` 过于频繁导致 Wails WebView2 崩溃的问题

### Shell 路径解析

通过 `ResolveShellPath` 函数将 Shell 名称解析为 64 位绝对路径：
- 64 位进程直接使用 `System32` 目录
- 32 位进程通过 `Sysnative` 虚拟目录访问真正的 64 位 `System32`
- 解决了 32 位进程被 Windows 文件系统重定向导致无法使用 `ssh` 等命令的问题

### 子应用加载

子应用通过内置的 HTTP 静态服务器加载：
- 服务器监听 `127.0.0.1` 的随机端口
- 静态目录下的每个子目录为一个应用，入口为 `index.html`
- 可通过 `xxx.name` 文件自定义应用名称，`icon.png` 作为应用图标
- 前端通过 iframe 加载应用 URL

### 文件对话框

文件/目录选择对话框通过 PowerShell 调用 Windows Forms 实现，避免了 Wails COM 对话框在某些场景下的崩溃问题。

## 代码规范

- 每个 Vue 文件不超过 1000 行，超过建议拆分
- 每个 Go 文件不超过 600 行，超过建议拆分
- Go 代码符合 Go 语言规范
- Vue 代码符合 Vue3 Composition API 规范
- UI 组件使用 Element Plus 规范

## 作者

angeya — 1571858518@qq.com
