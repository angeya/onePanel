<template>
  <Transition name="search-bar">
    <div v-if="visible" class="search-bar" @keydown.stop>
      <div class="search-input-wrapper">
        <input
          ref="inputRef"
          v-model="keyword"
          class="search-input"
          placeholder="搜索..."
          @input="onInput"
          @keydown.enter="onEnter"
          @keydown.shift.enter="onShiftEnter"
          @keydown.esc="close"
        />
        <span v-if="matchInfo" class="match-info" :class="{ 'unsupported': isUnsupported }">{{ matchInfo }}</span>
      </div>
      <button class="search-btn" title="上一个 (Shift+Enter)" @click="findPrev" :disabled="!keyword">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="currentColor"><path d="M7 4.5L2.5 9l1 1L7 6.5 10.5 10l1-1z"/></svg>
      </button>
      <button class="search-btn" title="下一个 (Enter)" @click="findNext" :disabled="!keyword">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="currentColor"><path d="M7 9.5L2.5 5l1-1L7 7.5 10.5 4l1 1z"/></svg>
      </button>
      <button class="search-close" title="关闭 (Esc)" @click="close">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="currentColor"><path d="M7 5.59L2.41 1 1 2.41 5.59 7 1 11.59 2.41 13 7 8.41 11.59 13 13 11.59 8.41 7 13 2.41 11.59 1z"/></svg>
      </button>
    </div>
  </Transition>
</template>

<script setup>
import { ref, nextTick, watch } from 'vue'

const props = defineProps({
  visible: { type: Boolean, default: false }
})

const emit = defineEmits(['update:visible', 'search', 'findNext', 'findPrev', 'close'])

const inputRef = ref(null)
const keyword = ref('')
const matchInfo = ref('')
const isUnsupported = ref(false)

watch(() => props.visible, (val) => {
  if (val) {
    nextTick(() => {
      inputRef.value?.focus()
      inputRef.value?.select()
    })
  } else {
    keyword.value = ''
    matchInfo.value = ''
    isUnsupported.value = false
  }
})

const onInput = () => {
  emit('search', keyword.value)
}

const onEnter = () => {
  emit('findNext', keyword.value)
}

const onShiftEnter = () => {
  emit('findPrev', keyword.value)
}

const findNext = () => {
  if (keyword.value) {
    emit('findNext', keyword.value)
  }
}

const findPrev = () => {
  if (keyword.value) {
    emit('findPrev', keyword.value)
  }
}

const close = () => {
  emit('close')
  emit('update:visible', false)
}

/**
 * 更新匹配信息显示
 * @param {number} current - 当前匹配索引（从1开始）
 * @param {number} total - 总匹配数
 */
const updateMatchInfo = (current, total) => {
  isUnsupported.value = false
  if (total === 0) {
    matchInfo.value = '无结果'
  } else {
    matchInfo.value = `${current}/${total}`
  }
}

/**
 * 清除匹配信息
 */
const clearMatchInfo = () => {
  matchInfo.value = ''
  isUnsupported.value = false
}

/**
 * 设置为不支持搜索状态
 */
const setUnsupported = () => {
  isUnsupported.value = true
  matchInfo.value = '不支持搜索'
}

defineExpose({ updateMatchInfo, clearMatchInfo, setUnsupported, focus: () => inputRef.value?.focus() })
</script>

<style scoped>
.search-bar {
  position: absolute;
  top: 8px;
  right: 16px;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 2px;
  background-color: var(--bg-tertiary);
  border: 1px solid var(--border-light);
  border-radius: 6px;
  padding: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 6px;
  min-width: 200px;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 13px;
  font-family: inherit;
  min-width: 120px;
  padding: 4px 0;
}

.search-input::placeholder {
  color: var(--text-faint);
}

.match-info {
  font-size: 11px;
  color: var(--text-muted);
  white-space: nowrap;
  flex-shrink: 0;
  user-select: none;
}

.match-info.unsupported {
  color: var(--text-faint);
  font-style: italic;
}

.search-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 4px;
  flex-shrink: 0;
  padding: 0;
}

.search-btn:hover:not(:disabled) {
  background-color: var(--bg-hover);
  color: var(--text-primary);
}

.search-btn:disabled {
  color: var(--text-faint);
  cursor: not-allowed;
}

.search-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 4px;
  flex-shrink: 0;
  padding: 0;
}

.search-close:hover {
  background-color: var(--bg-hover);
  color: var(--text-primary);
}

.search-bar-enter-active {
  transition: all 0.15s ease-out;
}

.search-bar-leave-active {
  transition: all 0.1s ease-in;
}

.search-bar-enter-from {
  opacity: 0;
  transform: translateY(-8px);
}

.search-bar-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
