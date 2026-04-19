<!-- System install wizard — standalone page, no auth required -->
<template>
  <div class="install-page">
    <!-- Language bar -->
    <div class="install-lang-bar">
      <ElSelect v-model="locale" size="small" style="width: 120px">
        <ElOption value="zh" :label="$t('install.langZh')" />
        <ElOption value="en" :label="$t('install.langEn')" />
      </ElSelect>
    </div>

    <div class="install-card-wrap">
      <ElCard class="install-card" shadow="always">
        <!-- Header -->
        <div class="install-header">
          <h2 class="install-title">{{ $t('install.title') }}</h2>
        </div>

        <!-- Steps indicator -->
        <ElSteps :active="currentStep" finish-status="success" align-center class="install-steps">
          <ElStep :title="$t('install.stepDb')" />
          <ElStep :title="$t('install.stepAdmin')" />
          <ElStep :title="$t('install.stepEmail')" />
        </ElSteps>

        <!-- Step 1: Database -->
        <ElForm
          v-show="currentStep === 0"
          ref="formDbRef"
          :model="form"
          :rules="rulesDb"
          label-width="140px"
          class="install-form"
          @submit.prevent
        >
          <ElFormItem :label="$t('install.dbType')" prop="db_type">
            <ElSelect v-model="form.db_type" style="width: 100%" @change="handleDbTypeChange">
              <ElOption v-for="db in dbList" :key="db.value" :value="db.value" :label="db.label" />
            </ElSelect>
          </ElFormItem>

          <template v-if="form.db_type !== 'sqlite'">
            <ElRow :gutter="16">
              <ElCol :span="14">
                <ElFormItem :label="$t('install.dbHost')" prop="db_host">
                  <ElInput v-model="form.db_host" />
                </ElFormItem>
              </ElCol>
              <ElCol :span="10">
                <ElFormItem :label="$t('install.dbPort')" prop="db_port">
                  <ElInputNumber
                    v-model="form.db_port"
                    :min="1"
                    :max="65535"
                    controls-position="right"
                    style="width: 100%"
                  />
                </ElFormItem>
              </ElCol>
            </ElRow>

            <ElRow :gutter="16">
              <ElCol :span="12">
                <ElFormItem :label="$t('install.dbUsername')" prop="db_username">
                  <ElInput v-model="form.db_username" autocomplete="off" />
                </ElFormItem>
              </ElCol>
              <ElCol :span="12">
                <ElFormItem :label="$t('install.dbPassword')" prop="db_password">
                  <ElInput v-model="form.db_password" type="password" show-password autocomplete="off" />
                </ElFormItem>
              </ElCol>
            </ElRow>
          </template>

          <ElRow :gutter="16">
            <ElCol :span="12">
              <ElFormItem
                :label="form.db_type === 'sqlite' ? $t('install.dbFilePath') : $t('install.dbName')"
                prop="db_name"
              >
                <ElInput
                  v-model="form.db_name"
                  :placeholder="form.db_type === 'sqlite' ? $t('install.dbFilePathPlaceholder') : ''"
                />
              </ElFormItem>
            </ElCol>
            <ElCol :span="12">
              <ElFormItem :label="$t('install.dbTablePrefix')" prop="db_table_prefix">
                <ElInput v-model="form.db_table_prefix" />
              </ElFormItem>
            </ElCol>
          </ElRow>

          <ElFormItem>
            <ElButton type="primary" :loading="submitting" @click="goNext(0)" v-ripple>
              {{ $t('install.next') }}
            </ElButton>
          </ElFormItem>
        </ElForm>

        <!-- Step 2: Admin account -->
        <ElForm
          v-show="currentStep === 1"
          ref="formAdminRef"
          :model="form"
          :rules="rulesAdmin"
          label-width="140px"
          class="install-form"
          @submit.prevent
        >
          <ElRow :gutter="16">
            <ElCol :span="12">
              <ElFormItem :label="$t('install.adminUsername')" prop="admin_username">
                <ElInput v-model="form.admin_username" autocomplete="username" />
              </ElFormItem>
            </ElCol>
            <ElCol :span="12">
              <ElFormItem :label="$t('install.adminEmail')" prop="admin_email">
                <ElInput v-model="form.admin_email" autocomplete="email" />
              </ElFormItem>
            </ElCol>
          </ElRow>

          <ElRow :gutter="16">
            <ElCol :span="12">
              <ElFormItem :label="$t('install.adminPassword')" prop="admin_password">
                <ElInput
                  v-model="form.admin_password"
                  type="password"
                  show-password
                  :placeholder="$t('install.passwordPlaceholder')"
                  autocomplete="new-password"
                />
              </ElFormItem>
            </ElCol>
            <ElCol :span="12">
              <ElFormItem :label="$t('install.confirmPassword')" prop="confirm_admin_password">
                <ElInput
                  v-model="form.confirm_admin_password"
                  type="password"
                  show-password
                  :placeholder="$t('install.passwordPlaceholder')"
                  autocomplete="new-password"
                />
              </ElFormItem>
            </ElCol>
          </ElRow>

          <ElFormItem>
            <ElButton @click="goPrev">{{ $t('install.prev') }}</ElButton>
            <ElButton type="primary" @click="goNext(1)" v-ripple>{{ $t('install.next') }}</ElButton>
          </ElFormItem>
        </ElForm>

        <!-- Step 3: Email config (optional) -->
        <ElForm
          v-show="currentStep === 2"
          ref="formEmailRef"
          :model="form"
          label-width="140px"
          class="install-form"
          @submit.prevent
        >
          <ElAlert
            :title="$t('install.emailOptionalTip')"
            type="info"
            :closable="false"
            show-icon
            style="margin-bottom: 20px"
          />

          <ElRow :gutter="16">
            <ElCol :span="14">
              <ElFormItem :label="$t('install.emailHost')">
                <ElInput v-model="form.email_host" />
              </ElFormItem>
            </ElCol>
            <ElCol :span="10">
              <ElFormItem :label="$t('install.emailPort')">
                <ElInputNumber
                  v-model="form.email_port"
                  :min="1"
                  :max="65535"
                  controls-position="right"
                  style="width: 100%"
                />
              </ElFormItem>
            </ElCol>
          </ElRow>

          <ElRow :gutter="16">
            <ElCol :span="12">
              <ElFormItem :label="$t('install.emailUsername')">
                <ElInput v-model="form.email_username" autocomplete="off" />
              </ElFormItem>
            </ElCol>
            <ElCol :span="12">
              <ElFormItem :label="$t('install.emailPassword')">
                <ElInput v-model="form.email_password" type="password" show-password autocomplete="off" />
              </ElFormItem>
            </ElCol>
          </ElRow>

          <ElFormItem>
            <ElButton @click="goPrev">{{ $t('install.prev') }}</ElButton>
            <ElButton @click="handleSubmit" :loading="submitting" v-ripple>
              {{ $t('install.skip') }}
            </ElButton>
            <ElButton type="primary" @click="handleSubmit" :loading="submitting" v-ripple>
              {{ $t('install.submit') }}
            </ElButton>
          </ElFormItem>
        </ElForm>
      </ElCard>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed } from 'vue'
  import { useRouter } from 'vue-router'
  import { useI18n } from 'vue-i18n'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchInstall } from '@/api/install'

  defineOptions({ name: 'Install' })

  const router = useRouter()
  const { t, locale } = useI18n()

  // ── State ─────────────────────────────────────────────────────────────────────

  const currentStep = ref(0)
  const submitting = ref(false)

  const formDbRef = ref<FormInstance>()
  const formAdminRef = ref<FormInstance>()
  const formEmailRef = ref<FormInstance>()

  const form = reactive({
    // DB
    db_type: 'mysql',
    db_host: '127.0.0.1',
    db_port: 3306,
    db_username: '',
    db_password: '',
    db_name: '',
    db_table_prefix: '',
    // Admin
    admin_username: '',
    admin_password: '',
    confirm_admin_password: '',
    admin_email: '',
    // Email (optional)
    email_host: '',
    email_port: 465,
    email_username: '',
    email_password: ''
  })

  const dbList = [
    { value: 'mysql', label: 'MySQL' },
    { value: 'postgres', label: 'PostgreSQL' },
    { value: 'sqlite', label: 'SQLite' }
  ]

  const defaultPorts: Record<string, number> = {
    mysql: 3306,
    postgres: 5432,
    sqlite: 0
  }

  // ── Validation rules ─────────────────────────────────────────────────────────

  const rulesDb = computed<FormRules>(() => ({
    db_type: [{ required: true, message: t('install.selectDb'), trigger: 'change' }],
    db_host:
      form.db_type !== 'sqlite'
        ? [{ required: true, message: t('install.enterDbHost'), trigger: 'blur' }]
        : [],
    db_port:
      form.db_type !== 'sqlite'
        ? [{ required: true, type: 'number', message: t('install.enterDbPort'), trigger: 'blur' }]
        : [],
    db_username:
      form.db_type !== 'sqlite'
        ? [{ required: true, message: t('install.enterDbUser'), trigger: 'blur' }]
        : [],
    db_password:
      form.db_type !== 'sqlite'
        ? [{ required: true, message: t('install.enterDbPassword'), trigger: 'blur' }]
        : [],
    db_name: [{ required: true, message: t('install.enterDbName'), trigger: 'blur' }]
  }))

  const rulesAdmin = computed<FormRules>(() => ({
    admin_username: [{ required: true, message: t('install.enterAdminUsername'), trigger: 'blur' }],
    admin_email: [
      { required: true, type: 'email', message: t('install.enterAdminEmail'), trigger: 'blur' }
    ],
    admin_password: [
      { required: true, message: t('install.enterAdminPassword'), trigger: 'blur' },
      { min: 8, message: t('install.passwordMinLength'), trigger: 'blur' }
    ],
    confirm_admin_password: [
      { required: true, message: t('install.confirmAdminPassword'), trigger: 'blur' },
      {
        validator: (_rule: any, value: string, callback: any) => {
          if (value !== form.admin_password) {
            callback(new Error(t('install.passwordMismatch')))
          } else {
            callback()
          }
        },
        trigger: 'blur'
      }
    ]
  }))

  // ── Handlers ─────────────────────────────────────────────────────────────────

  function handleDbTypeChange(dbType: string) {
    form.db_port = defaultPorts[dbType] ?? 0
    if (dbType === 'sqlite') {
      form.db_host = ''
      form.db_username = ''
      form.db_password = ''
      form.db_name = './data/gocron.db'
    } else {
      form.db_host = '127.0.0.1'
      form.db_name = ''
    }
  }

  async function goNext(step: number) {
    if (step === 0) {
      const valid = await formDbRef.value?.validate().catch(() => false)
      if (!valid) return
    } else if (step === 1) {
      const valid = await formAdminRef.value?.validate().catch(() => false)
      if (!valid) return
    }
    currentStep.value = step + 1
  }

  function goPrev() {
    if (currentStep.value > 0) currentStep.value--
  }

  async function handleSubmit() {
    submitting.value = true
    try {
      await fetchInstall({
        db_type: form.db_type,
        db_host: form.db_host,
        db_port: form.db_port,
        db_username: form.db_username,
        db_password: form.db_password,
        db_name: form.db_name,
        db_table_prefix: form.db_table_prefix,
        admin_username: form.admin_username,
        admin_password: form.admin_password,
        confirm_admin_password: form.confirm_admin_password,
        admin_email: form.admin_email
      })
      ElMessage.success(t('install.installSuccess'))
      router.push('/auth/login')
    } catch {
      // error toast handled by http interceptor
    } finally {
      submitting.value = false
    }
  }
</script>

<style scoped>
  .install-page {
    min-height: 100vh;
    background: var(--el-bg-color-page, #f5f7fa);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 16px;
  }

  .install-lang-bar {
    position: fixed;
    top: 16px;
    right: 20px;
    z-index: 100;
  }

  .install-card-wrap {
    width: 100%;
    max-width: 760px;
  }

  .install-card {
    width: 100%;
  }

  .install-header {
    text-align: center;
    margin-bottom: 8px;
  }

  .install-title {
    font-size: 22px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    margin: 0 0 4px;
  }

  .install-steps {
    margin-bottom: 28px;
  }

  .install-form {
    padding: 0 8px;
  }

  @media (max-width: 768px) {
    .install-page {
      padding: 24px 8px;
      justify-content: flex-start;
    }

    .install-card-wrap {
      max-width: 100%;
    }

    .install-steps {
      margin-bottom: 20px;
    }

    /* Install form: ElRow columns stack full-width — handled globally,
       but also constrain the card's internal padding for small screens */
    .install-form {
      padding: 0;
    }
  }
</style>
