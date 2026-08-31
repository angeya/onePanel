<template>
  <div class="json-tree-view">
    <!-- 左侧：JSON 树配置列表 -->
    <div class="tree-sidebar">
      <div class="sidebar-header">
        <span class="sidebar-title">JSON树</span>
        <el-button size="small" type="primary" text @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新增
        </el-button>
      </div>

      <div class="sidebar-search">
        <el-input
          v-model="listKeyword"
          placeholder="搜索树名称"
          size="small"
          clearable
          :prefix-icon="Search"
        />
      </div>

      <div class="tree-list">
        <div
          v-for="tree in filteredTrees"
          :key="tree.id"
          class="tree-item"
          :class="{ active: currentTree && currentTree.id === tree.id }"
          @click="selectTree(tree)"
        >
          <div class="tree-item-main">
            <span class="tree-item-name" :title="tree.name">{{ tree.name }}</span>
            <span class="tree-item-tags">
              <el-tag size="small" :type="tree.treeType === 'logic' ? 'warning' : 'success'">
                {{ tree.treeType === 'logic' ? '逻辑树' : '结构树' }}
              </el-tag>
              <el-tag size="small" type="info">
                {{ tree.sourceType === 'file' ? '文件' : '文本' }}
              </el-tag>
            </span>
          </div>
          <div class="tree-item-actions" @click.stop>
            <el-icon :size="14" title="编辑" @click="openEditDialog(tree)"><Edit /></el-icon>
            <el-icon :size="14" title="删除" class="danger" @click="removeTree(tree)"><Delete /></el-icon>
          </div>
        </div>
        <div v-if="filteredTrees.length === 0" class="tree-list-empty">暂无配置</div>
      </div>
    </div>

    <!-- 右侧：内容区 -->
    <div class="tree-content">
      <template v-if="currentTree">
        <div class="content-toolbar">
          <div class="toolbar-left">
            <span class="content-title">{{ currentTree.name }}</span>
            <el-tag size="small" type="info">共 {{ treeData.length }} 个根节点</el-tag>
          </div>
          <div class="toolbar-right">
            <el-input
              v-model="nodeKeyword"
              placeholder="搜索节点"
              size="small"
              clearable
              style="width: 180px"
              :prefix-icon="Search"
            />
            <el-button size="small" plain @click="expandAll">
              <el-icon><Fold /></el-icon>
              展开全部
            </el-button>
            <el-button size="small" plain @click="collapseAll">
              <el-icon><Expand /></el-icon>
              收起全部
            </el-button>
            <el-button size="small" plain @click="reloadCurrentTree">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>

        <div class="content-body">
          <div class="tree-panel">
            <el-tree
              ref="treeRef"
              :data="treeData"
              node-key="key"
              :props="{ label: 'label', children: 'children' }"
              :default-expand-all="defaultExpandAll"
              :expand-on-click-node="false"
              :filter-node-method="filterNode"
              highlight-current
              @node-click="handleNodeClick"
            >
              <template #default="{ data }">
                <span class="tree-node">
                  <el-icon :size="13" color="#e6a23c" v-if="data.hasChildren"><FolderOpened /></el-icon>
                  <el-icon :size="13" color="#909399" v-else><Document /></el-icon>
                  <span class="tree-node-label" :title="data.label">{{ data.label }}</span>
                  <span class="tree-node-meta">{{ data.meta }}</span>
                </span>
              </template>
            </el-tree>
            <el-empty v-if="treeData.length === 0" description="无节点数据" :image-size="60" />
          </div>

          <div class="detail-panel">
            <div class="detail-header">
              <span class="detail-title">节点 JSON</span>
              <span v-if="selectedPath" class="detail-path">{{ selectedPath }}</span>
              <el-button
                v-if="selectedNode"
                size="small"
                text
                @click="copyNodeJson"
              >
                <el-icon><CopyDocument /></el-icon>
                复制
              </el-button>
            </div>
            <div class="detail-body">
              <pre v-if="selectedNode" class="json-pre" v-html="nodeJsonHtml"></pre>
              <el-empty v-else description="点击左侧树节点查看完整 JSON" :image-size="60" />
            </div>
          </div>
        </div>
      </template>

      <div v-else class="content-empty">
        <el-empty description="选择左侧 JSON 树，或点击新增创建配置" :image-size="100" />
      </div>
    </div>

    <!-- 新增 / 编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="form.id ? '编辑 JSON 树' : '新增 JSON 树'"
      width="620px"
      :close-on-click-modal="false"
    >
      <el-form label-width="100px" size="default">
        <el-form-item label="树名称" required>
          <el-input v-model="form.name" placeholder="例如：菜单配置" maxlength="50" />
        </el-form-item>
        <el-form-item label="树结构" required>
          <el-radio-group v-model="form.treeType">
            <el-radio value="structure">结构树（嵌套子节点数组）</el-radio>
            <el-radio value="logic">逻辑树（平铺 + 父节点关联）</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="数据来源" required>
          <el-radio-group v-model="form.sourceType">
            <el-radio value="text">JSON 文本</el-radio>
            <el-radio value="file">JSON 文件</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="form.sourceType === 'text'" label="JSON 内容" required>
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="10"
            placeholder='请输入 JSON 文本，如 [{"id":1,"name":"根节点","children":[...]}]'
            class="json-textarea"
          />
        </el-form-item>
        <el-form-item v-if="form.sourceType === 'file'" label="文件路径" required>
          <div class="file-path-row">
            <el-input v-model="form.filePath" placeholder="选择 JSON 文件路径" readonly />
            <el-button @click="chooseFile">浏览</el-button>
          </div>
        </el-form-item>

        <template v-if="form.treeType === 'logic'">
          <el-form-item label="节点ID字段" required>
            <el-input v-model="form.idField" placeholder="例如：id" />
          </el-form-item>
          <el-form-item label="父节点字段" required>
            <el-input v-model="form.parentField" placeholder="例如：parentId" />
          </el-form-item>
        </template>
        <el-form-item v-if="form.treeType === 'structure'" label="子节点字段">
          <el-input v-model="form.childrenField" placeholder="默认：children" />
        </el-form-item>
        <el-form-item label="显示名称字段">
          <el-input v-model="form.labelField" placeholder="树节点展示名称取值字段，默认：name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveTree">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus, Search, Edit, Delete, Refresh, Fold, Expand,
  FolderOpened, Document, CopyDocument
} from '@element-plus/icons-vue'
import {
  GetJsonTrees, GetJsonTree, GetJsonTreeData, SaveJsonTree, DeleteJsonTree
} from '../../../wailsjs/go/main/JsonTreeService'
import { OpenFileDialog } from '../../../wailsjs/go/main/App'

const props = defineProps({
  embedded: { type: Boolean, default: false }
})

// ---------------- 左侧列表 ----------------
const trees = ref([])
const listKeyword = ref('')
const currentTree = ref(null)

const filteredTrees = computed(() => {
  if (!listKeyword.value) return trees.value
  const keyword = listKeyword.value.toLowerCase()
  return trees.value.filter(t => t.name.toLowerCase().includes(keyword))
})

const loadTrees = async () => {
  try {
    trees.value = await GetJsonTrees()
  } catch (err) {
    ElMessage.error('获取 JSON 树列表失败: ' + err)
  }
}

const selectTree = async (tree) => {
  try {
    currentTree.value = tree
    selectedNode.value = null
    selectedPath.value = ''
    await buildTreeData(tree)
  } catch (err) {
    treeData.value = []
    ElMessage.error('加载 JSON 数据失败: ' + err)
  }
}

const reloadCurrentTree = async () => {
  if (!currentTree.value) return
  await selectTree(currentTree.value)
}

// ---------------- 树构建 ----------------
const treeData = ref([])
const treeRef = ref(null)
const nodeKeyword = ref('')
const selectedNode = ref(null)
const selectedPath = ref('')
const defaultExpandAll = ref(true)

watch(nodeKeyword, (val) => {
  treeRef.value && treeRef.value.filter(val)
})

const filterNode = (value, data) => {
  if (!value) return true
  return data.label.toLowerCase().includes(value.toLowerCase())
}

/**
 * 从原始节点取展示名称
 */
const pickLabel = (item, labelField) => {
  const value = item[labelField]
  if (value !== undefined && value !== null && value !== '') return String(value)
  const fallback = item.name || item.title || item.label || item.id
  return fallback !== undefined && fallback !== null ? String(fallback) : '未命名节点'
}

/**
 * 构建结构树（嵌套子节点数组）
 */
const buildStructureTree = (items, childrenField, labelField) => {
  let counter = 0
  const convert = (item, parentPath) => {
    counter++
    const key = parentPath ? `${parentPath}/${counter}` : `${counter}`
    const children = Array.isArray(item[childrenField]) ? item[childrenField] : []
    const node = {
      key,
      label: pickLabel(item, labelField),
      raw: item,
      meta: children.length ? `${children.length} 个子节点` : '',
      hasChildren: children.length > 0
    }
    if (children.length) {
      node.children = children.map(child => convert(child, key))
    }
    return node
  }
  return items.map(item => convert(item, ''))
}

/**
 * 构建逻辑树（平铺数组 + 父节点标识关联）
 */
const buildLogicTree = (items, idField, parentField, labelField) => {
  const nodeMap = new Map()
  const idOf = (item) => {
    const v = item[idField]
    return v === undefined || v === null ? '' : String(v)
  }
  const parentOf = (item) => {
    const v = item[parentField]
    return v === undefined || v === null || v === '' ? '' : String(v)
  }

  items.forEach((item, index) => {
    if (!nodeMap.has(idOf(item)) && idOf(item) !== '') {
      nodeMap.set(idOf(item), { item, children: [] })
    } else if (idOf(item) === '') {
      nodeMap.set(`__index_${index}`, { item, children: [] })
    }
  })

  const roots = []
  items.forEach((item, index) => {
    const selfKey = idOf(item) !== '' ? idOf(item) : `__index_${index}`
    const entry = nodeMap.get(selfKey)
    const parentKey = parentOf(item)
    const parentEntry = parentKey !== '' ? nodeMap.get(parentKey) : null
    if (parentEntry && parentEntry !== entry) {
      parentEntry.children.push(entry)
    } else {
      roots.push(entry)
    }
  })

  let counter = 0
  const convert = (entry, parentPath) => {
    counter++
    const key = parentPath ? `${parentPath}/${counter}` : `${counter}`
    const node = {
      key,
      label: pickLabel(entry.item, labelField),
      raw: entry.item,
      meta: idOf(entry.item) !== '' ? `id: ${idOf(entry.item)}` : '',
      hasChildren: entry.children.length > 0
    }
    if (entry.children.length) {
      node.children = entry.children.map(child => convert(child, key))
    }
    return node
  }
  return roots.map(entry => convert(entry, ''))
}

const buildTreeData = async (tree) => {
  const content = await GetJsonTreeData(tree.id)
  let parsed
  try {
    parsed = JSON.parse(content)
  } catch (err) {
    throw new Error('JSON 解析失败: ' + err.message)
  }

  const labelField = tree.labelField || 'name'
  if (tree.treeType === 'logic') {
    if (!Array.isArray(parsed)) {
      throw new Error('逻辑树数据根节点必须是 JSON 数组')
    }
    treeData.value = buildLogicTree(parsed, tree.idField || 'id', tree.parentField || 'parentId', labelField)
  } else {
    if (!Array.isArray(parsed)) {
      throw new Error('结构树数据根节点必须是 JSON 数组')
    }
    treeData.value = buildStructureTree(parsed, tree.childrenField || 'children', labelField)
  }

  await nextTick()
  if (defaultExpandAll.value && treeRef.value) {
    // default-expand-all 仅首次生效，重新加载后手动展开
    expandAll()
  }
}

/**
 * 计算节点路径（从根到当前节点）
 */
const getNodePath = (data) => {
  const parts = []
  const walk = (nodes, chain) => {
    for (const n of nodes) {
      const nextChain = [...chain, n.label]
      if (n.key === data.key) {
        parts.push(...nextChain)
        return true
      }
      if (n.children && walk(n.children, nextChain)) return true
    }
    return false
  }
  walk(treeData.value, [])
  return parts.join(' / ')
}

const handleNodeClick = (data) => {
  selectedNode.value = data.raw
  selectedPath.value = getNodePath(data)
}

// ---------------- 节点 JSON 详情 ----------------
const nodeJsonText = computed(() => {
  if (!selectedNode.value) return ''
  try {
    return JSON.stringify(selectedNode.value, null, 2)
  } catch (err) {
    return String(selectedNode.value)
  }
})

const nodeJsonHtml = computed(() => highlightJson(nodeJsonText.value))

/**
 * JSON 简易语法高亮
 */
const highlightJson = (json) => {
  if (!json) return ''
  return json
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (match) => {
      let cls = 'json-number'
      if (/^"/.test(match)) {
        cls = /:$/.test(match) ? 'json-key' : 'json-string'
      } else if (/true|false/.test(match)) {
        cls = 'json-boolean'
      } else if (/null/.test(match)) {
        cls = 'json-null'
      }
      return `<span class="${cls}">${match}</span>`
    })
}

const copyNodeJson = async () => {
  try {
    await navigator.clipboard.writeText(nodeJsonText.value)
    ElMessage.success('已复制节点 JSON')
  } catch (err) {
    ElMessage.error('复制失败: ' + err)
  }
}

// ---------------- 展开收起 ----------------
const collectKeys = (nodes, keys = []) => {
  nodes.forEach(n => {
    keys.push(n.key)
    if (n.children) collectKeys(n.children, keys)
  })
  return keys
}

const expandAll = () => {
  const keys = collectKeys(treeData.value)
  keys.forEach(key => {
    const node = treeRef.value && treeRef.value.getNode(key)
    if (node) node.expanded = true
  })
}

const collapseAll = () => {
  const keys = collectKeys(treeData.value)
  keys.forEach(key => {
    const node = treeRef.value && treeRef.value.getNode(key)
    if (node) node.expanded = false
  })
  // 收起时保留根节点展开
  treeData.value.forEach(root => {
    const node = treeRef.value && treeRef.value.getNode(root.key)
    if (node) node.expanded = true
  })
}

// ---------------- 新增 / 编辑对话框 ----------------
const dialogVisible = ref(false)
const saving = ref(false)
const form = ref(createEmptyForm())

function createEmptyForm() {
  return {
    id: 0,
    name: '',
    treeType: 'structure',
    sourceType: 'text',
    content: '',
    filePath: '',
    idField: 'id',
    parentField: 'parentId',
    childrenField: 'children',
    labelField: 'name'
  }
}

const openCreateDialog = () => {
  form.value = createEmptyForm()
  dialogVisible.value = true
}

const openEditDialog = async (tree) => {
  try {
    const detail = await GetJsonTree(tree.id)
    form.value = {
      id: detail.id,
      name: detail.name,
      treeType: detail.treeType || 'structure',
      sourceType: detail.sourceType || 'text',
      content: detail.content || '',
      filePath: detail.filePath || '',
      idField: detail.idField || 'id',
      parentField: detail.parentField || 'parentId',
      childrenField: detail.childrenField || 'children',
      labelField: detail.labelField || 'name'
    }
    dialogVisible.value = true
  } catch (err) {
    ElMessage.error('获取配置详情失败: ' + err)
  }
}

const chooseFile = async () => {
  try {
    const path = await OpenFileDialog('选择 JSON 文件', 'JSON 文件 (*.json;*.txt)|*.json;*.txt|所有文件|*.*')
    if (path) {
      form.value.filePath = path
    }
  } catch (err) {
    ElMessage.error('选择文件失败: ' + err)
  }
}

const validateForm = () => {
  if (!form.value.name.trim()) {
    ElMessage.warning('请输入树名称')
    return false
  }
  if (form.value.sourceType === 'text') {
    if (!form.value.content.trim()) {
      ElMessage.warning('请输入 JSON 内容')
      return false
    }
    try {
      JSON.parse(form.value.content)
    } catch (err) {
      ElMessage.warning('JSON 内容格式不合法: ' + err.message)
      return false
    }
  } else if (!form.value.filePath.trim()) {
    ElMessage.warning('请选择 JSON 文件')
    return false
  }
  if (form.value.treeType === 'logic' && !form.value.idField.trim()) {
    ElMessage.warning('请输入节点 ID 字段')
    return false
  }
  if (form.value.treeType === 'logic' && !form.value.parentField.trim()) {
    ElMessage.warning('请输入父节点字段')
    return false
  }
  return true
}

const saveTree = async () => {
  if (!validateForm()) return
  saving.value = true
  try {
    await SaveJsonTree({
      id: form.value.id,
      name: form.value.name.trim(),
      sourceType: form.value.sourceType,
      content: form.value.sourceType === 'text' ? form.value.content : '',
      filePath: form.value.sourceType === 'file' ? form.value.filePath.trim() : '',
      treeType: form.value.treeType,
      idField: form.value.idField.trim() || 'id',
      parentField: form.value.parentField.trim() || 'parentId',
      childrenField: form.value.childrenField.trim() || 'children',
      labelField: form.value.labelField.trim() || 'name'
    })
    ElMessage.success(form.value.id ? '保存成功' : '新增成功')
    dialogVisible.value = false
    await loadTrees()
  } catch (err) {
    ElMessage.error('保存失败: ' + err)
  } finally {
    saving.value = false
  }
}

const removeTree = async (tree) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除 JSON 树 "${tree.name}" 吗？`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
    )
    await DeleteJsonTree(tree.id)
    ElMessage.success('删除成功')
    if (currentTree.value && currentTree.value.id === tree.id) {
      currentTree.value = null
      treeData.value = []
      selectedNode.value = null
      selectedPath.value = ''
    }
    await loadTrees()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败: ' + err)
    }
  }
}

loadTrees()
</script>

<style scoped>
.json-tree-view {
  display: flex;
  height: 100%;
  background-color: var(--bg-secondary);
  overflow: hidden;
}

/* 左侧列表 */
.tree-sidebar {
  width: 250px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border-light);
  background-color: var(--bg-primary);
  overflow: hidden;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 12px 8px;
}

.sidebar-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.sidebar-search {
  padding: 0 12px 8px;
}

.tree-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 8px 8px;
}

.tree-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  border-left: 2px solid transparent;
  transition: background-color 0.15s, border-color 0.15s;
}

.tree-item:hover {
  background-color: var(--bg-hover);
}

.tree-item.active {
  background-color: var(--bg-hover);
  border-left-color: var(--accent);
}

.tree-item.active .tree-item-name {
  color: var(--accent);
  font-weight: 600;
}

.tree-item-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.tree-item-name {
  font-size: 13px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tree-item-tags {
  display: flex;
  gap: 4px;
}

.tree-item-actions {
  display: flex;
  gap: 6px;
  color: var(--text-dimmed);
  opacity: 0;
  transition: opacity 0.15s;
}

.tree-item:hover .tree-item-actions {
  opacity: 1;
}

.tree-item-actions .el-icon {
  cursor: pointer;
}

.tree-item-actions .el-icon:hover {
  color: var(--accent);
}

.tree-item-actions .el-icon.danger:hover {
  color: var(--el-color-danger);
}

.tree-list-empty {
  text-align: center;
  color: var(--text-faint);
  font-size: 12px;
  padding: 24px 0;
}

/* 右侧内容区 */
.tree-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.content-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.content-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.content-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* 树面板 */
.tree-panel {
  flex: 1.2;
  min-width: 0;
  overflow: auto;
  padding: 8px;
  border-right: 1px solid var(--border-light);
}

.tree-panel :deep(.el-tree) {
  --el-tree-node-content-height: 28px;
  --el-tree-node-hover-bg-color: var(--bg-hover);
  background-color: transparent;
  color: var(--text-primary);
  font-size: 13px;
}

.tree-panel :deep(.el-tree-node__content) {
  border-radius: 6px;
  transition: background-color 0.15s, color 0.15s;
}

.tree-panel :deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: var(--bg-active);
  color: var(--accent);
}

.tree-panel :deep(.el-tree-node.is-current > .el-tree-node__content .tree-node-label) {
  color: var(--accent);
  font-weight: 600;
}

.tree-panel :deep(.el-tree-node.is-current > .el-tree-node__content .tree-node-meta) {
  color: var(--accent-hover);
}

.tree-panel :deep(.el-tree-node__expand-icon) {
  color: var(--text-muted);
}

.tree-panel :deep(.el-tree-node__expand-icon:hover) {
  color: var(--accent);
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  padding-right: 8px;
}

.tree-node-label {
  font-size: 13px;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tree-node-meta {
  font-size: 11px;
  color: var(--text-faint);
  white-space: nowrap;
}

/* JSON 详情面板 */
.detail-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background-color: var(--bg-primary);
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-light);
  flex-shrink: 0;
}

.detail-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  flex-shrink: 0;
}

.detail-path {
  font-size: 12px;
  color: var(--text-faint);
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.detail-body {
  flex: 1;
  overflow: auto;
  padding: 12px 14px;
}

.json-pre {
  margin: 0;
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-all;
}

.json-pre :deep(.json-key) {
  color: var(--json-key);
}

.json-pre :deep(.json-string) {
  color: var(--json-string);
}

.json-pre :deep(.json-number) {
  color: var(--json-number);
}

.json-pre :deep(.json-boolean) {
  color: var(--json-boolean);
}

.json-pre :deep(.json-null) {
  color: var(--json-null);
}

/* 对话框 */
.file-path-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

.json-textarea :deep(.el-textarea__inner) {
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
}
</style>
