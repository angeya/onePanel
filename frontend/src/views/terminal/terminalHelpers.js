/**
 * detectTerminalSshState 根据终端输出内容推断 SSH 连接状态。
 * 返回值说明：
 * - disconnected: 已检测到断开连接，应清空 sshHost
 * - connected: 已检测到登录成功，可用于更新标签状态
 * - unchanged: 未检测到需要变更的状态
 */
export function detectTerminalSshState({
  sshHost,
  data,
  disconnectPatterns,
  connectIndicators
}) {
  if (sshHost) {
    const lowerData = data.toLowerCase()
    for (const pattern of disconnectPatterns) {
      if (lowerData.includes(pattern.toLowerCase())) {
        return 'disconnected'
      }
    }
    return 'unchanged'
  }

  if (sshHost === null) {
    for (const indicator of connectIndicators) {
      if (data.includes(indicator)) {
        return 'connected'
      }
    }
  }

  return 'unchanged'
}
