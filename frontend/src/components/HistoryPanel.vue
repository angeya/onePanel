<template>
  <div class="history-panel">
    <div class="panel-header">
      <el-input
        v-model="searchKeyword"
        size="small"
        placeholder="搜索历史命令..."
        clearable
        @input="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-button size="small" type="danger" plain @click="clearAllHistory">
        <el-icon><Delete /></el-icon>
        清空
      </el-button>
    </div>

    <div class="history-list" v-loading="loading">
      <div
        v-for="item in histories"
        :key="item.id"
        class="history-item"
        @click="executeCommand(item.command)"
      >
        <div class="history-command" :title="`${item.command}\n${item.executedAt}`">{{ item.command }}</div>
        <div class="history-meta">
          <span class="history-shell">{{ formatShell(item.shell) }}</span>
        </div>
        <el-icon class="history-delete" @click.stop="deleteItem(item.id)"><Delete /></el-icon>
      </div>

      <el-empty v-if="histories.length === 0 && !loading" description="暂无历史命令" :image-size="60" />
    </div>

    <div class="history-pagination" v-if="total > pageSize">
      <el-pagination
        v-model:current-page="currentPage"
        :page-size="pageSize"
        :total="total"
        layout="prev, next"
        small
        @current-change="loadHistory"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Search, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  GetHistory,
  SearchHistory,
  ClearHistory,
  DeleteHistory
} from '../../wailsjs/go/main/HistoryService'

const emit = defineEmits(['executeCommand'])

const formatShell = (shell) => {
  if (shell === 'cmd.exe') return 'cmd'
  if (shell === 'powershell') return 'ps'
  return shell.replace('.exe', '')
}

const histories = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = 50
const total = ref(0)
const searchKeyword = ref('')
let searchTimer = null

/**
 * 加载命令历史列表
 */
const loadHistory = async () => {
  loading.value = true
  try {
    const result = await GetHistory(currentPage.value, pageSize)
    histories.value = result.histories || []
    total.value = result.total || 0
  } catch (err) {
    ElMessage.error('加载历史失败: ' + err)
  } finally {
    loading.value = false
  }
}

/**
 * 搜索历史命令
 */
const handleSearch = () => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
  searchTimer = setTimeout(async () => {
    currentPage.value = 1
    loading.value = true
    try {
      if (searchKeyword.value.trim()) {
        const result = await SearchHistory(searchKeyword.value.trim(), currentPage.value, pageSize)
        histories.value = result.histories || []
        total.value = result.total || 0
      } else {
        await loadHistory()
      }
    } catch (err) {
      ElMessage.error('搜索失败: ' + err)
    } finally {
      loading.value = false
    }
  }, 300)
}

/**
 * 点击历史命令执行
 */
const executeCommand = (command) => {
  emit('executeCommand', command)
}

/**
 * 删除单条历史记录
 */
const deleteItem = async (id) => {
  try {
    await DeleteHistory(id)
    await loadHistory()
  } catch (err) {
    ElMessage.error('删除失败: ' + err)
  }
}

/**
 * 清空所有历史记录
 */
const clearAllHistory = async () => {
  try {
    await ElMessageBox.confirm('确定清空所有历史命令？此操作不可恢复。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await ClearHistory()
    ElMessage.success('历史已清空')
    await loadHistory()
  } catch {
    // 用户取消
  }
}

onMounted(() => {
  loadHistory()
})
</script>

<style scoped>
.history-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  gap: 6px;
  margin-bottom: 8px;
  flex-shrink: 0;
}

.panel-header .el-input {
  flex: 1;
}

.history-list {
  flex: 1;
  overflow-y: auto;
}

.history-list::-webkit-scrollbar {
  width: 4px;
}

.history-list::-webkit-scrollbar-thumb {
  background-color: var(--scrollbar-thumb);
  border-radius: 2px;
}

.history-item {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  padding: 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.15s;
  position: relative;
  gap: 4px;
}

.history-item:hover {
  background-color: var(--bg-hover);
}

.history-command {
  flex: 1;
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.history-meta {
  display: flex;
  gap: 8px;
  font-size: 11px;
  color: var(--text-faint);
  flex-shrink: 0;
}

.history-shell {
  color: var(--accent);
}

.history-delete {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  opacity: 0;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  border-radius: 4px;
}

.history-item:hover .history-delete {
  opacity: 1;
}

.history-delete:hover {
  color: var(--danger);
  background-color: var(--bg-active);
}

.history-pagination {
  display: flex;
  justify-content: center;
  padding: 8px 0;
  flex-shrink: 0;
}
</style>
