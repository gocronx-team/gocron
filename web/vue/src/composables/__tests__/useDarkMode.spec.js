import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { useDarkMode } from '../useDarkMode'

describe('useDarkMode', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.classList.remove('dark')
  })

  it('toggle flips state and DOM class', () => {
    const { isDark, toggle } = useDarkMode()
    const before = isDark.value
    toggle()
    expect(isDark.value).toBe(!before)
  })

  it('set forces state', async () => {
    const { isDark, set } = useDarkMode()
    set(true)
    await new Promise(r => setTimeout(r, 10))
    expect(isDark.value).toBe(true)
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })

  it('persists to localStorage', async () => {
    const { set } = useDarkMode()
    set(true)
    await new Promise(r => setTimeout(r, 10))
    expect(localStorage.getItem('theme')).toBe('dark')
  })
})
