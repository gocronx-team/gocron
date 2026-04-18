import { ref, watch, onMounted } from 'vue'

/**
 * Dark mode 切换 composable。
 *
 * - 初始值：localStorage('theme') → 'dark'/'light' → 否则跟随系统偏好
 * - 切换：添加/移除 <html class="dark">
 * - 持久化：localStorage
 *
 * Phase 1 阶段仅 shadcn 组件生效（通过 .dark class + CSS 变量）；
 * Element Plus 组件在 Phase 8 之后统一生效。
 */
const getInitial = () => {
  if (typeof localStorage === 'undefined') return false
  const saved = localStorage.getItem('theme')
  if (saved === 'dark') return true
  if (saved === 'light') return false
  return typeof window !== 'undefined' &&
    window.matchMedia?.('(prefers-color-scheme: dark)').matches
}

const isDark = ref(getInitial())

const apply = (dark) => {
  if (typeof document === 'undefined') return
  document.documentElement.classList.toggle('dark', dark)
}

const persist = (dark) => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('theme', dark ? 'dark' : 'light')
  }
}

let initialized = false

export function useDarkMode() {
  if (!initialized) {
    initialized = true
    apply(isDark.value)
    watch(isDark, (v) => {
      apply(v)
      persist(v)
    })
  }

  const toggle = () => { isDark.value = !isDark.value }
  const set = (v) => {
    isDark.value = !!v
    apply(isDark.value)
    persist(isDark.value)
  }

  return { isDark, toggle, set }
}
