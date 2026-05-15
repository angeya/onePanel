<template>
  <div>
    <QlCmdDialog ref="qlCmdRef" />
    <QlGroupDialog ref="qlGroupRef" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import QlCmdDialog from './QlCmdDialog.vue'
import QlGroupDialog from './QlGroupDialog.vue'

const qlCmdRef = ref(null)
const qlGroupRef = ref(null)

const showQlAddDialog = () => {
  qlCmdRef.value.openAdd()
}

const editQlCmd = (cmd) => {
  qlCmdRef.value.openEdit(cmd)
}

const showQlGroupDialog = () => {
  qlGroupRef.value.open()
}

defineExpose({
  showQlAddDialog,
  editQlCmd,
  showQlGroupDialog
})
</script>
