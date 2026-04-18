<template>
  <div class="tw-p-6">
    <div class="tw-max-w-lg tw-mx-auto">
      <Card>
        <CardHeader>
          <CardTitle>{{ isEdit ? t('host.list') : t('host.createNew') }}</CardTitle>
        </CardHeader>
        <CardContent>
          <form class="tw-space-y-5" @submit.prevent="onSubmit">
            <FormField v-slot="{ componentField }" name="alias">
              <FormItem>
                <FormLabel>{{ t('host.alias') }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    :placeholder="t('host.aliasPlaceholder')"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="name">
              <FormItem>
                <FormLabel>{{ t('host.name') }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    :placeholder="t('host.namePlaceholder')"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="port">
              <FormItem>
                <FormLabel>{{ t('host.port') }}</FormLabel>
                <FormControl>
                  <Input
                    v-bind="componentField"
                    type="number"
                    :placeholder="t('host.portPlaceholder')"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <FormField v-slot="{ componentField }" name="remark">
              <FormItem>
                <FormLabel>{{ t('host.remark') }}</FormLabel>
                <FormControl>
                  <textarea
                    v-bind="componentField"
                    rows="5"
                    class="tw-flex tw-min-h-[80px] tw-w-full tw-rounded-md tw-border tw-border-input tw-bg-background tw-px-3 tw-py-2 tw-text-sm tw-ring-offset-background placeholder:tw-text-muted-foreground focus-visible:tw-outline-none focus-visible:tw-ring-2 focus-visible:tw-ring-ring focus-visible:tw-ring-offset-2 disabled:tw-cursor-not-allowed disabled:tw-opacity-50"
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
import { computed, watch } from 'vue'
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

import hostService from '@/api/host'
import { useNotify } from '@/composables/useNotify'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const notify = useNotify()

const isEdit = computed(() => !!route.params.id)

const schema = toTypedSchema(
  z.object({
    alias: z.string().min(1, t('host.aliasRequired')),
    name: z.string().min(1, t('host.nameRequired')),
    port: z.coerce
      .number({ invalid_type_error: t('host.portInvalid') })
      .min(1, t('host.portInvalid'))
      .max(65535, t('host.portInvalid')),
    remark: z.string().optional()
  })
)

const { handleSubmit, setValues, resetForm } = useForm({
  validationSchema: schema,
  initialValues: {
    alias: '',
    name: '',
    port: 5921,
    remark: ''
  }
})

function loadForm() {
  resetForm({
    values: {
      alias: '',
      name: '',
      port: 5921,
      remark: ''
    }
  })
  const id = route.params.id
  if (!id) {
    return
  }
  hostService.detail(id, (data) => {
    if (!data) {
      notify.error(t('message.dataNotFound'))
      router.push('/host')
      return
    }
    setValues({
      alias: data.alias,
      name: data.name,
      port: data.port,
      remark: data.remark || ''
    })
  })
}

loadForm()

watch(
  () => route.params.id,
  () => {
    loadForm()
  }
)

const onSubmit = handleSubmit((values) => {
  const payload = {
    id: route.params.id || '',
    ...values
  }
  hostService.update(payload, () => {
    router.push('/host')
  })
})

function cancel() {
  router.push('/host')
}
</script>
