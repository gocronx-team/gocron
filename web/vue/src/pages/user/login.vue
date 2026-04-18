<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import { AlertCircle, Loader2 } from 'lucide-vue-next'

import { Form, FormField, FormItem, FormLabel, FormControl, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useNotify } from '@/composables/useNotify'
import { useUserStore } from '@/stores/user'
import { useLoading } from '@/composables/useLoading'
import userService from '@/api/user'
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const { loading, withLoading } = useLoading()
const notify = useNotify()

const require2FA = ref(false)
const errorMessage = ref('')

// Schema is computed so t() can pick up locale changes reactively
const validationSchema = computed(() =>
  toTypedSchema(
    z.object({
      username: z.string().min(1, t('login.usernameRequired')),
      password: z.string().min(1, t('login.passwordRequired')),
      verifyCode: require2FA.value
        ? z.string().min(1, t('login.verifyCodeRequired'))
        : z.string().optional()
    })
  )
)

async function onSubmit(values) {
  errorMessage.value = ''

  await withLoading(async () => {
    await new Promise((resolve) => {
      userService.login(
        values.username,
        values.password,
        values.verifyCode || undefined,
        (data) => {
          if (data.require_2fa) {
            require2FA.value = true
            errorMessage.value = ''
            resolve()
            return
          }

          userStore.setUser({
            token: data.token,
            uid: data.uid,
            username: data.username,
            isAdmin: data.is_admin
          })

          router.push(route.query.redirect || '/')
          resolve()
        },
        (_code, message) => {
          errorMessage.value = message || t('login.login') + ' ' + 'failed'
          resolve()
        }
      )
    })
  })
}
</script>

<template>
  <div
    class="tw-min-h-screen tw-flex tw-items-center tw-justify-center tw-bg-muted/30 tw-p-4"
  >
    <!-- Language switcher fixed to top-left -->
    <div class="tw-fixed tw-top-4 tw-left-4 tw-z-10">
      <LanguageSwitcher />
    </div>

    <Card class="tw-w-full tw-max-w-sm">
      <CardHeader class="tw-pb-4">
        <CardTitle class="tw-text-center tw-text-2xl tw-font-semibold">
          {{ t('login.title') }}
        </CardTitle>
      </CardHeader>

      <CardContent>
        <!-- Error alert -->
        <div
          v-if="errorMessage"
          class="tw-flex tw-items-start tw-gap-2 tw-rounded-md tw-border tw-border-destructive/50 tw-bg-destructive/10 tw-px-3 tw-py-2 tw-text-sm tw-text-destructive tw-mb-4"
        >
          <AlertCircle class="tw-size-4 tw-mt-0.5 tw-shrink-0" />
          <span>{{ errorMessage }}</span>
        </div>

        <Form
          :validation-schema="validationSchema"
          class="tw-space-y-4"
          @submit="onSubmit"
        >
          <!-- Username -->
          <FormField v-slot="{ componentField }" name="username">
            <FormItem>
              <FormLabel>{{ t('login.username') }}</FormLabel>
              <FormControl>
                <Input
                  v-bind="componentField"
                  :placeholder="t('login.usernamePlaceholder')"
                  autocomplete="username"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <!-- Password -->
          <FormField v-slot="{ componentField }" name="password">
            <FormItem>
              <FormLabel>{{ t('login.password') }}</FormLabel>
              <FormControl>
                <Input
                  v-bind="componentField"
                  type="password"
                  :placeholder="t('login.passwordPlaceholder')"
                  autocomplete="current-password"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <!-- 2FA code (conditionally shown) -->
          <FormField
            v-if="require2FA"
            v-slot="{ componentField }"
            name="verifyCode"
          >
            <FormItem>
              <FormLabel>{{ t('login.verifyCode') }}</FormLabel>
              <FormControl>
                <Input
                  v-bind="componentField"
                  :placeholder="t('login.verifyCodePlaceholder')"
                  maxlength="6"
                  autocomplete="one-time-code"
                  inputmode="numeric"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <!-- Submit button -->
          <Button
            type="submit"
            class="tw-w-full"
            :disabled="loading"
          >
            <Loader2
              v-if="loading"
              class="tw-size-4 tw-animate-spin"
            />
            {{ t('login.login') }}
          </Button>
        </Form>
      </CardContent>
    </Card>
  </div>
</template>
