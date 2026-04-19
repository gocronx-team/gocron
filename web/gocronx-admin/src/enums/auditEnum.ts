/**
 * Audit log enum definitions
 */

export const AUDIT_MODULES = [
  { value: 'task', labelKey: 'audit.module_task' },
  { value: 'host', labelKey: 'audit.module_host' },
  { value: 'user', labelKey: 'audit.module_user' },
  { value: 'system', labelKey: 'audit.module_system' }
] as const

export const AUDIT_ACTIONS = [
  { value: 'create', labelKey: 'audit.action_create' },
  { value: 'update', labelKey: 'audit.action_update' },
  { value: 'delete', labelKey: 'audit.action_delete' },
  { value: 'enable', labelKey: 'audit.action_enable' },
  { value: 'disable', labelKey: 'audit.action_disable' },
  { value: 'run', labelKey: 'audit.action_run' },
  { value: 'batch-enable', labelKey: 'audit.action_batch_enable' },
  { value: 'batch-disable', labelKey: 'audit.action_batch_disable' },
  { value: 'batch-remove', labelKey: 'audit.action_batch_remove' },
  { value: 'change-password', labelKey: 'audit.action_change_password' },
  { value: 'reset-password', labelKey: 'audit.action_reset_password' }
] as const

export const MODULE_TAG_TYPES: Record<
  string,
  'primary' | 'success' | 'warning' | 'danger' | 'info'
> = {
  task: 'primary',
  host: 'success',
  user: 'warning',
  system: 'danger'
}

export const ACTION_TAG_TYPES: Record<
  string,
  'primary' | 'success' | 'warning' | 'danger' | 'info'
> = {
  create: 'success',
  update: 'warning',
  delete: 'danger',
  enable: 'success',
  disable: 'info',
  run: 'primary',
  'batch-enable': 'success',
  'batch-disable': 'info',
  'batch-remove': 'danger',
  'change-password': 'warning',
  'reset-password': 'warning'
}
