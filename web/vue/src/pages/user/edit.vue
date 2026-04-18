<template>
  <div class="tw-p-6">
    <div class="tw-max-w-lg tw-mx-auto">
      <Card>
        <CardHeader>
          <CardTitle>
            {{ isEditMode ? t('common.edit') : t('common.add') }}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <form class="tw-space-y-5" @submit.prevent="onSubmit">
            <FormField v-slot="{ componentField }" name="name">
              <FormItem>
                <FormLabel>{{ t('user.username') }}</FormLabel>
                <FormControl>
                  <Input v-bind="componentField" autocomplete="username" />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="email">
              <FormItem>
                <FormLabel>{{ t('user.email') }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="email"
                    autocomplete="email"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <template v-if="!isEditMode">
              <FormField v-slot="{ componentField }" name="password">
                <FormItem>
                  <FormLabel>{{ t('user.password') }}</FormLabel>
                  <FormControl>
                    <Input
                      v-bind="componentField"
                      type="password"
                      :placeholder="t('user.passwordPlaceholder')"
                      autocomplete="new-password"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="confirm_password">
                <FormItem>
                  <FormLabel>{{ t('user.confirmPassword') }}</FormLabel>
                  <FormControl>
                    <Input
                      v-bind="componentField"
                      type="password"
                      :placeholder="t('user.passwordPlaceholder')"
                      autocomplete="new-password"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </template>

            <!-- Role: is_admin as Switch (true = admin, false = normal) -->
            <FormField v-slot="{ value, handleChange }" name="is_admin">
              <FormItem>
                <div class="tw-flex tw-items-center tw-justify-between tw-rounded-lg tw-border tw-p-3">
                  <div class="tw-space-y-0.5">
                    <FormLabel class="tw-text-base">{{ t('user.role') }}</FormLabel>
                    <p class="tw-text-sm tw-text-muted-foreground">
                      {{ value ? t('user.admin') : t('user.normalUser') }}
                    </p>
                  </div>
                  <FormControl>
                    <Switch
                      :model-value="value"
                      @update:model-value="handleChange"
                    />
                  </FormControl>
                </div>
                <FormMessage />
              </FormItem>
            </FormField>

            <!-- Status: as Switch (true = enabled, false = disabled) -->
            <FormField v-slot="{ value, handleChange }" name="status">
              <FormItem>
                <div class="tw-flex tw-items-center tw-justify-between tw-rounded-lg tw-border tw-p-3">
                  <div class="tw-space-y-0.5">
                    <FormLabel class="tw-text-base">{{ t('common.status') }}</FormLabel>
                    <p class="tw-text-sm tw-text-muted-foreground">
                      {{ value ? t('common.enabled') : t('common.disabled') }}
                    </p>
                  </div>
                  <FormControl>
                    <Switch
                      :model-value="value"
                      @update:model-value="handleChange"
                    />
                  </FormControl>
                </div>
                <FormMessage />
              </FormItem>
            </FormField>

            <div class="tw-flex tw-justify-center tw-gap-3 tw-pt-2">
              <Button type="submit">
                {{ t('common.save') }}
              </Button>
              <Button
                type="button"
                variant="outline"
                @click="cancel"
              >
                {{ t('common.cancel') }}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import { z } from 'zod'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle
} from '@/components/ui/card'
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@/components/ui/form'

import userService from '@/api/user'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const userId = ref(route.params.id || '')
const isEditMode = computed(() => !!userId.value)

// Build schema as a computed so it reacts to isEditMode
const validationSchema = computed(() =>
  toTypedSchema(
    z
      .object({
        name: z.string().min(1, t('user.usernameRequired')),
        email: z.string().email(t('user.emailRequired')),
        password: isEditMode.value
          ? z.string().optional()
          : z.string().min(1, t('user.passwordRequired')),
        confirm_password: isEditMode.value
          ? z.string().optional()
          : z.string().min(1, t('user.confirmPasswordRequired')),
        is_admin: z.boolean(),
        status: z.boolean()
      })
      .refine(
        data => {
          if (isEditMode.value) return true
          return data.password === data.confirm_password
        },
        {
          message: t('user.passwordMismatch'),
          path: ['confirm_password']
        }
      )
  )
)

const { handleSubmit, setValues } = useForm({
  validationSchema,
  initialValues: {
    name: '',
    email: '',
    password: '',
    confirm_password: '',
    is_admin: false,
    status: true
  }
})

onMounted(() => {
  if (!userId.value) return

  userService.detail(userId.value, data => {
    if (!data) return
    setValues({
      name: data.name,
      email: data.email,
      is_admin: data.is_admin === 1,
      status: data.status === 1
    })
  })
})

const onSubmit = handleSubmit(values => {
  const payload = {
    id: userId.value || undefined,
    name: values.name,
    email: values.email,
    is_admin: values.is_admin ? 1 : 0,
    status: values.status ? 1 : 0
  }

  if (!isEditMode.value) {
    payload.password = values.password
    payload.confirm_password = values.confirm_password
  }

  userService.update(payload, () => {
    router.push('/user')
  })
})

function cancel() {
  router.push('/user')
}
</script>
