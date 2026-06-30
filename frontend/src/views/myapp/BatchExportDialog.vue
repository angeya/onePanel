<template>
  <el-dialog
    v-model="visible"
    title="批量导出应用"
    width="460px"
    :close-on-click-modal="false"
  >
    <div class="batch-export-header">
      <el-checkbox
        :model-value="isAllSelected"
        :indeterminate="isPartialSelected"
        @change="toggleAll"
      >
        全选
      </el-checkbox>
      <span class="batch-export-count">已选 {{ selected.length }} 个</span>
    </div>
    <div class="batch-export-list">
      <div
        v-for="app in apps"
        :key="app.id"
        class="batch-export-item"
        :class="{ selected: selected.includes(app.id) }"
        @click="toggleItem(app.id)"
      >
        <el-checkbox
          :model-value="selected.includes(app.id)"
          @click.stop
          @change="toggleItem(app.id)"
        />
        <el-icon v-if="app.appType === 'web'" :size="16" color="#67c23a"><Link /></el-icon>
        <el-icon v-else :size="16" color="#409eff"><Document /></el-icon>
        <span class="batch-export-name">{{ app.name }}</span>
        <span class="batch-export-dir">{{ app.appType === 'web' ? '网页应用' : app.name }}</span>
      </div>
      <el-empty v-if="apps.length === 0" description="没有可导出的应用" :image-size="40" />
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleDoExport" :disabled="selected.length === 0">导出</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { Document, Link } from '@element-plus/icons-vue'

const appService = inject('appService')

const visible = ref(false)
const selected = ref([])

const apps = computed(() => appService.apps.value)

const isAllSelected = computed(() => {
  return apps.value.length > 0 && selected.value.length === apps.value.length
})

const isPartialSelected = computed(() => {
  return selected.value.length > 0 && selected.value.length < apps.value.length
})

const open = () => {
  if (apps.value.length === 0) return
  selected.value = []
  visible.value = true
}

const toggleItem = (appId) => {
  const idx = selected.value.indexOf(appId)
  if (idx >= 0) {
    selected.value.splice(idx, 1)
  } else {
    selected.value.push(appId)
  }
}

const toggleAll = () => {
  if (selected.value.length === apps.value.length) {
    selected.value = []
  } else {
    selected.value = apps.value.map(app => app.id)
  }
}

const handleDoExport = async () => {
  const ok = await appService.doBatchExport(selected.value)
  if (ok) visible.value = false
}

defineExpose({ open })
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
