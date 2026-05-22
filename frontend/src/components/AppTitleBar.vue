<template>
  <div class="titlebar">
    <div class="titlebar-left no-drag" @dblclick.stop>
      <img class="titlebar-icon" :src="appIcon" alt="oneWin" />
      <span class="titlebar-name">oneWin</span>
    </div>

    <div
      class="titlebar-drag-region"
      @dblclick="toggleMaximise"
    />

    <div class="titlebar-actions no-drag" @dblclick.stop>
      <button class="titlebar-btn" title="最小化" aria-label="最小化" @click="minimiseWindow">
        <span class="win-icon win-minimise"></span>
      </button>
      <button
        class="titlebar-btn"
        :title="isMaximised ? '向下还原' : '最大化'"
        :aria-label="isMaximised ? '向下还原' : '最大化'"
        @click="toggleMaximise"
      >
        <span v-if="!isMaximised" class="win-icon win-maximise"></span>
        <span v-else class="win-icon win-restore"></span>
      </button>
      <button class="titlebar-btn danger" title="关闭" aria-label="关闭" @click="closeWindow">
        <span class="win-icon win-close"></span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import {
  Quit,
  WindowMinimise,
  WindowToggleMaximise,
  WindowIsMaximised
} from '../../wailsjs/runtime/runtime'
import appIcon from '../assets/images/appicon.png'

const isMaximised = ref(false)

const syncMaximisedState = async () => {
  try {
    isMaximised.value = await WindowIsMaximised()
  } catch {
    isMaximised.value = false
  }
}

const minimiseWindow = () => {
  WindowMinimise()
}

const toggleMaximise = async () => {
  WindowToggleMaximise()
  window.setTimeout(syncMaximisedState, 60)
}

const closeWindow = () => {
  Quit()
}

const handleWindowResize = () => {
  syncMaximisedState()
}

onMounted(() => {
  syncMaximisedState()
  window.addEventListener('resize', handleWindowResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleWindowResize)
})
</script>

<style scoped>
.titlebar {
  height: 40px;
  display: flex;
  align-items: stretch;
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-light);
  user-select: none;
  -webkit-user-select: none;
  position: relative;
  z-index: 20;
  --wails-draggable: drag;
}

.titlebar-left {
  height: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
  flex-shrink: 0;
}

.titlebar-icon {
  width: 18px;
  height: 18px;
  object-fit: contain;
}

.titlebar-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.2px;
}

.titlebar-drag-region {
  flex: 1;
  min-width: 0;
}

.titlebar-actions {
  display: flex;
  align-items: stretch;
  flex-shrink: 0;
}

.titlebar-btn {
  width: 46px;
  height: 40px;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text-secondary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.15s, color 0.15s;
}

.titlebar-btn:hover {
  background-color: var(--bg-hover);
  color: var(--text-primary);
}

.titlebar-btn.danger:hover {
  background-color: #e81123;
  color: #ffffff;
}

.no-drag,
.no-drag * {
  --wails-draggable: no-drag;
}

.win-icon {
  position: relative;
  display: inline-block;
  width: 12px;
  height: 12px;
}

.win-minimise::before {
  content: '';
  position: absolute;
  left: 1px;
  right: 1px;
  bottom: 2px;
  border-top: 1.5px solid currentColor;
}

.win-maximise::before {
  content: '';
  position: absolute;
  inset: 1px;
  border: 1.5px solid currentColor;
}

.win-restore::before,
.win-restore::after {
  content: '';
  position: absolute;
  border: 1.5px solid currentColor;
  background: var(--bg-secondary);
}

.win-restore::before {
  width: 7px;
  height: 7px;
  top: 1px;
  right: 1px;
}

.win-restore::after {
  width: 7px;
  height: 7px;
  left: 1px;
  bottom: 1px;
}

.win-close::before,
.win-close::after {
  content: '';
  position: absolute;
  left: 5px;
  top: 0;
  width: 1.5px;
  height: 12px;
  background-color: currentColor;
  transform-origin: center;
}

.win-close::before {
  transform: rotate(45deg);
}

.win-close::after {
  transform: rotate(-45deg);
}
</style>
