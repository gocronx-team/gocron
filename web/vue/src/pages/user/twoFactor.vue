<template>
  <div class="two-factor-container">
    <el-card class="box-card">
      <template #header>
        <div class="clearfix">
          <span>{{ t('twoFactor.title') }}</span>
        </div>
      </template>
      
      <div v-if="!twoFactorEnabled">
        <el-alert
          :title="t('twoFactor.alertTitle')"
          type="info"
          :description="t('twoFactor.alertDescription')"
          :closable="false"
          show-icon
        />
        
        <el-button 
          type="primary" 
          style="margin-top: 20px;" 
          :loading="loading"
          @click="setup2FA"
        >
          {{ t('twoFactor.enable') }}
        </el-button>
      </div>

      <div v-else>
        <el-alert
          :title="t('twoFactor.enabledAlertTitle')"
          type="success"
          :description="t('twoFactor.enabledAlertDescription')"
          :closable="false"
          show-icon
        />
        
        <el-button 
          type="danger" 
          style="margin-top: 20px;" 
          @click="showDisableDialog"
        >
          {{ t('twoFactor.disable') }}
        </el-button>
      </div>
    </el-card>

    <el-dialog
      v-model="setupDialogVisible"
      :title="t('twoFactor.setup')"
      width="500px"
      :close-on-click-modal="false"
    >
      <div v-if="qrCode">
        <p>{{ t('twoFactor.scanQR') }}</p>
        <div style="text-align: center; margin: 20px 0;">
          <img
            :src="qrCode"
            alt="QR Code"
            style="width: 200px; height: 200px;"
          >
        </div>
        
        <p>{{ t('twoFactor.manualEntry') }}</p>
        <el-input
          v-model="secret"
          readonly
        >
          <template #append>
            <el-button @click="copySecret">
              {{ t('twoFactor.copySecret') }}
            </el-button>
          </template>
        </el-input>
        
        <p style="margin-top: 20px;">
          {{ t('twoFactor.verifyCodeStep') }}
        </p>
        <el-input 
          v-model="verifyCode" 
          :placeholder="t('twoFactor.verifyCodePlaceholder')"
          maxlength="6"
          @keyup.enter="enable2FA"
        />
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="setupDialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button
            type="primary"
            :loading="loading"
            @click="enable2FA"
          >{{ t('twoFactor.confirm') }}</el-button>
        </span>
      </template>
    </el-dialog>

    <el-dialog
      v-model="disableDialogVisible"
      :title="t('twoFactor.disableDialogTitle')"
      width="400px"
      :close-on-click-modal="false"
    >
      <p>{{ t('twoFactor.disableDialogDescription') }}</p>
      <el-input 
        v-model="disableCode" 
        :placeholder="t('twoFactor.verifyCodePlaceholder')"
        maxlength="6"
        @keyup.enter="disable2FA"
      />

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="disableDialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button
            type="danger"
            :loading="loading"
            @click="disable2FA"
          >{{ t('twoFactor.confirmDisable') }}</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useNotify } from '@/composables/useNotify'
import userApi from '@/api/user'

const { t } = useI18n()
const notify = useNotify()

const twoFactorEnabled = ref(false)
const loading = ref(false)
const setupDialogVisible = ref(false)
const disableDialogVisible = ref(false)
const qrCode = ref('')
const secret = ref('')
const verifyCode = ref('')
const disableCode = ref('')

onMounted(() => {
  check2FAStatus()
})

const check2FAStatus = () => {
  userApi.get2FAStatus((data) => {
    twoFactorEnabled.value = data.enabled
  })
}

const setup2FA = () => {
  loading.value = true
  userApi.setup2FA((data) => {
    qrCode.value = data.qr_code
    secret.value = data.secret
    setupDialogVisible.value = true
    loading.value = false
  })
}

const enable2FA = () => {
  if (!verifyCode.value || verifyCode.value.length !== 6) {
    notify.warning(t('twoFactor.verifyCodeLength'))
    return
  }

  loading.value = true
  userApi.enable2FA(secret.value, verifyCode.value, () => {
    notify.success(t('twoFactor.enableSuccess'))
    setupDialogVisible.value = false
    twoFactorEnabled.value = true
    verifyCode.value = ''
    loading.value = false
  })
}

const showDisableDialog = () => {
  disableCode.value = ''
  disableDialogVisible.value = true
}

const disable2FA = () => {
  if (!disableCode.value || disableCode.value.length !== 6) {
    notify.warning(t('twoFactor.verifyCodeLength'))
    return
  }

  loading.value = true
  userApi.disable2FA(disableCode.value, () => {
    notify.success(t('twoFactor.disableSuccess'))
    disableDialogVisible.value = false
    twoFactorEnabled.value = false
    disableCode.value = ''
    loading.value = false
  }, (code, msg) => {
    notify.error(msg || t('twoFactor.disableFailed'))
    loading.value = false
  })
}

const copySecret = () => {
  const input = document.createElement('input')
  input.value = secret.value
  document.body.appendChild(input)
  input.select()
  document.execCommand('copy')
  document.body.removeChild(input)
  notify.success(t('twoFactor.secretCopied'))
}
</script>

<style scoped>
.two-factor-container {
  padding: 20px;
}

.box-card {
  max-width: 600px;
}
</style>
