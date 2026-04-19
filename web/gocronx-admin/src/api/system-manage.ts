import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'

// 获取菜单列表（仅 backend 菜单模式使用；前端菜单模式不会调用此接口）
export function fetchGetMenuList() {
  return request.get<AppRouteRecord[]>({
    url: '/api/v3/system/menus/simple'
  })
}
