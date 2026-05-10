<template>
  <div class="shortcut-panel">
    <div class="panel-header">
      <el-button size="small" type="primary" @click="showAddDialog" plain>
        <el-icon><Plus /></el-icon>
        新增
      </el-button>
      <el-button size="small" @click="showCategoryDialog" plain>
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
              <el-icon class="action-icon" @click.stop="showEditDialog(cmd)"><Edit /></el-icon>
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
          <el-icon class="action-icon" @click.stop="showEditDialog(cmd)"><Edit /></el-icon>
          <el-icon class="action-icon" @click.stop="deleteCommand(cmd.id)"><Delete /></el-icon>
        </div>
      </div>

      <el-empty v-if="commands.length === 0" description="暂无快捷命令" :image-size="60" />
    </div>

    <el-dialog
      v-model="commandDialogVisible"
      :title="isEditing ? '编辑快捷命令' : '新增快捷命令'"
      width="450px"
      :close-on-click-modal="false"
      class="command-dialog"
    >
      <el-form :model="commandForm" label-width="80px" size="default">
        <el-form-item label="命令名称">
          <el-input v-model="commandForm.name" placeholder="请输入命令名称" />
        </el-form-item>
        <el-form-item label="所属分类">
          <el-select v-model="commandForm.categoryId" placeholder="请选择分类" clearable style="width: 100%">
            <el-option
              v-for="cat in categories"
              :key="cat.id"
              :label="cat.name"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="Shell">
          <el-select v-model="commandForm.shell" style="width: 100%">
            <el-option label="cmd.exe" value="cmd.exe" />
            <el-option label="powershell.exe" value="powershell.exe" />
          </el-select>
        </el-form-item>
        <el-form-item label="工作目录">
          <el-input v-model="commandForm.workDir" placeholder="留空则使用默认目录" />
        </el-form-item>
        <el-form-item label="命令内容">
          <el-input
            v-model="commandForm.commands"
            type="textarea"
            :rows="4"
            placeholder="每行一条命令"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="commandDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveCommand">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="categoryDialogVisible"
      title="管理分类"
      width="380px"
      class="command-dialog"
    >
      <div class="category-manage">
        <div class="category-add-row">
          <el-input v-model="newCategoryName" placeholder="分类名称" size="default" />
          <el-button type="primary" @click="addCategory" size="default">添加</el-button>
        </div>
        <div class="category-list">
          <div v-for="cat in categories" :key="cat.id" class="category-manage-item">
            <span>{{ cat.name }}</span>
            <el-icon class="action-icon" @click="deleteCategory(cat.id)"><Delete /></el-icon>
          </div>
          <el-empty v-if="categories.length === 0" description="暂无分类" :image-size="40" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Plus, FolderAdd, Edit, Delete, ArrowDown, ArrowRight } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  GetCategories,
  CreateCategory,
  DeleteCategory,
  GetCommands,
  CreateCommand,
  UpdateCommand,
  DeleteCommand as DeleteShortcutCommand
} from '../../../wailsjs/go/main/ShortcutService'

const emit = defineEmits(['executeCommand'])

const categories = ref([])
const commands = ref([])
const expandedCategories = ref(new Set())
const commandDialogVisible = ref(false)
const categoryDialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const newCategoryName = ref('')

const commandForm = ref({
  name: '',
  categoryId: null,
  shell: 'cmd.exe',
  workDir: '',
  commands: ''
})

/**
 * 获取未分类的命令列表
 */
const uncategorizedCommands = computed(() => {
  return commands.value.filter((cmd) => !cmd.categoryId)
})

/**
 * 按分类 ID 获取命令列表
 */
const getCommandsByCategory = (categoryId) => {
  return commands.value.filter((cmd) => cmd.categoryId === categoryId)
}

/**
 * 获取分类下命令数量
 */
const getCommandCount = (categoryId) => {
  return commands.value.filter((cmd) => cmd.categoryId === categoryId).length
}

/**
 * 展开/折叠分类
 */
const toggleCategory = (categoryId) => {
  const newSet = new Set(expandedCategories.value)
  if (newSet.has(categoryId)) {
    newSet.delete(categoryId)
  } else {
    newSet.add(categoryId)
  }
  expandedCategories.value = newSet
}

/**
 * 加载分类数据
 */
const loadCategories = async () => {
  try {
    const result = await GetCategories()
    categories.value = result || []
  } catch (err) {
    ElMessage.error('加载分类失败: ' + err)
  }
}

/**
 * 加载命令数据
 */
const loadCommands = async () => {
  try {
    const result = await GetCommands()
    commands.value = result || []
  } catch (err) {
    ElMessage.error('加载命令失败: ' + err)
  }
}

/**
 * 显示新增命令对话框
 */
const showAddDialog = () => {
  isEditing.value = false
  editingId.value = null
  commandForm.value = {
    name: '',
    categoryId: null,
    shell: 'cmd.exe',
    workDir: '',
    commands: ''
  }
  commandDialogVisible.value = true
}

/**
 * 显示编辑命令对话框
 */
const showEditDialog = (cmd) => {
  isEditing.value = true
  editingId.value = cmd.id
  commandForm.value = {
    name: cmd.name,
    categoryId: cmd.categoryId,
    shell: cmd.shell,
    workDir: cmd.workDir,
    commands: cmd.commands
  }
  commandDialogVisible.value = true
}

/**
 * 保存命令（新增或编辑）
 */
const saveCommand = async () => {
  if (!commandForm.value.name) {
    ElMessage.warning('请输入命令名称')
    return
  }
  if (!commandForm.value.commands) {
    ElMessage.warning('请输入命令内容')
    return
  }

  try {
    if (isEditing.value) {
      await UpdateCommand(
        editingId.value,
        commandForm.value.categoryId,
        commandForm.value.name,
        commandForm.value.shell,
        commandForm.value.workDir,
        commandForm.value.commands,
        0
      )
      ElMessage.success('更新成功')
    } else {
      await CreateCommand(
        commandForm.value.categoryId,
        commandForm.value.name,
        commandForm.value.shell,
        commandForm.value.workDir,
        commandForm.value.commands,
        0
      )
      ElMessage.success('创建成功')
    }
    commandDialogVisible.value = false
    await loadCommands()
  } catch (err) {
    ElMessage.error('保存失败: ' + err)
  }
}

/**
 * 删除快捷命令
 */
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

/**
 * 显示分类管理对话框
 */
const showCategoryDialog = () => {
  newCategoryName.value = ''
  categoryDialogVisible.value = true
}

/**
 * 添加分类
 */
const addCategory = async () => {
  if (!newCategoryName.value.trim()) {
    ElMessage.warning('请输入分类名称')
    return
  }
  try {
    await CreateCategory(newCategoryName.value.trim(), 0)
    newCategoryName.value = ''
    ElMessage.success('分类创建成功')
    await loadCategories()
  } catch (err) {
    ElMessage.error('创建分类失败: ' + err)
  }
}

/**
 * 删除分类
 */
const deleteCategory = async (id) => {
  try {
    await ElMessageBox.confirm('删除分类后，该分类下的命令将变为未分类，确定删除？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await DeleteCategory(id)
    ElMessage.success('分类删除成功')
    await loadCategories()
    await loadCommands()
  } catch {
    // 用户取消
  }
}

/**
 * 双击执行快捷命令
 */
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

.category-manage {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.category-add-row {
  display: flex;
  gap: 8px;
}

.category-add-row .el-input {
  flex: 1;
}

.category-list {
  max-height: 300px;
  overflow-y: auto;
}

.category-manage-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  border-radius: 4px;
  color: var(--text-primary);
  font-size: 13px;
}

.category-manage-item:hover {
  background-color: var(--bg-hover);
}
</style>
