<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ShieldCheck, ShieldOff, Copy, QrCode } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import {
  PinInput,
  PinInputGroup,
  PinInputSlot
} from '@/components/ui/pin-input'
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
// PinInput uses an array of single characters
const verifyCodeArr = ref([])
const disableCodeArr = ref([])

const getVerifyCode = () => verifyCodeArr.value.join('')
const getDisableCode = () => disableCodeArr.value.join('')

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
    verifyCodeArr.value = []
    setupDialogVisible.value = true
    loading.value = false
  })
}

const enable2FA = () => {
  const code = getVerifyCode()
  if (!code || code.length !== 6) {
    notify.warning(t('twoFactor.verifyCodeLength'))
    return
  }

  loading.value = true
  userApi.enable2FA(secret.value, code, () => {
    notify.success(t('twoFactor.enableSuccess'))
    setupDialogVisible.value = false
    twoFactorEnabled.value = true
    verifyCodeArr.value = []
    loading.value = false
  })
}

const showDisableDialog = () => {
  disableCodeArr.value = []
  disableDialogVisible.value = true
}

const disable2FA = () => {
  const code = getDisableCode()
  if (!code || code.length !== 6) {
    notify.warning(t('twoFactor.verifyCodeLength'))
    return
  }

  loading.value = true
  userApi.disable2FA(
    code,
    () => {
      notify.success(t('twoFactor.disableSuccess'))
      disableDialogVisible.value = false
      twoFactorEnabled.value = false
      disableCodeArr.value = []
      loading.value = false
    },
    (errorCode, msg) => {
      notify.error(msg || t('twoFactor.disableFailed'))
      loading.value = false
    }
  )
}

const copySecret = () => {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(secret.value).then(() => {
      notify.success(t('twoFactor.secretCopied'))
    })
  } else {
    // Fallback for browsers without Clipboard API
    const input = document.createElement('input')
    input.value = secret.value
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
    notify.success(t('twoFactor.secretCopied'))
  }
}
</script>

<template>
  <div class="tw-p-5">
    <Card class="tw-max-w-xl">
      <CardHeader>
        <CardTitle class="tw-flex tw-items-center tw-gap-2">
          <ShieldCheck
            v-if="twoFactorEnabled"
            class="tw-size-5 tw-text-green-600"
          />
          <ShieldOff
            v-else
            class="tw-size-5 tw-text-muted-foreground"
          />
          {{ t('twoFactor.title') }}
        </CardTitle>
      </CardHeader>

      <CardContent class="tw-space-y-4">
        <!-- 2FA disabled state -->
        <template v-if="!twoFactorEnabled">
          <div class="tw-rounded-md tw-border tw-border-blue-200 tw-bg-blue-50 tw-p-4 dark:tw-border-blue-800 dark:tw-bg-blue-950">
            <p class="tw-text-sm tw-font-medium tw-text-blue-800 dark:tw-text-blue-300">
              {{ t('twoFactor.alertTitle') }}
            </p>
            <p class="tw-mt-1 tw-text-sm tw-text-blue-700 dark:tw-text-blue-400">
              {{ t('twoFactor.alertDescription') }}
            </p>
          </div>

          <Button
            :disabled="loading"
            @click="setup2FA"
          >
            <QrCode class="tw-mr-2 tw-size-4" />
            {{ t('twoFactor.enable') }}
          </Button>
        </template>

        <!-- 2FA enabled state -->
        <template v-else>
          <div class="tw-rounded-md tw-border tw-border-green-200 tw-bg-green-50 tw-p-4 dark:tw-border-green-800 dark:tw-bg-green-950">
            <p class="tw-text-sm tw-font-medium tw-text-green-800 dark:tw-text-green-300">
              {{ t('twoFactor.enabledAlertTitle') }}
            </p>
            <p class="tw-mt-1 tw-text-sm tw-text-green-700 dark:tw-text-green-400">
              {{ t('twoFactor.enabledAlertDescription') }}
            </p>
          </div>

          <Button
            variant="destructive"
            @click="showDisableDialog"
          >
            <ShieldOff class="tw-mr-2 tw-size-4" />
            {{ t('twoFactor.disable') }}
          </Button>
        </template>
      </CardContent>
    </Card>

    <!-- Setup / Enable Dialog -->
    <Dialog v-model:open="setupDialogVisible">
      <DialogContent
        class="tw-max-w-md"
        @pointer-down-outside.prevent
      >
        <DialogHeader>
          <DialogTitle>{{ t('twoFactor.setup') }}</DialogTitle>
        </DialogHeader>

        <div
          v-if="qrCode"
          class="tw-space-y-4"
        >
          <!-- Step 1: scan QR -->
          <p class="tw-text-sm tw-text-muted-foreground">
            {{ t('twoFactor.scanQR') }}
          </p>
          <div class="tw-flex tw-justify-center">
            <img
              :src="qrCode"
              alt="QR Code"
              class="tw-size-48"
            >
          </div>

          <!-- Step 2: manual secret -->
          <div class="tw-space-y-1">
            <p class="tw-text-sm tw-text-muted-foreground">
              {{ t('twoFactor.manualEntry') }}
            </p>
            <div class="tw-flex tw-gap-2">
              <Input
                :model-value="secret"
                readonly
                class="tw-font-mono tw-text-xs"
              />
              <Button
                variant="outline"
                size="icon"
                :title="t('twoFactor.copySecret')"
                @click="copySecret"
              >
                <Copy class="tw-size-4" />
              </Button>
            </div>
          </div>

          <!-- Step 3: verify code -->
          <div class="tw-space-y-2">
            <Label>{{ t('twoFactor.verifyCodeStep') }}</Label>
            <PinInput
              v-model="verifyCodeArr"
              otp
              type="number"
              @complete="enable2FA"
            >
              <PinInputGroup>
                <PinInputSlot
                  v-for="i in 6"
                  :key="i"
                  :index="i - 1"
                />
              </PinInputGroup>
            </PinInput>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            @click="setupDialogVisible = false"
          >
            {{ t('common.cancel') }}
          </Button>
          <Button
            :disabled="loading"
            @click="enable2FA"
          >
            {{ t('twoFactor.confirm') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Disable Dialog -->
    <Dialog v-model:open="disableDialogVisible">
      <DialogContent
        class="tw-max-w-sm"
        @pointer-down-outside.prevent
      >
        <DialogHeader>
          <DialogTitle>{{ t('twoFactor.disableDialogTitle') }}</DialogTitle>
          <DialogDescription>
            {{ t('twoFactor.disableDialogDescription') }}
          </DialogDescription>
        </DialogHeader>

        <div class="tw-space-y-2">
          <Label>{{ t('twoFactor.verifyCode') }}</Label>
          <PinInput
            v-model="disableCodeArr"
            otp
            type="number"
            @complete="disable2FA"
          >
            <PinInputGroup>
              <PinInputSlot
                v-for="i in 6"
                :key="i"
                :index="i - 1"
              />
            </PinInputGroup>
          </PinInput>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            @click="disableDialogVisible = false"
          >
            {{ t('common.cancel') }}
          </Button>
          <Button
            variant="destructive"
            :disabled="loading"
            @click="disable2FA"
          >
            {{ t('twoFactor.confirmDisable') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
