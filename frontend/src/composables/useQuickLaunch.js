import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { OpenDirectoryDialog } from '../../wailsjs/go/main/App'
import {
  GetCategories as GetSCCategories, CreateCategory as CreateSCCategory, DeleteCategory as DeleteSCCategory,
  GetCommands as GetSCCommands, CreateCommand as CreateSCCommand,
  UpdateCommand as UpdateSCCommand, DeleteCommand as DeleteSCCommand,
  ExecuteCommand as ExecSCCommand
} from '../../wailsjs/go/main/ShortcutCmdService'

export function useQuickLaunch(addQuickLaunchTab) {
  const qlCategories = ref([])
  const qlCmds = ref([])
  const expandedQlCategories = ref(new Set())

  const uncategorizedQlCmds = computed(() => qlCmds.value.filter(cmd => !cmd.categoryId))

  const getQlCmdsByCategory = (categoryId) => qlCmds.value.filter(cmd => cmd.categoryId === categoryId)

  const getQlCmdCount = (categoryId) => qlCmds.value.filter(cmd => cmd.categoryId === categoryId).length

  const toggleQlCategory = (categoryId) => {
    const newSet = new Set(expandedQlCategories.value)
    if (newSet.has(categoryId)) newSet.delete(categoryId)
    else newSet.add(categoryId)
    expandedQlCategories.value = newSet
  }

  const loadQlCategories = async () => {
    try {
      qlCategories.value = await GetSCCategories() || []
    } catch (err) {
      ElMessage.error('加载分类失败: ' + err)
    }
  }

  const loadQlCmds = async () => {
    try {
      qlCmds.value = await GetSCCommands() || []
    } catch (err) {
      ElMessage.error('加载命令失败: ' + err)
    }
  }

  const executeQlCmd = async (cmd, quickLaunchTabRef) => {
    if (addQuickLaunchTab) addQuickLaunchTab()
    setTimeout(() => {
      if (quickLaunchTabRef && quickLaunchTabRef.value) {
        quickLaunchTabRef.value.execute(cmd)
      }
    }, 50)
  }

  const selectWorkDir = async () => {
    try {
      const dir = await OpenDirectoryDialog('选择工作目录')
      if (dir) return dir
    } catch (err) {
      console.error('选择目录失败:', err)
      ElMessage.warning('目录选择对话框打开失败，请手动输入路径')
    }
    return null
  }

  const saveQlCmd = async (isEditing, editingId, form) => {
    if (!form.name) { ElMessage.warning('请输入命令名称'); return }
    if (!form.commands) { ElMessage.warning('请输入命令内容'); return }

    try {
      if (isEditing) {
        await UpdateSCCommand(
          editingId,
          form.categoryId,
          form.name,
          form.shell,
          form.workDir,
          form.commands,
          0
        )
        ElMessage.success('更新成功')
      } else {
        await CreateSCCommand(
          form.categoryId,
          form.name,
          form.shell,
          form.workDir,
          form.commands,
          0
        )
        ElMessage.success('创建成功')
      }
      await loadQlCmds()
      return true
    } catch (err) {
      ElMessage.error('保存失败: ' + err)
      return false
    }
  }

  const deleteQlCmd = async (cmd) => {
    try {
      await ElMessageBox.confirm(
        `确定删除快速启动命令 "${cmd.name}" 吗？`,
        '确认删除',
        { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
      )
      await DeleteSCCommand(cmd.id)
      ElMessage.success('删除成功')
      await loadQlCmds()
    } catch (err) {
      if (err !== 'cancel') {
        ElMessage.error('删除失败: ' + err)
      }
    }
  }

  const addQlCategory = async (name) => {
    if (!name.trim()) { ElMessage.warning('请输入分类名称'); return }
    try {
      await CreateSCCategory(name.trim(), 0)
      ElMessage.success('分类创建成功')
      await loadQlCategories()
      return true
    } catch (err) {
      ElMessage.error('创建分类失败: ' + err)
      return false
    }
  }

  const deleteQlCategory = async (category) => {
    try {
      await ElMessageBox.confirm(
        `删除分类 "${category.name}" 后，该分类下的命令将变为未分类，确定删除？`,
        '确认删除',
        { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
      )
      await DeleteSCCategory(category.id)
      ElMessage.success('分类删除成功')
      await loadQlCategories()
      await loadQlCmds()
    } catch (err) {
      if (err !== 'cancel') {
        ElMessage.error('删除失败: ' + err)
      }
    }
  }

  return {
    qlCategories,
    qlCmds,
    expandedQlCategories,
    uncategorizedQlCmds,
    getQlCmdsByCategory,
    getQlCmdCount,
    toggleQlCategory,
    loadQlCategories,
    loadQlCmds,
    executeQlCmd,
    selectWorkDir,
    saveQlCmd,
    deleteQlCmd,
    addQlCategory,
    deleteQlCategory
  }
}
