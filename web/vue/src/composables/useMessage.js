import { useNotify } from './useNotify'
import { useI18n } from 'vue-i18n'

export function useMessage() {
  const { t } = useI18n()
  const notify = useNotify()

  const success = (message) => {
    notify.success(message)
  }

  const error = (message) => {
    notify.error(message)
  }

  const warning = (message) => {
    notify.warning(message)
  }

  const info = (message) => {
    notify.info(message)
  }

  const confirm = (message, title, options = {}) => {
    // options ignored in facade; use notify.confirm for consistent behavior
    return notify.confirm(message, title || t('common.tip'))
  }

  return {
    success,
    error,
    warning,
    info,
    confirm
  }
}
