import { describe, it, expect, beforeEach, vi } from 'vitest'

// mock element-plus 和 vue-sonner
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(), error: vi.fn(), info: vi.fn(), warning: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn(() => Promise.resolve())
  }
}))
vi.mock('vue-sonner', () => ({
  toast: Object.assign(vi.fn(), {
    success: vi.fn(), error: vi.fn(), warning: vi.fn()
  })
}))

import { ElMessage, ElMessageBox } from 'element-plus'
import { toast } from 'vue-sonner'
import { useNotify } from '../useNotify'

describe('useNotify', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  it('defaults to element backend', () => {
    const n = useNotify()
    n.success('hi')
    expect(ElMessage.success).toHaveBeenCalledWith('hi')
  })

  it('uses sonner when backend is set', () => {
    localStorage.setItem('notify.backend', 'sonner')
    const n = useNotify()
    n.success('hi')
    expect(toast.success).toHaveBeenCalledWith('hi')
  })

  it('confirm resolves true on accept', async () => {
    const n = useNotify()
    expect(await n.confirm('sure?')).toBe(true)
  })

  it('confirm resolves false on reject', async () => {
    ElMessageBox.confirm.mockRejectedValueOnce(new Error('cancel'))
    const n = useNotify()
    expect(await n.confirm('sure?')).toBe(false)
  })
})
