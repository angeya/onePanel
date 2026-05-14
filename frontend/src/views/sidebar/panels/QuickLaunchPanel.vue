<template>
  <div class="sub-panel-content">
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
    <div class="ql-list">
      <div v-for="group in qlService.qlGroups.value" :key="group.id" class="ql-group">
        <div class="ql-group-header" @click="qlService.toggleQlGroup(group.id)">
          <el-icon>
            <ArrowDown v-if="qlService.expandedQlGroups.value.has(group.id)" />
            <ArrowRight v-else />
          </el-icon>
          <span class="group-name">{{ group.name }}</span>
          <span class="group-count">({{ qlService.getQlCmdCount(group.id) }})</span>
        </div>
        <div v-show="qlService.expandedQlGroups.value.has(group.id)" class="ql-group-items">
          <div
            v-for="cmd in qlService.getQlCmdsByGroup(group.id)"
            :key="cmd.id"
            class="ql-item"
            @dblclick="$emit('executeQlCmd', cmd)"
          >
            <el-icon :size="16" :color="cmd.shell === 'powershell.exe' ? '#012456' : '#4cc2ff'">
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

      <div
        v-for="cmd in qlService.ungroupedQlCmds.value"
        :key="cmd.id"
        class="ql-item"
        @dblclick="$emit('executeQlCmd', cmd)"
      >
        <el-icon :size="16" :color="cmd.shell === 'powershell.exe' ? '#012456' : '#4cc2ff'">
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

      <el-empty v-if="qlService.qlCmds.value.length === 0" description="暂无快速启动命令" :image-size="40" />
    </div>
  </div>
</template>

<script setup>
import { inject } from 'vue'
import {
  Monitor, Plus, FolderAdd, ArrowDown, ArrowRight, Edit, Delete, VideoPlay
} from '@element-plus/icons-vue'

const qlService = inject('qlService')

defineEmits([
  'showQlAddDialog',
  'showQlGroupDialog',
  'executeQlCmd',
  'editQlCmd',
  'deleteQlCmd'
])
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

.ql-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 8px;
}

.ql-list::-webkit-scrollbar {
  width: 4px;
}

.ql-list::-webkit-scrollbar-thumb {
  background-color: var(--scrollbar-thumb);
  border-radius: 2px;
}

.ql-group {
  margin-bottom: 2px;
}

.ql-group-header {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  cursor: pointer;
  border-radius: 4px;
  color: var(--text-secondary);
  font-size: 13px;
}

.ql-group-header:hover {
  background-color: var(--bg-hover);
}

.ql-group-header .group-name {
  font-weight: 500;
  flex: 1;
}

.ql-group-header .group-count {
  color: var(--text-faint);
  font-size: 11px;
}

.ql-group-items {
  padding-left: 8px;
}

.ql-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.ql-item:hover {
  background-color: var(--bg-hover);
}

.ql-item-info {
  flex: 1;
  min-width: 0;
}

.ql-item-name {
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ql-item-cmd {
  font-size: 11px;
  color: var(--text-faint);
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

.ql-item:hover .ql-item-actions {
  opacity: 1;
}

.action-icon {
  cursor: pointer;
  color: var(--text-muted);
  padding: 2px;
  border-radius: 4px;
  font-size: 14px;
}

.action-icon:hover {
  color: var(--text-primary);
  background-color: var(--bg-active);
}
</style>
