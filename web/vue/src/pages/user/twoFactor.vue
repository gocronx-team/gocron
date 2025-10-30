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
          show-icon>
        </el-alert>
        
        <el-button 
          type="primary" 
          @click="setup2FA" 
          style="margin-top: 20px;"
          :loading="loading">
          {{ t('twoFactor.enable') }}
        </el-button>
      </div>

      <div v-else>
        <el-alert
          :title="t('twoFactor.enabledAlertTitle')"
          type="success"
          :description="t('twoFactor.enabledAlertDescription')"
          :closable="false"
          show-icon>
        </el-alert>
        
        <el-button 
          type="danger" 
          @click="showDisableDialog" 
          style="margin-top: 20px;">
          {{ t('twoFactor.disable') }}
        </el-button>
      </div>
    </el-card>

    <el-dialog
      :title="t('twoFactor.setup')"
      v-model="setupDialogVisible"
      width="500px"
      :close-on-click-modal="false">
      
      <div v-if="qrCode">
        <p>{{ t('twoFactor.scanQR') }}</p>
        <div style="text-align: center; margin: 20px 0;">
          <img :src="qrCode" alt="QR Code" style="width: 200px; height: 200px;">
        </div>
        
        <p>{{ t('twoFactor.manualEntry') }}</p>
        <el-input v-model="secret" readonly>
          <template #append>
            <el-button @click="copySecret">{{ t('twoFactor.copySecret') }}</el-button>
          </template>
        </el-input>
        
        <p style="margin-top: 20px;">{{ t('twoFactor.verifyCodeStep') }}</p>
        <el-input 
          v-model="verifyCode" 
          :placeholder="t('twoFactor.verifyCodePlaceholder')"
          maxlength="6"
          @keyup.enter="enable2FA">
        </el-input>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="setupDialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button type="primary" @click="enable2FA" :loading="loading">{{ t('twoFactor.confirm') }}</el-button>
        </span>
      </template>
    </el-dialog>

    <el-dialog
      :title="t('twoFactor.disableDialogTitle')"
      v-model="disableDialogVisible"
      width="400px"
      :close-on-click-modal="false">
      
      <p>{{ t('twoFactor.disableDialogDescription') }}</p>
      <el-input 
        v-model="disableCode" 
        :placeholder="t('twoFactor.verifyCodePlaceholder')"
        maxlength="6"
        @keyup.enter="disable2FA">
      </el-input>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="disableDialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button type="danger" @click="disable2FA" :loading="loading">{{ t('twoFactor.confirmDisable') }}</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import userApi from '@/api/user'

const { t } = useI18n()

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
    ElMessage.warning(t('twoFactor.verifyCodeLength'))
    return
  }

  loading.value = true
  userApi.enable2FA(secret.value, verifyCode.value, () => {
    ElMessage.success(t('twoFactor.enableSuccess'))
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
    ElMessage.warning(t('twoFactor.verifyCodeLength'))
    return
  }

  loading.value = true
  userApi.disable2FA(disableCode.value, () => {
    ElMessage.success(t('twoFactor.disableSuccess'))
    disableDialogVisible.value = false
    twoFactorEnabled.value = false
    disableCode.value = ''
    loading.value = false
  }, (code, msg) => {
    ElMessage.error(msg || t('twoFactor.disableFailed'))
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
  ElMessage.success(t('twoFactor.secretCopied'))
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
