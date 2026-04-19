/**
 * API 接口类型定义模块
 *
 * 提供所有后端接口的类型定义
 *
 * ## 主要功能
 *
 * - 通用类型（分页参数、响应结构等）
 * - 认证类型（登录、用户信息等）
 * - 系统管理类型（用户、角色等）
 * - 全局命名空间声明
 *
 * ## 使用场景
 *
 * - API 请求参数类型约束
 * - API 响应数据类型定义
 * - 接口文档类型同步
 *
 * ## 注意事项
 *
 * - 在 .vue 文件使用需要在 eslint.config.mjs 中配置 globals: { Api: 'readonly' }
 * - 使用全局命名空间，无需导入即可使用
 *
 * ## 使用方式
 *
 * ```typescript
 * const params: Api.Auth.LoginParams = { userName: 'admin', password: '123456' }
 * const response: Api.Auth.UserInfo = await fetchUserInfo()
 * ```
 *
 * @module types/api/api
 * @author GoCronX Team
 */

declare namespace Api {
  /** 通用类型 */
  namespace Common {
    /** 分页参数 */
    interface PaginationParams {
      /** 当前页码 */
      current: number
      /** 每页条数 */
      size: number
      /** 总条数 */
      total: number
    }

    /** 通用搜索参数 */
    type CommonSearchParams = Pick<PaginationParams, 'current' | 'size'>

    /** 分页响应基础结构 */
    interface PaginatedResponse<T = any> {
      records: T[]
      current: number
      size: number
      total: number
    }

    /** 启用状态 */
    type EnableStatus = '1' | '2'
  }

  /** 认证类型 */
  namespace Auth {
    /** 登录参数 - gocron /api/user/login (form body) */
    interface LoginParams {
      username: string
      password: string
      two_factor_code?: string
    }

    /** 登录响应 data 字段 - gocron success */
    interface LoginResponse {
      token: string
      uid: number
      username: string
      is_admin: number // 0 or 1
    }

    /** 2FA 中间步骤响应 data 字段 */
    interface Login2FARequired {
      require_2fa: true
    }

    /** 用户信息 - derived from JWT claims stored at login */
    interface UserInfo {
      /** gocron user id */
      userId: number
      userName: string
      /** is_admin: 1 = admin, 0 = regular user */
      isAdmin: number
      /** roles list derived from is_admin for template permission system */
      roles: string[]
      /** not used by gocron, kept for template compatibility */
      buttons: string[]
      email: string
      avatar?: string
    }
  }
}
