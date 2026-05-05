<template>
  <div class="shortcut-cmd-page">
    <div class="page-header">
      <span class="page-title">快捷命令</span>
      <div class="header-actions">
        <el-button size="small" type="primary" @click="showAddCommandDialog" plain>
          <el-icon><Plus /></el-icon>
          新增命令
        </el-button>
        <el-button size="small" @click="showGroupDialog" plain>
          <el-icon><FolderAdd /></el-icon>
          管理分组
        </el-button>
      </div>
    </div>

    <div class="page-body" v-loading="loading">
      <div v-if="commands.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无快捷命令，点击上方按钮新增" />
      </div>

      <div v-else class="cmd-groups">
        <div v-for="group in groups" :key="group.id" class="cmd-group">
          <div class="group-header" @click="toggleGroup(group.id)">
            <el-icon>
              <ArrowDown v-if="expandedGroups.has(group.id)" />
              <ArrowRight v-else />
            </el-icon>
            <span class="group-name">{{ group.name }}</span>
            <span class="group-count">({{ getCommandCount(group.id) }})</span>
          </div>
          <div v-show="expandedGroups.has(group.id)" class="group-commands">
            <div
              v-for="cmd in getCommandsByGroup(group.id)"
              :key="cmd.id"
              class="cmd-card"
              @dblclick="executeCommand(cmd)"
            >
              <div class="cmd-main">
                <div class="cmd-icon">
                  <el-icon :size="20" :color="cmd.shell === 'powershell' ? '#012456' : '#4cc2ff'">
                    <Monitor />
                  </el-icon>
                </div>
                <div class="cmd-info">
                  <div class="cmd-name">{{ cmd.name }}</div>
                  <div class="cmd-detail">
                    <el-tag size="small" :type="cmd.shell === 'powershell' ? 'info' : 'primary'" class="shell-tag">
                      {{ cmd.shell === 'powershell' ? 'PS' : 'CMD' }}
                    </el-tag>
                    <span class="cmd-commands" :title="cmd.commands">{{ cmd.commands }}</span>
                  </div>
                  <div v-if="cmd.workDir" class="cmd-workdir" :title="cmd.workDir">
                    <el-icon size="12"><Folder /></el-icon>
                    {{ cmd.workDir }}
                  </div>
                </div>
              </div>
              <div class="cmd-actions" @click.stop>
                <el-button size="small" type="primary" text @click="executeCommand(cmd)">
                  <el-icon><VideoPlay /></el-icon>
                </el-button>
                <el-button size="small" text @click="showEditCommandDialog(cmd)">
                  <el-icon><Edit /></el-icon>
                </el-button>
                <el-button size="small" type="danger" text @click="deleteCommand(cmd)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="ungroupedCommands.length > 0" class="cmd-group">
          <div class="group-header" @click="toggleGroup('none')">
            <el-icon>
              <ArrowDown v-if="expandedGroups.has('none')" />
              <ArrowRight v-else />
            </el-icon>
            <span class="group-name">未分组</span>
            <span class="group-count">({{ ungroupedCommands.length }})</span>
          </div>
          <div v-show="expandedGroups.has('none')" class="group-commands">
            <div
              v-for="cmd in ungroupedCommands"
              :key="cmd.id"
              class="cmd-card"
              @dblclick="executeCommand(cmd)"
            >
              <div class="cmd-main">
                <div class="cmd-icon">
                  <el-icon :size="20" :color="cmd.shell === 'powershell' ? '#012456' : '#4cc2ff'">
                    <Monitor />
                  </el-icon>
                </div>
                <div class="cmd-info">
                  <div class="cmd-name">{{ cmd.name }}</div>
                  <div class="cmd-detail">
                    <el-tag size="small" :type="cmd.shell === 'powershell' ? 'info' : 'primary'" class="shell-tag">
                      {{ cmd.shell === 'powershell' ? 'PS' : 'CMD' }}
                    </el-tag>
                    <span class="cmd-commands" :title="cmd.commands">{{ cmd.commands }}</span>
                  </div>
                  <div v-if="cmd.workDir" class="cmd-workdir" :title="cmd.workDir">
                    <el-icon size="12"><Folder /></el-icon>
                    {{ cmd.workDir }}
                  </div>
                </div>
              </div>
              <div class="cmd-actions" @click.stop>
                <el-button size="small" type="primary" text @click="executeCommand(cmd)">
                  <el-icon><VideoPlay /></el-icon>
                </el-button>
                <el-button size="small" text @click="showEditCommandDialog(cmd)">
                  <el-icon><Edit /></el-icon>
                </el-button>
                <el-button size="small" type="danger" text @click="deleteCommand(cmd)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="commandDialogVisible"
      :title="isEditing ? '编辑快捷命令' : '新增快捷命令'"
      width="520px"
      :close-on-click-modal="false"
    >
      <el-form :model="commandForm" label-width="90px" size="default">
        <el-form-item label="命令名称" required>
          <el-input v-model="commandForm.name" placeholder="请输入命令名称，如：启动 JMeter" />
        </el-form-item>
        <el-form-item label="所属分组">
          <el-select v-model="commandForm.groupId" placeholder="请选择分组（可选）" clearable style="width: 100%">
            <el-option
              v-for="group in groups"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="Shell 类型" required>
          <el-radio-group v-model="commandForm.shell">
            <el-radio value="cmd.exe">CMD</el-radio>
            <el-radio value="powershell">PowerShell</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="工作目录">
          <div style="display: flex; gap: 8px; width: 100%">
            <el-input v-model="commandForm.workDir" placeholder="留空则使用默认目录" />
            <el-button @click="selectWorkDir">选择</el-button>
          </div>
        </el-form-item>
        <el-form-item label="命令内容" required>
          <el-input
            v-model="commandForm.commands"
            type="textarea"
            :rows="5"
            placeholder="每行一条命令，如：&#10;cd D:\apache-jmeter\bin&#10;jmeter.bat"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="commandDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveCommand">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="groupDialogVisible"
      title="管理分组"
      width="420px"
      :close-on-click-modal="false"
    >
      <div class="group-manage">
        <div class="group-add-row">
          <el-input v-model="newGroupName" placeholder="输入分组名称" @keyup.enter="addGroup" />
          <el-button type="primary" @click="addGroup">添加</el-button>
        </div>
        <div class="group-list">
          <div v-for="group in groups" :key="group.id" class="group-manage-item">
            <span>{{ group.name }}</span>
            <el-icon class="action-icon" @click="deleteGroup(group)"><Delete /></el-icon>
          </div>
          <el-empty v-if="groups.length === 0" description="暂无分组" :image-size="40" />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Plus, FolderAdd, Edit, Delete, ArrowDown, ArrowRight, Monitor, Folder, VideoPlay } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  GetGroups, CreateGroup, DeleteGroup,
  GetCommands, CreateCommand, UpdateCommand, DeleteCommand as DeleteCmd, ExecuteCommand
} from '../../wailsjs/go/main/ShortcutCmdService'
import { OpenDirectoryDialog } from '../../wailsjs/go/main/App'

const groups = ref([])
const commands = ref([])
const loading = ref(false)
const expandedGroups = ref(new Set(['none']))

const commandDialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref(null)

const commandForm = ref({
  name: '',
  groupId: null,
  shell: 'cmd.exe',
  workDir: '',
  commands: ''
})

const groupDialogVisible = ref(false)
const newGroupName = ref('')

/**
 * 获取未分组的命令
 */
const ungroupedCommands = computed(() => {
  return commands.value.filter(cmd => !cmd.groupId)
})

/**
 * 按分组获取命令
 */
const getCommandsByGroup = (groupId) => {
  return commands.value.filter(cmd => cmd.groupId === groupId)
}

/**
 * 获取分组下命令数量
 */
const getCommandCount = (groupId) => {
  return commands.value.filter(cmd => cmd.groupId === groupId).length
}

/**
 * 展开/折叠分组
 */
const toggleGroup = (groupId) => {
  const newSet = new Set(expandedGroups.value)
  if (newSet.has(groupId)) {
    newSet.delete(groupId)
  } else {
    newSet.add(groupId)
  }
  expandedGroups.value = newSet
}

/**
 * 加载分组数据
 */
const loadGroups = async () => {
  try {
    groups.value = await GetGroups() || []
  } catch (err) {
    ElMessage.error('加载分组失败: ' + err)
  }
}

/**
 * 加载命令数据
 */
const loadCommands = async () => {
  loading.value = true
  try {
    commands.value = await GetCommands() || []
  } catch (err) {
    ElMessage.error('加载命令失败: ' + err)
  } finally {
    loading.value = false
  }
}

/**
 * 显示新增命令对话框
 */
const showAddCommandDialog = () => {
  isEditing.value = false
  editingId.value = null
  commandForm.value = {
    name: '',
    groupId: null,
    shell: 'cmd.exe',
    workDir: '',
    commands: ''
  }
  commandDialogVisible.value = true
}

/**
 * 显示编辑命令对话框
 */
const showEditCommandDialog = (cmd) => {
  isEditing.value = true
  editingId.value = cmd.id
  commandForm.value = {
    name: cmd.name,
    groupId: cmd.groupId || null,
    shell: cmd.shell || 'cmd.exe',
    workDir: cmd.workDir || '',
    commands: cmd.commands
  }
  commandDialogVisible.value = true
}

/**
 * 选择工作目录
 */
const selectWorkDir = async () => {
  try {
    const dir = await OpenDirectoryDialog('选择工作目录')
    if (dir) {
      commandForm.value.workDir = dir
    }
  } catch (err) {
    console.error('选择目录失败:', err)
  }
}

/**
 * 保存命令
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
        commandForm.value.groupId,
        commandForm.value.name,
        commandForm.value.shell,
        commandForm.value.workDir,
        commandForm.value.commands,
        0
      )
      ElMessage.success('更新成功')
    } else {
      await CreateCommand(
        commandForm.value.groupId,
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
 * 删除命令
 */
const deleteCommand = async (cmd) => {
  try {
    await ElMessageBox.confirm(
      `确定删除快捷命令 "${cmd.name}" 吗？`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteCmd(cmd.id)
    ElMessage.success('删除成功')
    await loadCommands()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败: ' + err)
    }
  }
}

/**
 * 执行快捷命令
 */
const executeCommand = async (cmd) => {
  try {
    await ExecuteCommand(cmd.id)
    ElMessage.success(`已执行: ${cmd.name}`)
  } catch (err) {
    ElMessage.error('执行失败: ' + err)
  }
}

/**
 * 显示分组管理对话框
 */
const showGroupDialog = () => {
  newGroupName.value = ''
  groupDialogVisible.value = true
}

/**
 * 添加分组
 */
const addGroup = async () => {
  if (!newGroupName.value.trim()) {
    ElMessage.warning('请输入分组名称')
    return
  }
  try {
    await CreateGroup(newGroupName.value.trim(), 0)
    newGroupName.value = ''
    ElMessage.success('分组创建成功')
    await loadGroups()
  } catch (err) {
    ElMessage.error('创建分组失败: ' + err)
  }
}

/**
 * 删除分组
 */
const deleteGroup = async (group) => {
  try {
    await ElMessageBox.confirm(
      `删除分组 "${group.name}" 后，该分组下的命令将变为未分组，确定删除？`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteGroup(group.id)
    ElMessage.success('分组删除成功')
    await loadGroups()
    await loadCommands()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败: ' + err)
    }
  }
}

onMounted(() => {
  loadGroups()
  loadCommands()
})
</script>

<style scoped>
.shortcut-cmd-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #252526;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
}

.page-title {
  font-size: 16px;
  font-weight: 600;
  color: #e5e5e5;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.page-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.cmd-groups {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 6px;
  color: #c0c0c0;
  font-size: 14px;
  background-color: #2d2d2d;
}

.group-header:hover {
  background-color: #363636;
}

.group-name {
  font-weight: 500;
}

.group-count {
  color: #666;
  font-size: 12px;
}

.group-commands {
  padding-left: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 4px;
}

.cmd-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background-color: #2d2d2d;
  border: 1px solid #3d3d3d;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.cmd-card:hover {
  background-color: #363636;
  border-color: #409eff;
}

.cmd-main {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.cmd-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background-color: #1e1e1e;
  flex-shrink: 0;
}

.cmd-info {
  flex: 1;
  min-width: 0;
}

.cmd-name {
  font-size: 14px;
  font-weight: 500;
  color: #e5e5e5;
  margin-bottom: 4px;
}

.cmd-detail {
  display: flex;
  align-items: center;
  gap: 8px;
}

.shell-tag {
  flex-shrink: 0;
}

.cmd-commands {
  font-size: 12px;
  color: #888;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: Consolas, 'Courier New', monospace;
}

.cmd-workdir {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #666;
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cmd-actions {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.2s;
}

.cmd-card:hover .cmd-actions {
  opacity: 1;
}

.group-manage {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.group-add-row {
  display: flex;
  gap: 8px;
}

.group-add-row .el-input {
  flex: 1;
}

.group-list {
  max-height: 300px;
  overflow-y: auto;
}

.group-manage-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  color: #e5e5e5;
  font-size: 13px;
}

.group-manage-item:hover {
  background-color: #2d2d2d;
}

.action-icon {
  cursor: pointer;
  color: #a0a0a0;
  padding: 4px;
  border-radius: 4px;
}

.action-icon:hover {
  color: #f56c6c;
  background-color: #3d3d3d;
}
</style>
