<template>
  <div class="tw-p-6">
    <div class="tw-max-w-lg tw-mx-auto">
      <Card>
        <CardHeader>
          <CardTitle>{{ t('user.changePassword') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <form class="tw-space-y-5" @submit.prevent="onSubmit">
            <FormField v-slot="{ componentField }" name="new_password">
              <FormItem>
                <FormLabel>{{ t('user.newPassword') }}</FormLabel>
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

            <FormField v-slot="{ componentField }" name="confirm_new_password">
              <FormItem>
                <FormLabel>{{ t('user.confirmNewPassword') }}</FormLabel>
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

            <div class="tw-flex tw-justify-center tw-gap-3 tw-pt-2">
              <Button type="submit">
                {{ t('common.save') }}
              </Button>
              <Button type="button" variant="outline" @click="cancel">
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
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import { z } from 'zod'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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

const userId = route.params.id

const schema = toTypedSchema(
  z.object({
    new_password: z.string().min(6, t('user.newPasswordRequired')),
    confirm_new_password: z.string().min(1, t('user.confirmPasswordRequired'))
  }).refine(
    data => data.new_password === data.confirm_new_password,
    {
      message: t('user.passwordMismatch'),
      path: ['confirm_new_password']
    }
  )
)

const { handleSubmit } = useForm({ validationSchema: schema })

const onSubmit = handleSubmit(values => {
  userService.editPassword(
    { id: userId, new_password: values.new_password, confirm_new_password: values.confirm_new_password },
    () => {
      router.push('/user')
    }
  )
})

function cancel() {
  router.push('/user')
}
</script>
