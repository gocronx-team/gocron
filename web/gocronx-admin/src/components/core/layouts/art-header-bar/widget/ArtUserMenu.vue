<!-- 用户菜单 -->
<template>
  <ElPopover
    ref="userMenuPopover"
    placement="bottom-end"
    :width="240"
    :hide-after="0"
    :offset="10"
    trigger="hover"
    :show-arrow="false"
    popper-class="user-menu-popover"
    popper-style="padding: 5px 16px;"
  >
    <template #reference>
      <div
        class="user-avatar size-8.5 mr-5 c-p rounded-full max-sm:w-6.5 max-sm:h-6.5 max-sm:mr-[16px]"
        :style="{ background: avatarColor }"
      >
        <span class="avatar-initial">{{ avatarInitial }}</span>
      </div>
    </template>
    <template #default>
      <div class="pt-3">
        <div class="flex-c pb-1 px-0">
          <div
            class="user-avatar w-10 h-10 mr-3 ml-0 rounded-full"
            :style="{ background: avatarColor }"
          >
            <span class="avatar-initial text-base">{{ avatarInitial }}</span>
          </div>
          <div class="w-[calc(100%-60px)] h-full">
            <span class="block text-sm font-medium text-g-800 truncate">{{
              userInfo.userName
            }}</span>
            <span class="block mt-0.5 text-xs text-g-500 truncate">{{ userInfo.email }}</span>
          </div>
        </div>
        <ul class="py-4 mt-3 border-t border-g-300/80">
          <li class="btn-item" @click="goPage('/system/user-center')">
            <ArtSvgIcon icon="ri:user-3-line" />
            <span>{{ $t('topBar.user.userCenter') }}</span>
          </li>
          <li class="btn-item" @click="toDocs()">
            <ArtSvgIcon icon="ri:book-2-line" />
            <span>{{ $t('topBar.user.docs') }}</span>
          </li>
          <li class="btn-item" @click="toGithub()">
            <ArtSvgIcon icon="ri:github-line" />
            <span>{{ $t('topBar.user.github') }}</span>
          </li>
          <div class="w-full h-px my-2 bg-g-300/80"></div>
          <div class="log-out c-p" @click="loginOut">
            {{ $t('topBar.user.logout') }}
          </div>
        </ul>
      </div>
    </template>
  </ElPopover>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { ElMessageBox } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'
  import { WEB_LINKS } from '@/utils/constants'

  defineOptions({ name: 'ArtUserMenu' })

  const router = useRouter()
  const { t, locale } = useI18n()
  const userStore = useUserStore()

  const { getUserInfo: userInfo } = storeToRefs(userStore)
  const userMenuPopover = ref()

  // Deterministic avatar: first letter of username on a color hashed from the name.
  const avatarPalette = [
    '#409EFF', '#67C23A', '#E6A23C', '#F56C6C',
    '#9B59B6', '#1ABC9C', '#34495E', '#E67E22'
  ]
  const avatarInitial = computed(() => {
    const n = userInfo.value?.userName || ''
    return n ? n.charAt(0).toUpperCase() : '?'
  })
  const avatarColor = computed(() => {
    const n = userInfo.value?.userName || ''
    let hash = 0
    for (let i = 0; i < n.length; i++) hash = (hash * 31 + n.charCodeAt(i)) & 0xffffffff
    return avatarPalette[Math.abs(hash) % avatarPalette.length]
  })

  /**
   * 页面跳转
   * @param {string} path - 目标路径
   */
  const goPage = (path: string): void => {
    router.push(path)
  }

  /**
   * 打开文档页面（跟随当前语言切换中英文站点）
   */
  const toDocs = (): void => {
    const url = locale.value === 'zh' ? WEB_LINKS.DOCS_ZH : WEB_LINKS.DOCS_EN
    window.open(url)
  }

  /**
   * 打开 GitHub 页面
   */
  const toGithub = (): void => {
    window.open(WEB_LINKS.GITHUB)
  }

  /**
   * 用户登出确认
   */
  const loginOut = (): void => {
    closeUserMenu()
    setTimeout(() => {
      ElMessageBox.confirm(t('common.logOutTips'), t('common.tips'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        customClass: 'login-out-dialog'
      }).then(() => {
        userStore.logOut()
      })
    }, 200)
  }

  /**
   * 关闭用户菜单弹出层
   */
  const closeUserMenu = (): void => {
    setTimeout(() => {
      userMenuPopover.value.hide()
    }, 100)
  }
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  @layer components {
    .btn-item {
      @apply flex items-center p-2 mb-3 select-none rounded-md cursor-pointer last:mb-0;

      span {
        @apply text-sm;
      }

      .art-svg-icon {
        @apply mr-2 text-base;
      }

      &:hover {
        background-color: var(--art-gray-200);
      }
    }
  }

  .log-out {
    @apply py-1.5
    mt-5
    text-xs
    text-center
    border
    border-g-400
    rounded-md
    transition-all
    duration-200
    hover:shadow-xl;
  }

  .user-avatar {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-weight: 600;
    line-height: 1;
    user-select: none;
    flex-shrink: 0;
  }

  .avatar-initial {
    font-size: 14px;
  }
</style>
