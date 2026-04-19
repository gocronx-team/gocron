<template>
  <div class="two-factor-page">
    <!-- Status card -->
    <ElCard class="status-card">
      <template #header>
        <span class="card-title">{{ t('twoFactor.title') }}</span>
      </template>

      <!-- Loading skeleton -->
      <div v-if="statusLoading" class="flex-c gap-3 py-4">
        <ElSkeleton :rows="2" animated />
      </div>

      <!-- Not enabled -->
      <div v-else-if="!twoFactorEnabled">
        <ElAlert
          :title="t('twoFactor.disabledAlertTitle')"
          type="info"
          :description="t('twoFactor.disabledAlertDesc')"
          :closable="false"
          show-icon
        />
        <ElButton
          type="primary"
          class="mt-5"
          :loading="setupLoading"
          @click="handleSetup"
        >
          {{ t('twoFactor.enableBtn') }}
        </ElButton>
      </div>

      <!-- Already enabled -->
      <div v-else>
        <ElAlert
          :title="t('twoFactor.enabledAlertTitle')"
          type="success"
          :description="t('twoFactor.enabledAlertDesc')"
          :closable="false"
          show-icon
        />
        <ElButton
          type="danger"
          class="mt-5"
          @click="showDisableDialog"
        >
          {{ t('twoFactor.disableBtn') }}
        </ElButton>
      </div>
    </ElCard>

    <!-- Setup dialog (scan QR + enter code to enable) -->
    <ElDialog
      v-model="setupDialogVisible"
      :title="t('twoFactor.setupDialogTitle')"
      width="500px"
      :close-on-click-modal="false"
      @closed="resetSetupState"
    >
      <div v-if="qrCode" class="setup-content">
        <p class="mb-3 text-sm">{{ t('twoFactor.scanQr') }}</p>

        <div class="qr-wrapper">
          <img :src="qrCode" alt="2FA QR Code" class="qr-img" />
        </div>

        <p class="mt-4 mb-2 text-sm">{{ t('twoFactor.manualSecret') }}</p>
        <ElInput v-model="secret" readonly>
          <template #append>
            <ElButton @click="copySecret">{{ t('twoFactor.copySecret') }}</ElButton>
          </template>
        </ElInput>

        <p class="mt-5 mb-2 text-sm">{{ t('twoFactor.inputCode') }}</p>
        <ElInput
          v-model="verifyCode"
          :placeholder="t('twoFactor.codePlaceholder')"
          maxlength="6"
          @keyup.enter="handleEnable"
        />
      </div>

      <template #footer>
        <ElButton @click="setupDialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton
          type="primary"
          :loading="enableLoading"
          @click="handleEnable"
        >
          {{ t('twoFactor.save') }}
        </ElButton>
      </template>
    </ElDialog>

    <!-- Disable dialog (enter code to disable) -->
    <ElDialog
      v-model="disableDialogVisible"
      :title="t('twoFactor.disableDialogTitle')"
      width="420px"
      :close-on-click-modal="false"
      @closed="disableCode = ''"
    >
      <p class="mb-3 text-sm">{{ t('twoFactor.disableDialogDesc') }}</p>
      <ElInput
        v-model="disableCode"
        :placeholder="t('twoFactor.codePlaceholder')"
        maxlength="6"
        @keyup.enter="handleDisable"
      />

      <template #footer>
        <ElButton @click="disableDialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton
          type="danger"
          :loading="disableLoading"
          @click="handleDisable"
        >
          {{ t('twoFactor.disableBtn') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { get2FAStatus, setup2FA, enable2FA, disable2FA } from '@/api/user'

  defineOptions({ name: 'TwoFactor' })

  const { t } = useI18n()

  // ── state ──────────────────────────────────────────────────────────────────
  const statusLoading = ref(false)
  const setupLoading = ref(false)
  const enableLoading = ref(false)
  const disableLoading = ref(false)

  const twoFactorEnabled = ref(false)
  const setupDialogVisible = ref(false)
  const disableDialogVisible = ref(false)

  const qrCode = ref('')
  const secret = ref('')
  const verifyCode = ref('')
  const disableCode = ref('')

  // ── lifecycle ──────────────────────────────────────────────────────────────
  onMounted(() => {
    fetchStatus()
  })

  // ── methods ────────────────────────────────────────────────────────────────
  async function fetchStatus() {
    statusLoading.value = true
    try {
      const res = await get2FAStatus()
      twoFactorEnabled.value = (res as any)?.data?.enabled ?? false
    } finally {
      statusLoading.value = false
    }
  }

  async function handleSetup() {
    setupLoading.value = true
    try {
      const res = await setup2FA()
      const data = (res as any)?.data
      qrCode.value = data?.qr_code ?? ''
      secret.value = data?.secret ?? ''
      setupDialogVisible.value = true
    } finally {
      setupLoading.value = false
    }
  }

  async function handleEnable() {
    if (!verifyCode.value || verifyCode.value.length !== 6) {
      ElMessage.warning(t('twoFactor.codeLength'))
      return
    }
    enableLoading.value = true
    try {
      await enable2FA(secret.value, verifyCode.value)
      ElMessage.success(t('twoFactor.enableSuccess'))
      twoFactorEnabled.value = true
      setupDialogVisible.value = false
    } finally {
      enableLoading.value = false
    }
  }

  function showDisableDialog() {
    disableCode.value = ''
    disableDialogVisible.value = true
  }

  async function handleDisable() {
    if (!disableCode.value || disableCode.value.length !== 6) {
      ElMessage.warning(t('twoFactor.codeLength'))
      return
    }
    disableLoading.value = true
    try {
      await disable2FA(disableCode.value)
      ElMessage.success(t('twoFactor.disableSuccess'))
      twoFactorEnabled.value = false
      disableDialogVisible.value = false
    } finally {
      disableLoading.value = false
    }
  }

  function copySecret() {
    navigator.clipboard.writeText(secret.value).then(
      () => ElMessage.success(t('twoFactor.secretCopied')),
      () => {
        // Fallback for older browsers
        const el = document.createElement('input')
        el.value = secret.value
        document.body.appendChild(el)
        el.select()
        document.execCommand('copy')
        document.body.removeChild(el)
        ElMessage.success(t('twoFactor.secretCopied'))
      }
    )
  }

  function resetSetupState() {
    qrCode.value = ''
    secret.value = ''
    verifyCode.value = ''
  }
</script>

<style scoped>
.two-factor-page {
  padding: 20px;
}

.status-card {
  max-width: 620px;
}

.card-title {
  font-size: 16px;
  font-weight: 500;
}

.setup-content {
  padding: 4px 0;
}

.qr-wrapper {
  display: flex;
  justify-content: center;
  margin: 16px 0;
}

.qr-img {
  width: 200px;
  height: 200px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
}
</style>
