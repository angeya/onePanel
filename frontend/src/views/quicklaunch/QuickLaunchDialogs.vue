<template>
  <div>
    <QlCmdDialog ref="qlCmdRef" />
    <QlCategoryDialog ref="qlCategoryRef" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import QlCmdDialog from './QlCmdDialog.vue'
import QlCategoryDialog from './QlCategoryDialog.vue'

const qlCmdRef = ref(null)
const qlCategoryRef = ref(null)

const showQlAddDialog = () => {
  qlCmdRef.value.openAdd()
}

const editQlCmd = (cmd) => {
  qlCmdRef.value.openEdit(cmd)
}

const showQlCategoryDialog = () => {
  qlCategoryRef.value.open()
}

defineExpose({
  showQlAddDialog,
  editQlCmd,
  showQlCategoryDialog
})
</script>
