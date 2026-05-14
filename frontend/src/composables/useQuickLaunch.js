import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { OpenDirectoryDialog } from '../../wailsjs/go/main/App'
import {
  GetGroups as GetSCGroups, CreateGroup as CreateSCGroup, DeleteGroup as DeleteSCGroup,
  GetCommands as GetSCCommands, CreateCommand as CreateSCCommand,
  UpdateCommand as UpdateSCCommand, DeleteCommand as DeleteSCCommand,
  ExecuteCommand as ExecSCCommand
} from '../../wailsjs/go/main/ShortcutCmdService'

export function useQuickLaunch(addQuickLaunchTab) {
  const qlGroups = ref([])
  const qlCmds = ref([])
  const expandedQlGroups = ref(new Set())

  const ungroupedQlCmds = computed(() => qlCmds.value.filter(cmd => !cmd.groupId))

  const getQlCmdsByGroup = (groupId) => qlCmds.value.filter(cmd => cmd.groupId === groupId)

  const getQlCmdCount = (groupId) => qlCmds.value.filter(cmd => cmd.groupId === groupId).length

  const toggleQlGroup = (groupId) => {
    const newSet = new Set(expandedQlGroups.value)
    if (newSet.has(groupId)) newSet.delete(groupId)
    else newSet.add(groupId)
    expandedQlGroups.value = newSet
  }

  const loadQlGroups = async () => {
    try {
      qlGroups.value = await GetSCGroups() || []
    } catch (err) {
      ElMessage.error('加载分组失败: ' + err)
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
          form.groupId,
          form.name,
          form.shell,
          form.workDir,
          form.commands,
          0
        )
        ElMessage.success('更新成功')
      } else {
        await CreateSCCommand(
          form.groupId,
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

  const addQlGroup = async (name) => {
    if (!name.trim()) { ElMessage.warning('请输入分组名称'); return }
    try {
      await CreateSCGroup(name.trim(), 0)
      ElMessage.success('分组创建成功')
      await loadQlGroups()
      return true
    } catch (err) {
      ElMessage.error('创建分组失败: ' + err)
      return false
    }
  }

  const deleteQlGroup = async (group) => {
    try {
      await ElMessageBox.confirm(
        `删除分组 "${group.name}" 后，该分组下的命令将变为未分组，确定删除？`,
        '确认删除',
        { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
      )
      await DeleteSCGroup(group.id)
      ElMessage.success('分组删除成功')
      await loadQlGroups()
      await loadQlCmds()
    } catch (err) {
      if (err !== 'cancel') {
        ElMessage.error('删除失败: ' + err)
      }
    }
  }

  return {
    qlGroups,
    qlCmds,
    expandedQlGroups,
    ungroupedQlCmds,
    getQlCmdsByGroup,
    getQlCmdCount,
    toggleQlGroup,
    loadQlGroups,
    loadQlCmds,
    executeQlCmd,
    selectWorkDir,
    saveQlCmd,
    deleteQlCmd,
    addQlGroup,
    deleteQlGroup
  }
}
