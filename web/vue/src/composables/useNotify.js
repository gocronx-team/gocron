import { ElMessage, ElMessageBox } from 'element-plus'
import { toast as sonnerToast } from 'vue-sonner'

/**
 * 迁移期间的通知 facade：通过 localStorage 'notify.backend' 切换底层
 * - 'element'（默认）：用 Element Plus 的 ElMessage / ElMessageBox
 * - 'sonner'：用 vue-sonner（shadcn 默认 toast）
 *
 * 迁移完成（Phase 8）时移除 Element Plus 分支。
 */
const getBackend = () => (typeof localStorage !== 'undefined' && localStorage.getItem('notify.backend')) || 'element'

export function useNotify() {
  const backend = getBackend()
  const useSonner = backend === 'sonner'

  return {
    success(msg) {
      useSonner ? sonnerToast.success(msg) : ElMessage.success(msg)
    },
    error(msg) {
      useSonner ? sonnerToast.error(msg) : ElMessage.error(msg)
    },
    info(msg) {
      useSonner ? sonnerToast(msg) : ElMessage.info(msg)
    },
    warning(msg) {
      useSonner ? sonnerToast.warning(msg) : ElMessage.warning(msg)
    },
    /**
     * 确认对话框：返回 Promise<boolean>
     * @returns {Promise<boolean>} 用户点击确定返回 true，取消返回 false
     */
    async confirm(message, title = '提示') {
      if (useSonner) {
        // TODO(Phase 7): 接入 shadcn AlertDialog 全局实例
        // 过渡期间仍通过 ElMessageBox 实现，保证行为一致
      }
      try {
        await ElMessageBox.confirm(message, title)
        return true
      } catch {
        return false
      }
    }
  }
}
