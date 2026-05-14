<template>
  <div class="sub-panel-content">
    <div class="sub-panel-title">实用工具</div>
    <div class="tool-list">
      <div class="tool-item" @click="$emit('openTool', 'port', '网络端口')">
        <el-icon :size="18" color="#409eff"><Connection /></el-icon>
        <span class="tool-name">网络端口</span>
        <el-icon class="tool-arrow"><ArrowRight /></el-icon>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Connection, ArrowRight } from '@element-plus/icons-vue'

defineEmits(['openTool'])
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

.tool-list {
  padding: 8px;
}

.tool-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: all 0.15s;
}

.tool-item:hover {
  background-color: var(--bg-hover);
  color: var(--text-primary);
}

.tool-name {
  flex: 1;
  font-size: 13px;
}

.tool-arrow {
  color: var(--text-dimmed);
  font-size: 12px;
}
</style>
