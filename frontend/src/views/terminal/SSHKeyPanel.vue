<template>
  <div class="ssh-panel">
    <div class="panel-header">
      <el-button size="small" type="primary" @click="showAddDialog" plain>
        <el-icon><Plus /></el-icon>
        新增会话
      </el-button>
      <el-button size="small" @click="showCategoryDialog" plain>
        <el-icon><FolderAdd /></el-icon>
        新增分类
      </el-button>
    </div>

    <div class="session-list">
      <div v-for="category in categories" :key="category.id" class="category-group">
        <div class="category-header" @click="toggleCategory(category.id)">
          <el-icon>
            <ArrowDown v-if="expandedCategories.has(category.id)" />
            <ArrowRight v-else />
          </el-icon>
          <span class="category-name">{{ category.name }}</span>
          <span class="category-count">({{ getSessionCount(category.id) }})</span>
        </div>
        <div v-show="expandedCategories.has(category.id)" class="category-sessions">
          <div
            v-for="server in getSessionsByCategory(category.id)"
            :key="server.id"
            class="session-item"
            @dblclick="handleLogin(server)"
            @contextmenu.prevent="showContextMenu($event, server)"
          >
            <div class="session-info">
              <span class="session-name">{{ server.sessionName }}</span>
              <span class="session-detail">{{ server.user }}@{{ server.host }}{{ server.port !== 22 ? ':' + server.port : '' }}</span>
            </div>
            <div class="session-badges">
              <el-tag v-if="server.keyDeployed" type="success" size="small" effect="dark">密钥</el-tag>
              <el-tag v-else-if="server.useKeyLogin" type="warning" size="small" effect="dark">待部署</el-tag>
            </div>
            <el-icon class="more-icon" @click.stop="showContextMenu($event, server)"><MoreFilled /></el-icon>
          </div>
        </div>
      </div>

      <div
        v-for="server in uncategorizedSessions"
        :key="server.id"
        class="session-item"
        @dblclick="handleLogin(server)"
        @contextmenu.prevent="showContextMenu($event, server)"
      >
        <div class="session-info">
          <span class="session-name">{{ server.sessionName }}</span>
          <span class="session-detail">{{ server.user }}@{{ server.host }}{{ server.port !== 22 ? ':' + server.port : '' }}</span>
        </div>
        <div class="session-badges">
          <el-tag v-if="server.keyDeployed" type="success" size="small" effect="dark">密钥</el-tag>
          <el-tag v-else-if="server.useKeyLogin" type="warning" size="small" effect="dark">待部署</el-tag>
        </div>
        <el-icon class="more-icon" @click.stop="showContextMenu($event, server)"><MoreFilled /></el-icon>
      </div>

      <el-empty v-if="servers.length === 0" description="暂无会话" :image-size="50" />
    </div>

    <el-dialog
      v-model="sessionDialogVisible"
      :title="isEditing ? '编辑会话' : '新增会话'"
      width="420px"
      :close-on-click-modal="false"
    >
      <el-form :model="sessionForm" label-width="100px" size="default">
        <el-form-item label="会话名称">
          <el-input v-model="sessionForm.sessionName" placeholder="可选，默认使用 用户名@主机名" />
        </el-form-item>
        <el-form-item label="所属分类">
          <el-select v-model="sessionForm.categoryId" placeholder="请选择分类" clearable style="width: 100%">
            <el-option
              v-for="cat in categories"
              :key="cat.id"
              :label="cat.name"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="主机地址">
          <el-input v-model="sessionForm.host" placeholder="如 192.168.1.100" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="sessionForm.port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="sessionForm.user" placeholder="如 root" />
        </el-form-item>
        <el-form-item label="密钥登录">
          <el-checkbox v-model="sessionForm.useKeyLogin">
            以后使用密钥免密登录
          </el-checkbox>
          <div v-if="sessionForm.useKeyLogin" class="key-hint">
            首次登录需输入密码，登录成功后自动部署公钥
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="sessionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveSession" :loading="savingSession">
          {{ isEditing ? '保存' : '确认并登录' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="renameDialogVisible"
      title="重命名会话"
      width="360px"
      :close-on-click-modal="false"
    >
      <el-input v-model="renameValue" placeholder="请输入新的会话名称" @keyup.enter="handleRename" />
      <template #footer>
        <el-button @click="renameDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRename">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="categoryDialogVisible"
      title="管理分类"
      width="380px"
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

    <div
      v-if="contextMenuVisible"
      class="context-menu"
      :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
    >
      <div class="context-menu-item" @click="handleContextMenuAction('rename')">
        <el-icon><Edit /></el-icon>
        <span>重命名</span>
      </div>
      <div class="context-menu-item" @click="handleContextMenuAction('edit')">
        <el-icon><Setting /></el-icon>
        <span>编辑</span>
      </div>
      <div class="context-menu-item danger" @click="handleContextMenuAction('delete')">
        <el-icon><Delete /></el-icon>
        <span>删除</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, onBeforeUnmount } from 'vue'
import {
  Plus, FolderAdd, Edit, Delete, ArrowDown, ArrowRight, MoreFilled, Setting
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  GetServers,
  AddServer,
  UpdateServer,
  DeleteServer,
  RenameServer,
  GetLoginCommand,
  DeployKey,
  GetSessionCategories,
  CreateSessionCategory,
  DeleteSessionCategory
} from '../../../wailsjs/go/main/SSHKeyService'

const emit = defineEmits(['executeCommand'])

const categories = ref([])
const servers = ref([])
const expandedCategories = ref(new Set())

const sessionDialogVisible = ref(false)
const isEditing = ref(false)
const editingId = ref(null)
const savingSession = ref(false)
const sessionForm = ref({
  sessionName: '',
  categoryId: null,
  host: '',
  port: 22,
  user: '',
  useKeyLogin: true
})

const renameDialogVisible = ref(false)
const renameValue = ref('')
const renameTarget = ref(null)

const categoryDialogVisible = ref(false)
const newCategoryName = ref('')

const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuTarget = ref(null)

/**
 * 获取未分类的会话列表
 */
const uncategorizedSessions = computed(() => {
  return servers.value.filter((s) => !s.categoryId)
})

/**
 * 按分类 ID 获取会话列表
 */
const getSessionsByCategory = (categoryId) => {
  return servers.value.filter((s) => s.categoryId === categoryId)
}

/**
 * 获取分类下会话数量
 */
const getSessionCount = (categoryId) => {
  return servers.value.filter((s) => s.categoryId === categoryId).length
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
    const result = await GetSessionCategories()
    categories.value = result || []
  } catch (err) {
    ElMessage.error('加载分类失败: ' + err)
  }
}

/**
 * 加载服务器列表
 */
const loadServers = async () => {
  try {
    const result = await GetServers()
    servers.value = result || []
  } catch (err) {
    ElMessage.error('加载服务器列表失败: ' + err)
  }
}

/**
 * 显示新增会话对话框
 */
const showAddDialog = () => {
  isEditing.value = false
  editingId.value = null
  sessionForm.value = {
    sessionName: '',
    categoryId: null,
    host: '',
    port: 22,
    user: '',
    useKeyLogin: true
  }
  sessionDialogVisible.value = true
}

/**
 * 显示编辑会话对话框
 */
const showEditDialog = (server) => {
  isEditing.value = true
  editingId.value = server.id
  sessionForm.value = {
    sessionName: server.sessionName,
    categoryId: server.categoryId,
    host: server.host,
    port: server.port,
    user: server.user,
    useKeyLogin: server.useKeyLogin
  }
  sessionDialogVisible.value = true
}

/**
 * 保存会话（新增或编辑）
 * 新增时确认后直接发起 SSH 登录
 */
const handleSaveSession = async () => {
  if (!sessionForm.value.host) {
    ElMessage.warning('请输入主机地址')
    return
  }
  if (!sessionForm.value.user) {
    ElMessage.warning('请输入用户名')
    return
  }

  savingSession.value = true
  try {
    if (isEditing.value) {
      await UpdateServer(
        editingId.value,
        sessionForm.value.categoryId,
        sessionForm.value.sessionName,
        sessionForm.value.host,
        sessionForm.value.port,
        sessionForm.value.user,
        sessionForm.value.useKeyLogin
      )
      ElMessage.success('更新成功')
      sessionDialogVisible.value = false
      await loadServers()
    } else {
      const server = await AddServer(
        sessionForm.value.categoryId,
        sessionForm.value.sessionName,
        sessionForm.value.host,
        sessionForm.value.port,
        sessionForm.value.user,
        sessionForm.value.useKeyLogin
      )
      sessionDialogVisible.value = false
      await loadServers()
      handleLogin(server)
    }
  } catch (err) {
    ElMessage.error('保存失败: ' + err)
  } finally {
    savingSession.value = false
  }
}

/**
 * 双击登录服务器
 * 向终端发送 SSH 登录命令
 * 如果勾选了密钥登录且密钥未部署，登录后自动部署公钥
 */
const handleLogin = async (server) => {
  try {
    const cmd = await GetLoginCommand(server.id)
    emit('executeCommand', cmd + '\r')

    if (server.useKeyLogin && !server.keyDeployed) {
      promptDeployKey(server)
    }
  } catch (err) {
    ElMessage.error('获取登录命令失败: ' + err)
  }
}

/**
 * 提示用户输入密码以部署公钥
 * 在终端中用户已输入密码登录后，再通过后端部署公钥
 */
const promptDeployKey = async (server) => {
  try {
    await ElMessageBox.prompt(
      `已向终端发送 SSH 登录命令。如需自动部署公钥以实现免密登录，请输入服务器密码：`,
      '部署公钥 - ' + server.sessionName,
      {
        confirmButtonText: '部署',
        cancelButtonText: '跳过',
        inputType: 'password',
        inputPlaceholder: '输入服务器登录密码'
      }
    ).then(async ({ value }) => {
      if (value) {
        try {
          await DeployKey(server.id, value)
          ElMessage.success('公钥部署成功，后续可免密登录')
          await loadServers()
        } catch (err) {
          ElMessage.error('部署失败: ' + err)
        }
      }
    }).catch(() => {})
  } catch {
    // 用户跳过
  }
}

/**
 * 显示右键菜单
 */
const showContextMenu = (event, server) => {
  contextMenuTarget.value = server
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
}

/**
 * 处理右键菜单操作
 */
const handleContextMenuAction = (action) => {
  contextMenuVisible.value = false
  const server = contextMenuTarget.value
  if (!server) return

  switch (action) {
    case 'rename':
      renameValue.value = server.sessionName
      renameTarget.value = server
      renameDialogVisible.value = true
      break
    case 'edit':
      showEditDialog(server)
      break
    case 'delete':
      handleDeleteServer(server)
      break
  }
}

/**
 * 重命名会话
 */
const handleRename = async () => {
  if (!renameValue.value.trim()) {
    ElMessage.warning('会话名称不能为空')
    return
  }

  try {
    await RenameServer(renameTarget.value.id, renameValue.value.trim())
    ElMessage.success('重命名成功')
    renameDialogVisible.value = false
    await loadServers()
  } catch (err) {
    ElMessage.error('重命名失败: ' + err)
  }
}

/**
 * 删除会话
 */
const handleDeleteServer = async (server) => {
  try {
    await ElMessageBox.confirm(
      `确定删除会话 "${server.sessionName}"？`,
      '提示',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteServer(server.id)
    ElMessage.success('删除成功')
    await loadServers()
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
    await CreateSessionCategory(newCategoryName.value.trim(), 0)
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
    await ElMessageBox.confirm(
      '删除分类后，该分类下的会话将变为未分类，确定删除？',
      '提示',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteSessionCategory(id)
    ElMessage.success('分类删除成功')
    await loadCategories()
    await loadServers()
  } catch {
    // 用户取消
  }
}

/**
 * 点击其他区域关闭右键菜单
 */
const handleClickOutside = () => {
  contextMenuVisible.value = false
}

onMounted(() => {
  loadCategories()
  loadServers()
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.ssh-panel {
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

.session-list {
  flex: 1;
  overflow-y: auto;
}

.session-list::-webkit-scrollbar {
  width: 4px;
}

.session-list::-webkit-scrollbar-thumb {
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

.category-sessions {
  padding-left: 12px;
}

.session-item {
  display: flex;
  align-items: center;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.15s;
  gap: 6px;
}

.session-item:hover {
  background-color: var(--bg-hover);
}

.session-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex: 1;
}

.session-name {
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-detail {
  font-size: 11px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.session-badges {
  flex-shrink: 0;
}

.more-icon {
  flex-shrink: 0;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  border-radius: 4px;
  opacity: 0;
  transition: opacity 0.15s;
}

.session-item:hover .more-icon {
  opacity: 1;
}

.more-icon:hover {
  color: var(--text-primary);
  background-color: var(--bg-active);
}

.key-hint {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
  line-height: 1.4;
}

.context-menu {
  position: fixed;
  z-index: 9999;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 4px 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 120px;
}

.context-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-primary);
  transition: background-color 0.15s;
}

.context-menu-item:hover {
  background-color: var(--bg-hover);
}

.context-menu-item.danger {
  color: var(--el-color-danger);
}

.context-menu-item.danger:hover {
  background-color: var(--el-color-danger-light-9);
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
