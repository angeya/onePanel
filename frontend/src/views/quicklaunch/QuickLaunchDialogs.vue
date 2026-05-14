<template>
  <div>
    <QlCmdDialog
      :ql-cmd-dialog-visible="cmdState.visible"
      :is-editing-ql-cmd="cmdState.isEditing"
      :ql-cmd-form="cmdState.form"
      :ql-groups="qlService.qlGroups.value"
      @update:ql-cmd-dialog-visible="cmdState.visible = $event"
      @update-ql-cmd-form="handleUpdateQlCmdForm"
      @select-work-dir="handleSelectWorkDir"
      @save-ql-cmd="handleSaveQlCmd"
    />

    <QlGroupDialog
      :ql-group-dialog-visible="groupState.visible"
      :ql-groups="qlService.qlGroups.value"
      :new-group-name="groupState.newName"
      @update:ql-group-dialog-visible="groupState.visible = $event"
      @update:new-group-name="groupState.newName = $event"
      @add-ql-group="handleAddQlGroup"
      @delete-ql-group="handleDeleteQlGroup"
    />
  </div>
</template>

<script setup>
import { reactive, inject } from 'vue'
import QlCmdDialog from './QlCmdDialog.vue'
import QlGroupDialog from './QlGroupDialog.vue'

const qlService = inject('qlService')

const cmdState = reactive({
  visible: false,
  isEditing: false,
  editingId: null,
  form: { name: '', groupId: null, shell: 'cmd.exe', workDir: '', commands: '' }
})

const groupState = reactive({
  visible: false,
  newName: ''
})

const showQlAddDialog = () => {
  cmdState.isEditing = false
  cmdState.editingId = null
  cmdState.form = { name: '', groupId: null, shell: 'cmd.exe', workDir: '', commands: '' }
  cmdState.visible = true
}

const editQlCmd = (cmd) => {
  cmdState.isEditing = true
  cmdState.editingId = cmd.id
  cmdState.form = {
    name: cmd.name,
    groupId: cmd.groupId || null,
    shell: cmd.shell || 'cmd.exe',
    workDir: cmd.workDir || '',
    commands: cmd.commands
  }
  cmdState.visible = true
}

const showQlGroupDialog = () => {
  groupState.newName = ''
  groupState.visible = true
}

const handleUpdateQlCmdForm = ({ key, value }) => {
  cmdState.form[key] = value
}

const handleSelectWorkDir = async () => {
  const dir = await qlService.selectWorkDir()
  if (dir) cmdState.form.workDir = dir
}

const handleSaveQlCmd = async () => {
  const ok = await qlService.saveQlCmd(cmdState.isEditing, cmdState.editingId, cmdState.form)
  if (ok) cmdState.visible = false
}

const handleAddQlGroup = async () => {
  const ok = await qlService.addQlGroup(groupState.newName)
  if (ok) groupState.newName = ''
}

const handleDeleteQlGroup = (group) => {
  qlService.deleteQlGroup(group)
}

defineExpose({
  showQlAddDialog,
  editQlCmd,
  showQlGroupDialog
})
</script>
