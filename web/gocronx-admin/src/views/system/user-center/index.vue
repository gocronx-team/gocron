<!-- Personal center — real gocron user info + change password -->
<template>
  <div class="user-center-page">
    <ElRow :gutter="20">
      <!-- Profile card -->
      <ElCol :xs="24" :md="10">
        <ElCard shadow="never" class="profile-card">
          <div class="profile-body">
            <div class="avatar-wrap">
              <ElAvatar :size="80">
                {{ initials }}
              </ElAvatar>
            </div>
            <h2 class="username">{{ userInfo.userName || '-' }}</h2>
            <ElTag
              :type="userInfo.isAdmin === 1 ? 'danger' : 'info'"
              size="small"
              effect="light"
              class="role-tag"
            >
              {{ userInfo.isAdmin === 1 ? t('user.admin') : t('user.normalUser') }}
            </ElTag>

            <div class="info-list">
              <div v-if="userInfo.email" class="info-row">
                <ArtSvgIcon icon="ri:mail-line" class="info-icon" />
                <span>{{ userInfo.email }}</span>
              </div>
            </div>

            <ElDivider />

            <div class="quick-actions">
              <ElButton
                type="primary"
                plain
                @click="goTwoFactor"
                v-ripple
              >
                <ArtSvgIcon icon="ri:shield-keyhole-line" />
                <span class="ml-1">{{ t('menus.system.twoFactor') }}</span>
              </ElButton>
            </div>
          </div>
        </ElCard>
      </ElCol>

      <!-- Change password -->
      <ElCol :xs="24" :md="14">
        <ElCard shadow="never">
          <template #header>
            <span class="card-title">{{ t('changePassword.title') }}</span>
          </template>

          <ElForm
            :model="pwdForm"
            :rules="pwdRules"
            ref="pwdFormRef"
            label-position="top"
            class="pwd-form"
            @submit.prevent
          >
            <ElFormItem :label="t('changePassword.oldPassword')" prop="old_password">
              <ElInput
                v-model="pwdForm.old_password"
                type="password"
                show-password
                autocomplete="current-password"
              />
            </ElFormItem>

            <ElFormItem :label="t('changePassword.newPassword')" prop="new_password">
              <ElInput
                v-model="pwdForm.new_password"
                type="password"
                show-password
                autocomplete="new-password"
              />
            </ElFormItem>

            <ElFormItem :label="t('changePassword.confirmNewPassword')" prop="confirm_new_password">
              <ElInput
                v-model="pwdForm.confirm_new_password"
                type="password"
                show-password
                autocomplete="new-password"
              />
            </ElFormItem>

            <ElFormItem>
              <ElButton
                type="primary"
                :loading="pwdSaving"
                class="submit-btn"
                @click="submitPwd"
                v-ripple
              >
                {{ t('changePassword.save') }}
              </ElButton>
            </ElFormItem>
          </ElForm>
        </ElCard>
      </ElCol>
    </ElRow>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { useUserStore } from '@/store/modules/user'
  import { fetchUserEditMyPassword } from '@/api/user'

  defineOptions({ name: 'UserCenter' })

  const { t } = useI18n()
  const router = useRouter()
  const userStore = useUserStore()
  const userInfo = computed(() => userStore.getUserInfo)

  const initials = computed(() => {
    const name = userInfo.value?.userName || ''
    return name.slice(0, 1).toUpperCase() || '?'
  })

  const pwdFormRef = ref<FormInstance>()
  const pwdSaving = ref(false)

  const pwdForm = reactive({
    old_password: '',
    new_password: '',
    confirm_new_password: ''
  })

  const pwdRules = computed<FormRules>(() => ({
    old_password: [
      { required: true, message: t('changePassword.oldPasswordRequired'), trigger: 'blur' }
    ],
    new_password: [
      { required: true, message: t('changePassword.newPasswordRequired'), trigger: 'blur' },
      { min: 8, message: t('changePassword.tooShort'), trigger: 'blur' }
    ],
    confirm_new_password: [
      { required: true, message: t('changePassword.confirmNewPasswordRequired'), trigger: 'blur' },
      {
        validator: (_rule: any, value: string, callback: (e?: Error) => void) => {
          if (value !== pwdForm.new_password) {
            callback(new Error(t('changePassword.mismatch')))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ]
  }))

  async function submitPwd() {
    if (!pwdFormRef.value) return
    const valid = await pwdFormRef.value.validate().catch(() => false)
    if (!valid) return

    pwdSaving.value = true
    try {
      await fetchUserEditMyPassword({
        old_password: pwdForm.old_password,
        new_password: pwdForm.new_password,
        confirm_new_password: pwdForm.confirm_new_password
      })
      ElMessage.success(t('changePassword.saveSuccess'))
      pwdFormRef.value.resetFields()
    } catch {
      // error handled by http interceptor
    } finally {
      pwdSaving.value = false
    }
  }

  function goTwoFactor() {
    router.push('/system/two-factor')
  }
</script>

<style scoped>
  .user-center-page {
    padding: 16px;
  }

  .profile-card {
    height: 100%;
  }

  .profile-body {
    padding: 16px 8px;
    text-align: center;
  }

  .avatar-wrap {
    display: flex;
    justify-content: center;
    margin-bottom: 16px;
  }

  .username {
    font-size: 20px;
    font-weight: 500;
    margin: 0 0 8px;
  }

  .role-tag {
    margin-bottom: 16px;
  }

  .info-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: 260px;
    margin: 0 auto;
    text-align: left;
  }

  .info-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: var(--el-text-color-regular);
  }

  .info-icon {
    color: var(--el-text-color-secondary);
  }

  .quick-actions {
    display: flex;
    justify-content: center;
    gap: 10px;
  }

  .card-title {
    font-size: 15px;
    font-weight: 500;
  }

  .pwd-form {
    max-width: 420px;
  }

  .pwd-form :deep(.el-form-item__label) {
    padding-bottom: 6px;
    font-size: 13px;
    color: var(--el-text-color-regular);
    line-height: 1.4;
  }

  .pwd-form :deep(.el-form-item) {
    margin-bottom: 18px;
  }

  .submit-btn {
    min-width: 120px;
  }
</style>
