<template>
  <div class="sub-panel-content">
    <div class="sub-panel-title">终端</div>
    <el-tabs v-model="localSubTab" class="sub-tabs">
      <el-tab-pane label="快捷命令" name="shortcuts">
        <ShortcutPanel @execute-command="handleTerminalCommand" />
      </el-tab-pane>
      <el-tab-pane label="历史" name="history">
        <HistoryPanel @execute-command="handleTerminalCommand" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, inject } from 'vue'
import ShortcutPanel from '../terminal/ShortcutPanel.vue'
import HistoryPanel from '../terminal/HistoryPanel.vue'

const handleTerminalCommand = inject('handleTerminalCommand')

const localSubTab = ref('shortcuts')
</script>

<style scoped>
.sub-panel-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.sub-panel-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  padding: 12px 12px 0;
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
  background-color: var(--border-color);
}

.sub-tabs :deep(.el-tabs__item) {
  color: var(--text-muted);
  font-size: 12px;
  padding: 0 12px;
  height: 32px;
  line-height: 32px;
}

.sub-tabs :deep(.el-tabs__item.is-active) {
  color: var(--accent);
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
</style>
