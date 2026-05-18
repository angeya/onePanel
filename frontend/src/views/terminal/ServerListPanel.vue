<template>
  <div class="server-list-panel">
    <div class="panel-header">
      <el-button size="small" type="primary" @click="sessionDialogRef.show()" plain>
        <el-icon><Plus /></el-icon>
        新增会话
      </el-button>
      <el-button size="small" @click="categoryDialogRef.show()" plain>
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

    <ServerSessionDialog
      ref="sessionDialogRef"
      :categories="categories"
      @saved="onDialogSaved"
      @login="handleLogin"
    />
    <ServerRenameDialog
      ref="renameDialogRef"
      @saved="onDialogSaved"
    />
    <ServerCategoryDialog
      ref="categoryDialogRef"
      :categories="categories"
      @saved="onDialogSaved"
    />

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
  DeleteServer,
  GetLoginCommand,
  DeployKey,
  GetSessionCategories
} from '../../../wailsjs/go/main/ServerListService'
import ServerSessionDialog from './ServerSessionDialog.vue'
import ServerRenameDialog from './ServerRenameDialog.vue'
import ServerCategoryDialog from './ServerCategoryDialog.vue'

const emit = defineEmits(['executeCommand'])

const categories = ref([])
const servers = ref([])
const expandedCategories = ref(new Set())
const sessionDialogRef = ref(null)
const renameDialogRef = ref(null)
const categoryDialogRef = ref(null)

const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuTarget = ref(null)

const uncategorizedSessions = computed(() => {
  return servers.value.filter((s) => !s.categoryId)
})

const getSessionsByCategory = (categoryId) => {
  return servers.value.filter((s) => s.categoryId === categoryId)
}

const getSessionCount = (categoryId) => {
  return servers.value.filter((s) => s.categoryId === categoryId).length
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
    const result = await GetSessionCategories()
    categories.value = result || []
  } catch (err) {
    ElMessage.error('加载分类失败: ' + err)
  }
}

const loadServers = async () => {
  try {
    const result = await GetServers()
    servers.value = result || []
  } catch (err) {
    ElMessage.error('加载服务器列表失败: ' + err)
  }
}

const onDialogSaved = async () => {
  await Promise.all([loadCategories(), loadServers()])
}

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

const showContextMenu = (event, server) => {
  contextMenuTarget.value = server
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
}

const handleContextMenuAction = (action) => {
  contextMenuVisible.value = false
  const server = contextMenuTarget.value
  if (!server) return

  switch (action) {
    case 'rename':
      renameDialogRef.value.show(server)
      break
    case 'edit':
      sessionDialogRef.value.show(server)
      break
    case 'delete':
      handleDeleteServer(server)
      break
  }
}

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
.server-list-panel {
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
</style>
