<template>
  <div class="shortcut-panel">
    <div class="panel-header">
      <el-button size="small" type="primary" @click="commandDialogRef.show()" plain>
        <el-icon><Plus /></el-icon>
        新增
      </el-button>
      <el-button size="small" @click="categoryDialogRef.show()" plain>
        <el-icon><FolderAdd /></el-icon>
        管理分类
      </el-button>
    </div>

    <div class="shortcut-list">
      <div v-for="category in categories" :key="category.id" class="category-group">
        <div class="category-header" @click="toggleCategory(category.id)">
          <el-icon>
            <ArrowDown v-if="expandedCategories.has(category.id)" />
            <ArrowRight v-else />
          </el-icon>
          <span class="category-name">{{ category.name }}</span>
          <span class="category-count">({{ getCommandCount(category.id) }})</span>
        </div>
        <div v-show="expandedCategories.has(category.id)" class="category-commands">
          <div
            v-for="cmd in getCommandsByCategory(category.id)"
            :key="cmd.id"
            class="command-item"
            @dblclick="executeCommand(cmd)"
          >
            <div class="command-info">
              <span class="command-name">{{ cmd.name }}</span>
              <span class="command-text">{{ cmd.commands }}</span>
            </div>
            <div class="command-actions">
              <el-icon class="action-icon" @click.stop="commandDialogRef.show(cmd)"><Edit /></el-icon>
              <el-icon class="action-icon" @click.stop="deleteCommand(cmd.id)"><Delete /></el-icon>
            </div>
          </div>
        </div>
      </div>

      <div
        v-for="cmd in uncategorizedCommands"
        :key="cmd.id"
        class="command-item"
        @dblclick="executeCommand(cmd)"
      >
        <div class="command-info">
          <span class="command-name">{{ cmd.name }}</span>
          <span class="command-text">{{ cmd.commands }}</span>
        </div>
        <div class="command-actions">
          <el-icon class="action-icon" @click.stop="commandDialogRef.show(cmd)"><Edit /></el-icon>
          <el-icon class="action-icon" @click.stop="deleteCommand(cmd.id)"><Delete /></el-icon>
        </div>
      </div>

      <el-empty v-if="commands.length === 0" description="暂无快捷命令" :image-size="60" />
    </div>

    <ShortcutCommandDialog
      ref="commandDialogRef"
      :categories="categories"
      @saved="onDialogSaved"
    />
    <ShortcutCategoryDialog
      ref="categoryDialogRef"
      :categories="categories"
      @saved="onDialogSaved"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Plus, FolderAdd, Edit, Delete, ArrowDown, ArrowRight } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  GetCategories,
  GetCommands,
  DeleteCommand as DeleteShortcutCommand
} from '../../../wailsjs/go/main/ShortcutService'
import ShortcutCommandDialog from './ShortcutCommandDialog.vue'
import ShortcutCategoryDialog from './ShortcutCategoryDialog.vue'

const emit = defineEmits(['executeCommand'])

const categories = ref([])
const commands = ref([])
const expandedCategories = ref(new Set())
const commandDialogRef = ref(null)
const categoryDialogRef = ref(null)

const uncategorizedCommands = computed(() => {
  return commands.value.filter((cmd) => !cmd.categoryId)
})

const getCommandsByCategory = (categoryId) => {
  return commands.value.filter((cmd) => cmd.categoryId === categoryId)
}

const getCommandCount = (categoryId) => {
  return commands.value.filter((cmd) => cmd.categoryId === categoryId).length
}

const toggleCategory = (categoryId) => {
  const newSet = new Set(expandedCategories.value)
  if (newSet.has(categoryId)) {
    newSet.delete(categoryId)
  } else {
    newSet.add(categoryId)
  }
  expandedCategories.value = newSet
}

const loadCategories = async () => {
  try {
    const result = await GetCategories()
    categories.value = result || []
  } catch (err) {
    ElMessage.error('加载分类失败: ' + err)
  }
}

const loadCommands = async () => {
  try {
    const result = await GetCommands()
    commands.value = result || []
  } catch (err) {
    ElMessage.error('加载命令失败: ' + err)
  }
}

const onDialogSaved = async () => {
  await Promise.all([loadCategories(), loadCommands()])
}

const deleteCommand = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除该快捷命令？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await DeleteShortcutCommand(id)
    ElMessage.success('删除成功')
    await loadCommands()
  } catch {
    // 用户取消
  }
}

const executeCommand = (cmd) => {
  const commandLines = cmd.commands.split('\n').filter((line) => line.trim())
  commandLines.forEach((line) => {
    emit('executeCommand', line.trim())
  })
}

onMounted(() => {
  loadCategories()
  loadCommands()
})
</script>

<style scoped>
.shortcut-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  gap: 6px;
  padding: 8px 8px 0;
  margin-bottom: 8px;
  flex-shrink: 0;
}

.shortcut-list {
  flex: 1;
  overflow-y: auto;
}

.shortcut-list::-webkit-scrollbar {
  width: 4px;
}

.shortcut-list::-webkit-scrollbar-thumb {
  background-color: var(--scrollbar-thumb);
  border-radius: 2px;
}

.category-group {
  margin-bottom: 4px;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  cursor: pointer;
  border-radius: 4px;
  color: var(--text-secondary);
  font-size: 13px;
}

.category-header:hover {
  background-color: var(--bg-hover);
}

.category-name {
  font-weight: 500;
}

.category-count {
  color: var(--text-faint);
  font-size: 12px;
}

.category-commands {
  padding-left: 12px;
}

.command-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.command-item:hover {
  background-color: var(--bg-hover);
}

.command-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.command-name {
  font-size: 13px;
  color: var(--text-primary);
}

.command-text {
  font-size: 11px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.command-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.15s;
  flex-shrink: 0;
}

.command-item:hover .command-actions {
  opacity: 1;
}

.action-icon {
  cursor: pointer;
  color: var(--text-muted);
  padding: 2px;
  border-radius: 4px;
}

.action-icon:hover {
  color: var(--text-primary);
  background-color: var(--bg-active);
}
</style>
