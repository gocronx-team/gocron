import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * 合并 class name 的工具（shadcn-vue 标配）
 * @param {...import('clsx').ClassValue} inputs
 * @returns {string}
 */
export function cn(...inputs) {
  return twMerge(clsx(inputs))
}
