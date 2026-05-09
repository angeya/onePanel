import { AddHistory } from '../../wailsjs/go/main/HistoryService'

/**
 * 终端事件管理组合式函数
 * 负责终端间命令传递和搜索事件的全局事件总线
 * 替代直接使用 window.dispatchEvent 的原始方式，提供类型安全和集中管理
 */
export function useTerminalEvent() {
  /**
   * 向指定终端发送命令
   * 通过 CustomEvent 实现跨组件通信
   */
  const sendCommand = (tabId, command) => {
    const event = new CustomEvent('terminal-send-command', {
      detail: { tabId, command }
    })
    window.dispatchEvent(event)
  }

  /**
   * 记录命令执行历史
   * 忽略记录失败的错误，不影响主流程
   */
  const recordHistory = (command) => {
    if (command && command.trim()) {
      AddHistory(command, 'cmd.exe', '').catch(() => {})
    }
  }

  /**
   * 发送搜索事件到当前激活的终端
   */
  const dispatchSearch = (tabId, action, keyword) => {
    const event = new CustomEvent('terminal-search', {
      detail: { tabId, action, keyword }
    })
    window.dispatchEvent(event)
  }

  return {
    sendCommand,
    recordHistory,
    dispatchSearch
  }
}
