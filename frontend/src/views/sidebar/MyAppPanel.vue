<template>
  <div class="sub-panel-content">
    <div class="sub-panel-header">
      <span class="sub-panel-title">我的应用</span>
    </div>
    <div class="sub-panel-toolbar">
      <el-button size="small" @click="openMyAppDialog('addWebApp')" plain>
        <el-icon><Link /></el-icon>
        网页
      </el-button>
      <el-button size="small" @click="openMyAppDialog('import')" plain>
        <el-icon><Bottom /></el-icon>
        导入
      </el-button>
      <el-button size="small" @click="openMyAppDialog('batchExport')" plain>
        <el-icon><Top /></el-icon>
        导出
      </el-button>
      <el-button size="small" @click="appService.refreshApps()" plain>
        <el-icon><Refresh /></el-icon>
      </el-button>
    </div>
    <div class="app-list" v-loading="appService.appsLoading.value">
      <div
        v-for="app in appService.apps.value"
        :key="app.id"
        class="app-item"
        @click="openApp(app)"
      >
        <div class="app-icon">
          <el-icon v-if="app.appType === 'web'" :size="22" color="#67c23a"><Link /></el-icon>
          <el-icon v-else :size="22" color="#409eff"><Document /></el-icon>
        </div>
        <div class="app-info">
          <div class="app-name">{{ app.name }}</div>
          <div class="app-dir">{{ app.appType === 'web' ? app.entryUrl : app.name }}</div>
        </div>
        <el-dropdown trigger="click" @command="(cmd) => handleAppCmd(cmd, app)" @click.stop>
          <el-icon class="app-more" @click.stop><MoreFilled /></el-icon>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="edit">{{ app.appType === 'web' ? '编辑' : '编辑名称' }}</el-dropdown-item>
              <el-dropdown-item command="export">导出</el-dropdown-item>
              <el-dropdown-item command="delete" divided>
                <span style="color: #f56c6c">删除</span>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <el-empty v-if="appService.apps.value.length === 0 && !appService.appsLoading.value" description="暂无应用" :image-size="40" />
    </div>
  </div>
</template>

<script setup>
import { inject } from 'vue'
import { Bottom, Top, Refresh, Document, MoreFilled, Link } from '@element-plus/icons-vue'

const appService = inject('appService')
const openMyAppDialog = inject('openMyAppDialog')
const addAppTab = inject('addAppTab')

const openApp = (app) => {
  appService.openApp(app, addAppTab)
}

const handleAppCmd = (cmd, app) => {
  const result = appService.handleAppCmd(cmd, app)
  if (result) {
    openMyAppDialog('handleCmd', { cmd: result.cmd, app: result.app })
  }
}
</script>

<style scoped>
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
  color: var(--text-primary);
}

.sub-panel-toolbar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  flex-shrink: 0;
}

.app-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 8px;
}

.app-list::-webkit-scrollbar {
  width: 4px;
}

.app-list::-webkit-scrollbar-thumb {
  background-color: var(--scrollbar-thumb);
  border-radius: 2px;
}

.app-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  position: relative;
}

.app-item:hover {
  background-color: var(--bg-hover);
}

.app-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background-color: var(--bg-secondary);
  flex-shrink: 0;
}

.app-info {
  flex: 1;
  min-width: 0;
}

.app-name {
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-dir {
  font-size: 11px;
  color: var(--text-faint);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-top: 1px;
}

.app-more {
  color: var(--text-dimmed);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  opacity: 0;
  transition: opacity 0.15s;
}

.app-item:hover .app-more {
  opacity: 1;
}

.app-more:hover {
  color: var(--text-primary);
  background-color: var(--bg-active);
}
</style>
