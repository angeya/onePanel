import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { OpenDirectoryDialog } from '../../wailsjs/go/main/App'
import {
  GetGroups as GetSCGroups, CreateGroup as CreateSCGroup, DeleteGroup as DeleteSCGroup,
  GetCommands as GetSCCommands, CreateCommand as CreateSCCommand,
  UpdateCommand as UpdateSCCommand, DeleteCommand as DeleteSCCommand,
  ExecuteCommand as ExecSCCommand
} from '../../wailsjs/go/main/ShortcutCmdService'

/**
 * 快速启动组合式函数
 * 负责快速启动的分组、命令、执行等逻辑
 */
export function useQuickLaunch(addQuickLaunchTab) {
  const qlGroups = ref([])
  const qlCmds = ref([])
  const expandedQlGroups = ref(new Set(['none']))

  const qlCmdDialogVisible = ref(false)
  const isEditingQlCmd = ref(false)
  const editingQlCmdId = ref(null)
  const qlCmdForm = ref({
    name: '',
    groupId: null,
    shell: 'cmd.exe',
    workDir: '',
    commands: ''
  })

  const qlGroupDialogVisible = ref(false)
  const newGroupName = ref('')

  const ungroupedQlCmds = computed(() => qlCmds.value.filter(cmd => !cmd.groupId))

  /**
   * 按分组获取命令列表
   */
  const getQlCmdsByGroup = (groupId) => qlCmds.value.filter(cmd => cmd.groupId === groupId)

  /**
   * 获取分组下命令数量
   */
  const getQlCmdCount = (groupId) => qlCmds.value.filter(cmd => cmd.groupId === groupId).length

  /**
   * 展开/折叠分组
   */
  const toggleQlGroup = (groupId) => {
    const newSet = new Set(expandedQlGroups.value)
    if (newSet.has(groupId)) newSet.delete(groupId)
    else newSet.add(groupId)
    expandedQlGroups.value = newSet
  }

  /**
   * 加载分组数据
   */
  const loadQlGroups = async () => {
    try {
      qlGroups.value = await GetSCGroups() || []
    } catch (err) {
      ElMessage.error('加载分组失败: ' + err)
    }
  }

  /**
   * 加载命令数据
   */
  const loadQlCmds = async () => {
    try {
      qlCmds.value = await GetSCCommands() || []
    } catch (err) {
      ElMessage.error('加载命令失败: ' + err)
    }
  }

  /**
   * 执行快速启动命令
   */
  const executeQlCmd = async (cmd, quickLaunchTabRef) => {
    if (addQuickLaunchTab) addQuickLaunchTab()
    setTimeout(() => {
      if (quickLaunchTabRef && quickLaunchTabRef.value) {
        quickLaunchTabRef.value.execute(cmd)
      }
    }, 50)
  }

  /**
   * 显示新增命令对话框
   */
  const showQlAddDialog = () => {
    isEditingQlCmd.value = false
    editingQlCmdId.value = null
    qlCmdForm.value = { name: '', groupId: null, shell: 'cmd.exe', workDir: '', commands: '' }
    qlCmdDialogVisible.value = true
  }

  /**
   * 编辑命令
   */
  const editQlCmd = (cmd) => {
    isEditingQlCmd.value = true
    editingQlCmdId.value = cmd.id
    qlCmdForm.value = {
      name: cmd.name,
      groupId: cmd.groupId || null,
      shell: cmd.shell || 'cmd.exe',
      workDir: cmd.workDir || '',
      commands: cmd.commands
    }
    qlCmdDialogVisible.value = true
  }

  /**
   * 选择工作目录
   */
  const selectWorkDir = async () => {
    try {
      const dir = await OpenDirectoryDialog('选择工作目录')
      if (dir) qlCmdForm.value.workDir = dir
    } catch (err) {
      console.error('选择目录失败:', err)
      ElMessage.warning('目录选择对话框打开失败，请手动输入路径')
    }
  }

  /**
   * 保存命令
   */
  const saveQlCmd = async () => {
    if (!qlCmdForm.value.name) { ElMessage.warning('请输入命令名称'); return }
    if (!qlCmdForm.value.commands) { ElMessage.warning('请输入命令内容'); return }

    try {
      if (isEditingQlCmd.value) {
        await UpdateSCCommand(
          editingQlCmdId.value,
          qlCmdForm.value.groupId,
          qlCmdForm.value.name,
          qlCmdForm.value.shell,
          qlCmdForm.value.workDir,
          qlCmdForm.value.commands,
          0
        )
        ElMessage.success('更新成功')
      } else {
        await CreateSCCommand(
          qlCmdForm.value.groupId,
          qlCmdForm.value.name,
          qlCmdForm.value.shell,
          qlCmdForm.value.workDir,
          qlCmdForm.value.commands,
          0
        )
        ElMessage.success('创建成功')
      }
      qlCmdDialogVisible.value = false
      await loadQlCmds()
    } catch (err) {
      ElMessage.error('保存失败: ' + err)
    }
  }

  /**
   * 删除命令
   */
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

  /**
   * 显示分组管理对话框
   */
  const showQlGroupDialog = () => {
    newGroupName.value = ''
    qlGroupDialogVisible.value = true
  }

  /**
   * 添加分组
   */
  const addQlGroup = async () => {
    if (!newGroupName.value.trim()) { ElMessage.warning('请输入分组名称'); return }
    try {
      await CreateSCGroup(newGroupName.value.trim(), 0)
      newGroupName.value = ''
      ElMessage.success('分组创建成功')
      await loadQlGroups()
    } catch (err) {
      ElMessage.error('创建分组失败: ' + err)
    }
  }

  /**
   * 删除分组
   */
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
    qlCmdDialogVisible,
    isEditingQlCmd,
    qlCmdForm,
    qlGroupDialogVisible,
    newGroupName,
    ungroupedQlCmds,
    getQlCmdsByGroup,
    getQlCmdCount,
    toggleQlGroup,
    loadQlGroups,
    loadQlCmds,
    executeQlCmd,
    showQlAddDialog,
    editQlCmd,
    selectWorkDir,
    saveQlCmd,
    deleteQlCmd,
    showQlGroupDialog,
    addQlGroup,
    deleteQlGroup
  }
}
