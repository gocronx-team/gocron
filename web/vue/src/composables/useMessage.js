import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

// 注意：此 composable 保留 Element Plus 的 ElMessageBox 行为，
// 不走 useNotify facade，以保持 confirm 对话框的样式 options。
// Phase 8 卸载 Element Plus 时再迁移到 shadcn AlertDialog。
export function useMessage() {
  const { t } = useI18n()

  const success = (message) => {
    ElMessage.success(message)
  }

  const error = (message) => {
    ElMessage.error(message)
  }

  const warning = (message) => {
    ElMessage.warning(message)
  }

  const info = (message) => {
    ElMessage.info(message)
  }

  const confirm = (message, title, options = {}) => {
    return ElMessageBox.confirm(
      message,
      title || t('common.tip'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
        center: true,
        ...options
      }
    )
  }

  return {
    success,
    error,
    warning,
    info,
    confirm
  }
}
