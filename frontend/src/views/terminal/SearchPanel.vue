<template>
  <div class="search-panel">
    <el-input
      v-model="searchKeyword"
      size="small"
      placeholder="搜索终端内容..."
      @input="handleSearch"
      clearable
    />
    <div class="search-actions">
      <el-button size="small" @click="findPrevious" :disabled="!searchKeyword" plain>上一个</el-button>
      <el-button size="small" @click="findNext" :disabled="!searchKeyword" plain>下一个</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, onUnmounted } from 'vue'

const props = defineProps({
  activeTabId: { type: String, default: '' }
})

const searchKeyword = ref('')

/**
 * 发送搜索事件到当前激活的终端
 */
const dispatchSearch = (action) => {
  const event = new CustomEvent('terminal-search', {
    detail: {
      tabId: props.activeTabId,
      action,
      keyword: searchKeyword.value
    }
  })
  window.dispatchEvent(event)
}

/**
 * 执行搜索
 */
const handleSearch = () => {
  dispatchSearch('findNext')
}

/**
 * 查找上一个匹配项
 */
const findPrevious = () => {
  dispatchSearch('findPrevious')
}

/**
 * 查找下一个匹配项
 */
const findNext = () => {
  dispatchSearch('findNext')
}

onUnmounted(() => {
  dispatchSearch('clearDecorations')
})
</script>

<style scoped>
.search-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
}

.search-actions {
  display: flex;
  gap: 8px;
}
</style>
