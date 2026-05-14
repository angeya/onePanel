<template>
  <el-dialog
    :model-value="batchExportVisible"
    @update:model-value="$emit('update:batchExportVisible', $event)"
    title="批量导出应用"
    width="460px"
    :close-on-click-modal="false"
  >
    <div class="batch-export-header">
      <el-checkbox
        :model-value="isAllSelected"
        :indeterminate="isPartialSelected"
        @change="$emit('toggleBatchExportAll')"
      >
        全选
      </el-checkbox>
      <span class="batch-export-count">已选 {{ batchExportSelected.length }} 个</span>
    </div>
    <div class="batch-export-list">
      <div
        v-for="app in apps"
        :key="app.id"
        class="batch-export-item"
        :class="{ selected: batchExportSelected.includes(app.id) }"
        @click="$emit('toggleBatchExportItem', app.id)"
      >
        <el-checkbox
          :model-value="batchExportSelected.includes(app.id)"
          @click.stop
          @change="$emit('toggleBatchExportItem', app.id)"
        />
        <el-icon v-if="app.appType === 'web'" :size="16" color="#67c23a"><Link /></el-icon>
        <el-icon v-else :size="16" color="#409eff"><Document /></el-icon>
        <span class="batch-export-name">{{ app.displayName }}</span>
        <span class="batch-export-dir">{{ app.appType === 'web' ? '网页应用' : app.dirName }}</span>
      </div>
      <el-empty v-if="apps.length === 0" description="没有可导出的应用" :image-size="40" />
    </div>
    <template #footer>
      <el-button @click="$emit('update:batchExportVisible', false)">取消</el-button>
      <el-button type="primary" @click="$emit('doBatchExport')" :disabled="batchExportSelected.length === 0">导出</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'
import { Document, Link } from '@element-plus/icons-vue'

const props = defineProps({
  apps: { type: Array, default: () => [] },
  batchExportVisible: { type: Boolean, default: false },
  batchExportSelected: { type: Array, default: () => [] }
})

defineEmits([
  'update:batchExportVisible',
  'doBatchExport',
  'toggleBatchExportItem',
  'toggleBatchExportAll'
])

const isAllSelected = computed(() => {
  return props.apps.length > 0 && props.batchExportSelected.length === props.apps.length
})

const isPartialSelected = computed(() => {
  return props.batchExportSelected.length > 0 && props.batchExportSelected.length < props.apps.length
})
</script>

<style scoped>
.batch-export-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.batch-export-count {
  font-size: 12px;
  color: var(--text-muted);
}

.batch-export-list {
  max-height: 300px;
  overflow-y: auto;
}

.batch-export-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.batch-export-item:hover {
  background-color: var(--bg-hover);
}

.batch-export-item.selected {
  background-color: var(--bg-active);
}

.batch-export-name {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
}

.batch-export-dir {
  font-size: 11px;
  color: var(--text-faint);
}
</style>
