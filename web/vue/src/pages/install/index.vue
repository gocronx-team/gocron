<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import { Loader2, Database, User } from 'lucide-vue-next'

import { FormField, FormItem, FormLabel, FormControl, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { useNotify } from '@/composables/useNotify'
import { useLoading } from '@/composables/useLoading'
import installService from '@/api/install'
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'

const { t, locale } = useI18n()
const router = useRouter()
const notify = useNotify()
const { loading, withLoading } = useLoading()

// ─── Language selection dialog ─────────────────────────────────────────────
const showLanguageDialog = ref(false)
const selectedLanguage = ref('')

const availableLanguages = [
  { value: 'zh-CN', label: '简体中文', icon: '🇨🇳' },
  { value: 'en-US', label: 'English', icon: '🇺🇸' }
]

// Dialog text is bilingual (keyed off selectedLanguage, not i18n locale)
const dialogTitle = computed(() =>
  selectedLanguage.value === 'en-US' ? 'Select Language' : '选择语言'
)
const dialogPrompt = computed(() =>
  selectedLanguage.value === 'en-US'
    ? 'Please select your preferred language'
    : '请选择您的首选语言'
)
const dialogConfirmText = computed(() =>
  selectedLanguage.value === 'en-US' ? 'Confirm' : '确认'
)

onMounted(() => {
  const savedLocale = localStorage.getItem('locale')
  selectedLanguage.value = savedLocale || 'en-US'
  showLanguageDialog.value = true
})

function selectLanguage(lang) {
  selectedLanguage.value = lang
}

function confirmLanguage() {
  if (!selectedLanguage.value) return
  locale.value = selectedLanguage.value
  localStorage.setItem('locale', selectedLanguage.value)
  showLanguageDialog.value = false
}

// ─── DB type state ─────────────────────────────────────────────────────────
const DEFAULT_PORTS = { mysql: 3306, postgres: 5432, sqlite: 0 }

// Track db_type locally so isSqlite and label computeds can react
const dbType = ref('mysql')
const isSqlite = computed(() => dbType.value === 'sqlite')

const dbNameLabel = computed(() =>
  isSqlite.value ? t('install.dbFilePath') : t('install.dbName')
)
const dbNamePlaceholder = computed(() =>
  isSqlite.value ? t('install.dbFilePathPlaceholder') : t('install.dbNamePlaceholder')
)

// ─── Form setup ────────────────────────────────────────────────────────────
// Schema is a computed ref so t() messages stay reactive on locale change
const validationSchema = computed(() =>
  toTypedSchema(
    z
      .object({
        db_type: z.enum(['mysql', 'postgres', 'sqlite']),
        db_host: z.string().optional(),
        db_port: z.coerce.number().optional(),
        db_username: z.string().optional(),
        db_password: z.string().optional(),
        db_name: z.string().min(1, t('install.enterDbName')),
        db_table_prefix: z.string().optional(),
        admin_username: z.string().min(1, t('install.enterAdminUsername')),
        admin_email: z.string().email(t('install.enterAdminEmail')),
        admin_password: z
          .string()
          .min(1, t('install.enterAdminPassword'))
          .min(8, t('install.passwordMinLength')),
        confirm_admin_password: z
          .string()
          .min(1, t('install.confirmAdminPassword'))
          .min(8, t('install.passwordMinLength'))
      })
      .superRefine((val, ctx) => {
        if (val.db_type !== 'sqlite') {
          if (!val.db_host) {
            ctx.addIssue({ path: ['db_host'], code: 'custom', message: t('install.enterDbHost') })
          }
          if (!val.db_port) {
            ctx.addIssue({ path: ['db_port'], code: 'custom', message: t('install.enterDbPort') })
          }
          if (!val.db_username) {
            ctx.addIssue({ path: ['db_username'], code: 'custom', message: t('install.enterDbUser') })
          }
          if (!val.db_password) {
            ctx.addIssue({ path: ['db_password'], code: 'custom', message: t('install.enterDbPassword') })
          }
        }
        if (val.admin_password && val.confirm_admin_password &&
            val.admin_password !== val.confirm_admin_password) {
          ctx.addIssue({
            path: ['confirm_admin_password'],
            code: 'custom',
            message: t('message.passwordMismatch')
          })
        }
      })
  )
)

const { handleSubmit, setFieldValue } = useForm({
  validationSchema,
  initialValues: {
    db_type: 'mysql',
    db_host: '127.0.0.1',
    db_port: 3306,
    db_username: '',
    db_password: '',
    db_name: '',
    db_table_prefix: '',
    admin_username: '',
    admin_email: '',
    admin_password: '',
    confirm_admin_password: ''
  }
})

// ─── DB type switching ─────────────────────────────────────────────────────
function handleDbTypeChange(val) {
  dbType.value = val
  setFieldValue('db_port', DEFAULT_PORTS[val])
  if (val === 'sqlite') {
    setFieldValue('db_host', '')
    setFieldValue('db_username', '')
    setFieldValue('db_password', '')
    setFieldValue('db_name', './data/gocron.db')
  } else {
    setFieldValue('db_host', '127.0.0.1')
    setFieldValue('db_name', '')
  }
}

// ─── Submit ────────────────────────────────────────────────────────────────
const onSubmit = handleSubmit(async (values) => {
  await withLoading(async () => {
    await new Promise((resolve) => {
      installService.store(values, () => {
        notify.success(t('install.installSuccess'))
        router.push('/')
        resolve()
      })
    })
  })
})
</script>

<template>
  <!-- Language selection dialog (shown on mount, not closable) -->
  <Dialog :open="showLanguageDialog">
    <DialogContent
      class="tw-max-w-sm"
      @interact-outside.prevent
      @escape-key-down.prevent
    >
      <DialogHeader>
        <DialogTitle class="tw-text-center">
          {{ dialogTitle }}
        </DialogTitle>
      </DialogHeader>

      <p class="tw-text-center tw-text-sm tw-text-muted-foreground tw-py-2">
        {{ dialogPrompt }}
      </p>

      <div class="tw-flex tw-flex-col tw-gap-3 tw-items-center tw-py-2">
        <button
          v-for="lang in availableLanguages"
          :key="lang.value"
          type="button"
          :class="[
            'tw-w-64 tw-h-14 tw-rounded-lg tw-border tw-flex tw-items-center tw-justify-center tw-gap-3 tw-text-base tw-font-medium tw-transition-all tw-duration-200',
            selectedLanguage === lang.value
              ? 'tw-border-primary tw-bg-primary/10 tw-text-primary'
              : 'tw-border-border tw-bg-background hover:tw-border-primary/50 hover:tw-bg-muted'
          ]"
          @click="selectLanguage(lang.value)"
        >
          <span class="tw-text-2xl">{{ lang.icon }}</span>
          <span>{{ lang.label }}</span>
        </button>
      </div>

      <DialogFooter>
        <Button
          class="tw-w-full"
          :disabled="!selectedLanguage"
          @click="confirmLanguage"
        >
          {{ dialogConfirmText }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- Main install page -->
  <div class="tw-min-h-screen tw-bg-muted/30 tw-p-6">
    <!-- Language switcher top-right -->
    <div class="tw-fixed tw-top-4 tw-right-4 tw-z-10">
      <LanguageSwitcher />
    </div>

    <div class="tw-max-w-2xl tw-mx-auto">
      <h1 class="tw-text-2xl tw-font-bold tw-text-center tw-mb-6">
        {{ t('install.welcome') }}
      </h1>

      <form @submit="onSubmit">
        <!-- Database Configuration -->
        <Card class="tw-mb-6">
          <CardHeader class="tw-pb-3">
            <CardTitle class="tw-flex tw-items-center tw-gap-2 tw-text-lg">
              <Database class="tw-size-5" />
              {{ t('install.dbConfig') }}
            </CardTitle>
          </CardHeader>
          <CardContent class="tw-space-y-4">
            <!-- DB Type -->
            <FormField v-slot="{ componentField }" name="db_type">
              <FormItem>
                <FormLabel>{{ t('install.dbType') }}</FormLabel>
                <FormControl>
                  <Select
                    v-bind="componentField"
                    @update:model-value="handleDbTypeChange"
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="mysql">
                        MySQL
                      </SelectItem>
                      <SelectItem value="postgres">
                        PostgreSQL
                      </SelectItem>
                      <SelectItem value="sqlite">
                        SQLite
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>

            <!-- Host + Port (hidden for sqlite) -->
            <div
              v-if="!isSqlite"
              class="tw-grid tw-grid-cols-2 tw-gap-4"
            >
              <FormField v-slot="{ componentField }" name="db_host">
                <FormItem>
                  <FormLabel>{{ t('install.dbHost') }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="db_port">
                <FormItem>
                  <FormLabel>{{ t('install.dbPort') }}</FormLabel>
                  <FormControl>
                    <Input
                      v-bind="componentField"
                      type="number"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>

            <!-- Username + Password (hidden for sqlite) -->
            <div
              v-if="!isSqlite"
              class="tw-grid tw-grid-cols-2 tw-gap-4"
            >
              <FormField v-slot="{ componentField }" name="db_username">
                <FormItem>
                  <FormLabel>{{ t('install.dbUser') }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="db_password">
                <FormItem>
                  <FormLabel>{{ t('install.dbPassword') }}</FormLabel>
                  <FormControl>
                    <Input
                      v-bind="componentField"
                      type="password"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>

            <!-- DB Name + Table Prefix (always shown) -->
            <div class="tw-grid tw-grid-cols-2 tw-gap-4">
              <FormField v-slot="{ componentField }" name="db_name">
                <FormItem>
                  <FormLabel>{{ dbNameLabel }}</FormLabel>
                  <FormControl>
                    <Input
                      v-bind="componentField"
                      :placeholder="dbNamePlaceholder"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="db_table_prefix">
                <FormItem>
                  <FormLabel>{{ t('install.dbTablePrefix') }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>
          </CardContent>
        </Card>

        <!-- Admin Account Configuration -->
        <Card class="tw-mb-6">
          <CardHeader class="tw-pb-3">
            <CardTitle class="tw-flex tw-items-center tw-gap-2 tw-text-lg">
              <User class="tw-size-5" />
              {{ t('install.adminConfig') }}
            </CardTitle>
          </CardHeader>
          <CardContent class="tw-space-y-4">
            <!-- Username + Email -->
            <div class="tw-grid tw-grid-cols-2 tw-gap-4">
              <FormField v-slot="{ componentField }" name="admin_username">
                <FormItem>
                  <FormLabel>{{ t('install.adminUsername') }}</FormLabel>
                  <FormControl>
                    <Input v-bind="componentField" />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="admin_email">
                <FormItem>
                  <FormLabel>{{ t('install.adminEmail') }}</FormLabel>
                  <FormControl>
                    <Input
                      v-bind="componentField"
                      type="email"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>

            <!-- Password + Confirm Password -->
            <div class="tw-grid tw-grid-cols-2 tw-gap-4">
              <FormField v-slot="{ componentField }" name="admin_password">
                <FormItem>
                  <FormLabel>{{ t('install.adminPassword') }}</FormLabel>
                  <FormControl>
                    <Input
                      v-bind="componentField"
                      type="password"
                      :placeholder="t('install.passwordPlaceholder')"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>

              <FormField v-slot="{ componentField }" name="confirm_admin_password">
                <FormItem>
                  <FormLabel>{{ t('install.confirmPassword') }}</FormLabel>
                  <FormControl>
                    <Input
                      v-bind="componentField"
                      type="password"
                      :placeholder="t('install.passwordPlaceholder')"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>
          </CardContent>
        </Card>

        <Separator class="tw-my-4" />

        <!-- Submit -->
        <Button
          type="submit"
          class="tw-w-full"
          :disabled="loading"
        >
          <Loader2
            v-if="loading"
            class="tw-size-4 tw-animate-spin tw-mr-2"
          />
          {{ loading ? t('install.installing') : t('install.install') }}
        </Button>
      </form>
    </div>
  </div>
</template>
