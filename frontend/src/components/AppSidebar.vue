<template>
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
          @click="$emit('switchNav', item.key)"
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
        <el-tabs v-model="localTerminalSubTab" class="sub-tabs">
          <el-tab-pane label="快捷命令" name="shortcuts">
            <ShortcutPanel @execute-command="$emit('terminalQlExec', $event)" />
          </el-tab-pane>
          <el-tab-pane label="历史" name="history">
            <HistoryPanel @execute-command="$emit('terminalHistoryExec', $event)" />
          </el-tab-pane>
          <el-tab-pane label="搜索" name="search">
            <SearchPanel :active-tab-id="activeTabId" />
          </el-tab-pane>
        </el-tabs>
      </div>

      <div v-if="activeNav === 'apps'" class="sub-panel-content">
        <div class="sub-panel-header">
          <span class="sub-panel-title">我的应用</span>
          <el-button size="small" @click="$emit('showAppSettings')" plain>
            <el-icon><Setting /></el-icon>
          </el-button>
        </div>
        <div class="sub-panel-toolbar">
          <el-tag v-if="serverStatus.running" type="success" size="small">
            :{{ serverStatus.port }}
          </el-tag>
          <el-tag v-else type="info" size="small">未启动</el-tag>
          <el-button size="small" @click="$emit('showAppImport')" plain>
            <el-icon><Upload /></el-icon>
          </el-button>
          <el-button size="small" @click="$emit('refreshApps')" plain>
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
        <div class="app-sidebar-list" v-loading="appsLoading">
          <div
            v-for="app in apps"
            :key="app.id"
            class="app-sidebar-item"
            @click="$emit('openApp', app)"
          >
            <div class="app-sidebar-icon">
              <img v-if="app.iconPath" :src="getAppIconUrl(app)" alt="" />
              <el-icon v-else :size="22" color="#409eff"><Document /></el-icon>
            </div>
            <div class="app-sidebar-info">
              <div class="app-sidebar-name">{{ app.displayName }}</div>
              <div class="app-sidebar-dir">{{ app.dirName }}</div>
            </div>
            <el-dropdown trigger="click" @command="(cmd) => $emit('handleAppCmd', cmd, app)" @click.stop>
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
          <el-button size="small" @click="$emit('showQlAddDialog')" plain>
            <el-icon><Plus /></el-icon>
          </el-button>
        </div>
        <div class="sub-panel-toolbar">
          <el-button size="small" @click="$emit('showQlGroupDialog')" plain>
            <el-icon><FolderAdd /></el-icon>
            管理分组
          </el-button>
        </div>
        <div class="ql-sidebar-list">
          <div v-for="group in qlGroups" :key="group.id" class="ql-sidebar-group">
            <div class="ql-group-header" @click="$emit('toggleQlGroup', group.id)">
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
                @dblclick="$emit('executeQlCmd', cmd)"
              >
                <el-icon :size="16" :color="cmd.shell === 'powershell' ? '#012456' : '#4cc2ff'">
                  <Monitor />
                </el-icon>
                <div class="ql-item-info">
                  <div class="ql-item-name">{{ cmd.name }}</div>
                  <div class="ql-item-cmd" :title="cmd.commands">{{ cmd.commands }}</div>
                </div>
                <div class="ql-item-actions" @click.stop>
                  <el-icon class="action-icon" @click="$emit('executeQlCmd', cmd)"><VideoPlay /></el-icon>
                  <el-icon class="action-icon" @click="$emit('editQlCmd', cmd)"><Edit /></el-icon>
                  <el-icon class="action-icon" @click="$emit('deleteQlCmd', cmd)"><Delete /></el-icon>
                </div>
              </div>
            </div>
          </div>

          <div v-if="ungroupedQlCmds.length > 0" class="ql-sidebar-group">
            <div class="ql-group-header" @click="$emit('toggleQlGroup', 'none')">
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
                @dblclick="$emit('executeQlCmd', cmd)"
              >
                <el-icon :size="16" :color="cmd.shell === 'powershell' ? '#012456' : '#4cc2ff'">
                  <Monitor />
                </el-icon>
                <div class="ql-item-info">
                  <div class="ql-item-name">{{ cmd.name }}</div>
                  <div class="ql-item-cmd" :title="cmd.commands">{{ cmd.commands }}</div>
                </div>
                <div class="ql-item-actions" @click.stop>
                  <el-icon class="action-icon" @click="$emit('executeQlCmd', cmd)"><VideoPlay /></el-icon>
                  <el-icon class="action-icon" @click="$emit('editQlCmd', cmd)"><Edit /></el-icon>
                  <el-icon class="action-icon" @click="$emit('deleteQlCmd', cmd)"><Delete /></el-icon>
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
          <div class="tool-sidebar-item" @click="$emit('openTool', 'port', '网络端口')">
            <el-icon :size="18" color="#409eff"><Connection /></el-icon>
            <span class="tool-name">网络端口</span>
            <el-icon class="tool-arrow"><ArrowRight /></el-icon>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import {
  Monitor, Grid, Promotion, SetUp, Plus,
  Setting, Upload, Refresh, Document, MoreFilled,
  FolderAdd, ArrowDown, ArrowRight, Edit, Delete,
  VideoPlay, Connection
} from '@element-plus/icons-vue'
import ShortcutPanel from './ShortcutPanel.vue'
import HistoryPanel from './HistoryPanel.vue'
import SearchPanel from './SearchPanel.vue'

const props = defineProps({
  activeNav: { type: String, required: true },
  activeTabId: { type: String, default: '' },
  terminalSubTab: { type: String, default: 'shortcuts' },
  navItems: { type: Array, required: true },
  apps: { type: Array, default: () => [] },
  appsLoading: { type: Boolean, default: false },
  serverStatus: { type: Object, default: () => ({ running: false, port: 0 }) },
  getAppIconUrl: { type: Function, default: () => '' },
  qlGroups: { type: Array, default: () => [] },
  qlCmds: { type: Array, default: () => [] },
  expandedQlGroups: { type: Set, default: () => new Set() },
  ungroupedQlCmds: { type: Array, default: () => [] },
  getQlCmdsByGroup: { type: Function, default: () => [] },
  getQlCmdCount: { type: Function, default: () => 0 }
})

const emit = defineEmits([
  'update:terminalSubTab',
  'switchNav',
  'terminalQlExec',
  'terminalHistoryExec',
  'showAppSettings',
  'showAppImport',
  'refreshApps',
  'openApp',
  'handleAppCmd',
  'showQlAddDialog',
  'showQlGroupDialog',
  'toggleQlGroup',
  'executeQlCmd',
  'editQlCmd',
  'deleteQlCmd',
  'openTool'
])

const localTerminalSubTab = computed({
  get: () => props.terminalSubTab,
  set: (val) => emit('update:terminalSubTab', val)
})
</script>

<style scoped>
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
</style>
