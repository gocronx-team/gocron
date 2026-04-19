/**
 * 全局事件总线模块
 *
 * 基于 mitt 库实现的类型安全的事件总线
 *
 * ## 用法示例
 *
 * ```typescript
 * mittBus.on('openSetting', () => { ... })
 * mittBus.emit('openSetting')
 * ```
 *
 * @module utils/sys/mittBus
 * @author GoCronX Team
 */
import mitt, { type Emitter } from 'mitt'

// 定义事件类型映射
type Events = {
  // 打开设置面板事件 - 无参数
  openSetting: void
}

// 创建类型安全的事件总线实例
const mittBus: Emitter<Events> = mitt<Events>()

export default mittBus
