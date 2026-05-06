<template>
  <div class="app-container">
    <div class="left-panel">
      <div class="nav-column">
        <div class="nav-logo">
          <span class="logo-text">onePanel</span>
        </div>
        <div class="nav-menu">
          <div
            v-for="item in navItems"
            :key="item.key"
            class="nav-item"
            :class="{ active: activeNav === item.key }"
            :title="item.label"
            @click="switchNav(item.key)"
          >
            <el-icon :size="20"><component :is="item.icon" /></el-icon>
          </div>
        </div>
        <div class="nav-bottom">
          <div class="version-info" title="onePanel v0.0.1">v0.0.1</div>
        </div>
      </div>
      <div class="sub-panel">
        <div v-if="activeNav === 'terminal'" class="sub-panel-content">
          <div class="sub-panel-title">终端</div>
          <el-tabs v-model="terminalSubTab" class="sub-tabs">
            <el-tab-pane label="快速启动" name="shortcuts">
              <ShortcutPanel @execute-command="handleTerminalQlExec" />
            </el-tab-pane>
            <el-tab-pane label="历史" name="history">
              <HistoryPanel @execute-command="handleTerminalHistoryExec" />
            </el-tab-pane>
          </el-tabs>
        </div>

        <div v-if="activeNav === 'apps'" class="sub-panel-content">
          <div class="sub-panel-header">
            <span class="sub-panel-title">我的应用</span>
            <el-button size="small" @click="showAppSettings" plain>
              <el-icon><Setting /></el-icon>
            </el-button>
          </div>
          <div class="sub-panel-toolbar">
            <el-tag v-if="serverStatus.running" type="success" size="small">
              :{{ serverStatus.port }}
            </el-tag>
            <el-tag v-else type="info" size="small">未启动</el-tag>
            <el-button size="small" @click="showAppImport" plain>
              <el-icon><Upload /></el-icon>
            </el-button>
            <el-button size="small" @click="refreshApps" plain>
              <el-icon><Refresh /></el-icon>
            </el-button>
          </div>
          <div class="app-sidebar-list" v-loading="appsLoading">
            <div
              v-for="app in apps"
              :key="app.id"
              class="app-sidebar-item"
              @click="openApp(app)"
            >
              <div class="app-sidebar-icon">
                <img v-if="app.iconPath" :src="getAppIconUrl(app)" alt="" />
                <el-icon v-else :size="22" color="#409eff"><Document /></el-icon>
              </div>
              <div class="app-sidebar-info">
                <div class="app-sidebar-name">{{ app.displayName }}</div>
                <div class="app-sidebar-dir">{{ app.dirName }}</div>
              </div>
              <el-dropdown trigger="click" @command="(cmd) => handleAppCmd(cmd, app)" @click.stop>
                <el-icon class="app-sidebar-more" @click.stop><MoreFilled /></el-icon>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit">编辑名称</el-dropdown-item>
                    <el-dropdown-item command="rename">修改目录名</el-dropdown-item>
                    <el-dropdown-item command="icon">上传图标</el-dropdown-item>
                    <el-dropdown-item command="export">导出</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>
                      <span style="color: #f56c6c">删除</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
            <el-empty v-if="apps.length === 0 && !appsLoading" description="暂无应用" :image-size="40" />
          </div>
        </div>

        <div v-if="activeNav === 'shortcuts'" class="sub-panel-content">
          <div class="sub-panel-header">
            <span class="sub-panel-title">快速启动</span>
            <el-button size="small" @click="showQlAddDialog" plain>
              <el-icon><Plus /></el-icon>
            </el-button>
          </div>
          <div class="sub-panel-toolbar">
            <el-button size="small" @click="showQlGroupDialog" plain>
              <el-icon><FolderAdd /></el-icon>
              管理分组
            </el-button>
          </div>
          <div class="ql-sidebar-list">
            <div v-for="group in qlGroups" :key="group.id" class="ql-sidebar-group">
              <div class="ql-group-header" @click="toggleQlGroup(group.id)">
                <el-icon>
                  <ArrowDown v-if="expandedQlGroups.has(group.id)" />
                  <ArrowRight v-else />
                </el-icon>
                <span class="group-name">{{ group.name }}</span>
                <span class="group-count">({{ getQlCmdCount(group.id) }})</span>
              </div>
              <div v-show="expandedQlGroups.has(group.id)" class="ql-group-items">
                <div
                  v-for="cmd in getQlCmdsByGroup(group.id)"
                  :key="cmd.id"
                  class="ql-sidebar-item"
                  @dblclick="executeQlCmd(cmd)"
                >
                  <el-icon :size="16" :color="cmd.shell === 'powershell' ? '#012456' : '#4cc2ff'">
                    <Monitor />
                  </el-icon>
                  <div class="ql-item-info">
                    <div class="ql-item-name">{{ cmd.name }}</div>
                    <div class="ql-item-cmd" :title="cmd.commands">{{ cmd.commands }}</div>
                  </div>
                  <div class="ql-item-actions" @click.stop>
                    <el-icon class="action-icon" @click="executeQlCmd(cmd)"><VideoPlay /></el-icon>
                    <el-icon class="action-icon" @click="editQlCmd(cmd)"><Edit /></el-icon>
                    <el-icon class="action-icon" @click="deleteQlCmd(cmd)"><Delete /></el-icon>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="ungroupedQlCmds.length > 0" class="ql-sidebar-group">
              <div class="ql-group-header" @click="toggleQlGroup('none')">
                <el-icon>
                  <ArrowDown v-if="expandedQlGroups.has('none')" />
                  <ArrowRight v-else />
                </el-icon>
                <span class="group-name">未分组</span>
                <span class="group-count">({{ ungroupedQlCmds.length }})</span>
              </div>
              <div v-show="expandedQlGroups.has('none')" class="ql-group-items">
                <div
                  v-for="cmd in ungroupedQlCmds"
                  :key="cmd.id"
                  class="ql-sidebar-item"
                  @dblclick="executeQlCmd(cmd)"
                >
                  <el-icon :size="16" :color="cmd.shell === 'powershell' ? '#012456' : '#4cc2ff'">
                    <Monitor />
                  </el-icon>
                  <div class="ql-item-info">
                    <div class="ql-item-name">{{ cmd.name }}</div>
                    <div class="ql-item-cmd" :title="cmd.commands">{{ cmd.commands }}</div>
                  </div>
                  <div class="ql-item-actions" @click.stop>
                    <el-icon class="action-icon" @click="executeQlCmd(cmd)"><VideoPlay /></el-icon>
                    <el-icon class="action-icon" @click="editQlCmd(cmd)"><Edit /></el-icon>
                    <el-icon class="action-icon" @click="deleteQlCmd(cmd)"><Delete /></el-icon>
                  </div>
                </div>
              </div>
            </div>

            <el-empty v-if="qlCmds.length === 0" description="暂无快速启动命令" :image-size="40" />
          </div>
        </div>

        <div v-if="activeNav === 'tools'" class="sub-panel-content">
          <div class="sub-panel-title">实用工具</div>
          <div class="tool-sidebar-list">
            <div class="tool-sidebar-item" @click="openTool('port', '网络端口')">
              <el-icon :size="18" color="#409eff"><Connection /></el-icon>
              <span class="tool-name">网络端口</span>
              <el-icon class="tool-arrow"><ArrowRight /></el-icon>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="right-panel">
      <div v-if="tabs.length === 0" class="empty-main">
        <el-empty description="选择左侧功能开始使用" :image-size="80" />
      </div>
      <template v-else>
        <div class="main-tabs-header">
          <div class="tabs-list" ref="tabsListRef">
            <div
              v-for="tab in tabs"
              :key="tab.id"
              class="main-tab-item"
              :class="{ active: activeTabId === tab.id }"
              @click="switchTab(tab.id)"
            >
              <el-icon size="12"><component :is="getTabIcon(tab)" /></el-icon>
              <span class="tab-name">{{ tab.title }}</span>
              <el-icon
                v-if="tab.closable !== false"
                class="tab-close"
                size="12"
                @click.stop="closeTab(tab.id)"
              >
                <Close />
              </el-icon>
            </div>
          </div>
          <el-button
            v-if="activeNav === 'terminal'"
            class="tab-add"
            size="small"
            @click="addTerminalTab"
            :icon="Plus"
            circle
          />
        </div>
        <div class="main-tabs-body">
          <TerminalTab
            v-for="tab in terminalTabs"
            :key="tab.id"
            :tab-id="tab.id"
            :shell="tab.shell || 'cmd.exe'"
            v-show="activeTabId === tab.id"
            @command-executed="handleCommandExecuted"
            @send-command="handleSendCommand"
          />
          <iframe
            v-for="tab in appTabs"
            :key="tab.id"
            v-show="activeTabId === tab.id"
            :src="tab.url"
            class="app-iframe"
            sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
          />
          <QuickLaunchTab
            v-if="quickLaunchTab"
            v-show="activeTabId === quickLaunchTab.id"
            ref="quickLaunchTabRef"
          />
          <ToolsPage
            v-for="tab in toolTabs"
            :key="tab.id"
            v-show="activeTabId === tab.id"
            :embedded="true"
          />
        </div>
      </template>
    </div>

    <el-dialog v-model="appSettingsVisible" title="应用设置" width="500px" :close-on-click-modal="false">
      <el-form label-width="100px">
        <el-form-item label="静态目录">
          <div style="display: flex; gap: 8px; width: 100%">
            <el-input v-model="staticDir" placeholder="选择静态文件目录或手动输入路径" />
            <el-button @click="selectDirectory">选择</el-button>
          </div>
        </el-form-item>
        <el-form-item label="服务状态">
          <div style="display: flex; align-items: center; gap: 12px">
            <el-tag v-if="serverStatus.running" type="success" size="small">
              运行中，端口: {{ serverStatus.port }}
            </el-tag>
            <el-tag v-else type="info" size="small">未启动</el-tag>
            <el-button
              v-if="!serverStatus.running"
              size="small"
              type="primary"
              @click="startServer"
              :disabled="!staticDir"
            >启动服务</el-button>
            <el-button v-else size="small" type="danger" @click="stopServer">停止服务</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="appSettingsVisible = false">关闭</el-button>
        <el-button type="primary" @click="saveStaticDir">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="appImportVisible" title="导入应用" width="500px" :close-on-click-modal="false">
      <el-tabs v-model="appImportTab">
        <el-tab-pane label="导入 ZIP" name="zip">
          <el-form label-width="80px">
            <el-form-item label="ZIP 文件">
              <div style="display: flex; gap: 8px; width: 100%">
                <el-input v-model="importZipPath" placeholder="选择 ZIP 压缩包或手动输入路径" />
                <el-button @click="selectZipFile">选择</el-button>
              </div>
            </el-form-item>
          </el-form>
          <div style="text-align: right">
            <el-button type="primary" @click="doImportZip" :disabled="!importZipPath">导入</el-button>
          </div>
        </el-tab-pane>
        <el-tab-pane label="导入目录" name="dir">
          <el-form label-width="80px">
            <el-form-item label="应用目录">
              <div style="display: flex; gap: 8px; width: 100%">
                <el-input v-model="importDirPath" placeholder="选择包含 index.html 的目录或手动输入路径" />
                <el-button @click="selectImportDir">选择</el-button>
              </div>
            </el-form-item>
            <el-form-item label="应用名称">
              <el-input v-model="importAppName" placeholder="留空则使用目录名称" />
            </el-form-item>
          </el-form>
          <div style="text-align: right">
            <el-button type="primary" @click="doImportDir" :disabled="!importDirPath">导入</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>

    <el-dialog v-model="appEditNameVisible" title="编辑应用名称" width="400px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="应用名称">
          <el-input v-model="appEditNameValue" placeholder="请输入应用名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="appEditNameVisible = false">取消</el-button>
        <el-button type="primary" @click="saveAppDisplayName">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="appRenameDirVisible" title="修改目录名称" width="400px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="目录名称">
          <el-input v-model="appRenameDirValue" placeholder="请输入新的目录名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="appRenameDirVisible = false">取消</el-button>
        <el-button type="primary" @click="saveAppDirName">保存</el-button>
      </template>
    </el-dialog>

    <input ref="iconInputRef" type="file" accept="image/png" style="display: none" @change="handleIconUpload" />

    <el-dialog
      v-model="qlCmdDialogVisible"
      :title="isEditingQlCmd ? '编辑快速启动' : '新增快速启动'"
      width="520px"
      :close-on-click-modal="false"
    >
      <el-form :model="qlCmdForm" label-width="90px" size="default">
        <el-form-item label="命令名称" required>
          <el-input v-model="qlCmdForm.name" placeholder="请输入命令名称" />
        </el-form-item>
        <el-form-item label="所属分组">
          <el-select v-model="qlCmdForm.groupId" placeholder="请选择分组（可选）" clearable style="width: 100%">
            <el-option v-for="group in qlGroups" :key="group.id" :label="group.name" :value="group.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="Shell 类型" required>
          <el-radio-group v-model="qlCmdForm.shell">
            <el-radio value="cmd.exe">CMD</el-radio>
            <el-radio value="powershell">PowerShell</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="工作目录">
          <div style="display: flex; gap: 8px; width: 100%">
            <el-input v-model="qlCmdForm.workDir" placeholder="留空则使用默认目录" />
            <el-button @click="selectWorkDir">选择</el-button>
          </div>
        </el-form-item>
        <el-form-item label="命令内容" required>
          <el-input
            v-model="qlCmdForm.commands"
            type="textarea"
            :rows="5"
            placeholder="每行一条命令"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="qlCmdDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveQlCmd">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="qlGroupDialogVisible" title="管理分组" width="420px" :close-on-click-modal="false">
      <div class="group-manage">
        <div class="group-add-row">
          <el-input v-model="newGroupName" placeholder="输入分组名称" @keyup.enter="addQlGroup" />
          <el-button type="primary" @click="addQlGroup">添加</el-button>
        </div>
        <div class="group-list">
          <div v-for="group in qlGroups" :key="group.id" class="group-manage-item">
            <span>{{ group.name }}</span>
            <el-icon class="action-icon" @click="deleteQlGroup(group)"><Delete /></el-icon>
          </div>
          <el-empty v-if="qlGroups.length === 0" description="暂无分组" :image-size="40" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  Monitor, Grid, Promotion, SetUp, Close, Plus,
  Setting, Upload, Refresh, Document, MoreFilled,
  FolderAdd, ArrowDown, ArrowRight, Edit, Delete,
  VideoPlay, Connection
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import TerminalTab from './components/TerminalTab.vue'
import ShortcutPanel from './components/ShortcutPanel.vue'
import HistoryPanel from './components/HistoryPanel.vue'
import QuickLaunchTab from './components/QuickLaunchTab.vue'
import ToolsPage from './views/ToolsPage.vue'
import {
  GetStaticDir, SetStaticDir, GetServerStatus, StartServer, StopServer,
  GetApps, ScanApps, UpdateDisplayName, UpdateDirName, UploadIcon,
  DeleteApp, ExportApp, ImportZip, ImportDir
} from '../wailsjs/go/main/AppService'
import { OpenDirectoryDialog, OpenFileDialog } from '../wailsjs/go/main/App'
import {
  GetGroups as GetSCGroups, CreateGroup as CreateSCGroup, DeleteGroup as DeleteSCGroup,
  GetCommands as GetSCCommands, CreateCommand as CreateSCCommand,
  UpdateCommand as UpdateSCCommand, DeleteCommand as DeleteSCCommand,
  ExecuteCommand as ExecSCCommand
} from '../wailsjs/go/main/ShortcutCmdService'
import { AddHistory } from '../wailsjs/go/main/HistoryService'

const navItems = [
  { key: 'terminal', label: '终端', icon: Monitor },
  { key: 'apps', label: '我的应用', icon: Grid },
  { key: 'shortcuts', label: '快速启动', icon: Promotion },
  { key: 'tools', label: '实用工具', icon: SetUp }
]

const activeNav = ref('terminal')
const terminalSubTab = ref('shortcuts')

const tabs = ref([])
const activeTabId = ref('')
let tabCounter = 0

const terminalTabs = computed(() => tabs.value.filter(t => t.type === 'terminal'))
const appTabs = computed(() => tabs.value.filter(t => t.type === 'app'))
const quickLaunchTab = computed(() => tabs.value.find(t => t.type === 'quick-launch'))
const toolTabs = computed(() => tabs.value.filter(t => t.type === 'tool'))
const quickLaunchTabRef = ref(null)

/**
 * 切换左侧导航
 */
const switchNav = (key) => {
  activeNav.value = key
  if (key === 'terminal') {
    loadQlPanelData()
  } else if (key === 'apps') {
    loadApps()
    loadServerStatus()
  } else if (key === 'shortcuts') {
    loadQlCmds()
    loadQlGroups()
  }
}

/**
 * 获取tab图标
 */
const getTabIcon = (tab) => {
  switch (tab.type) {
    case 'terminal': return Monitor
    case 'app': return Grid
    case 'quick-launch': return Promotion
    case 'tool': return SetUp
    default: return Monitor
  }
}

/**
 * 添加终端tab
 */
const addTerminalTab = (shell = 'cmd.exe') => {
  tabCounter++
  const id = `terminal-${Date.now()}-${tabCounter}`
  const tab = {
    id,
    type: 'terminal',
    title: `终端 ${tabCounter}`,
    shell,
    closable: true
  }
  tabs.value.push(tab)
  activeTabId.value = id
}

/**
 * 切换tab
 */
const switchTab = (id) => {
  activeTabId.value = id
  const tab = tabs.value.find(t => t.id === id)
  if (tab) {
    if (tab.type === 'terminal') activeNav.value = 'terminal'
    else if (tab.type === 'app') activeNav.value = 'apps'
    else if (tab.type === 'quick-launch') activeNav.value = 'shortcuts'
    else if (tab.type === 'tool') activeNav.value = 'tools'
  }
}

/**
 * 关闭tab
 */
const closeTab = (id) => {
  const index = tabs.value.findIndex(t => t.id === id)
  if (index === -1) return
  tabs.value.splice(index, 1)
  if (tabs.value.length > 0 && activeTabId.value === id) {
    const newIndex = Math.min(index, tabs.value.length - 1)
    activeTabId.value = tabs.value[newIndex].id
  } else if (tabs.value.length === 0) {
    activeTabId.value = ''
  }
}

/**
 * 查找或切换到已存在的tab
 */
const findOrSwitchTab = (type, matchKey, matchValue) => {
  const existing = tabs.value.find(t => t.type === type && t[matchKey] === matchValue)
  if (existing) {
    activeTabId.value = existing.id
    return true
  }
  return false
}

/**
 * 终端快速启动执行
 */
const handleTerminalQlExec = (command) => {
  if (terminalTabs.value.length === 0) {
    addTerminalTab()
  }
  handleSendCommand(activeTabId.value, command)
}

/**
 * 终端历史命令执行
 */
const handleTerminalHistoryExec = (command) => {
  if (terminalTabs.value.length === 0) {
    addTerminalTab()
  }
  handleSendCommand(activeTabId.value, command)
}

/**
 * 命令执行完成
 */
const handleCommandExecuted = (data) => {
  if (data && data.command) {
    AddHistory(data.command, 'cmd.exe', '').catch(() => {})
  }
}

/**
 * 向终端发送命令
 */
const handleSendCommand = (tabId, command) => {
  const event = new CustomEvent('terminal-send-command', {
    detail: { tabId, command }
  })
  window.dispatchEvent(event)
}

/**
 * 加载快捷面板数据（终端侧边栏用）
 */
const loadQlPanelData = () => {}

/**
 * ============== 我的应用相关 ==============
 */
const apps = ref([])
const appsLoading = ref(false)
const serverStatus = ref({ running: false, port: 0, dir: '' })

const appSettingsVisible = ref(false)
const staticDir = ref('')

const appImportVisible = ref(false)
const appImportTab = ref('zip')
const importZipPath = ref('')
const importDirPath = ref('')
const importAppName = ref('')

const appEditNameVisible = ref(false)
const appEditNameValue = ref('')
const editingAppId = ref(null)

const appRenameDirVisible = ref(false)
const appRenameDirValue = ref('')
const renamingAppId = ref(null)

const iconInputRef = ref(null)
const iconUploadingAppId = ref(null)

/**
 * 加载应用列表
 */
const loadApps = async () => {
  appsLoading.value = true
  try {
    apps.value = await GetApps()
  } catch (err) {
    ElMessage.error('加载应用列表失败: ' + err)
  } finally {
    appsLoading.value = false
  }
}

/**
 * 刷新应用列表
 */
const refreshApps = async () => {
  appsLoading.value = true
  try {
    apps.value = await ScanApps()
    ElMessage.success('刷新成功')
  } catch (err) {
    ElMessage.error('刷新失败: ' + err)
  } finally {
    appsLoading.value = false
  }
}

/**
 * 加载服务器状态
 */
const loadServerStatus = async () => {
  try {
    serverStatus.value = await GetServerStatus()
  } catch (err) {
    console.error('获取服务器状态失败:', err)
  }
}

/**
 * 打开应用 - 在右侧tab中展示
 */
const openApp = (app) => {
  if (!serverStatus.value.running) {
    ElMessage.warning('请先启动静态服务器')
    return
  }
  if (findOrSwitchTab('app', 'appId', app.id)) return

  tabCounter++
  const id = `app-${Date.now()}-${tabCounter}`
  tabs.value.push({
    id,
    type: 'app',
    title: app.displayName,
    appId: app.id,
    url: `http://127.0.0.1:${serverStatus.value.port}${app.entryUrl}`,
    closable: true
  })
  activeTabId.value = id
}

/**
 * 获取应用图标URL
 */
const getAppIconUrl = (app) => {
  if (!app.iconPath || !serverStatus.value.running) return ''
  const dir = app.entryUrl.replace('/index.html', '')
  return `http://127.0.0.1:${serverStatus.value.port}${dir}/icon.png`
}

/**
 * 显示应用设置
 */
const showAppSettings = async () => {
  try {
    staticDir.value = await GetStaticDir() || ''
  } catch (err) {
    staticDir.value = ''
  }
  await loadServerStatus()
  appSettingsVisible.value = true
}

/**
 * 选择目录
 */
const selectDirectory = async () => {
  try {
    const dir = await OpenDirectoryDialog('选择静态文件目录')
    if (dir) staticDir.value = dir
  } catch (err) {
    console.error('选择目录失败:', err)
    ElMessage.warning('目录选择对话框打开失败，请手动输入路径')
  }
}

/**
 * 保存静态目录
 */
const saveStaticDir = async () => {
  try {
    await SetStaticDir(staticDir.value)
    ElMessage.success('保存成功')
    await loadServerStatus()
    if (staticDir.value) await refreshApps()
    appSettingsVisible.value = false
  } catch (err) {
    ElMessage.error('保存失败: ' + err)
  }
}

/**
 * 启动服务器
 */
const startServer = async () => {
  try {
    await StartServer()
    await loadServerStatus()
    ElMessage.success('服务已启动')
  } catch (err) {
    ElMessage.error('启动失败: ' + err)
  }
}

/**
 * 停止服务器
 */
const stopServer = async () => {
  try {
    await StopServer()
    await loadServerStatus()
    ElMessage.success('服务已停止')
  } catch (err) {
    ElMessage.error('停止失败: ' + err)
  }
}

/**
 * 显示导入对话框
 */
const showAppImport = () => {
  importZipPath.value = ''
  importDirPath.value = ''
  importAppName.value = ''
  appImportVisible.value = true
}

const selectZipFile = async () => {
  try {
    const path = await OpenFileDialog('选择 ZIP 文件', 'ZIP 文件 (*.zip)')
    if (path) importZipPath.value = path
  } catch (err) {
    console.error('选择文件失败:', err)
    ElMessage.warning('文件选择对话框打开失败，请手动输入路径')
  }
}

const selectImportDir = async () => {
  try {
    const dir = await OpenDirectoryDialog('选择应用目录')
    if (dir) importDirPath.value = dir
  } catch (err) {
    console.error('选择目录失败:', err)
    ElMessage.warning('目录选择对话框打开失败，请手动输入路径')
  }
}

const doImportZip = async () => {
  try {
    await ImportZip(importZipPath.value)
    ElMessage.success('导入成功')
    appImportVisible.value = false
    await refreshApps()
  } catch (err) {
    ElMessage.error('导入失败: ' + err)
  }
}

const doImportDir = async () => {
  try {
    await ImportDir(importDirPath.value, importAppName.value)
    ElMessage.success('导入成功')
    appImportVisible.value = false
    await refreshApps()
  } catch (err) {
    ElMessage.error('导入失败: ' + err)
  }
}

/**
 * 处理应用操作命令
 */
const handleAppCmd = (command, app) => {
  switch (command) {
    case 'edit':
      editingAppId.value = app.id
      appEditNameValue.value = app.displayName
      appEditNameVisible.value = true
      break
    case 'rename':
      renamingAppId.value = app.id
      appRenameDirValue.value = app.dirName
      appRenameDirVisible.value = true
      break
    case 'icon':
      iconUploadingAppId.value = app.id
      iconInputRef.value?.click()
      break
    case 'export':
      doExportApp(app)
      break
    case 'delete':
      doDeleteApp(app)
      break
  }
}

const saveAppDisplayName = async () => {
  try {
    await UpdateDisplayName(editingAppId.value, appEditNameValue.value)
    ElMessage.success('修改成功')
    appEditNameVisible.value = false
    await loadApps()
  } catch (err) {
    ElMessage.error('修改失败: ' + err)
  }
}

const saveAppDirName = async () => {
  try {
    await UpdateDirName(renamingAppId.value, appRenameDirValue.value)
    ElMessage.success('修改成功')
    appRenameDirVisible.value = false
    await loadApps()
  } catch (err) {
    ElMessage.error('修改失败: ' + err)
  }
}

const handleIconUpload = async (event) => {
  const file = event.target.files[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = async (e) => {
    const data = new Uint8Array(e.target.result)
    try {
      await UploadIcon(iconUploadingAppId.value, Array.from(data))
      ElMessage.success('图标上传成功')
      await loadApps()
    } catch (err) {
      ElMessage.error('上传失败: ' + err)
    }
  }
  reader.readAsArrayBuffer(file)
  event.target.value = ''
}

const doExportApp = async (app) => {
  try {
    const zipPath = await ExportApp(app.id)
    ElMessage.success(`已导出到: ${zipPath}`)
  } catch (err) {
    ElMessage.error('导出失败: ' + err)
  }
}

const doDeleteApp = async (app) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除应用 "${app.displayName}" 吗？此操作将同时删除应用文件，不可恢复。`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteApp(app.id)
    ElMessage.success('删除成功')
    const appTab = tabs.value.find(t => t.type === 'app' && t.appId === app.id)
    if (appTab) closeTab(appTab.id)
    await loadApps()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败: ' + err)
    }
  }
}

/**
 * ============== 快速启动相关 ==============
 */
const qlGroups = ref([])
const qlCmds = ref([])
const expandedQlGroups = ref(new Set(['none']))

const qlCmdDialogVisible = ref(false)
const isEditingQlCmd = ref(false)
const editingQlCmdId = ref(null)
const qlCmdForm = ref({
  name: '',
  groupId: null,
  shell: 'cmd.exe',
  workDir: '',
  commands: ''
})

const qlGroupDialogVisible = ref(false)
const newGroupName = ref('')

const ungroupedQlCmds = computed(() => qlCmds.value.filter(cmd => !cmd.groupId))

const getQlCmdsByGroup = (groupId) => qlCmds.value.filter(cmd => cmd.groupId === groupId)
const getQlCmdCount = (groupId) => qlCmds.value.filter(cmd => cmd.groupId === groupId).length

const toggleQlGroup = (groupId) => {
  const newSet = new Set(expandedQlGroups.value)
  if (newSet.has(groupId)) newSet.delete(groupId)
  else newSet.add(groupId)
  expandedQlGroups.value = newSet
}

const loadQlGroups = async () => {
  try {
    qlGroups.value = await GetSCGroups() || []
  } catch (err) {
    ElMessage.error('加载分组失败: ' + err)
  }
}

const loadQlCmds = async () => {
  try {
    qlCmds.value = await GetSCCommands() || []
  } catch (err) {
    ElMessage.error('加载命令失败: ' + err)
  }
}

/**
 * 执行快速启动命令 - 确保快速启动tab存在并调用执行
 */
const executeQlCmd = async (cmd) => {
  let tab = quickLaunchTab.value
  if (!tab) {
    tabCounter++
    tab = {
      id: `quick-launch-${Date.now()}-${tabCounter}`,
      type: 'quick-launch',
      title: '快速启动',
      closable: true
    }
    tabs.value.push(tab)
  }
  activeTabId.value = tab.id

  setTimeout(() => {
    if (quickLaunchTabRef.value) {
      quickLaunchTabRef.value.execute(cmd)
    }
  }, 50)
}

const showQlAddDialog = () => {
  isEditingQlCmd.value = false
  editingQlCmdId.value = null
  qlCmdForm.value = { name: '', groupId: null, shell: 'cmd.exe', workDir: '', commands: '' }
  qlCmdDialogVisible.value = true
}

const editQlCmd = (cmd) => {
  isEditingQlCmd.value = true
  editingQlCmdId.value = cmd.id
  qlCmdForm.value = {
    name: cmd.name,
    groupId: cmd.groupId || null,
    shell: cmd.shell || 'cmd.exe',
    workDir: cmd.workDir || '',
    commands: cmd.commands
  }
  qlCmdDialogVisible.value = true
}

const selectWorkDir = async () => {
  try {
    const dir = await OpenDirectoryDialog('选择工作目录')
    if (dir) qlCmdForm.value.workDir = dir
  } catch (err) {
    console.error('选择目录失败:', err)
    ElMessage.warning('目录选择对话框打开失败，请手动输入路径')
  }
}

const saveQlCmd = async () => {
  if (!qlCmdForm.value.name) { ElMessage.warning('请输入命令名称'); return }
  if (!qlCmdForm.value.commands) { ElMessage.warning('请输入命令内容'); return }

  try {
    if (isEditingQlCmd.value) {
      await UpdateSCCommand(
        editingQlCmdId.value,
        qlCmdForm.value.groupId,
        qlCmdForm.value.name,
        qlCmdForm.value.shell,
        qlCmdForm.value.workDir,
        qlCmdForm.value.commands,
        0
      )
      ElMessage.success('更新成功')
    } else {
      await CreateSCCommand(
        qlCmdForm.value.groupId,
        qlCmdForm.value.name,
        qlCmdForm.value.shell,
        qlCmdForm.value.workDir,
        qlCmdForm.value.commands,
        0
      )
      ElMessage.success('创建成功')
    }
    qlCmdDialogVisible.value = false
    await loadQlCmds()
  } catch (err) {
    ElMessage.error('保存失败: ' + err)
  }
}

const deleteQlCmd = async (cmd) => {
  try {
    await ElMessageBox.confirm(
      `确定删除快速启动命令 "${cmd.name}" 吗？`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteSCCommand(cmd.id)
    ElMessage.success('删除成功')
    await loadQlCmds()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败: ' + err)
    }
  }
}

const showQlGroupDialog = () => {
  newGroupName.value = ''
  qlGroupDialogVisible.value = true
}

const addQlGroup = async () => {
  if (!newGroupName.value.trim()) { ElMessage.warning('请输入分组名称'); return }
  try {
    await CreateSCGroup(newGroupName.value.trim(), 0)
    newGroupName.value = ''
    ElMessage.success('分组创建成功')
    await loadQlGroups()
  } catch (err) {
    ElMessage.error('创建分组失败: ' + err)
  }
}

const deleteQlGroup = async (group) => {
  try {
    await ElMessageBox.confirm(
      `删除分组 "${group.name}" 后，该分组下的命令将变为未分组，确定删除？`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteSCGroup(group.id)
    ElMessage.success('分组删除成功')
    await loadQlGroups()
    await loadQlCmds()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败: ' + err)
    }
  }
}

/**
 * ============== 实用工具相关 ==============
 */
const openTool = (toolKey, toolName) => {
  if (findOrSwitchTab('tool', 'toolKey', toolKey)) return

  tabCounter++
  const id = `tool-${Date.now()}-${tabCounter}`
  tabs.value.push({
    id,
    type: 'tool',
    title: toolName,
    toolKey,
    closable: true
  })
  activeTabId.value = id
}

onMounted(() => {
  addTerminalTab()
  loadServerStatus()
  loadApps()
  loadQlGroups()
  loadQlCmds()
})
</script>

<style scoped>
.app-container {
  display: flex;
  height: 100vh;
  width: 100%;
  overflow: hidden;
}

.left-panel {
  display: flex;
  width: auto;
  flex-shrink: 0;
  background-color: #1e1e1e;
  border-right: 1px solid #2d2d2d;
}

.nav-column {
  width: 56px;
  display: flex;
  flex-direction: column;
  align-items: center;
  border-right: 1px solid #2d2d2d;
  flex-shrink: 0;
}

.nav-logo {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #2d2d2d;
  width: 100%;
}

.logo-text {
  font-size: 11px;
  font-weight: 700;
  color: #409eff;
  letter-spacing: 0.5px;
  writing-mode: vertical-rl;
  text-orientation: mixed;
}

.nav-menu {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
  gap: 4px;
}

.nav-item {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  cursor: pointer;
  color: #808080;
  transition: all 0.15s;
}

.nav-item:hover {
  color: #c0c0c0;
  background-color: #2d2d2d;
}

.nav-item.active {
  color: #409eff;
  background-color: #2d2d2d;
}

.nav-bottom {
  padding: 8px 0 12px;
}

.version-info {
  font-size: 10px;
  color: #555;
  text-align: center;
  cursor: default;
}

.sub-panel {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sub-panel-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.sub-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 12px 0;
}

.sub-panel-title {
  font-size: 13px;
  font-weight: 600;
  color: #e5e5e5;
  padding: 12px 12px 0;
}

.sub-panel-header .sub-panel-title {
  padding: 0;
}

.sub-panel-toolbar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  flex-shrink: 0;
}

.sub-tabs {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding: 0 8px;
}

.sub-tabs :deep(.el-tabs__header) {
  margin: 0;
  flex-shrink: 0;
}

.sub-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #2d2d2d;
}

.sub-tabs :deep(.el-tabs__item) {
  color: #a0a0a0;
  font-size: 12px;
  padding: 0 12px;
  height: 32px;
  line-height: 32px;
}

.sub-tabs :deep(.el-tabs__item.is-active) {
  color: #409eff;
}

.sub-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
  padding: 0;
}

.sub-tabs :deep(.el-tab-pane) {
  height: 100%;
  overflow: hidden;
}

/* 应用侧边栏列表 */
.app-sidebar-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 8px;
}

.app-sidebar-list::-webkit-scrollbar {
  width: 4px;
}

.app-sidebar-list::-webkit-scrollbar-thumb {
  background-color: #555;
  border-radius: 2px;
}

.app-sidebar-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  position: relative;
}

.app-sidebar-item:hover {
  background-color: #2d2d2d;
}

.app-sidebar-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background-color: #252526;
  flex-shrink: 0;
  overflow: hidden;
}

.app-sidebar-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.app-sidebar-info {
  flex: 1;
  min-width: 0;
}

.app-sidebar-name {
  font-size: 13px;
  color: #e5e5e5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-sidebar-dir {
  font-size: 11px;
  color: #666;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-top: 1px;
}

.app-sidebar-more {
  color: #555;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  opacity: 0;
  transition: opacity 0.15s;
}

.app-sidebar-item:hover .app-sidebar-more {
  opacity: 1;
}

.app-sidebar-more:hover {
  color: #e5e5e5;
  background-color: #3d3d3d;
}

/* 快速启动侧边栏列表 */
.ql-sidebar-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 8px;
}

.ql-sidebar-list::-webkit-scrollbar {
  width: 4px;
}

.ql-sidebar-list::-webkit-scrollbar-thumb {
  background-color: #555;
  border-radius: 2px;
}

.ql-sidebar-group {
  margin-bottom: 2px;
}

.ql-group-header {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  cursor: pointer;
  border-radius: 4px;
  color: #c0c0c0;
  font-size: 13px;
}

.ql-group-header:hover {
  background-color: #2d2d2d;
}

.ql-group-header .group-name {
  font-weight: 500;
  flex: 1;
}

.ql-group-header .group-count {
  color: #666;
  font-size: 11px;
}

.ql-group-items {
  padding-left: 8px;
}

.ql-sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.ql-sidebar-item:hover {
  background-color: #2d2d2d;
}

.ql-item-info {
  flex: 1;
  min-width: 0;
}

.ql-item-name {
  font-size: 13px;
  color: #e5e5e5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ql-item-cmd {
  font-size: 11px;
  color: #666;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: Consolas, 'Courier New', monospace;
  margin-top: 1px;
}

.ql-item-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.15s;
  flex-shrink: 0;
}

.ql-sidebar-item:hover .ql-item-actions {
  opacity: 1;
}

.action-icon {
  cursor: pointer;
  color: #a0a0a0;
  padding: 2px;
  border-radius: 4px;
  font-size: 14px;
}

.action-icon:hover {
  color: #e5e5e5;
  background-color: #3d3d3d;
}

/* 实用工具侧边栏列表 */
.tool-sidebar-list {
  padding: 8px;
}

.tool-sidebar-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  color: #c0c0c0;
  transition: all 0.15s;
}

.tool-sidebar-item:hover {
  background-color: #2d2d2d;
  color: #e5e5e5;
}

.tool-name {
  flex: 1;
  font-size: 13px;
}

.tool-arrow {
  color: #555;
  font-size: 12px;
}

/* 右侧面板 */
.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background-color: #252526;
  min-width: 0;
}

.empty-main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.main-tabs-header {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  background-color: #1e1e1e;
  border-bottom: 1px solid #2d2d2d;
  gap: 4px;
  flex-shrink: 0;
}

.tabs-list {
  display: flex;
  gap: 2px;
  flex: 1;
  overflow-x: auto;
}

.tabs-list::-webkit-scrollbar {
  height: 3px;
}

.tabs-list::-webkit-scrollbar-thumb {
  background-color: #555;
  border-radius: 2px;
}

.main-tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  white-space: nowrap;
  font-size: 13px;
  color: #a0a0a0;
  background-color: #2d2d2d;
  transition: all 0.15s;
  min-width: 0;
}

.main-tab-item:hover {
  background-color: #3d3d3d;
  color: #e5e5e5;
}

.main-tab-item.active {
  background-color: #3d3d3d;
  color: #e5e5e5;
  border-bottom: 2px solid #409eff;
}

.tab-name {
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  border-radius: 50%;
  padding: 2px;
  opacity: 0;
  transition: opacity 0.15s;
}

.main-tab-item:hover .tab-close {
  opacity: 1;
}

.tab-close:hover {
  background-color: #555;
  color: #fff;
}

.tab-add {
  flex-shrink: 0;
  background-color: #2d2d2d !important;
  border-color: #3d3d3d !important;
  color: #a0a0a0 !important;
}

.tab-add:hover {
  color: #409eff !important;
}

.main-tabs-body {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.app-iframe {
  width: 100%;
  height: 100%;
  border: none;
  position: absolute;
  top: 0;
  left: 0;
}

/* 分组管理弹窗 */
.group-manage {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.group-add-row {
  display: flex;
  gap: 8px;
}

.group-add-row .el-input {
  flex: 1;
}

.group-list {
  max-height: 300px;
  overflow-y: auto;
}

.group-manage-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  color: #e5e5e5;
  font-size: 13px;
}

.group-manage-item:hover {
  background-color: #2d2d2d;
}
</style>
