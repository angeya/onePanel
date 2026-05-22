import { ref } from 'vue'

/**
 * useBootProgress 管理应用启动阶段的进度条与显示状态。
 * 通过平滑推进的方式减少启动过程中的空白等待感。
 */
export function useBootProgress() {
  const appReady = ref(false)
  const bootProgress = ref(0)
  const bootStatus = ref('正在初始化界面')

  let bootProgressTimer = null

  /**
   * updateBootProgress 更新当前进度值和展示文案。
   * 进度值会被限制在 0 到 100 之间，避免出现非法状态。
   */
  const updateBootProgress = (value, status) => {
    bootProgress.value = Math.max(0, Math.min(100, value))
    if (status) {
      bootStatus.value = status
    }
  }

  /**
   * startBootProgress 启动预加载阶段的平滑推进动画。
   * 在真实初始化完成前，进度条会缓慢前进到安全阈值附近。
   */
  const startBootProgress = () => {
    clearInterval(bootProgressTimer)
    updateBootProgress(8, '正在初始化界面')
    bootProgressTimer = window.setInterval(() => {
      if (bootProgress.value < 88) {
        bootProgress.value += bootProgress.value < 40 ? 8 : 4
      }
    }, 120)
  }

  /**
   * finishBootProgress 将进度条补满并切换到主界面。
   * 额外保留一小段延迟，让进度完成动画更自然。
   */
  const finishBootProgress = () => {
    clearInterval(bootProgressTimer)
    updateBootProgress(100, '初始化完成')
    window.setTimeout(() => {
      appReady.value = true
    }, 180)
  }

  /**
   * cleanupBootProgress 停止启动进度条内部定时器。
   * 在组件卸载时调用，避免残留异步任务。
   */
  const cleanupBootProgress = () => {
    clearInterval(bootProgressTimer)
  }

  return {
    appReady,
    bootProgress,
    bootStatus,
    updateBootProgress,
    startBootProgress,
    finishBootProgress,
    cleanupBootProgress
  }
}
